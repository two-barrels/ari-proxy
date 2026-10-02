package client

import (
	"errors"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type endpoint struct {
	c *Client
}

func (e *endpoint) Data(key *ari.Key) (*ari.EndpointData, error) {
	data, err := e.c.dataRequest(&proxy.Request{
		Kind: "EndpointData",
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	return data.Endpoint, nil
}

func (e *endpoint) Get(key *ari.Key) *ari.EndpointHandle {
	k, err := e.c.getRequest(&proxy.Request{
		Kind: "EndpointGet",
		Key:  key,
	})
	if err != nil {
		e.c.log.Warn("failed to get endpoint for handle", "error", err)
		return ari.NewEndpointHandle(key, e)
	}
	return ari.NewEndpointHandle(k, e)
}

func (e *endpoint) List(filter *ari.Key) ([]*ari.Key, error) {
	return e.c.listRequest(&proxy.Request{
		Kind: "EndpointList",
		Key:  filter,
	})
}

func (e *endpoint) ListByTech(tech string, filter *ari.Key) ([]*ari.Key, error) {
	return e.c.listRequest(&proxy.Request{
		Kind: "EndpointListByTech",
		Key:  filter,
		EndpointListByTech: &proxy.EndpointListByTech{
			Tech: tech,
		},
	})
}

func (e *endpoint) Refer(referenceKey *ari.Key, opts ari.EndpointReferOptions) error {
	if referenceKey == nil || referenceKey.App == "" || referenceKey.Node == "" {
		return errors.New("endpoint REFER requires an application and target node")
	}
	return e.c.commandRequest(&proxy.Request{Kind: "EndpointRefer", Key: referenceKey, EndpointRefer: &opts})
}

func (e *endpoint) ReferToEndpoint(key *ari.Key, opts ari.EndpointReferOptions) error {
	if key == nil || key.App == "" || key.Node == "" {
		return errors.New("endpoint REFER requires an application and target node")
	}
	return e.c.commandRequest(&proxy.Request{Kind: "EndpointReferToEndpoint", Key: key, EndpointRefer: &opts})
}
