// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type queryWireClient struct {
	ari.Client
	application ari.Application
	channel     ari.Channel
	asterisk    ari.Asterisk
}

func (c queryWireClient) Application() ari.Application { return c.application }
func (c queryWireClient) Channel() ari.Channel         { return c.channel }
func (c queryWireClient) Asterisk() ari.Asterisk       { return c.asterisk }

type queryWireApplication struct {
	ari.Application
	eventSource string
}

func (a *queryWireApplication) Subscribe(_ *ari.Key, eventSource string) error {
	a.eventSource = eventSource
	return nil
}

func (a *queryWireApplication) Unsubscribe(_ *ari.Key, eventSource string) error {
	a.eventSource = eventSource
	return nil
}

type queryWireChannel struct {
	ari.Channel
	variable string
}

func (c *queryWireChannel) GetVariable(_ *ari.Key, variable string) (string, error) {
	c.variable = variable
	return "value-1", nil
}

type queryWireAsterisk struct {
	ari.Asterisk
	variables ari.AsteriskVariables
}

func (a queryWireAsterisk) Variables() ari.AsteriskVariables { return a.variables }

type queryWireVariables struct {
	ari.AsteriskVariables
	variable string
	value    string
}

func (v *queryWireVariables) Set(key *ari.Key, value string) error {
	v.variable, v.value = key.ID, value
	return nil
}

func (v *queryWireVariables) Get(key *ari.Key) (string, error) {
	v.variable = key.ID
	return "value-1", nil
}

func TestServerForwardsQueryValues(t *testing.T) {
	application := &queryWireApplication{}
	channel := &queryWireChannel{}
	variables := &queryWireVariables{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = queryWireClient{
		application: application, channel: channel,
		asterisk: queryWireAsterisk{variables: variables},
	}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}

	s.dispatchRequest(context.Background(), "reply", &proxy.Request{
		Kind: "ApplicationUnsubscribe", Key: key,
		ApplicationSubscribe: &proxy.ApplicationSubscribe{EventSource: "channel:call+leg"},
	})
	if application.eventSource != "channel:call+leg" || bus.response == nil || bus.response.Err() != nil {
		t.Errorf("application event source = %q, response = %+v", application.eventSource, bus.response)
	}

	s.dispatchRequest(context.Background(), "reply", &proxy.Request{
		Kind: "ChannelVariableGet", Key: key,
		ChannelVariable: &proxy.ChannelVariable{Name: "CUSTOM+NAME"},
	})
	if channel.variable != "CUSTOM+NAME" || bus.response == nil || bus.response.Data == nil || bus.response.Data.Variable != "value-1" {
		t.Errorf("channel variable = %q, response = %+v", channel.variable, bus.response)
	}

	variableKey := &ari.Key{App: "demo", Node: "node-1", Kind: ari.VariableKey, ID: "CUSTOM+NAME"}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "AsteriskVariableGet", Key: variableKey})
	if variables.variable != "CUSTOM+NAME" || bus.response == nil || bus.response.Data == nil || bus.response.Data.Variable != "value-1" {
		t.Errorf("global variable = %q, response = %+v", variables.variable, bus.response)
	}
}

func TestServerForwardsQueryWrites(t *testing.T) {
	application := &queryWireApplication{}
	variables := &queryWireVariables{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = queryWireClient{application: application, asterisk: queryWireAsterisk{variables: variables}}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ApplicationKey, ID: "demo"}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "ApplicationSubscribe", Key: key,
		ApplicationSubscribe: &proxy.ApplicationSubscribe{EventSource: "channel:call+leg/2"}})
	if application.eventSource != "channel:call+leg/2" || bus.response == nil || bus.response.Err() != nil {
		t.Fatalf("subscribe source=%q response=%+v", application.eventSource, bus.response)
	}
	variableKey := key.New(ari.VariableKey, "CUSTOM+NAME")
	for _, value := range []string{"", "a=b&c=d"} {
		bus.response = nil
		s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "AsteriskVariableSet", Key: variableKey,
			AsteriskVariableSet: &proxy.AsteriskVariableSet{Value: value}})
		if variables.variable != "CUSTOM+NAME" || variables.value != value || bus.response == nil || bus.response.Err() != nil {
			t.Fatalf("variable=%q value=%q response=%+v", variables.variable, variables.value, bus.response)
		}
	}
}
