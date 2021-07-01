#!/usr/bin/env bash

# Set the PATHS
GOPATH="$(go env GOPATH)"

# Set default binary location
binary_path="$GOPATH/src/github.com/ava-labs/avalanchego/build/avalanchego-latest/plugins/timestampvm"

# Avalabs docker hub
dockerhub_repo="avaplatform/avalanchego"

# Current branch
current_branch=${CURRENT_BRANCH:-$(git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD)}
echo "Using branch: ${current_branch}"

# Image build id
# Use an abbreviated version of the full commit to tag the image.

# WARNING: this will use the most recent commit even if there are un-committed changes present
timerpc_commit="$(git --git-dir="$TIMERPC_PATH/.git" rev-parse HEAD)"
timerpc_commit_id="${timerpc_commit::8}"

build_image_id=${BUILD_IMAGE_ID:-"$avalanche_version-$timerpc_commit_id"}
