// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type simpleCommandState struct {
	logging, device, playback, copy, moveApp, moveArgs, moh string
	oldMessages, newMessages                                int
	mute, unmute                                            ari.Direction
	fields                                                  []ari.ConfigTuple
	soundFilters                                            map[string]string
}
type simpleCommandClient struct {
	ari.Client
	state *simpleCommandState
}

func (c simpleCommandClient) Asterisk() ari.Asterisk       { return simpleAsterisk{state: c.state} }
func (c simpleCommandClient) DeviceState() ari.DeviceState { return simpleDevice{state: c.state} }
func (c simpleCommandClient) Mailbox() ari.Mailbox         { return simpleMailbox{state: c.state} }
func (c simpleCommandClient) Playback() ari.Playback       { return simplePlayback{state: c.state} }
func (c simpleCommandClient) StoredRecording() ari.StoredRecording {
	return simpleRecording{state: c.state}
}
func (c simpleCommandClient) Channel() ari.Channel { return simpleChannel{state: c.state} }
func (c simpleCommandClient) Sound() ari.Sound     { return simpleSound{state: c.state} }

type simpleAsterisk struct {
	ari.Asterisk
	state *simpleCommandState
}

func (a simpleAsterisk) Logging() ari.Logging { return simpleLogging{state: a.state} }
func (a simpleAsterisk) Config() ari.Config   { return simpleConfig{state: a.state} }

type simpleConfig struct {
	ari.Config
	state *simpleCommandState
}

func (c simpleConfig) Update(_ *ari.Key, tuples []ari.ConfigTuple) error {
	c.state.fields = tuples
	return nil
}

type simpleSound struct {
	ari.Sound
	state *simpleCommandState
}

func (s simpleSound) List(filters map[string]string, _ *ari.Key) ([]*ari.Key, error) {
	s.state.soundFilters = filters
	return nil, nil
}

type simpleLogging struct {
	ari.Logging
	state *simpleCommandState
}

func (l simpleLogging) Create(key *ari.Key, levels string) (*ari.LogHandle, error) {
	l.state.logging = levels
	return ari.NewLogHandle(key, l), nil
}

type simpleDevice struct {
	ari.DeviceState
	state *simpleCommandState
}

func (d simpleDevice) Update(_ *ari.Key, value string) error { d.state.device = value; return nil }

type simpleMailbox struct {
	ari.Mailbox
	state *simpleCommandState
}

func (m simpleMailbox) Update(_ *ari.Key, old, newer int) error {
	m.state.oldMessages, m.state.newMessages = old, newer
	return nil
}

type simplePlayback struct {
	ari.Playback
	state *simpleCommandState
}

func (p simplePlayback) Control(_ *ari.Key, op string) error { p.state.playback = op; return nil }

type simpleRecording struct {
	ari.StoredRecording
	state *simpleCommandState
}

func (r simpleRecording) Copy(key *ari.Key, dest string) (*ari.StoredRecordingHandle, error) {
	r.state.copy = dest
	return ari.NewStoredRecordingHandle(key.New(ari.StoredRecordingKey, dest), r, nil), nil
}

type simpleChannel struct {
	ari.Channel
	state *simpleCommandState
}

func (c simpleChannel) Move(_ *ari.Key, app, args string) error {
	c.state.moveApp, c.state.moveArgs = app, args
	return nil
}
func (c simpleChannel) Mute(_ *ari.Key, dir ari.Direction) error   { c.state.mute = dir; return nil }
func (c simpleChannel) Unmute(_ *ari.Key, dir ari.Direction) error { c.state.unmute = dir; return nil }
func (c simpleChannel) MOH(_ *ari.Key, class string) error         { c.state.moh = class; return nil }

func TestSimpleCommandsReachNativeFromProxyServer(t *testing.T) {
	state := &simpleCommandState{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = simpleCommandClient{state: state}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "item-1"}
	tests := []struct {
		name    string
		request *proxy.Request
		check   func() bool
	}{
		{"config fields", &proxy.Request{Kind: "AsteriskConfigUpdate", Key: key, AsteriskConfig: &proxy.AsteriskConfig{Tuples: []ari.ConfigTuple{{Attribute: "callerid", Value: "Alice <100>"}}}}, func() bool {
			return reflect.DeepEqual(state.fields, []ari.ConfigTuple{{Attribute: "callerid", Value: "Alice <100>"}})
		}},
		{"sound filters", &proxy.Request{Kind: "SoundList", Key: key, SoundList: &proxy.SoundList{Filters: map[string]string{"lang": "en+US", "format": "wav"}}}, func() bool {
			return reflect.DeepEqual(state.soundFilters, map[string]string{"lang": "en+US", "format": "wav"})
		}},
		{"logging", &proxy.Request{Kind: "AsteriskLoggingCreate", Key: key, AsteriskLoggingChannel: &proxy.AsteriskLoggingChannel{Levels: "notice,error+debug"}}, func() bool { return state.logging == "notice,error+debug" }},
		{"device", &proxy.Request{Kind: "DeviceStateUpdate", Key: key, DeviceStateUpdate: &proxy.DeviceStateUpdate{State: "INUSE+BUSY"}}, func() bool { return state.device == "INUSE+BUSY" }},
		{"mailbox", &proxy.Request{Kind: "MailboxUpdate", Key: key, MailboxUpdate: &proxy.MailboxUpdate{Old: 2, New: 3}}, func() bool { return state.oldMessages == 2 && state.newMessages == 3 }},
		{"playback", &proxy.Request{Kind: "PlaybackControl", Key: key, PlaybackControl: &proxy.PlaybackControl{Command: "forward"}}, func() bool { return state.playback == "forward" }},
		{"copy", &proxy.Request{Kind: "RecordingStoredCopy", Key: key, RecordingStoredCopy: &proxy.RecordingStoredCopy{Destination: "target+1"}}, func() bool { return state.copy == "target+1" }},
		{"move", &proxy.Request{Kind: "ChannelMove", Key: key, ChannelMove: &proxy.ChannelMove{App: "demo+app", AppArgs: "one,two"}}, func() bool { return state.moveApp == "demo+app" && state.moveArgs == "one,two" }},
		{"mute", &proxy.Request{Kind: "ChannelMute", Key: key, ChannelMute: &proxy.ChannelMute{Direction: ari.DirectionIn}}, func() bool { return state.mute == ari.DirectionIn }},
		{"unmute", &proxy.Request{Kind: "ChannelUnmute", Key: key, ChannelMute: &proxy.ChannelMute{Direction: ari.DirectionOut}}, func() bool { return state.unmute == ari.DirectionOut }},
		{"moh", &proxy.Request{Kind: "ChannelMOH", Key: key, ChannelMOH: &proxy.ChannelMOH{Music: "custom+class"}}, func() bool { return state.moh == "custom+class" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bus.response = nil
			s.dispatchRequest(context.Background(), "reply", test.request)
			if bus.response == nil || bus.response.Err() != nil || !test.check() {
				t.Fatalf("response=%+v state=%+v", bus.response, state)
			}
		})
	}
}
