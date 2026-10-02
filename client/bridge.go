// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/rid"
)

func requireRouteTarget(key *ari.Key) error {
	if key == nil || key.App == "" || key.Node == "" {
		return errors.New("explicit route requires an application and target node")
	}
	return nil
}

type bridge struct {
	c *Client
}

func (b *bridge) GetVariable(key *ari.Key, name string) (string, error) {
	data, err := b.c.dataRequest(&proxy.Request{Kind: "BridgeVariableGet", Key: key,
		BridgeVariable: &proxy.BridgeVariable{Name: name}})
	if err != nil {
		return "", err
	}
	return data.Variable, nil
}

func (b *bridge) SetVariable(key *ari.Key, name, value string, reportEvents *bool) error {
	return b.c.commandRequest(&proxy.Request{Kind: "BridgeVariableSet", Key: key,
		BridgeVariable: &proxy.BridgeVariable{Name: name, Value: value, ReportEvents: reportEvents}})
}

func (b *bridge) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	data, err := b.c.dataRequest(&proxy.Request{Kind: "BridgeVariablesGet", Key: key,
		BridgeVariables: &proxy.BridgeVariables{Names: names}})
	if err != nil {
		return nil, err
	}
	return data.Variables, nil
}

func (b *bridge) SetVariables(key *ari.Key, values map[string]ari.BridgeVariableAssignment) error {
	return b.c.commandRequest(&proxy.Request{Kind: "BridgeVariablesSet", Key: key,
		BridgeVariables: &proxy.BridgeVariables{Values: values}})
}

func (b *bridge) Create(key *ari.Key, btype, name string) (*ari.BridgeHandle, error) {
	return b.CreateWithOptions(key, ari.BridgeCreateOptions{Type: btype, Name: name})
}

