// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/timestampvm/sdk/stack"
)

var noopDecider = stack.NewNoopDecider()

// BlockExecutor bundles all of the functionality necessary to execute [Block] on a generic state definition.
// In this case, [*Block] and [database.Database]
type BlockExecutor struct{}

func (b *BlockExecutor) SyntacticVerify(ctx context.Context, parent *Block, block *Block) error {
	// Ensure [b]'s height comes right after its parent's height
	if expectedHeight := parent.Height() + 1; expectedHeight != block.Hght {
		return fmt.Errorf(
			"expected block to have height %d, but found %d",
			expectedHeight,
			block.Hght,
		)
	}

	// Ensure [b]'s timestamp is >= its parent's timestamp.
	if block.Timestamp().Unix() < parent.Timestamp().Unix() {
		return fmt.Errorf("block cannot have timestamp (%s) < parent timestamp (%s)", block.Timestamp(), parent.Timestamp())
	}

	// Ensure [b]'s timestamp is not more than an hour
	// ahead of this node's time
	if block.Timestamp().Unix() >= time.Now().Add(futureBlockLimit).Unix() {
		return fmt.Errorf("block cannot have timestamp (%s) further than (%s) past current time (%s)", block.Timestamp(), futureBlockLimit, time.Now())
	}

	return nil
}

func (b *BlockExecutor) ExecuteStateChanges(ctx context.Context, parent *Block, block *Block, state database.Database) (stack.Decider, error) {
	timestampBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(timestampBytes, uint64(block.Tmstmp))
	if err := state.Put(timestampBytes, block.DataHash[:]); err != nil {
		return nil, fmt.Errorf("failed to put timestamped hash in database for block %s: %w", block.id, err)
	}
	return noopDecider, nil
}

// Execute executes verifies and executes [block]
func (b *BlockExecutor) Execute(ctx context.Context, parent *Block, block *Block, state database.Database) (stack.Decider, error) {
	if err := b.SyntacticVerify(ctx, parent, block); err != nil {
		return nil, err
	}

	return b.ExecuteStateChanges(ctx, parent, block, state)
}
