#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Root directory
ROOT_DIR=$(
    cd "$(dirname "${BASH_SOURCE[0]}")"
    cd .. && pwd
)

# Load the constants
source "$ROOT_DIR"/scripts/constants.sh

go test -timeout=$UNIT_TEST_TIMEOUT -coverprofile="coverage.out" -covermode="atomic" $(go list ./... | grep -v tests)
