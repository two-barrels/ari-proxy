// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type recordingFileClient struct {
	ari.Client
	recording ari.StoredRecording
}

func (c recordingFileClient) StoredRecording() ari.StoredRecording { return c.recording }

type recordingFileSource struct {
	ari.StoredRecording
	key  *ari.Key
	body *trackedFileBody
}

type trackedFileBody struct {
	*bytes.Reader
	closed bool
}

func (b *trackedFileBody) Close() error { b.closed = true; return nil }

func (s *recordingFileSource) File(_ context.Context, key *ari.Key) (*ari.RecordingFile, error) {
	s.key = key
	return &ari.RecordingFile{Body: s.body, ContentType: "audio/wav", Size: int64(s.body.Len())}, nil
}

func TestRecordingFileServerChunksAndReplays(t *testing.T) {
	payload := bytes.Repeat([]byte{0, 1, 2, 3}, recordingFileChunkSize/2+17)
	source := &recordingFileSource{body: &trackedFileBody{Reader: bytes.NewReader(payload)}}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = recordingFileClient{recording: source}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.StoredRecordingKey, ID: "call-1"}
	request := func(kind string, sequence uint64, session string) *proxy.Response {
		bus.response = nil
		s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: kind, Key: key,
			RecordingFile: &proxy.RecordingFileRequest{Session: session, Sequence: sequence}})
		if bus.response == nil {
			t.Fatal("missing proxy response")
		}
		return bus.response
	}
	opened := request("RecordingStoredFileOpen", 0, "")
	if opened.Err() != nil || opened.RecordingFile == nil || opened.RecordingFile.Session == "" || opened.RecordingFile.ContentType != "audio/wav" || opened.RecordingFile.Size != int64(len(payload)) || source.key != key {
		t.Fatalf("open = %+v", opened)
	}
	id := opened.RecordingFile.Session
	first := request("RecordingStoredFileRead", 0, id)
	if first.Err() != nil || len(first.RecordingFile.Data) != recordingFileChunkSize || first.RecordingFile.EOF {
		t.Fatalf("first chunk = %+v", first)
	}
	replayed := request("RecordingStoredFileRead", 0, id)
	if !reflect.DeepEqual(replayed.RecordingFile, first.RecordingFile) {
		t.Fatal("retry advanced the stream")
	}
	if err := request("RecordingStoredFileRead", 2, id).Err(); err == nil {
		t.Fatal("expected out-of-sequence error")
	}
	second := request("RecordingStoredFileRead", 1, id)
	third := request("RecordingStoredFileRead", 2, id)
	if second.Err() != nil || third.Err() != nil || !third.RecordingFile.EOF {
		t.Fatalf("second=%+v third=%+v", second, third)
	}
	var got []byte
	got = append(got, first.RecordingFile.Data...)
	got = append(got, second.RecordingFile.Data...)
	got = append(got, third.RecordingFile.Data...)
	if !bytes.Equal(got, payload) {
		t.Fatalf("transferred %d bytes, want %d", len(got), len(payload))
	}
	otherKey := *key
	otherKey.ID = "other-call"
	bus.response = nil
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "RecordingStoredFileClose", Key: &otherKey,
		RecordingFile: &proxy.RecordingFileRequest{Session: id}})
	if bus.response == nil || bus.response.Err() == nil || source.body.closed {
		t.Fatalf("mismatched close response=%+v closed=%v", bus.response, source.body.closed)
	}
	if response := request("RecordingStoredFileClose", 0, id); response.Err() != nil || !source.body.closed || len(s.recordingFiles) != 0 {
		t.Fatalf("close=%+v closed=%v sessions=%d", response, source.body.closed, len(s.recordingFiles))
	}
	if response := request("RecordingStoredFileClose", 0, id); response.Err() != nil {
		t.Fatalf("repeat close=%+v", response)
	}
}

func TestRecordingFileIdleSessionCleanup(t *testing.T) {
	source := &recordingFileSource{body: &trackedFileBody{Reader: bytes.NewReader([]byte("sample"))}}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = recordingFileClient{recording: source}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.StoredRecordingKey, ID: "call-1"}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "RecordingStoredFileOpen", Key: key})
	if bus.response == nil || bus.response.RecordingFile == nil {
		t.Fatalf("open response=%+v", bus.response)
	}
	id := bus.response.RecordingFile.Session
	session := s.recordingFiles[id]
	session.mu.Lock()
	session.lastUsed = time.Now().Add(-recordingFileIdleTimeout - time.Second)
	session.mu.Unlock()
	s.expireRecordingFile(id)
	if !source.body.closed || len(s.recordingFiles) != 0 {
		t.Fatalf("expired session: closed=%v count=%d", source.body.closed, len(s.recordingFiles))
	}
}
