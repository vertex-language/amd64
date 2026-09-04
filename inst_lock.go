package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// The locking forms, opposite isa/table_lock.go's buildLock, and the
// three fences.
//
// LOCK is a prefix and not part of a mnemonic, so it is not something
// Emit can be asked for: "lock add" is not an instruction name the way
// "add" is. Naming one of these helpers is how a caller asks for it,
// which is the same reason the table declares them as rows of their own.
//
// Every one refuses a register destination at the call. LOCK on a
// register is #UD, and that is a fact about the value rather than its
// class — the class is r/m64 either way and the operand is perfectly
// well formed — so it is ErrOperand from lockInst rather than ErrForm
// from the encoder, which sees a locking clone and its base row as the
// same shape.

// The read-modify-write arithmetic and logic. Every one of these
// reads memory, computes and writes it back, and LOCK is what makes the
// three steps one indivisible operation rather than three chances for
// another processor to interleave.

var (
	lockAddRM8R8     = form("LockAddRM8R8")
	lockAddRM16R16   = form("LockAddRM16R16")
	lockAddRM32R32   = form("LockAddRM32R32")
	lockAddRM64R64   = form("LockAddRM64R64")
	lockAddRM16Imm8  = form("LockAddRM16Imm8")
	lockAddRM32Imm8  = form("LockAddRM32Imm8")
	lockAddRM64Imm8  = form("LockAddRM64Imm8")
	lockAddRM8Imm8   = form("LockAddRM8Imm8")
	lockAddRM16Imm16 = form("LockAddRM16Imm16")
	lockAddRM32Imm32 = form("LockAddRM32Imm32")
	lockAddRM64Imm32 = form("LockAddRM64Imm32")

	lockOrRM8R8     = form("LockOrRM8R8")
	lockOrRM16R16   = form("LockOrRM16R16")
	lockOrRM32R32   = form("LockOrRM32R32")
	lockOrRM64R64   = form("LockOrRM64R64")
	lockOrRM16Imm8  = form("LockOrRM16Imm8")
	lockOrRM32Imm8  = form("LockOrRM32Imm8")
	lockOrRM64Imm8  = form("LockOrRM64Imm8")
	lockOrRM8Imm8   = form("LockOrRM8Imm8")
	lockOrRM16Imm16 = form("LockOrRM16Imm16")
	lockOrRM32Imm32 = form("LockOrRM32Imm32")
	lockOrRM64Imm32 = form("LockOrRM64Imm32")

	lockAdcRM8R8     = form("LockAdcRM8R8")
	lockAdcRM16R16   = form("LockAdcRM16R16")
	lockAdcRM32R32   = form("LockAdcRM32R32")
	lockAdcRM64R64   = form("LockAdcRM64R64")
	lockAdcRM16Imm8  = form("LockAdcRM16Imm8")
	lockAdcRM32Imm8  = form("LockAdcRM32Imm8")
	lockAdcRM64Imm8  = form("LockAdcRM64Imm8")
	lockAdcRM8Imm8   = form("LockAdcRM8Imm8")
	lockAdcRM16Imm16 = form("LockAdcRM16Imm16")
	lockAdcRM32Imm32 = form("LockAdcRM32Imm32")
	lockAdcRM64Imm32 = form("LockAdcRM64Imm32")

	lockSbbRM8R8     = form("LockSbbRM8R8")
	lockSbbRM16R16   = form("LockSbbRM16R16")
	lockSbbRM32R32   = form("LockSbbRM32R32")
	lockSbbRM64R64   = form("LockSbbRM64R64")
	lockSbbRM16Imm8  = form("LockSbbRM16Imm8")
	lockSbbRM32Imm8  = form("LockSbbRM32Imm8")
	lockSbbRM64Imm8  = form("LockSbbRM64Imm8")
	lockSbbRM8Imm8   = form("LockSbbRM8Imm8")
	lockSbbRM16Imm16 = form("LockSbbRM16Imm16")
	lockSbbRM32Imm32 = form("LockSbbRM32Imm32")
	lockSbbRM64Imm32 = form("LockSbbRM64Imm32")

	lockAndRM8R8     = form("LockAndRM8R8")
	lockAndRM16R16   = form("LockAndRM16R16")
	lockAndRM32R32   = form("LockAndRM32R32")
	lockAndRM64R64   = form("LockAndRM64R64")
	lockAndRM16Imm8  = form("LockAndRM16Imm8")
	lockAndRM32Imm8  = form("LockAndRM32Imm8")
	lockAndRM64Imm8  = form("LockAndRM64Imm8")
	lockAndRM8Imm8   = form("LockAndRM8Imm8")
	lockAndRM16Imm16 = form("LockAndRM16Imm16")
	lockAndRM32Imm32 = form("LockAndRM32Imm32")
	lockAndRM64Imm32 = form("LockAndRM64Imm32")

	lockSubRM8R8     = form("LockSubRM8R8")
	lockSubRM16R16   = form("LockSubRM16R16")
	lockSubRM32R32   = form("LockSubRM32R32")
	lockSubRM64R64   = form("LockSubRM64R64")
	lockSubRM16Imm8  = form("LockSubRM16Imm8")
	lockSubRM32Imm8  = form("LockSubRM32Imm8")
	lockSubRM64Imm8  = form("LockSubRM64Imm8")
	lockSubRM8Imm8   = form("LockSubRM8Imm8")
	lockSubRM16Imm16 = form("LockSubRM16Imm16")
	lockSubRM32Imm32 = form("LockSubRM32Imm32")
	lockSubRM64Imm32 = form("LockSubRM64Imm32")

	lockXorRM8R8     = form("LockXorRM8R8")
	lockXorRM16R16   = form("LockXorRM16R16")
	lockXorRM32R32   = form("LockXorRM32R32")
	lockXorRM64R64   = form("LockXorRM64R64")
	lockXorRM16Imm8  = form("LockXorRM16Imm8")
	lockXorRM32Imm8  = form("LockXorRM32Imm8")
	lockXorRM64Imm8  = form("LockXorRM64Imm8")
	lockXorRM8Imm8   = form("LockXorRM8Imm8")
	lockXorRM16Imm16 = form("LockXorRM16Imm16")
	lockXorRM32Imm32 = form("LockXorRM32Imm32")
	lockXorRM64Imm32 = form("LockXorRM64Imm32")
)

