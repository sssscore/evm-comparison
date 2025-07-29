// Copyright (c) 2025 Sonic Operations Ltd
// SPDX-License-Identifier: BSL-1.1
//
// Direct BSC EVM Interpreter Benchmarks - Fair comparison with LFVM
// This benchmarks the BSC EVMInterpreter.Run method directly without
// the overhead of full EVM context processing.
package main

import (
	"encoding/hex"
	"fmt"
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

// Test patterns that benefit from super instructions
var superInstructionPatternsFromTosca = map[string]string{
	// Stack Manipulation Super Instructions
	"SWAP1_POP":   "60016002809150600360048091506005600680915060076008809150600960108091506000", // Multiple SWAP1_POP
	"POP_POP":     "6001600250506003600450506005600650506007600850506009600a50506000",           // Multiple POP_POP
	"SWAP2_SWAP1": "60016002600391906004600560069190600760086009919060106011601291906000",       // Multiple SWAP2_SWAP1
	//"SWAP2_POP":   "60016002600392506004600592506006600792506008600992506010601192506000",       // Multiple SWAP2_POP

	// Control Flow Super Instructions
	//"POP_JUMP":           "6010506056600c565b6001601c565b6002602c565b60006035565b6001600101005b6000", // POP + JUMP patterns
	//"PUSH2_JUMP": "6100565b6001601c6100565b6002602c6100565b60006035565b6000", // PUSH2 + JUMP
	//"PUSH2_JUMPI":        "600161005760026020576003603057600460405760006045565b6000",                 // PUSH2 + JUMPI
	"ISZERO_PUSH2_JUMPI": "60001561005760011561205760021562005760031563005760006000", // ISZERO + PUSH2 + JUMPI

	// Memory Operations Super Instructions
	"DUP2_MSTORE": "600160028152600360048152600560068152600760088152600960108152600060005160205160405160605160805160a05160c05160e0516000", // Multiple DUP2_MSTORE

	// Arithmetic Super Instructions
	//"PUSH1_ADD": "6001600101600201600301600401600501600601600701600801600901600a01600b01600c01600d01600e01600f01601001601101601201601301601401601501601601601701601801601901601a01601b01601c01601d01601e01601f01602001600050505050505050505050505050505050505050505050505050505050505050505050505050506000", // Multiple PUSH1_ADD
	"PUSH1_SHL": "6001601b6002601b6003601b6004601b6005601b6006601b6007601b6008601b6009601b600a601b600b601b600c601b600d601b600e601b600f601b6010601b60115050505050505050505050505050506000", // Multiple PUSH1_SHL
	"DUP2_LT":   "6001600281106003600481106005600681106007600881106009601081105050505050506000",                                                                                           // Multiple DUP2_LT

	// Complex Multi-Instruction Super Instructions
	//"SWAP2_SWAP1_POP_JUMP":      "600160026003919050566005600660079190505660086009601091905056600b600c600d9190505660006000",     // Complex pattern
	"SWAP1_POP_SWAP2_SWAP1": "60016002600380915091906004600560068091509190600760086009809150919060106011601280915091906000", // Complex stack ops
	//"POP_SWAP2_SWAP1_POP":   "60016002600350919050600460056006509190506007600860095091905060106011601250919050600060006000", // Complex cleanup
	//"AND_SWAP1_POP_SWAP2_SWAP1": "600f601016809150919060ff60201680915091906001600281680915091906003600481680915091906000",       // Bitwise + stack

	// Constant Loading Super Instructions
	//"PUSH1_PUSH1": "60016002600360046005600660076008600960106011601260136014601560166017601860196020600050505050505050505050505050505050505050505050506000",                                                         // Multiple PUSH1_PUSH1
	"PUSH1_DUP1": "6001806002806003806004806005806006806007806008806009806010806011806012806013806014806015806016806017806018806019806020805050505050505050505050505050505050505050505050505050505050505050506000", // Multiple PUSH1_DUP1

	// Complex Real-World Patterns
	"FUNCTION_CALL_CLEANUP": "60016002600391805091906004600560068091509190600760086009809150919060106011601280915091905050505050506000", // Function epilogue pattern
	//"LOOP_COUNTER_PATTERN":  "6001600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600101600150505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050506000", // Heavy loop pattern
	//"MEMORY_COPY_PATTERN":   "600160008152602060208152604060408152606060608152608060808152602060006020600037604060206040604037606060406060606037608060606080608037600060805160605160405160205160005160c05160a05160805160605160405160205160005160c05160a05160805160605160405160205160005160c05160a051608051606051604051602051600051600050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050505050506000",                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               // Memory operations

	// Edge Cases for Super Instructions
	"MIXED_SUPER_PATTERNS":  "600160028091506003600481526005600682526007600810600901600a1b600b600c91906000",                     // Mixed super instruction types
	"NESTED_SUPER_PATTERNS": "600160026003919080915060046005600691908091506007600860099190809150600a600b600c919080915060006000", // Nested patterns
	//"CASCADING_SUPER_INSTR": "600160020180915060036004815260056006825260076008109060090160001b600c600d9190506000",               // Cascading optimizations
}

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

// Extensive opcode coverage benchmarks
func BenchmarkExtensiveOpcodesCoverage(b *testing.B) {
	testCases := map[string]string{
		// Arithmetic Operations
		"ArithmeticIntensive": "60036002016004600260020160056003600401600660046002016007600560030160086006600401505050505050506000", // Multiple arithmetic ops with proper stack management
		"ModularArithmetic":   "600a600b600c0960080a600d600e0b6000",                                                                 // ADDMOD, MULMOD, MOD
		"BitwiseOperations":   "600f601016601117601218601319601a1a601b1b6000",                                                       // AND, OR, XOR, NOT, BYTE

		// Comparison Operations
		"ComparisonOps": "60056006106007600810600960081260056003136000", // LT, GT, SLT, SGT, EQ, ISZERO

		// Stack Operations - More Complex
		"DeepStackOps":    "6001600260036004600560066007600880818283848586878889808100",   // Multiple DUPs and SWAPs
		"StackManipHeavy": "600160028060038160048260058360068460078590919293949596979899", // Heavy stack manipulation
		"StackBoundaries": "6001808080808080808080808080808080505050505050505050506000",   // Push to near-limit, then pop

		// Memory Operations
		"MemoryIntensive":   "60206000526040600052606060005260806000526000516020516040516060516000",                                                                       // Multiple MSTORE/MLOAD
		"MemoryExpansion":   "60ff60ff5260ff61010052610f005160e005160d005160c005160b005160a00516090051608005160700516060051605005160400516030051602005160100516000516000", // Fixed memory expansion
		"MemoryCopyPattern": "60206000526040602052606060405260806060526000602060006020600037602060406020604037604060806040606037608060a060806080376000",                   // CODECOPY/CALLDATACOPY patterns

		// Hashing Operations
		"HashingIntensive": "6020600052602060006020600020602060005260206000602060002060206000526020600060206000206000", // Multiple SHA3 calls
		"HashWithMemory":   "6001600052600260205260036040526004606052600560805260a0600060a0206000",                     // SHA3 with memory expansion

		// Jump Operations - Fixed to use valid jump destinations
		"JumpPattern":      "6008565b600160010160006000",                                     // Simple valid jump pattern
		"ConditionalJumps": "6001600014156011576002600201600060006013565b600360030160006000", // Fixed conditional jumps

		// Gas Operations
		"GasOpsPattern": "5a60015a0360025a0360035a0360045a0360055a036000", // GAS opcode with arithmetic

		// Environment/Context Operations - Fixed to avoid GASPRICE issues
		"EnvironmentOps": "30343332333641424344454650505050506000", // ADDRESS, ORIGIN, CALLER, removed problematic opcodes
		"BlockOps":       "42434041444548496000",                   // TIMESTAMP, NUMBER, GASLIMIT, etc.

		// Complex Mixed Operations - Fixed stack management
		"MixedComplex": "600360020160040260050360060460070560080660090760100850505050505050506000",

		// Performance Edge Cases
		"LargeStackDepth":  buildLargeStackCode(50),          // 50 items on stack
		"MemoryBoundary":   buildMemoryBoundaryCode(),        // Memory at boundary conditions
		"JumpTableStress":  buildJumpTableStressFixed(),      // Fixed stress jump table
		"SuperInstruction": buildSuperInstructionTestFixed(), // Fixed operations
	}

	// Add super-instruction pattern benchmarks for fair comparison with LFVM - Fixed patterns
	superInstructionPatterns := map[string]string{
		"SWAP1_POP_Pattern": "6001600280915060036004809150600560068091506000",                                                                                       // Multiple SWAP1_POP patterns
		"PUSH1_ADD_Pattern": "6001600101600201600301600401600501600601600701600801600901600a01600b01600c01600d01600e01600f0150505050505050505050505050505050506000", // Fixed multiple PUSH1_ADD patterns
		"PUSH1_SHL_Pattern": "6001601b6002601b6003601b6004601b6005601b50505050506000",                                                                               // Fixed multiple PUSH1_SHL patterns
		"DUP1_POP_Pattern":  "60018050600280506003805060048050600580506000",                                                                                         // Multiple DUP1_POP patterns
		"SWAP2_POP_Pattern": "6001600260036004600592506006600792506008600992506000",                                                                                 // Fixed SWAP2_POP patterns
	}

	// Merge super-instruction patterns with regular test cases
	for name, hexCode := range superInstructionPatterns {
		testCases[name] = hexCode
	}

	testCases = superInstructionPatternsFromTosca

	for name, hexCode := range testCases {
		b.Run(name, func(b *testing.B) {
			interpreter, contract := createInterpreterAndContract()
			code, err := hex.DecodeString(hexCode)
			if err != nil {
				b.Fatalf("Failed to decode %s: %v", name, err)
			}

			// Increase gas limit for complex operations
			gasLimit := uint64(100_000)
			if name == "MemoryExpansion" || name == "LargeStackDepth" {
				gasLimit = 1_000_000
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := execInterpreterDirect(interpreter, contract, code, gasLimit); err != nil {
					b.Fatalf("%s failed: %v", name, err)
				}
			}
		})
	}
}

// Helper functions to build complex test patterns
func buildLargeStackCode(depth int) string {
	// Push numbers 1 to depth, then pop them all
	code := ""
	for i := 1; i <= depth; i++ {
		if i <= 255 {
			code += fmt.Sprintf("60%02x", i) // PUSH1 i
		} else {
			code += fmt.Sprintf("61%04x", i) // PUSH2 i
		}
	}
	// Pop all items
	for i := 0; i < depth; i++ {
		code += "50" // POP
	}
	code += "6000" // PUSH1 0 (to avoid empty stack)
	return code
}

func buildMemoryBoundaryCode() string {
	// Test memory operations at various boundaries
	return "60ff60ff52" + // PUSH1 0xff, PUSH1 0xff, MSTORE (store at 0xff)
		"61ffff61ffff52" + // PUSH2 0xffff, PUSH2 0xffff, MSTORE (store at 0xffff)
		"60206000526040602052" + // Store at 0x00, 0x20
		"60ff516101ff516102ff516000" // Load from various positions
}

func buildJumpTableStress() string {
	// Create code that uses many different opcodes to stress jump table
	return "6001" + "6002" + "01" + "6003" + "02" + "6004" + "04" + "6005" + "06" +
		"6006" + "10" + "6007" + "11" + "6008" + "12" + "6009" + "14" + "600a" + "15" +
		"600b" + "16" + "600c" + "17" + "600d" + "18" + "600e" + "19" + "600f" + "1a" +
		"80" + "81" + "82" + "83" + "90" + "91" + "92" + "93" + "a0" + "a1" + "a2" + "a3" +
		"50505050505050505050505050505050505050506000"
}

func buildJumpTableStressFixed() string {
	// Create code that uses many different opcodes with proper stack management
	return "6001" + "6002" + "01" + "6003" + "02" + "6004" + "04" + "6005" + "06" +
		"6006" + "10" + "6007" + "11" + "6008" + "12" + "6009" + "14" + "600a" + "15" +
		"600b" + "16" + "600c" + "17" + "600d" + "18" + "600e" + "19" + "600f" + "1a" +
		"6001" + "80" + "81" + "82" + "83" + "90" + "91" + "92" + "93" +
		"505050505050505050505050505050505050506000"
}

func buildSuperInstructionTest() string {
	// Test patterns that might be optimized as super-instructions
	return "80" + "50" + // DUP1, POP (common pattern)
		"91" + "50" + // SWAP2, POP
		"6001" + "01" + // PUSH1 1, ADD
		"6002" + "1b" + // PUSH1 2, SHL
		"80" + "80" + // DUP1, DUP1 (duplication pattern)
		"82" + "91" + "50" + "50" + // SWAP3, SWAP2, POP, POP (complex swap pattern)
		"6000"
}

func buildSuperInstructionTestFixed() string {
	// Test patterns that might be optimized as super-instructions - fixed version
	return "6001" + "80" + "50" + // PUSH1 1, DUP1, POP (common pattern)
		"6002" + "6003" + "91" + "50" + // PUSH1 2, PUSH1 3, SWAP2, POP
		"6001" + "01" + // PUSH1 1, ADD
		"6002" + "1b" + // PUSH1 2, SHL
		"6004" + "80" + "80" + "50" + "50" + // PUSH1 4, DUP1, DUP1, POP, POP
		"50" + "6000" // Final cleanup
}
