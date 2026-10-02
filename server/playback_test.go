// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestPlaybackData(t *testing.T) {
	integration.TestPlaybackData(t, &srv{})
}

func TestPlaybackControl(t *testing.T) {
	integration.TestPlaybackControl(t, &srv{})
}

func TestPlaybackStop(t *testing.T) {
	integration.TestPlaybackStop(t, &srv{})
}
