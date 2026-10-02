// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

const (
	recordingFileChunkSize   = 32 * 1024
	maxRecordingFileSessions = 64
	recordingFileIdleTimeout = 2 * time.Minute
)

type recordingFileSession struct {
	mu           sync.Mutex
	key          ari.Key
	file         *ari.RecordingFile
	cancel       context.CancelFunc
	timer        *time.Timer
	lastUsed     time.Time
	nextSequence uint64
	last         *proxy.RecordingFileResponse
	closed       bool
}

func (s *Server) recordingStoredFileOpen(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key == nil || req.Key.ID == "" {
		s.sendError(reply, errors.New("stored recording key not supplied"))
		return
	}
	s.recordingFilesMu.Lock()
	if len(s.recordingFiles) >= maxRecordingFileSessions {
		s.recordingFilesMu.Unlock()
		s.sendError(reply, errors.New("too many open recording files"))
		return
	}
	s.recordingFilesMu.Unlock()
	fileCtx, cancel := context.WithCancel(ctx)
	file, err := s.ari.StoredRecording().File(fileCtx, req.Key)
	if err != nil {
		cancel()
		s.sendError(reply, err)
		return
	}
	id := rid.New("recording")
	session := &recordingFileSession{key: *req.Key, file: file, cancel: cancel, lastUsed: time.Now()}
	s.recordingFilesMu.Lock()
	if s.recordingFiles == nil {
		s.recordingFiles = make(map[string]*recordingFileSession)
	}
	if len(s.recordingFiles) >= maxRecordingFileSessions {
		s.recordingFilesMu.Unlock()
		file.Body.Close()
		cancel()
		s.sendError(reply, errors.New("too many open recording files"))
		return
	}
	s.recordingFiles[id] = session
	session.timer = time.AfterFunc(recordingFileIdleTimeout, func() { s.expireRecordingFile(id) })
	s.recordingFilesMu.Unlock()
	s.publish(reply, &proxy.Response{RecordingFile: &proxy.RecordingFileResponse{
		Session: id, ContentType: file.ContentType, Size: file.Size,
	}})
}

func (s *Server) recordingStoredFileRead(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key == nil || req.RecordingFile == nil || req.RecordingFile.Session == "" {
		s.sendError(reply, errors.New("recording file session not supplied"))
		return
	}
	s.recordingFilesMu.Lock()
	session := s.recordingFiles[req.RecordingFile.Session]
	s.recordingFilesMu.Unlock()
	if session == nil {
		s.sendError(reply, errors.New("recording file session not found"))
		return
	}
	session.mu.Lock()
	if session.closed || session.key.ID != req.Key.ID || session.key.App != req.Key.App || session.key.Node != req.Key.Node {
		session.mu.Unlock()
		s.sendError(reply, errors.New("recording file session does not match key"))
		return
	}
	session.lastUsed = time.Now()
	if session.last != nil && req.RecordingFile.Sequence == session.last.Sequence {
		response := session.last
		session.mu.Unlock()
		s.publish(reply, &proxy.Response{RecordingFile: response})
		return
	}
	if req.RecordingFile.Sequence != session.nextSequence {
		session.mu.Unlock()
		s.sendError(reply, errors.New("recording file chunk out of sequence"))
		return
	}
	chunk := make([]byte, recordingFileChunkSize)
	n, err := io.ReadFull(session.file.Body, chunk)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		session.mu.Unlock()
		s.closeRecordingFile(req.RecordingFile.Session)
		s.sendError(reply, err)
		return
	}
	response := &proxy.RecordingFileResponse{Session: req.RecordingFile.Session,
		Sequence: req.RecordingFile.Sequence, Data: chunk[:n], EOF: err != nil}
	session.last = response
	session.nextSequence++
	session.mu.Unlock()
	s.publish(reply, &proxy.Response{RecordingFile: response})
}

func (s *Server) recordingStoredFileClose(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key == nil || req.RecordingFile == nil || req.RecordingFile.Session == "" {
		s.sendError(reply, errors.New("recording file session not supplied"))
		return
	}
	s.recordingFilesMu.Lock()
	session := s.recordingFiles[req.RecordingFile.Session]
	s.recordingFilesMu.Unlock()
	if session != nil {
		session.mu.Lock()
		matches := session.key.ID == req.Key.ID && session.key.App == req.Key.App && session.key.Node == req.Key.Node
		session.mu.Unlock()
		if !matches {
			s.sendError(reply, errors.New("recording file session does not match key"))
			return
		}
	}
	s.closeRecordingFile(req.RecordingFile.Session)
	s.sendError(reply, nil)
}

func (s *Server) expireRecordingFile(id string) {
	s.recordingFilesMu.Lock()
	session := s.recordingFiles[id]
	s.recordingFilesMu.Unlock()
	if session == nil {
		return
	}
	session.mu.Lock()
	remaining := recordingFileIdleTimeout - time.Since(session.lastUsed)
	if remaining > 0 && !session.closed {
		session.timer.Reset(remaining)
		session.mu.Unlock()
		return
	}
	if !session.closed {
		session.closed = true
		session.file.Body.Close()
		session.cancel()
	}
	session.mu.Unlock()
	s.recordingFilesMu.Lock()
	if s.recordingFiles[id] == session {
		delete(s.recordingFiles, id)
	}
	s.recordingFilesMu.Unlock()
}

func (s *Server) closeRecordingFile(id string) {
	s.recordingFilesMu.Lock()
	session := s.recordingFiles[id]
	delete(s.recordingFiles, id)
	s.recordingFilesMu.Unlock()
	if session == nil {
		return
	}
	session.mu.Lock()
	if !session.closed {
		session.closed = true
		session.timer.Stop()
		session.file.Body.Close()
		session.cancel()
	}
	session.mu.Unlock()
}

func (s *Server) closeAllRecordingFiles() {
	s.recordingFilesMu.Lock()
	ids := make([]string, 0, len(s.recordingFiles))
	for id := range s.recordingFiles {
		ids = append(ids, id)
	}
	s.recordingFilesMu.Unlock()
	for _, id := range ids {
		s.closeRecordingFile(id)
	}
}
