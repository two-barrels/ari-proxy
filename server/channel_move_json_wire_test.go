// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6/client/native"
)

func TestMoveProxyJSONTranslatesToAsteriskJSON(t *testing.T) {
	for _, args := range []string{"agent,ONCALL+value&x=1", ""} {
		t.Run(args, func(t *testing.T) {
			received := make(chan map[string]string, 1)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/ari/channels/channel-1/move" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var body map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				received <- body
				w.WriteHeader(http.StatusNoContent)
			}))
			defer upstream.Close()
			body := map[string]any{"kind": "ChannelMove", "key": map[string]string{"app": "demo", "node": "node-1", "kind": "channel", "id": "channel-1"}, "channel_move": map[string]string{"app": "destination", "app_args": args}}
			wire, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			var req proxy.Request
			if err := json.Unmarshal(wire, &req); err != nil {
				t.Fatal(err)
			}
			s := New()
			s.ari = native.New(&native.Options{URL: upstream.URL + "/ari"})
			bus := &bridgeOptionsResponseBus{}
			s.mbus = bus
			s.dispatchRequest(context.Background(), "reply", &req)
			if bus.response == nil || bus.response.Err() != nil {
				t.Fatalf("response=%+v", bus.response)
			}
			want := map[string]string{"app": "destination"}
			if args != "" {
				want["appArgs"] = args
			}
			select {
			case got := <-received:
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("Asterisk JSON=%v, want %v", got, want)
				}
			default:
				t.Fatal("no Asterisk request")
			}
		})
	}
}
