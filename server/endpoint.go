// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"

	"github.com/two-barrels/ari-proxy/v6/proxy"
)

func (s *Server) endpointRefer(ctx context.Context, reply string, req *proxy.Request) {
	if req.EndpointRefer == nil {
		s.sendError(reply, errors.New("EndpointRefer is mandatory"))
		return
	}
	s.sendError(reply, s.ari.Endpoint().Refer(req.Key, *req.EndpointRefer))
}

func (s *Server) endpointReferToEndpoint(ctx context.Context, reply string, req *proxy.Request) {
	if req.EndpointRefer == nil {
		s.sendError(reply, errors.New("EndpointRefer is mandatory"))
		return
	}
	s.sendError(reply, s.ari.Endpoint().ReferToEndpoint(req.Key, *req.EndpointRefer))
}

func (s *Server) endpointData(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Endpoint().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Data: &proxy.EntityData{
			Endpoint: data,
		},
	})
}

func (s *Server) endpointGet(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Endpoint().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: data.Key,
	})
}

func (s *Server) endpointList(ctx context.Context, reply string, req *proxy.Request) {
	list, err := s.ari.Endpoint().List(nil)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Keys: list,
	})
}

func (s *Server) endpointListByTech(ctx context.Context, reply string, req *proxy.Request) {
	list, err := s.ari.Endpoint().ListByTech(req.EndpointListByTech.Tech, req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Keys: list,
	})
}
