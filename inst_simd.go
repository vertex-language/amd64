package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// SSE2's packed integer instructions, opposite isa/table_simd.go's buildSIMD,
// which says what each family is for.
//
// Every one of these is destructive in the same way the two-address integer
// instructions are: the first operand is read and written, so a lowering that
// needs the destination's old value elsewhere copies it first. That is the
// architecture's shape and not this package's — there is no three-operand
// form until AVX, where the same instruction gains one.

var (
	paddbXmmRM128      = form("PaddbXmmRM128")
	paddwXmmRM128      = form("PaddwXmmRM128")
	padddXmmRM128      = form("PadddXmmRM128")
	paddqXmmRM128      = form("PaddqXmmRM128")
	psubbXmmRM128      = form("PsubbXmmRM128")
	psubwXmmRM128      = form("PsubwXmmRM128")
	psubdXmmRM128      = form("PsubdXmmRM128")
	psubqXmmRM128      = form("PsubqXmmRM128")
	paddsbXmmRM128     = form("PaddsbXmmRM128")
	paddswXmmRM128     = form("PaddswXmmRM128")
	paddusbXmmRM128    = form("PaddusbXmmRM128")
	padduswXmmRM128    = form("PadduswXmmRM128")
	psubsbXmmRM128     = form("PsubsbXmmRM128")
	psubswXmmRM128     = form("PsubswXmmRM128")
	psubusbXmmRM128    = form("PsubusbXmmRM128")
	psubuswXmmRM128    = form("PsubuswXmmRM128")
	pmullwXmmRM128     = form("PmullwXmmRM128")
	pmulhwXmmRM128     = form("PmulhwXmmRM128")
	pmulhuwXmmRM128    = form("PmulhuwXmmRM128")
	pmuludqXmmRM128    = form("PmuludqXmmRM128")
	pavgbXmmRM128      = form("PavgbXmmRM128")
	pavgwXmmRM128      = form("PavgwXmmRM128")
	pminubXmmRM128     = form("PminubXmmRM128")
	pmaxubXmmRM128     = form("PmaxubXmmRM128")
	pminswXmmRM128     = form("PminswXmmRM128")
	pmaxswXmmRM128     = form("PmaxswXmmRM128")
	psadbwXmmRM128     = form("PsadbwXmmRM128")
	pmaddwdXmmRM128    = form("PmaddwdXmmRM128")
	pandXmmRM128       = form("PandXmmRM128")
	pandnXmmRM128      = form("PandnXmmRM128")
	porXmmRM128        = form("PorXmmRM128")
	pxorXmmRM128       = form("PxorXmmRM128")
	pcmpeqbXmmRM128    = form("PcmpeqbXmmRM128")
	pcmpeqwXmmRM128    = form("PcmpeqwXmmRM128")
	pcmpeqdXmmRM128    = form("PcmpeqdXmmRM128")
	pcmpgtbXmmRM128    = form("PcmpgtbXmmRM128")
	pcmpgtwXmmRM128    = form("PcmpgtwXmmRM128")
	pcmpgtdXmmRM128    = form("PcmpgtdXmmRM128")
	punpcklbwXmmRM128  = form("PunpcklbwXmmRM128")
	punpcklwdXmmRM128  = form("PunpcklwdXmmRM128")
	punpckldqXmmRM128  = form("PunpckldqXmmRM128")
	punpcklqdqXmmRM128 = form("PunpcklqdqXmmRM128")
	punpckhbwXmmRM128  = form("PunpckhbwXmmRM128")
	punpckhwdXmmRM128  = form("PunpckhwdXmmRM128")
	punpckhdqXmmRM128  = form("PunpckhdqXmmRM128")
	punpckhqdqXmmRM128 = form("PunpckhqdqXmmRM128")
	packsswbXmmRM128   = form("PacksswbXmmRM128")
	packssdwXmmRM128   = form("PackssdwXmmRM128")
	packuswbXmmRM128   = form("PackuswbXmmRM128")
	psllwXmmRM128      = form("PsllwXmmRM128")
	psllwXmmImm8       = form("PsllwXmmImm8")
	pslldXmmRM128      = form("PslldXmmRM128")
	pslldXmmImm8       = form("PslldXmmImm8")
	psllqXmmRM128      = form("PsllqXmmRM128")
	psllqXmmImm8       = form("PsllqXmmImm8")
	psrlwXmmRM128      = form("PsrlwXmmRM128")
	psrlwXmmImm8       = form("PsrlwXmmImm8")
	psrldXmmRM128      = form("PsrldXmmRM128")
	psrldXmmImm8       = form("PsrldXmmImm8")
	psrlqXmmRM128      = form("PsrlqXmmRM128")
	psrlqXmmImm8       = form("PsrlqXmmImm8")
	psrawXmmRM128      = form("PsrawXmmRM128")
	psrawXmmImm8       = form("PsrawXmmImm8")
	psradXmmRM128      = form("PsradXmmRM128")
	psradXmmImm8       = form("PsradXmmImm8")
	pslldqXmmImm8      = form("PslldqXmmImm8")
	psrldqXmmImm8      = form("PsrldqXmmImm8")

	movdqaXmmRM128 = form("MovdqaXmmRM128")
	movdqaRM128Xmm = form("MovdqaRM128Xmm")
	movdquXmmRM128 = form("MovdquXmmRM128")
	movdquRM128Xmm = form("MovdquRM128Xmm")

	pshufdXmmRM128Imm8  = form("PshufdXmmRM128Imm8")
	pshuflwXmmRM128Imm8 = form("PshuflwXmmRM128Imm8")
	pshufhwXmmRM128Imm8 = form("PshufhwXmmRM128Imm8")

	pmovmskbR32Xmm    = form("PmovmskbR32Xmm")
	pinsrwXmmRM32Imm8 = form("PinsrwXmmRM32Imm8")
	pextrwR32XmmImm8  = form("PextrwR32XmmImm8")
)

