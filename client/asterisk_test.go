// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestAsteriskInfo(t *testing.T) {
	integration.TestAsteriskInfo(t, &srv{})
}

func TestAsteriskVariablesGet(t *testing.T) {
	integration.TestAsteriskVariablesGet(t, &srv{})
}

func TestAsteriskVariablesSet(t *testing.T) {
	integration.TestAsteriskVariablesSet(t, &srv{})
}