func (s *Section) LockAddRM8R8(dst operand.RM8, src reg.R8)     { s.lockInst(lockAddRM8R8, dst, src) }
func (s *Section) LockAddRM16R16(dst operand.RM16, src reg.R16) { s.lockInst(lockAddRM16R16, dst, src) }
func (s *Section) LockAddRM32R32(dst operand.RM32, src reg.R32) { s.lockInst(lockAddRM32R32, dst, src) }
func (s *Section) LockAddRM64R64(dst operand.RM64, src reg.R64) { s.lockInst(lockAddRM64R64, dst, src) }
func (s *Section) LockAddRM16Imm8(dst operand.RM16, v int64) {
	s.lockInst(lockAddRM16Imm8, dst, imm(v))
}
func (s *Section) LockAddRM32Imm8(dst operand.RM32, v int64) {
	s.lockInst(lockAddRM32Imm8, dst, imm(v))
}
func (s *Section) LockAddRM64Imm8(dst operand.RM64, v int64) {
	s.lockInst(lockAddRM64Imm8, dst, imm(v))
}
func (s *Section) LockAddRM8Imm8(dst operand.RM8, v int64) { s.lockInst(lockAddRM8Imm8, dst, imm(v)) }
func (s *Section) LockAddRM16Imm16(dst operand.RM16, v int64) {
	s.lockInst(lockAddRM16Imm16, dst, imm(v))
}
func (s *Section) LockAddRM32Imm32(dst operand.RM32, v int64) {
	s.lockInst(lockAddRM32Imm32, dst, imm(v))
}
func (s *Section) LockAddRM64Imm32(dst operand.RM64, v int64) {
	s.lockInst(lockAddRM64Imm32, dst, imm(v))
}

