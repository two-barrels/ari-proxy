// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestLoggingList(t *testing.T) {
	integration.TestLoggingList(t, &srv{})
}

func TestLoggingCreate(t *testing.T) {
	integration.TestLoggingCreate(t, &srv{})
}

func TestLoggingRotate(t *testing.T) {
	integration.TestLoggingRotate(t, &srv{})
}

func TestLoggingDelete(t *testing.T) {
	integration.TestLoggingDelete(t, &srv{})
}
