// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

func (s *Server) channelAnswer(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
	}

	s.sendError(reply, s.ari.Channel().Answer(req.Key))
}

func (s *Server) channelBusy(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
	}

	s.sendError(reply, s.ari.Channel().Busy(req.Key))
}

func (s *Server) channelCongestion(ctx context.Context, reply string, req *proxy.Request) {
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
	}

	s.sendError(reply, s.ari.Channel().Congestion(req.Key))
}

func (s *Server) channelCreate(ctx context.Context, reply string, req *proxy.Request) {
	create := req.ChannelCreate.ChannelCreateRequest

	if create.ChannelID == "" {
		create.ChannelID = rid.New(rid.Channel)
	}

	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", create.ChannelID)
		s.Dialog.Bind(req.Key.Dialog, "channel", create.OtherChannelID)
	}

	h, err := s.ari.Channel().Create(req.Key, create)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) channelData(ctx context.Context, reply string, req *proxy.Request) {
	d, err := s.ari.Channel().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Data: &proxy.EntityData{
			Channel: d,
		},
	})
}

func (s *Server) channelGet(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Channel().Data(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: data.Key,
	})
}

func (s *Server) channelContinue(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelContinue.Options != nil {
		s.sendError(reply, s.ari.Channel().ContinueWithOptions(req.Key, *req.ChannelContinue.Options))
		return
	}
	s.sendError(reply, s.ari.Channel().Continue(req.Key, req.ChannelContinue.Context, req.ChannelContinue.Extension, req.ChannelContinue.Priority))
}

func (s *Server) channelMove(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelMove == nil {
		s.sendError(reply, errors.New("ChannelMove is mandatory"))
		return
	}
	s.sendError(reply, s.ari.Channel().Move(req.Key, req.ChannelMove.App, req.ChannelMove.AppArgs))
}

func (s *Server) channelDial(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().Dial(req.Key, req.ChannelDial.Caller, req.ChannelDial.Timeout))
}

func (s *Server) channelHangup(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelHangup.ReasonCode != "" {
		s.sendError(reply, s.ari.Channel().HangupWithOptions(req.Key, ari.ChannelHangupOptions{
			Reason: req.ChannelHangup.Reason, ReasonCode: req.ChannelHangup.ReasonCode,
		}))
		return
	}
	s.sendError(reply, s.ari.Channel().Hangup(req.Key, req.ChannelHangup.Reason))
}

func (s *Server) channelHold(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().Hold(req.Key))
}

func (s *Server) channelList(ctx context.Context, reply string, req *proxy.Request) {
	list, err := s.ari.Channel().List(nil)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Keys: list,
	})
}

func (s *Server) channelMOH(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().MOH(req.Key, req.ChannelMOH.Music))
}

func (s *Server) channelMute(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().Mute(req.Key, req.ChannelMute.Direction))
}

func (s *Server) channelOriginate(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelOriginate == nil {
		s.sendError(reply, errors.New("OriginateRequest is mandatory"))
		return
	}
	orig := req.ChannelOriginate.OriginateRequest

	if orig.ChannelID == "" {
		orig.ChannelID = rid.New(rid.Channel)
	}

	if req.Key != nil && req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", orig.ChannelID)
		if orig.OtherChannelID != "" {
			s.Dialog.Bind(req.Key.Dialog, "channel", orig.OtherChannelID)
		}
		if orig.Originator != "" {
			s.Dialog.Bind(req.Key.Dialog, "channel", orig.Originator)
		}
		if req.Key.ID != orig.Originator {
			s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
		}
	}

	h, err := s.ari.Channel().Originate(req.Key, orig)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) channelOriginateWithID(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelOriginate == nil {
		s.sendError(reply, errors.New("OriginateRequest is mandatory"))
		return
	}
	orig := req.ChannelOriginate.OriginateRequest
	h, err := s.ari.Channel().OriginateWithID(req.Key, orig)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", h.ID())
	}
	s.publish(reply, &proxy.Response{Key: req.Key.New(ari.ChannelKey, h.ID())})
}