// ---- whole-register moves -------------------------------------------------

// MovdqaXmmRM128 emits MOVDQA xmm, xmm/m128: sixteen bytes from a
// sixteen-byte-aligned address, or a register copy. An unaligned address
// faults, which is the difference from MOVDQU and the reason a compiler that
// cannot prove alignment reaches for the other one.
func (s *Section) MovdqaXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(movdqaXmmRM128, dst, src) }

// MovdqaRM128Xmm emits MOVDQA xmm/m128, xmm: the store.
func (s *Section) MovdqaRM128Xmm(dst operand.RM128, src reg.Xmm) { s.inst(movdqaRM128Xmm, dst, src) }

// MovdquXmmRM128 emits MOVDQU xmm, xmm/m128, which any alignment satisfies.
func (s *Section) MovdquXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(movdquXmmRM128, dst, src) }

// MovdquRM128Xmm emits MOVDQU xmm/m128, xmm.
func (s *Section) MovdquRM128Xmm(dst operand.RM128, src reg.Xmm) { s.inst(movdquRM128Xmm, dst, src) }

// ---- arithmetic, bitwise, comparison, interleave and pack -----------------
//
// One shape for all of them: the destination is also the first source.

func (s *Section) PaddbXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(paddbXmmRM128, dst, src) }
func (s *Section) PaddwXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(paddwXmmRM128, dst, src) }
func (s *Section) PadddXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(padddXmmRM128, dst, src) }
func (s *Section) PaddqXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(paddqXmmRM128, dst, src) }
func (s *Section) PsubbXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(psubbXmmRM128, dst, src) }
func (s *Section) PsubwXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(psubwXmmRM128, dst, src) }
func (s *Section) PsubdXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(psubdXmmRM128, dst, src) }
func (s *Section) PsubqXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(psubqXmmRM128, dst, src) }
func (s *Section) PaddsbXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(paddsbXmmRM128, dst, src) }
func (s *Section) PaddswXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(paddswXmmRM128, dst, src) }
func (s *Section) PaddusbXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(paddusbXmmRM128, dst, src) }
func (s *Section) PadduswXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(padduswXmmRM128, dst, src) }
func (s *Section) PsubsbXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(psubsbXmmRM128, dst, src) }
func (s *Section) PsubswXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(psubswXmmRM128, dst, src) }
func (s *Section) PsubusbXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(psubusbXmmRM128, dst, src) }
func (s *Section) PsubuswXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(psubuswXmmRM128, dst, src) }
func (s *Section) PmullwXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(pmullwXmmRM128, dst, src) }
func (s *Section) PmulhwXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(pmulhwXmmRM128, dst, src) }
func (s *Section) PmulhuwXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pmulhuwXmmRM128, dst, src) }
func (s *Section) PmuludqXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pmuludqXmmRM128, dst, src) }
func (s *Section) PavgbXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(pavgbXmmRM128, dst, src) }
func (s *Section) PavgwXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(pavgwXmmRM128, dst, src) }
func (s *Section) PminubXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(pminubXmmRM128, dst, src) }
func (s *Section) PmaxubXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(pmaxubXmmRM128, dst, src) }
func (s *Section) PminswXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(pminswXmmRM128, dst, src) }
func (s *Section) PmaxswXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(pmaxswXmmRM128, dst, src) }
func (s *Section) PsadbwXmmRM128(dst reg.Xmm, src operand.RM128)  { s.inst(psadbwXmmRM128, dst, src) }
func (s *Section) PmaddwdXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pmaddwdXmmRM128, dst, src) }
func (s *Section) PandXmmRM128(dst reg.Xmm, src operand.RM128)    { s.inst(pandXmmRM128, dst, src) }
func (s *Section) PandnXmmRM128(dst reg.Xmm, src operand.RM128)   { s.inst(pandnXmmRM128, dst, src) }
func (s *Section) PorXmmRM128(dst reg.Xmm, src operand.RM128)     { s.inst(porXmmRM128, dst, src) }
func (s *Section) PxorXmmRM128(dst reg.Xmm, src operand.RM128)    { s.inst(pxorXmmRM128, dst, src) }
func (s *Section) PcmpeqbXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pcmpeqbXmmRM128, dst, src) }
func (s *Section) PcmpeqwXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pcmpeqwXmmRM128, dst, src) }
func (s *Section) PcmpeqdXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pcmpeqdXmmRM128, dst, src) }
func (s *Section) PcmpgtbXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pcmpgtbXmmRM128, dst, src) }
func (s *Section) PcmpgtwXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pcmpgtwXmmRM128, dst, src) }
func (s *Section) PcmpgtdXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pcmpgtdXmmRM128, dst, src) }
func (s *Section) PunpcklbwXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpcklbwXmmRM128, dst, src)
}
func (s *Section) PunpcklwdXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpcklwdXmmRM128, dst, src)
}
func (s *Section) PunpckldqXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpckldqXmmRM128, dst, src)
}
func (s *Section) PunpcklqdqXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpcklqdqXmmRM128, dst, src)
}
func (s *Section) PunpckhbwXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpckhbwXmmRM128, dst, src)
}
func (s *Section) PunpckhwdXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpckhwdXmmRM128, dst, src)
}
func (s *Section) PunpckhdqXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpckhdqXmmRM128, dst, src)
}
func (s *Section) PunpckhqdqXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(punpckhqdqXmmRM128, dst, src)
}
func (s *Section) PacksswbXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(packsswbXmmRM128, dst, src)
}
func (s *Section) PackssdwXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(packssdwXmmRM128, dst, src)
}
func (s *Section) PackuswbXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(packuswbXmmRM128, dst, src)
}

