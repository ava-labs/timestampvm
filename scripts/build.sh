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

GOPATH="$(go env GOPATH)"

# Set default binary directory location
NAME="timestampchain"
PLUGIN_DIR="$GOPATH/src/github.com/ava-labs/avalanchego/build/plugins"
BINARY_NAME="tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH"

./scripts/build_vm.sh $NAME $PLUGIN_DIR $BINARY_NAME