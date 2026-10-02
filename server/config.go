// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"github.com/two-barrels/ari-proxy/v6/proxy"
)

func (s *Server) asteriskConfigData(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Asterisk().Config().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Data: &proxy.EntityData{
			Config: data,
		},
	})
}

func (s *Server) asteriskConfigDelete(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Asterisk().Config().Delete(req.Key))
}

func (s *Server) asteriskConfigUpdate(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Asterisk().Config().Update(req.Key, req.AsteriskConfig.Tuples))
}
