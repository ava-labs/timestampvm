// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// network implements an interface for setting up a default network for testing purposes.
package network

import (
	"context"
	"time"

	runner_sdk "github.com/ava-labs/avalanche-network-runner/client"
	"github.com/ava-labs/avalanche-network-runner/rpcpb"
	"github.com/ava-labs/avalanchego/utils/logging"
	log "github.com/inconshreveable/log15"
)

var (
	_ StaticNetwork = (*existingNetwork)(nil)
	_ StaticNetwork = (*networkRunner)(nil)
)

// Network supports a basic interface for setting up, interacting with, and destructing a network.
// This interface is intended to be used by tests that do not need to change the underlying state
// of the net
type StaticNetwork interface {
	CreateDefault(context.Context) error
	URIs(context.Context) ([]string, error)
	Teardown(context.Context) error
}

// existingNetwork implements the StaticNetwork interface and assumes that the network
// has already been constructed and does not require any startup/teardown.
type existingNetwork struct {
	uris []string
}

func NewExistingNetwork(uris []string) *existingNetwork {
	return &existingNetwork{
		uris: uris,
	}
}

func (e *existingNetwork) CreateDefault(context.Context) error    { return nil }
func (e *existingNetwork) URIs(context.Context) ([]string, error) { return e.uris, nil }
func (e *existingNetwork) Teardown(context.Context) error         { return nil }

type NetworkRunnerConfig struct { //nolint
	NetworkRunnerLogLevel string                  `json:"network-runner-log-level"`
	NetworkRunnerEndpoint string                  `json:"network-runner-endpoint"`
	AvalancheGoExecPath   string                  `json:"avalanchego-exec-path"`
	PluginDir             string                  `json:"plugin-dir"`
	VMID                  string                  `json:"vm-id"`
	BlockchainSpecs       []*rpcpb.BlockchainSpec `json:"blockchain-specs"`
	GlobalNodeConfig      string                  `json:"global-node-config"`
}

type networkRunner struct {
	client runner_sdk.Client
	config NetworkRunnerConfig
}

func NewStaticNetworkRunnerNetwork(config NetworkRunnerConfig) (*networkRunner, error) {
	logLevel, err := logging.ToLevel(config.NetworkRunnerLogLevel)
	if err != nil {
		return nil, err
	}
	logFactory := logging.NewFactory(logging.Config{
		DisplayLevel: logLevel,
		LogLevel:     logLevel,
	})
	log, err := logFactory.Make("main")
	if err != nil {
		return nil, err
	}

	client, err := runner_sdk.New(runner_sdk.Config{
		Endpoint:    config.NetworkRunnerEndpoint,
		DialTimeout: 10 * time.Second,
	}, log)
	if err != nil {
		return nil, err
	}

	return &networkRunner{
		client: client,
		config: config,
	}, nil
}

func (n *networkRunner) CreateDefault(ctx context.Context) error {
	log.Info("Starting network runner", "execPath", n.config.AvalancheGoExecPath, "vmID", n.config.VMID)

	resp, err := n.client.Start(
		ctx,
		n.config.AvalancheGoExecPath,
		runner_sdk.WithPluginDir(n.config.PluginDir),
		runner_sdk.WithBlockchainSpecs(
			n.config.BlockchainSpecs,
		),
		// Disable all rate limiting
		runner_sdk.WithGlobalNodeConfig(n.config.GlobalNodeConfig),
	)
	if err != nil {
		return err
	}
	log.Info("Successfully started", "node names", resp.ClusterInfo.NodeNames)

	// TODO: network runner health should imply custom VM healthiness
	// or provide a separate API for custom VM healthiness
	// "start" is async, so wait some time for cluster health
	log.Info("Waiting for network to report healthy...")
	for {
		healthRes, err := n.client.Health(ctx)
		if err != nil || !healthRes.ClusterInfo.Healthy {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	log.Info("Network reporting healthy.")
	return nil
}

func (n *networkRunner) URIs(ctx context.Context) ([]string, error) {
	return n.client.URIs(ctx)
}

func (n *networkRunner) Teardown(ctx context.Context) error {
	log.Info("Shutting down network.")
	_, err := n.client.Stop(ctx)
	if err != nil {
		return err
	}
	log.Info("Successfully stopped the network.")

	err = n.client.Close()
	if err != nil {
		return err
	}
	log.Info("Closed client")

	return nil
}
