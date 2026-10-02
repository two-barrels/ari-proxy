// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestSoundData(t *testing.T) {
	integration.TestSoundData(t, &srv{})
}

func TestSoundList(t *testing.T) {
	integration.TestSoundList(t, &srv{})
}
