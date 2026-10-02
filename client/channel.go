// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"time"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

type channel struct {
	c *Client
}

func (c *channel) Get(key *ari.Key) *ari.ChannelHandle {
	k, err := c.c.getRequest(&proxy.Request{
		Kind: "ChannelGet",
		Key:  key,
	})
	if err != nil {
		c.c.log.Warn("failed to make data request for channel", "error", err)
		return ari.NewChannelHandle(key, c, nil)
	}
	return ari.NewChannelHandle(k, c, nil)
}

func (c *channel) List(filter *ari.Key) ([]*ari.Key, error) {
	return c.c.listRequest(&proxy.Request{
		Kind: "ChannelList",
		Key:  filter,
	})
}

func (c *channel) Originate(referenceKey *ari.Key, o ari.OriginateRequest) (*ari.ChannelHandle, error) {
	k, err := c.c.createRequest(&proxy.Request{
		Kind: "ChannelOriginate",
		Key:  referenceKey,
		ChannelOriginate: &proxy.ChannelOriginate{
			OriginateRequest: o,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k, c, nil), nil
}

func (c *channel) OriginateWithID(referenceKey *ari.Key, o ari.OriginateRequest) (*ari.ChannelHandle, error) {
	if err := requireRouteTarget(referenceKey); err != nil {
		return nil, err
	}
	if o.ChannelID == "" {
		return nil, errors.New("channel ID required for path-ID originate")
	}
	k, err := c.c.createRequest(&proxy.Request{Kind: "ChannelOriginateWithID", Key: referenceKey,
		ChannelOriginate: &proxy.ChannelOriginate{OriginateRequest: o}})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k, c, nil), nil
}

func (c *channel) StageOriginate(referenceKey *ari.Key, o ari.OriginateRequest) (*ari.ChannelHandle, error) {
	if o.ChannelID == "" {
		o.ChannelID = rid.New(rid.Channel)
	}

	// We go ahead an call the createRequest on the server so that we lock in an
	// Asterisk box at the time of staging even though this staging call will
	// never actually be used.
	k, err := c.c.createRequest(&proxy.Request{
		Kind: "ChannelStageOriginate",
		Key:  referenceKey,
		ChannelOriginate: &proxy.ChannelOriginate{
			OriginateRequest: o,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k.New(ari.ChannelKey, o.ChannelID), c, func(h *ari.ChannelHandle) error {
		_, err := c.Originate(referenceKey, o)
		return err
	}), nil
}

func (c *channel) Create(key *ari.Key, o ari.ChannelCreateRequest) (*ari.ChannelHandle, error) {
	k, err := c.c.createRequest(&proxy.Request{
		Kind: "ChannelCreate",
		Key:  key,
		ChannelCreate: &proxy.ChannelCreate{
			ChannelCreateRequest: o,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k.New(ari.ChannelKey, o.ChannelID), c, nil), nil
}

func (c *channel) Data(key *ari.Key) (*ari.ChannelData, error) {
	data, err := c.c.dataRequest(&proxy.Request{
		Kind: "ChannelData",
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	return data.Channel, nil
}

func (c *channel) Continue(key *ari.Key, context string, extension string, priority int) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelContinue",
		Key:  key,
		ChannelContinue: &proxy.ChannelContinue{
			Context:   context,
			Extension: extension,
			Priority:  priority,
		},
	})
}

func (c *channel) ContinueWithOptions(key *ari.Key, opts ari.ChannelContinueOptions) error {
	return c.c.commandRequest(&proxy.Request{Kind: "ChannelContinue", Key: key,
		ChannelContinue: &proxy.ChannelContinue{Options: &opts}})
}

func (c *channel) Move(key *ari.Key, app, appArgs string) error {
	return c.c.commandRequest(&proxy.Request{
		Kind:        "ChannelMove",
		Key:         key,
		ChannelMove: &proxy.ChannelMove{App: app, AppArgs: appArgs},
	})
}

func (c *channel) Busy(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelBusy",
		Key:  key,
	})
}

func (c *channel) Congestion(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelCongestion",
		Key:  key,
	})
}

func (c *channel) Hangup(key *ari.Key, reason string) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelHangup",
		Key:  key,
		ChannelHangup: &proxy.ChannelHangup{
			Reason: reason,
		},
	})
}

