package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// Scalar SSE and SSE2: the floating-point instruction surface, opposite
// isa/table_sse.go's buildSSE.
//
// None of these is gated. SSE and SSE2 are inside x86-64-v1, which is the
// smallest feature set a module can have, so they are baseline in the same
// sense ADD is.

var (
	movssXmmXM32 = form("MovssXmmXM32")
	movssXM32Xmm = form("MovssXM32Xmm")
	movsdXmmXM64 = form("MovsdXmmXM64")
	movsdXM64Xmm = form("MovsdXM64Xmm")

	movapsXmmRM128 = form("MovapsXmmRM128")
	movapsRM128Xmm = form("MovapsRM128Xmm")
	movapdXmmRM128 = form("MovapdXmmRM128")
	movapdRM128Xmm = form("MovapdRM128Xmm")
	movupsXmmRM128 = form("MovupsXmmRM128")
	movupsRM128Xmm = form("MovupsRM128Xmm")
	movupdXmmRM128 = form("MovupdXmmRM128")
	movupdRM128Xmm = form("MovupdRM128Xmm")

	movdXmmRM32 = form("MovdXmmRM32")
	movdRM32Xmm = form("MovdRM32Xmm")
	movqXmmRM64 = form("MovqXmmRM64")
	movqRM64Xmm = form("MovqRM64Xmm")
)

// MovssXmmXM32 emits MOVSS xmm, xmm/m32. From memory it writes the low four
// bytes and zeroes the rest of the register; from a register it merges,
// leaving the upper three elements of the destination alone. MovapsXmmRM128
// is the whole-register copy.
func (s *Section) MovssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(movssXmmXM32, dst, src) }

// MovssXM32Xmm emits MOVSS xmm/m32, xmm: four bytes, and no more.
func (s *Section) MovssXM32Xmm(dst operand.XM32, src reg.Xmm) { s.inst(movssXM32Xmm, dst, src) }

// MovsdXmmXM64 emits MOVSD xmm, xmm/m64.
func (s *Section) MovsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(movsdXmmXM64, dst, src) }

// MovsdXM64Xmm emits MOVSD xmm/m64, xmm.
func (s *Section) MovsdXM64Xmm(dst operand.XM64, src reg.Xmm) { s.inst(movsdXM64Xmm, dst, src) }

// MovapsXmmRM128 emits MOVAPS xmm, xmm/m128. A memory operand must be
// 16-byte aligned or the instruction faults; MovupsXmmRM128 is the row for
// an address that has not been proven.
func (s *Section) MovapsXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(movapsXmmRM128, dst, src)
}

// MovapsRM128Xmm emits MOVAPS xmm/m128, xmm.
func (s *Section) MovapsRM128Xmm(dst operand.RM128, src reg.Xmm) {
	s.inst(movapsRM128Xmm, dst, src)
}

// MovapdXmmRM128 emits MOVAPD xmm, xmm/m128.
func (s *Section) MovapdXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(movapdXmmRM128, dst, src)
}

// MovapdRM128Xmm emits MOVAPD xmm/m128, xmm.
func (s *Section) MovapdRM128Xmm(dst operand.RM128, src reg.Xmm) {
	s.inst(movapdRM128Xmm, dst, src)
}

// MovupsXmmRM128 emits MOVUPS xmm, xmm/m128.
func (s *Section) MovupsXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(movupsXmmRM128, dst, src)
}

// MovupsRM128Xmm emits MOVUPS xmm/m128, xmm.
func (s *Section) MovupsRM128Xmm(dst operand.RM128, src reg.Xmm) {
	s.inst(movupsRM128Xmm, dst, src)
}

// MovupdXmmRM128 emits MOVUPD xmm, xmm/m128.
func (s *Section) MovupdXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(movupdXmmRM128, dst, src)
}

// MovupdRM128Xmm emits MOVUPD xmm/m128, xmm.
func (s *Section) MovupdRM128Xmm(dst operand.RM128, src reg.Xmm) {
	s.inst(movupdRM128Xmm, dst, src)
}

