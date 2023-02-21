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
4. Write unit tests

e2e tests and load test

5. Implement desired interface for BlockStack with chain indexing abstracted away
6. Implement BlockStack
7. Write unit tests, e2e tests, and load test
8. TxStack 