// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stack

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/cache/metercacher"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/prometheus/client_golang/prometheus"
)

var DefaultBlockCacheConfig = BlockCacheConfig{
	Decided:    1024,
	Unverified: 1024,
	Missing:    1024,
	BytesToID:  1024,
}

type BlockCacheConfig struct {
	Decided    int
	Unverified int
	Missing    int
	BytesToID  int
}

// BlockCache serves as a cache for blocks to serve to the Snowman Consensus Engine
type BlockCache[B StatelessBlock] struct {
	backend BlockBackend[B]

	// verifiedBlocks is a map of blocks that have been verified and are currently in consensus
	verifiedBlocks map[ids.ID]*Block[B]
	// decidedBlocks is a cache of blocks that have been marked as decided
	decidedBlocks cache.Cacher
	// unverifiedBlocks is an LRU cache of blocks with status processing
	// that have not yet passed verification.
	// Every value in [unverifiedBlocks] is a (*Block)
	unverifiedBlocks cache.Cacher
	// missingBlocks is an LRU cache of missing blocks
	// Every value in [missingBlocks] is an empty struct.
	missingBlocks cache.Cacher
	// string([byte repr. of block]) --> the block's ID
	bytesToIDCache cache.Cacher

	lastAcceptedBlock *Block[B]
	preferredBlock    *Block[B]
}

func NewBlockCache[B StatelessBlock](backend BlockBackend[B], lastAcceptedStatelessBlock B, config BlockCacheConfig, registerer prometheus.Registerer) (*BlockCache[B], error) {
	decidedCache, err := metercacher.New(
		"decided_cache",
		registerer,
		&cache.LRU{Size: config.Decided},
	)
	if err != nil {
		return nil, err
	}
	missingCache, err := metercacher.New(
		"missing_cache",
		registerer,
		&cache.LRU{Size: config.Missing},
	)
	if err != nil {
		return nil, err
	}
	unverifiedCache, err := metercacher.New(
		"unverified_cache",
		registerer,
		&cache.LRU{Size: config.Unverified},
	)
	if err != nil {
		return nil, err
	}
	bytesToIDCache, err := metercacher.New(
		"bytes_to_id_cache",
		registerer,
		&cache.LRU{Size: config.BytesToID},
	)
	if err != nil {
		return nil, err
	}

	blockCache := &BlockCache[B]{
		backend:          backend,
		verifiedBlocks:   make(map[ids.ID]*Block[B]),
		decidedBlocks:    decidedCache,
		missingBlocks:    missingCache,
		unverifiedBlocks: unverifiedCache,
		bytesToIDCache:   bytesToIDCache,
	}

	lastAcceptedBlock := &Block[B]{
		// Note: we are guaranteed not to call Verify on lastAcceptedBlock so we can leave parentBlock unpopulated here
		innerBlock: lastAcceptedStatelessBlock,
		cache:      blockCache,
		backend:    backend,
		status:     choices.Accepted,
	}
	blockCache.lastAcceptedBlock = lastAcceptedBlock
	blockCache.preferredBlock = lastAcceptedBlock
	blockCache.decidedBlocks.Put(lastAcceptedBlock.ID(), lastAcceptedBlock)

	return blockCache, nil
}

// Flush each block cache
func (bc *BlockCache[B]) Flush() {
	bc.decidedBlocks.Flush()
	bc.missingBlocks.Flush()
	bc.unverifiedBlocks.Flush()
	bc.bytesToIDCache.Flush()
}

// getCachedBlock checks the caches for [blkID] by priority. Returning
// true if [blkID] is found in one of the caches.
func (bc *BlockCache[B]) getCachedBlock(blkID ids.ID) (snowman.Block, bool) {
	if blk, ok := bc.verifiedBlocks[blkID]; ok {
		return blk, true
	}

	if blk, ok := bc.decidedBlocks.Get(blkID); ok {
		return blk.(snowman.Block), true
	}

	if blk, ok := bc.unverifiedBlocks.Get(blkID); ok {
		return blk.(snowman.Block), true
	}

	return nil, false
}

// GetBlock returns the BlockWrapper as snowman.Block corresponding to [blkID]
func (bc *BlockCache[B]) GetBlock(ctx context.Context, blkID ids.ID) (snowman.Block, error) {
	if blk, ok := bc.getCachedBlock(blkID); ok {
		return blk, nil
	}

	if _, ok := bc.missingBlocks.Get(blkID); ok {
		return nil, database.ErrNotFound
	}

	blk, err := bc.backend.GetBlock(ctx, blkID)
	// If getBlock returns [database.ErrNotFound], State considers
	// this a cacheable miss.
	if err == database.ErrNotFound {
		bc.missingBlocks.Put(blkID, struct{}{})
		return nil, err
	} else if err != nil {
		return nil, err
	}

	// Since this block is not in consensus, addBlockOutsideConsensus
	// is called to add [blk] to the correct cache.
	return bc.addBlockOutsideConsensus(ctx, blk)
}

