package client

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/messagebus"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/inconshreveable/log15"
)

type parityRequestBus struct {
	messagebus.Client
	request *proxy.Request
}

func (b *parityRequestBus) Request(_ string, req *proxy.Request) (*proxy.Response, error) {
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	b.request = new(proxy.Request)
	if err := json.Unmarshal(encoded, b.request); err != nil {
		return nil, err
	}
	return &proxy.Response{Key: b.request.Key}, nil
}

func TestNewOptionsSurviveProxyClientEncoding(t *testing.T) {
	bus := &parityRequestBus{}
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler())
	c := &Client{core: &core{mbus: bus, log: logger, prefix: "test."}}
	channelKey := &ari.Key{App: "demo", Node: "node-1", Kind: ari.ChannelKey, ID: "channel-1"}
	bridgeKey := &ari.Key{App: "demo", Node: "node-1", Kind: ari.BridgeKey, ID: "bridge-1"}
	falseValue, zero := false, 0
	if err := c.Channel().SetVariableWithOptions(channelKey, "STATE+NAME", "", &ari.ChannelVariableSetOptions{ReportEvents: &falseValue}); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelVariable.Name != "STATE+NAME" || bus.request.ChannelVariable.ReportEvents == nil || *bus.request.ChannelVariable.ReportEvents || bus.request.ChannelVariable.Value != "" {
		t.Fatalf("set variable=%+v", bus.request)
	}
	if err := c.Channel().HangupWithOptions(channelKey, ari.ChannelHangupOptions{Reason: "busy", ReasonCode: "17"}); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelHangup.Reason != "busy" || bus.request.ChannelHangup.ReasonCode != "17" {
		t.Fatalf("hangup=%+v", bus.request)
	}
	opts := ari.ChannelContinueOptions{Context: "support", Label: "start+1", Priority: &zero}
	opts.Extension = "s"
	if err := c.Channel().ContinueWithOptions(channelKey, opts); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelContinue.Options == nil || !reflect.DeepEqual(*bus.request.ChannelContinue.Options, opts) {
		t.Fatalf("continue=%+v", bus.request)
	}
	if err := c.Channel().Dial(channelKey, "call+leg/2", 7*time.Second); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelDial == nil || bus.request.ChannelDial.Caller != "call+leg/2" || bus.request.ChannelDial.Timeout != 7*time.Second {
		t.Fatalf("dial=%+v", bus.request)
	}
	create := ari.BridgeCreateOptions{Type: "mixing", Name: "support", Variables: map[string]ari.BridgeCreateVariable{"STATE": {Value: "Ready"}, "ALERT": {Value: "Off", ReportEvents: &falseValue}}}
	if _, err := c.Bridge().CreateWithOptions(bridgeKey, create); err != nil {
		t.Fatal(err)
	}
	if bus.request.BridgeCreate.Type != create.Type || bus.request.BridgeCreate.Name != create.Name || !reflect.DeepEqual(bus.request.BridgeCreate.Variables, create.Variables) {
		t.Fatalf("bridge create=%+v", bus.request)
	}
	if err := c.Bridge().MOH(bridgeKey, "custom+class"); err != nil {
		t.Fatal(err)
	}
	if bus.request.BridgeMOH == nil || bus.request.BridgeMOH.Class != "custom+class" {
		t.Fatalf("bridge moh=%+v", bus.request)
	}
	if err := c.Bridge().RemoveChannel(bridgeKey, "channel+1/2"); err != nil {
		t.Fatal(err)
	}
	if bus.request.BridgeRemoveChannel == nil || bus.request.BridgeRemoveChannel.Channel != "channel+1/2" {
		t.Fatalf("bridge remove channel=%+v", bus.request)
	}
	if err := c.Bridge().AddChannelWithOptions(bridgeKey, "channel-1", &ari.BridgeAddChannelOptions{InhibitConnectedLineUpdates: &falseValue}); err != nil {
		t.Fatal(err)
	}
	if bus.request.BridgeAddChannel.InhibitConnectedLineUpdates == nil || *bus.request.BridgeAddChannel.InhibitConnectedLineUpdates {
		t.Fatalf("bridge add=%+v", bus.request)
	}
	play := ari.BridgePlayOptions{Media: []string{"sound:one", "sound:two"}, AnnouncerFormat: "slin16", Lang: "en", OffsetMS: &zero}
	if _, err := c.Bridge().PlayWithOptions(bridgeKey, "playback-1", play); err != nil {
		t.Fatal(err)
	}
	if bus.request.BridgePlay == nil || bus.request.BridgePlay.Options == nil || !reflect.DeepEqual(*bus.request.BridgePlay.Options, play) {
		t.Fatalf("bridge play=%+v", bus.request)
	}
	record := &ari.RecordingOptions{Format: "wav", RecorderFormat: "slin16", MaxDuration: 10 * time.Second, MaxSilence: 2 * time.Second, Exists: "overwrite", Beep: true, Terminate: "#"}
	if _, err := c.Bridge().Record(bridgeKey, "recording-1", record); err != nil {
		t.Fatal(err)
	}
	if bus.request.BridgeRecord == nil || bus.request.BridgeRecord.Name != "recording-1" || !reflect.DeepEqual(bus.request.BridgeRecord.Options, record) {
		t.Fatalf("bridge record=%+v", bus.request)
	}
	createChannel := ari.ChannelCreateRequest{Endpoint: "PJSIP/alice", App: "demo", AppArgs: "one,two", ChannelID: "channel-2", OtherChannelID: "other-2", Originator: "parent-1", Formats: "ulaw,slin16", Variables: map[string]string{"STATE": "Ready"}}
	if _, err := c.Channel().Create(channelKey, createChannel); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelCreate == nil || !reflect.DeepEqual(bus.request.ChannelCreate.ChannelCreateRequest, createChannel) {
		t.Fatalf("channel create=%+v", bus.request)
	}
	channelPlay := ari.ChannelPlayOptions{Media: []string{"sound:one"}, Lang: "en", OffsetMS: &zero}
	if _, err := c.Channel().PlayWithOptions(channelKey, "playback-2", channelPlay); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelPlay == nil || bus.request.ChannelPlay.Options == nil || !reflect.DeepEqual(*bus.request.ChannelPlay.Options, channelPlay) {
		t.Fatalf("channel play=%+v", bus.request)
	}
	external := ari.ExternalMediaOptions{ChannelID: "external-1", App: "demo", ExternalHost: "127.0.0.1:5000", Format: "slin16", TransportData: "a=b&c=d"}
	if _, err := c.Channel().ExternalMedia(channelKey, external); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelExternalMedia == nil || !reflect.DeepEqual(bus.request.ChannelExternalMedia.Options, external) {
		t.Fatalf("external media=%+v", bus.request)
	}
	userEvent := &ari.ChannelUserevent{Eventname: "customer/alert", Userevent: map[string]any{"ticket": "42"}}
	if err := c.Channel().UserEvent(channelKey, userEvent); err != nil {
		t.Fatal(err)
	}
	if bus.request.ChannelUserevent == nil || bus.request.ChannelUserevent.UserEvent.Eventname != userEvent.Eventname || !reflect.DeepEqual(bus.request.ChannelUserevent.UserEvent.Userevent, userEvent.Userevent) {
		t.Fatalf("user event=%+v", bus.request)
	}
}