// MovdXmmRM32 emits MOVD xmm, r/m32: four bytes across the register files,
// zeroing the rest of the destination.
func (s *Section) MovdXmmRM32(dst reg.Xmm, src operand.RM32) { s.inst(movdXmmRM32, dst, src) }

// MovdRM32Xmm emits MOVD r/m32, xmm.
func (s *Section) MovdRM32Xmm(dst operand.RM32, src reg.Xmm) { s.inst(movdRM32Xmm, dst, src) }

// MovqXmmRM64 emits MOVQ xmm, r/m64. With MovdXmmRM32 it is how a float
// constant reaches a vector register without a constant pool: build the bit
// pattern in a general-purpose register and move it across.
func (s *Section) MovqXmmRM64(dst reg.Xmm, src operand.RM64) { s.inst(movqXmmRM64, dst, src) }

// MovqRM64Xmm emits MOVQ r/m64, xmm.
func (s *Section) MovqRM64Xmm(dst operand.RM64, src reg.Xmm) { s.inst(movqRM64Xmm, dst, src) }

// ---- scalar arithmetic ----------------------------------------------------
//
// Two-address: the destination is also the first source. Minss and Maxss
// are not the minimum and the maximum — both return the second operand when
// either is NaN and when both are zero, so neither is commutative and
// neither propagates NaN the way IEEE's minNum does.

var (
	addssXmmXM32 = form("AddssXmmXM32")
	addsdXmmXM64 = form("AddsdXmmXM64")
	subssXmmXM32 = form("SubssXmmXM32")
	subsdXmmXM64 = form("SubsdXmmXM64")
	mulssXmmXM32 = form("MulssXmmXM32")
	mulsdXmmXM64 = form("MulsdXmmXM64")
	divssXmmXM32 = form("DivssXmmXM32")
	divsdXmmXM64 = form("DivsdXmmXM64")
	minssXmmXM32 = form("MinssXmmXM32")
	minsdXmmXM64 = form("MinsdXmmXM64")
	maxssXmmXM32 = form("MaxssXmmXM32")
	maxsdXmmXM64 = form("MaxsdXmmXM64")

	sqrtssXmmXM32 = form("SqrtssXmmXM32")
	sqrtsdXmmXM64 = form("SqrtsdXmmXM64")
)

func (s *Section) AddssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(addssXmmXM32, dst, src) }
func (s *Section) AddsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(addsdXmmXM64, dst, src) }
func (s *Section) SubssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(subssXmmXM32, dst, src) }
func (s *Section) SubsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(subsdXmmXM64, dst, src) }
func (s *Section) MulssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(mulssXmmXM32, dst, src) }
func (s *Section) MulsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(mulsdXmmXM64, dst, src) }
func (s *Section) DivssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(divssXmmXM32, dst, src) }
func (s *Section) DivsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(divsdXmmXM64, dst, src) }
func (s *Section) MinssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(minssXmmXM32, dst, src) }
func (s *Section) MinsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(minsdXmmXM64, dst, src) }
func (s *Section) MaxssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(maxssXmmXM32, dst, src) }
func (s *Section) MaxsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(maxsdXmmXM64, dst, src) }

// SqrtssXmmXM32 emits SQRTSS xmm, xmm/m32. Unlike the rows above, the
// destination is not also a source.
func (s *Section) SqrtssXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(sqrtssXmmXM32, dst, src) }

// SqrtsdXmmXM64 emits SQRTSD xmm, xmm/m64.
func (s *Section) SqrtsdXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(sqrtsdXmmXM64, dst, src) }

// ---- the logical rows -----------------------------------------------------
//
// Packed, because there is no scalar form of them to have instead. They are
// how a sign bit is cleared, set or flipped — fabs, fneg and copysign are a
// mask and one of these — and the operation being packed is invisible when
// the mask has zeroes everywhere but the low element.

