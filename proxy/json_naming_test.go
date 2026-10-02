// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/two-barrels/ari/v6"
)

func TestProxyJSONFieldNames(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"move arguments", ChannelMove{App: "demo", AppArgs: "agent,ONCALL"}, `{"app":"demo","app_args":"agent,ONCALL"}`},
		{"empty move arguments", ChannelMove{App: "demo"}, `{"app":"demo"}`},
		{"playback envelope", ChannelPlay{PlaybackID: "pb-1", MediaURI: "sound:hello"}, `{"playback_id":"pb-1","media_uri":"sound:hello"}`},
		{"embedded creation options", ChannelCreate{ChannelCreateRequest: ari.ChannelCreateRequest{Endpoint: "PJSIP/alice", App: "demo", AppArgs: "agent,ONCALL", ChannelID: "ch-1", OtherChannelID: "ch-2"}}, `{"channel_create_request":{"endpoint":"PJSIP/alice","app":"demo","appArgs":"agent,ONCALL","channelId":"ch-1","otherChannelId":"ch-2"}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wire, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := json.Unmarshal(wire, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("wire=%s; want=%s", wire, test.want)
			}
		})
	}
}
