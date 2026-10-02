// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type optionParityClient struct {
	ari.Client
	bridge  ari.Bridge
	channel ari.Channel
}

func (c optionParityClient) Bridge() ari.Bridge   { return c.bridge }
func (c optionParityClient) Channel() ari.Channel { return c.channel }

type optionBridgeRecorder struct {
	ari.Bridge
	created  ari.BridgeCreateOptions
	added    *ari.BridgeAddChannelOptions
	mohClass string
	removed  string
	played   ari.BridgePlayOptions
	recorded *ari.RecordingOptions
}

func (r *optionBridgeRecorder) CreateWithOptions(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	r.created = opts
	return ari.NewBridgeHandle(key, r, nil), nil
}
func (r *optionBridgeRecorder) AddChannelWithOptions(_ *ari.Key, _ string, opts *ari.BridgeAddChannelOptions) error {
	r.added = opts
	return nil
}
func (r *optionBridgeRecorder) MOH(_ *ari.Key, class string) error {
	r.mohClass = class
	return nil
}
func (r *optionBridgeRecorder) RemoveChannel(_ *ari.Key, channel string) error {
	r.removed = channel
	return nil
}
func (r *optionBridgeRecorder) PlayWithOptions(key *ari.Key, id string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	r.played = opts
	return ari.NewPlaybackHandle(key.New(ari.PlaybackKey, id), nil, nil), nil
}
func (r *optionBridgeRecorder) Record(key *ari.Key, name string, opts *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	r.recorded = opts
	return ari.NewLiveRecordingHandle(key.New(ari.LiveRecordingKey, name), nil, nil), nil
}
func (r *optionBridgeRecorder) Data(key *ari.Key) (*ari.BridgeData, error) {
	return &ari.BridgeData{Key: key, ID: key.ID}, nil
}

type optionChannelRecorder struct {
	ari.Channel
	continued   ari.ChannelContinueOptions
	hungup      ari.ChannelHangupOptions
	report      *bool
	varName     string
	varValue    string
	created     ari.ChannelCreateRequest
	played      ari.ChannelPlayOptions
	external    ari.ExternalMediaOptions
	userEvent   *ari.ChannelUserevent
	dialCaller  string
	dialTimeout time.Duration
}

func (r *optionChannelRecorder) Dial(_ *ari.Key, caller string, timeout time.Duration) error {
	r.dialCaller, r.dialTimeout = caller, timeout
	return nil
}

func (r *optionChannelRecorder) ContinueWithOptions(_ *ari.Key, opts ari.ChannelContinueOptions) error {
	r.continued = opts
	return nil
}
func (r *optionChannelRecorder) HangupWithOptions(_ *ari.Key, opts ari.ChannelHangupOptions) error {
	r.hungup = opts
	return nil
}
func (r *optionChannelRecorder) SetVariableWithOptions(_ *ari.Key, name, value string, opts *ari.ChannelVariableSetOptions) error {
	r.varName, r.varValue = name, value
	r.report = opts.ReportEvents
	return nil
}
func (r *optionChannelRecorder) Create(key *ari.Key, opts ari.ChannelCreateRequest) (*ari.ChannelHandle, error) {
	r.created = opts
	return ari.NewChannelHandle(key.New(ari.ChannelKey, opts.ChannelID), r, nil), nil
}
func (r *optionChannelRecorder) Data(key *ari.Key) (*ari.ChannelData, error) {
	return &ari.ChannelData{Key: key, ID: key.ID}, nil
}
func (r *optionChannelRecorder) PlayWithOptions(key *ari.Key, id string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	r.played = opts
	return ari.NewPlaybackHandle(key.New(ari.PlaybackKey, id), nil, nil), nil
}
func (r *optionChannelRecorder) ExternalMedia(key *ari.Key, opts ari.ExternalMediaOptions) (*ari.ChannelHandle, error) {
	r.external = opts
	return ari.NewChannelHandle(key.New(ari.ChannelKey, opts.ChannelID), r, nil), nil
}
func (r *optionChannelRecorder) UserEvent(_ *ari.Key, event *ari.ChannelUserevent) error {
	r.userEvent = event
	return nil
}

