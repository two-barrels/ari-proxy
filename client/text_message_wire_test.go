package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/client/cluster"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

func TestTextMessageProxyRequests(t *testing.T) {
	bus := &ari23RequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	cl := cluster.New()
	cl.Update("node-1", "demo")
	c := &Client{appName: "demo", core: &core{mbus: bus, log: logger, prefix: "test.", cluster: cl, clusterMaxAge: time.Minute}}
	if c.TextMessage() == nil {
		t.Fatal("text message accessor is nil")
	}
	vars := map[string]string{"X-Trace": "a/b", "display_name": "A + B"}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.EndpointKey, ID: "PJSIP/alice"}
	if err := c.TextMessage().SendWithKey(key, "sip:operator@example.com", "hello & goodbye", vars); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "TextMessageSend" || bus.request.Key.ID != key.ID || bus.request.Key.Node != key.Node || bus.request.TextMessageSend.From != "sip:operator@example.com" || bus.request.TextMessageSend.Body != "hello & goodbye" || !reflect.DeepEqual(bus.request.TextMessageSend.Variables, vars) {
		t.Fatalf("keyed endpoint request=%+v", bus.request)
	}
	if err := c.TextMessage().SendByURIWithKey(ari.NodeKey("demo", "node-1"), "operator", "PJSIP/bob?x=1&y=2", "hello", nil); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "TextMessageSendByURI" || bus.request.TextMessageSend.To != "PJSIP/bob?x=1&y=2" || bus.request.TextMessageSend.Variables == nil {
		t.Fatalf("keyed URI request=%+v", bus.request)
	}
	if err := c.TextMessage().Send("operator", "PJSIP", "alice", "hello", vars); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "TextMessageSend" || bus.request.Key.Node != "node-1" || bus.request.Key.ID != "PJSIP/alice" {
		t.Fatalf("single-node endpoint request=%+v", bus.request)
	}
	if err := c.TextMessage().SendByURI("operator", "PJSIP/bob", "hello", nil); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "TextMessageSendByURI" || bus.request.Key.Node != "node-1" {
		t.Fatalf("single-node URI request=%+v", bus.request)
	}
	cl.Update("node-2", "demo")
	if err := c.TextMessage().Send("operator", "PJSIP", "alice", "hello", nil); err == nil {
		t.Fatal("expected ambiguous target error")
	}
	if err := c.TextMessage().SendByURI("operator", "PJSIP/bob", "hello", nil); err == nil {
		t.Fatal("expected ambiguous URI target error")
	}
}
