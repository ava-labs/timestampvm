// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// load implements the load tests.
package load_test

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	runner_sdk "github.com/ava-labs/avalanche-network-runner/client"
	"github.com/ava-labs/avalanche-network-runner/rpcpb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/timestampvm/tests/network"
	log "github.com/inconshreveable/log15"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/gomega"
)

func TestLoad(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "timestampvm load test suites")
}

var (
	vmGenesisPath    string
	vmConfigPath     string
	subnetConfigPath string

	// Comma separated list of client URIs
	// If the length is non-zero, this will skip using the network runner to start and stop a network.
	commaSeparatedClientURIs    string
	commaSeparatedBlockchainIDs string
	// Specifies the full timestampvm client URIs to use for load test.
	// Populated in BeforeSuite
	clientURIs    []string
	blockchainIDs []string

	terminalHeight uint64
	numBlockchains int

	// testNetwork provides the static network to execute the load test on.
	// Set in BeforeSuite and taken down in AfterSuite.
	testNetwork network.StaticNetwork
	config      network.NetworkRunnerConfig
)

func init() {
	// Network runner flags
	flag.StringVar(
		&config.NetworkRunnerLogLevel,
		"network-runner-log-level",
		"info",
		"gRPC server endpoint",
	)

	flag.StringVar(
		&config.NetworkRunnerEndpoint,
		"network-runner-grpc-endpoint",
		"0.0.0.0:8080",
		"gRPC server endpoint",
	)

	flag.StringVar(
		&config.AvalancheGoExecPath,
		"avalanchego-path",
		"",
		"avalanchego executable path",
	)

	flag.StringVar(
		&config.PluginDir,
		"avalanchego-plugin-dir",
		"",
		"avalanchego plugin directory",
	)

	// Blockchain specification arguments
	flag.StringVar(
		&vmGenesisPath,
		"vm-genesis-path",
		"",
		"VM genesis file path",
	)

	flag.StringVar(
		&vmConfigPath,
		"vm-config-path",
		"",
		"VM configfile path",
	)

	flag.StringVar(
		&subnetConfigPath,
		"subnet-config-path",
		"",
		"Subnet configfile path",
	)

	// Test parameters
	flag.Uint64Var(
		&terminalHeight,
		"terminal-height",
		1_000_000,
		"height to quit at",
	)
	flag.IntVar(
		&numBlockchains,
		"num-blockchains",
		1,
		"Sets the number of blockchains to create and throughput test.",
	)

	// Override flag to set the client URIs manually instead of constructing a network.
	flag.StringVar(
		&commaSeparatedClientURIs,
		"client-uris",
		"",
		"Specifies a comma separated list of full timestampvm client URIs to use in place of orchestrating a network. (Ex. 127.0.0.1:9650/ext/bc/q2aTwKuyzgs8pynF7UXBZCU7DejbZbZ6EUyHr3JQzYgwNPUPi/rpc,127.0.0.1:9652/ext/bc/q2aTwKuyzgs8pynF7UXBZCU7DejbZbZ6EUyHr3JQzYgwNPUPi/rpc",
	)
	flag.StringVar(
		&commaSeparatedBlockchainIDs,
		"blockchain-ids",
		"",
		"Specifies a set of blockchainIDs on which to perform a throughput test (assumes all clients are validating every blockchainID). Must be populaeted if client-uris is populated.",
	)
}

const vmName = "timestamp"

var vmID ids.ID

func init() {
	// TODO: add "getVMID" util function in avalanchego and import from "avalanchego"
	b := make([]byte, 32)
	copy(b, []byte(vmName))
	var err error
	vmID, err = ids.ToID(b)
	if err != nil {
		panic(err)
	}
}

var (
	cli               runner_sdk.Client
	timestampvmRPCEps []string
)

var _ = ginkgo.BeforeSuite(func() {
	if len(commaSeparatedClientURIs) != 0 {
		clientURIs = strings.Split(commaSeparatedClientURIs, ",")
		blockchainIDs = strings.Split(commaSeparatedBlockchainIDs, ",")

		outf("{{green}}creating %d clients from manually specified URIs:{{/}}\n", len(clientURIs))
		testNetwork = network.NewExistingNetwork(clientURIs)
		return
	}

	// Create [numBlockchains] as specified
	for i := 0; i < numBlockchains; i++ {
		config.BlockchainSpecs = append(config.BlockchainSpecs, &rpcpb.BlockchainSpec{
			VmName:       vmName,
			Genesis:      vmGenesisPath,
			ChainConfig:  vmConfigPath,
			SubnetConfig: subnetConfigPath,
		})
	}

	runner, err := network.NewStaticNetworkRunnerNetwork(config)
	gomega.Expect(err).Should(gomega.BeNil())

	err = runner.CreateDefault(context.Background())
	gomega.Expect(err).Should(gomega.BeNil())

	blockchainIDs, err = runner.BlockchainIDs(context.Background())
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(blockchainIDs).Should(gomega.Not(gomega.BeEmpty()))
})

var _ = ginkgo.AfterSuite(func() {
	log.Info("Tearing down network")
	err := testNetwork.Teardown(context.Background())
	gomega.Expect(err).Should(gomega.BeNil())
	log.Info("Finished tearing down network.")
})

// Tests only assumes that [instances] has been populated by BeforeSuite
var _ = ginkgo.Describe("[ProposeBlock]", func() {
	ginkgo.It("load test", func() {
		workers := newLoadWorkers(clientURIs, blockchainIDs[0])

		err := RunLoadTest(context.Background(), workers, terminalHeight, 2*time.Minute)
		gomega.Ω(err).Should(gomega.BeNil())
		log.Info("Load test completed successfully")
	})
})

// Outputs to stdout.
//
// e.g.,
//
//	Out("{{green}}{{bold}}hi there %q{{/}}", "aa")
//	Out("{{magenta}}{{bold}}hi therea{{/}} {{cyan}}{{underline}}b{{/}}")
//
// ref.
// https://github.com/onsi/ginkgo/blob/v2.0.0/formatter/formatter.go#L52-L73
func outf(format string, args ...interface{}) {
	s := formatter.F(format, args...)
	fmt.Fprint(formatter.ColorableStdOut, s)
}
