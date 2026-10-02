package client

import (
	"net/http"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/testfixtures"
	"github.com/inconshreveable/log15"
)

func TestVersionedProxyResponses(t *testing.T) {
	for _, fixture := range testfixtures.Asterisk20_22_23 {
		if fixture.Path != "/bridges/{bridgeId}/variable" && fixture.Path != "/channels/{channelId}/progress" {
			continue
		}
		t.Run(fixture.Path+"@"+fixture.Version, func(t *testing.T) {
			response := &proxy.Response{}
			if fixture.Available && fixture.Method == http.MethodGet {
				response.Data = &proxy.EntityData{Variable: "enabled"}
			} else if !fixture.Available {
				response = proxy.NewErrorResponse(&proxy.StatusError{Status: http.StatusNotFound, Message: "Not Found"})
			}
			bus := &responseContractBus{response: response}
			logger := log15.New()
			logger.SetHandler(log15.DiscardHandler())
			client := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
			key := &ari.Key{App: "demo", Node: "node-1", ID: "bridge-1", Kind: ari.BridgeKey}
			var err error
			if fixture.Method == http.MethodGet {
				var value string
				value, err = client.Bridge().GetVariable(key, "TEST")
				if bus.request == nil || bus.request.Kind != "BridgeVariableGet" {
					t.Fatalf("request=%+v, want BridgeVariableGet", bus.request)
				}
				if fixture.Available && value != "enabled" {
					t.Errorf("value=%q, want enabled", value)
				}
			} else {
				key.Kind, key.ID = ari.ChannelKey, "channel-1"
				err = client.Channel().Progress(key)
				if bus.request == nil || bus.request.Kind != "ChannelProgress" {
					t.Fatalf("request=%+v, want ChannelProgress", bus.request)
				}
			}
			if fixture.Available && err != nil || !fixture.Available && proxyCode(err) != http.StatusNotFound {
				t.Errorf("available=%t, error=%v, code=%d", fixture.Available, err, proxyCode(err))
			}
		})
	}
}

func proxyCode(err error) int {
	if coded, ok := err.(interface{ Code() int }); ok {
		return coded.Code()
	}
	return 0
}
