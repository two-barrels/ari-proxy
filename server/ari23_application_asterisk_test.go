package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type ari23ServerClient struct {
	ari.Client
	application ari.Application
	asterisk    ari.Asterisk
}

func (c ari23ServerClient) Application() ari.Application { return c.application }
func (c ari23ServerClient) Asterisk() ari.Asterisk       { return c.asterisk }

type ari23Application struct {
	ari.Application
	key    *ari.Key
	filter *ari.ApplicationEventFilter
}

func (a *ari23Application) FilterEvents(key *ari.Key, filter *ari.ApplicationEventFilter) (*ari.ApplicationData, error) {
	a.key, a.filter = key, filter
	return &ari.ApplicationData{Name: key.ID, EventsAllowed: []ari.EventTypeFilter{{Type: "StasisStart"}}}, nil
}

type ari23Asterisk struct {
	ari.Asterisk
	key  *ari.Key
	only string
}

func (a *ari23Asterisk) Ping(key *ari.Key) (*ari.AsteriskPing, error) {
	a.key = key
	return &ari.AsteriskPing{AsteriskID: "node-1", Ping: "pong"}, nil
}

func (a *ari23Asterisk) InfoWithOptions(key *ari.Key, opts ari.AsteriskInfoOptions) (*ari.AsteriskInfo, error) {
	a.key, a.only = key, opts.Only
	return &ari.AsteriskInfo{SystemInfo: ari.SystemInfo{Version: "23.0.0"}}, nil
}

func TestServerForwardsAsteriskInfoOnly(t *testing.T) {
	asterisk := &ari23Asterisk{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = ari23ServerClient{asterisk: asterisk}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1"}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "AsteriskInfo", Key: key, AsteriskInfoOptions: &ari.AsteriskInfoOptions{Only: "build,system+status"}})
	if asterisk.key != key || asterisk.only != "build,system+status" || bus.response == nil || bus.response.Data == nil || bus.response.Data.Asterisk == nil || bus.response.Data.Asterisk.SystemInfo.Version != "23.0.0" {
		t.Fatalf("forwarded only=%q response=%+v", asterisk.only, bus.response)
	}
}

func TestServerForwardsAsterisk23ApplicationAndPing(t *testing.T) {
	application := &ari23Application{}
	asterisk := &ari23Asterisk{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = ari23ServerClient{application: application, asterisk: asterisk}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ApplicationKey, ID: "demo"}
	empty := []ari.EventTypeFilter{}
	filter := &ari.ApplicationEventFilter{Allowed: &empty}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{
		Kind: "ApplicationFilterEvents", Key: key, ApplicationEventFilter: filter,
	})
	if application.key != key || !reflect.DeepEqual(application.filter, filter) || bus.response == nil ||
		bus.response.Data == nil || bus.response.Data.Application == nil || bus.response.Data.Application.Name != "demo" {
		t.Fatalf("filter forwarding: key=%+v filter=%+v response=%+v", application.key, application.filter, bus.response)
	}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "AsteriskPing", Key: key})
	if asterisk.key != key || bus.response == nil || bus.response.Data == nil || bus.response.Data.AsteriskPing == nil || bus.response.Data.AsteriskPing.Ping != "pong" {
		t.Fatalf("ping forwarding: key=%+v response=%+v", asterisk.key, bus.response)
	}
}
