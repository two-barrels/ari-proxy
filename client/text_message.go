// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"

	"github.com/two-barrels/ari-proxy/v6/proxy"
	"github.com/two-barrels/ari/v6"
)

type textMessage struct{ c *Client }

func (t *textMessage) uniqueNode() (*ari.Key, error) {
	if t.c.appName == "" || t.c.cluster == nil {
		return nil, errors.New("text message requires an application and target node")
	}
	members := t.c.cluster.Matching("", t.c.appName, t.c.clusterMaxAge)
	if len(members) != 1 {
		return nil, errors.New("text message target is ambiguous; use a keyed send method")
	}
	return ari.NodeKey(t.c.appName, members[0].ID), nil
}

func (t *textMessage) Send(from, tech, resource, body string, vars map[string]string) error {
	key, err := t.uniqueNode()
	if err != nil {
		return err
	}
	return t.SendWithKey(ari.NewEndpointKey(tech, resource, ari.WithApp(key.App), ari.WithNode(key.Node)), from, body, vars)
}

func (t *textMessage) SendByURI(from, to, body string, vars map[string]string) error {
	key, err := t.uniqueNode()
	if err != nil {
		return err
	}
	return t.SendByURIWithKey(key, from, to, body, vars)
}

func (t *textMessage) SendWithKey(key *ari.Key, from, body string, vars map[string]string) error {
	if key == nil || key.Kind != ari.EndpointKey || key.ID == "" || key.App == "" || key.Node == "" || from == "" {
		return errors.New("text message requires an endpoint, source, application, and target node")
	}
	if vars == nil {
		vars = map[string]string{}
	}
	return t.c.commandRequest(&proxy.Request{Kind: "TextMessageSend", Key: key,
		TextMessageSend: &proxy.TextMessageSend{From: from, Body: body, Variables: vars}})
}

func (t *textMessage) SendByURIWithKey(key *ari.Key, from, to, body string, vars map[string]string) error {
	if key == nil || key.App == "" || key.Node == "" || from == "" || to == "" {
		return errors.New("text message requires a source, destination, application, and target node")
	}
	if vars == nil {
		vars = map[string]string{}
	}
	return t.c.commandRequest(&proxy.Request{Kind: "TextMessageSendByURI", Key: key,
		TextMessageSend: &proxy.TextMessageSend{From: from, To: to, Body: body, Variables: vars}})
}
