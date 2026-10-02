// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/client/cluster"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

func (b *parityRequestBus) MultipleRequest(subject string, req *proxy.Request, _ int) ([]*proxy.Response, error) {
	response, err := b.Request(subject, req)
	if err != nil {
		return nil, err
	}
	return []*proxy.Response{response}, nil
}

func TestSimpleCommandsSurviveProxyEncoding(t *testing.T) {
	bus := &parityRequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test.", cluster: cluster.New()}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "item-1"}
	tests := []struct {
		name, kind string
		call       func() error
		check      func(*proxy.Request) bool
	}{
		{"config fields", "AsteriskConfigUpdate", func() error {
			return c.Asterisk().Config().Update(key.New("config", "sorcery/endpoint/alice"), []ari.ConfigTuple{{Attribute: "callerid", Value: "Alice <100>"}})
		}, func(r *proxy.Request) bool {
			return r.AsteriskConfig != nil && reflect.DeepEqual(r.AsteriskConfig.Tuples, []ari.ConfigTuple{{Attribute: "callerid", Value: "Alice <100>"}})
		}},
		{"sound filters", "SoundList", func() error {
			_, err := c.Sound().List(map[string]string{"lang": "en+US", "format": "wav"}, key)
			return err
		}, func(r *proxy.Request) bool {
			return r.SoundList != nil && reflect.DeepEqual(r.SoundList.Filters, map[string]string{"lang": "en+US", "format": "wav"})
		}},
		{"logging", "AsteriskLoggingCreate", func() error { _, err := c.Asterisk().Logging().Create(key, "notice,error+debug"); return err }, func(r *proxy.Request) bool {
			return r.AsteriskLoggingChannel != nil && r.AsteriskLoggingChannel.Levels == "notice,error+debug"
		}},
		{"device", "DeviceStateUpdate", func() error { return c.DeviceState().Update(key, "INUSE+BUSY") }, func(r *proxy.Request) bool {
			return r.DeviceStateUpdate != nil && r.DeviceStateUpdate.State == "INUSE+BUSY"
		}},
		{"mailbox", "MailboxUpdate", func() error { return c.Mailbox().Update(key, 2, 3) }, func(r *proxy.Request) bool {
			return r.MailboxUpdate != nil && r.MailboxUpdate.Old == 2 && r.MailboxUpdate.New == 3
		}},
		{"playback", "PlaybackControl", func() error { return c.Playback().Control(key, "forward") }, func(r *proxy.Request) bool { return r.PlaybackControl != nil && r.PlaybackControl.Command == "forward" }},
		{"recording copy", "RecordingStoredCopy", func() error { _, err := c.StoredRecording().Copy(key, "target+1"); return err }, func(r *proxy.Request) bool {
			return r.RecordingStoredCopy != nil && r.RecordingStoredCopy.Destination == "target+1"
		}},
		{"move", "ChannelMove", func() error { return c.Channel().Move(key, "demo+app", "one,two") }, func(r *proxy.Request) bool {
			return r.ChannelMove != nil && r.ChannelMove.App == "demo+app" && r.ChannelMove.AppArgs == "one,two"
		}},
		{"mute", "ChannelMute", func() error { return c.Channel().Mute(key, ari.DirectionIn) }, func(r *proxy.Request) bool { return r.ChannelMute != nil && r.ChannelMute.Direction == ari.DirectionIn }},
		{"unmute", "ChannelUnmute", func() error { return c.Channel().Unmute(key, ari.DirectionOut) }, func(r *proxy.Request) bool {
			return r.ChannelMute != nil && r.ChannelMute.Direction == ari.DirectionOut
		}},
		{"moh", "ChannelMOH", func() error { return c.Channel().MOH(key, "custom+class") }, func(r *proxy.Request) bool { return r.ChannelMOH != nil && r.ChannelMOH.Music == "custom+class" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err != nil {
				t.Fatal(err)
			}
			if bus.request == nil || bus.request.Kind != test.kind || !test.check(bus.request) {
				t.Fatalf("request=%+v", bus.request)
			}
		})
	}
}
