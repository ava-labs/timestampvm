#!/usr/bin/env bash
set -e

# This script assumes that AvalancheGo and Plugin binaries are available in the standard location
# within the $GOPATH
# The AvalancheGo and PluginDir paths can alternatively be specified by the environment variables
# used in scripts/run.sh

# Load constants
ROOT_DIR=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)

source "$ROOT_DIR"/scripts/constants.sh

echo "Installing ginkgo@$GINKGO_VERSION"
go install -v github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION}

ACK_GINKGO_RC=true ginkgo build ./examples/...

for exampleVM in ./examples/*/
do
  echo "Running Ginkgo tests for ExampleVM: $exampleVM"
  $exampleVM/tests/tests.test
done
