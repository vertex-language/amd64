// Package isa is the form table: what instructions exist, what operand
// shapes they take, what bytes they are, and what gates them.
//
// It is internal because it is implementation. The typed helpers and Emit
// are the instruction surface, and nothing a caller writes holds an isa
// type. Encoder and resolution failures reach a caller through *obj.Error
// as a sentinel, a message and notes — data, not internal types.
package isa

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// Class is one operand position's accepted shape.
type Class uint8

const (
	ClassNone Class = iota

	// General-purpose registers.
	R8
	R16
	R32
	R64

	// Register or memory, at one width.
	RM8
	RM16
	RM32
	RM64

	// An address with no access width. LEA's operand, and INVLPG's.
	M

	// Vector registers, and register-or-memory at vector widths.
	Xmm
	Ymm
	Zmm
	RM128
	RM256
	RM512

	// Scalar SSE's r/m: an xmm register, or four or eight bytes of memory.
	//
	// RM128 spells xmm/m128 because 128 bits names the register file by
	// itself. At 32 and 64 it does not — r/m32 is a general-purpose
	// operand and xmm/m32 is not — so these carry the file in the name.
	// A scalar instruction reading memory reads exactly its own operand's
	// width, which is what makes MOVSS's load four bytes and not sixteen.
	XM32
	XM64

	// Other register files.
	Sreg
	St
	MmReg
	KReg
	Cr
	Dr
	Tmm

	// Immediates. A form pins the field; a value that does not fit the
	// field is a range failure at the call, not a different form.
	//
	// Imm8 is a byte: -128 to 255, either sign, because a byte-wide
	// operation is what it is written on and 0xff is an ordinary mask.
	// Imm8S is the sign-extended byte of the 83 group and of PUSH imm8,
	// where the field is one byte and the operand is four or eight — so
	// 255 is not that operand's 255 and the wider form has to be used
	// instead. Keeping them apart is what makes Emit's shortest-wins rule
	// safe: without it, addl $200 would match the four-byte form and mean
	// -56.
	Imm8
	Imm8S
	Imm16
	Imm32
	Imm64

	// Branch displacements, resolved at Finalize.
	Rel8
	Rel32

	// Fixed operands. The form names the register and leaves no field to
	// put another one in, so the typed helper carries it in its name and
	// not in its parameters. Emit still passes it, because Emit is handed
	// the instruction as written.
	FixAL
	FixCL
	FixAX
	FixDX
	FixEAX
	FixRAX
	FixOne // the literal 1 of SHL r/m64, 1
)

var classNames = [...]string{
	ClassNone: "",
	R8:        "r8",
	R16:       "r16",
	R32:       "r32",
	R64:       "r64",
	RM8:       "r/m8",
	RM16:      "r/m16",
	RM32:      "r/m32",
	RM64:      "r/m64",
	M:         "m",
	Xmm:       "xmm",
	Ymm:       "ymm",
	Zmm:       "zmm",
	RM128:     "xmm/m128",
	RM256:     "ymm/m256",
	RM512:     "zmm/m512",
	XM32:      "xmm/m32",
	XM64:      "xmm/m64",
	Sreg:      "sreg",
	St:        "st",
	MmReg:     "mm",
	KReg:      "k",
	Cr:        "cr",
	Dr:        "dr",
	Tmm:       "tmm",
	Imm8:      "imm8",
	Imm8S:     "imm8",
	Imm16:     "imm16",
	Imm32:     "imm32",
	Imm64:     "imm64",
	Rel8:      "rel8",
	Rel32:     "rel32",
	FixAL:     "al",
	FixCL:     "cl",
	FixAX:     "ax",
	FixDX:     "dx",
	FixEAX:    "eax",
	FixRAX:    "rax",
	FixOne:    "1",
}

// String is the Intel spelling, which is what a form name is built from and
// what a diagnostic prints: "ADD r/m64, imm8".
func (c Class) String() string {
	if int(c) < len(classNames) {
		return classNames[c]
	}
	return "class?"
}

// Fixed reports whether the class names one specific operand, so the typed
// helper spells it in its name rather than taking a parameter.
func (c Class) Fixed() bool { return c >= FixAL }