func (s *Server) channelStageOriginate(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelOriginate == nil {
		s.sendError(reply, errors.New("OriginateRequest is mandatory"))
		return
	}
	orig := req.ChannelOriginate.OriginateRequest

	if orig.ChannelID == "" {
		orig.ChannelID = rid.New(rid.Channel)
	}

	if req.Key != nil && req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", orig.ChannelID)
		if orig.OtherChannelID != "" {
			s.Dialog.Bind(req.Key.Dialog, "channel", orig.OtherChannelID)
		}
		if orig.Originator != "" {
			s.Dialog.Bind(req.Key.Dialog, "channel", orig.Originator)
		}
		if req.Key.ID != orig.Originator {
			s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
		}
	}

	h, err := s.ari.Channel().StageOriginate(req.Key, orig)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) channelPlay(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Channel().Data(req.Key)
	if err != nil || data == nil {
		s.sendError(reply, err)
		return
	}

	if req.ChannelPlay.PlaybackID == "" {
		req.ChannelPlay.PlaybackID = rid.New(rid.Playback)
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "playback", req.ChannelPlay.PlaybackID)
	}

	var ph *ari.PlaybackHandle
	if req.ChannelPlay.Options != nil {
		ph, err = s.ari.Channel().PlayWithOptions(req.Key, req.ChannelPlay.PlaybackID, *req.ChannelPlay.Options)
	} else {
		ph, err = s.ari.Channel().Play(req.Key, req.ChannelPlay.PlaybackID, req.ChannelPlay.URIs()...)
	}
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: ph.Key(),
	})
}

func (s *Server) channelPlayWithoutID(ctx context.Context, reply string, req *proxy.Request) {
	opts := ari.ChannelPlayOptions{Media: req.ChannelPlay.URIs()}
	if req.ChannelPlay.Options != nil {
		opts = *req.ChannelPlay.Options
	}
	var h *ari.PlaybackHandle
	var err error
	if req.ChannelPlay.PlaybackID != "" {
		h, err = s.ari.Channel().PlayOnCollection(req.Key, req.ChannelPlay.PlaybackID, opts)
	} else {
		h, err = s.ari.Channel().PlayWithoutID(req.Key, opts)
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

func (s *Server) channelStagePlay(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Channel().Data(req.Key)
	if err != nil || data == nil {
		s.Log.Debug("failed to get channel data", "channel", req.Key)
		s.sendError(reply, err)
		return
	}

	if req.ChannelPlay.PlaybackID == "" {
		req.ChannelPlay.PlaybackID = rid.New(rid.Playback)
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", data.ID)
		s.Dialog.Bind(req.Key.Dialog, "playback", req.ChannelPlay.PlaybackID)
	}

	s.publish(reply, &proxy.Response{
		Key: s.ari.Playback().Get(ari.NewKey(ari.PlaybackKey, req.ChannelPlay.PlaybackID)).Key(),
	})
}

func (s *Server) channelRecord(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelRecord.Name == "" {
		req.ChannelRecord.Name = rid.New(rid.Recording)
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
		s.Dialog.Bind(req.Key.Dialog, "recording", req.ChannelRecord.Name)
	}

	h, err := s.ari.Channel().Record(req.Key, req.ChannelRecord.Name, req.ChannelRecord.Options)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) channelStageRecord(ctx context.Context, reply string, req *proxy.Request) {
	data, err := s.ari.Channel().Data(req.Key)
	if err != nil || data == nil {
		s.sendError(reply, err)
		return
	}

	if req.ChannelRecord.Name == "" {
		req.ChannelRecord.Name = rid.New(rid.Recording)
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", data.ID)
		s.Dialog.Bind(req.Key.Dialog, "recording", req.ChannelRecord.Name)
	}

	s.publish(reply, &proxy.Response{
		Key: data.Key.New(ari.LiveRecordingKey, req.ChannelRecord.Name),
	})
}

func (s *Server) channelRing(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().Ring(req.Key))
}

func (s *Server) channelSendDTMF(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().SendDTMF(req.Key, req.ChannelSendDTMF.DTMF, req.ChannelSendDTMF.Options))
}

func (s *Server) channelSilence(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().Silence(req.Key))
}

func (s *Server) channelSnoop(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelSnoop.SnoopID == "" {
		req.ChannelSnoop.SnoopID = rid.New(rid.Snoop)
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.ChannelSnoop.SnoopID)
	}

	h, err := s.ari.Channel().Snoop(req.Key, req.ChannelSnoop.SnoopID, req.ChannelSnoop.Options)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) channelSnoopWithoutID(ctx context.Context, reply string, req *proxy.Request) {
	var h *ari.ChannelHandle
	var err error
	if req.ChannelSnoop.SnoopID != "" {
		h, err = s.ari.Channel().SnoopOnCollection(req.Key, req.ChannelSnoop.SnoopID, req.ChannelSnoop.Options)
	} else {
		h, err = s.ari.Channel().SnoopWithoutID(req.Key, req.ChannelSnoop.Options)
	}
	if err != nil {
		s.sendError(reply, err)
		return
	}
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", h.ID())
	}
	s.publish(reply, &proxy.Response{Key: req.Key.New(ari.ChannelKey, h.ID())})
}

