// Package asm assembles x86-64 source text.
//
// It is the second front door onto the encoder in the parent package, the same
// arrangement arm64/asm has: the typed surface is for when the instruction is
// known while the Go is written, and this is for when the instruction is text.
// Both reach one ISA table and one encoder.
//
// # Syntax
//
// AT&T, as GNU as writes it and as every inline asm statement in a C header
// is written: source before destination, `%` on registers, `$` on immediates,
// `disp(base, index, scale)` addressing, and the operand width carried in the
// mnemonic's suffix. This is the one architecture in the tree where the
// operand grammar genuinely forks — Intel syntax is the same instruction set
// under a different reading — and the fork is handled as a mode on one
// matcher rather than as a second package.
//
// The suffix is not part of the mnemonic the table declares. `movq` and
// `movl` are both `mov` there, distinguished by their operands, which is the
// right way round: the width of `mov %rax, %rbx` is not in doubt and the
// suffix would be redundant. The suffix matters exactly when the operands
// leave the width open — `movq $1, (%rax)` — and it is read as a width for
// the memory operand rather than as part of the name.
package asm
