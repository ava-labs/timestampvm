#!/usr/bin/env bash
# // (c) 2019-2022, Ava Labs, Inc. All rights reserved.
# // See the file LICENSE for licensing terms.


set -o errexit
set -o nounset
set -o pipefail

go test -race -timeout="3m" -coverprofile="coverage.out" -covermode="atomic" ./...
