#!/usr/bin/env bash
# // (c) 2019-2022, Ava Labs, Inc. All rights reserved.
# // See the file LICENSE for licensing terms.

set -e

# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"

# e.g.,
# ./scripts/tests.load.sh /tmp/avalanchego
#
# run without e2e tests
# ./scripts/tests.load.sh /tmp/avalanchego
#
# to run E2E tests (terminates cluster afterwards)
# E2E=true ./scripts/tests.load.sh 1.7.13
if ! [[ "$0" =~ scripts/tests.load.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

AVALANCHEGO_PATH=$1
if [[ -z "${AVALANCHEGO_PATH}" ]]; then
  echo "Missing avalanchego argument!"
  echo "Usage: ${0} [AVALANCHEGO_PATH]" >> /dev/stderr
  exit 255
fi

############################
echo "copying avalanchego"
LOAD_PATH=/tmp/timestampvm-load
rm -rf ${LOAD_PATH}
mkdir ${LOAD_PATH}
cp ${AVALANCHEGO_PATH} ${LOAD_PATH}

############################
echo "building timestampvm"
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
LOAD_PLUGIN_DIR=${LOAD_PATH}/plugins

# delete previous (if exists)
rm -f ${LOAD_PLUGIN_DIR}/tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH

go build \
-o ${LOAD_PLUGIN_DIR}/tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH \
./main/


############################
echo "updating perms"
chmod -R 755 ${LOAD_PATH}

############################
echo "creating genesis file"
echo -n "e2e" >> ${LOAD_PATH}/.genesis

############################

############################

echo "creating vm config"
cat <<EOF > ${LOAD_PATH}/.config
{}
EOF

############################

############################
echo "building load.test"
# to install the ginkgo binary (required for test build and run)
go install -v github.com/onsi/ginkgo/v2/ginkgo@v2.1.4
ACK_GINKGO_RC=true ginkgo build ./tests/load

#################################
# download avalanche-network-runner
# https://github.com/ava-labs/avalanche-network-runner
ANR_REPO_PATH=github.com/ava-labs/avalanche-network-runner
ANR_VERSION=e3f5816ca8a7508d359115a9c75e6bcb54a546a8
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
--avalanchego-path=${LOAD_PATH}/avalanchego \
--avalanchego-plugin-dir=${LOAD_PLUGIN_DIR} \
--vm-genesis-path=${LOAD_PATH}/.genesis \
--vm-config-path=${LOAD_PATH}/.config

############################
# load.test" already terminates the cluster
# just in case load tests are aborted, manually terminate them again
echo "network-runner RPC server was running on PID ${PID} as test mode; terminating the process..."
pkill -P ${PID} || true
kill -2 ${PID} || true
pkill -9 -f tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH || true # in case pkill didn't work
exit ${STATUS}
