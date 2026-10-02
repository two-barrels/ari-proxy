package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/client/native"
)

// TestLiveRecordingFileDispatch compares a bounded existing file through native
// and proxy streaming, including a retried chunk and EOF.
// It never publishes content or modifies the stored recording.
func TestLiveRecordingFileDispatch(t *testing.T) {
	if os.Getenv("ARI_LIVE_RECORDING_TEST") == "" {
		t.Skip("set ARI_LIVE_RECORDING_TEST=1 to run against a live test PBX")
	}
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	upstream := native.New(&native.Options{URL: strings.TrimRight(os.Getenv("ARI_LIVE_URL"), "/"), Username: os.Getenv("ARI_LIVE_USERNAME"), Password: strings.TrimSuffix(password, "\n"), Application: "ari-testing", HTTPClient: &http.Client{Timeout: 10 * time.Second}})
	keys, err := upstream.StoredRecording().List(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		t.Skip("no stored recordings available")
	}
	info, err := upstream.Asterisk().Info(nil)
	if err != nil {
		t.Fatal(err)
	}
	key := &ari.Key{App: "ari-testing", Node: info.SystemInfo.EntityID, Kind: ari.StoredRecordingKey, ID: keys[0].ID}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari, s.mbus = upstream, bus
	request := func(kind string, file *proxy.RecordingFileRequest) *proxy.Response {
		t.Helper()
		bus.response = nil
		s.dispatchRequest(ctx, "reply", &proxy.Request{Kind: kind, Key: key, RecordingFile: file})
		if bus.response == nil {
			t.Fatal("no proxy response")
		}
		wire, err := json.Marshal(bus.response)
		if err != nil {
			t.Fatal(err)
		}
		var response proxy.Response
		if err := json.Unmarshal(wire, &response); err != nil {
			t.Fatal(err)
		}
		if err := response.Err(); err != nil {
			t.Fatal(err)
		}
		return &response
	}
	opened := request("RecordingStoredFileOpen", nil)
	if opened.RecordingFile == nil || opened.RecordingFile.Session == "" {
		t.Fatal("missing file session")
	}
	session := opened.RecordingFile.Session
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		bus.response = nil
		s.dispatchRequest(cleanupCtx, "reply", &proxy.Request{Kind: "RecordingStoredFileClose", Key: key, RecordingFile: &proxy.RecordingFileRequest{Session: session}})
		if bus.response == nil || bus.response.Err() != nil {
			t.Errorf("recording session close failed: %+v", bus.response)
		}
		if len(s.recordingFiles) != 0 {
			t.Errorf("recording session was not closed")
		}
	}()
	chunk := request("RecordingStoredFileRead", &proxy.RecordingFileRequest{Session: session, Sequence: 0})
	if chunk.RecordingFile == nil || len(chunk.RecordingFile.Data) > recordingFileChunkSize {
		t.Fatal("invalid first chunk")
	}
	t.Logf("read %d bytes of first chunk; content type %q; declared size %d", len(chunk.RecordingFile.Data), opened.RecordingFile.ContentType, opened.RecordingFile.Size)
	retry := request("RecordingStoredFileRead", &proxy.RecordingFileRequest{Session: session, Sequence: 0})
	if retry.RecordingFile == nil || !bytes.Equal(chunk.RecordingFile.Data, retry.RecordingFile.Data) || chunk.RecordingFile.EOF != retry.RecordingFile.EOF {
		t.Fatal("retry did not return the same chunk")
	}
	const maxBytes = 4 * 1024 * 1024
	if opened.RecordingFile.Size < 0 || opened.RecordingFile.Size > maxBytes {
		t.Log("full comparison omitted: size is unknown or exceeds 4 MiB limit")
		return
	}
	hash := sha256.New()
	var total int64
	for sequence := uint64(0); ; sequence++ {
		if sequence > 0 {
			chunk = request("RecordingStoredFileRead", &proxy.RecordingFileRequest{Session: session, Sequence: sequence})
		}
		if chunk.RecordingFile == nil || chunk.RecordingFile.Sequence != sequence || len(chunk.RecordingFile.Data) > recordingFileChunkSize {
			t.Fatal("invalid recording chunk")
		}
		total += int64(len(chunk.RecordingFile.Data))
		if total > maxBytes {
			t.Fatal("recording exceeded comparison limit")
		}
		_, _ = hash.Write(chunk.RecordingFile.Data)
		if chunk.RecordingFile.EOF {
			break
		}
		if len(chunk.RecordingFile.Data) == 0 {
			t.Fatal("empty chunk without EOF")
		}
	}
	file, err := upstream.StoredRecording().File(ctx, keys[0])
	if err != nil {
		t.Fatal(err)
	}
	defer file.Body.Close()
	nativeHash := sha256.New()
	nativeSize, err := io.Copy(nativeHash, io.LimitReader(file.Body, maxBytes+1))
	if err != nil {
		t.Fatal(err)
	}
	if nativeSize != total || !bytes.Equal(hash.Sum(nil), nativeHash.Sum(nil)) || total != opened.RecordingFile.Size {
		t.Fatalf("native/proxy content mismatch: native=%d, proxy=%d, declared=%d", nativeSize, total, opened.RecordingFile.Size)
	}
	t.Logf("native/proxy bytes match: %d bytes; retry and EOF passed", total)
}
