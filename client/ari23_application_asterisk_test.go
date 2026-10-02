package client

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/client/cluster"
	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type ari23RequestBus struct {
	messagebus.Client
	request *proxy.Request
}

func (b *ari23RequestBus) respond(request *proxy.Request) (*proxy.Response, error) {
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var decoded proxy.Request
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, err
	}
	b.request = &decoded
	response := &proxy.Response{Data: &proxy.EntityData{
		Application:  &ari.ApplicationData{Name: "demo", EventsAllowed: []ari.EventTypeFilter{{Type: "StasisStart"}}, EventsDisallowed: []ari.EventTypeFilter{}},
		AsteriskPing: &ari.AsteriskPing{AsteriskID: "node-1", Ping: "pong", Timestamp: "2026-09-25T12:00:00Z"},
		Asterisk:     &ari.AsteriskInfo{SystemInfo: ari.SystemInfo{Version: "23.0.0"}},
		Variable:     "Waiting",
		Variables:    map[string]any{"BRIDGE_STATE": "Waiting", "SUPPORT_LEVEL": "Premium"},
		RTPStats:     &ari.RTPStats{TxCount: 20, RxCount: 19, RTT: 2.5, ChannelUniqueID: "channel-1"},
	}}
	encoded, err = json.Marshal(response)
	if err != nil {
		return nil, err
	}
	var delivered proxy.Response
	if err := json.Unmarshal(encoded, &delivered); err != nil {
		return nil, err
	}
	return &delivered, nil
}

func TestClientAsteriskInfoOnlySurvivesMessageBus(t *testing.T) {
	bus := &ari23RequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	key := &ari.Key{App: "demo", Node: "node-1"}
	info, err := c.Asterisk().InfoWithOptions(key, ari.AsteriskInfoOptions{Only: "build,system+status"})
	if err != nil || info == nil || info.SystemInfo.Version != "23.0.0" || bus.request.AsteriskInfoOptions == nil || bus.request.AsteriskInfoOptions.Only != "build,system+status" {
		t.Fatalf("info=%+v request=%+v error=%v", info, bus.request, err)
	}
}

func TestClientQueryWritesSurviveMessageBus(t *testing.T) {
	bus := &ari23RequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	appKey := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ApplicationKey, ID: "demo"}
	if err := c.Application().Subscribe(appKey, "channel:call+leg/2"); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "ApplicationSubscribe" || bus.request.ApplicationSubscribe.EventSource != "channel:call+leg/2" {
		t.Fatalf("subscribe request=%+v", bus.request)
	}
	variableKey := appKey.New(ari.VariableKey, "CUSTOM+NAME")
	for _, value := range []string{"", "a=b&c=d"} {
		if err := c.Asterisk().Variables().Set(variableKey, value); err != nil {
			t.Fatal(err)
		}
		if bus.request.Kind != "AsteriskVariableSet" || bus.request.Key.ID != "CUSTOM+NAME" || bus.request.AsteriskVariableSet.Value != value {
			t.Fatalf("variable request=%+v want value=%q", bus.request, value)
		}
	}
}

func TestClientBridgeVariablesSurviveMessageBusEncoding(t *testing.T) {
	bus := &ari23RequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"}
	report := false
	value, err := c.Bridge().GetVariable(key, "BRIDGE+STATE")
	if err != nil || value != "Waiting" || bus.request.Kind != "BridgeVariableGet" || bus.request.BridgeVariable.Name != "BRIDGE+STATE" {
		t.Fatalf("single read: value=%q request=%+v error=%v", value, bus.request, err)
	}
	if err := c.Bridge().SetVariable(key, "BRIDGE+STATE", "", &report); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "BridgeVariableSet" || bus.request.BridgeVariable.ReportEvents == nil || *bus.request.BridgeVariable.ReportEvents {
		t.Fatalf("single write request = %+v", bus.request)
	}
	variables, err := c.Bridge().GetVariables(key, "BRIDGE+STATE", "SUPPORT_LEVEL")
	if err != nil || len(variables) != 2 || !reflect.DeepEqual(bus.request.BridgeVariables.Names, []string{"BRIDGE+STATE", "SUPPORT_LEVEL"}) {
		t.Fatalf("bulk read: values=%+v request=%+v error=%v", variables, bus.request, err)
	}
	assignments := map[string]ari.BridgeVariableAssignment{
		"BRIDGE_STATE":  {Value: "Waiting"},
		"SUPPORT_LEVEL": {Value: "Premium", ReportEvents: &report},
	}
	if err := c.Bridge().SetVariables(key, assignments); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "BridgeVariablesSet" || !reflect.DeepEqual(bus.request.BridgeVariables.Values, assignments) {
		t.Fatalf("bulk write request = %+v, want %+v", bus.request, assignments)
	}
}

func (b *ari23RequestBus) Request(_ string, request *proxy.Request) (*proxy.Response, error) {
	return b.respond(request)
}

func (b *ari23RequestBus) MultipleRequestReturnFirstGoodResponse(_ string, request *proxy.Request, _ int) (*proxy.Response, error) {
	return b.respond(request)
}

func TestClientAsterisk23RequestsSurviveMessageBusEncoding(t *testing.T) {
	bus := &ari23RequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test.", cluster: cluster.New(), clusterMaxAge: time.Minute}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ApplicationKey, ID: "demo"}
	empty := []ari.EventTypeFilter{}
	blocked := []ari.EventTypeFilter{{Type: "ChannelDestroyed"}}
	for _, test := range []struct {
		name   string
		filter *ari.ApplicationEventFilter
	}{
		{"clear both", nil},
		{"clear allowed", &ari.ApplicationEventFilter{Allowed: &empty}},
		{"set disallowed", &ari.ApplicationEventFilter{Disallowed: &blocked}},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := c.Application().FilterEvents(key, test.filter)
			if err != nil {
				t.Fatal(err)
			}
			if bus.request == nil || bus.request.Kind != "ApplicationFilterEvents" || !reflect.DeepEqual(bus.request.ApplicationEventFilter, test.filter) {
				t.Fatalf("application request = %+v", bus.request)
			}
			if data == nil || data.Name != "demo" || len(data.EventsAllowed) != 1 || data.EventsDisallowed == nil {
				t.Fatalf("application response = %+v", data)
			}
		})
	}
	ping, err := c.Asterisk().Ping(key)
	if err != nil {
		t.Fatal(err)
	}
	if bus.request == nil || bus.request.Kind != "AsteriskPing" || ping == nil || ping.Ping != "pong" || ping.AsteriskID != "node-1" {
		t.Fatalf("ping request = %+v, response = %+v", bus.request, ping)
	}
	if bus.request.Key == nil || bus.request.Key.Node != "node-1" {
		t.Fatalf("targeted ping key = %+v", bus.request.Key)
	}
	if _, err := c.Asterisk().Ping(nil); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "AsteriskPing" || bus.request.Key == nil || bus.request.Key.Node != "" {
		t.Fatalf("broadcast ping request = %+v", bus.request)
	}
}