// ---- shifts ---------------------------------------------------------------
//
// Two forms each: a count in the low quadword of a register, and a count in
// an immediate. A count past the lane's width yields zero rather than
// wrapping, which is the opposite of what the scalar shifts do.

func (s *Section) PsllwXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(psllwXmmRM128, dst, count) }
func (s *Section) PsllwXmmImm8(dst reg.Xmm, count int64)          { s.inst(psllwXmmImm8, dst, imm(count)) }
func (s *Section) PslldXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(pslldXmmRM128, dst, count) }
func (s *Section) PslldXmmImm8(dst reg.Xmm, count int64)          { s.inst(pslldXmmImm8, dst, imm(count)) }
func (s *Section) PsllqXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(psllqXmmRM128, dst, count) }
func (s *Section) PsllqXmmImm8(dst reg.Xmm, count int64)          { s.inst(psllqXmmImm8, dst, imm(count)) }
func (s *Section) PsrlwXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(psrlwXmmRM128, dst, count) }
func (s *Section) PsrlwXmmImm8(dst reg.Xmm, count int64)          { s.inst(psrlwXmmImm8, dst, imm(count)) }
func (s *Section) PsrldXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(psrldXmmRM128, dst, count) }
func (s *Section) PsrldXmmImm8(dst reg.Xmm, count int64)          { s.inst(psrldXmmImm8, dst, imm(count)) }
func (s *Section) PsrlqXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(psrlqXmmRM128, dst, count) }
func (s *Section) PsrlqXmmImm8(dst reg.Xmm, count int64)          { s.inst(psrlqXmmImm8, dst, imm(count)) }
func (s *Section) PsrawXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(psrawXmmRM128, dst, count) }
func (s *Section) PsrawXmmImm8(dst reg.Xmm, count int64)          { s.inst(psrawXmmImm8, dst, imm(count)) }
func (s *Section) PsradXmmRM128(dst reg.Xmm, count operand.RM128) { s.inst(psradXmmRM128, dst, count) }
func (s *Section) PsradXmmImm8(dst reg.Xmm, count int64)          { s.inst(psradXmmImm8, dst, imm(count)) }

