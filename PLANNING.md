# Stacks SDK

The Stacks SDK project aims to provide a generic VM Development Framework for VMs on Avalanche.

Stacks SDK takes the approach of building VMs through composable and interchangeable modules based on the needs of the VM developer.

## Stacks

1. ChainStack - VM handles its own block indexing, active state, and state transitions
2. BlockStack - VM handles its own active state and state transitions (uses a state module to track state for processing blocks)
3. TxStack - VM handles definition of its own block format, transaction format, and Transaction Execution Engine
4. Define different transaction execution engines

## Examples

1. ChainStack - TimestampVM that handles its own block indexing
2. BlockStack - TimestampVM that handles its own active state and state transitions
3. TxStack - TimestampVM with timestamp per transaction
4. ParallelTxStack - TimestampVM with timestamp per transaction executed in parallel

## TODO

1. Finish TimestampVM ChainStack example v1 with mempool
2. Abstract Accept/Reject from calling on the VM to verify returns a `Decidable`
3. Re-write example
4. Write basic unit tests and e2e tests

5. Implement desired interface for BlockStack with chain indexing abstracted and generic state management abstracted away
6. Implement BlockStack
7. Write basic unit tests and e2e tests
8. TxStack and Execution Engines

## What are the requirements of a VM Implementation

1. GitHub actions for linting, unit tests, CodeQL, E2E tests
2. Dockerfile and build image script
3. Simple CLI to interact with it
4. WAILS app to interact with the blockchain built on top of the CLI/grpc layer to call APIs from JS and build a simple site
5. Integration with Core (what does integration with tooling look like)
6. Integrating with data ingestion 

## Functional Requirements of a Useful VM

1. State Sync
2. Light client
3. Developer tooling to interact with it: grpc interface for calling existing APIs
4. Ledger Wallet integration
5. 
