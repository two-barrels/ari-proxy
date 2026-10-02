package main

import (
	"testing"

	"github.com/spf13/viper"
)

func TestServerARIOptionsEventSubscription(t *testing.T) {
	viper.Set("ari.application", "demo+app")
	viper.Set("ari.subscribe_all", true)
	t.Cleanup(func() { viper.Set("ari.application", ""); viper.Set("ari.subscribe_all", false) })
	opts := serverARIOptions()
	if opts.Application != "demo+app" || !opts.SubscribeAll {
		t.Errorf("application=%q subscribeAll=%t", opts.Application, opts.SubscribeAll)
	}
}
