// (c) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/version"
	"github.com/stretchr/testify/require"
)

var blockchainID = ids.ID{1, 2, 3}

func newTestVM(genesisDataHash ids.ID) (*VM, *Service, chan common.Message, error) {
	dbManager := manager.NewMemDB(&version.Semantic{
		Major: 1,
		Minor: 0,
		Patch: 0,
	})
	msgChan := make(chan common.Message, 1)
	vm := &VM{}
	snowCtx := snow.DefaultContextTest()
	snowCtx.ChainID = blockchainID
	err := vm.Initialize(context.Background(), snowCtx, dbManager, genesisDataHash[:], nil, nil, msgChan, nil, nil)
	return vm, &Service{vm: vm}, msgChan, err
}

func TestGenesis(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	// Initialize the vm
	genesisDataHash := ids.ID{1, 2, 3, 4, 5}
	vm, _, _, err := newTestVM(genesisDataHash)
	require.NoError(err)

	// Get lastAccepted
	lastAccepted, err := vm.LastAccepted(ctx)
	require.NoError(err)
	require.NotEqual(ids.Empty, lastAccepted)

	genesisBlock, err := vm.GetBlock(ctx, lastAccepted)
	require.NoError(err)
	require.Equal(genesisDataHash, genesisBlock.DataHash)

	genesisBlockID, err := vm.GetBlockIDAtHeight(ctx, 0)
	require.NoError(err)
	require.Equal(lastAccepted, genesisBlockID)
}

func TestBuildBlock(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	// Initialize the vm
	genesisDataHash := ids.ID{1}
	vm, service, toEngine, err := newTestVM(genesisDataHash)
	require.NoError(err)

	// Get lastAccepted
	lastAccepted, err := vm.LastAccepted(ctx)
	require.NoError(err)
	require.NotEqual(ids.Empty, lastAccepted)

	genesisBlock, err := vm.GetBlock(ctx, lastAccepted)
	require.NoError(err)
	dataHash := ids.ID{2}
	require.NoError(service.ProposeBlock(nil, &BlockIDArgs{ID: dataHash}, &api.EmptyReply{}))
	<-toEngine

	block, err := vm.BuildBlock(ctx, genesisBlock)
	require.NoError(err)
	require.Equal(dataHash, block.DataHash)

	decider, err := vm.Verify(ctx, genesisBlock, block)
	require.NoError(err)
	require.NoError(decider.Accept(ctx))

	blockID, err := vm.GetBlockIDAtHeight(ctx, 1)
	require.NoError(err)
	require.Equal(block.ID(), blockID)

	fetchedBlockByID := new(Block)
	require.NoError(service.GetBlock(nil, &BlockIDArgs{blockID}, fetchedBlockByID))
	require.Equal(block, fetchedBlockByID)

	fetchedLastAcceptedBlock := new(Block)
	require.NoError(service.GetBlock(nil, &BlockIDArgs{ids.Empty}, fetchedLastAcceptedBlock))
	require.Equal(block, fetchedLastAcceptedBlock)
}
