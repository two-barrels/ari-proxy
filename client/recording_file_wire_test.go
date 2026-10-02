package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type recordingFileBus struct {
	messagebus.Client
	payload []byte
	reads   int
	closes  int
	status  int
}

func (b *recordingFileBus) Request(_ string, request *proxy.Request) (*proxy.Response, error) {
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var req proxy.Request
	if err := json.Unmarshal(encoded, &req); err != nil {
		return nil, err
	}
	var response proxy.Response
	switch req.Kind {
	case "RecordingStoredFileOpen":
		if b.status != 0 {
			response = proxy.Response{Error: "recording not found", StatusCode: b.status}
			break
		}
		response.RecordingFile = &proxy.RecordingFileResponse{Session: "session-1", ContentType: "audio/wav", Size: int64(len(b.payload))}
	case "RecordingStoredFileRead":
		b.reads++
		if req.RecordingFile == nil || req.RecordingFile.Session != "session-1" {
			return nil, errors.New("bad file session")
		}
		start := int(req.RecordingFile.Sequence) * maxRecordingFileChunk
		end := start + maxRecordingFileChunk
		if end > len(b.payload) {
			end = len(b.payload)
		}
		if start > len(b.payload) {
			return nil, errors.New("bad chunk sequence")
		}
		response.RecordingFile = &proxy.RecordingFileResponse{Session: "session-1", Sequence: req.RecordingFile.Sequence,
			Data: b.payload[start:end], EOF: end == len(b.payload)}
	case "RecordingStoredFileClose":
		b.closes++
	default:
		return nil, errors.New("unexpected request kind")
	}
	encoded, err = json.Marshal(response)
	if err != nil {
		return nil, err
	}
	var delivered proxy.Response
	if err := json.Unmarshal(encoded, &delivered); err != nil {
		return nil, err
	}
	return &delivered, nil
}

func TestRecordingFileClientReadsBoundedChunks(t *testing.T) {
	payload := bytes.Repeat([]byte{0, 1, 2, 0xff}, maxRecordingFileChunk/2+7)
	bus := &recordingFileBus{payload: payload}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.StoredRecordingKey, ID: "call-1"}
	file, err := c.StoredRecording().File(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	if file.ContentType != "audio/wav" || file.Size != int64(len(payload)) || bus.reads != 0 {
		t.Fatalf("metadata=%+v reads=%d", file, bus.reads)
	}
	got, err := io.ReadAll(file.Body)
	if err != nil || !bytes.Equal(got, payload) || bus.reads != 3 {
		t.Fatalf("bytes=%d reads=%d error=%v", len(got), bus.reads, err)
	}
	if err := file.Body.Close(); err != nil || bus.closes != 1 {
		t.Fatalf("close error=%v closes=%d", err, bus.closes)
	}
	if err := file.Body.Close(); err != nil || bus.closes != 1 {
		t.Fatalf("second close error=%v closes=%d", err, bus.closes)
	}
	if _, err := c.StoredRecording().File(context.Background(), ari.NewKey(ari.StoredRecordingKey, "call-1")); err == nil {
		t.Fatal("expected target node error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.StoredRecording().File(ctx, key); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled open error=%v", err)
	}
	active, err := c.StoredRecording().File(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	part := make([]byte, 7)
	if _, err := active.Body.Read(part); err != nil {
		t.Fatal(err)
	}
	reads := bus.reads
	if err := active.Body.Close(); err != nil || bus.reads != reads || bus.closes != 2 {
		t.Fatalf("partial close error=%v reads=%d closes=%d", err, bus.reads, bus.closes)
	}
	activeCtx, cancelActive := context.WithCancel(context.Background())
	active, err = c.StoredRecording().File(activeCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	cancelActive()
	if _, err := active.Body.Read(part); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled read error=%v", err)
	}
	if err := active.Body.Close(); err != nil || bus.reads != reads || bus.closes != 3 {
		t.Fatalf("canceled close error=%v reads=%d closes=%d", err, bus.reads, bus.closes)
	}
	bus.status = http.StatusNotFound
	if _, err := c.StoredRecording().File(context.Background(), key); err == nil {
		t.Fatal("expected not found error")
	} else {
		var coded interface{ Code() int }
		if !errors.As(err, &coded) || coded.Code() != http.StatusNotFound {
			t.Fatalf("error=%v", err)
		}
	}
}
