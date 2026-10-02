package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type bridgeVariableClient struct {
	ari.Client
	bridge ari.Bridge
}

func (c bridgeVariableClient) Bridge() ari.Bridge { return c.bridge }

type bridgeVariableRecorder struct {
	ari.Bridge
	key         *ari.Key
	name, value string
	report      *bool
	names       []string
	assignments map[string]ari.BridgeVariableAssignment
}

func (b *bridgeVariableRecorder) GetVariable(key *ari.Key, name string) (string, error) {
	b.key, b.name = key, name
	return "Waiting", nil
}
func (b *bridgeVariableRecorder) SetVariable(key *ari.Key, name, value string, report *bool) error {
	b.key, b.name, b.value, b.report = key, name, value, report
	return nil
}
func (b *bridgeVariableRecorder) GetVariables(key *ari.Key, names ...string) (map[string]any, error) {
	b.key, b.names = key, names
	return map[string]any{"BRIDGE_STATE": "Waiting"}, nil
}
func (b *bridgeVariableRecorder) SetVariables(key *ari.Key, values map[string]ari.BridgeVariableAssignment) error {
	b.key, b.assignments = key, values
	return nil
}

func TestServerForwardsBridgeVariables(t *testing.T) {
	recorder := &bridgeVariableRecorder{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = bridgeVariableClient{bridge: recorder}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"}
	report := false
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "BridgeVariableGet", Key: key,
		BridgeVariable: &proxy.BridgeVariable{Name: "BRIDGE+STATE"}})
	if recorder.key != key || recorder.name != "BRIDGE+STATE" || bus.response == nil || bus.response.Data == nil || bus.response.Data.Variable != "Waiting" {
		t.Fatalf("single read: recorder=%+v response=%+v", recorder, bus.response)
	}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "BridgeVariableSet", Key: key,
		BridgeVariable: &proxy.BridgeVariable{Name: "BRIDGE+STATE", Value: "", ReportEvents: &report}})
	if recorder.report == nil || *recorder.report || recorder.value != "" || bus.response == nil || bus.response.Err() != nil {
		t.Fatalf("single write: recorder=%+v response=%+v", recorder, bus.response)
	}
	names := []string{"BRIDGE+STATE", "SUPPORT_LEVEL"}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "BridgeVariablesGet", Key: key,
		BridgeVariables: &proxy.BridgeVariables{Names: names}})
	if !reflect.DeepEqual(recorder.names, names) || bus.response == nil || bus.response.Data == nil || bus.response.Data.Variables["BRIDGE_STATE"] != "Waiting" {
		t.Fatalf("bulk read: recorder=%+v response=%+v", recorder, bus.response)
	}
	assignments := map[string]ari.BridgeVariableAssignment{"BRIDGE_STATE": {Value: "Waiting"}, "SUPPORT_LEVEL": {Value: "Premium", ReportEvents: &report}}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "BridgeVariablesSet", Key: key,
		BridgeVariables: &proxy.BridgeVariables{Values: assignments}})
	if !reflect.DeepEqual(recorder.assignments, assignments) || bus.response == nil || bus.response.Err() != nil {
		t.Fatalf("bulk write: recorder=%+v response=%+v", recorder, bus.response)
	}
}
