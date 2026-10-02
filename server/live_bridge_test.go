// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/client/native"
)

// TestLiveBridgeDispatch checks the proxy's state-changing request path on a
// uniquely named bridge, with direct cleanup even if an ARI create returns an
// error after allocating the bridge.
func TestLiveBridgeDispatch(t *testing.T) {
	if os.Getenv("ARI_LIVE_BRIDGE_TEST") == "" {
		t.Skip("set ARI_LIVE_BRIDGE_TEST=1 to run against a live test PBX")
	}
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	url := strings.TrimRight(os.Getenv("ARI_LIVE_URL"), "/")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	upstream := native.New(&native.Options{URL: url, WebsocketURL: "ws://" + strings.TrimPrefix(url, "http://") + "/events", Application: "ari-testing", Username: os.Getenv("ARI_LIVE_USERNAME"), Password: strings.TrimSuffix(password, "\n"), HTTPClient: &http.Client{Timeout: 10 * time.Second}})
	if err := upstream.ConnectWithContext(ctx); err != nil {
		t.Fatal(err)
	}
	defer upstream.Close()
	info, err := upstream.Asterisk().Info(nil)
	if err != nil {
		t.Fatal(err)
	}
	bus := &bridgeOptionsResponseBus{}
	s := New()
	s.ari, s.mbus = upstream, bus
	id := fmt.Sprintf("codex-ari-testing-proxy-%d", time.Now().UnixNano())
	key := &ari.Key{App: "ari-testing", Node: info.SystemInfo.EntityID, Kind: ari.BridgeKey, ID: id}
	created := true
	defer func() {
		if created {
			if err := upstream.Bridge().Delete(key); err != nil && native.CodeFromError(err) != http.StatusNotFound {
				t.Errorf("cleanup bridge %s: %v", id, err)
			}
		}
	}()
	request := func(req *proxy.Request) *proxy.Response {
		t.Helper()
		bus.response = nil
		s.dispatchRequest(ctx, "reply", req)
		if bus.response == nil {
			t.Fatal("missing proxy response")
		}
		data, err := json.Marshal(bus.response)
		if err != nil {
			t.Fatal(err)
		}
		var response proxy.Response
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatal(err)
		}
		if err := response.Err(); err != nil {
			t.Fatal(err)
		}
		return &response
	}
	createdResp := request(&proxy.Request{Kind: "BridgeCreateOnCollection", Key: &ari.Key{App: key.App, Node: key.Node, Kind: key.Kind}, BridgeCreate: &proxy.BridgeCreate{BridgeID: id, Type: "mixing", Name: id, Variables: map[string]ari.BridgeCreateVariable{"ARI_TEST_MARKER": {Value: "created"}}}})
	if createdResp.Key == nil || createdResp.Key.ID != id {
		t.Fatalf("created key=%+v", createdResp.Key)
	}
	got := request(&proxy.Request{Kind: "BridgeVariableGet", Key: key, BridgeVariable: &proxy.BridgeVariable{Name: "ARI_TEST_MARKER"}})
	if got.Data == nil || got.Data.Variable != "created" {
		t.Fatalf("created variable=%+v", got.Data)
	}
	request(&proxy.Request{Kind: "BridgeVariablesSet", Key: key, BridgeVariables: &proxy.BridgeVariables{Values: map[string]ari.BridgeVariableAssignment{"ARI_TEST_MARKER": {Value: "updated"}}}})
	got = request(&proxy.Request{Kind: "BridgeVariableGet", Key: key, BridgeVariable: &proxy.BridgeVariable{Name: "ARI_TEST_MARKER"}})
	if got.Data == nil || got.Data.Variable != "updated" {
		t.Fatalf("updated variable=%+v", got.Data)
	}
	request(&proxy.Request{Kind: "BridgeDelete", Key: key})
	created = false
	if _, err := upstream.Bridge().Data(key); native.CodeFromError(err) != http.StatusNotFound {
		t.Fatalf("post-delete status=%d err=%v", native.CodeFromError(err), err)
	}
}
