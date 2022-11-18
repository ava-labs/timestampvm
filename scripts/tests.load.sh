#!/usr/bin/env bash
# // (c) 2019-2022, Ava Labs, Inc. All rights reserved.
# // See the file LICENSE for licensing terms.

set -e

# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"

# Ensure we are in the right location
if ! [[ "$0" =~ scripts/tests.load.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

VERSION=$1
if [[ -z "${VERSION}" ]]; then
  echo "Missing version argument!"
  echo "Usage: ${0} [VERSION]" >> /dev/stderr
  exit 255
fi

############################
echo "downloading avalanchego"
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
AVALANCHEGO_PATH=/tmp/avalanchego-v${VERSION}/avalanchego
AVALANCHEGO_PLUGIN_DIR=/tmp/avalanchego-v${VERSION}/plugins

if [ ! -f "$AVALANCHEGO_PATH" ]; then
  DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/v${VERSION}/avalanchego-linux-${GOARCH}-v${VERSION}.tar.gz
  DOWNLOAD_PATH=/tmp/avalanchego.tar.gz
  if [[ ${GOOS} == "darwin" ]]; then
    DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/v${VERSION}/avalanchego-macos-v${VERSION}.zip
    DOWNLOAD_PATH=/tmp/avalanchego.zip
  fi

  rm -rf /tmp/avalanchego-v${VERSION}
  rm -rf /tmp/avalanchego-build
  rm -f ${DOWNLOAD_PATH}

  echo "downloading avalanchego ${VERSION} at ${DOWNLOAD_URL}"
  curl -L ${DOWNLOAD_URL} -o ${DOWNLOAD_PATH}

  echo "extracting downloaded avalanchego"
  if [[ ${GOOS} == "linux" ]]; then
    tar xzvf ${DOWNLOAD_PATH} -C /tmp
  elif [[ ${GOOS} == "darwin" ]]; then
    unzip ${DOWNLOAD_PATH} -d /tmp/avalanchego-build
    mv /tmp/avalanchego-build/build /tmp/avalanchego-v${VERSION}
  fi
  find /tmp/avalanchego-v${VERSION}
fi

############################
echo "building timestampvm"

# delete previous (if exists)
rm -f /tmp/avalanchego-v${VERSION}/plugins/tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH

go build \
-o /tmp/avalanchego-v${VERSION}/plugins/tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH \
./main/
find /tmp/avalanchego-v${VERSION}


############################
echo "creating genesis file"
echo -n "e2e" >> /tmp/.genesis

############################
echo "creating vm config"
echo -n "{}" >> /tmp/.config

############################
echo "building load.test"
# to install the ginkgo binary (required for test build and run)
go install -v github.com/onsi/ginkgo/v2/ginkgo@v2.1.4
ACK_GINKGO_RC=true ginkgo build ./tests/load

#################################
# download avalanche-network-runner
# https://github.com/ava-labs/avalanche-network-runner
ANR_REPO_PATH=github.com/ava-labs/avalanche-network-runner
ANR_VERSION=v1.3.2
# version set
go install -v ${ANR_REPO_PATH}@${ANR_VERSION}

#################################
# run "avalanche-network-runner" server
GOPATH=$(go env GOPATH)
if [[ -z ${GOBIN+x} ]]; then
  # no gobin set
  BIN=${GOPATH}/bin/avalanche-network-runner
else
  # gobin set
  BIN=${GOBIN}/avalanche-network-runner
fi

echo "launch avalanche-network-runner in the background"
$BIN server \
--log-level warn \
--port=":12342" \
--disable-grpc-gateway &
PID=${!}

############################
# By default, it runs all e2e test cases!
# Use "--ginkgo.skip" to skip tests.
# Use "--ginkgo.focus" to select tests.
echo "running load tests"
./tests/load/load.test \
--ginkgo.v \
--network-runner-log-level warn \
--network-runner-grpc-endpoint="0.0.0.0:12342" \
--avalanchego-path=${AVALANCHEGO_PATH} \
--avalanchego-plugin-dir=${AVALANCHEGO_PLUGIN_DIR} \
--vm-genesis-path=/tmp/.genesis \
--vm-config-path=/tmp/.config \
--terminal-height=1000000

############################
# load.test" already terminates the cluster
# just in case load tests are aborted, manually terminate them again
echo "network-runner RPC server was running on PID ${PID} as test mode; terminating the process..."
pkill -P ${PID} || true
kill -2 ${PID} || true
pkill -9 -f tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH || true # in case pkill didn't work
