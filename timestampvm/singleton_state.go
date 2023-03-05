// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"github.com/ava-labs/avalanchego/database"
)

const (
	IsInitializedKey byte = iota
)

var (
	isInitializedKey                  = []byte{IsInitializedKey}
	_                InitializedState = (*initializedState)(nil)
)

// InitializedState is a thin wrapper around a database to provide, caching,
// serialization, and de-serialization of the initialization status.
type InitializedState interface {
	IsInitialized() (bool, error)
	SetInitialized() error
}

type initializedState struct {
	singletonDB database.Database
}

func NewInitializedState(db database.Database) InitializedState {
	return &initializedState{
		singletonDB: db,
	}
}

func (s *initializedState) IsInitialized() (bool, error) {
	return s.singletonDB.Has(isInitializedKey)
}

func (s *initializedState) SetInitialized() error {
	return s.singletonDB.Put(isInitializedKey, nil)
}
