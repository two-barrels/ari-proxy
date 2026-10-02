package client

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type routeRequestBus struct {
	messagebus.Client
	request *proxy.Request
	calls   int
}

func (b *routeRequestBus) Request(_ string, req *proxy.Request) (*proxy.Response, error) {
	b.calls++
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	b.request = new(proxy.Request)
	if err := json.Unmarshal(data, b.request); err != nil {
		return nil, err
	}
	id, kind := "", ""
	switch req.Kind {
	case "BridgeCreateWithoutID":
		id, kind = "server-bridge", ari.BridgeKey
	case "BridgeCreateOnCollection":
		id, kind = req.BridgeCreate.BridgeID, ari.BridgeKey
	case "BridgePlayWithoutID", "ChannelPlayWithoutID":
		id, kind = "server-playback", ari.PlaybackKey
	case "BridgePlayOnCollection":
		id, kind = req.BridgePlay.PlaybackID, ari.PlaybackKey
	case "ChannelPlayOnCollection":
		id, kind = req.ChannelPlay.PlaybackID, ari.PlaybackKey
	case "ChannelOriginateWithID":
		id, kind = req.ChannelOriginate.OriginateRequest.ChannelID, ari.ChannelKey
	case "ChannelSnoopWithoutID":
		id, kind = "server-snoop", ari.ChannelKey
	case "ChannelSnoopOnCollection", "ChannelSnoop":
		id, kind = req.ChannelSnoop.SnoopID, ari.ChannelKey
	}
	return &proxy.Response{Key: req.Key.New(kind, id)}, nil
}

