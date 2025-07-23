// Copyright (c) 2025 Sonic Operations Ltd
// SPDX-License-Identifier: BSL-1.1
//
// Direct BSC EVM Interpreter Benchmarks - Fair comparison with LFVM
// This benchmarks the BSC EVMInterpreter.Run method directly without
// the overhead of full EVM context processing.
package main

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

// --- Test vectors ----------------------------------------------------------

const (
	// PUSH1 1; PUSH1 2; ADD; POP; STOP - simpler arithmetic without jump
	simpleArithmeticCode  = "60016002015000"
	simpleArithmeticCode2 = "6001600201600357"
)

// createMinimalBSCEVM creates a BSC EVM with minimal overhead for interpreter benchmarking
func createMinimalBSCEVM() *vm.EVM {
	// Initialize EVM with proper BSC configuration but minimal state
	db := rawdb.NewMemoryDatabase()
	trieDB := triedb.NewDatabase(db, nil)
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(trieDB, nil))

	// Create test account with some BNB for gas
	testAddr := common.HexToAddress("0x1000000000000000000000000000000000000001")
	statedb.CreateAccount(testAddr)
	balance := uint256.NewInt(1000000000000000000) // 1 BNB
	statedb.SetBalance(testAddr, balance, tracing.BalanceChangeUnspecified)

	// BSC Chain configuration (BSC Testnet config)
	chainConfig := &params.ChainConfig{
		ChainID:             big.NewInt(97), // BSC Testnet
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		ArrowGlacierBlock:   big.NewInt(0),
		GrayGlacierBlock:    big.NewInt(0),
		MergeNetsplitBlock:  big.NewInt(0),
		ShanghaiTime:        new(uint64), // Enable Shanghai for PUSH0
		CancunTime:          new(uint64), // Enable Cancun for PUSH0
	}
	vmConfig := vm.Config{}

	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(1),
		Time:        uint64(1681338455), // Set to a time after Shanghai activation
		Difficulty:  big.NewInt(1),
		GasLimit:    10000000000,   // 10B gas limit
		BaseFee:     big.NewInt(0), // BSC has 0 base fee
	}

	return vm.NewEVM(blockContext, statedb, chainConfig, vmConfig)
}

// execInterpreterDirect calls the BSC EVMInterpreter.Run method directly with pre-created interpreter and contract
func execInterpreterDirect(interpreter *vm.EVMInterpreter, contract *vm.Contract, code []byte, gas uint64) ([]byte, error) {
	// Reset contract state for fresh execution
	contract.Code = code
	contract.Input = []byte{}
	contract.Gas = gas

	// Execute directly through interpreter - this is the equivalent of tosca's interpreter.Run()
	return interpreter.Run(contract, []byte{}, false)
}

// createInterpreterAndContract creates the interpreter and contract outside the benchmark loop
func createInterpreterAndContract() (*vm.EVMInterpreter, *vm.Contract) {
	evm := createMinimalBSCEVM()
	interpreter := evm.Interpreter()

	// Create contract once
	callerAddr := common.HexToAddress("0x1000000000000000000000000000000000000001")
	contractAddr := common.HexToAddress("0x0100000000000000000000000000000000000000")
	contract := vm.NewContract(
		vm.AccountRef(callerAddr),
		vm.AccountRef(contractAddr),
		uint256.NewInt(0),
		10000, // default gas, will be reset per execution
	)

	return interpreter, contract
}

// ---------------------------------------------------------------------------
// Benchmarks – Direct interpreter benchmarks for fair comparison with LFVM
// ---------------------------------------------------------------------------

func BenchmarkInterpreterSimpleOperations(b *testing.B) {
	interpreter, contract := createInterpreterAndContract()
	code, _ := hex.DecodeString(simpleArithmeticCode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := execInterpreterDirect(interpreter, contract, code, 10_000); err != nil {
			b.Fatalf("Interpreter exec failed: %v", err)
		}
	}
}

func BenchmarkInterpreterCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		evm := createMinimalBSCEVM()
		_ = evm.Interpreter() // Access interpreter to ensure it's created
	}
}

func BenchmarkInterpreterBasicOperations(b *testing.B) {
	testCases := map[string]string{
		"PUSH_POP": "60016000506000",       // PUSH1 1; PUSH1 0; POP; PUSH1 0
		"ADD_SUB":  "60016002016001036000", // PUSH1 1; PUSH1 2; ADD; PUSH1 1; SUB; PUSH1 0
		"MUL_DIV":  "60036002026003046000", // PUSH1 3; PUSH1 2; MUL; PUSH1 3; DIV; PUSH1 0
		"DUP_SWAP": "600180908100",         // PUSH1 1; DUP1; SWAP1; POP; STOP
	}

	for name, hexCode := range testCases {
		b.Run(name, func(b *testing.B) {
			interpreter, contract := createInterpreterAndContract()
			code, _ := hex.DecodeString(hexCode)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := execInterpreterDirect(interpreter, contract, code, 10_000); err != nil {
					b.Fatalf("%s failed: %v", name, err)
				}
			}
		})
	}
}

// Benchmark interpreter with simple storage operation (like BEP20 benchmark in other tests)
func BenchmarkInterpreterStorageOperation(b *testing.B) {
	interpreter, contract := createInterpreterAndContract()
	// PUSH1 1, PUSH1 2, SSTORE - simple storage operation
	code, _ := hex.DecodeString("6001600255")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := execInterpreterDirect(interpreter, contract, code, 25_000); err != nil {
			b.Fatalf("Storage operation failed: %v", err)
		}
	}
}

// Benchmark for pure interpreter execution (most equivalent to Tosca benchmark)
func BenchmarkPureInterpreterExecution(b *testing.B) {
	interpreter, contract := createInterpreterAndContract()
	code, _ := hex.DecodeString(simpleArithmeticCode)

	// Pre-set contract code to eliminate any setup overhead in the loop
	contract.Code = code
	contract.Input = []byte{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Only reset gas - this is the absolute minimal overhead
		contract.Gas = 10_000
		if _, err := interpreter.Run(contract, []byte{}, false); err != nil {
			b.Fatalf("Pure execution failed: %v", err)
		}
	}
}
