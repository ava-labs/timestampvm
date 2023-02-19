// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stack

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/api/health"
	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
)

type VMBackend[Block StatelessBlock] interface {
	Initialize(
		ctx context.Context,
		chainCtx *snow.Context,
		dbManager manager.Manager,
		genesisBytes []byte,
		upgradeBytes []byte,
		configBytes []byte,
		toEngine chan<- common.Message,
		fxs []*common.Fx,
		appSender common.AppSender,
	) error

	// Returns nil if the VM is healthy.
	// Periodically called and reported via the node's Health API.
	health.Checker

	// SetState communicates to VM its next state it starts
	SetState(ctx context.Context, state snow.State) error

	// Shutdown is called when the node is shutting down.
	Shutdown(context.Context) error

	// Version returns the version of the VM.
	Version(context.Context) (string, error)

	BlockBackend[Block]

	CreateStaticHandlers(context.Context) (map[string]*common.HTTPHandler, error)
	CreateHandlers(context.Context) (map[string]*common.HTTPHandler, error)
}

type BlockBackend[Block StatelessBlock] interface {
	LastAccepted(context.Context) (ids.ID, error)
	ParseBlock(context.Context, []byte) (Block, error)
	BuildBlock(context.Context, Block) (Block, error)
	GetBlockIDAtHeight(context.Context, uint64) (ids.ID, error) // TODO Handle missing blocks with database.ErrNotFound
	GetBlock(context.Context, ids.ID) (Block, error)
	BlockDecisioner[Block]
}

type BlockDecisioner[Block StatelessBlock] interface {
	Verify(ctx context.Context, parent Block, block Block) error
	Accept(context.Context, Block) error
	Reject(context.Context, Block) error
}

type StatelessBlock interface {
	ID() ids.ID
	Parent() ids.ID
	Bytes() []byte
	Height() uint64
	Timestamp() time.Time
}
