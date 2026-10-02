// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

const maxRecordingFileChunk = 32 * 1024

// File opens a stored recording through bounded message-bus chunks.
func (s *storedRecording) File(ctx context.Context, key *ari.Key) (*ari.RecordingFile, error) {
	if ctx == nil {
		return nil, errors.New("context not supplied")
	}
	if key == nil || key.ID == "" || key.App == "" || key.Node == "" {
		return nil, errors.New("recording download requires a recording key, application, and target node")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	response, err := s.c.makeRequest("data", &proxy.Request{Kind: "RecordingStoredFileOpen", Key: key})
	if err != nil {
		return nil, err
	}
	if err := response.Err(); err != nil {
		return nil, err
	}
	if response.RecordingFile == nil || response.RecordingFile.Session == "" {
		return nil, ErrNil
	}
	reader := &recordingFileReader{client: s.c, ctx: ctx, key: key,
		session: response.RecordingFile.Session}
	return &ari.RecordingFile{Body: reader, ContentType: response.RecordingFile.ContentType,
		Size: response.RecordingFile.Size}, nil
}

type recordingFileReader struct {
	mu       sync.Mutex
	client   *Client
	ctx      context.Context
	key      *ari.Key
	session  string
	sequence uint64
	pending  []byte
	eof      bool
	closed   bool
}

func (r *recordingFileReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	if len(r.pending) == 0 {
		if r.eof {
			return 0, io.EOF
		}
		if err := r.ctx.Err(); err != nil {
			return 0, err
		}
		response, err := r.client.makeRequest("data", &proxy.Request{Kind: "RecordingStoredFileRead", Key: r.key,
			RecordingFile: &proxy.RecordingFileRequest{Session: r.session, Sequence: r.sequence}})
		if err != nil {
			return 0, err
		}
		if err := response.Err(); err != nil {
			return 0, err
		}
		chunk := response.RecordingFile
		if chunk == nil || chunk.Session != r.session || chunk.Sequence != r.sequence || len(chunk.Data) > maxRecordingFileChunk {
			return 0, errors.New("invalid recording file chunk")
		}
		r.sequence++
		r.pending = chunk.Data
		r.eof = chunk.EOF
		if len(r.pending) == 0 {
			if r.eof {
				return 0, io.EOF
			}
			return 0, io.ErrNoProgress
		}
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

func (r *recordingFileReader) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	r.pending = nil
	r.mu.Unlock()
	return r.client.commandRequest(&proxy.Request{Kind: "RecordingStoredFileClose", Key: r.key,
		RecordingFile: &proxy.RecordingFileRequest{Session: r.session}})
}
