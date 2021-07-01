#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Avalanche root directory
TIMERPC_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )

# Load the versions
source "$TIMERPC_PATH"/scripts/versions.sh

# Load the constants
source "$TIMERPC_PATH"/scripts/constants.sh

if [[ $# -eq 1 ]]; then
    binary_path=$1
elif [[ $# -ne 0 ]]; then
    echo "Invalid arguments to build timerpc. Requires either no arguments (default) or one arguments to specify binary location."
    exit 1
fi

# Check if TIMERPC_COMMIT is set, if not retrieve the last commit from the repo.
# This is used in the Dockerfile to allow a commit hash to be passed in without
# including the .git/ directory within the Docker image.
timerpc_commit=${TIMERPC_COMMIT:-$( git rev-list -1 HEAD )}

# Build Timerpc, which is run as a subprocess
echo "Building Timerpc Version: $timerpc_version; GitCommit: $timerpc_commit"
go build -ldflags "-X github.com/ava-labs/timerpc/plugin/timerpc.GitCommit=$timerpc_commit -X github.com/ava-labs/timerpc/plugin/timerpc.Version=$timerpc_version" -o "$binary_path" "plugin/"*.go
