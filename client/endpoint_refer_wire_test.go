// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

func TestEndpointReferProxyRequests(t *testing.T) {
	bus := &ari23RequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.EndpointKey, ID: "PJSIP/alice"}
	falseValue := false
	opts := ari.EndpointReferOptions{To: "PJSIP/alice+1", From: "sip:operator@example.com", ReferTo: "PJSIP/bob?x=1&y=2", ToSelf: &falseValue, Variables: map[string]string{"display_name": "A + B"}}
	if err := c.Endpoint().Refer(key, opts); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "EndpointRefer" || bus.request.Key.ID != key.ID || bus.request.EndpointRefer == nil || !reflect.DeepEqual(*bus.request.EndpointRefer, opts) {
		t.Fatalf("URI refer request = %+v", bus.request)
	}
	if err := c.Endpoint().ReferToEndpoint(key, opts); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "EndpointReferToEndpoint" || bus.request.Key.ID != key.ID || bus.request.EndpointRefer == nil || !reflect.DeepEqual(*bus.request.EndpointRefer, opts) {
		t.Fatalf("endpoint refer request = %+v", bus.request)
	}
	if err := c.Endpoint().Refer(nil, opts); err == nil {
		t.Fatal("expected target node requirement for URI REFER")
	}
	if err := c.Endpoint().ReferToEndpoint(ari.NewEndpointKey("PJSIP", "alice"), opts); err == nil {
		t.Fatal("expected target node requirement for endpoint REFER")
	}
}