func (s *Section) LockOrRM8R8(dst operand.RM8, src reg.R8)     { s.lockInst(lockOrRM8R8, dst, src) }
func (s *Section) LockOrRM16R16(dst operand.RM16, src reg.R16) { s.lockInst(lockOrRM16R16, dst, src) }
func (s *Section) LockOrRM32R32(dst operand.RM32, src reg.R32) { s.lockInst(lockOrRM32R32, dst, src) }
func (s *Section) LockOrRM64R64(dst operand.RM64, src reg.R64) { s.lockInst(lockOrRM64R64, dst, src) }
func (s *Section) LockOrRM16Imm8(dst operand.RM16, v int64)    { s.lockInst(lockOrRM16Imm8, dst, imm(v)) }
func (s *Section) LockOrRM32Imm8(dst operand.RM32, v int64)    { s.lockInst(lockOrRM32Imm8, dst, imm(v)) }
func (s *Section) LockOrRM64Imm8(dst operand.RM64, v int64)    { s.lockInst(lockOrRM64Imm8, dst, imm(v)) }
func (s *Section) LockOrRM8Imm8(dst operand.RM8, v int64)      { s.lockInst(lockOrRM8Imm8, dst, imm(v)) }
func (s *Section) LockOrRM16Imm16(dst operand.RM16, v int64) {
	s.lockInst(lockOrRM16Imm16, dst, imm(v))
}
func (s *Section) LockOrRM32Imm32(dst operand.RM32, v int64) {
	s.lockInst(lockOrRM32Imm32, dst, imm(v))
}
func (s *Section) LockOrRM64Imm32(dst operand.RM64, v int64) {
	s.lockInst(lockOrRM64Imm32, dst, imm(v))
}

func (s *Section) LockAdcRM8R8(dst operand.RM8, src reg.R8)     { s.lockInst(lockAdcRM8R8, dst, src) }
func (s *Section) LockAdcRM16R16(dst operand.RM16, src reg.R16) { s.lockInst(lockAdcRM16R16, dst, src) }
func (s *Section) LockAdcRM32R32(dst operand.RM32, src reg.R32) { s.lockInst(lockAdcRM32R32, dst, src) }
func (s *Section) LockAdcRM64R64(dst operand.RM64, src reg.R64) { s.lockInst(lockAdcRM64R64, dst, src) }
func (s *Section) LockAdcRM16Imm8(dst operand.RM16, v int64) {
	s.lockInst(lockAdcRM16Imm8, dst, imm(v))
}
func (s *Section) LockAdcRM32Imm8(dst operand.RM32, v int64) {
	s.lockInst(lockAdcRM32Imm8, dst, imm(v))
}
func (s *Section) LockAdcRM64Imm8(dst operand.RM64, v int64) {
	s.lockInst(lockAdcRM64Imm8, dst, imm(v))
}
func (s *Section) LockAdcRM8Imm8(dst operand.RM8, v int64) { s.lockInst(lockAdcRM8Imm8, dst, imm(v)) }
func (s *Section) LockAdcRM16Imm16(dst operand.RM16, v int64) {
	s.lockInst(lockAdcRM16Imm16, dst, imm(v))
}
func (s *Section) LockAdcRM32Imm32(dst operand.RM32, v int64) {
	s.lockInst(lockAdcRM32Imm32, dst, imm(v))
}
func (s *Section) LockAdcRM64Imm32(dst operand.RM64, v int64) {
	s.lockInst(lockAdcRM64Imm32, dst, imm(v))
}

func (s *Section) LockSbbRM8R8(dst operand.RM8, src reg.R8)     { s.lockInst(lockSbbRM8R8, dst, src) }
func (s *Section) LockSbbRM16R16(dst operand.RM16, src reg.R16) { s.lockInst(lockSbbRM16R16, dst, src) }
func (s *Section) LockSbbRM32R32(dst operand.RM32, src reg.R32) { s.lockInst(lockSbbRM32R32, dst, src) }
func (s *Section) LockSbbRM64R64(dst operand.RM64, src reg.R64) { s.lockInst(lockSbbRM64R64, dst, src) }
func (s *Section) LockSbbRM16Imm8(dst operand.RM16, v int64) {
	s.lockInst(lockSbbRM16Imm8, dst, imm(v))
}
func (s *Section) LockSbbRM32Imm8(dst operand.RM32, v int64) {
	s.lockInst(lockSbbRM32Imm8, dst, imm(v))
}
func (s *Section) LockSbbRM64Imm8(dst operand.RM64, v int64) {
	s.lockInst(lockSbbRM64Imm8, dst, imm(v))
}
func (s *Section) LockSbbRM8Imm8(dst operand.RM8, v int64) { s.lockInst(lockSbbRM8Imm8, dst, imm(v)) }
func (s *Section) LockSbbRM16Imm16(dst operand.RM16, v int64) {
	s.lockInst(lockSbbRM16Imm16, dst, imm(v))
}
func (s *Section) LockSbbRM32Imm32(dst operand.RM32, v int64) {
	s.lockInst(lockSbbRM32Imm32, dst, imm(v))
}
func (s *Section) LockSbbRM64Imm32(dst operand.RM64, v int64) {
	s.lockInst(lockSbbRM64Imm32, dst, imm(v))
}

