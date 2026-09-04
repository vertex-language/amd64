package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// Data movement: MOV and its width-extending relatives, LEA, XCHG, MOVBE.

var (
	movRM8R8   = form("MovRM8R8")
	movRM16R16 = form("MovRM16R16")
	movRM32R32 = form("MovRM32R32")
	movRM64R64 = form("MovRM64R64")

	movR8RM8   = form("MovR8RM8")
	movR16RM16 = form("MovR16RM16")
	movR32RM32 = form("MovR32RM32")
	movR64RM64 = form("MovR64RM64")

	movR8Imm8    = form("MovR8Imm8")
	movR16Imm16  = form("MovR16Imm16")
	movR32Imm32  = form("MovR32Imm32")
	movR64Imm64  = form("MovR64Imm64")
	movRM8Imm8   = form("MovRM8Imm8")
	movRM16Imm16 = form("MovRM16Imm16")
	movRM32Imm32 = form("MovRM32Imm32")
	movRM64Imm32 = form("MovRM64Imm32")
)

func (s *Section) MovRM8R8(dst operand.RM8, src reg.R8)     { s.inst(movRM8R8, dst, src) }
func (s *Section) MovRM16R16(dst operand.RM16, src reg.R16) { s.inst(movRM16R16, dst, src) }
func (s *Section) MovRM32R32(dst operand.RM32, src reg.R32) { s.inst(movRM32R32, dst, src) }
func (s *Section) MovRM64R64(dst operand.RM64, src reg.R64) { s.inst(movRM64R64, dst, src) }

func (s *Section) MovR8RM8(dst reg.R8, src operand.RM8)     { s.inst(movR8RM8, dst, src) }
func (s *Section) MovR16RM16(dst reg.R16, src operand.RM16) { s.inst(movR16RM16, dst, src) }
func (s *Section) MovR32RM32(dst reg.R32, src operand.RM32) { s.inst(movR32RM32, dst, src) }
func (s *Section) MovR64RM64(dst reg.R64, src operand.RM64) { s.inst(movR64RM64, dst, src) }

// The immediate forms, and the three that a lowering has to choose between
// deliberately.
//
// MovR32Imm32 is five bytes and zeroes the upper half of the destination.
// MovRM64Imm32 is seven and sign-extends. MovR64Imm64 is ten and is the only
// imm64 in the architecture — movabs. They are different instructions with
// different destination classes, which is why Emit never substitutes one for
// another however much shorter it is: picking the five-byte version means
// knowing that a 32-bit write zeroes the upper half, and that is a claim
// about the architecture, so it is stated here rather than inferred there.

func (s *Section) MovR8Imm8(dst reg.R8, v int64)    { s.inst(movR8Imm8, dst, imm(v)) }
func (s *Section) MovR16Imm16(dst reg.R16, v int64) { s.inst(movR16Imm16, dst, imm(v)) }
func (s *Section) MovR32Imm32(dst reg.R32, v int64) { s.inst(movR32Imm32, dst, imm(v)) }

// MovR64Imm64 takes a uint64 because what goes here is usually a bit pattern
// rather than a number, and 0xdeadbeefcafef00d is an awkward int64 literal to
// write. It is the one helper in the tree whose parameter is unsigned.
func (s *Section) MovR64Imm64(dst reg.R64, v uint64) { s.inst(movR64Imm64, dst, immU(v)) }

func (s *Section) MovRM8Imm8(dst operand.RM8, v int64)    { s.inst(movRM8Imm8, dst, imm(v)) }
func (s *Section) MovRM16Imm16(dst operand.RM16, v int64) { s.inst(movRM16Imm16, dst, imm(v)) }
func (s *Section) MovRM32Imm32(dst operand.RM32, v int64) { s.inst(movRM32Imm32, dst, imm(v)) }
func (s *Section) MovRM64Imm32(dst operand.RM64, v int64) { s.inst(movRM64Imm32, dst, imm(v)) }

// ---- width extension ------------------------------------------------------
//
// Seven helpers and not nine. There is no MovzxR64RM32, because a write to a
// 32-bit register already zeroes the upper half and the zero-extending
// instruction is therefore MOV — MovR32RM32 with a 64-bit destination in mind.
// There is no MovzxR16RM8 either; the table declares the 32-bit destination
// and a 16-bit one would need the 66 prefix to say something nobody wants.
//
// The RM8 forms are where the high-byte rule bites hardest. MovzxR64RM8(RAX,
// AH) has no encoding — the 64-bit destination forces REX, and REX turns
// encoding 4 into SPL — and it is ErrOperand at the call naming both
// registers, because the kinds were right and the combination is what the
// silicon cannot express.

