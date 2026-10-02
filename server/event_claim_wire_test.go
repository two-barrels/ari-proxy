// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type claimClient struct {
	ari.Client
	application ari.Application
}

func (c claimClient) Application() ari.Application { return c.application }

type claimApplication struct {
	ari.Application
	key       *ari.Key
	channelID string
	err       error
}

func (a *claimApplication) ClaimChannel(key *ari.Key, channelID string) error {
	a.key, a.channelID = key, channelID
	return a.err
}

func TestClaimChannelServerDispatchAndStatus(t *testing.T) {
	application := &claimApplication{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = claimClient{application: application}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ApplicationKey, ID: "demo"}
	req := &proxy.Request{Kind: "EventClaimChannel", Key: key, EventClaim: &proxy.EventClaim{ChannelID: "channel+1"}}
	s.dispatchRequest(context.Background(), "reply", req)
	if application.key != key || application.channelID != "channel+1" || bus.response == nil || bus.response.Err() != nil {
		t.Fatalf("forwarded: key=%v channel=%q response=%+v", application.key, application.channelID, bus.response)
	}
	application.err = &proxy.StatusError{Message: "already claimed", Status: http.StatusConflict}
	s.dispatchRequest(context.Background(), "reply", req)
	if bus.response == nil || bus.response.StatusCode != http.StatusConflict {
		t.Fatalf("conflict response=%+v", bus.response)
	}
}
