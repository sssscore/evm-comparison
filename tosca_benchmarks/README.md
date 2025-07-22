# TOSCA LFVM Standalone Benchmarks

This folder contains standalone TOSCA LFVM benchmarks rewritten from `bep20_simple_benchmark_test.go` for independent execution.

## Files

- `tosca_benchmark_test.go` - Main benchmark file with TOSCA LFVM performance tests
- `go.mod` - Go module configuration with local TOSCA dependency
- `README.md` - This documentation file

## Benchmarks Included

1. **BenchmarkSimpleOperations** - Basic arithmetic operations
2. **BenchmarkBEP20BytecodeConversion** - Bytecode conversion performance
3. **BenchmarkInterpreterCreation** - Interpreter initialization overhead
4. **BenchmarkBasicEVMOperations** - Core EVM opcodes (PUSH, POP, ADD, SUB, MUL, DIV, DUP, SWAP)

## Usage

Run all benchmarks:
```bash
go test -bench=.
```

Run specific benchmark:
```bash
go test -bench=BenchmarkBasicEVMOperations
```

Run with memory profiling:
```bash
go test -bench=. -benchmem
```

Run with CPU profiling:
```bash
go test -bench=. -cpuprofile=cpu.prof
```

## Key Features

- **Standalone execution** - No external dependencies beyond TOSCA
- **Pure computational tests** - Avoids storage operations that require state management
- **Apples-to-apples comparison ready** - Can be compared directly with BSC benchmarks
- **Gas tracking** - Monitors gas usage for performance analysis

## Expected Performance

Based on previous testing:
- TOSCA LFVM operations: ~170-220 ns/op
- 4-5x faster than traditional EVM implementations
- Low memory overhead and efficient bytecode execution

## Fully Standalone

This benchmark is **completely standalone** and will:
- ✅ Download TOSCA dependencies automatically from GitHub
- ✅ No need for local TOSCA installation
- ✅ Works anywhere with Go 1.24+ installed
- ✅ All dependencies resolved via `go mod tidy`

## Sample Results

```
BenchmarkInterpreterCreation-8                  8289200    138.9 ns/op
BenchmarkBasicEVMOperations/PUSH_POP-8          2762385    196.9 ns/op
BenchmarkBasicEVMOperations/ADD_SUB-8           2550252    215.7 ns/op
BenchmarkBasicEVMOperations/MUL_DIV-8           2755003    236.2 ns/op
BenchmarkBasicEVMOperations/DUP_SWAP-8          3414596    170.6 ns/op
```