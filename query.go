package amd64

import "github.com/vertex-language/amd64/internal/isa"

// MemBytes is the width, in bytes, that every form of a mnemonic gives a
// memory operand — or false when its forms disagree and the width has to come
// from somewhere else.
//
// It is the one question about the table that an assembler cannot answer from
// the instruction it is reading. AT&T syntax puts a memory operand's width in
// the mnemonic's suffix or in a register beside it, and some instructions
// offer neither: CLFLUSH takes one operand and it is an address, LEA's operand
// has no width at all, and CMPXCHG16B's is sixteen bytes because there is no
// other size it could be. For those the table is what knows.
//
// Zero with ok true is an address with no access width, which is a real
// answer and not a missing one. False means the mnemonic's forms disagree —
// MOV, ADD, and most of the table — and a caller reading assembly should
// require the suffix its user left out rather than choose.
//
// The isa package is internal, and this is a question about it rather than a
// value from it: the answer is an int and a bool.
func MemBytes(mnemonic string) (n int, ok bool) { return isa.MemBytes(mnemonic) }

// HasMnemonic reports whether the table declares any form of this mnemonic.
//
// It is the question an AT&T front end has to ask before it peels a width
// letter off a name: CMOVL is an instruction and not CMOV with an "l", and
// the two are told apart by the table declaring the first. Asking is also
// what makes the peeling rule complete in the other direction — MULL is MUL
// with a suffix precisely because the table has no MULL.
func HasMnemonic(name string) bool { return len(isa.ByMnemonic(name)) > 0 }