func (b *bridge) CreateWithOptions(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	k, err := b.c.createRequest(&proxy.Request{
		Kind: "BridgeCreate",
		Key:  key,
		BridgeCreate: &proxy.BridgeCreate{
			Type:      opts.Type,
			Name:      opts.Name,
			Variables: opts.Variables,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewBridgeHandle(k, b, nil), nil
}

func (b *bridge) CreateWithoutID(reference *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	if err := requireRouteTarget(reference); err != nil {
		return nil, err
	}
	k, err := b.c.createRequest(&proxy.Request{
		Kind: "BridgeCreateWithoutID", Key: reference,
		BridgeCreate: &proxy.BridgeCreate{Type: opts.Type, Name: opts.Name, Variables: opts.Variables},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewBridgeHandle(k, b, nil), nil
}

func (b *bridge) CreateOnCollection(reference *ari.Key, bridgeID string, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	if err := requireRouteTarget(reference); err != nil {
		return nil, err
	}
	k, err := b.c.createRequest(&proxy.Request{Kind: "BridgeCreateOnCollection", Key: reference,
		BridgeCreate: &proxy.BridgeCreate{BridgeID: bridgeID, Type: opts.Type, Name: opts.Name, Variables: opts.Variables}})
	if err != nil {
		return nil, err
	}
	return ari.NewBridgeHandle(k, b, nil), nil
}

func (b *bridge) StageCreate(key *ari.Key, btype, name string) (*ari.BridgeHandle, error) {
	return b.StageCreateWithOptions(key, ari.BridgeCreateOptions{Type: btype, Name: name})
}

func (b *bridge) StageCreateWithOptions(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	k, err := b.c.createRequest(&proxy.Request{
		Kind: "BridgeStageCreate",
		Key:  key,
		BridgeCreate: &proxy.BridgeCreate{
			Type:      opts.Type,
			Name:      opts.Name,
			Variables: opts.Variables,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewBridgeHandle(k, b, func(h *ari.BridgeHandle) error {
		_, err := b.CreateWithOptions(k, opts)
		return err
	}), nil
}

func (b *bridge) Get(key *ari.Key) *ari.BridgeHandle {
	k, err := b.c.getRequest(&proxy.Request{
		Kind: "BridgeGet",
		Key:  key,
	})
	if err != nil {
		b.c.log.Warn("failed to get bridge for handle", "error", err)
		return ari.NewBridgeHandle(key, b, nil)
	}
	return ari.NewBridgeHandle(k, b, nil)
}

func (b *bridge) List(filter *ari.Key) ([]*ari.Key, error) {
	return b.c.listRequest(&proxy.Request{
		Kind: "BridgeList",
		Key:  filter,
	})
}

func (b *bridge) Data(key *ari.Key) (*ari.BridgeData, error) {
	resp, err := b.c.dataRequest(&proxy.Request{
		Kind: "BridgeData",
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	return resp.Bridge, nil
}

func (b *bridge) AddChannel(key *ari.Key, channelID string) error {
	return b.AddChannelWithOptions(key, channelID, nil)
}

func (b *bridge) AddChannelWithOptions(key *ari.Key, channelID string, options *ari.BridgeAddChannelOptions) error {
	if options == nil {
		options = new(ari.BridgeAddChannelOptions)
	}

	return b.c.commandRequest(&proxy.Request{
		Kind: "BridgeAddChannel",
		Key:  key,
		BridgeAddChannel: &proxy.BridgeAddChannel{
			Channel:                     channelID,
			AbsorbDTMF:                  options.AbsorbDTMF,
			Mute:                        options.Mute,
			Role:                        options.Role,
			InhibitConnectedLineUpdates: options.InhibitConnectedLineUpdates,
		},
	})
}

func (b *bridge) RemoveChannel(key *ari.Key, channelID string) error {
	return b.c.commandRequest(&proxy.Request{
		Kind: "BridgeRemoveChannel",
		Key:  key,
		BridgeRemoveChannel: &proxy.BridgeRemoveChannel{
			Channel: channelID,
		},
	})
}

func (b *bridge) Delete(key *ari.Key) error {
	return b.c.commandRequest(&proxy.Request{
		Kind: "BridgeDelete",
		Key:  key,
	})
}

func (b *bridge) MOH(key *ari.Key, class string) error {
	return b.c.commandRequest(&proxy.Request{
		Kind: "BridgeMOH",
		Key:  key,
		BridgeMOH: &proxy.BridgeMOH{
			Class: class,
		},
	})
}

func (b *bridge) StopMOH(key *ari.Key) error {
	return b.c.commandRequest(&proxy.Request{
		Kind: "BridgeStopMOH",
		Key:  key,
	})
}

func (b *bridge) Play(key *ari.Key, id string, uri ...string) (*ari.PlaybackHandle, error) {
	return b.PlayWithOptions(key, id, ari.BridgePlayOptions{Media: uri})
}

func (b *bridge) PlayWithOptions(key *ari.Key, id string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	if id == "" {
		id = rid.New(rid.Playback)
	}
	k, err := b.c.createRequest(&proxy.Request{
		Kind:       "BridgePlay",
		Key:        key,
		BridgePlay: bridgePlayPayload(id, opts),
	})
	if err != nil {
		return nil, err
	}
	return ari.NewPlaybackHandle(k.New(ari.PlaybackKey, id), b.c.Playback(), nil), nil
}

func (b *bridge) PlayWithoutID(key *ari.Key, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	if err := requireRouteTarget(key); err != nil {
		return nil, err
	}
	k, err := b.c.createRequest(&proxy.Request{Kind: "BridgePlayWithoutID", Key: key, BridgePlay: bridgePlayPayload("", opts)})
	if err != nil {
		return nil, err
	}
	return ari.NewPlaybackHandle(k, b.c.Playback(), nil), nil
}

func (b *bridge) PlayOnCollection(key *ari.Key, playbackID string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	if err := requireRouteTarget(key); err != nil {
		return nil, err
	}
	k, err := b.c.createRequest(&proxy.Request{Kind: "BridgePlayOnCollection", Key: key, BridgePlay: bridgePlayPayload(playbackID, opts)})
	if err != nil {
		return nil, err
	}
	return ari.NewPlaybackHandle(k, b.c.Playback(), nil), nil
}

func (b *bridge) StagePlay(key *ari.Key, id string, uri ...string) (*ari.PlaybackHandle, error) {
	return b.StagePlayWithOptions(key, id, ari.BridgePlayOptions{Media: uri})
}

func (b *bridge) StagePlayWithOptions(key *ari.Key, id string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	if id == "" {
		id = rid.New(rid.Playback)
	}
	k, err := b.c.getRequest(&proxy.Request{
		Kind:       "BridgeStagePlay",
		Key:        key,
		BridgePlay: bridgePlayPayload(id, opts),
	})
	if err != nil {
		return nil, err
	}

	return ari.NewPlaybackHandle(k.New(ari.PlaybackKey, id), b.c.Playback(), func(h *ari.PlaybackHandle) error {
		_, err := b.PlayWithOptions(k.New(ari.BridgeKey, key.ID), id, opts)
		return err
	}), nil
}

func bridgePlayPayload(id string, opts ari.BridgePlayOptions) *proxy.BridgePlay {
	media := opts.Media
	p := &proxy.BridgePlay{PlaybackID: id, Options: &opts}
	if len(media) > 0 {
		p.MediaURI = media[0]
	}
	if len(media) > 1 {
		p.MediaURIs = media
	}
	return p
}

func (b *bridge) Record(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	if opts == nil {
		opts = &ari.RecordingOptions{}
	}
	if name == "" {
		name = rid.New(rid.Recording)
	}

	k, err := b.c.createRequest(&proxy.Request{
		Kind: "BridgeRecord",
		Key:  key,
		BridgeRecord: &proxy.BridgeRecord{
			Name:    name,
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}
	return ari.NewLiveRecordingHandle(k.New(ari.LiveRecordingKey, name), b.c.LiveRecording(), nil), nil
}

func (b *bridge) StageRecord(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	if opts == nil {
		opts = &ari.RecordingOptions{}
	}
	if name == "" {
		name = rid.New(rid.Recording)
	}

	k, err := b.c.getRequest(&proxy.Request{
		Kind: "BridgeStageRecord",
		Key:  key,
		BridgeRecord: &proxy.BridgeRecord{
			Name:    name,
			Options: opts,
		},
	})
	if err != nil {
		return nil, err
	}

	return ari.NewLiveRecordingHandle(k.New(ari.LiveRecordingKey, name), b.c.LiveRecording(), func(h *ari.LiveRecordingHandle) error {
		_, err := b.Record(k.New(ari.BridgeKey, key.ID), name, opts)
		return err
	}), nil
}

func (b *bridge) Subscribe(key *ari.Key, n ...string) ari.Subscription {
	err := b.c.commandRequest(&proxy.Request{
		Kind: "BridgeSubscribe",
		Key:  key,
	})
	if err != nil {
		b.c.log.Warn("failed to call bridge subscribe")
		if key.Dialog != "" {
			b.c.log.Error("dialog present; failing")
			return nil
		}
	}

	return b.c.Bus().Subscribe(key, n...)
}

func (b *bridge) VideoSource(key *ari.Key, channelID string) error {
	return b.c.commandRequest(&proxy.Request{
		Kind: "BridgeVideoSource",
		Key:  key,
		BridgeVideoSource: &proxy.BridgeVideoSource{
			Channel: channelID,
		},
	})
}

func (b *bridge) VideoSourceDelete(key *ari.Key) error {
	return b.c.commandRequest(&proxy.Request{
		Kind: "BridgeVideoSourceDelete",
		Key:  key,
	})
}
