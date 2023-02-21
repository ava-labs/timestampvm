#!/usr/bin/env bash
set -e

ROOT_DIR=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)
# Load the constants
source "$ROOT_DIR"/scripts/constants.sh

############################
# Download an AvalancheGo release from GitHub
# https://github.com/ava-labs/avalanchego/releases
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
BASEDIR=${BASE_DIR:-"/tmp/avalanchego-release"}
mkdir -p ${BASEDIR}
AVAGO_DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/${AVALANCHEGO_VERSION}/avalanchego-linux-${GOARCH}-${AVALANCHEGO_VERSION}.tar.gz
AVAGO_DOWNLOAD_PATH=${BASEDIR}/avalanchego-linux-${GOARCH}-${AVALANCHEGO_VERSION}.tar.gz
if [[ ${GOOS} == "darwin" ]]; then
  AVAGO_DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/${AVALANCHEGO_VERSION}/avalanchego-macos-${AVALANCHEGO_VERSION}.zip
  AVAGO_DOWNLOAD_PATH=${BASEDIR}/avalanchego-macos-${AVALANCHEGO_VERSION}.zip
fi

AVALANCHEGO_BUILD_PATH=${AVALANCHEGO_BUILD_PATH:-${BASEDIR}/avalanchego-${AVALANCHEGO_VERSION}}
mkdir -p $AVALANCHEGO_BUILD_PATH

if [[ ! -f ${AVAGO_DOWNLOAD_PATH} ]]; then
  echo "downloading avalanchego ${AVALANCHEGO_VERSION} at ${AVAGO_DOWNLOAD_URL} to ${AVAGO_DOWNLOAD_PATH}"
  curl -L ${AVAGO_DOWNLOAD_URL} -o ${AVAGO_DOWNLOAD_PATH}
fi
echo "extracting downloaded avalanchego to ${AVALANCHEGO_BUILD_PATH}"
if [[ ${GOOS} == "linux" ]]; then
  mkdir -p ${AVALANCHEGO_BUILD_PATH} && tar xzvf ${AVAGO_DOWNLOAD_PATH} --directory ${AVALANCHEGO_BUILD_PATH} --strip-components 1
elif [[ ${GOOS} == "darwin" ]]; then
  unzip ${AVAGO_DOWNLOAD_PATH} -d ${AVALANCHEGO_BUILD_PATH}
  mv ${AVALANCHEGO_BUILD_PATH}/build/* ${AVALANCHEGO_BUILD_PATH}
  rm -rf ${AVALANCHEGO_BUILD_PATH}/build/
fi

AVALANCHEGO_PATH=${AVALANCHEGO_BUILD_PATH}/avalanchego
AVALANCHEGO_PLUGIN_DIR=${AVALANCHEGO_BUILD_PATH}/plugins

echo "Installed AvalancheGo release ${AVALANCHEGO_VERSION}"
echo "AvalancheGo Path: ${AVALANCHEGO_PATH}"
echo "Plugin Dir: ${AVALANCHEGO_PLUGIN_DIR}"
