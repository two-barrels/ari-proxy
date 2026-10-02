package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	clientbus "github.com/two-barrels/ari-proxy/v6/client/bus"
	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/stdbus"
	"github.com/inconshreveable/log15"
)

// These fakes replace the network transport only. Event decoding, routing,
// serialization, and client-side decoding use the production implementations.
type eventTestClient struct {
	ari.Client
	bus ari.Bus
}

func (c eventTestClient) Bus() ari.Bus { return c.bus }

type eventTestBus struct {
	ari.Bus
	ready chan struct{}
}

func (b eventTestBus) Subscribe(key *ari.Key, types ...string) ari.Subscription {
	sub := b.Bus.Subscribe(key, types...)
	close(b.ready)
	return sub
}

type eventTestPublication struct {
	subject string
	data    []byte
}

type eventTestPublisher struct {
	messagebus.Server
	publications chan eventTestPublication
	deliver      func(string, []byte)
}

func (p *eventTestPublisher) PublishEvent(subject string, event ari.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	p.publications <- eventTestPublication{subject: subject, data: data}
	if p.deliver != nil {
		p.deliver(subject, data)
	}
	return nil
}

type eventTestSubscription struct{}

func (eventTestSubscription) Unsubscribe() error { return nil }

type eventTestReceiver struct {
	messagebus.Client
	callback messagebus.EventHandler
}

func (r *eventTestReceiver) GetWildcardString(messagebus.WildcardType) string { return ">" }

func (r *eventTestReceiver) SubscribeEvent(_ string, _ string, callback messagebus.EventHandler) (messagebus.Subscription, error) {
	r.callback = callback
	return eventTestSubscription{}, nil
}

type eventFidelityHarness struct {
	source       ari.Bus
	server       *Server
	clientEvents ari.Subscription
	publications chan eventTestPublication
}

func newEventFidelityHarness(t *testing.T) *eventFidelityHarness {
	t.Helper()

	source := stdbus.New()
	ready := make(chan struct{})
	publications := make(chan eventTestPublication, 16)
	receiver := &eventTestReceiver{}
	clientEvents := clientbus.New("test.", receiver, log15.New()).Subscribe(nil, ari.Events.All)
	if clientEvents == nil {
		t.Fatal("failed to subscribe proxy client to events")
	}

	publisher := &eventTestPublisher{publications: publications}
	publisher.deliver = func(subject string, data []byte) {
		if subject == "test.event.demo.node-1" {
			receiver.callback(data)
		}
	}

	server := New()
	server.Application = "demo"
	server.AsteriskID = "node-1"
	server.MBPrefix = "test."
	server.ari = eventTestClient{bus: eventTestBus{Bus: source, ready: ready}}
	server.mbus = publisher

	ctx, stop := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		server.runEventHandler(ctx)
	}()
	select {
	case <-ready:
	case <-time.After(time.Second):
		stop()
		t.Fatal("proxy did not subscribe to source events")
	}
	t.Cleanup(func() {
		stop()
		<-done
		clientEvents.Cancel()
	})

	return &eventFidelityHarness{source: source, server: server, clientEvents: clientEvents, publications: publications}
}

func (h *eventFidelityHarness) sendRaw(t *testing.T, raw string) {
	t.Helper()
	event, err := ari.DecodeEvent([]byte(raw))
	if err != nil {
		t.Fatalf("ARI decoder rejected raw event: %v", err)
	}
	h.source.Send(event)
}

func (h *eventFidelityHarness) receiveClientEvent(t *testing.T) ari.Event {
	t.Helper()
	select {
	case event := <-h.clientEvents.Events():
		return event
	case <-time.After(250 * time.Millisecond):
		t.Fatal("raw event did not reach the proxy client")
		return nil
	}
}

func TestEventRoundTripPreservesKnownAndFutureFields(t *testing.T) {
	h := newEventFidelityHarness(t)
	raw := `{"type":"StasisStart","application":"demo","asterisk_id":"node-1","args":["first"],"channel":{"id":"channel-1","name":"PJSIP/1","state":"Up","creationtime":"2025-01-01T00:00:00.000+0000","future_channel_field":{"nested":[1,"two"]}},"future_event_field":{"enabled":true}}`
	h.sendRaw(t, raw)

	event := h.receiveClientEvent(t)
	publication := <-h.publications
	if publication.subject != "test.event.demo.node-1" {
		t.Fatalf("first publication subject = %q, want canonical event subject", publication.subject)
	}
	t.Run("published", func(t *testing.T) {
		assertJSONFieldsPreserved(t, []byte(raw), publication.data)
	})
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("client", func(t *testing.T) {
		assertJSONFieldsPreserved(t, []byte(raw), encoded)
	})
}

