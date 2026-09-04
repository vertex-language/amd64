// Package reg declares the AMD64 register file: one Go type per register
// class, one constant per register, and nothing else.
//
// Each width is a distinct type. That is the whole point: a helper declared
// as AddR64RM64(dst R64, ...) cannot be handed an EAX, because EAX is an R32,
// and the mistake is a compile error rather than a runtime ErrForm.
//
// This package imports nothing, from this tree or the standard library. It is
// the bottom of the import graph and it stays there.
//
// Spelling: numbered registers always carry their width suffix here — R8Q,
// R8D, R8W, R8B — because the type name R8 already claims the identifier the
// 64-bit r8 would want. The root package re-exports R8Q as amd64.R8, which is
// the spelling callers actually write.
package reg

// Operand is the seal on the operand interface. Only this package can
// declare the unexported method, so the set of things that can be an
// instruction operand is closed at the bottom of the tree — where the
// registers are — rather than above it, where a register could not satisfy it.
//
// Package operand embeds Seal in Imm, Mem, Label and SymRef; the register
// types below implement operand() directly.
type Operand interface {
	operand()
}

// Seal is embedded by every operand type declared outside this package.
//
//	type Imm struct {
//	    reg.Seal
//	    ...
//	}
type Seal struct{}

func (Seal) operand() {}
