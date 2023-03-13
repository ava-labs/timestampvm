// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/utils/hashing"
	avalancheJSON "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/utils/timer/mockable"
	avalancheRPC "github.com/gorilla/rpc/v2"
)

// Name/Version
var (
	Name    = "TimestampChainVM"
	Version = "v0.0.1"
	ID      = ids.ID{'t', 'i', 'm', 'e', 's', 't', 'a', 'm', 'p'}
)

// Type assertions
var (
// _ stack.ChainBackend[*Block] = (*VM)(nil)
// _ stack.ChainVM[*Block]      = (*VM)(nil)
)

var futureBlockLimit = time.Minute // Maximum amount of time that a block can be in the future

type VM struct {
	// Clock used for block building and verification
	clock mockable.Clock

	state database.Database

	BlockExecutor

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
	vm.state = dbManager.Current().Database

	vm.mempool = newMempool(toEngine)
	vm.builder = newBuilder(&vm.clock, vm.mempool)

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

	decider, err := vm.ExecuteStateChanges(ctx, nil, genesisBlock, vm.state)
	if err != nil {
		return err
	}
	if err := decider.Accept(ctx); err != nil {
		return err
	}
	return nil
}

func (vm *VM) ParseBlock(ctx context.Context, b []byte) (*Block, error) {
	return ParseBlock(ctx, b)
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
