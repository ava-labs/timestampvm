// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stack

import (
	"context"

	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/prometheus/client_golang/prometheus"
)

// Type assertions
var _ block.ChainVM = (*VM[StatelessBlock])(nil)

type VM[Block StatelessBlock] struct {
	chainCtx *snow.Context

	ChainVM VMBackend[Block]

	*BlockCache[Block]
	*Network
}

func (vm *VM[B]) Initialize(
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
	vm.chainCtx = chainCtx
	if err := vm.ChainVM.Initialize(
		ctx,
		chainCtx,
		dbManager,
		genesisBytes,
		upgradeBytes,
		configBytes,
		toEngine,
		fxs,
		appSender,
	); err != nil {
		return err
	}

	// Initialize chain
	lastAcceptedID, err := vm.ChainVM.LastAccepted(ctx)
	if err != nil {
		return err
	}
	lastAcceptedBlock, err := vm.ChainVM.GetBlock(ctx, lastAcceptedID)
	if err != nil {
		return err
	}

	blockCacheRegistry := prometheus.NewRegistry()
	vm.BlockCache, err = NewBlockCache[B](chainCtx, vm.ChainVM, lastAcceptedBlock, DefaultBlockCacheConfig, blockCacheRegistry)
	if err != nil {
		return err
	}

	return nil
}

func (vm *VM[B]) HealthCheck(ctx context.Context) (interface{}, error) {
	return vm.ChainVM.HealthCheck(ctx)
}

// Connector represents a handler that is called on connection connect/disconnect
// validators.Connector

// SetState communicates to VM its next state it starts
func (vm *VM[B]) SetState(ctx context.Context, state snow.State) error {
	return vm.ChainVM.SetState(ctx, state)
}

// Shutdown is called when the node is shutting down.
func (vm *VM[B]) Shutdown(ctx context.Context) error {
	vm.BlockCache.Shutdown(ctx)
	return vm.ChainVM.Shutdown(ctx)
}

// Version returns the version of the VM.
func (vm *VM[B]) Version(ctx context.Context) (string, error) {
	return vm.ChainVM.Version(ctx)
}

func (vm *VM[B]) CreateStaticHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	return vm.ChainVM.CreateStaticHandlers(ctx)
}

func (vm *VM[B]) CreateHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	return vm.ChainVM.CreateHandlers(ctx)
}
