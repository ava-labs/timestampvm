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

type ChainVM[Block StatelessBlock] interface {
	VMBackend
	ChainBackend[Block]
}

type BlockVM[Block StatelessBlock, State any] interface {
	VMBackend
	Parser[Block]
	Builder[Block]
	Execute(ctx context.Context, parent Block, block Block, state State) (Decider, error)
}

type VMBackend interface {
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

	CreateStaticHandlers(context.Context) (map[string]*common.HTTPHandler, error)
	CreateHandlers(context.Context) (map[string]*common.HTTPHandler, error)
}

type ChainBackend[Block StatelessBlock] interface {
	// VM functionality required to provide chain indexing of accepted blocks
	LastAccepted(ctx context.Context) (ids.ID, error)
	GetBlockIDAtHeight(ctx context.Context, height uint64) (ids.ID, error)
	GetBlock(ctx context.Context, blockID ids.ID) (Block, error)
	Parser[Block]
	Builder[Block]
	Verify(ctx context.Context, parent Block, block Block) (Decider, error)
}

type Parser[Block StatelessBlock] interface {
	ParseBlock(ctx context.Context, bytes []byte) (Block, error)
}

type Builder[Block StatelessBlock] interface {
	BuildBlock(ctx context.Context, parent Block) (Block, error)
}

type Decider interface {
	Accept(context.Context) error
	Abandon(context.Context) error
}

type NoopDecider struct{}

func NewNoopDecider() *NoopDecider { return &NoopDecider{} }

func (NoopDecider) Accept(context.Context) error  { return nil }
func (NoopDecider) Abandon(context.Context) error { return nil }

type StatelessBlock interface {
	ID() ids.ID
	Parent() ids.ID
	Bytes() []byte
	Height() uint64
	Timestamp() time.Time
	Rooter
}

type Rooter interface {
	Root() ids.ID
}
