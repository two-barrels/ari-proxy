package server

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type textMessageClient struct {
	ari.Client
	message ari.TextMessage
}

func (c textMessageClient) TextMessage() ari.TextMessage { return c.message }

type textMessageRecorder struct {
	ari.TextMessage
	called         string
	key            *ari.Key
	from, to, body string
	vars           map[string]string
	err            error
}

func (r *textMessageRecorder) SendWithKey(key *ari.Key, from, body string, vars map[string]string) error {
	r.called, r.key, r.from, r.to, r.body, r.vars = "endpoint", key, from, "", body, vars
	return r.err
}

func (r *textMessageRecorder) SendByURIWithKey(key *ari.Key, from, to, body string, vars map[string]string) error {
	r.called, r.key, r.from, r.to, r.body, r.vars = "URI", key, from, to, body, vars
	return r.err
}

func TestTextMessageServerDispatch(t *testing.T) {
	recorder := &textMessageRecorder{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = textMessageClient{message: recorder}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.EndpointKey, ID: "PJSIP/alice"}
	vars := map[string]string{"X-Trace": "a/b"}
	for _, test := range []struct{ kind, called, to string }{{"TextMessageSend", "endpoint", ""}, {"TextMessageSendByURI", "URI", "PJSIP/bob"}} {
		t.Run(test.kind, func(t *testing.T) {
			bus.response = nil
			s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: test.kind, Key: key,
				TextMessageSend: &proxy.TextMessageSend{From: "operator", To: test.to, Body: "hello", Variables: vars}})
			if recorder.called != test.called || recorder.key != key || recorder.from != "operator" || recorder.to != test.to || recorder.body != "hello" || !reflect.DeepEqual(recorder.vars, vars) {
				t.Errorf("forwarded: %+v", recorder)
			}
			if bus.response == nil || bus.response.Err() != nil {
				t.Fatalf("response=%+v", bus.response)
			}
		})
	}
	recorder.err = &proxy.StatusError{Message: "endpoint missing", Status: http.StatusNotFound}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: "TextMessageSend", Key: key,
		TextMessageSend: &proxy.TextMessageSend{From: "operator", Body: "hello"}})
	if bus.response == nil || bus.response.StatusCode != http.StatusNotFound {
		t.Fatalf("error response=%+v", bus.response)
	}
}
