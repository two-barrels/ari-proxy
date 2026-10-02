// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/two-barrels/ari/v6"
)

// Benchmark the current per-dialog event copy and serialization path.
func BenchmarkDialogEventCopies(b *testing.B) {
	raw := []byte(`{"type":"StasisStart","application":"demo","asterisk_id":"node-1","args":["first","second"],"channel":{"id":"channel-1","name":"PJSIP/100-00000001","state":"Up","creationtime":"2025-01-01T00:00:00.000+0000","future_channel_field":{"codec":"opus","media":{"rx":1234,"tx":5678}}},"future_event_field":{"trace":"abc123","metadata":{"queue":"support","region":"west"}}}`)
	for _, dialogs := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("dialogs=%d", dialogs), func(b *testing.B) {
			b.SetBytes(int64(len(raw)))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				event, err := ari.DecodeEvent(raw)
				if err != nil {
					b.Fatal(err)
				}
				for dialog := 0; dialog < dialogs; dialog++ {
					copy, err := ari.CloneEvent(event)
					if err != nil {
						b.Fatal(err)
					}
					copy.SetDialog("dialog-1")
					if _, err := json.Marshal(copy); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
