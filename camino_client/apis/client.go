// Copyright (C) 2022, Chain4Travel AG. All rights reserved.
//
// This file is a derived work, based on ava-labs code
//
// It is distributed under the same license conditions as the
// original code from which it is derived.
//
// Much love to the original authors for their work.

package apis

import (
	"time"

	"github.com/chain4travel/caminogo/api/admin"
	"github.com/chain4travel/caminogo/api/health"
	"github.com/chain4travel/caminogo/api/info"
	"github.com/chain4travel/caminogo/api/ipcs"
	"github.com/chain4travel/caminogo/api/keystore"
	"github.com/chain4travel/caminogo/vms/avm"
	"github.com/chain4travel/caminogo/vms/platformvm"
)

const (
	XChain = "X"
)

type Client struct {
	admin    admin.Client
	xChain   avm.Client
	health   health.Client
	info     info.Client
	ipcs     ipcs.Client
	keystore keystore.Client
	platform platformvm.Client
}

// Returns a Client for interacting with the P Chain endpoint
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		admin:    admin.NewClient(uri),
		xChain:   avm.NewClient(uri, XChain),
		health:   health.NewClient(uri),
		info:     info.NewClient(uri),
		ipcs:     ipcs.NewClient(uri),
		keystore: keystore.NewClient(uri),
		platform: platformvm.NewClient(uri),
	}
}

func (c *Client) PChainAPI() platformvm.Client {
	return c.platform
}

func (c *Client) XChainAPI() avm.Client {
	return c.xChain
}

func (c *Client) InfoAPI() info.Client {
	return c.info
}

func (c *Client) HealthAPI() health.Client {
	return c.health
}

func (c *Client) IpcsAPI() ipcs.Client {
	return c.ipcs
}

func (c *Client) KeystoreAPI() keystore.Client {
	return c.keystore
}

func (c *Client) AdminAPI() admin.Client {
	return c.admin
}
