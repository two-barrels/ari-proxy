// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

func TestChannelARI23ProxyRequests(t *testing.T) {
	bus := &ari23RequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}
	channel := c.Channel()
	values, err := channel.GetVariables(key, "BRIDGE_STATE", "SUPPORT_LEVEL")
	if err != nil || len(values) != 2 || bus.request.Kind != "ChannelVariablesGet" || !reflect.DeepEqual(bus.request.ChannelVariables.Names, []string{"BRIDGE_STATE", "SUPPORT_LEVEL"}) {
		t.Fatalf("get variables: values=%v request=%+v error=%v", values, bus.request, err)
	}
	report := false
	assignments := map[string]ari.VariableAssignment{"STATE": {Value: "Ready"}, "LEVEL": {Value: "Gold", ReportEvents: &report}}
	if err := channel.SetVariables(key, assignments); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "ChannelVariablesSet" || !reflect.DeepEqual(bus.request.ChannelVariables.Values, assignments) {
		t.Fatalf("set variables request=%+v", bus.request)
	}
	if err := channel.Redirect(key, "PJSIP/100+1"); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "ChannelRedirect" || bus.request.ChannelRedirect.Endpoint != "PJSIP/100+1" {
		t.Fatalf("redirect request=%+v", bus.request)
	}
	if err := channel.Progress(key); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "ChannelProgress" {
		t.Fatalf("progress request=%+v", bus.request)
	}
	if err := channel.TransferProgress(key, "channel_progress"); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "ChannelTransferProgress" || bus.request.ChannelTransferProgress.States != "channel_progress" {
		t.Fatalf("transfer progress request=%+v", bus.request)
	}
	stats, err := channel.RTPStatistics(key)
	if err != nil || stats == nil || stats.TxCount != 20 || stats.RTT != 2.5 || stats.ChannelUniqueID != "channel-1" || bus.request.Kind != "ChannelRTPStatistics" {
		t.Fatalf("rtp stats=%+v request=%+v error=%v", stats, bus.request, err)
	}
}
