package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// SSSE3's byte permutes, SHUFPS, and the cryptographic instructions,
// opposite isa/table_crypto.go. All are destructive two-address forms:
// the destination is the first source.

var (
	pshufbXmmRM128      = form("PshufbXmmRM128")
	palignrXmmRM128Imm8 = form("PalignrXmmRM128Imm8")
	shufpsXmmRM128Imm8  = form("ShufpsXmmRM128Imm8")

	aesencXmmRM128              = form("AesencXmmRM128")
	aesenclastXmmRM128          = form("AesenclastXmmRM128")
	aesdecXmmRM128              = form("AesdecXmmRM128")
	aesdeclastXmmRM128          = form("AesdeclastXmmRM128")
	aesimcXmmRM128              = form("AesimcXmmRM128")
	aeskeygenassistXmmRM128Imm8 = form("AeskeygenassistXmmRM128Imm8")
	pclmulqdqXmmRM128Imm8       = form("PclmulqdqXmmRM128Imm8")

	sha1rnds4XmmRM128Imm8   = form("Sha1rnds4XmmRM128Imm8")
	sha1nexteXmmRM128       = form("Sha1nexteXmmRM128")
	sha1msg1XmmRM128        = form("Sha1msg1XmmRM128")
	sha1msg2XmmRM128        = form("Sha1msg2XmmRM128")
	sha256rnds2XmmRM128Xmm0 = form("Sha256rnds2XmmRM128Xmm0")
	sha256msg1XmmRM128      = form("Sha256msg1XmmRM128")
	sha256msg2XmmRM128      = form("Sha256msg2XmmRM128")

	crc32R32RM8  = form("Crc32R32RM8")
	crc32R32RM16 = form("Crc32R32RM16")
	crc32R32RM32 = form("Crc32R32RM32")
	crc32R64RM64 = form("Crc32R64RM64")
)

// ---- permutes -------------------------------------------------------------

// PshufbXmmRM128 emits PSHUFB xmm, xmm/m128 (SSSE3): each byte of dst
// replaced by the byte of dst its index in src names, or zero where the
// index has its top bit set.
func (s *Section) PshufbXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(pshufbXmmRM128, dst, src) }

// PalignrXmmRM128Imm8 emits PALIGNR xmm, xmm/m128, imm8 (SSSE3): dst:src
// shifted right by n bytes, the low sixteen kept.
func (s *Section) PalignrXmmRM128Imm8(dst reg.Xmm, src operand.RM128, n int64) {
	s.inst(palignrXmmRM128Imm8, dst, src, imm(n))
}

// ShufpsXmmRM128Imm8 emits SHUFPS xmm, xmm/m128, imm8: lanes 0 and 1 from
// dst and lanes 2 and 3 from src, each by two bits of order.
func (s *Section) ShufpsXmmRM128Imm8(dst reg.Xmm, src operand.RM128, order int64) {
	s.inst(shufpsXmmRM128Imm8, dst, src, imm(order))
}

// ---- AES and carry-less multiplication ------------------------------------

func (s *Section) AesencXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(aesencXmmRM128, dst, src) }
func (s *Section) AesenclastXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(aesenclastXmmRM128, dst, src)
}
func (s *Section) AesdecXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(aesdecXmmRM128, dst, src) }
func (s *Section) AesdeclastXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(aesdeclastXmmRM128, dst, src)
}

// AesimcXmmRM128 emits AESIMC xmm, xmm/m128, which writes dst without
// reading it.
func (s *Section) AesimcXmmRM128(dst reg.Xmm, src operand.RM128) { s.inst(aesimcXmmRM128, dst, src) }

// AeskeygenassistXmmRM128Imm8 emits AESKEYGENASSIST xmm, xmm/m128, imm8,
// which writes dst without reading it.
func (s *Section) AeskeygenassistXmmRM128Imm8(dst reg.Xmm, src operand.RM128, rcon int64) {
	s.inst(aeskeygenassistXmmRM128Imm8, dst, src, imm(rcon))
}

// PclmulqdqXmmRM128Imm8 emits PCLMULQDQ xmm, xmm/m128, imm8: bit 0 of sel
// picks dst's quadword and bit 4 src's.
func (s *Section) PclmulqdqXmmRM128Imm8(dst reg.Xmm, src operand.RM128, sel int64) {
	s.inst(pclmulqdqXmmRM128Imm8, dst, src, imm(sel))
}

// ---- SHA ------------------------------------------------------------------

// Sha1rnds4XmmRM128Imm8 emits SHA1RNDS4 xmm, xmm/m128, imm8: four rounds
// of the function and constant fn (0 to 3) names.
func (s *Section) Sha1rnds4XmmRM128Imm8(dst reg.Xmm, src operand.RM128, fn int64) {
	s.inst(sha1rnds4XmmRM128Imm8, dst, src, imm(fn))
}

func (s *Section) Sha1nexteXmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(sha1nexteXmmRM128, dst, src)
}
func (s *Section) Sha1msg1XmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(sha1msg1XmmRM128, dst, src)
}
func (s *Section) Sha1msg2XmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(sha1msg2XmmRM128, dst, src)
}
func (s *Section) Sha256msg1XmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(sha256msg1XmmRM128, dst, src)
}
func (s *Section) Sha256msg2XmmRM128(dst reg.Xmm, src operand.RM128) {
	s.inst(sha256msg2XmmRM128, dst, src)
}

// Sha256rnds2XmmRM128Xmm0 emits SHA256RNDS2 xmm, xmm/m128, <XMM0>: two
// rounds, dst the state's CDGH half going in and its ABEF half coming out,
// src the ABEF half, and XMM0's low quadword the two words of W+K.
func (s *Section) Sha256rnds2XmmRM128Xmm0(dst reg.Xmm, src operand.RM128) {
	s.inst(sha256rnds2XmmRM128Xmm0, dst, src, fixXMM0)
}

// ---- CRC-32C --------------------------------------------------------------

func (s *Section) Crc32R32RM8(dst reg.R32, src operand.RM8)   { s.inst(crc32R32RM8, dst, src) }
func (s *Section) Crc32R32RM16(dst reg.R32, src operand.RM16) { s.inst(crc32R32RM16, dst, src) }
func (s *Section) Crc32R32RM32(dst reg.R32, src operand.RM32) { s.inst(crc32R32RM32, dst, src) }
func (s *Section) Crc32R64RM64(dst reg.R64, src operand.RM64) { s.inst(crc32R64RM64, dst, src) }
