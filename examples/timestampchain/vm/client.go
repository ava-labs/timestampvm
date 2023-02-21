// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/rpc"
)

// Client defines timestampvm client operations.
type Client interface {
	// ProposeBlock submits [dataHash] as data to be timestamped in a future block
	ProposeBlock(ctx context.Context, dataHash ids.ID) error

	// GetBlock fetches the block corresponding to [blockID].
	// Fetches the last accepted block if [blockID] is the empty ID
	GetBlock(ctx context.Context, blockID ids.ID) (*Block, error)
}

// NewClient creates a new client object.
func NewClient(uri string) Client {
	req := rpc.NewEndpointRequester(uri)
	return &client{req: req}
}

type client struct {
	req rpc.EndpointRequester
}

func (c *client) ProposeBlock(ctx context.Context, dataHash ids.ID) error {
	return c.req.SendRequest(ctx,
		"timestamp.proposeBlock",
		&BlockIDArgs{ID: dataHash},
		&api.EmptyReply{},
	)
}

func (c *client) GetBlock(ctx context.Context, blockID ids.ID) (*Block, error) {
	block := new(Block)
	return block, c.req.SendRequest(ctx,
		"timestamp.getBlock",
		&BlockIDArgs{ID: blockID},
		block,
	)
}
