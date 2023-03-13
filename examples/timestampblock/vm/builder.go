// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"

	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/utils/timer/mockable"
)

type builder struct {
	clock *mockable.Clock

	mempool *mempool
}

func newBuilder(clock *mockable.Clock, mempool *mempool) *builder {
	return &builder{
		clock:   clock,
		mempool: mempool,
	}
}

func (b *builder) BuildBlock(ctx context.Context, parentBlock *Block) (*Block, error) {
	defer b.mempool.NotifyBuildBlock()

	dataHash, err := b.mempool.Pending()
	if err != nil {
		return nil, err
	}
	block := &Block{
		PrntID:   parentBlock.id,
		Hght:     parentBlock.Hght + 1,
		Tmstmp:   b.clock.Time().Unix(),
		DataHash: dataHash,
	}

	bytes, err := Codec.Marshal(0, block)
	if err != nil {
		return nil, err
	}
	block.bytes = bytes
	block.id = hashing.ComputeHash256Array(bytes)

	return block, nil
}
