// (c) 2021, Ava Labs, Inc. All rights reserved.
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
	blockCacheSize = 8192
)

var lastAcceptedKey = []byte{lastAcceptedByte}

var _ BlockState = &blockState{}

type BlockState interface {
	GetBlock(blkID ids.ID) (Block, error)
	PutBlock(blk Block) error

	GetLastAccepted() (ids.ID, error)
	SetLastAccepted(ids.ID) error

	ClearCache()
}

type blockState struct {
	blkCache cache.Cacher
	blockDB  database.Database
	vm       *VM

	lastAccepted ids.ID
}

type blkWrapper struct {
	Blk    []byte         `serialize:"true"`
	Status choices.Status `serialize:"true"`

	block Block
}

func NewBlockState(db database.Database, vm *VM) BlockState {
	return &blockState{
		blkCache: &cache.LRU{Size: blockCacheSize},
		blockDB:  db,
		vm:       vm,
	}
}

func (s *blockState) GetBlock(blkID ids.ID) (Block, error) {
	if blkIntf, cached := s.blkCache.Get(blkID); cached {
		if blkIntf == nil {
			return nil, database.ErrNotFound
		}
		return blkIntf.(Block), nil
	}

	blkBytes, err := s.blockDB.Get(blkID[:])
	if err != nil {
		if err == database.ErrNotFound {
			s.blkCache.Put(blkID, nil)
		}
		return nil, err
	}

	blkw := blkWrapper{}
	if _, err := Codec.Unmarshal(blkBytes, &blkw); err != nil {
		return nil, err
	}

	var blk Block
	if _, err := Codec.Unmarshal(blkw.Blk, &blk); err != nil {
		return nil, err
	}

	blk.Initialize(blkw.Blk, blkw.Status, s.vm)

	s.blkCache.Put(blkID, blk)

	return blk, nil
}

func (s *blockState) PutBlock(blk Block) error {
	blkw := blkWrapper{
		Blk:    blk.Bytes(),
		Status: blk.Status(),
		block:  blk,
	}

	bytes, err := Codec.Marshal(CodecVersion, &blkw)
	if err != nil {
		return err
	}

	blkID := blk.ID()
	s.blkCache.Put(blkID, blk)

	return s.blockDB.Put(blkID[:], bytes)
}

func (s *blockState) DeleteBlock(blkID ids.ID) error {
	s.blkCache.Put(blkID, nil)
	return s.blockDB.Delete(blkID[:])
}

func (s *blockState) GetLastAccepted() (ids.ID, error) {
	if s.lastAccepted != ids.Empty {
		return s.lastAccepted, nil
	}

	lastAcceptedBytes, err := s.blockDB.Get(lastAcceptedKey)
	if err != nil {
		return ids.ID{}, err
	}
	lastAccepted, err := ids.ToID(lastAcceptedBytes)
	if err != nil {
		return ids.ID{}, err
	}
	s.lastAccepted = lastAccepted
	return lastAccepted, nil
}

func (s *blockState) SetLastAccepted(lastAccepted ids.ID) error {
	if s.lastAccepted == lastAccepted {
		return nil
	}
	s.lastAccepted = lastAccepted
	return s.blockDB.Put(lastAcceptedKey, lastAccepted[:])
}

func (s *blockState) ClearCache() {
	s.blkCache.Flush()
}