func (s *Section) LockAndRM8R8(dst operand.RM8, src reg.R8)     { s.lockInst(lockAndRM8R8, dst, src) }
func (s *Section) LockAndRM16R16(dst operand.RM16, src reg.R16) { s.lockInst(lockAndRM16R16, dst, src) }
func (s *Section) LockAndRM32R32(dst operand.RM32, src reg.R32) { s.lockInst(lockAndRM32R32, dst, src) }
func (s *Section) LockAndRM64R64(dst operand.RM64, src reg.R64) { s.lockInst(lockAndRM64R64, dst, src) }
func (s *Section) LockAndRM16Imm8(dst operand.RM16, v int64) {
	s.lockInst(lockAndRM16Imm8, dst, imm(v))
}
func (s *Section) LockAndRM32Imm8(dst operand.RM32, v int64) {
	s.lockInst(lockAndRM32Imm8, dst, imm(v))
}
func (s *Section) LockAndRM64Imm8(dst operand.RM64, v int64) {
	s.lockInst(lockAndRM64Imm8, dst, imm(v))
}
func (s *Section) LockAndRM8Imm8(dst operand.RM8, v int64) { s.lockInst(lockAndRM8Imm8, dst, imm(v)) }
func (s *Section) LockAndRM16Imm16(dst operand.RM16, v int64) {
	s.lockInst(lockAndRM16Imm16, dst, imm(v))
}
func (s *Section) LockAndRM32Imm32(dst operand.RM32, v int64) {
	s.lockInst(lockAndRM32Imm32, dst, imm(v))
}
func (s *Section) LockAndRM64Imm32(dst operand.RM64, v int64) {
	s.lockInst(lockAndRM64Imm32, dst, imm(v))
}

func (s *Section) LockSubRM8R8(dst operand.RM8, src reg.R8)     { s.lockInst(lockSubRM8R8, dst, src) }
func (s *Section) LockSubRM16R16(dst operand.RM16, src reg.R16) { s.lockInst(lockSubRM16R16, dst, src) }
func (s *Section) LockSubRM32R32(dst operand.RM32, src reg.R32) { s.lockInst(lockSubRM32R32, dst, src) }
func (s *Section) LockSubRM64R64(dst operand.RM64, src reg.R64) { s.lockInst(lockSubRM64R64, dst, src) }
func (s *Section) LockSubRM16Imm8(dst operand.RM16, v int64) {
	s.lockInst(lockSubRM16Imm8, dst, imm(v))
}
func (s *Section) LockSubRM32Imm8(dst operand.RM32, v int64) {
	s.lockInst(lockSubRM32Imm8, dst, imm(v))
}
func (s *Section) LockSubRM64Imm8(dst operand.RM64, v int64) {
	s.lockInst(lockSubRM64Imm8, dst, imm(v))
}
func (s *Section) LockSubRM8Imm8(dst operand.RM8, v int64) { s.lockInst(lockSubRM8Imm8, dst, imm(v)) }
func (s *Section) LockSubRM16Imm16(dst operand.RM16, v int64) {
	s.lockInst(lockSubRM16Imm16, dst, imm(v))
}
func (s *Section) LockSubRM32Imm32(dst operand.RM32, v int64) {
	s.lockInst(lockSubRM32Imm32, dst, imm(v))
}
func (s *Section) LockSubRM64Imm32(dst operand.RM64, v int64) {
	s.lockInst(lockSubRM64Imm32, dst, imm(v))
}

func (s *Section) LockXorRM8R8(dst operand.RM8, src reg.R8)     { s.lockInst(lockXorRM8R8, dst, src) }
func (s *Section) LockXorRM16R16(dst operand.RM16, src reg.R16) { s.lockInst(lockXorRM16R16, dst, src) }
func (s *Section) LockXorRM32R32(dst operand.RM32, src reg.R32) { s.lockInst(lockXorRM32R32, dst, src) }
func (s *Section) LockXorRM64R64(dst operand.RM64, src reg.R64) { s.lockInst(lockXorRM64R64, dst, src) }
func (s *Section) LockXorRM16Imm8(dst operand.RM16, v int64) {
	s.lockInst(lockXorRM16Imm8, dst, imm(v))
}
func (s *Section) LockXorRM32Imm8(dst operand.RM32, v int64) {
	s.lockInst(lockXorRM32Imm8, dst, imm(v))
}
func (s *Section) LockXorRM64Imm8(dst operand.RM64, v int64) {
	s.lockInst(lockXorRM64Imm8, dst, imm(v))
}
func (s *Section) LockXorRM8Imm8(dst operand.RM8, v int64) { s.lockInst(lockXorRM8Imm8, dst, imm(v)) }
func (s *Section) LockXorRM16Imm16(dst operand.RM16, v int64) {
	s.lockInst(lockXorRM16Imm16, dst, imm(v))
}
func (s *Section) LockXorRM32Imm32(dst operand.RM32, v int64) {
	s.lockInst(lockXorRM32Imm32, dst, imm(v))
}
func (s *Section) LockXorRM64Imm32(dst operand.RM64, v int64) {
	s.lockInst(lockXorRM64Imm32, dst, imm(v))
}

