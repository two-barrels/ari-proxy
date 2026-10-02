// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestDeviceData(t *testing.T) {
	integration.TestDeviceData(t, &srv{})
}

func TestDeviceDelete(t *testing.T) {
	integration.TestDeviceDelete(t, &srv{})
}

func TestDeviceUpdate(t *testing.T) {
	integration.TestDeviceUpdate(t, &srv{})
}

func TestDeviceList(t *testing.T) {
	integration.TestDeviceList(t, &srv{})
}
