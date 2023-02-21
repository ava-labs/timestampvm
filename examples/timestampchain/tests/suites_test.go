// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/api/health"
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

var _ = ginkgo.Describe("[Ping]", ginkgo.Ordered, func() {
	ginkgo.It("ping the network", ginkgo.Label("setup"), func() {
		client := health.NewClient(utils.DefaultLocalNodeURI)
		healthy, err := client.Readiness(context.Background())
		gomega.Expect(err).Should(gomega.BeNil())
		gomega.Expect(healthy.Healthy).Should(gomega.BeTrue())
	})
})