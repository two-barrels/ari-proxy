package client

import (
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type externalMediaRequestBus struct {
	messagebus.Client
	requests []*proxy.Request
}

func (b *externalMediaRequestBus) Request(_ string, req *proxy.Request) (*proxy.Response, error) {
	b.requests = append(b.requests, req)
	return &proxy.Response{Key: &ari.Key{
		App:  req.Key.App,
		Node: req.Key.Node,
		Kind: ari.ChannelKey,
		ID:   req.ChannelExternalMedia.Options.ChannelID,
	}}, nil
}

func TestStageExternalMediaUsesExternalMediaRequest(t *testing.T) {
	bus := &externalMediaRequestBus{}
	c := &Client{core: &core{mbus: bus, log: log15.New(), prefix: "test."}}
	reference := &ari.Key{App: "demo", Node: "node-1"}
	opts := ari.ExternalMediaOptions{
		ChannelID:    "external-1",
		App:          "demo",
		ExternalHost: "127.0.0.1:5000",
		Format:       "slin16",
	}

	handle, err := c.Channel().StageExternalMedia(reference, opts)
	if err != nil {
		t.Fatal(err)
	}
	if handle.ID() != opts.ChannelID {
		t.Fatalf("staged handle ID = %q, want %q", handle.ID(), opts.ChannelID)
	}
	if len(bus.requests) != 1 || bus.requests[0].Kind != "ChannelStageExternalMedia" {
		t.Fatalf("staging requests = %v, want one ChannelStageExternalMedia request", bus.requests)
	}
	if bus.requests[0].ChannelExternalMedia == nil || !reflect.DeepEqual(bus.requests[0].ChannelExternalMedia.Options, opts) {
		t.Fatalf("staging request lost external media options: %+v", bus.requests[0].ChannelExternalMedia)
	}

	if err := handle.Exec(); err != nil {
		t.Fatal(err)
	}
	if len(bus.requests) != 2 || bus.requests[1].Kind != "ChannelExternalMedia" {
		t.Fatalf("execution requests = %v, want ChannelExternalMedia after staging", bus.requests)
	}
}
