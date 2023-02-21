// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/timestampvm/sdk/stack"
)

// Type assertions
var (
	_ stack.StatelessBlock = (*Block)(nil)
	_ stack.Decider        = (*blockDecider)(nil)
)

// Block defines a stateless block
type Block struct {
	PrntID   ids.ID `serialize:"true" json:"parentID"`  // parent's ID
	Hght     uint64 `serialize:"true" json:"height"`    // This block's height. The genesis block is at height 0.
	Tmstmp   int64  `serialize:"true" json:"timestamp"` // Time this block was proposed at
	DataHash ids.ID `serialize:"true" json:"dataHash"`  // hash of some arbitrary data to timestamp

	id    ids.ID // hold this block's ID
	bytes []byte // this block's encoded bytes
}

// ID returns the ID of this block
func (b *Block) ID() ids.ID { return b.id }

// ParentID returns [b]'s parent's ID
func (b *Block) Parent() ids.ID { return b.PrntID }

// Height returns this block's height. The genesis block has height 0.
func (b *Block) Height() uint64 { return b.Hght }

// Timestamp returns this block's time. The genesis block has time 0.
func (b *Block) Timestamp() time.Time { return time.Unix(b.Tmstmp, 0) }

// Bytes returns the byte repr. of this block
func (b *Block) Bytes() []byte { return b.bytes }

// blockDecider implements the stack.Decider interface
type blockDecider struct {
	*Block

	vm *VM
}

func (d *blockDecider) Accept(ctx context.Context) error {
	return d.vm.acceptBlock(d.Block)
}

// Abandon is a no-op since there is nothing to clean up on Abandon
func (d *blockDecider) Abandon(ctx context.Context) error {
	return nil
}
