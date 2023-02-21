// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"context"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/version"
)

var _ validators.Connector = (*peerTracker)(nil)

type peer struct {
	version *version.Application
}

type peerTracker struct {
	lock  sync.RWMutex
	peers map[ids.NodeID]*peer
}

func newPeerTracker() *peerTracker {
	return &peerTracker{
		peers: make(map[ids.NodeID]*peer),
	}
}

func (p *peerTracker) Connected(_ context.Context, nodeID ids.NodeID, nodeVersion *version.Application) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.peers[nodeID] = &peer{
		version: nodeVersion,
	}
	return nil
}

func (p *peerTracker) Disconnected(_ context.Context, nodeID ids.NodeID) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.peers, nodeID)
	return nil
}
