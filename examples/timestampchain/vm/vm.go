// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	commonEng "github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/utils/timer/mockable"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/timestampvm/sdk/stack"
)

// Name/Version
var (
	Name    string = "TimestampChainVM"
	Version string = "v0.0.1"
)

// Type assertions
var (
	_ stack.BlockBackend[*Block] = (*VM)(nil)
	_ stack.VMBackend[*Block]    = (*VM)(nil)
)

var (
	// Database prefixes
	heightPrefix   = []byte("height")
	blockPrefix    = []byte("block")
	acceptedPrefix = []byte("accepted")
	statePrefix    = []byte("state")

	// Database markers
	acceptedKey = []byte("acceptedBlock")

	futureBlockLimit = time.Minute // Maximum amount of time that a block can be in the future
)

type VM struct {
	// Clock used for block building and verification
	clock mockable.Clock

	// State management
	vDB           *versiondb.Database
	heightIndex   database.Database
	blockIndex    database.Database
	acceptedIndex database.Database
	state         database.Database

	// Mempool
	mempool *mempool
}

// Initialize implements the snowman.ChainVM interface
func (vm *VM) Initialize(
	ctx context.Context,
	_ *snow.Context,
	dbManager manager.Manager,
	genesisBytes []byte,
	_ []byte,
	_ []byte,
	toEngine chan<- commonEng.Message,
	_ []*commonEng.Fx,
	appSender commonEng.AppSender,
) error {
	vm.vDB = versiondb.New(dbManager.Current().Database)
	vm.heightIndex = prefixdb.New(heightPrefix, vm.vDB)
	vm.blockIndex = prefixdb.New(blockPrefix, vm.vDB)
	vm.acceptedIndex = prefixdb.New(acceptedPrefix, vm.vDB)
	vm.state = prefixdb.New(statePrefix, vm.vDB)

	vm.mempool = NewMempool()

	return vm.initGenesis(ctx, genesisBytes)
}

func (vm *VM) initGenesis(ctx context.Context, genesisBytes []byte) error {
	genesisBlock, err := vm.ParseBlock(ctx, genesisBytes)
	if err != nil {
		return fmt.Errorf("failed to parse genesis block bytes: %w", err)
	}

	if genesisBlock.Hght != 0 {
		return fmt.Errorf("cannot use genesis block with height: %d", genesisBlock.Hght)
	}
	if genesisBlock.PrntID != ids.Empty {
		return fmt.Errorf("cannot use genesis block with non-empty parentID: %s", genesisBlock.PrntID)
	}

	genesisBlkID, err := vm.GetBlockIDAtHeight(ctx, 0)
	// If the block on disk matches what we parsed, return early
	if err == nil && genesisBlkID == genesisBlock.id {
		return nil
	}

	switch {
	case err == nil && genesisBlkID == genesisBlock.id:
		return nil
	case errors.Is(err, database.ErrNotFound):
		if err := vm.putBlock(genesisBlock); err != nil {
			return fmt.Errorf("failed to put genesis block: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("failed to get blockID for genesis: %w", err)
	}
}

func (vm *VM) ParseBlock(ctx context.Context, b []byte) (*Block, error) {
	// A new empty block
	block := &Block{}

	// Unmarshal the byte repr. of the block into our empty block
	_, err := Codec.Unmarshal(b, block)
	if err != nil {
		return nil, err
	}

	block.id = hashing.ComputeHash256Array(b)
	block.bytes = b

	return block, nil
}

// BuildBlock builds a block out of the necessary components
func (vm *VM) BuildBlock(ctx context.Context, parentBlock *Block) (*Block, error) {
	block := &Block{
		PrntID:   parentBlock.id,
		Hght:     parentBlock.Hght + 1,
		Tmstmp:   vm.clock.Time().Unix(),
		DataHash: ids.GenerateTestID(),
	}

	bytes, err := Codec.Marshal(0, block)
	if err != nil {
		return nil, err
	}
	block.bytes = bytes
	block.id = hashing.ComputeHash256Array(bytes)

	return block, nil
}

func (vm *VM) putBlock(block *Block) error {
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, block.Height())

	if err := vm.heightIndex.Put(heightBytes, block.id[:]); err != nil {
		return fmt.Errorf("failed to put block %s into height index: %w", block.ID(), err)
	}

	if err := vm.blockIndex.Put(block.id[:], block.bytes); err != nil {
		return fmt.Errorf("failed to put block %s into block index: %w", block.ID(), err)
	}

	return nil
}

func (vm *VM) GetBlockIDAtHeight(ctx context.Context, height uint64) (ids.ID, error) {
	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, height)

	blkIDBytes, err := vm.heightIndex.Get(heightBytes)
	if err != nil {
		return ids.ID{}, fmt.Errorf("failed to get height index at %d: %w", height, err)
	}

	blkID, err := ids.ToID(blkIDBytes)
	if err != nil {
		return ids.ID{}, fmt.Errorf("failed to parse blkIDBytes at height %d: %w", height, err)
	}

	return blkID, nil
}

func (vm *VM) GetBlock(ctx context.Context, blkID ids.ID) (*Block, error) {
	blkBytes, err := vm.blockIndex.Get(blkID[:])
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %w", blkID, err)
	}

	blk, err := vm.ParseBlock(ctx, blkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block from disk %s: %w", blkID, err)
	}

	return blk, nil
}

// Verify verifies that [block] to be added to consensus with the given parent block.
// This function assumes that [parentBlock] is guaranteed to be the actual parent of
// [block].
func (vm *VM) Verify(ctx context.Context, parent *Block, block *Block) error {
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

// Accept marks [block] as accepted and performs all DB IO necessary on accept.
func (vm *VM) Accept(ctx context.Context, block *Block) error {
	defer vm.vDB.Abort()

	if err := vm.putBlock(block); err != nil {
		return err
	}

	// Add timestamp to the current state
	timestampBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(timestampBytes, uint64(block.Tmstmp))
	if err := vm.state.Put(timestampBytes, block.DataHash[:]); err != nil {
		return fmt.Errorf("failed to put timestamped hash in database for block %s: %w", block.id, err)
	}

	if err := vm.acceptedIndex.Put(acceptedKey, block.id[:]); err != nil {
		return fmt.Errorf("failed to update last accepted block to %s: %w", block.id, err)
	}

	if err := vm.vDB.Commit(); err != nil {
		return fmt.Errorf("failed to commit database accepting block %s: %w", block.id, err)
	}

	return nil
}

// Reject is called by the engine when a block is marked as rejected.
// TimestampVM does not need to perform any cleanup on Reject, since there is no garbage collection
// necessary from Verify.
func (vm *VM) Reject(ctx context.Context, block *Block) error {
	return nil
}

func (vm *VM) LastAccepted(ctx context.Context) (ids.ID, error) {
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

func (vm *VM) HealthCheck(ctx context.Context) (interface{}, error) {
	return nil, nil
}

// SetState communicates to VM its next state it starts
func (vm *VM) SetState(ctx context.Context, state snow.State) error {
	return nil
}

// Shutdown is called when the node is shutting down.
func (vm *VM) Shutdown(ctx context.Context) error {
	return nil
}

// Version returns the version of the VM.
func (vm *VM) Version(ctx context.Context) (string, error) {
	return Version, nil
}

func (vm *VM) CreateStaticHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	return nil, nil
}

func (vm *VM) CreateHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	return nil, nil
}
