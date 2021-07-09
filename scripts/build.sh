#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Load the constants
# Set the PATHS
GOPATH="$(go env GOPATH)"

# Set default binary location
binary_path="$GOPATH/src/github.com/ava-labs/avalanchego/build/avalanchego-latest/plugins/timestampvm"

if [[ $# -eq 1 ]]; then
    binary_path=$1
elif [[ $# -ne 0 ]]; then
    echo "Invalid arguments to build timestampvm. Requires either no arguments (default) or one arguments to specify binary location."
    exit 1
fi

# Check if timestampvm_COMMIT is set, if not retrieve the last commit from the repo.
# This is used in the Dockerfile to allow a commit hash to be passed in without
# including the .git/ directory within the Docker image.
timestampvm_commit=${timestampvm_COMMIT:-$( git rev-list -1 HEAD )}

# Build timestampvm, which is run as a subprocess
echo "Building timestampvm; GitCommit: $timestampvm_commit"
go build -ldflags "-X github.com/ava-labs/timestampvm/main/timestampvm.GitCommit=$timestampvm_commit" -o "$binary_path" "main/"*.go
