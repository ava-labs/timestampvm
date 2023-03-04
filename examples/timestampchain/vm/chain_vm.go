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
	"github.com/ava-labs/avalanchego/utils/hashing"
	avalancheJSON "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/utils/timer/mockable"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/timestampvm/sdk/stack"
	avalancheRPC "github.com/gorilla/rpc/v2"
)

// Name/Version
var (
	Name    string = "TimestampChainVM"
	Version string = "v0.0.1"
	ID             = ids.ID{'t', 'i', 'm', 'e', 's', 't', 'a', 'm', 'p'}
)

// Type assertions
var (
	_ stack.ChainBackend[*Block] = (*VM)(nil)
	_ stack.ChainVM[*Block]      = (*VM)(nil)
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

	mempool *mempool
	*builder
}

// Initialize implements the snowman.ChainVM interface
func (vm *VM) Initialize(
	ctx context.Context,
	_ *snow.Context,
	dbManager manager.Manager,
	genesisBytes []byte,
	_ []byte,
	_ []byte,
	toEngine chan<- common.Message,
	_ []*common.Fx,
	appSender common.AppSender,
) error {
	vm.vDB = versiondb.New(dbManager.Current().Database)
	vm.heightIndex = prefixdb.New(heightPrefix, vm.vDB)
	vm.blockIndex = prefixdb.New(blockPrefix, vm.vDB)
	vm.acceptedIndex = prefixdb.New(acceptedPrefix, vm.vDB)
	vm.state = prefixdb.New(statePrefix, vm.vDB)

	vm.mempool = NewMempool(toEngine)
	vm.builder = NewBuilder(&vm.clock, vm.mempool)

	return vm.initGenesis(ctx, genesisBytes)
}

func (vm *VM) initGenesis(ctx context.Context, genesisBytes []byte) error {
	genesisDataHash, err := ids.ToID(genesisBytes)
	if err != nil {
		return fmt.Errorf("failed to convert supplied genesis bytes to data hash: %w", err)
	}

	genesisBlock := &Block{
		PrntID:   ids.Empty,
		Hght:     0,
		Tmstmp:   0,
		DataHash: genesisDataHash,
	}

	bytes, err := Codec.Marshal(0, genesisBlock)
	if err != nil {
		return fmt.Errorf("failed to marshal genesis block: %w", err)
	}
	genesisBlock.bytes = bytes
	genesisBlock.id = hashing.ComputeHash256Array(bytes)

	genesisBlkID, err := vm.GetBlockIDAtHeight(ctx, 0)
	switch {
	case err == nil && genesisBlkID == genesisBlock.id: // If the block on disk matches what we parsed, return early
		return nil
	case errors.Is(err, database.ErrNotFound):
		if err := vm.Accept(genesisBlock); err != nil {
			return fmt.Errorf("failed to put genesis block: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("failed to get blockID for genesis: %w", err)
	}
}

func (vm *VM) ParseBlock(ctx context.Context, b []byte) (*Block, error) {
	return ParseBlock(ctx, b)
}

func (vm *VM) Accept(block *Block) error {
	defer vm.vDB.Abort()

	heightBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(heightBytes, block.Height())

	if err := vm.heightIndex.Put(heightBytes, block.id[:]); err != nil {
		return fmt.Errorf("failed to put block %s into height index: %w", block.ID(), err)
	}

	if err := vm.blockIndex.Put(block.id[:], block.bytes); err != nil {
		return fmt.Errorf("failed to put block %s into block index: %w", block.ID(), err)
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

func (vm *VM) GetBlockIDAtHeight(ctx context.Context, height uint64) (ids.ID, error) {
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
func (vm *VM) Verify(ctx context.Context, parent *Block, block *Block) (stack.Decider, error) {
	// Ensure [b]'s height comes right after its parent's height
	if expectedHeight := parent.Height() + 1; expectedHeight != block.Hght {
		return nil, fmt.Errorf(
			"expected block to have height %d, but found %d",
			expectedHeight,
			block.Hght,
		)
	}

	// Ensure [b]'s timestamp is >= its parent's timestamp.
	if block.Timestamp().Unix() < parent.Timestamp().Unix() {
		return nil, fmt.Errorf("block cannot have timestamp (%s) < parent timestamp (%s)", block.Timestamp(), parent.Timestamp())
	}

	// Ensure [b]'s timestamp is not more than an hour
	// ahead of this node's time
	if block.Timestamp().Unix() >= time.Now().Add(futureBlockLimit).Unix() {
		return nil, fmt.Errorf("block cannot have timestamp (%s) further than (%s) past current time (%s)", block.Timestamp(), futureBlockLimit, time.Now())
	}

	return &chainDecider{
		Block:    block,
		acceptor: vm,
	}, nil
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
	server := avalancheRPC.NewServer()
	server.RegisterCodec(avalancheJSON.NewCodec(), "application/json")
	server.RegisterCodec(avalancheJSON.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(&Service{vm: vm}, "timestamp"); err != nil {
		return nil, err
	}

	handlers := map[string]*common.HTTPHandler{
		"/timestamp": {LockOptions: common.NoLock, Handler: server},
	}
	return handlers, nil
}
