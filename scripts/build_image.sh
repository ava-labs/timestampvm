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

echo "Building Docker Image: $dockerhub_repo:$build_image_id based of $avalanche_version"
docker build -t "$dockerhub_repo:$build_image_id" "$TIMERPC_PATH" -f "$TIMERPC_PATH/Dockerfile" \
  --build-arg AVALANCHE_VERSION="$avalanche_version" \
  --build-arg TIMERPC_COMMIT="$timerpc_commit" \
  --build-arg CURRENT_BRANCH="$current_branch"
