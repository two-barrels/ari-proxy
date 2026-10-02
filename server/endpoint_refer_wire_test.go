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

type endpointReferClient struct {
	ari.Client
	endpoint ari.Endpoint
}

func (c endpointReferClient) Endpoint() ari.Endpoint { return c.endpoint }

type endpointReferRecorder struct {
	ari.Endpoint
	called string
	key    *ari.Key
	opts   ari.EndpointReferOptions
}

func (r *endpointReferRecorder) Refer(key *ari.Key, opts ari.EndpointReferOptions) error {
	r.called, r.key, r.opts = "URI", key, opts
	return nil
}

func (r *endpointReferRecorder) ReferToEndpoint(key *ari.Key, opts ari.EndpointReferOptions) error {
	r.called, r.key, r.opts = "endpoint", key, opts
	return nil
}

func TestEndpointReferServerDispatch(t *testing.T) {
	recorder := &endpointReferRecorder{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = endpointReferClient{endpoint: recorder}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.EndpointKey, ID: "PJSIP/alice"}
	falseValue := false
	opts := ari.EndpointReferOptions{To: "PJSIP/alice+1", From: "operator", ReferTo: "PJSIP/bob", ToSelf: &falseValue, Variables: map[string]string{"display_name": "A + B"}}
	for _, test := range []struct{ kind, called string }{{"EndpointRefer", "URI"}, {"EndpointReferToEndpoint", "endpoint"}} {
		t.Run(test.kind, func(t *testing.T) {
			bus.response = nil
			s.dispatchRequest(context.Background(), "reply", &proxy.Request{Kind: test.kind, Key: key, EndpointRefer: &opts})
			if recorder.called != test.called || recorder.key != key || !reflect.DeepEqual(recorder.opts, opts) {
				t.Errorf("forwarded: called=%q key=%v opts=%+v", recorder.called, recorder.key, recorder.opts)
			}
			if bus.response == nil || bus.response.Err() != nil {
				t.Fatalf("response=%+v", bus.response)
			}
		})
	}
}
