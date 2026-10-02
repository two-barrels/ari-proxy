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

type channelARI23Client struct {
	ari.Client
	channel ari.Channel
}

func (c channelARI23Client) Channel() ari.Channel { return c.channel }

type channelARI23Recorder struct {
	ari.Channel
	kind   string
	key    *ari.Key
	names  []string
	values map[string]ari.VariableAssignment
	option string
}

func (r *channelARI23Recorder) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	r.kind, r.key, r.names = "get variables", key, names
	return map[string]any{"STATE": "Ready"}, nil
}
func (r *channelARI23Recorder) SetVariables(key *ari.Key, values map[string]ari.VariableAssignment) error {
	r.kind, r.key, r.values = "set variables", key, values
	return nil
}
func (r *channelARI23Recorder) Redirect(key *ari.Key, endpoint string) error {
	r.kind, r.key, r.option = "redirect", key, endpoint
	return nil
}
func (r *channelARI23Recorder) Progress(key *ari.Key) error {
	r.kind, r.key = "progress", key
	return nil
}
func (r *channelARI23Recorder) TransferProgress(key *ari.Key, state string) error {
	r.kind, r.key, r.option = "transfer progress", key, state
	return nil
}
func (r *channelARI23Recorder) RTPStatistics(key *ari.Key) (*ari.RTPStats, error) {
	r.kind, r.key = "rtp statistics", key
	return &ari.RTPStats{TxCount: 20, RTT: 2.5, ChannelUniqueID: "channel-1"}, nil
}

func TestChannelARI23ServerDispatch(t *testing.T) {
	recorder := &channelARI23Recorder{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = channelARI23Client{channel: recorder}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}
	report := false
	assignments := map[string]ari.VariableAssignment{"LEVEL": {Value: "Gold", ReportEvents: &report}}
	tests := []struct {
		kind, called, option string
		request              *proxy.Request
		check                func(*testing.T, *proxy.Response)
	}{
		{"ChannelVariablesGet", "get variables", "", &proxy.Request{ChannelVariables: &proxy.ChannelVariables{Names: []string{"STATE", "LEVEL"}}}, func(t *testing.T, response *proxy.Response) {
			if !reflect.DeepEqual(recorder.names, []string{"STATE", "LEVEL"}) || response.Data == nil || response.Data.Variables["STATE"] != "Ready" {
				t.Errorf("get variables: names=%v response=%+v", recorder.names, response)
			}
		}},
		{"ChannelVariablesSet", "set variables", "", &proxy.Request{ChannelVariables: &proxy.ChannelVariables{Values: assignments}}, func(t *testing.T, _ *proxy.Response) {
			if !reflect.DeepEqual(recorder.values, assignments) {
				t.Errorf("assignments=%+v", recorder.values)
			}
		}},
		{"ChannelRedirect", "redirect", "PJSIP/100+1", &proxy.Request{ChannelRedirect: &proxy.ChannelRedirect{Endpoint: "PJSIP/100+1"}}, nil},
		{"ChannelProgress", "progress", "", &proxy.Request{}, nil},
		{"ChannelTransferProgress", "transfer progress", "channel_progress", &proxy.Request{ChannelTransferProgress: &proxy.ChannelTransferProgress{States: "channel_progress"}}, nil},
		{"ChannelRTPStatistics", "rtp statistics", "", &proxy.Request{}, func(t *testing.T, response *proxy.Response) {
			if response.Data == nil || response.Data.RTPStats == nil || response.Data.RTPStats.TxCount != 20 || response.Data.RTPStats.RTT != 2.5 {
				t.Errorf("rtp response=%+v", response)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			bus.response = nil
			test.request.Kind, test.request.Key = test.kind, key
			s.dispatchRequest(context.Background(), "reply", test.request)
			if recorder.kind != test.called || recorder.key != key || recorder.option != test.option && test.option != "" {
				t.Errorf("call=%s key=%v option=%q", recorder.kind, recorder.key, recorder.option)
			}
			if bus.response == nil || bus.response.Err() != nil {
				t.Fatalf("response=%+v", bus.response)
			}
			if test.check != nil {
				test.check(t, bus.response)
			}
		})
	}
}