// ParseBlock attempts to parse [b] into an internal Block and adds it to the appropriate
// caching layer if successful.
func (bc *BlockCache[B]) ParseBlock(ctx context.Context, b []byte) (snowman.Block, error) {
	// See if we've cached this block's ID by its byte repr.
	blkIDIntf, blkIDCached := bc.bytesToIDCache.Get(string(b))
	if blkIDCached {
		blkID := blkIDIntf.(ids.ID)
		// See if we have this block cached
		if cachedBlk, ok := bc.getCachedBlock(blkID); ok {
			return cachedBlk, nil
		}
	}

	// We don't have this block cached by its byte repr.
	// Parse the block from bytes
	blk, err := bc.backend.ParseBlock(ctx, b)
	if err != nil {
		return nil, err
	}
	blkID := blk.ID()
	bc.bytesToIDCache.Put(string(b), blkID)

	// Only check the caches if we didn't do so above
	if !blkIDCached {
		// Check for an existing block, so we can return a unique block
		// if processing or simply allow this block to be immediately
		// garbage collected if it is already cached.
		if cachedBlk, ok := bc.getCachedBlock(blkID); ok {
			return cachedBlk, nil
		}
	}

	bc.missingBlocks.Evict(blkID)

	// Since this block is not in consensus, addBlockOutsideConsensus
	// is called to add [blk] to the correct cache.
	return bc.addBlockOutsideConsensus(ctx, blk)
}

// BuildBlock attempts to build a new internal Block, wraps it, and adds it
// to the appropriate caching layer if successful.
func (bc *BlockCache[B]) BuildBlock(ctx context.Context) (snowman.Block, error) {
	blk, err := bc.backend.BuildBlock(ctx, bc.preferredBlock.innerBlock)
	if err != nil {
		return nil, err
	}

	blkID := blk.ID()
	// Defensive: buildBlock should not return a block that has already been verified.
	// If it does, make sure to return the existing reference to the block.
	if existingBlk, ok := bc.getCachedBlock(blkID); ok {
		return existingBlk, nil
	}
	// Evict the produced block from missing blocks in case it was previously
	// marked as missing.
	bc.missingBlocks.Evict(blkID)

	// wrap the returned block and add it to the correct cache
	return bc.addBlockOutsideConsensus(ctx, blk)
}

// addBlockOutsideConsensus adds [blk] to the correct cache and returns
// a wrapped version of [blk]
// assumes [blk] is a known, non-wrapped block that is not currently
// in consensus. [blk] could be either decided or a block that has not yet
// been verified and added to consensus.
func (bc *BlockCache[B]) addBlockOutsideConsensus(ctx context.Context, blk B) (snowman.Block, error) {
	blkID := blk.ID()
	status, err := bc.getStatus(ctx, blk)
	if err != nil {
		return nil, fmt.Errorf("could not get block status for %s due to %w", blkID, err)
	}
	wrappedBlk := &Block[B]{
		innerBlock: blk,
		cache:      bc,
		backend:    bc.backend,
		status:     status,
	}

	switch status {
	case choices.Accepted, choices.Rejected:
		bc.decidedBlocks.Put(blkID, wrappedBlk)
	case choices.Processing:
		bc.unverifiedBlocks.Put(blkID, wrappedBlk)
	default:
		return nil, fmt.Errorf("found unexpected status for blk %s: %s", blkID, status)
	}

	return wrappedBlk, nil
}

func (bc *BlockCache[B]) LastAccepted(ctx context.Context) (ids.ID, error) {
	return bc.lastAcceptedBlock.ID(), nil
}

// LastAcceptedBlock returns the last accepted wrapped block
func (bc *BlockCache[B]) LastAcceptedBlock() *Block[B] {
	return bc.lastAcceptedBlock
}

func (bc *BlockCache[B]) SetPreference(ctx context.Context, blkID ids.ID) error {
	preferredBlock, err := bc.GetBlock(ctx, blkID)
	if err != nil {
		return fmt.Errorf("failed to get preferred block %s: %w", blkID, err)
	}

	bc.preferredBlock = preferredBlock.(*Block[B])
	return nil
}

// getStatus returns the status of [blk]
func (bc *BlockCache[B]) getStatus(ctx context.Context, blk B) (choices.Status, error) {
	lastAcceptedHeight := bc.lastAcceptedBlock.Height()
	blkHeight := blk.Height()
	if blkHeight > lastAcceptedHeight {
		return choices.Processing, nil
	}

	acceptedID, err := bc.backend.GetBlockIDAtHeight(ctx, blkHeight)
	switch err {
	case nil:
		if acceptedID == blk.ID() {
			return choices.Accepted, nil
		}
		return choices.Rejected, nil
	case database.ErrNotFound:
		return choices.Processing, nil
	default:
		return choices.Unknown, fmt.Errorf("failed to get accepted blkID at height: %d: %w", blkHeight, err)
	}
}
