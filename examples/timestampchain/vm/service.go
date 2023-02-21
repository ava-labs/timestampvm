// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"errors"
	"net/http"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
)

var errCannotGetLastAccepted = errors.New("cannot get last accepted block")

// Service is the API service for this VM
type Service struct{ vm *VM }

// BlockIDArgs is an API request where the only argument is a single block ID
type BlockIDArgs struct {
	// DataHash to include in the block
	ID ids.ID `json:"data"`
}

// ProposeBlock is an API method to propose a new block whose data is [args].Data.
// [args].Data must be a string repr. of a 32 byte array
func (s *Service) ProposeBlock(_ *http.Request, args *BlockIDArgs, reply *api.EmptyReply) error {
	return s.vm.mempool.Add(args.ID)
}

// GetBlock gets the block whose ID is [args.ID]
// If [args.ID] is empty, get the latest block
func (s *Service) GetBlock(_ *http.Request, args *BlockIDArgs, block *Block) error {
	var (
		requestedBlockID = args.ID
		err              error
	)
	if requestedBlockID == ids.Empty {
		requestedBlockID, err = s.vm.LastAccepted(context.Background())
		if err != nil {
			return errCannotGetLastAccepted
		}
	}
	retrievedBlock, err := s.vm.GetBlock(context.TODO(), requestedBlockID)
	if err != nil {
		return err
	}

	*block = *retrievedBlock
	return nil
}
