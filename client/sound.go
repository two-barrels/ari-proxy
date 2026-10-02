// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type sound struct {
	c *Client
}

func (s *sound) List(filters map[string]string, keyFilter *ari.Key) ([]*ari.Key, error) {
	return s.c.listRequest(&proxy.Request{
		Kind: "SoundList",
		Key:  keyFilter,
		SoundList: &proxy.SoundList{
			Filters: filters,
		},
	})
}

func (s *sound) Data(key *ari.Key) (*ari.SoundData, error) {
	data, err := s.c.dataRequest(&proxy.Request{
		Kind: "SoundData",
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	return data.Sound, nil
}
