// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/database/versiondb"
)

var (
	// These are prefixes for db keys.
	// It's important to set different prefixes for each separate database objects.
	singletonStatePrefix = []byte("singleton")
	blockStatePrefix     = []byte("block")

	_ State = &state{}
)

// State is a wrapper around avax.SingleTonState and BlockState
// State also exposes a few methods needed for managing database commits and close.
type State interface {
	InitializedState
	BlockState

	Commit() error
	Close() error
}

type state struct {
	InitializedState
	BlockState

	baseDB *versiondb.Database
}

func NewState(db database.Database, vm *VM) State {
	// create a new baseDB
	baseDB := versiondb.New(db)

	// create a prefixed "singletonDB" from baseDB
	singletonDB := prefixdb.New(singletonStatePrefix, baseDB)
	// create a prefixed "blockDB" from baseDB
	blockDB := prefixdb.New(blockStatePrefix, baseDB)

	// return state with created sub state components
	return &state{
		InitializedState: NewInitializedState(singletonDB),
		BlockState:       NewBlockState(blockDB, vm),
		baseDB:           baseDB,
	}
}

// Commit commits pending operations to baseDB
func (s *state) Commit() error {
	return s.baseDB.Commit()
}

// Close closes the underlying base database
func (s *state) Close() error {
	return s.baseDB.Close()
}