// Accepts reports whether the operand can fill this position.
//
// Immediate classes test the value, not just the type: an Imm holding 300
// is not an imm8. That is what makes Emit's shortest-wins rule pick the
// four-byte sign-extended form for "add rax, 1" and the seven-byte form for
// "add rax, 300" without a size estimator anywhere.
func (c Class) Accepts(op operand.Operand) bool {
	switch c {
	case R8:
		_, ok := op.(reg.R8)
		return ok
	case R16:
		_, ok := op.(reg.R16)
		return ok
	case R32:
		_, ok := op.(reg.R32)
		return ok
	case R64:
		_, ok := op.(reg.R64)
		return ok

	case RM8:
		return isReg[reg.R8](op) || isMem[operand.M8](op)
	case RM16:
		return isReg[reg.R16](op) || isMem[operand.M16](op)
	case RM32:
		return isReg[reg.R32](op) || isMem[operand.M32](op)
	case RM64:
		return isReg[reg.R64](op) || isMem[operand.M64](op)

	case M:
		_, ok := op.(operand.Memory)
		return ok

	case Xmm:
		_, ok := op.(reg.Xmm)
		return ok
	case Ymm:
		_, ok := op.(reg.Ymm)
		return ok
	case Zmm:
		_, ok := op.(reg.Zmm)
		return ok
	case RM128:
		return isReg[reg.Xmm](op) || isMem[operand.M128](op)
	case RM256:
		return isReg[reg.Ymm](op) || isMem[operand.M256](op)
	case RM512:
		return isReg[reg.Zmm](op) || isMem[operand.M512](op)
	case XM32:
		return isReg[reg.Xmm](op) || isMem[operand.M32](op)
	case XM64:
		return isReg[reg.Xmm](op) || isMem[operand.M64](op)

	case Sreg:
		_, ok := op.(reg.Sreg)
		return ok
	case St:
		_, ok := op.(reg.St)
		return ok
	case MmReg:
		_, ok := op.(reg.Mm)
		return ok
	case KReg:
		_, ok := op.(reg.K)
		return ok
	case Cr:
		_, ok := op.(reg.Cr)
		return ok
	case Dr:
		_, ok := op.(reg.Dr)
		return ok
	case Tmm:
		_, ok := op.(reg.Tmm)
		return ok

	case Imm8:
		i, ok := op.(operand.Imm)
		return ok && (i.FitsSigned(8) || i.FitsUnsigned(8))
	case Imm8S:
		i, ok := op.(operand.Imm)
		return ok && i.FitsSigned(8)
	case Imm16:
		i, ok := op.(operand.Imm)
		return ok && i.FitsSigned(16)
	case Imm32:
		i, ok := op.(operand.Imm)
		return ok && i.FitsSigned(32)
	case Imm64:
		_, ok := op.(operand.Imm)
		return ok

	case Rel8, Rel32:
		switch op.(type) {
		case operand.Label, operand.SymRef:
			return true
		}
		return false

	case FixAL:
		r, ok := op.(reg.R8)
		return ok && r == reg.AL
	case FixCL:
		r, ok := op.(reg.R8)
		return ok && r == reg.CL
	case FixAX:
		r, ok := op.(reg.R16)
		return ok && r == reg.AX
	case FixDX:
		r, ok := op.(reg.R16)
		return ok && r == reg.DX
	case FixEAX:
		r, ok := op.(reg.R32)
		return ok && r == reg.EAX
	case FixRAX:
		r, ok := op.(reg.R64)
		return ok && r == reg.RAX
	case FixOne:
		i, ok := op.(operand.Imm)
		return ok && i.Int64() == 1
	}
	return false
}

func isReg[T reg.Value](op operand.Operand) bool {
	_, ok := op.(T)
	return ok
}

func isMem[T operand.Memory](op operand.Operand) bool {
	_, ok := op.(T)
	return ok
}

// MemBytes is the width in bytes of the memory operand this class admits, and
// whether it admits one at all. Zero bytes with ok true is an address with no
// access width: LEA's operand and INVLPG's.
func (c Class) MemBytes() (n int, ok bool) {
	switch c {
	case RM8:
		return 1, true
	case RM16:
		return 2, true
	case RM32, XM32:
		return 4, true
	case RM64, XM64:
		return 8, true
	case RM128:
		return 16, true
	case RM256:
		return 32, true
	case RM512:
		return 64, true
	case M:
		return 0, true
	}
	return 0, false
}

// MemBytes is the width every form of a mnemonic gives a memory operand, when
// they agree on one.
//
// It exists for an assembler reading AT&T, where the width of a memory operand
// normally comes from the mnemonic's suffix or from a register beside it. Some
// instructions have neither to offer — CLFLUSH takes one operand and it is an
// address — and for those the table is the only thing that knows. Asking it is
// better than a list of exceptions in the parser, and better than guessing:
// a mnemonic whose forms disagree, which is most of them, returns false and
// the caller still demands a suffix.
func MemBytes(mnemonic string) (n int, ok bool) {
	first := true
	for _, f := range ByMnemonic(mnemonic) {
		for _, c := range f.Ops {
			w, isMem := c.MemBytes()
			if !isMem {
				continue
			}
			if first {
				n, first = w, false
				continue
			}
			if w != n {
				return 0, false
			}
		}
	}
	return n, !first
}
