package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6/client/native"
)

// TestLiveReadOnlyDispatch is opt-in. It reads the ARI password from stdin so
// the credential need not appear in a shell argument or repository file.
func TestLiveReadOnlyDispatch(t *testing.T) {
	if os.Getenv("ARI_LIVE_READONLY") == "" {
		t.Skip("set ARI_LIVE_READONLY=1 to run against a live test PBX")
	}
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	client := native.New(&native.Options{
		URL: os.Getenv("ARI_LIVE_URL"), Username: os.Getenv("ARI_LIVE_USERNAME"),
		Password:   strings.TrimSuffix(password, "\n"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	})
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari, s.mbus = client, bus
	for _, fixture := range []struct {
		kind string
		req  *proxy.Request
	}{
		{"AsteriskInfo", &proxy.Request{Kind: "AsteriskInfo"}},
		{"AsteriskPing", &proxy.Request{Kind: "AsteriskPing"}},
		{"ApplicationList", &proxy.Request{Kind: "ApplicationList"}},
		{"BridgeList", &proxy.Request{Kind: "BridgeList"}},
		{"ChannelList", &proxy.Request{Kind: "ChannelList"}},
		{"EndpointList", &proxy.Request{Kind: "EndpointList"}},
		{"SoundList", &proxy.Request{Kind: "SoundList", SoundList: &proxy.SoundList{}}},
		{"RecordingStoredList", &proxy.Request{Kind: "RecordingStoredList"}},
		{"DeviceStateList", &proxy.Request{Kind: "DeviceStateList"}},
		{"AsteriskModuleList", &proxy.Request{Kind: "AsteriskModuleList"}},
	} {
		t.Run(fixture.kind, func(t *testing.T) {
			bus.response = nil
			s.dispatchRequest(context.Background(), "reply", fixture.req)
			if bus.response == nil {
				t.Fatal("no proxy response")
			}
			wire, err := json.Marshal(bus.response)
			if err != nil {
				t.Fatal(err)
			}
			var got proxy.Response
			if err := json.Unmarshal(wire, &got); err != nil {
				t.Fatal(err)
			}
			if err := got.Err(); err != nil {
				t.Fatal(err)
			}
			if fixture.kind == "AsteriskInfo" {
				if got.Data == nil || got.Data.Asterisk == nil || got.Data.Asterisk.SystemInfo.Version == "" {
					t.Fatal("missing Asterisk version")
				}
			} else if fixture.kind == "AsteriskPing" {
				if got.Data == nil || got.Data.AsteriskPing == nil || got.Data.AsteriskPing.Ping == "" {
					t.Fatal("missing ping response")
				}
			} else {
				t.Logf("returned %d keys", len(got.Keys))
			}
		})
	}
}
