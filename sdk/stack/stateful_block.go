// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stack

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
)

// Type assertion
var _ snowman.Block = (*Block[StatelessBlock])(nil)

// Block implements the snowman.Block interface
type Block[B StatelessBlock] struct {
	innerBlock B
	cache      *BlockCache[B]
	backend    BlockBackend[B]

	// decider is populated when Verify is called on the block.
	// When the block is marked as Accepted/Rejected by the consensus engine,
	// Accept/Abandon will be called on the block.
	// If the VM shuts down, Abandon will be called on all verified blocks, but this
	// is not guaranteed in the event of an ungraceful shutdown.
	decider Decider

	status choices.Status
}

func (b *Block[B]) ID() ids.ID {
	return b.innerBlock.ID()
}

func (b *Block[B]) Parent() ids.ID {
	return b.innerBlock.Parent()
}

func (b *Block[B]) Bytes() []byte {
	return b.innerBlock.Bytes()
}

func (b *Block[B]) Height() uint64 {
	return b.innerBlock.Height()
}

func (b *Block[B]) Timestamp() time.Time {
	return b.innerBlock.Timestamp()
}

func (b *Block[B]) Status() choices.Status {
	return b.status
}

func (b *Block[B]) Verify(ctx context.Context) error {
	// Fetch the parent block
	parentBlock, err := b.cache.GetBlock(ctx, b.innerBlock.Parent())
	if err != nil {
		return fmt.Errorf("failed to get parent of %s for verification: %w", b.innerBlock.ID(), err)
	}

	// Verify the block with the backend
	b.decider, err = b.backend.Verify(ctx, parentBlock.(*Block[B]).innerBlock, b.innerBlock)
	if err != nil {
		return err
	}
	if b.decider == nil {
		return fmt.Errorf("block %s returned invalid nil decider during verification", b.ID())
	}

	// Update caches if verification passes
	blkID := b.innerBlock.ID()
	b.cache.unverifiedBlocks.Evict(blkID)
	b.cache.verifiedBlocks[blkID] = b

	return nil
}

func (b *Block[B]) Accept(ctx context.Context) error {
	if err := b.decider.Accept(ctx); err != nil {
		return nil
	}

	b.status = choices.Accepted

	blkID := b.innerBlock.ID()
	b.cache.decidedBlocks.Put(blkID, b)
	delete(b.cache.verifiedBlocks, blkID)

	return nil
}

func (b *Block[B]) Reject(ctx context.Context) error {
	if err := b.decider.Abandon(ctx); err != nil {
		return err
	}

	b.status = choices.Rejected

	blkID := b.innerBlock.ID()
	delete(b.cache.verifiedBlocks, blkID)
	b.cache.decidedBlocks.Put(blkID, b)
	return nil
}
