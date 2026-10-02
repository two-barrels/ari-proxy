// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package session

import "github.com/two-barrels/ari/v6"

// Transport defines how the commands and events are sent.
type Transport interface {

	// Command sends a command and waits for a response
	Command(name string, body interface{}, resp interface{}) error

	// Event dispatches an event
	Event(evt ari.Event) error
}
