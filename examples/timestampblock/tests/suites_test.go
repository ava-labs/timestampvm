// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/timestampvm/examples/timestampchain/vm"
	"github.com/ava-labs/timestampvm/sdk/utils"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var beforeSuite, afterSuite = utils.GinkgoSetup()

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Timestampchain Ginkgo Suite")
}

var (
	_ = ginkgo.BeforeSuite(beforeSuite)
	_ = ginkgo.AfterSuite(afterSuite)
)

var _ = ginkgo.Describe("[Workflow]", ginkgo.Ordered, ginkgo.Label("Timestamp"), ginkgo.Label("Chain"), func() {
	ginkgo.It("ping the network", ginkgo.Label("setup"), func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		genesisDataHash := ids.ID{1, 2, 3, 4, 5}
		blockchainID := utils.CreateNewBlockchain(ctx, vm.ID, genesisDataHash[:])

		client := vm.NewClient(fmt.Sprintf("%s/ext/bc/%s/timestamp", utils.DefaultLocalNodeURI, blockchainID))
		genesisBlock, err := client.GetBlock(ctx, ids.Empty)
		gomega.Expect(err).Should(gomega.BeNil())
		gomega.Expect(genesisBlock.DataHash).Should(gomega.Equal(genesisDataHash))

		nextDataHash := ids.ID{5, 4, 3, 2, 1}
		gomega.Expect(client.ProposeBlock(ctx, nextDataHash)).Should(gomega.BeNil())
		for {
			block, err := client.GetBlock(ctx, ids.Empty)
			// If we encounter an unexpected API error, fail early
			if err != nil {
				gomega.Expect(err).Should(gomega.BeNil())
			}

			if block.Height() == 0 {
				continue
			}
			// Assert that the block is produced and accepted and then allow the test to terminate
			gomega.Expect(block.Height()).Should(gomega.Equal(uint64(1)))
			gomega.Expect(block.DataHash).Should(gomega.Equal(nextDataHash))
			break
		}
	})
})
