// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stack

import (
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/validators"
)

type Network struct {
	common.AppHandler
	validators.Connector
}

// TODO:
// track connected valdiators
// Allow VM to implement a validators.Connector hook
// implement general protocol registry
// implement general mempool protocol for gossiping data hashes
// integrate into timestampvm chain