var (
	andpsXmmRM128  = form("AndpsXmmRM128")
	andpdXmmRM128  = form("AndpdXmmRM128")
	andnpsXmmRM128 = form("AndnpsXmmRM128")
	andnpdXmmRM128 = form("AndnpdXmmRM128")
	orpsXmmRM128   = form("OrpsXmmRM128")
	orpdXmmRM128   = form("OrpdXmmRM128")
	xorpsXmmRM128  = form("XorpsXmmRM128")
	xorpdXmmRM128  = form("XorpdXmmRM128")
)

func (s *Section) AndpsXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(andpsXmmRM128, dst, src) }
func (s *Section) AndpdXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(andpdXmmRM128, dst, src) }

// AndnpsXmmRM128 emits ANDNPS xmm, xmm/m128, which is (NOT dst) AND src —
// the destination is the operand that gets inverted, which is the opposite
// of what the name reads like.
func (s *Section) AndnpsXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(andnpsXmmRM128, dst, src) }

// AndnpdXmmRM128 emits ANDNPD xmm, xmm/m128.
func (s *Section) AndnpdXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(andnpdXmmRM128, dst, src) }

func (s *Section) OrpsXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(orpsXmmRM128, dst, src) }
func (s *Section) OrpdXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(orpdXmmRM128, dst, src) }

// XorpsXmmRM128 emits XORPS xmm, xmm/m128. With both operands the same
// register it is the shortest zeroing of a vector register, which is what
// the integer XOR is for a general-purpose one.
func (s *Section) XorpsXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(xorpsXmmRM128, dst, src) }

// XorpdXmmRM128 emits XORPD xmm, xmm/m128.
func (s *Section) XorpdXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(xorpdXmmRM128, dst, src) }

// ---- comparison -----------------------------------------------------------
//
// These write EFLAGS rather than a register, which is what lets a float
// comparison feed the same Jcc and SETcc an integer one does. They set ZF,
// PF and CF and clear OF, SF and AF, so the conditions that read them are
// the unsigned ones — above, below, equal — and there is no signed reading
// of a float compare.
//
// PF is set when either operand is NaN, which is the unordered case and the
// reason a float `==` is not one instruction.

var (
	ucomissXmmXM32 = form("UcomissXmmXM32")
	ucomisdXmmXM64 = form("UcomisdXmmXM64")
	comissXmmXM32  = form("ComissXmmXM32")
	comisdXmmXM64  = form("ComisdXmmXM64")
)

// UcomissXmmXM32 emits UCOMISS xmm, xmm/m32, which raises an exception only
// on a signalling NaN. ComissXmmXM32 raises on any NaN.
func (s *Section) UcomissXmmXM32(a reg.Xmm, b operand.XM32) { s.inst(ucomissXmmXM32, a, b) }

// UcomisdXmmXM64 emits UCOMISD xmm, xmm/m64.
func (s *Section) UcomisdXmmXM64(a reg.Xmm, b operand.XM64) { s.inst(ucomisdXmmXM64, a, b) }

// ComissXmmXM32 emits COMISS xmm, xmm/m32.
func (s *Section) ComissXmmXM32(a reg.Xmm, b operand.XM32) { s.inst(comissXmmXM32, a, b) }

// ComisdXmmXM64 emits COMISD xmm, xmm/m64.
func (s *Section) ComisdXmmXM64(a reg.Xmm, b operand.XM64) { s.inst(comisdXmmXM64, a, b) }

// ---- conversion -----------------------------------------------------------
//
// The double T is the difference between truncating toward zero, which is
// what a C cast means, and rounding by MXCSR, which is what lrint means.
// Both are here and spelled out, because one letter is an easy thing to
// misread and the two are different instructions.
//
// There is no unsigned row, and the silicon has none before AVX-512: an
// unsigned conversion is a sequence, and which sequence depends on the
// width.

