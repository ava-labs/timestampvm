#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Root directory
ROOT_DIRECTORY_PATH=$(
    cd "$(dirname "${BASH_SOURCE[0]}")"
    cd .. && pwd
)

# Load the constants
source "$ROOT_DIRECTORY_PATH"/scripts/constants.sh

if [[ $# -ne 3 ]]; then
    echo "Invalid arguments to scripts/build_vm.sh, expected three arguments: [NAME] [PLUGIN_DIR] [BINARY_NAME]"
fi

NAME=$1
PLUGIN_DIR=$2
BINARY_NAME=$3

# Build the specififed VM
echo "Building $NAME @ $PLUGIN_DIR/$BINARY_NAME"
go build -o "$PLUGIN_DIR/$BINARY_NAME" "examples/$NAME/main/"*.go
