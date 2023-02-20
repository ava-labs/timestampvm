// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

// The goal of this project is to build a VM SDK based on stackable modules.
// This comes from experience working on and maintaining a variety of different VMs.
// When there are multiple VMs that contain vastly different logic it is difficult to compare and
// assess them for correctness. Therefore, this project is intended to build layers of abstraction
// on top of each other, so that any type of VM project can be built on top of the same modular stack.
// This is an alternative to the hyper-sdk approach, where and entire SDK is built from the ground up
// with the intention of supporting a highly optimized use case.
// This project's primary goal is to design for obvious correctness and simple to use and understand
// abstractions for obvious and clear blockchain development.
// Whereas the Hyper-SDK takes on complexity, but hides it from the user, so that the user never needs
// to worry about any of the optimizations underneath, this project aims to make it simple to understand
// the full codebase for anyone that wants to understand the full stack.

// This project is intended to be built to support the following types of VM projects:
// 1. A VM that handles its own block indexing and needs to fit the Avalanche Consensus engine invariants (MoveVM, Solana, etc. fork friendly interface)
// 2. A VM that only defines a block state transition (fork friendly potentially deeper changes)
// 3. A VM that defines a block format and a transaction format
// 4. Transaction execution engines - sequential execution engine, OCC execution engine, parallel execution engine with explicit R/W sets declared by txs

// TimestampVM implementations on simple AvalancheGo database:
// 1. Implement TimestampVM that handles its own block indexing on accept. (implement networking protocol registry and mempool) (full)
// 2. Implement TimestampVM that only handles its own live state and uses an SDK for access to accepted chain index. (block)
// 3. Implement TimestampVM that defines a block format and a transaction format. (tx)
// 4. Implement TimestampVM that uses an off the shelf block format and a fill in the blank transaction format. (simpletx)
// 5. Implement TimestampVM with different transaction execution engines. (paralleltx)

// Re-implement on top of arbitrary state backend that supports state sync.

// Separate concerns:
// State sync on top of a given state - if you use an out of the box state module, you should be able to import state sync with only a few lines of code
// Mempool - if you implement a simple interface for quick verification of blobs, you should be able to import a mempool with various settings
// Networking - it should be simple to register a networking protocol that does not conflict with existing protocols
// ChainRules - ease of upgradability
// Light client support - ease of supporting light clients
// Host calls - customizability on top of whatever you create WASM Host Calls or EVM Precompiles - this is a powerful notion
// Metrics and Observability - the SDK should provide metrics and examples that enforce precise metrics about a VM
// Performance - tracing and e2e testing for performance testing should be a first class concern
// Testing - testing is a first class concern and far from an after though. The SDK will provide detailed tooling for every layer of the stack
// and abstraction, so that VM implementers can easily test their code for correctness.

// What do I want the VM to implement
// Define a Database format for the VM
// Define block building logic
// Define block execution and accept/reject logic
// Define APIs that can build on top of that
// Define Networking with libp2p style protocol registration
// Define a CLI tool that can fit neatly into a broader interface
