package server

import (
	"testing"

	"github.com/two-barrels/ari-proxy/v6/internal/integration"
)

func TestPlaybackData(t *testing.T) {
	integration.TestPlaybackData(t, &srv{})
}

func TestPlaybackControl(t *testing.T) {
	integration.TestPlaybackControl(t, &srv{})
}

func TestPlaybackStop(t *testing.T) {
	integration.TestPlaybackStop(t, &srv{})
}
