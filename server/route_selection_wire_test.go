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

type routeForwardingClient struct {
	ari.Client
	bridge  ari.Bridge
	channel ari.Channel
}

func (c routeForwardingClient) Bridge() ari.Bridge   { return c.bridge }
func (c routeForwardingClient) Channel() ari.Channel { return c.channel }

type routeBridge struct {
	ari.Bridge
	create   ari.BridgeCreateOptions
	play     ari.BridgePlayOptions
	createID string
	playID   string
}

func (b *routeBridge) CreateWithoutID(key *ari.Key, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	b.create = opts
	return ari.NewBridgeHandle(key.New(ari.BridgeKey, "server-bridge"), b, nil), nil
}
func (b *routeBridge) PlayWithoutID(key *ari.Key, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	b.play = opts
	return ari.NewPlaybackHandle(key.New(ari.PlaybackKey, "server-playback"), nil, nil), nil
}
func (b *routeBridge) CreateOnCollection(key *ari.Key, id string, opts ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	b.createID, b.create = id, opts
	return ari.NewBridgeHandle(key.New(ari.BridgeKey, id), b, nil), nil
}
func (b *routeBridge) PlayOnCollection(key *ari.Key, id string, opts ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	b.playID, b.play = id, opts
	return ari.NewPlaybackHandle(key.New(ari.PlaybackKey, id), nil, nil), nil
}

type routeChannel struct {
	ari.Channel
	originate ari.OriginateRequest
	play      ari.ChannelPlayOptions
	snoop     *ari.SnoopOptions
	playID    string
	snoopID   string
}

func (c *routeChannel) Snoop(key *ari.Key, id string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	c.snoopID, c.snoop = id, opts
	return ari.NewChannelHandle(key.New(ari.ChannelKey, id), c, nil), nil
}

func (c *routeChannel) OriginateWithID(key *ari.Key, req ari.OriginateRequest) (*ari.ChannelHandle, error) {
	c.originate = req
	return ari.NewChannelHandle(key.New(ari.ChannelKey, req.ChannelID), c, nil), nil
}
func (c *routeChannel) PlayWithoutID(key *ari.Key, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	c.play = opts
	return ari.NewPlaybackHandle(key.New(ari.PlaybackKey, "server-playback"), nil, nil), nil
}
func (c *routeChannel) SnoopWithoutID(key *ari.Key, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	c.snoop = opts
	return ari.NewChannelHandle(key.New(ari.ChannelKey, "server-snoop"), c, nil), nil
}
func (c *routeChannel) PlayOnCollection(key *ari.Key, id string, opts ari.ChannelPlayOptions) (*ari.PlaybackHandle, error) {
	c.playID, c.play = id, opts
	return ari.NewPlaybackHandle(key.New(ari.PlaybackKey, id), nil, nil), nil
}
func (c *routeChannel) SnoopOnCollection(key *ari.Key, id string, opts *ari.SnoopOptions) (*ari.ChannelHandle, error) {
	c.snoopID, c.snoop = id, opts
	return ari.NewChannelHandle(key.New(ari.ChannelKey, id), c, nil), nil
}

