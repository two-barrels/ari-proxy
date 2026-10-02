// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"

	"github.com/two-barrels/ari-proxy/v6/proxy"
)

func (s *Server) textMessageSend(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key == nil || req.TextMessageSend == nil {
		s.sendError(reply, errors.New("endpoint key and text message are mandatory"))
		return
	}
	msg := req.TextMessageSend
	s.sendError(reply, s.ari.TextMessage().SendWithKey(req.Key, msg.From, msg.Body, msg.Variables))
}

func (s *Server) textMessageSendByURI(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key == nil || req.TextMessageSend == nil {
		s.sendError(reply, errors.New("target node and text message are mandatory"))
		return
	}
	msg := req.TextMessageSend
	s.sendError(reply, s.ari.TextMessage().SendByURIWithKey(req.Key, msg.From, msg.To, msg.Body, msg.Variables))
}
