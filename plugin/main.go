// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"github.com/hashicorp/go-plugin"

	"github.com/ava-labs/avalanchego/vms/rpcchainvm"
	"github.com/ava-labs/timerpc/plugin/timestampvm"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: rpcchainvm.Handshake,
		Plugins: map[string]plugin.Plugin{
			"vm": rpcchainvm.New(&timestampvm.VM{}),
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