func (c *channel) HangupWithOptions(key *ari.Key, opts ari.ChannelHangupOptions) error {
	return c.c.commandRequest(&proxy.Request{Kind: "ChannelHangup", Key: key,
		ChannelHangup: &proxy.ChannelHangup{Reason: opts.Reason, ReasonCode: opts.ReasonCode}})
}

func (c *channel) Answer(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelAnswer",
		Key:  key,
	})
}

func (c *channel) Ring(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelRing",
		Key:  key,
	})
}

func (c *channel) StopRing(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelStopRing",
		Key:  key,
	})
}

func (c *channel) SendDTMF(key *ari.Key, dtmf string, opts *ari.DTMFOptions) error {
	if opts == nil {
		opts = &ari.DTMFOptions{}
	}
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelSendDTMF",
		Key:  key,
		ChannelSendDTMF: &proxy.ChannelSendDTMF{
			DTMF:    dtmf,
			Options: opts,
		},
	})
}

func (c *channel) Hold(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelHold",
		Key:  key,
	})
}

func (c *channel) StopHold(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelStopHold",
		Key:  key,
	})
}

func (c *channel) Mute(key *ari.Key, dir ari.Direction) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelMute",
		Key:  key,
		ChannelMute: &proxy.ChannelMute{
			Direction: dir,
		},
	})
}

func (c *channel) Unmute(key *ari.Key, dir ari.Direction) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelUnmute",
		Key:  key,
		ChannelMute: &proxy.ChannelMute{
			Direction: dir,
		},
	})
}

func (c *channel) MOH(key *ari.Key, moh string) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelMOH",
		Key:  key,
		ChannelMOH: &proxy.ChannelMOH{
			Music: moh,
		},
	})
}

func (c *channel) StopMOH(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelStopMOH",
		Key:  key,
	})
}

func (c *channel) Silence(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelSilence",
		Key:  key,
	})
}

func (c *channel) StopSilence(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelStopSilence",
		Key:  key,
	})
}

func (c *channel) Snoop(key *ari.Key, snoopID string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	k, err := c.c.createRequest(&proxy.Request{
		Kind: "ChannelSnoop",
		Key:  key,
		ChannelSnoop: &proxy.ChannelSnoop{
			SnoopID: snoopID,
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k.New(ari.ChannelKey, snoopID), c, nil), nil
}

func (c *channel) SnoopWithoutID(key *ari.Key, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	if err := requireRouteTarget(key); err != nil {
		return nil, err
	}
	k, err := c.c.createRequest(&proxy.Request{Kind: "ChannelSnoopWithoutID", Key: key,
		ChannelSnoop: &proxy.ChannelSnoop{Options: opts}})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k, c, nil), nil
}

func (c *channel) SnoopOnCollection(key *ari.Key, snoopID string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	if err := requireRouteTarget(key); err != nil {
		return nil, err
	}
	k, err := c.c.createRequest(&proxy.Request{Kind: "ChannelSnoopOnCollection", Key: key,
		ChannelSnoop: &proxy.ChannelSnoop{SnoopID: snoopID, Options: opts}})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k, c, nil), nil
}