// The unary group. CMP is not among them and cannot be: it writes
// nothing, so LOCK on it is #UD and the base row says so.

var (
	lockNotRM8  = form("LockNotRM8")
	lockNotRM16 = form("LockNotRM16")
	lockNotRM32 = form("LockNotRM32")
	lockNotRM64 = form("LockNotRM64")

	lockNegRM8  = form("LockNegRM8")
	lockNegRM16 = form("LockNegRM16")
	lockNegRM32 = form("LockNegRM32")
	lockNegRM64 = form("LockNegRM64")

	lockIncRM8  = form("LockIncRM8")
	lockIncRM16 = form("LockIncRM16")
	lockIncRM32 = form("LockIncRM32")
	lockIncRM64 = form("LockIncRM64")

	lockDecRM8  = form("LockDecRM8")
	lockDecRM16 = form("LockDecRM16")
	lockDecRM32 = form("LockDecRM32")
	lockDecRM64 = form("LockDecRM64")
)

func (s *Section) LockNotRM8(dst operand.RM8)   { s.lockInst(lockNotRM8, dst) }
func (s *Section) LockNotRM16(dst operand.RM16) { s.lockInst(lockNotRM16, dst) }
func (s *Section) LockNotRM32(dst operand.RM32) { s.lockInst(lockNotRM32, dst) }
func (s *Section) LockNotRM64(dst operand.RM64) { s.lockInst(lockNotRM64, dst) }

func (s *Section) LockNegRM8(dst operand.RM8)   { s.lockInst(lockNegRM8, dst) }
func (s *Section) LockNegRM16(dst operand.RM16) { s.lockInst(lockNegRM16, dst) }
func (s *Section) LockNegRM32(dst operand.RM32) { s.lockInst(lockNegRM32, dst) }
func (s *Section) LockNegRM64(dst operand.RM64) { s.lockInst(lockNegRM64, dst) }

func (s *Section) LockIncRM8(dst operand.RM8)   { s.lockInst(lockIncRM8, dst) }
func (s *Section) LockIncRM16(dst operand.RM16) { s.lockInst(lockIncRM16, dst) }
func (s *Section) LockIncRM32(dst operand.RM32) { s.lockInst(lockIncRM32, dst) }
func (s *Section) LockIncRM64(dst operand.RM64) { s.lockInst(lockIncRM64, dst) }

func (s *Section) LockDecRM8(dst operand.RM8)   { s.lockInst(lockDecRM8, dst) }
func (s *Section) LockDecRM16(dst operand.RM16) { s.lockInst(lockDecRM16, dst) }
func (s *Section) LockDecRM32(dst operand.RM32) { s.lockInst(lockDecRM32, dst) }
func (s *Section) LockDecRM64(dst operand.RM64) { s.lockInst(lockDecRM64, dst) }

// The atomic primitives. XCHG with a memory operand is locked whether or
// not the prefix is written — that is the one instruction on this
// architecture with an implicit LOCK — so LockXchg emits a redundant
// byte and exists because a caller writing an atomic exchange should be
// able to say so. The rest are not implicitly anything: an unlocked
// CMPXCHG on shared memory is a bug in almost every case, which is why
// both spellings exist and the caller says which.
var (
	lockXchgRM8R8   = form("LockXchgRM8R8")
	lockXchgRM16R16 = form("LockXchgRM16R16")
	lockXchgRM32R32 = form("LockXchgRM32R32")
	lockXchgRM64R64 = form("LockXchgRM64R64")

	lockCmpxchgRM8R8   = form("LockCmpxchgRM8R8")
	lockCmpxchgRM16R16 = form("LockCmpxchgRM16R16")
	lockCmpxchgRM32R32 = form("LockCmpxchgRM32R32")
	lockCmpxchgRM64R64 = form("LockCmpxchgRM64R64")

	lockXaddRM8R8   = form("LockXaddRM8R8")
	lockXaddRM16R16 = form("LockXaddRM16R16")
	lockXaddRM32R32 = form("LockXaddRM32R32")
	lockXaddRM64R64 = form("LockXaddRM64R64")

	lockCmpxchg16bM128 = form("LockCmpxchg16bM128")
)

