// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestEndpointList(t *testing.T) {
	integration.TestEndpointList(t, &srv{})
}

func TestEndpointListByTech(t *testing.T) {
	integration.TestEndpointListByTech(t, &srv{})
}

func TestEndpointData(t *testing.T) {
	integration.TestEndpointData(t, &srv{})
}
