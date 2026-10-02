// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type responseContractBus struct {
	messagebus.Client
	response *proxy.Response
	request  *proxy.Request
}

func (b *responseContractBus) Request(_ string, request *proxy.Request) (*proxy.Response, error) {
	b.request = request
	data, err := json.Marshal(b.response)
	if err != nil {
		return nil, err
	}
	var response proxy.Response
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func TestProxyResponseContractWire(t *testing.T) {
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}
	fixtures := []struct {
		name, kind string
		response   *proxy.Response
		wantCode   int
		wantID     string
		wantError  bool
		call       func(*Client) (string, error)
	}{
		{"bridge create assigned ID", "BridgeCreateWithoutID", &proxy.Response{Key: key.New(ari.BridgeKey, "assigned-bridge")}, 0, "assigned-bridge", false,
			func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithoutID(key, ari.BridgeCreateOptions{Type: "mixing"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge create missing key", "BridgeCreateWithoutID", &proxy.Response{}, 0, "", true,
			func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithoutID(key, ari.BridgeCreateOptions{Type: "mixing"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge create conflict", "BridgeCreateWithoutID", proxy.NewErrorResponse(&proxy.StatusError{Status: http.StatusConflict, Message: "bridge already exists"}), 409, "", true,
			func(c *Client) (string, error) {
				h, err := c.Bridge().CreateWithoutID(key, ari.BridgeCreateOptions{Type: "mixing"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"originate invalid options", "ChannelOriginateWithID", proxy.NewErrorResponse(&proxy.StatusError{Status: http.StatusBadRequest, Message: "invalid endpoint"}), 400, "", true,
			func(c *Client) (string, error) {
				h, err := c.Channel().OriginateWithID(key, ari.OriginateRequest{ChannelID: "channel-9", Endpoint: "PJSIP/alice", App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"bridge play not found", "BridgePlayWithoutID", proxy.NewErrorResponse(&proxy.StatusError{Status: http.StatusNotFound, Message: "Bridge not found"}), 404, "", true,
			func(c *Client) (string, error) {
				h, err := c.Bridge().PlayWithoutID(key.New(ari.BridgeKey, "bridge-1"), ari.BridgePlayOptions{Media: []string{"sound:one"}})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"snoop conflict", "ChannelSnoopWithoutID", proxy.NewErrorResponse(&proxy.StatusError{Status: http.StatusConflict, Message: "Channel already exists"}), 409, "", true,
			func(c *Client) (string, error) {
				h, err := c.Channel().SnoopWithoutID(key, &ari.SnoopOptions{App: "demo"})
				if err != nil {
					return "", err
				}
				return h.ID(), nil
			}},
		{"answer empty success", "ChannelAnswer", &proxy.Response{}, 0, "", false,
			func(c *Client) (string, error) { return "", c.Channel().Answer(key) }},
		{"answer not found", "ChannelAnswer", proxy.NewErrorResponse(&proxy.StatusError{Status: http.StatusNotFound, Message: "Channel not found"}), 404, "", true,
			func(c *Client) (string, error) { return "", c.Channel().Answer(key) }},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			bus := &responseContractBus{response: fixture.response}
			logger := log15.New()
			logger.SetHandler(log15.DiscardHandler())
			client := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
			id, err := fixture.call(client)
			if bus.request == nil || bus.request.Kind != fixture.kind {
				t.Fatalf("request=%+v, want kind %s", bus.request, fixture.kind)
			}
			var coded interface{ Code() int }
			code := 0
			if errors.As(err, &coded) {
				code = coded.Code()
			}
			if id != fixture.wantID || (err != nil) != fixture.wantError || code != fixture.wantCode {
				t.Fatalf("id=%q error=%v code=%d, want id=%q error=%t code=%d", id, err, code, fixture.wantID, fixture.wantError, fixture.wantCode)
			}
			if fixture.response.Error != "" && !strings.Contains(err.Error(), fixture.response.Error) {
				t.Errorf("error %q omits %q", err, fixture.response.Error)
			}
		})
	}
}
