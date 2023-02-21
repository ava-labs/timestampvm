#!/usr/bin/env bash

# Versions
AVALANCHE_VERSION=${AVALANCHE_VERSION:-'v1.9.7'}

# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"

# Test parameters
UNIT_TEST_TIMEOUT=${UNIT_TEST_TIMEOUT:-"3m"}
GINKGO_VERSION=${GINKGO_VERSION:-'v2.2.0'}