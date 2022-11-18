// (c) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
)

const (
	lastAcceptedByte byte = iota
)

const (
	// maximum block capacity of the cache
	blockCacheSize = 8192
)

// persists lastAccepted block IDs with this key
var lastAcceptedKey = []byte{lastAcceptedByte}

var _ BlockState = &blockState{}

// BlockState defines methods to manage state with Blocks and LastAcceptedIDs.
type BlockState interface {
	GetBlock(blkID ids.ID) (*Block, error)
	PutBlock(blk *Block) error
	GetLastAccepted() (ids.ID, error)
	SetLastAccepted(ids.ID) error
}

// blockState implements BlocksState interface with database and cache.
type blockState struct {
	// cache to store blocks
	blkCache cache.Cacher
	// block database
	blockDB      database.Database
	lastAccepted ids.ID

	// vm reference
	vm *VM
}

// NewBlockState returns BlockState with a new cache and given db
func NewBlockState(db database.Database, vm *VM) BlockState {
	return &blockState{
		blkCache: &cache.LRU{Size: blockCacheSize},
		blockDB:  db,
		vm:       vm,
	}
}

// GetBlock gets Block from either cache or database
func (s *blockState) GetBlock(blkID ids.ID) (*Block, error) {
	// Check if cache has this blkID
	if blkIntf, cached := s.blkCache.Get(blkID); cached {
		// there is a key but value is nil, so return an error
		if blkIntf == nil {
			return nil, database.ErrNotFound
		}
		// We found it return the block in cache
		return blkIntf.(*Block), nil
	}

	// get block bytes from db with the blkID key
	bytes, err := s.blockDB.Get(blkID[:])
	if err != nil {
		// we could not find it in the db, let's cache this blkID with nil value
		// so next time we try to fetch the same key we can return error
		// without hitting the database
		if err == database.ErrNotFound {
			s.blkCache.Put(blkID, nil)
		}
		// could not find the block, return error
		return nil, err
	}

	// decode/unmarshal the block bytes to block
	blk, err := UnmarshalBlock(bytes)
	if err != nil {
		return nil, err
	}

	// initialize block with block bytes, status and vm
	blk.Initialize(bytes, choices.Accepted, s.vm)

	// put block into cache
	s.blkCache.Put(blkID, blk)

	return blk, nil
}

// PutBlock puts block into both database and cache
func (s *blockState) PutBlock(blk *Block) error {
	// encode block to its byte representation
	bytes := MarshalBlock(blk)

	blkID := blk.ID()
	// put actual block to cache, so we can directly fetch it from cache
	s.blkCache.Put(blkID, blk)

	// put wrapped block bytes into database
	return s.blockDB.Put(blkID[:], bytes)
}

// DeleteBlock deletes block from both cache and database
func (s *blockState) DeleteBlock(blkID ids.ID) error {
	s.blkCache.Put(blkID, nil)
	return s.blockDB.Delete(blkID[:])
}

// GetLastAccepted returns last accepted block ID
func (s *blockState) GetLastAccepted() (ids.ID, error) {
	// check if we already have lastAccepted ID in state memory
	if s.lastAccepted != ids.Empty {
		return s.lastAccepted, nil
	}

	// get lastAccepted bytes from database with the fixed lastAcceptedKey
	lastAcceptedBytes, err := s.blockDB.Get(lastAcceptedKey)
	if err != nil {
		return ids.ID{}, err
	}
	// parse bytes to ID
	lastAccepted, err := ids.ToID(lastAcceptedBytes)
	if err != nil {
		return ids.ID{}, err
	}
	// put lastAccepted ID into memory
	s.lastAccepted = lastAccepted
	return lastAccepted, nil
}

// SetLastAccepted persists lastAccepted ID into both cache and database
func (s *blockState) SetLastAccepted(lastAccepted ids.ID) error {
	// if the ID in memory and the given memory are same don't do anything
	if s.lastAccepted == lastAccepted {
		return nil
	}
	// put lastAccepted ID to memory
	s.lastAccepted = lastAccepted
	// persist lastAccepted ID to database with fixed lastAcceptedKey
	return s.blockDB.Put(lastAcceptedKey, lastAccepted[:])
}