func (s *Section) LockXchgRM8R8(dst operand.RM8, src reg.R8) {
	s.lockInst(lockXchgRM8R8, dst, src)
}

func (s *Section) LockXchgRM16R16(dst operand.RM16, src reg.R16) {
	s.lockInst(lockXchgRM16R16, dst, src)
}

func (s *Section) LockXchgRM32R32(dst operand.RM32, src reg.R32) {
	s.lockInst(lockXchgRM32R32, dst, src)
}

func (s *Section) LockXchgRM64R64(dst operand.RM64, src reg.R64) {
	s.lockInst(lockXchgRM64R64, dst, src)
}

// LockCmpxchgRM8R8 emits LOCK CMPXCHG r/m8, r8: the destination is
// compared against AL and replaced by the source if they match, and AL
// takes the destination's old value if they do not. The fixed register
// is the whole reason a compare-and-swap needs no third operand.
func (s *Section) LockCmpxchgRM8R8(dst operand.RM8, src reg.R8) {
	s.lockInst(lockCmpxchgRM8R8, dst, src)
}

// LockCmpxchgRM16R16 emits LOCK CMPXCHG r/m16, r16, against AX.
func (s *Section) LockCmpxchgRM16R16(dst operand.RM16, src reg.R16) {
	s.lockInst(lockCmpxchgRM16R16, dst, src)
}

// LockCmpxchgRM32R32 emits LOCK CMPXCHG r/m32, r32, against EAX.
func (s *Section) LockCmpxchgRM32R32(dst operand.RM32, src reg.R32) {
	s.lockInst(lockCmpxchgRM32R32, dst, src)
}

// LockCmpxchgRM64R64 emits LOCK CMPXCHG r/m64, r64, against RAX.
func (s *Section) LockCmpxchgRM64R64(dst operand.RM64, src reg.R64) {
	s.lockInst(lockCmpxchgRM64R64, dst, src)
}

// LockXaddRM8R8 emits LOCK XADD r/m8, r8.
func (s *Section) LockXaddRM8R8(dst operand.RM8, src reg.R8) {
	s.lockInst(lockXaddRM8R8, dst, src)
}

// LockXaddRM16R16 emits LOCK XADD r/m16, r16.
func (s *Section) LockXaddRM16R16(dst operand.RM16, src reg.R16) {
	s.lockInst(lockXaddRM16R16, dst, src)
}

// LockXaddRM32R32 emits LOCK XADD r/m32, r32: the sum goes to memory and
// memory's old value goes to the register, which is what makes an atomic
// fetch-and-add one instruction.
func (s *Section) LockXaddRM32R32(dst operand.RM32, src reg.R32) {
	s.lockInst(lockXaddRM32R32, dst, src)
}

// LockXaddRM64R64 emits LOCK XADD r/m64, r64.
func (s *Section) LockXaddRM64R64(dst operand.RM64, src reg.R64) {
	s.lockInst(lockXaddRM64R64, dst, src)
}

// LockCmpxchg16bM128 emits LOCK CMPXCHG16B m128, the sixteen-byte
// compare-and-swap: RDX:RAX against memory, RCX:RBX written on a match.
func (s *Section) LockCmpxchg16bM128(dst operand.M128) {
	s.lockInst(lockCmpxchg16bM128, dst)
}

// ---- the fences -----------------------------------------------------------
//
// The other half of what LOCK is for. LOCK orders one access against
// everything; a fence orders everything before it against everything
// after, and needs no operand to do it.
//
// On this architecture only MFENCE is usually the one wanted. Ordinary
// loads and stores are already ordered against each other — the memory
// model reorders only a store followed by a load — so LFENCE and SFENCE
// are for the non-temporal stores and the streaming loads that opt out
// of that, and not for the code a compiler ordinarily emits.

var (
	lfence = form("Lfence")
	mfence = form("Mfence")
	sfence = form("Sfence")
)

func (s *Section) Lfence() { s.inst(lfence) }
func (s *Section) Mfence() { s.inst(mfence) }
func (s *Section) Sfence() { s.inst(sfence) }
