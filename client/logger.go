// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"

	"github.com/inconshreveable/log15"
	"golang.org/x/exp/slog"
)

// Logger defaults to a discard handler (null output).
// If you wish to enable logging, you can set your own
// handler like so:
//
//	ari.Logger.SetHandler(log15.StderrHandler)
var Logger = log15.New()

func init() {
	// Null logger, by default
	Logger.SetHandler(log15.DiscardHandler())
}

// SetLogger adapts the proxy's log15 records to the ari/v6 logging API.
func (c *Client) SetLogger(logger *slog.Logger) {
	if logger == nil {
		return
	}
	c.log.SetHandler(log15.FuncHandler(func(record *log15.Record) error {
		level := slog.LevelInfo
		switch record.Lvl {
		case log15.LvlCrit, log15.LvlError:
			level = slog.LevelError
		case log15.LvlWarn:
			level = slog.LevelWarn
		case log15.LvlDebug:
			level = slog.LevelDebug
		}
		logger.Log(context.Background(), level, record.Msg, record.Ctx...)
		return nil
	}))
}
