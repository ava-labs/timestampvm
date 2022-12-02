// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ava-labs/timestampvm/tests/network"
	"github.com/ethereum/go-ethereum/log"
)

var (
	networkRunnerConfigString string
	config                    network.NetworkRunnerConfig
)

func init() {
	flag.StringVar(
		&networkRunnerConfigString,
		"network-runner-config",
		`
		"network-runner-log-level": ,
		"network-runner-endpoint": ,
		"avalanchego-exec-path": ,
		"plugin-dir": ,
		"vm-id": ,
		"blockchain-specs": ,
		"global-node-config": ,
`,
		"Full network runner config",
	)
}

func run(ctx context.Context, quit <-chan struct{}) error {
	if len(networkRunnerConfigString) == 0 {
		return errors.New("cannot start network with empty config")
	}

	err := json.Unmarshal([]byte(networkRunnerConfigString), &config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal network runner config: %w", err)
	}

	runner, err := network.NewStaticNetworkRunnerNetwork(config)
	if err != nil {
		return fmt.Errorf("failed to create network client: %w", err)
	}

	if err := runner.CreateDefault(ctx); err != nil {
		return fmt.Errorf("failed to construct default network: %w", err)
	}

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-quit: // Leave the error nil
	}

	teardownErr := runner.Teardown(ctx)
	if err == nil {
		err = teardownErr
	}

	return err

}

func main() {
	ctx := context.Background()
	quit := make(chan struct{})

	// register signals to kill the application
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	signal.Notify(signals, syscall.SIGTERM)

	// Close the quit channel after receiving a signal for a graceful shutdown.
	go func() {
		<-signals
		close(quit)
	}()

	if err := run(ctx, quit); err != nil {
		log.Error("network runner failed", "err", err)
		return
	}

	log.Info("Terminated successfully.")
}
