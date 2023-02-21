#!/usr/bin/env bash
set -e

# Start a single node non-staking network to perform tests
if ! [[ "$0" =~ scripts/run.sh ]]; then
  echo "must be run from repository root, but got $0"
  exit 255
fi

# Load the versions
ROOT_DIR=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)
# Load the constants
source "$ROOT_DIR"/scripts/constants.sh

# Set up avalanche binary path and assume build directory is set
AVALANCHEGO_BUILD_PATH=${AVALANCHEGO_BUILD_PATH:-"$GOPATH/src/github.com/ava-labs/avalanchego/build"}
AVALANCHEGO_PATH=${AVALANCHEGO_PATH:-"$AVALANCHEGO_BUILD_PATH/avalanchego"}
AVALANCHEGO_PLUGIN_DIR=${AVALANCHEGO_PLUGIN_DIR:-"$AVALANCHEGO_BUILD_PATH/plugins"}
DATA_DIR=${DATA_DIR:-/tmp/subnet-evm-start-node/$(date "+%Y-%m-%d%:%H:%M:%S")}

mkdir -p $DATA_DIR

# Set the config file contents for the path passed in as the first argument
function _set_config(){
  cat <<EOF >$1
  {
    "network-id": "local",
    "staking-enabled": false,
    "health-check-frequency": "5s",
    "plugin-dir": "$AVALANCHEGO_PLUGIN_DIR"
  }
EOF
}

function execute_cmd() {
  echo "Executing command: $@"
  $@
}

NODE_NAME="node1"
NODE_DATA_DIR="$DATA_DIR/$NODE_NAME"
echo "Creating data directory: $NODE_DATA_DIR"
mkdir -p $NODE_DATA_DIR
NODE_CONFIG_FILE_PATH="$NODE_DATA_DIR/config.json"
_set_config $NODE_CONFIG_FILE_PATH

CMD="$AVALANCHEGO_PATH --data-dir=$NODE_DATA_DIR --config-file=$NODE_CONFIG_FILE_PATH"

execute_cmd $CMD
