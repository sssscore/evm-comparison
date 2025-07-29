### LFVM vs BSC — nanoseconds per operation (ns/op)

| Test case                 | LFVM (ns/op) | BSC (ns/op) | Speed‑up<sup>†</sup> |
| ------------------------- | ------------ | ----------- | -------------------- |
| SWAP1\_POP                | **378.7**    | **370.8**   |  1.02 ×              |
| POP\_POP                  | 355.5        | 190.5       |  1.87 ×              |
| SWAP2\_SWAP1              | 350.6        | 203.7       |  1.72 ×              |
| ISZERO\_PUSH2\_JUMPI      | 288.8        | 160.1       |  1.80 ×              |
| PUSH1\_SHL                | 635.9        | 355.2       |  1.79 ×              |
| DUP2\_LT                  | 397.9        | 238.2       |  1.67 ×              |
| SWAP1\_POP → SWAP2\_SWAP1 | 448.7        | 266.7       |  1.68 ×              |
| FUNCTION\_CALL\_CLEANUP   | 492.4        | 294.9       |  1.67 ×              |
| DUP2\_MSTORE              | 651.6        | 454.8       |  1.43 ×              |
| PUSH1\_DUP1               | 690.9        | 471.7       |  1.46 ×              |
| MIXED\_SUPER\_PATTERNS    | 409.4        | 259.2       |  1.58 ×              |
| NESTED\_SUPER\_PATTERNS   | 443.1        | 274.6       |  1.61 ×              |

<sup>† Speed‑up = LFVM ns/op ÷ BSC ns/op. Values > 1 mean BSC executes the pattern that many times faster.</sup>
