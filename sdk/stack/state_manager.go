// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stack

import "github.com/ava-labs/avalanchego/ids"

type StateManager[State Rooter] interface {
	// OpenState opens a read/write eligible copy of the state at a specific [root]
	OpenState(root ids.ID) (State, error)
	// Commit commits the state to disk (the guarantees provided by this action depend on the implementation)
	Commit(State) error
}
