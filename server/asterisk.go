package server

import (
	"context"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

func (s *Server) asteriskInfo(ctx context.Context, reply string, req *proxy.Request) {
	var data *ari.AsteriskInfo
	var err error
	if req.AsteriskInfoOptions != nil {
		data, err = s.ari.Asterisk().InfoWithOptions(req.Key, *req.AsteriskInfoOptions)
	} else {
		data, err = s.ari.Asterisk().Info(req.Key)
	}
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Data: &proxy.EntityData{
			Asterisk: data,
		},
	})
}

func (s *Server) asteriskPing(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Asterisk().Ping(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	s.publish(reply, &proxy.Response{Data: &proxy.EntityData{AsteriskPing: data}})
}

func (s *Server) asteriskVariableGet(ctx context.Context, reply string, req *proxy.Request) {
	val, err := s.ari.Asterisk().Variables().Get(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Data: &proxy.EntityData{
			Variable: val,
		},
	})
}

func (s *Server) asteriskVariableSet(ctx context.Context, reply string, req *proxy.Request) {
	err := s.ari.Asterisk().Variables().Set(req.Key, req.AsteriskVariableSet.Value)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.sendError(reply, nil)
}
