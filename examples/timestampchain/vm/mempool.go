// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/engine/common"
)

var (
	errEmptyMempool = errors.New("empty mempool")
	mempoolSize     = 100
)

// var _ Mempool[ids.ID, ids.ID] = (*mempool)(nil)

type mempool struct {
	toEngine   chan<- common.Message
	dataHashes chan ids.ID
}

// TODO
// type Mempool[Pending any, Item any] interface {
// 	Pending() (Pending, error)
// 	Add(Item) error
// }

func NewMempool(toEngine chan<- common.Message) *mempool {
	return &mempool{
		dataHashes: make(chan ids.ID, mempoolSize),
		toEngine:   toEngine,
	}
}

func (m *mempool) Add(dataHash ids.ID) error {
	m.notifyEngine()

	select {
	case m.dataHashes <- dataHash:
		return nil
	default:
		return fmt.Errorf("failed to add DataHash(%s) to mempool due to full at size (%d)", dataHash, mempoolSize)
	}
}

func (m *mempool) Pending() (ids.ID, error) {
	select {
	case nextDataHash := <-m.dataHashes:
		return nextDataHash, nil
	default:
		return ids.Empty, errEmptyMempool
	}
}

func (m *mempool) NotifyBuildBlock() {
	if m.Len() > 0 {
		m.notifyEngine()
	}
}

func (m *mempool) notifyEngine() {
	select {
	case m.toEngine <- common.PendingTxs:
	default:
	}
}

func (m *mempool) Len() int {
	return len(m.dataHashes)
}
