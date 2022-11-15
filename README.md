# Timestamp Virtual Machine

[![Lint+Test+Build](https://github.com/ava-labs/timestampvm/actions/workflows/lint_test_build.yml/badge.svg)](https://github.com/ava-labs/timestampvm/actions/workflows/lint_test_build.yml)

Avalanche is a network composed of multiple blockchains. Each blockchain is an instance of a [Virtual Machine (VM)](https://docs.avax.network/learn/platform-overview#virtual-machines), much like an object in an object-oriented language is an instance of a class. That is, the VM defines the behavior of the blockchain.

TimestampVM defines a blockchain that is a timestamp server. Each block in the blockchain contains the timestamp when it was created along with a 32-byte piece of data (payload). Each block’s timestamp is after its parent’s timestamp. This VM demonstrates capabilities of custom VMs and custom blockchains. For more information, see: [Create a Virtual Machine](https://docs.avax.network/build/tutorials/platform/create-a-virtual-machine-vm)

## Running the VM
[`scripts/run.sh`](scripts/run.sh) automatically installs [avalanchego], sets up a local network,
and creates a `blobvm` genesis file. To build and run E2E tests, you need to set the variable `E2E` before it: `E2E=true ./scripts/run.sh 1.7.11`

_See [`tests/e2e`](tests/e2e) to see how it's set up and how its client requests are made._

```bash
# to startup a local cluster (good for development)
cd ${HOME}/go/src/github.com/ava-labs/blobvm
./scripts/run.sh 1.9.2

# to run full e2e tests and shut down cluster afterwards
cd ${HOME}/go/src/github.com/ava-labs/blobvm
E2E=true ./scripts/run.sh 1.9.2

# inspect cluster endpoints when ready
cat /tmp/avalanchego-v1.9.2/output.yaml
<<COMMENT
endpoint: /ext/bc/2VCAhX6vE3UnXC6s1CBPE6jJ4c4cHWMfPgCptuWS59pQ9vbeLM
logsDir: ...
pid: 12811
uris:
- http://127.0.0.1:9650
- http://127.0.0.1:9652
- http://127.0.0.1:9654
- http://127.0.0.1:9656
- http://127.0.0.1:9658
network-runner RPC server is running on PID 66810...

use the following command to terminate:

pkill -P 66810 && kill -2 66810 && pkill -9 -f tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH

# propose a block
curl -X POST --data '{
    "jsonrpc": "2.0",
    "method": "timestampvm.proposeBlock",
    "params":{
        "data":"0x01020304000000000000000000000000000000000000000000000000000000003f004e9c"
    },
    "id": 1
}' -H 'content-type:application/json;' http://127.0.0.1:9652/ext/bc/2W3Gn3E3xKSeHQZP47iybpgH6pk3JRWbNQs9P2FrKvXcHSNteB
<<COMMENT
{"jsonrpc":"2.0","result":{"Success":true},"id":1}
COMMENT

# view last accepted block
curl -X POST --data '{
    "jsonrpc": "2.0",
    "method": "timestampvm.getBlock",
    "params":{},
    "id": 1
}' -H 'content-type:application/json;' http://127.0.0.1:9652/ext/bc/2W3Gn3E3xKSeHQZP47iybpgH6pk3JRWbNQs9P2FrKvXcHSNteB
<<COMMENT
{"jsonrpc":"2.0","result":{"timestamp":"1668475950","data":"0x01020304000000000000000000000000000000000000000000000000000000003f004e9c","height":"1","id":"2RbyqtZcr8DWnxWjD2jLaPUsjd2cxMFbjz1kmJjR7gDpp3txvz","parentID":"SdVstz8FpkYxsneD2XQDk2CK7d1EBe4YVqkhftgbvUiyFfeHJ"},"id":1}
COMMENT

# terminate cluster
pkill -P 66810 && kill -2 66810 && pkill -9 -f tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH
```

## Load Testing VM
```bash
./scripts/tests.load.sh [AVALANCHEGO_PATH]
```
