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

type optionForwardingClient struct {
	ari.Client
	channel ari.Channel
	bridge  ari.Bridge
}

func (c optionForwardingClient) Channel() ari.Channel { return c.channel }
func (c optionForwardingClient) Bridge() ari.Bridge   { return c.bridge }

type optionForwardingChannel struct {
	ari.Channel
	originate     ari.OriginateRequest
	externalMedia ari.ExternalMediaOptions
	recording     *ari.RecordingOptions
	recordingName string
	dtmfDigits    string
	dtmfOptions   *ari.DTMFOptions
	called        string
}

func (c *optionForwardingChannel) Originate(_ *ari.Key, options ari.OriginateRequest) (*ari.ChannelHandle, error) {
	c.called, c.originate = "Originate", options
	return ari.NewChannelHandle(&ari.Key{Kind: ari.ChannelKey, ID: options.ChannelID}, c, nil), nil
}

func (c *optionForwardingChannel) ExternalMedia(_ *ari.Key, options ari.ExternalMediaOptions) (*ari.ChannelHandle, error) {
	c.called, c.externalMedia = "ExternalMedia", options
	return ari.NewChannelHandle(&ari.Key{Kind: ari.ChannelKey, ID: options.ChannelID}, c, nil), nil
}

func (c *optionForwardingChannel) Record(_ *ari.Key, name string, options *ari.RecordingOptions) (*ari.LiveRecordingHandle, error) {
	c.called, c.recordingName, c.recording = "Record", name, options
	return ari.NewLiveRecordingHandle(&ari.Key{Kind: ari.LiveRecordingKey, ID: name}, nil, nil), nil
}

func (c *optionForwardingChannel) SendDTMF(_ *ari.Key, digits string, options *ari.DTMFOptions) error {
	c.called, c.dtmfDigits, c.dtmfOptions = "SendDTMF", digits, options
	return nil
}

type optionForwardingBridge struct {
	ari.Bridge
	playbackID string
	media      []string
}

func (b *optionForwardingBridge) Play(_ *ari.Key, playbackID string, media ...string) (*ari.PlaybackHandle, error) {
	b.playbackID, b.media = playbackID, media
	return ari.NewPlaybackHandle(&ari.Key{Kind: ari.PlaybackKey, ID: playbackID}, nil, nil), nil
}

func TestServerForwardsMappedOptions(t *testing.T) {
	originate := ari.OriginateRequest{
		Endpoint: "PJSIP/alice", Extension: "s+1", Context: "support/main", Priority: 2, Label: "start+1",
		App: "demo", AppArgs: "one,two", ChannelID: "channel-1", OtherChannelID: "channel-2",
		CallerID: "Alice <100>", Timeout: 15, Originator: "parent+1",
		Formats: "ulaw,slin16", Variables: map[string]string{"ticket": "42"},
	}
	external := ari.ExternalMediaOptions{
		ChannelID: "external-1", App: "demo", ExternalHost: "127.0.0.1:5000",
		Encapsulation: "rtp", Transport: "udp", ConnectionType: "client", Format: "slin16",
		Direction: "both", Data: "stream-1", Variables: map[string]string{"ticket": "42"},
	}
	recording := &ari.RecordingOptions{
		Format: "wav", MaxDuration: 10 * time.Second, MaxSilence: 2 * time.Second,
		Exists: "overwrite", Beep: true, Terminate: "#",
	}
	dtmf := &ari.DTMFOptions{
		Before: 50 * time.Millisecond, Between: 75 * time.Millisecond,
		Duration: 120 * time.Millisecond, After: 25 * time.Millisecond,
	}
	tests := []struct {
		name    string
		request *proxy.Request
		check   func(*testing.T, *optionForwardingChannel, *optionForwardingBridge)
	}{
		{
			name:    "originate",
			request: &proxy.Request{Kind: "ChannelOriginate", ChannelOriginate: &proxy.ChannelOriginate{OriginateRequest: originate}},
			check: func(t *testing.T, channel *optionForwardingChannel, _ *optionForwardingBridge) {
				if channel.called != "Originate" || !reflect.DeepEqual(channel.originate, originate) {
					t.Errorf("forwarded originate = %+v", channel.originate)
				}
			},
		},
		{
			name:    "external media",
			request: &proxy.Request{Kind: "ChannelExternalMedia", ChannelExternalMedia: &proxy.ChannelExternalMedia{Options: external}},
			check: func(t *testing.T, channel *optionForwardingChannel, _ *optionForwardingBridge) {
				if channel.called != "ExternalMedia" || !reflect.DeepEqual(channel.externalMedia, external) {
					t.Errorf("forwarded external media = %+v", channel.externalMedia)
				}
			},
		},
		{
			name:    "recording",
			request: &proxy.Request{Kind: "ChannelRecord", ChannelRecord: &proxy.ChannelRecord{Name: "recording-1", Options: recording}},
			check: func(t *testing.T, channel *optionForwardingChannel, _ *optionForwardingBridge) {
				if channel.called != "Record" || channel.recordingName != "recording-1" || !reflect.DeepEqual(channel.recording, recording) {
					t.Errorf("forwarded recording = %+v", channel.recording)
				}
			},
		},
		{
			name: "bridge playback",
			request: &proxy.Request{Kind: "BridgePlay", BridgePlay: &proxy.BridgePlay{
				PlaybackID: "playback-1", MediaURIs: []string{"sound:one", "sound:two"},
			}},
			check: func(t *testing.T, _ *optionForwardingChannel, bridge *optionForwardingBridge) {
				if bridge.playbackID != "playback-1" || !reflect.DeepEqual(bridge.media, []string{"sound:one", "sound:two"}) {
					t.Errorf("forwarded playback = %q %v", bridge.playbackID, bridge.media)
				}
			},
		},
		{
			name: "DTMF",
			request: &proxy.Request{Kind: "ChannelSendDTMF", ChannelSendDTMF: &proxy.ChannelSendDTMF{
				DTMF: "12#", Options: dtmf,
			}},
			check: func(t *testing.T, channel *optionForwardingChannel, _ *optionForwardingBridge) {
				if channel.called != "SendDTMF" || channel.dtmfDigits != "12#" || !reflect.DeepEqual(channel.dtmfOptions, dtmf) {
					t.Errorf("forwarded DTMF = %q %+v", channel.dtmfDigits, channel.dtmfOptions)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			channel, bridge := &optionForwardingChannel{}, &optionForwardingBridge{}
			bus := &bridgeOptionsResponseBus{}
			s := New()
			s.ari = optionForwardingClient{channel: channel, bridge: bridge}
			s.mbus = bus
			test.request.Key = &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"}
			s.dispatchRequest(context.Background(), "reply", test.request)
			test.check(t, channel, bridge)
			if bus.response == nil || bus.response.Err() != nil {
				t.Errorf("proxy response = %+v", bus.response)
			}
		})
	}
}
