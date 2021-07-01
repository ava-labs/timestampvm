// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"fmt"
	"net/http"

	"github.com/ava-labs/avalanchego/utils/formatting"
)

// StaticService defines the base service for the timestamp vm
type StaticService struct{}

// CreateStaticService ...
func CreateStaticService() *StaticService {
	return &StaticService{}
}

// BuildGenesisArgs are arguments for BuildGenesis
type BuildGenesisArgs struct {
	GenesisData string              `json:"genesisData"`
	Encoding    formatting.Encoding `json:"encoding"`
}

// BuildGenesisReply is the reply from BuildGenesis
type BuildGenesisReply struct {
	Bytes    string              `json:"bytes"`
	Encoding formatting.Encoding `json:"encoding"`
}

// BuildGenesis returns the encoded genesisData
func (ss *StaticService) BuildGenesis(_ *http.Request, args *BuildGenesisArgs, reply *BuildGenesisReply) error {
	bytes, err := formatting.Encode(args.Encoding, []byte(args.GenesisData))
	if err != nil {
		return fmt.Errorf("couldn't encode genesis as string: %s", err)
	}
	reply.Bytes = bytes
	reply.Encoding = args.Encoding
	return nil
}
