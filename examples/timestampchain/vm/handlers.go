// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
)

// Service is the API service for this VM
type Service struct{ vm *VM }

// ProposeBlockArgs are the arguments to function ProposeValue
type ProposeBlockArgs struct {
	// DataHash to include in the block
	DataHash ids.ID `json:"data"`
}

// ProposeBlock is an API method to propose a new block whose data is [args].Data.
// [args].Data must be a string repr. of a 32 byte array
func (s *Service) ProposeBlock(_ *http.Request, args *ProposeBlockArgs, reply *api.EmptyReply) error {
	return s.vm.mempool.Add(args.DataHash)
}

// // GetBlockArgs are the arguments to GetBlock
// type GetBlockArgs struct {
// 	// ID of the block we're getting.
// 	// If left blank, gets the latest block
// 	ID *ids.ID `json:"id"`
// }

// // GetBlockReply is the reply from GetBlock
// type GetBlockReply struct {
// 	Timestamp json.Uint64 `json:"timestamp"` // Timestamp of block
// 	Data      string      `json:"data"`      // Data (hex-encoded) in block
// 	Height    json.Uint64 `json:"height"`    // Height of block
// 	ID        ids.ID      `json:"id"`        // String repr. of ID of block
// 	ParentID  ids.ID      `json:"parentID"`  // String repr. of ID of block's parent
// }

// // GetBlock gets the block whose ID is [args.ID]
// // If [args.ID] is empty, get the latest block
// func (s *Service) GetBlock(_ *http.Request, args *GetBlockArgs, reply *GetBlockReply) error {
// 	// If an ID is given, parse its string representation to an ids.ID
// 	// If no ID is given, ID becomes the ID of last accepted block
// 	var (
// 		id  ids.ID
// 		err error
// 	)

// 	if args.ID == nil {
// 		id, err = s.vm.state.GetLastAccepted()
// 		if err != nil {
// 			return errCannotGetLastAccepted
// 		}
// 	} else {
// 		id = *args.ID
// 	}

// 	// Get the block from the database
// 	block, err := s.vm.getBlock(id)
// 	if err != nil {
// 		return errNoSuchBlock
// 	}

// 	// Fill out the response with the block's data
// 	reply.Timestamp = json.Uint64(block.Timestamp().Unix())
// 	data := block.Data()
// 	reply.Data, err = formatting.Encode(formatting.Hex, data[:])
// 	reply.Height = json.Uint64(block.Hght)
// 	reply.ID = block.ID()
// 	reply.ParentID = block.Parent()

// 	return err
// }
