// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stack

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

var _ ChainVM[StatelessBlock] = (*BlockVMImpl[StatelessBlock, Rooter])(nil)

var (
	// Database prefixes
	wrapperIndexPrefix       = []byte("wrapper")
	vmDBPrefix               = []byte("vm")
	heightPrefix             = []byte("height")
	blockIndexPrefix         = []byte("index")
	acceptedBlockIndexPrefix = []byte("accepted")

	// Database markers
	acceptedKey = []byte("acceptedBlock")
)

type BlockVMImpl[Block StatelessBlock, State Rooter] struct {
	BlockVM[Block, State]

	StateManager StateManager[State]

	wrapperDBManager manager.Manager
	heightIndex      database.Database
	blockIndex       database.Database
	acceptedIndex    database.Database
}

func (vm *BlockVMImpl[Block, State]) Initialize(
	ctx context.Context,
	chainCtx *snow.Context,
	dbManager manager.Manager,
	genesisBytes []byte,
	upgradeBytes []byte,
	configBytes []byte,
	toEngine chan<- common.Message,
	fxs []*common.Fx,
	appSender common.AppSender,
) error {
	vm.wrapperDBManager = dbManager.NewPrefixDBManager(wrapperIndexPrefix)

	db := vm.wrapperDBManager.Current().Database
	vm.blockIndex = prefixdb.New(blockIndexPrefix, db)
	vm.acceptedIndex = prefixdb.New(acceptedBlockIndexPrefix, db)
	vm.heightIndex = prefixdb.New(heightPrefix, db)
	vmDBManager := dbManager.NewPrefixDBManager(vmDBPrefix)
	return vm.BlockVM.Initialize(ctx, chainCtx, vmDBManager, genesisBytes, upgradeBytes, configBytes, toEngine, fxs, appSender)
}

func (vm *BlockVMImpl[Block, State]) LastAccepted(ctx context.Context) (ids.ID, error) {
	blkIDBytes, err := vm.acceptedIndex.Get(acceptedKey)
	if err != nil {
		return ids.ID{}, fmt.Errorf("failed to get last accepted blockID: %w", err)
	}

	blkID, err := ids.ToID(blkIDBytes)
	if err != nil {
		return ids.ID{}, fmt.Errorf("failed to parse last accepted blockID from disk: %w", err)
	}

	return blkID, nil
}

func (vm *BlockVMImpl[Block, State]) GetBlockIDAtHeight(ctx context.Context, height uint64) (ids.ID, error) {
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, height)

	blkIDBytes, err := vm.heightIndex.Get(heightBytes)
	switch {
	case err == database.ErrNotFound:
		return ids.ID{}, err
	case err != nil:
		return ids.ID{}, fmt.Errorf("failed to get height index at %d: %w", height, err)
	}

	blkID, err := ids.ToID(blkIDBytes)
	if err != nil {
		return ids.ID{}, fmt.Errorf("failed to parse blkIDBytes at height %d: %w", height, err)
	}

	return blkID, nil
}

func (vm *BlockVMImpl[Block, State]) GetBlock(ctx context.Context, blkID ids.ID) (Block, error) {
	blkBytes, err := vm.blockIndex.Get(blkID[:])
	if err != nil {
		return *new(Block), fmt.Errorf("failed to get block %s: %w", blkID, err)
	}

	blk, err := vm.ParseBlock(ctx, blkBytes)
	if err != nil {
		return *new(Block), fmt.Errorf("failed to parse block from disk %s: %w", blkID, err)
	}

	return blk, nil
}

// Verify is where the interesting logic happens
func (vm *BlockVMImpl[Block, State]) Verify(ctx context.Context, parent Block, block Block) (Decider, error) {
	// TODO: create interface for either pinning the state in memory or writing it straight to disk
	// ideally leaving the choice to be decided by the underlying VM.
	// TODO: implement TimestampVM on top of the go-ethereum trie database
	state, err := vm.StateManager.OpenState(parent.Root())
	if err != nil {
		return nil, err
	}

	underlyingDecider, err := vm.BlockVM.Execute(ctx, parent, block, state)
	if err != nil {
		return nil, err
	}

	return &blockDecider[Block, State]{
		underlyingDecider: underlyingDecider,
		vm:                vm,
		block:             block,
	}, nil
}

type blockDecider[Block StatelessBlock, State Rooter] struct {
	underlyingDecider Decider
	block             Block
	vm                *BlockVMImpl[Block, State]
}

func (bd *blockDecider[Block, State]) Accept(ctx context.Context) error {
	if err := bd.underlyingDecider.Accept(ctx); err != nil {
		return err
	}

	// TODO: this code can be broken out into a separate accepted block indexing component
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, bd.block.Height())

	blkID := bd.block.ID()
	if err := bd.vm.heightIndex.Put(heightBytes, blkID[:]); err != nil {
		return fmt.Errorf("failed to put block %s into height index: %w", blkID, err)
	}

	if err := bd.vm.blockIndex.Put(blkID[:], bd.block.Bytes()); err != nil {
		return fmt.Errorf("failed to put block %s into block index: %w", blkID, err)
	}

	if err := bd.vm.acceptedIndex.Put(acceptedKey, blkID[:]); err != nil {
		return fmt.Errorf("failed to update last accepted block to %s: %w", blkID, err)
	}
	return nil
}

func (bd *blockDecider[Block, State]) Abandon(ctx context.Context) error {
	return bd.underlyingDecider.Abandon(ctx)
}
