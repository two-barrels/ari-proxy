// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestConfigData(t *testing.T) {
	integration.TestConfigData(t, &srv{})
}

func TestConfigDelete(t *testing.T) {
	integration.TestConfigDelete(t, &srv{})
}

func TestConfigUpdate(t *testing.T) {
	integration.TestConfigUpdate(t, &srv{})
}