var (
	movzxR32RM8  = form("MovzxR32RM8")
	movzxR64RM8  = form("MovzxR64RM8")
	movzxR32RM16 = form("MovzxR32RM16")
	movzxR64RM16 = form("MovzxR64RM16")

	movsxR32RM8  = form("MovsxR32RM8")
	movsxR64RM8  = form("MovsxR64RM8")
	movsxR32RM16 = form("MovsxR32RM16")
	movsxR64RM16 = form("MovsxR64RM16")

	movsxdR64RM32 = form("MovsxdR64RM32")
)

func (s *Section) MovzxR32RM8(dst reg.R32, src operand.RM8)   { s.inst(movzxR32RM8, dst, src) }
func (s *Section) MovzxR64RM8(dst reg.R64, src operand.RM8)   { s.inst(movzxR64RM8, dst, src) }
func (s *Section) MovzxR32RM16(dst reg.R32, src operand.RM16) { s.inst(movzxR32RM16, dst, src) }
func (s *Section) MovzxR64RM16(dst reg.R64, src operand.RM16) { s.inst(movzxR64RM16, dst, src) }

func (s *Section) MovsxR32RM8(dst reg.R32, src operand.RM8)   { s.inst(movsxR32RM8, dst, src) }
func (s *Section) MovsxR64RM8(dst reg.R64, src operand.RM8)   { s.inst(movsxR64RM8, dst, src) }
func (s *Section) MovsxR32RM16(dst reg.R32, src operand.RM16) { s.inst(movsxR32RM16, dst, src) }
func (s *Section) MovsxR64RM16(dst reg.R64, src operand.RM16) { s.inst(movsxR64RM16, dst, src) }

// MovsxdR64RM32 is the one MOVSXD this package declares. The manual lists
// 16- and 32-bit destinations too and then says not to use them, because
// without REX.W the instruction is a MOV with extra bytes.
func (s *Section) MovsxdR64RM32(dst reg.R64, src operand.RM32) {
	s.inst(movsxdR64RM32, dst, src)
}

// ---- LEA ------------------------------------------------------------------
//
// LEA takes Memory rather than an RM class, because its operand is an address
// and has no access width. A register in that slot would not be an
// instruction. It is also how a RIP-relative symbol address is loaded:
//
//	t.LeaR64M(amd64.RDI, amd64.Rip(amd64.Ref("msg", amd64.RefPC32)))
//
// which is four bytes and position-independent, and needs no thunk, no GOT
// and no decision.

var (
	leaR64M = form("LeaR64M")
	leaR32M = form("LeaR32M")
)

func (s *Section) LeaR64M(dst reg.R64, src operand.Memory) { s.inst(leaR64M, dst, src) }
func (s *Section) LeaR32M(dst reg.R32, src operand.Memory) { s.inst(leaR32M, dst, src) }

// ---- XCHG -----------------------------------------------------------------
//
// XCHG with a memory destination is implicitly locked whether or not the LOCK
// prefix is written — the only instruction in the architecture that is — so
// the locking clones in inst_lock.go exist for symmetry and cost a byte.

var (
	xchgRM8R8   = form("XchgRM8R8")
	xchgRM16R16 = form("XchgRM16R16")
	xchgRM32R32 = form("XchgRM32R32")
	xchgRM64R64 = form("XchgRM64R64")
)

func (s *Section) XchgRM8R8(dst operand.RM8, src reg.R8)     { s.inst(xchgRM8R8, dst, src) }
func (s *Section) XchgRM16R16(dst operand.RM16, src reg.R16) { s.inst(xchgRM16R16, dst, src) }
func (s *Section) XchgRM32R32(dst operand.RM32, src reg.R32) { s.inst(xchgRM32R32, dst, src) }
func (s *Section) XchgRM64R64(dst operand.RM64, src reg.R64) { s.inst(xchgRM64R64, dst, src) }

// ---- MOVBE ----------------------------------------------------------------
//
// Byte-swapping load and store, gated on its own CPUID bit even though the
// v3 level requires it. A module at v1 that calls one of these gets
// ErrFeature naming MOVBE, which is the answer whether the caller thinks in
// levels or in extensions.

var (
	movbeR32RM32 = form("MovbeR32RM32")
	movbeR64RM64 = form("MovbeR64RM64")
)

func (s *Section) MovbeR32RM32(dst reg.R32, src operand.RM32) { s.inst(movbeR32RM32, dst, src) }
func (s *Section) MovbeR64RM64(dst reg.R64, src operand.RM64) { s.inst(movbeR64RM64, dst, src) }