// PslldqXmmImm8 emits PSLLDQ xmm, imm8: the whole register left by a count in
// bytes, zeroes shifted in. There is no by-register form.
func (s *Section) PslldqXmmImm8(dst reg.Xmm, bytes int64) { s.inst(pslldqXmmImm8, dst, imm(bytes)) }

// PsrldqXmmImm8 emits PSRLDQ xmm, imm8: the same to the right.
func (s *Section) PsrldqXmmImm8(dst reg.Xmm, bytes int64) { s.inst(psrldqXmmImm8, dst, imm(bytes)) }

// ---- shuffles, masks and lane access --------------------------------------

// PshufdXmmRM128Imm8 emits PSHUFD xmm, xmm/m128, imm8: four doublewords
// permuted by four two-bit fields, low field selecting the low result lane.
// Unlike the instructions above it reads its source rather than its
// destination, so it is the one that can permute without a copy first.
func (s *Section) PshufdXmmRM128Imm8(dst reg.Xmm, src operand.RM128, order int64) {
	s.inst(pshufdXmmRM128Imm8, dst, src, imm(order))
}

// PshuflwXmmRM128Imm8 emits PSHUFLW: the low four words permuted, the high
// four copied through.
func (s *Section) PshuflwXmmRM128Imm8(dst reg.Xmm, src operand.RM128, order int64) {
	s.inst(pshuflwXmmRM128Imm8, dst, src, imm(order))
}

// PshufhwXmmRM128Imm8 emits PSHUFHW: the high four words permuted, the low
// four copied through.
func (s *Section) PshufhwXmmRM128Imm8(dst reg.Xmm, src operand.RM128, order int64) {
	s.inst(pshufhwXmmRM128Imm8, dst, src, imm(order))
}

// PmovmskbR32Xmm emits PMOVMSKB r32, xmm: the top bit of each of sixteen
// bytes, gathered into the low sixteen bits of a general register. It is how
// a packed compare's mask becomes something a branch can read.
func (s *Section) PmovmskbR32Xmm(dst reg.R32, src reg.Xmm) { s.inst(pmovmskbR32Xmm, dst, src) }

// PinsrwXmmRM32Imm8 emits PINSRW xmm, r32/m16, imm8: one word placed at the
// index the immediate names, the other seven left alone.
func (s *Section) PinsrwXmmRM32Imm8(dst reg.Xmm, src operand.RM32, index int64) {
	s.inst(pinsrwXmmRM32Imm8, dst, src, imm(index))
}

// PextrwR32XmmImm8 emits PEXTRW r32, xmm, imm8: one word out, zero-extended
// into the general register.
func (s *Section) PextrwR32XmmImm8(dst reg.R32, src reg.Xmm, index int64) {
	s.inst(pextrwR32XmmImm8, dst, src, imm(index))
}
