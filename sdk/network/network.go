// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"context"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/version"
)

var (
	_ common.NetworkAppHandler = (*Network)(nil)
	_ validators.Connector     = (*Network)(nil)
)

type Network struct {
	common.AppHandler

	peerTracker *peerTracker

	connectorsLock sync.RWMutex
	connectors     []validators.Connector
}

func NewNetwork(connectors []validators.Connector) *Network {
	peerTracker := newPeerTracker()
	connectors = append(connectors, peerTracker)
	network := &Network{
		AppHandler:  common.NewNoOpAppHandler(logging.NoLog{}), // TODO: replace no-op handler with protocol registry
		peerTracker: peerTracker,
		connectors:  connectors,
	}
	return network
}

func (n *Network) Connected(ctx context.Context, nodeID ids.NodeID, nodeVersion *version.Application) error {
	n.connectorsLock.RLock()
	defer n.connectorsLock.RUnlock()

	for _, connector := range n.connectors {
		if err := connector.Connected(ctx, nodeID, nodeVersion); err != nil {
			return err
		}
	}
	return nil
}

func (n *Network) Disconnected(ctx context.Context, nodeID ids.NodeID) error {
	n.connectorsLock.RLock()
	defer n.connectorsLock.RUnlock()

	for _, connector := range n.connectors {
		if err := connector.Disconnected(ctx, nodeID); err != nil {
			return err
		}
	}
	return nil
}
