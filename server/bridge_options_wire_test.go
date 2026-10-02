package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type bridgeOptionsClient struct {
	ari.Client
	bridge ari.Bridge
}

func (c bridgeOptionsClient) Bridge() ari.Bridge { return c.bridge }

type bridgeOptionsRecorder struct {
	ari.Bridge
	key       *ari.Key
	channelID string
	options   *ari.BridgeAddChannelOptions
	called    bool
}

func (b *bridgeOptionsRecorder) AddChannelWithOptions(key *ari.Key, channelID string, options *ari.BridgeAddChannelOptions) error {
	b.key, b.channelID, b.options, b.called = key, channelID, options, true
	return nil
}

func (b *bridgeOptionsRecorder) AddChannel(*ari.Key, string) error {
	return nil
}

type bridgeOptionsResponseBus struct {
	messagebus.Server
	response *proxy.Response
}

func (b *bridgeOptionsResponseBus) PublishResponse(_ string, response *proxy.Response) error {
	b.response = response
	return nil
}

func TestBridgeAddChannelForwardsOptions(t *testing.T) {
	recorder := &bridgeOptionsRecorder{}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari = bridgeOptionsClient{bridge: recorder}
	s.mbus = bus
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"}
	s.dispatchRequest(context.Background(), "reply", &proxy.Request{
		Kind: "BridgeAddChannel",
		Key:  key,
		BridgeAddChannel: &proxy.BridgeAddChannel{
			Channel: "channel-1", Role: "caller", AbsorbDTMF: true, Mute: true,
		},
	})

	want := &ari.BridgeAddChannelOptions{Role: "caller", AbsorbDTMF: true, Mute: true}
	if !recorder.called || recorder.key != key || recorder.channelID != "channel-1" || !reflect.DeepEqual(recorder.options, want) {
		t.Fatalf("forwarded bridge options: called=%v key=%v channel=%q options=%+v", recorder.called, recorder.key, recorder.channelID, recorder.options)
	}
	if bus.response == nil || bus.response.Err() != nil {
		t.Fatalf("proxy response = %+v", bus.response)
	}
}
