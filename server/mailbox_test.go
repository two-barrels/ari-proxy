// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestMailboxList(t *testing.T) {
	integration.TestMailboxList(t, &srv{})
}

func TestMailboxUpdate(t *testing.T) {
	integration.TestMailboxUpdate(t, &srv{})
}

func TestMailboxDelete(t *testing.T) {
	integration.TestMailboxDelete(t, &srv{})
}

func TestMailboxData(t *testing.T) {
	integration.TestMailboxData(t, &srv{})
}
