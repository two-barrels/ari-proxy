// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestModulesData(t *testing.T) {
	integration.TestModulesData(t, &srv{})
}

func TestModulesLoad(t *testing.T) {
	integration.TestModulesLoad(t, &srv{})
}

func TestModulesReload(t *testing.T) {
	integration.TestModulesReload(t, &srv{})
}

func TestModulesUnload(t *testing.T) {
	integration.TestModulesUnload(t, &srv{})
}

func TestModulesList(t *testing.T) {
	integration.TestModulesList(t, &srv{})
}
