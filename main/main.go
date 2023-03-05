// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/vms/rpcchainvm"

	"github.com/ava-labs/timestampvm/timestampvm"
)

func main() {
	version, err := PrintVersion()
	if err != nil {
		fmt.Printf("couldn't get config: %s\n", err)
		os.Exit(1)
	}
	// Print VM ID and exit
	if version {
		fmt.Printf("%s@%s\n", timestampvm.Name, timestampvm.Version)
		os.Exit(0)
	}

	err = rpcchainvm.Serve(context.Background(), &timestampvm.VM{})
	if err != nil {
		fmt.Printf("serve returned an error: %s\n", err)
	}
}
