package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type claimRequestBus struct {
	messagebus.Client
	request  *proxy.Request
	response *proxy.Response
}

func (b *claimRequestBus) Request(_ string, req *proxy.Request) (*proxy.Response, error) {
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	b.request = new(proxy.Request)
	if err := json.Unmarshal(encoded, b.request); err != nil {
		return nil, err
	}
	encoded, err = json.Marshal(b.response)
	if err != nil {
		return nil, err
	}
	var response proxy.Response
	if err := json.Unmarshal(encoded, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func TestClaimChannelProxyRequestAndStatus(t *testing.T) {
	bus := &claimRequestBus{response: &proxy.Response{}}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ApplicationKey, ID: "demo"}
	if err := c.Application().ClaimChannel(key, "channel+1"); err != nil {
		t.Fatal(err)
	}
	if bus.request.Kind != "EventClaimChannel" || bus.request.Key.ID != key.ID || bus.request.Key.App != key.App || bus.request.Key.Node != key.Node || bus.request.EventClaim == nil || bus.request.EventClaim.ChannelID != "channel+1" {
		t.Fatalf("request = %+v", bus.request)
	}
	if err := c.Application().ClaimChannel(ari.NewKey(ari.ApplicationKey, "demo"), "channel+1"); err == nil {
		t.Fatal("expected target node requirement")
	}
	bus.response = &proxy.Response{Error: "already claimed", StatusCode: http.StatusConflict}
	err := c.Application().ClaimChannel(key, "channel+1")
	var coded interface{ Code() int }
	if !errors.As(err, &coded) || coded.Code() != http.StatusConflict {
		t.Fatalf("error = %v", err)
	}
}
