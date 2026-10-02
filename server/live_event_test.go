package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	clientbus "github.com/two-barrels/ari-proxy/v6/client/bus"
	"github.com/two-barrels/ari/v6"
	"github.com/two-barrels/ari/v6/client/native"
	"github.com/inconshreveable/log15"
)

// TestLiveEventForwarding uses a real ARI websocket and an isolated external
// media channel. The proxy message-bus transport is replaced with an in-memory
// JSON transport so the event path can be checked without a NATS deployment.
func TestLiveEventForwarding(t *testing.T) {
	if os.Getenv("ARI_LIVE_EVENT_TEST") == "" {
		t.Skip("set ARI_LIVE_EVENT_TEST=1 to run against a live test PBX")
	}
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	url := strings.TrimRight(os.Getenv("ARI_LIVE_URL"), "/")
	if !strings.HasPrefix(url, "http://") {
		t.Fatal("ARI_LIVE_URL must use http:// for this test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	upstream := native.New(&native.Options{
		URL: url, WebsocketURL: "ws://" + strings.TrimPrefix(url, "http://") + "/events",
		Application: "ari-testing", Username: os.Getenv("ARI_LIVE_USERNAME"),
		Password: strings.TrimSuffix(password, "\n"), HTTPClient: &http.Client{Timeout: 10 * time.Second},
	})
	if err := upstream.ConnectWithContext(ctx); err != nil {
		t.Fatal(err)
	}
	defer upstream.Close()
	info, err := upstream.Asterisk().Info(nil)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan struct{})
	receiver := &eventTestReceiver{}
	clientEvents := clientbus.New("live.", receiver, log15.New()).Subscribe(ari.AppKey("ari-testing"), ari.Events.All)
	if clientEvents == nil {
		t.Fatal("proxy client subscription failed")
	}
	defer clientEvents.Cancel()
	publications := make(chan eventTestPublication, 32)
	publisher := &eventTestPublisher{publications: publications}
	wantSubject := fmt.Sprintf("live.event.ari-testing.%s", info.SystemInfo.EntityID)
	publisher.deliver = func(subject string, data []byte) {
		if subject == wantSubject {
			receiver.callback(data)
		}
	}
	s := New()
	s.Application, s.AsteriskID, s.MBPrefix = "ari-testing", info.SystemInfo.EntityID, "live."
	s.ari = eventTestClient{Client: upstream, bus: eventTestBus{Bus: upstream.Bus(), ready: ready}}
	s.mbus = publisher
	eventCtx, stopEvents := context.WithCancel(ctx)
	eventsDone := make(chan struct{})
	go func() { defer close(eventsDone); s.runEventHandler(eventCtx) }()
	defer func() { stopEvents(); <-eventsDone }()
	select {
	case <-ready:
	case <-ctx.Done():
		t.Fatal("proxy did not subscribe to native events")
	}
	id := fmt.Sprintf("codex-ari-testing-event-%d", time.Now().UnixNano())
	key := ari.NewKey(ari.ChannelKey, id)
	created := true // Always attempt cleanup: some ARI errors occur after creation.
	defer func() {
		if created {
			if err := upstream.Channel().Hangup(key, "normal"); err != nil && native.CodeFromError(err) != http.StatusNotFound {
				t.Errorf("cleanup channel %s: %v", id, err)
			}
		}
	}()
	if _, err := upstream.Channel().ExternalMedia(nil, ari.ExternalMediaOptions{
		ChannelID: id, App: "ari-testing", ExternalHost: "127.0.0.1:9", Format: "ulaw",
	}); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(5 * time.Second)
	for {
		select {
		case event := <-clientEvents.Events():
			start, ok := event.(*ari.StasisStart)
			if !ok || start.Channel.ID != id {
				continue
			}
			encoded, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			var clientJSON any
			if err := json.Unmarshal(encoded, &clientJSON); err != nil {
				t.Fatal(err)
			}
			found := false
			for !found {
				select {
				case published := <-publications:
					if published.subject != wantSubject {
						continue
					}
					var serverJSON any
					if err := json.Unmarshal(published.data, &serverJSON); err != nil {
						t.Fatal(err)
					}
					if reflect.DeepEqual(serverJSON, clientJSON) {
						found = true
					}
				case <-deadline:
					t.Fatal("matching proxy publication not received")
				}
			}
			if err := upstream.Channel().Hangup(key, "normal"); err != nil {
				t.Fatal(err)
			}
			created = false
			return
		case <-deadline:
			t.Fatal("StasisStart was not forwarded through proxy event path")
		}
	}
}