func TestNewOptionsReachNativeFromProxyServer(t *testing.T) {
	bridge := &optionBridgeRecorder{}
	channel := &optionChannelRecorder{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = optionParityClient{bridge: bridge, channel: channel}
	s.mbus = bus
	bridgeKey := &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"}
	channelKey := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}
	falseValue, zero := false, 0
	request := func(req *proxy.Request) {
		bus.response = nil
		s.dispatchRequest(context.Background(), "reply", req)
		if bus.response == nil || bus.response.Err() != nil {
			t.Fatalf("response=%+v", bus.response)
		}
	}
	variables := map[string]ari.BridgeCreateVariable{"STATE": {Value: "Ready", ReportEvents: &falseValue}}
	request(&proxy.Request{Kind: "BridgeCreate", Key: bridgeKey, BridgeCreate: &proxy.BridgeCreate{Type: "mixing", Name: "support+one", Variables: variables}})
	if bridge.created.Type != "mixing" || bridge.created.Name != "support+one" || !reflect.DeepEqual(bridge.created.Variables, variables) {
		t.Errorf("bridge create=%+v", bridge.created)
	}
	request(&proxy.Request{Kind: "BridgeAddChannel", Key: bridgeKey, BridgeAddChannel: &proxy.BridgeAddChannel{Channel: "channel-1", InhibitConnectedLineUpdates: &falseValue}})
	if bridge.added == nil || bridge.added.InhibitConnectedLineUpdates == nil || *bridge.added.InhibitConnectedLineUpdates {
		t.Errorf("bridge add=%+v", bridge.added)
	}
	request(&proxy.Request{Kind: "BridgeMOH", Key: bridgeKey, BridgeMOH: &proxy.BridgeMOH{Class: "custom+class"}})
	if bridge.mohClass != "custom+class" {
		t.Errorf("bridge moh=%q", bridge.mohClass)
	}
	request(&proxy.Request{Kind: "BridgeRemoveChannel", Key: bridgeKey, BridgeRemoveChannel: &proxy.BridgeRemoveChannel{Channel: "channel+1/2"}})
	if bridge.removed != "channel+1/2" {
		t.Errorf("bridge removed=%q", bridge.removed)
	}
	request(&proxy.Request{Kind: "ChannelContinue", Key: channelKey, ChannelContinue: &proxy.ChannelContinue{Options: &ari.ChannelContinueOptions{Context: "support", Extension: "s", Label: "start", Priority: &zero}}})
	if channel.continued.Context != "support" || channel.continued.Extension != "s" || channel.continued.Label != "start" || channel.continued.Priority == nil || *channel.continued.Priority != 0 {
		t.Errorf("continue=%+v", channel.continued)
	}
	request(&proxy.Request{Kind: "ChannelDial", Key: channelKey, ChannelDial: &proxy.ChannelDial{Caller: "call+leg/2", Timeout: 7 * time.Second}})
	if channel.dialCaller != "call+leg/2" || channel.dialTimeout != 7*time.Second {
		t.Errorf("dial caller=%q timeout=%s", channel.dialCaller, channel.dialTimeout)
	}
	request(&proxy.Request{Kind: "ChannelHangup", Key: channelKey, ChannelHangup: &proxy.ChannelHangup{Reason: "busy", ReasonCode: "17"}})
	if channel.hungup.Reason != "busy" || channel.hungup.ReasonCode != "17" {
		t.Errorf("hangup=%+v", channel.hungup)
	}
	request(&proxy.Request{Kind: "ChannelVariableSet", Key: channelKey, ChannelVariable: &proxy.ChannelVariable{Name: "STATE+NAME", Value: "", ReportEvents: &falseValue}})
	if channel.varName != "STATE+NAME" || channel.varValue != "" || channel.report == nil || *channel.report {
		t.Errorf("variable=%q value=%q report events=%v", channel.varName, channel.varValue, channel.report)
	}
	play := ari.BridgePlayOptions{Media: []string{"sound:one"}, AnnouncerFormat: "slin16", Lang: "en", OffsetMS: &zero}
	request(&proxy.Request{Kind: "BridgePlay", Key: bridgeKey, BridgePlay: &proxy.BridgePlay{PlaybackID: "playback-1", Options: &play}})
	if !reflect.DeepEqual(bridge.played, play) {
		t.Errorf("bridge play=%+v", bridge.played)
	}
	record := &ari.RecordingOptions{Format: "wav", RecorderFormat: "slin16", MaxDuration: 10 * time.Second, MaxSilence: 2 * time.Second, Exists: "overwrite", Beep: true, Terminate: "#"}
	request(&proxy.Request{Kind: "BridgeRecord", Key: bridgeKey, BridgeRecord: &proxy.BridgeRecord{Name: "recording-1", Options: record}})
	if !reflect.DeepEqual(bridge.recorded, record) {
		t.Errorf("bridge record=%+v", bridge.recorded)
	}
	createChannel := ari.ChannelCreateRequest{Endpoint: "PJSIP/alice", App: "demo", AppArgs: "one,two", ChannelID: "channel-2", OtherChannelID: "other-2", Originator: "parent-1", Formats: "ulaw,slin16", Variables: map[string]string{"STATE": "Ready"}}
	request(&proxy.Request{Kind: "ChannelCreate", Key: channelKey, ChannelCreate: &proxy.ChannelCreate{ChannelCreateRequest: createChannel}})
	if !reflect.DeepEqual(channel.created, createChannel) {
		t.Errorf("channel create=%+v", channel.created)
	}
	channelPlay := ari.ChannelPlayOptions{Media: []string{"sound:one"}, Lang: "en", OffsetMS: &zero}
	request(&proxy.Request{Kind: "ChannelPlay", Key: channelKey, ChannelPlay: &proxy.ChannelPlay{PlaybackID: "playback-2", Options: &channelPlay}})
	if !reflect.DeepEqual(channel.played, channelPlay) {
		t.Errorf("channel play=%+v", channel.played)
	}
	external := ari.ExternalMediaOptions{ChannelID: "external-1", App: "demo", ExternalHost: "127.0.0.1:5000", Format: "slin16", TransportData: "a=b&c=d"}
	request(&proxy.Request{Kind: "ChannelExternalMedia", Key: channelKey, ChannelExternalMedia: &proxy.ChannelExternalMedia{Options: external}})
	if !reflect.DeepEqual(channel.external, external) {
		t.Errorf("external media=%+v", channel.external)
	}
	userEvent := &ari.ChannelUserevent{Eventname: "customer/alert", Userevent: map[string]any{"ticket": "42"}}
	request(&proxy.Request{Kind: "ChannelUserEvent", Key: channelKey, ChannelUserevent: &proxy.ChannelUserevent{UserEvent: *userEvent}})
	if channel.userEvent == nil || channel.userEvent.Eventname != userEvent.Eventname || !reflect.DeepEqual(channel.userEvent.Userevent, userEvent.Userevent) {
		t.Errorf("user event=%+v", channel.userEvent)
	}
}
