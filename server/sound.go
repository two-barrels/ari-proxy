// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"github.com/two-barrels/ari-proxy/v6/proxy"
)

func (s *Server) soundData(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Sound().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Data: &proxy.EntityData{
			Sound: data,
		},
	})
}

func (s *Server) soundList(ctx context.Context, reply string, req *proxy.Request) {
	filters := req.SoundList.Filters

	if len(filters) == 0 {
		filters = nil // just send nil to upstream if empty. makes tests easier
	}

	list, err := s.ari.Sound().List(filters, req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Keys: list,
	})
}