func TestExplicitRoutesSurviveProxyRequest(t *testing.T) {
	bus := &routeRequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	key := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}
	zero := 0
	bridgeOpts := ari.BridgeCreateOptions{Type: "mixing", Name: "support+one", Variables: map[string]ari.BridgeCreateVariable{"STATE": {Value: "Ready"}}}
	bridge, err := c.Bridge().CreateWithoutID(key, bridgeOpts)
	if err != nil || bridge.ID() != "server-bridge" || bus.request.Kind != "BridgeCreateWithoutID" || bus.request.BridgeCreate.Type != bridgeOpts.Type || bus.request.BridgeCreate.Name != bridgeOpts.Name || !reflect.DeepEqual(bus.request.BridgeCreate.Variables, bridgeOpts.Variables) {
		t.Fatalf("bridge create=%v request=%+v error=%v", bridge, bus.request, err)
	}
	bridge, err = c.Bridge().CreateOnCollection(key, "bridge+9", bridgeOpts)
	if err != nil || bridge.ID() != "bridge+9" || bus.request.Kind != "BridgeCreateOnCollection" || bus.request.BridgeCreate.BridgeID != "bridge+9" || bus.request.BridgeCreate.Type != bridgeOpts.Type || bus.request.BridgeCreate.Name != bridgeOpts.Name {
		t.Fatalf("bridge create with ID=%v request=%+v error=%v", bridge, bus.request, err)
	}
	skip, offset := 4500, 300
	bplay := ari.BridgePlayOptions{Media: []string{"sound:one"}, AnnouncerFormat: "slin16", Lang: "en", OffsetMS: &zero, SkipMS: &skip}
	playback, err := c.Bridge().PlayWithoutID(key.New(ari.BridgeKey, "bridge-1"), bplay)
	if err != nil || playback.ID() != "server-playback" || bus.request.Kind != "BridgePlayWithoutID" || !reflect.DeepEqual(*bus.request.BridgePlay.Options, bplay) {
		t.Fatalf("bridge playback=%v request=%+v error=%v", playback, bus.request, err)
	}
	playback, err = c.Bridge().PlayOnCollection(key.New(ari.BridgeKey, "bridge-1"), "play+9", bplay)
	if err != nil || playback.ID() != "play+9" || bus.request.Kind != "BridgePlayOnCollection" || bus.request.BridgePlay.PlaybackID != "play+9" {
		t.Fatalf("bridge play with ID=%v request=%+v error=%v", playback, bus.request, err)
	}
	orig := ari.OriginateRequest{ChannelID: "channel-9", Endpoint: "PJSIP/alice", Extension: "s+1", Context: "support/main", Priority: 2, Label: "start+1", App: "demo", AppArgs: "one,two", CallerID: "Alice <100>", Timeout: 15, OtherChannelID: "other-9", Originator: "parent-1", Formats: "ulaw,slin16", Variables: map[string]string{"STATE": "Ready"}}
	channel, err := c.Channel().OriginateWithID(key, orig)
	if err != nil || channel.ID() != "channel-9" || bus.request.Kind != "ChannelOriginateWithID" || !reflect.DeepEqual(bus.request.ChannelOriginate.OriginateRequest, orig) {
		t.Fatalf("channel originate=%v request=%+v error=%v", channel, bus.request, err)
	}
	cplay := ari.ChannelPlayOptions{Media: []string{"sound:two"}, Lang: "en+US", OffsetMS: &offset, SkipMS: &zero}
	playback, err = c.Channel().PlayWithoutID(key, cplay)
	if err != nil || playback.ID() != "server-playback" || bus.request.Kind != "ChannelPlayWithoutID" || !reflect.DeepEqual(*bus.request.ChannelPlay.Options, cplay) {
		t.Fatalf("channel playback=%v request=%+v error=%v", playback, bus.request, err)
	}
	playback, err = c.Channel().PlayOnCollection(key, "play+9", cplay)
	if err != nil || playback.ID() != "play+9" || bus.request.Kind != "ChannelPlayOnCollection" || bus.request.ChannelPlay.PlaybackID != "play+9" {
		t.Fatalf("channel play with ID=%v request=%+v error=%v", playback, bus.request, err)
	}
	snoopOpts := &ari.SnoopOptions{App: "demo", AppArgs: "one,two", Spy: ari.DirectionIn, Whisper: ari.DirectionOut}
	snoop, err := c.Channel().SnoopWithoutID(key, snoopOpts)
	if err != nil || snoop.ID() != "server-snoop" || bus.request.Kind != "ChannelSnoopWithoutID" || !reflect.DeepEqual(bus.request.ChannelSnoop.Options, snoopOpts) {
		t.Fatalf("channel snoop=%v request=%+v error=%v", snoop, bus.request, err)
	}
	snoop, err = c.Channel().Snoop(key, "snoop-9", snoopOpts)
	if err != nil || snoop.ID() != "snoop-9" || bus.request.Kind != "ChannelSnoop" || bus.request.ChannelSnoop.SnoopID != "snoop-9" || !reflect.DeepEqual(bus.request.ChannelSnoop.Options, snoopOpts) {
		t.Fatalf("channel snoop path=%v request=%+v error=%v", snoop, bus.request, err)
	}
	snoop, err = c.Channel().SnoopOnCollection(key, "snoop+9", &ari.SnoopOptions{App: "demo"})
	if err != nil || snoop.ID() != "snoop+9" || bus.request.Kind != "ChannelSnoopOnCollection" || bus.request.ChannelSnoop.SnoopID != "snoop+9" {
		t.Fatalf("channel snoop with ID=%v request=%+v error=%v", snoop, bus.request, err)
	}
	before := bus.calls
	if _, err := c.Bridge().CreateWithoutID(&ari.Key{App: "demo"}, bridgeOpts); err == nil || bus.calls != before {
		t.Fatalf("unscoped bridge create error=%v calls=%d", err, bus.calls)
	}
	if _, err := c.Channel().OriginateWithID(&ari.Key{App: "demo"}, orig); err == nil || bus.calls != before {
		t.Fatalf("unscoped originate error=%v calls=%d", err, bus.calls)
	}
}
