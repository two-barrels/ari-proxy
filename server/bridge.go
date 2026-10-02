package server

import (
	"context"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

func (s *Server) bridgeVariableGet(ctx context.Context, reply string, req *proxy.Request) {
	value, err := s.ari.Bridge().GetVariable(req.Key, req.BridgeVariable.Name)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	s.publish(reply, &proxy.Response{Data: &proxy.EntityData{Variable: value}})
}

func (s *Server) bridgeVariableSet(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Bridge().SetVariable(req.Key, req.BridgeVariable.Name,
		req.BridgeVariable.Value, req.BridgeVariable.ReportEvents))
}

func (s *Server) bridgeVariablesGet(ctx context.Context, reply string, req *proxy.Request) {
	values, err := s.ari.Bridge().GetVariables(req.Key, req.BridgeVariables.Names...)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	s.publish(reply, &proxy.Response{Data: &proxy.EntityData{Variables: values}})
}

func (s *Server) bridgeVariablesSet(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Bridge().SetVariables(req.Key, req.BridgeVariables.Values))
}

func (s *Server) bridgeAddChannel(ctx context.Context, reply string, req *proxy.Request) {
	channel := req.BridgeAddChannel.Channel

	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "channel", channel)
	}

	err := s.ari.Bridge().AddChannelWithOptions(req.Key, channel, &ari.BridgeAddChannelOptions{
		Role:                        req.BridgeAddChannel.Role,
		AbsorbDTMF:                  req.BridgeAddChannel.AbsorbDTMF,
		Mute:                        req.BridgeAddChannel.Mute,
		InhibitConnectedLineUpdates: req.BridgeAddChannel.InhibitConnectedLineUpdates,
	})
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.sendError(reply, nil)
}

func (s *Server) bridgeCreate(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	h, err := s.ari.Bridge().CreateWithOptions(req.Key, ari.BridgeCreateOptions{
		Type: req.BridgeCreate.Type, Name: req.BridgeCreate.Name, Variables: req.BridgeCreate.Variables,
	})
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) bridgeCreateWithoutID(ctx context.Context, reply string, req *proxy.Request) {
	opts := ari.BridgeCreateOptions{Type: req.BridgeCreate.Type, Name: req.BridgeCreate.Name, Variables: req.BridgeCreate.Variables}
	var h *ari.BridgeHandle
	var err error
	if req.BridgeCreate.BridgeID != "" {
		h, err = s.ari.Bridge().CreateOnCollection(req.Key, req.BridgeCreate.BridgeID, opts)
	} else {
		h, err = s.ari.Bridge().CreateWithoutID(req.Key, opts)
	}
	if err != nil {
		s.sendError(reply, err)
		return
	}
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", h.ID())
	}
	s.publish(reply, &proxy.Response{Key: req.Key.New(ari.BridgeKey, h.ID())})
}

func (s *Server) bridgeStageCreate(ctx context.Context, reply string, req *proxy.Request) {
	bh := s.ari.Bridge().Get(req.Key)

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	s.publish(reply, &proxy.Response{
		Key: bh.Key(),
	})
}

func (s *Server) bridgeData(ctx context.Context, reply string, req *proxy.Request) {
	bd, err := s.ari.Bridge().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Data: &proxy.EntityData{
			Bridge: bd,
		},
	})
}

func (s *Server) bridgeDelete(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	err := s.ari.Bridge().Delete(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.sendError(reply, nil)
}

func (s *Server) bridgeGet(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Bridge().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	s.publish(reply, &proxy.Response{
		Key: data.Key,
	})
}

func (s *Server) bridgeList(ctx context.Context, reply string, req *proxy.Request) {
	list, err := s.ari.Bridge().List(nil)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Keys: list,
	})
}

func (s *Server) bridgeMOH(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	s.sendError(
		reply,
		s.ari.Bridge().MOH(req.Key, req.BridgeMOH.Class),
	)
}

func (s *Server) bridgeStopMOH(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	s.sendError(
		reply,
		s.ari.Bridge().StopMOH(req.Key),
	)
}