var (
	cvtsi2ssXmmRM32 = form("Cvtsi2ssXmmRM32")
	cvtsi2ssXmmRM64 = form("Cvtsi2ssXmmRM64")
	cvtsi2sdXmmRM32 = form("Cvtsi2sdXmmRM32")
	cvtsi2sdXmmRM64 = form("Cvtsi2sdXmmRM64")

	cvttss2siR32XM32 = form("Cvttss2siR32XM32")
	cvttss2siR64XM32 = form("Cvttss2siR64XM32")
	cvttsd2siR32XM64 = form("Cvttsd2siR32XM64")
	cvttsd2siR64XM64 = form("Cvttsd2siR64XM64")

	cvtss2siR32XM32 = form("Cvtss2siR32XM32")
	cvtss2siR64XM32 = form("Cvtss2siR64XM32")
	cvtsd2siR32XM64 = form("Cvtsd2siR32XM64")
	cvtsd2siR64XM64 = form("Cvtsd2siR64XM64")

	cvtss2sdXmmXM32 = form("Cvtss2sdXmmXM32")
	cvtsd2ssXmmXM64 = form("Cvtsd2ssXmmXM64")
)

func (s *Section) Cvtsi2ssXmmRM32(dst reg.Xmm, src operand.RM32) { s.inst(cvtsi2ssXmmRM32, dst, src) }
func (s *Section) Cvtsi2ssXmmRM64(dst reg.Xmm, src operand.RM64) { s.inst(cvtsi2ssXmmRM64, dst, src) }
func (s *Section) Cvtsi2sdXmmRM32(dst reg.Xmm, src operand.RM32) { s.inst(cvtsi2sdXmmRM32, dst, src) }
func (s *Section) Cvtsi2sdXmmRM64(dst reg.Xmm, src operand.RM64) { s.inst(cvtsi2sdXmmRM64, dst, src) }

func (s *Section) Cvttss2siR32XM32(dst reg.R32, src operand.XM32) {
	s.inst(cvttss2siR32XM32, dst, src)
}

func (s *Section) Cvttss2siR64XM32(dst reg.R64, src operand.XM32) {
	s.inst(cvttss2siR64XM32, dst, src)
}

func (s *Section) Cvttsd2siR32XM64(dst reg.R32, src operand.XM64) {
	s.inst(cvttsd2siR32XM64, dst, src)
}

func (s *Section) Cvttsd2siR64XM64(dst reg.R64, src operand.XM64) {
	s.inst(cvttsd2siR64XM64, dst, src)
}

func (s *Section) Cvtss2siR32XM32(dst reg.R32, src operand.XM32) { s.inst(cvtss2siR32XM32, dst, src) }
func (s *Section) Cvtss2siR64XM32(dst reg.R64, src operand.XM32) { s.inst(cvtss2siR64XM32, dst, src) }
func (s *Section) Cvtsd2siR32XM64(dst reg.R32, src operand.XM64) { s.inst(cvtsd2siR32XM64, dst, src) }
func (s *Section) Cvtsd2siR64XM64(dst reg.R64, src operand.XM64) { s.inst(cvtsd2siR64XM64, dst, src) }

// Cvtss2sdXmmXM32 emits CVTSS2SD xmm, xmm/m32: a float widened to a double.
func (s *Section) Cvtss2sdXmmXM32(dst reg.Xmm, src operand.XM32) { s.inst(cvtss2sdXmmXM32, dst, src) }

// Cvtsd2ssXmmXM64 emits CVTSD2SS xmm, xmm/m64: a double narrowed to a float.
func (s *Section) Cvtsd2ssXmmXM64(dst reg.Xmm, src operand.XM64) { s.inst(cvtsd2ssXmmXM64, dst, src) }

// ---- rounding -------------------------------------------------------------

var (
	roundssXmmXM32Imm8 = form("RoundssXmmXM32Imm8")
	roundsdXmmXM64Imm8 = form("RoundsdXmmXM64Imm8")
)

func (s *Section) RoundssXmmXM32Imm8(dst reg.Xmm, src operand.XM32, mode int64) {
	s.inst(roundssXmmXM32Imm8, dst, src, imm(mode))
}

func (s *Section) RoundsdXmmXM64Imm8(dst reg.Xmm, src operand.XM64, mode int64) {
	s.inst(roundsdXmmXM64Imm8, dst, src, imm(mode))
}
