// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"context"
	"os"
	"time"

	"github.com/ava-labs/avalanchego/api/health"
	"github.com/go-cmd/cmd"
	"github.com/onsi/gomega"
)

// GinkgoSetup returns a BeforeSuite and AfterSuite function to set up and teardown a single
// node local network with staking disabled.
func GinkgoSetup() (beforeSuite func(), afterSuite func()) {
	var startCmd *cmd.Cmd

	beforeSuite = func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		wd, err := os.Getwd()
		gomega.Expect(err).Should(gomega.BeNil())
		_ = wd
		// log.Info("Starting AvalancheGo node", "wd", wd)
		startCmd, err = RunCommand("./scripts/run.sh")
		gomega.Expect(err).Should(gomega.BeNil())

		// Assumes that startCmd will launch a node with HTTP Port at [utils.DefaultLocalNodeURI]
		healthClient := health.NewClient(DefaultLocalNodeURI)
		healthy, err := health.AwaitReady(ctx, healthClient, 5*time.Second)
		gomega.Expect(err).Should(gomega.BeNil())
		gomega.Expect(healthy).Should(gomega.BeTrue())
		// log.Info("AvalancheGo node is healthy")
	}

	afterSuite = func() {
		gomega.Expect(startCmd).ShouldNot(gomega.BeNil())
		gomega.Expect(startCmd.Stop()).Should(gomega.BeNil())
	}

	return beforeSuite, afterSuite
}