func TestEventRoundTripSupportsAsterisk23Types(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want ari.Event
	}{
		{"ApplicationRegistered", `{"type":"ApplicationRegistered","application":"demo","asterisk_id":"node-1"}`, &ari.ApplicationRegistered{}},
		{"ApplicationUnregistered", `{"type":"ApplicationUnregistered","application":"demo","asterisk_id":"node-1"}`, &ari.ApplicationUnregistered{}},
		{"CallBroadcast", `{"type":"CallBroadcast","application":"demo","asterisk_id":"node-1","channel":{"id":"channel-1"},"caller":"100","called":"200"}`, &ari.CallBroadcast{}},
		{"CallClaimed", `{"type":"CallClaimed","application":"demo","asterisk_id":"node-1","channel":{"id":"channel-1"},"winner_app":"winner"}`, &ari.CallClaimed{}},
		{"ChannelToneDetected", `{"type":"ChannelToneDetected","application":"demo","asterisk_id":"node-1","channel":{"id":"channel-1"},"new_detail":{"value":42}}`, &ari.ChannelToneDetected{}},
		{"ChannelTransfer", `{"type":"ChannelTransfer","application":"demo","asterisk_id":"node-1","refer_to":{"requested_destination":{"destination":"200","future_destination":{"token":7}},"destination_channel":{"id":"channel-2"}},"referred_by":{"source_channel":{"id":"channel-1"}}}`, &ari.ChannelTransfer{}},
		{"RESTResponse", `{"type":"RESTResponse","application":"demo","asterisk_id":"node-1","transaction_id":"tx-1","request_id":"req-1","status_code":200,"reason_phrase":"OK","uri":"/channels"}`, &ari.RESTResponse{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newEventFidelityHarness(t)
			h.sendRaw(t, test.raw)
			event := h.receiveClientEvent(t)
			if reflect.TypeOf(event) != reflect.TypeOf(test.want) {
				t.Fatalf("proxy client received %T, want %T", event, test.want)
			}
			encoded, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONFieldsPreserved(t, []byte(test.raw), encoded)
		})
	}
}

func TestEventRoundTripAcceptsUnknownFutureType(t *testing.T) {
	h := newEventFidelityHarness(t)
	raw := `{"type":"FutureChannelEvent","application":"demo","asterisk_id":"node-1","channel":{"id":"channel-1"},"new_detail":{"value":42}}`
	h.sendRaw(t, raw)
	event := h.receiveClientEvent(t)
	if _, ok := event.(*ari.UnknownEvent); !ok {
		t.Fatalf("proxy client received %T, want *ari.UnknownEvent", event)
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	assertJSONFieldsPreserved(t, []byte(raw), encoded)
}

func TestEventRoundTripAcceptsEventWithoutResourceKey(t *testing.T) {
	h := newEventFidelityHarness(t)
	raw := `{"type":"ApplicationReplaced","application":"demo","asterisk_id":"node-1"}`
	h.sendRaw(t, raw)
	h.receiveClientEvent(t)
}

func TestEventRoundTripPublishesEachDialogOnce(t *testing.T) {
	h := newEventFidelityHarness(t)

	// Bind the same dialog to two resources named in one event, as happens
	// when a channel enters a bridge. Two different dialogs must each get one copy.
	h.server.Dialog.Bind("dialog-a", ari.ChannelKey, "channel-1")
	h.server.Dialog.Bind("dialog-a", ari.BridgeKey, "bridge-1")
	h.server.Dialog.Bind("dialog-b", ari.ChannelKey, "channel-1")

	raw := `{"type":"ChannelEnteredBridge","application":"demo","asterisk_id":"node-1","channel":{"id":"channel-1"},"bridge":{"id":"bridge-1","channels":["channel-1"]}}`
	h.sendRaw(t, raw)

	counts := map[string]int{}
	deadline := time.NewTimer(100 * time.Millisecond)
	defer deadline.Stop()
	for collecting := true; collecting; {
		var publication eventTestPublication
		select {
		case publication = <-h.publications:
		case <-deadline.C:
			collecting = false
			continue
		}
		if publication.subject == "test.event.demo.node-1" {
			counts[publication.subject]++
			continue
		}
		var body struct {
			Dialog string `json:"dialog"`
		}
		if err := json.Unmarshal(publication.data, &body); err != nil {
			t.Fatal(err)
		}
		if body.Dialog != publication.subject[len("test.dialogevent."):] {
			t.Errorf("subject %q carries dialog %q", publication.subject, body.Dialog)
		}
		counts[publication.subject]++
	}
	want := map[string]int{"test.event.demo.node-1": 1, "test.dialogevent.dialog-a": 1, "test.dialogevent.dialog-b": 1}
	if !reflect.DeepEqual(counts, want) {
		t.Errorf("dialog publications = %v, want %v", counts, want)
	}
}

func assertJSONFieldsPreserved(t *testing.T, original, delivered []byte) {
	t.Helper()
	var want, got any
	if err := json.Unmarshal(original, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(delivered, &got); err != nil {
		t.Fatal(err)
	}
	assertJSONSubset(t, "$", want, got)
}

func assertJSONSubset(t *testing.T, path string, want, got any) {
	t.Helper()
	switch expected := want.(type) {
	case map[string]any:
		actual, ok := got.(map[string]any)
		if !ok {
			t.Errorf("%s: got %T, want object", path, got)
			return
		}
		for key, value := range expected {
			assertJSONSubset(t, path+"."+key, value, actual[key])
		}
	case []any:
		actual, ok := got.([]any)
		if !ok || len(actual) != len(expected) {
			t.Errorf("%s: got %v, want %v", path, got, want)
			return
		}
		for index, value := range expected {
			assertJSONSubset(t, fmt.Sprintf("%s[%d]", path, index), value, actual[index])
		}
	default:
		if !reflect.DeepEqual(want, got) {
			t.Errorf("%s: got %v, want %v", path, got, want)
		}
	}
}
