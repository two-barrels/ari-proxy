// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestLiveRecordingData(t *testing.T) {
	integration.TestLiveRecordingData(t, &srv{})
}

/*
func TestLiveRecordingDelete(t *testing.T) {
	integration.TestLiveRecordingDelete(t, &srv{})
}
*/

func TestLiveRecordingMute(t *testing.T) {
	integration.TestLiveRecordingMute(t, &srv{})
}

func TestLiveRecordingUnmute(t *testing.T) {
	integration.TestLiveRecordingUnmute(t, &srv{})
}

func TestLiveRecordingPause(t *testing.T) {
	integration.TestLiveRecordingPause(t, &srv{})
}

func TestLiveRecordingStop(t *testing.T) {
	integration.TestLiveRecordingStop(t, &srv{})
}

func TestLiveRecordingResume(t *testing.T) {
	integration.TestLiveRecordingResume(t, &srv{})
}

func TestLiveRecordingScrap(t *testing.T) {
	integration.TestLiveRecordingScrap(t, &srv{})
}