func TestExplicitRoutesForwardThroughServer(t *testing.T) {
	bridge := &routeBridge{}
	channel := &routeChannel{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = routeForwardingClient{bridge: bridge, channel: channel}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}
	check := func(req *proxy.Request, kind, id string) {
		t.Helper()
		bus.response = nil
		s.dispatchRequest(context.Background(), "reply", req)
		if bus.response == nil || bus.response.Err() != nil || bus.response.Key == nil || bus.response.Key.Kind != kind || bus.response.Key.ID != id {
			t.Fatalf("request=%s response=%+v", req.Kind, bus.response)
		}
	}
	zero := 0
	create := ari.BridgeCreateOptions{Type: "mixing", Name: "support+one", Variables: map[string]ari.BridgeCreateVariable{"STATE": {Value: "Ready"}}}
	check(&proxy.Request{Kind: "BridgeCreateWithoutID", Key: key, BridgeCreate: &proxy.BridgeCreate{Type: create.Type, Name: create.Name, Variables: create.Variables}}, ari.BridgeKey, "server-bridge")
	if !reflect.DeepEqual(bridge.create, create) {
		t.Errorf("bridge create=%+v", bridge.create)
	}
	check(&proxy.Request{Kind: "BridgeCreateOnCollection", Key: key, BridgeCreate: &proxy.BridgeCreate{BridgeID: "bridge+9", Type: create.Type, Name: create.Name, Variables: create.Variables}}, ari.BridgeKey, "bridge+9")
	if bridge.createID != "bridge+9" || !reflect.DeepEqual(bridge.create, create) {
		t.Errorf("bridge create ID=%q options=%+v", bridge.createID, bridge.create)
	}
	skip, offset := 4500, 300
	bplay := ari.BridgePlayOptions{Media: []string{"sound:one"}, AnnouncerFormat: "slin16", Lang: "en", OffsetMS: &zero, SkipMS: &skip}
	check(&proxy.Request{Kind: "BridgePlayWithoutID", Key: key.New(ari.BridgeKey, "bridge-1"), BridgePlay: &proxy.BridgePlay{Options: &bplay}}, ari.PlaybackKey, "server-playback")
	if !reflect.DeepEqual(bridge.play, bplay) {
		t.Errorf("bridge play=%+v", bridge.play)
	}
	check(&proxy.Request{Kind: "BridgePlayOnCollection", Key: key.New(ari.BridgeKey, "bridge-1"), BridgePlay: &proxy.BridgePlay{PlaybackID: "play+9", Options: &bplay}}, ari.PlaybackKey, "play+9")
	if bridge.playID != "play+9" {
		t.Errorf("bridge play ID=%q", bridge.playID)
	}
	orig := ari.OriginateRequest{ChannelID: "channel-9", Endpoint: "PJSIP/alice", Extension: "s+1", Context: "support/main", Priority: 2, Label: "start+1", App: "demo", AppArgs: "one,two", CallerID: "Alice <100>", Timeout: 15, OtherChannelID: "other-9", Originator: "parent-1", Formats: "ulaw,slin16", Variables: map[string]string{"STATE": "Ready"}}
	check(&proxy.Request{Kind: "ChannelOriginateWithID", Key: key, ChannelOriginate: &proxy.ChannelOriginate{OriginateRequest: orig}}, ari.ChannelKey, "channel-9")
	if !reflect.DeepEqual(channel.originate, orig) {
		t.Errorf("originate=%+v", channel.originate)
	}
	cplay := ari.ChannelPlayOptions{Media: []string{"sound:two"}, Lang: "en+US", OffsetMS: &offset, SkipMS: &zero}
	check(&proxy.Request{Kind: "ChannelPlayWithoutID", Key: key, ChannelPlay: &proxy.ChannelPlay{Options: &cplay}}, ari.PlaybackKey, "server-playback")
	if !reflect.DeepEqual(channel.play, cplay) {
		t.Errorf("channel play=%+v", channel.play)
	}
	check(&proxy.Request{Kind: "ChannelPlayOnCollection", Key: key, ChannelPlay: &proxy.ChannelPlay{PlaybackID: "play+9", Options: &cplay}}, ari.PlaybackKey, "play+9")
	if channel.playID != "play+9" {
		t.Errorf("channel play ID=%q", channel.playID)
	}
	snoop := &ari.SnoopOptions{App: "demo", AppArgs: "one,two", Spy: ari.DirectionIn, Whisper: ari.DirectionOut}
	check(&proxy.Request{Kind: "ChannelSnoopWithoutID", Key: key, ChannelSnoop: &proxy.ChannelSnoop{Options: snoop}}, ari.ChannelKey, "server-snoop")
	if !reflect.DeepEqual(channel.snoop, snoop) {
		t.Errorf("snoop=%+v", channel.snoop)
	}
	check(&proxy.Request{Kind: "ChannelSnoop", Key: key, ChannelSnoop: &proxy.ChannelSnoop{SnoopID: "snoop-9", Options: snoop}}, ari.ChannelKey, "snoop-9")
	if channel.snoopID != "snoop-9" || !reflect.DeepEqual(channel.snoop, snoop) {
		t.Errorf("snoop path ID=%q options=%+v", channel.snoopID, channel.snoop)
	}
	check(&proxy.Request{Kind: "ChannelSnoopOnCollection", Key: key, ChannelSnoop: &proxy.ChannelSnoop{SnoopID: "snoop+9", Options: snoop}}, ari.ChannelKey, "snoop+9")
	if channel.snoopID != "snoop+9" {
		t.Errorf("snoop ID=%q", channel.snoopID)
	}
}
