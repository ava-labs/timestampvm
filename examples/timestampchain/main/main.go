// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"

	log "github.com/inconshreveable/log15"

	"github.com/ava-labs/avalanchego/vms/rpcchainvm"
	timestampvm "github.com/ava-labs/timestampvm/examples/timestampchain/vm"
	"github.com/ava-labs/timestampvm/sdk/stack"
)

func main() {
	version, err := PrintVersion()
	if err != nil {
		fmt.Printf("couldn't get config: %s", err)
		os.Exit(1)
	}
	// Print VM ID and exit
	if version {
		fmt.Printf("%s@%s\n", timestampvm.Name, timestampvm.Version)
		os.Exit(0)
	}

	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat())))

	rpcchainvm.Serve(&stack.VM[*timestampvm.Block]{ChainVM: &timestampvm.VM{}})
}
