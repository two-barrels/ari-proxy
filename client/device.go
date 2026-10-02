// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type deviceState struct {
	c *Client
}

func (ds *deviceState) Get(key *ari.Key) *ari.DeviceStateHandle {
	k, err := ds.c.getRequest(&proxy.Request{
		Kind: "DeviceStateGet",
		Key:  key,
	})
	if err != nil {
		ds.c.log.Warn("failed to get device state for handle")
		return ari.NewDeviceStateHandle(key, ds)
	}
	return ari.NewDeviceStateHandle(k, ds)
}

func (ds *deviceState) List(filter *ari.Key) ([]*ari.Key, error) {
	return ds.c.listRequest(&proxy.Request{
		Kind: "DeviceStateList",
		Key:  filter,
	})
}

func (ds *deviceState) Data(key *ari.Key) (*ari.DeviceStateData, error) {
	data, err := ds.c.dataRequest(&proxy.Request{
		Kind: "DeviceStateData",
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	return data.DeviceState, nil
}

func (ds *deviceState) Update(key *ari.Key, state string) error {
	return ds.c.commandRequest(&proxy.Request{
		Kind: "DeviceStateUpdate",
		Key:  key,
		DeviceStateUpdate: &proxy.DeviceStateUpdate{
			State: state,
		},
	})
}

func (ds *deviceState) Delete(key *ari.Key) error {
	return ds.c.commandRequest(&proxy.Request{
		Kind: "DeviceStateDelete",
		Key:  key,
	})
}
