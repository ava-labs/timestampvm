// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"testing"

	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/version"
	"github.com/stretchr/testify/assert"
)

var blockchainID = ids.ID{1, 2, 3}

// Assert that after initialization, the vm has the state we expect
func TestGenesis(t *testing.T) {
	// Initialize the vm
	dbManager := manager.NewMemDB(version.DefaultVersion1_0_0)
	msgChan := make(chan common.Message, 1)
	vm := &VM{}
	ctx := snow.DefaultContextTest()
	ctx.ChainID = blockchainID
	assert := assert.New(t)
	assert.NoError(vm.Initialize(ctx, dbManager, []byte{0, 0, 0, 0, 0}, nil, nil, msgChan, nil, nil))
	// Verify that the db is initialized
	ok, err := vm.state.IsInitialized()
	assert.NoError(err)
	assert.True(ok)

	// Get lastAccepted
	lastAccepted, err := vm.LastAccepted()
	assert.NoError(err)
	assert.NotEqual(ids.Empty, lastAccepted)

	// Verify that getBlock returns the genesis block, and the genesis block
	// is the type we expect
	genesisBlock, err := vm.getBlock(lastAccepted) // genesisBlock as snowman.Block
	assert.NoError(err)

	// Verify that the genesis block has the data we expect
	assert.Equal(ids.Empty, genesisBlock.Parent())
	assert.Equal([32]byte{0, 0, 0, 0, 0}, genesisBlock.Data())
}

func TestHappyPath(t *testing.T) {
	// Initialize the vm
	dbManager := manager.NewMemDB(version.DefaultVersion1_0_0)
	msgChan := make(chan common.Message, 1)
	vm := &VM{}
	ctx := snow.DefaultContextTest()
	ctx.ChainID = blockchainID
	assert := assert.New(t)
	assert.NoError(vm.Initialize(ctx, dbManager, []byte{0, 0, 0, 0, 0}, nil, nil, msgChan, nil, nil))

	lastAcceptedID, err := vm.LastAccepted()
	assert.NoError(err)
	genesisBlock, err := vm.getBlock(lastAcceptedID)
	assert.NoError(err)

	// in an actual execution, the engine would set the preference
	assert.NoError(vm.SetPreference(genesisBlock.ID()))

	ctx.Lock.Lock()
	vm.proposeBlock([dataLen]byte{0, 0, 0, 0, 1}) // propose a value
	ctx.Lock.Unlock()

	select { // assert there is a pending tx message to the engine
	case msg := <-msgChan:
		assert.Equal(common.PendingTxs, msg)
	default:
		assert.FailNow("should have been pendingTxs message on channel")
	}

	// build the block
	ctx.Lock.Lock()
	snowmanBlock2, err := vm.BuildBlock()
	assert.NoError(err)

	assert.NoError(snowmanBlock2.Verify())
	assert.NoError(snowmanBlock2.Accept())
	assert.NoError(vm.SetPreference(snowmanBlock2.ID()))

	lastAcceptedID, err = vm.LastAccepted()
	assert.NoError(err)

	// Should be the block we just accepted
	block2, err := vm.getBlock(lastAcceptedID)
	assert.NoError(err)

	// Assert the block we accepted has the data we expect
	assert.Equal(genesisBlock.ID(), block2.Parent())
	assert.Equal([dataLen]byte{0, 0, 0, 0, 1}, block2.Data())
	assert.Equal(snowmanBlock2.ID(), block2.ID())
	assert.NoError(block2.Verify())

	vm.proposeBlock([dataLen]byte{0, 0, 0, 0, 2}) // propose a block
	ctx.Lock.Unlock()

	select { // verify there is a pending tx message to the engine
	case msg := <-msgChan:
		assert.Equal(common.PendingTxs, msg)
	default:
		assert.FailNow("should have been pendingTxs message on channel")
	}

	ctx.Lock.Lock()

	// build the block
	snowmanBlock3, err := vm.BuildBlock()
	assert.NoError(err)
	assert.NoError(snowmanBlock3.Verify())
	assert.NoError(snowmanBlock3.Accept())
	assert.NoError(vm.SetPreference(snowmanBlock3.ID()))

	lastAcceptedID, err = vm.LastAccepted()
	assert.NoError(err)
	// The block we just accepted
	block3, err := vm.getBlock(lastAcceptedID)
	assert.NoError(err)

	// Assert the block we accepted has the data we expect
	assert.Equal(snowmanBlock2.ID(), block3.Parent())
	assert.Equal([dataLen]byte{0, 0, 0, 0, 2}, block3.Data())
	assert.Equal(snowmanBlock3.ID(), block3.ID())
	assert.NoError(block3.Verify())

	// Next, check the blocks we added are there
	block2FromState, err := vm.getBlock(block2.ID())
	assert.NoError(err)
	assert.Equal(block2.ID(), block2FromState.ID())

	block3FromState, err := vm.getBlock(snowmanBlock3.ID())
	assert.NoError(err)
	assert.Equal(snowmanBlock3.ID(), block3FromState.ID())

	ctx.Lock.Unlock()
}

func TestService(t *testing.T) {
	// Initialize the vm
	dbManager := manager.NewMemDB(version.DefaultVersion1_0_0)
	msgChan := make(chan common.Message, 1)
	vm := &VM{}
	ctx := snow.DefaultContextTest()
	ctx.ChainID = blockchainID
	assert.NoError(t, vm.Initialize(ctx, dbManager, []byte{0, 0, 0, 0, 0}, nil, nil, msgChan, nil, nil))
	service := Service{vm}
	assert.NoError(t, service.GetBlock(nil, &GetBlockArgs{}, &GetBlockReply{}))
}
