package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type responseErrorBridge struct{ ari.Bridge }

func (b responseErrorBridge) CreateWithoutID(*ari.Key, ari.BridgeCreateOptions) (*ari.BridgeHandle, error) {
	return nil, &proxy.StatusError{Status: http.StatusConflict, Message: "bridge already exists"}
}

func (b responseErrorBridge) PlayWithoutID(*ari.Key, ari.BridgePlayOptions) (*ari.PlaybackHandle, error) {
	return nil, &proxy.StatusError{Status: http.StatusNotFound, Message: "Bridge not found"}
}

func TestServerResponseContractWire(t *testing.T) {
	fixtures := []struct {
		name, kind string
		status     int
		message    string
		request    *proxy.Request
	}{
		{"create conflict", "BridgeCreateWithoutID", http.StatusConflict, "bridge already exists", &proxy.Request{
			Kind: "BridgeCreateWithoutID", Key: &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey},
			BridgeCreate: &proxy.BridgeCreate{Type: "mixing"},
		}},
		{"play missing bridge", "BridgePlayWithoutID", http.StatusNotFound, "Bridge not found", &proxy.Request{
			Kind: "BridgePlayWithoutID", Key: &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"},
			BridgePlay: &proxy.BridgePlay{Options: &ari.BridgePlayOptions{Media: []string{"sound:one"}}},
		}},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			bus := &bridgeOptionsResponseBus{}
			server := New()
			server.ari = routeForwardingClient{bridge: responseErrorBridge{}}
			server.mbus = bus
			server.dispatchRequest(context.Background(), "reply", fixture.request)
			if bus.response == nil || bus.response.StatusCode != fixture.status || bus.response.Error != fixture.message {
				t.Fatalf("%s response=%+v, want status %d and %q", fixture.kind, bus.response, fixture.status, fixture.message)
			}
		})
	}
}