func (s *Server) bridgePlay(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "playback", req.BridgePlay.PlaybackID)
	}

	var ph *ari.PlaybackHandle
	var err error
	if req.BridgePlay.Options != nil {
		ph, err = s.ari.Bridge().PlayWithOptions(req.Key, req.BridgePlay.PlaybackID, *req.BridgePlay.Options)
	} else {
		ph, err = s.ari.Bridge().Play(req.Key, req.BridgePlay.PlaybackID, req.BridgePlay.URIs()...)
	}
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: ph.Key(),
	})
}

func (s *Server) bridgePlayWithoutID(ctx context.Context, reply string, req *proxy.Request) {
	opts := ari.BridgePlayOptions{Media: req.BridgePlay.URIs()}
	if req.BridgePlay.Options != nil {
		opts = *req.BridgePlay.Options
	}
	var h *ari.PlaybackHandle
	var err error
	if req.BridgePlay.PlaybackID != "" {
		h, err = s.ari.Bridge().PlayOnCollection(req.Key, req.BridgePlay.PlaybackID, opts)
	} else {
		h, err = s.ari.Bridge().PlayWithoutID(req.Key, opts)
	}
	if err != nil {
		s.sendError(reply, err)
		return
	}
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "playback", h.ID())
	}
	s.publish(reply, &proxy.Response{Key: req.Key.New(ari.PlaybackKey, h.ID())})
}

func (s *Server) bridgeStagePlay(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Bridge().Data(req.Key)
	if err != nil || data == nil {
		s.sendError(reply, err)
		return
	}

	if req.BridgePlay.PlaybackID == "" {
		req.BridgePlay.PlaybackID = rid.New(rid.Playback)
	}

	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "playback", req.BridgePlay.PlaybackID)
	}

	s.publish(reply, &proxy.Response{
		Key: s.ari.Playback().Get(ari.NewKey(ari.PlaybackKey, req.BridgePlay.PlaybackID)).Key(),
	})
}

func (s *Server) bridgeRecord(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Bridge().Data(req.Key)
	if err != nil || data == nil {
		s.sendError(reply, err)
		return
	}

	if req.BridgeRecord.Name == "" {
		req.BridgeRecord.Name = rid.New(rid.Recording)
	}

	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "recording", req.BridgeRecord.Name)
	}

	h, err := s.ari.Bridge().Record(req.Key, req.BridgeRecord.Name, req.BridgeRecord.Options)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) bridgeStageRecord(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Bridge().Data(req.Key)
	if err != nil || data == nil {
		s.sendError(reply, err)
		return
	}

	if req.BridgeRecord.Name == "" {
		req.BridgeRecord.Name = rid.New(rid.Recording)
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", data.ID)
		s.Dialog.Bind(req.Key.Dialog, "recording", req.BridgeRecord.Name)
	}

	s.publish(reply, &proxy.Response{
		Key: data.Key.New(ari.LiveRecordingKey, req.BridgeRecord.Name),
	})
}

func (s *Server) bridgeRemoveChannel(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "channel", req.BridgeRemoveChannel.Channel)
	}

	err := s.ari.Bridge().RemoveChannel(req.Key, req.BridgeRemoveChannel.Channel)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.sendError(reply, nil)
}

func (s *Server) bridgeSubscribe(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	s.sendError(reply, nil)
}

func (s *Server) bridgeUnsubscribe(ctx context.Context, reply string, req *proxy.Request) {
	// no-op for now; may want to eventually optimize away the dialog subscription
	s.sendError(reply, nil)
}

func (s *Server) bridgeVideoSource(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "channel", req.BridgeVideoSource.Channel)
	}

	err := s.ari.Bridge().VideoSource(req.Key, req.BridgeVideoSource.Channel)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.sendError(reply, nil)
}

func (s *Server) bridgeVideoSourceDelete(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "bridge", req.Key.ID)
	}

	err := s.ari.Bridge().VideoSourceDelete(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.sendError(reply, nil)
}
