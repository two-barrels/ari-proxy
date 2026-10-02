// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

func (a *application) ClaimChannel(key *ari.Key, channelID string) error {
	if key == nil || key.ID == "" || key.App == "" || key.Node == "" {
		return errors.New("channel claim requires an application and target node")
	}
	if channelID == "" {
		return errors.New("channel ID not supplied")
	}
	return a.c.commandRequest(&proxy.Request{Kind: "EventClaimChannel", Key: key,
		EventClaim: &proxy.EventClaim{ChannelID: channelID}})
}

type application struct {
	c *Client
}

func (a *application) List(filter *ari.Key) ([]*ari.Key, error) {
	return a.c.listRequest(&proxy.Request{
		Kind: "ApplicationList",
		Key:  filter,
	})
}

func (a *application) Data(key *ari.Key) (*ari.ApplicationData, error) {
	ret, err := a.c.dataRequest(&proxy.Request{
		Kind: "ApplicationData",
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	return ret.Application, nil
}

func (a *application) Get(key *ari.Key) *ari.ApplicationHandle {
	k, err := a.c.getRequest(&proxy.Request{
		Kind: "ApplicationGet",
		Key:  key,
	})
	if err != nil {
		a.c.log.Warn("failed to make data request for application", "error", err)
		return ari.NewApplicationHandle(key, a)
	}
	return ari.NewApplicationHandle(k, a)
}

func (a *application) Subscribe(key *ari.Key, eventSource string) (err error) {
	return a.c.commandRequest(&proxy.Request{
		Kind: "ApplicationSubscribe",
		Key:  key,
		ApplicationSubscribe: &proxy.ApplicationSubscribe{
			EventSource: eventSource,
		},
	})
}

func (a *application) Unsubscribe(key *ari.Key, eventSource string) (err error) {
	return a.c.commandRequest(&proxy.Request{
		Kind: "ApplicationUnsubscribe",
		Key:  key,
		ApplicationSubscribe: &proxy.ApplicationSubscribe{
			EventSource: eventSource,
		},
	})
}

func (a *application) FilterEvents(key *ari.Key, filter *ari.ApplicationEventFilter) (*ari.ApplicationData, error) {
	data, err := a.c.dataRequest(&proxy.Request{
		Kind: "ApplicationFilterEvents", Key: key, ApplicationEventFilter: filter,
	})
	if err != nil {
		return nil, err
	}
	if data.Application == nil {
		return nil, ErrNil
	}
	return data.Application, nil
}