func (s *Server) channelStageSnoop(ctx context.Context, reply string, req *proxy.Request) {
	// Snoop requires a reference channel to exist
	data, err := s.ari.Channel().Data(req.Key)
	if err != nil || data == nil {
		s.sendError(reply, err)
		return
	}

	if req.ChannelSnoop.SnoopID == "" {
		req.ChannelSnoop.SnoopID = rid.New(rid.Snoop)
	}

	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.ChannelSnoop.SnoopID)
	}

	s.publish(reply, &proxy.Response{
		Key: s.ari.Channel().Get(ari.NewKey(ari.ChannelKey, req.ChannelSnoop.SnoopID)).Key(),
	})
}

func (s *Server) channelExternalMedia(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelExternalMedia == nil {
		s.sendError(reply, errors.New("ExternalMediaOptions is required"))
		return
	}
	opts := req.ChannelExternalMedia.Options

	if opts.ChannelID == "" {
		opts.ChannelID = rid.New(rid.Channel)
	}

	if req.Key != nil && req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", opts.ChannelID)
	}

	h, err := s.ari.Channel().ExternalMedia(req.Key, opts)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) channelStageExternalMedia(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelExternalMedia == nil {
		s.sendError(reply, errors.New("ExternalMediaOptions is required"))
		return
	}
	opts := req.ChannelExternalMedia.Options

	if opts.ChannelID == "" {
		opts.ChannelID = rid.New(rid.Channel)
	}

	if req.Key != nil && req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", opts.ChannelID)
	}

	h, err := s.ari.Channel().StageExternalMedia(req.Key, opts)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.publish(reply, &proxy.Response{
		Key: h.Key(),
	})
}

func (s *Server) channelStopHold(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().StopHold(req.Key))
}

func (s *Server) channelStopMOH(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().StopMOH(req.Key))
}

func (s *Server) channelStopRing(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().StopRing(req.Key))
}

func (s *Server) channelStopSilence(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().StopSilence(req.Key))
}

func (s *Server) channelSubscribe(ctx context.Context, reply string, req *proxy.Request) {
	// bind dialog
	if req.Key.Dialog != "" {
		s.Dialog.Bind(req.Key.Dialog, "channel", req.Key.ID)
	}

	s.sendError(reply, nil)
}

func (s *Server) channelUnmute(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().Unmute(req.Key, req.ChannelMute.Direction))
}

func (s *Server) channelVariableGet(ctx context.Context, reply string, req *proxy.Request) {
	val, err := s.ari.Channel().GetVariable(req.Key, req.ChannelVariable.Name)
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

func (s *Server) channelVariableSet(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelVariable.ReportEvents != nil {
		s.sendError(reply, s.ari.Channel().SetVariableWithOptions(req.Key, req.ChannelVariable.Name, req.ChannelVariable.Value,
			&ari.ChannelVariableSetOptions{ReportEvents: req.ChannelVariable.ReportEvents}))
		return
	}
	s.sendError(reply, s.ari.Channel().SetVariable(req.Key, req.ChannelVariable.Name, req.ChannelVariable.Value))
}

func (s *Server) channelVariablesGet(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelVariables == nil {
		s.sendError(reply, errors.New("ChannelVariables is mandatory"))
		return
	}
	values, err := s.ari.Channel().GetVariables(req.Key, req.ChannelVariables.Names...)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	s.publish(reply, &proxy.Response{Data: &proxy.EntityData{Variables: values}})
}

func (s *Server) channelVariablesSet(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelVariables == nil {
		s.sendError(reply, errors.New("ChannelVariables is mandatory"))
		return
	}
	s.sendError(reply, s.ari.Channel().SetVariables(req.Key, req.ChannelVariables.Values))
}

func (s *Server) channelRedirect(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelRedirect == nil {
		s.sendError(reply, errors.New("ChannelRedirect is mandatory"))
		return
	}
	s.sendError(reply, s.ari.Channel().Redirect(req.Key, req.ChannelRedirect.Endpoint))
}

func (s *Server) channelProgress(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().Progress(req.Key))
}

func (s *Server) channelTransferProgress(ctx context.Context, reply string, req *proxy.Request) {
	if req.ChannelTransferProgress == nil {
		s.sendError(reply, errors.New("ChannelTransferProgress is mandatory"))
		return
	}
	s.sendError(reply, s.ari.Channel().TransferProgress(req.Key, req.ChannelTransferProgress.States))
}

func (s *Server) channelRTPStatistics(ctx context.Context, reply string, req *proxy.Request) {
	stats, err := s.ari.Channel().RTPStatistics(req.Key)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	s.publish(reply, &proxy.Response{Data: &proxy.EntityData{RTPStats: stats}})
}

func (s *Server) channelUserEvent(ctx context.Context, reply string, req *proxy.Request) {
	s.sendError(reply, s.ari.Channel().UserEvent(req.Key, &req.ChannelUserevent.UserEvent))
}
