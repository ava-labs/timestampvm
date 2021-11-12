// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"errors"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
)

const (
	blockCacheSize = 8192
)

var (
	errBlockWrongVersion = errors.New("wrong version")

	_ BlockState = &blockState{}
)

type BlockState interface {
	GetBlock(blkID ids.ID) (Block, error)
	PutBlock(blk Block) error

	GetLastAccepted() ids.ID
	SetLastAccepted(ids.ID)

	ClearCache()
}

type blockState struct {
	blkCache cache.Cacher
	blockDB  database.Database

	lastAccepted ids.ID
}

func NewBlockState(db database.Database) BlockState {
	return &blockState{
		blkCache: &cache.LRU{Size: blockCacheSize},
		blockDB:  db,
	}
}

func (s *blockState) GetBlock(blkID ids.ID) (Block, error) {
	blkBytes, err := s.blockDB.Get(blkID[:])
	if err != nil {
		return nil, err
	}

	blk := TimeBlock{}
	parsedVersion, err := Codec.Unmarshal(blkBytes, &blk)
	if err != nil {
		return nil, err
	}

	if parsedVersion != codecVersion {
		return nil, errBlockWrongVersion
	}

	s.blkCache.Put(blkID, blk)

	return &blk, nil
}

func (s *blockState) PutBlock(blk Block) error {
	bytes, err := Codec.Marshal(codecVersion, &blk)
	if err != nil {
		return err
	}

	blkID := blk.ID()
	s.blkCache.Put(blkID, &blk)
	return s.blockDB.Put(blkID[:], bytes)
}

func (s *blockState) DeleteBlock(blkID ids.ID) error {
	s.blkCache.Put(blkID, nil)
	return s.blockDB.Delete(blkID[:])
}

func (s *blockState) GetLastAccepted() ids.ID             { return s.lastAccepted }
func (s *blockState) SetLastAccepted(lastAccepted ids.ID) { s.lastAccepted = lastAccepted }

func (s *blockState) ClearCache() {
	s.blkCache.Flush()
}
