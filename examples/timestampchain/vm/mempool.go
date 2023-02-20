// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
)

var (
	errEmptyMempool = errors.New("empty mempool")
	mempoolSize     = 100
)

type mempool struct {
	dataHashes chan ids.ID
}

func NewMempool() *mempool {
	return &mempool{
		dataHashes: make(chan ids.ID, mempoolSize),
	}
}

func (m *mempool) Add(dataHash ids.ID) error {
	select {
	case m.dataHashes <- dataHash:
		return nil
	default:
		return fmt.Errorf("failed to add DataHash(%s) to mempool due to full at size (%d)", dataHash, mempoolSize)
	}
}

func (m *mempool) Next() (ids.ID, error) {
	select {
	case nextDataHash := <-m.dataHashes:
		return nextDataHash, nil
	default:
		return ids.Empty, errEmptyMempool
	}
}
