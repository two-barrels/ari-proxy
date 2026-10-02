// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type optionWireBus struct {
	messagebus.Client
	requests []*proxy.Request
}

func (b *optionWireBus) Request(_ string, request *proxy.Request) (*proxy.Response, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var decoded proxy.Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	b.requests = append(b.requests, &decoded)
	return &proxy.Response{Data: &proxy.EntityData{Variable: "value-1"}, Key: &ari.Key{
		App: request.Key.App, Node: request.Key.Node, Kind: ari.ChannelKey, ID: "created-1",
	}}, nil
}

func TestClientOptionsSurviveMessageBusEncoding(t *testing.T) {
	bridgeOptions := &ari.BridgeAddChannelOptions{Role: "caller", AbsorbDTMF: true, Mute: true}
	originate := ari.OriginateRequest{
		Endpoint: "PJSIP/alice", Extension: "s+1", Context: "support/main", Priority: 2, Label: "start+1",
		App: "demo", AppArgs: "one,two", ChannelID: "channel-1", OtherChannelID: "channel-2",
		CallerID: "Alice <100>", Timeout: 15, Originator: "parent+1",
		Formats: "ulaw,slin16", Variables: map[string]string{"ticket": "42"},
	}
	external := ari.ExternalMediaOptions{
		ChannelID: "external-1", App: "demo", ExternalHost: "127.0.0.1:5000",
		Encapsulation: "rtp", Format: "slin16", Transport: "udp", ConnectionType: "client",
		Direction: "both", Data: "stream-1",
		Variables: map[string]string{"ticket": "42"},
	}
	record := &ari.RecordingOptions{
		Format: "wav", MaxDuration: 10 * time.Second, MaxSilence: 2 * time.Second,
		Exists: "overwrite", Beep: true, Terminate: "#",
	}
	dtmf := &ari.DTMFOptions{
		Before: 50 * time.Millisecond, Between: 75 * time.Millisecond,
		Duration: 120 * time.Millisecond, After: 25 * time.Millisecond,
	}
	tests := []struct {
		name  string
		kind  string
		call  func(*Client, *ari.Key) error
		check func(*testing.T, *proxy.Request)
	}{
		{
			name: "bridge add channel", kind: "BridgeAddChannel",
			call: func(c *Client, key *ari.Key) error {
				return c.Bridge().AddChannelWithOptions(key, "channel-1", bridgeOptions)
			},
			check: func(t *testing.T, request *proxy.Request) {
				want := &proxy.BridgeAddChannel{Channel: "channel-1", Role: "caller", AbsorbDTMF: true, Mute: true}
				if !reflect.DeepEqual(request.BridgeAddChannel, want) {
					t.Errorf("bridge options = %+v, want %+v", request.BridgeAddChannel, want)
				}
			},
		},
		{
			name: "channel originate", kind: "ChannelOriginate",
			call: func(c *Client, key *ari.Key) error {
				_, err := c.Channel().Originate(key, originate)
				return err
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.ChannelOriginate == nil || !reflect.DeepEqual(request.ChannelOriginate.OriginateRequest, originate) {
					t.Errorf("originate options = %+v, want %+v", request.ChannelOriginate, originate)
				}
			},
		},
		{
			name: "external media", kind: "ChannelExternalMedia",
			call: func(c *Client, key *ari.Key) error {
				_, err := c.Channel().ExternalMedia(key, external)
				return err
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.ChannelExternalMedia == nil || !reflect.DeepEqual(request.ChannelExternalMedia.Options, external) {
					t.Errorf("external media options = %+v, want %+v", request.ChannelExternalMedia, external)
				}
			},
		},
		{
			name: "bridge playback", kind: "BridgePlay",
			call: func(c *Client, key *ari.Key) error {
				_, err := c.Bridge().Play(key, "playback-1", "sound:one", "sound:two")
				return err
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.BridgePlay == nil || !reflect.DeepEqual(request.BridgePlay.URIs(), []string{"sound:one", "sound:two"}) {
					t.Errorf("playback media = %+v", request.BridgePlay)
				}
			},
		},
		{
			name: "channel record", kind: "ChannelRecord",
			call: func(c *Client, key *ari.Key) error {
				_, err := c.Channel().Record(key, "recording-1", record)
				return err
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.ChannelRecord == nil || request.ChannelRecord.Name != "recording-1" || !reflect.DeepEqual(request.ChannelRecord.Options, record) {
					t.Errorf("recording options = %+v, want %+v", request.ChannelRecord, record)
				}
			},
		},
		{
			name: "channel DTMF", kind: "ChannelSendDTMF",
			call: func(c *Client, key *ari.Key) error {
				return c.Channel().SendDTMF(key, "12#", dtmf)
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.ChannelSendDTMF == nil || request.ChannelSendDTMF.DTMF != "12#" ||
					!reflect.DeepEqual(request.ChannelSendDTMF.Options, dtmf) {
					t.Errorf("DTMF options = %+v, want %+v", request.ChannelSendDTMF, dtmf)
				}
			},
		},
		{
			name: "application unsubscribe", kind: "ApplicationUnsubscribe",
			call: func(c *Client, key *ari.Key) error {
				return c.Application().Unsubscribe(key, "channel:call+leg")
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.ApplicationSubscribe == nil || request.ApplicationSubscribe.EventSource != "channel:call+leg" {
					t.Errorf("event source = %+v", request.ApplicationSubscribe)
				}
			},
		},
		{
			name: "channel variable get", kind: "ChannelVariableGet",
			call: func(c *Client, key *ari.Key) error {
				_, err := c.Channel().GetVariable(key, "CUSTOM+NAME")
				return err
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.ChannelVariable == nil || request.ChannelVariable.Name != "CUSTOM+NAME" {
					t.Errorf("channel variable = %+v", request.ChannelVariable)
				}
			},
		},
		{
			name: "global variable get", kind: "AsteriskVariableGet",
			call: func(c *Client, key *ari.Key) error {
				variableKey := *key
				variableKey.Kind, variableKey.ID = ari.VariableKey, "CUSTOM+NAME"
				_, err := c.Asterisk().Variables().Get(&variableKey)
				return err
			},
			check: func(t *testing.T, request *proxy.Request) {
				if request.Key == nil || request.Key.ID != "CUSTOM+NAME" {
					t.Errorf("global variable key = %+v", request.Key)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bus := &optionWireBus{}
			logger := log15.New()
			logger.SetHandler(log15.DiscardHandler())
			client := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
			key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"}
			if err := test.call(client, key); err != nil {
				t.Fatal(err)
			}
			if len(bus.requests) != 1 || bus.requests[0].Kind != test.kind {
				t.Fatalf("requests = %+v, want one %s", bus.requests, test.kind)
			}
			test.check(t, bus.requests[0])
		})
	}
}