func (c *channel) StageSnoop(key *ari.Key, snoopID string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	// this getRequest is done merely to locate the Asterisk box on which the
	// snoop will be initiated.  It will never actually be Exec'd
	k, err := c.c.getRequest(&proxy.Request{
		Kind: "ChannelStageSnoop",
		Key:  key,
		ChannelSnoop: &proxy.ChannelSnoop{
			SnoopID: snoopID,
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k, c, func(h *ari.ChannelHandle) error {
		_, err := c.Snoop(k.New(ari.ChannelKey, key.ID), snoopID, opts)
		return err
	}), nil
}

func (c *channel) ExternalMedia(referenceKey *ari.Key, opts ari.ExternalMediaOptions) (*ari.ChannelHandle, error) {
	if opts.ChannelID == "" {
		opts.ChannelID = rid.New(rid.Channel)
	}
	k, err := c.c.createRequest(&proxy.Request{
		Kind: "ChannelExternalMedia",
		Key:  referenceKey,
		ChannelExternalMedia: &proxy.ChannelExternalMedia{
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k, c, nil), nil
}

func (c *channel) StageExternalMedia(referenceKey *ari.Key, opts ari.ExternalMediaOptions) (*ari.ChannelHandle, error) {
	if opts.ChannelID == "" {
		opts.ChannelID = rid.New(rid.Channel)
	}

	// We go ahead an call the createRequest on the server so that we lock in an
	// Asterisk box at the time of staging even though this staging call will
	// never actually be used.
	k, err := c.c.createRequest(&proxy.Request{
		Kind: "ChannelStageExternalMedia",
		Key:  referenceKey,
		ChannelExternalMedia: &proxy.ChannelExternalMedia{
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewChannelHandle(k, c, func(h *ari.ChannelHandle) error {
		_, err := c.ExternalMedia(k, opts)
		return err
	}), nil
}

func (c *channel) Dial(key *ari.Key, caller string, timeout time.Duration) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelDial",
		Key:  key,
		ChannelDial: &proxy.ChannelDial{
			Caller:  caller,
			Timeout: timeout,
		},
	})
}

func (c *channel) Play(key *ari.Key, playbackID string, mediaURI ...string) (*ari.PlaybackHandle, error) {
	return c.PlayWithOptions(key, playbackID, ari.ChannelPlayOptions{Media: mediaURI})
}

func (c *channel) PlayWithOptions(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	if playbackID == "" {
		playbackID = rid.New(rid.Playback)
	}

	k, err := c.c.createRequest(&proxy.Request{
		Kind:        "ChannelPlay",
		Key:         key,
		ChannelPlay: channelPlayPayload(playbackID, opts),
	})
	if err != nil {
		return nil, err
	}
	return ari.NewPlaybackHandle(k.New(ari.PlaybackKey, playbackID), c.c.Playback(), nil), nil
}

func (c *channel) PlayWithoutID(key *ari.Key, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	if err := requireRouteTarget(key); err != nil {
		return nil, err
	}
	k, err := c.c.createRequest(&proxy.Request{Kind: "ChannelPlayWithoutID", Key: key,
		ChannelPlay: channelPlayPayload("", opts)})
	if err != nil {
		return nil, err
	}
	return ari.NewPlaybackHandle(k, c.c.Playback(), nil), nil
}

func (c *channel) PlayOnCollection(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	if err := requireRouteTarget(key); err != nil {
		return nil, err
	}
	k, err := c.c.createRequest(&proxy.Request{Kind: "ChannelPlayOnCollection", Key: key,
		ChannelPlay: channelPlayPayload(playbackID, opts)})
	if err != nil {
		return nil, err
	}
	return ari.NewPlaybackHandle(k, c.c.Playback(), nil), nil
}

func (c *channel) StagePlay(key *ari.Key, playbackID string, mediaURI ...string) (*ari.PlaybackHandle, error) {
	return c.StagePlayWithOptions(key, playbackID, ari.ChannelPlayOptions{Media: mediaURI})
}

func (c *channel) StagePlayWithOptions(key *ari.Key, playbackID string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	if playbackID == "" {
		playbackID = rid.New(rid.Playback)
	}

	k, err := c.c.getRequest(&proxy.Request{
		Kind:        "ChannelStagePlay",
		Key:         key,
		ChannelPlay: channelPlayPayload(playbackID, opts),
	})
	if err != nil {
		return nil, err
	}
	return ari.NewPlaybackHandle(k.New(ari.PlaybackKey, playbackID), c.c.Playback(), func(h *ari.PlaybackHandle) error {
		_, err := c.PlayWithOptions(k.New(ari.ChannelKey, key.ID), playbackID, opts)
		return err
	}), nil
}

func channelPlayPayload(id string, opts ari.ChannelPlayOptions) *proxy.ChannelPlay {
	media := opts.Media
	p := &proxy.ChannelPlay{PlaybackID: id, Options: &opts}
	if len(media) > 0 {
		p.MediaURI = media[0]
	}
	if len(media) > 1 {
		p.MediaURIs = media
	}
	return p
}

func (c *channel) Record(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	rb, err := c.c.createRequest(&proxy.Request{
		Kind: "ChannelRecord",
		Key:  key,
		ChannelRecord: &proxy.ChannelRecord{
			Name:    name,
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewLiveRecordingHandle(rb.New(ari.LiveRecordingKey, name), c.c.LiveRecording(), nil), nil
}

func (c *channel) StageRecord(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	k, err := c.c.getRequest(&proxy.Request{
		Kind: "ChannelStageRecord",
		Key:  key,
		ChannelRecord: &proxy.ChannelRecord{
			Name:    name,
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewLiveRecordingHandle(k.New(ari.LiveRecordingKey, k.ID), c.c.LiveRecording(), func(h *ari.LiveRecordingHandle) error {
		_, err := c.Record(k.New(ari.ChannelKey, key.ID), k.ID, opts)
		return err
	}), nil
}

func (c *channel) Subscribe(key *ari.Key, n ...string) ari.Subscription {
	err := c.c.commandRequest(&proxy.Request{
		Kind: "ChannelSubscribe",
		Key:  key,
	})
	if err != nil {
		c.c.log.Warn("failed to call channel subscribe")
		if key.Dialog != "" {
			c.c.log.Error("dialog present; failing")
			return nil
		}
	}
	return c.c.Bus().Subscribe(key, n...)
}

func (c *channel) GetVariable(key *ari.Key, name string) (string, error) {
	data, err := c.c.dataRequest(&proxy.Request{
		Kind: "ChannelVariableGet",
		Key:  key,
		ChannelVariable: &proxy.ChannelVariable{
			Name: name,
		},
	})
	if err != nil {
		return "", err
	}
	return data.Variable, nil
}

func (c *channel) SetVariable(key *ari.Key, name, value string) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelVariableSet",
		Key:  key,
		ChannelVariable: &proxy.ChannelVariable{
			Name:  name,
			Value: value,
		},
	})
}

func (c *channel) SetVariableWithOptions(key *ari.Key, name, value string, opts *ari.ChannelVariableSetOptions) error {
	var reportEvents *bool
	if opts != nil {
		reportEvents = opts.ReportEvents
	}
	return c.c.commandRequest(&proxy.Request{Kind: "ChannelVariableSet", Key: key,
		ChannelVariable: &proxy.ChannelVariable{Name: name, Value: value, ReportEvents: reportEvents}})
}

func (c *channel) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	data, err := c.c.dataRequest(&proxy.Request{Kind: "ChannelVariablesGet", Key: key,
		ChannelVariables: &proxy.ChannelVariables{Names: names}})
	if err != nil {
		return nil, err
	}
	return data.Variables, nil
}

func (c *channel) SetVariables(key *ari.Key, values map[string]ari.VariableAssignment) error {
	return c.c.commandRequest(&proxy.Request{Kind: "ChannelVariablesSet", Key: key,
		ChannelVariables: &proxy.ChannelVariables{Values: values}})
}

func (c *channel) Redirect(key *ari.Key, endpoint string) error {
	return c.c.commandRequest(&proxy.Request{Kind: "ChannelRedirect", Key: key,
		ChannelRedirect: &proxy.ChannelRedirect{Endpoint: endpoint}})
}

func (c *channel) Progress(key *ari.Key) error {
	return c.c.commandRequest(&proxy.Request{Kind: "ChannelProgress", Key: key})
}

func (c *channel) TransferProgress(key *ari.Key, state string) error {
	return c.c.commandRequest(&proxy.Request{Kind: "ChannelTransferProgress", Key: key,
		ChannelTransferProgress: &proxy.ChannelTransferProgress{States: state}})
}

func (c *channel) RTPStatistics(key *ari.Key) (*ari.RTPStats, error) {
	data, err := c.c.dataRequest(&proxy.Request{Kind: "ChannelRTPStatistics", Key: key})
	if err != nil {
		return nil, err
	}
	return data.RTPStats, nil
}

func (c *channel) UserEvent(key *ari.Key, ue *ari.ChannelUserevent) error {
	return c.c.commandRequest(&proxy.Request{
		Kind: "ChannelUserEvent",
		Key:  key,
		ChannelUserevent: &proxy.ChannelUserevent{
			UserEvent: *ue,
		},
	})
}
