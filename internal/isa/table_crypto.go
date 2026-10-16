package isa

import "github.com/vertex-language/amd64/feature"

// buildSIMDExt declares the packed forms past SSE2 that a lowering of the
// byte permutes and the SHA state reaches for: SSSE3's PSHUFB and PALIGNR,
// and SSE's SHUFPS, which is a float instruction by its name and the only
// one that takes two doublewords from each of two registers.
//
// PSHUFB and PALIGNR are gated: SSSE3 is x86-64-v2, not the baseline.
// SHUFPS is SSE and in every AMD64 processor.
func buildSIMDExt() {
	ssse3 := []feature.Feature{feature.SSSE3}

	// Byte i of the destination is byte src[i]&15 of the destination, or
	// zero where src[i] has its top bit set.
	add(Form{Mnemonic: "pshufb", Helper: "PshufbXmmRM128", Ops: []Class{Xmm, RM128},
		Enc:  Enc{Pfx: 0x66, Map: Map0F38, Op: 0x00, Ext: SlashR, W: WAny, L: LNone},
		Gate: ssse3})
	// The 32 bytes dst:src shifted right by imm8 bytes, the low sixteen kept.
	add(Form{Mnemonic: "palignr", Helper: "PalignrXmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc:  Enc{Pfx: 0x66, Map: Map0F3A, Op: 0x0f, Ext: SlashR, W: WAny, L: LNone},
		Gate: ssse3})

	// The low two doublewords from the destination and the high two from
	// the source, each picked by two bits of imm8.
	add(Form{Mnemonic: "shufps", Helper: "ShufpsXmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc: Enc{Map: Map0F, Op: 0xc6, Ext: SlashR, W: WAny, L: LNone}})
}

// buildCrypto declares AES-NI, PCLMULQDQ, the SHA extensions and SSE4.2's
// CRC32: the instructions the IR's §V5 and §A8 verbs are on x86-64.
//
// Each is gated on its own CPUID bit, and none is in a level: AES and
// PCLMULQDQ are in nearly every processor since 2010 and SHA in the ones
// since 2017 (2019 on Intel's cores), but a level is a promise about every
// processor that claims it, and these were never part of one. CRC32 is
// SSE4.2's and so x86-64-v2.
//
// The SHA forms have no mandatory prefix, unlike every other form here;
// 66 0F 38 CC is not SHA256MSG1 but nothing at all.
func buildCrypto() {
	aes := []feature.Feature{feature.AES}
	for _, r := range []struct {
		n, h string
		op   byte
	}{
		{"aesenc", "AesencXmmRM128", 0xdc},
		{"aesenclast", "AesenclastXmmRM128", 0xdd},
		{"aesdec", "AesdecXmmRM128", 0xde},
		{"aesdeclast", "AesdeclastXmmRM128", 0xdf},
		{"aesimc", "AesimcXmmRM128", 0xdb},
	} {
		add(Form{Mnemonic: r.n, Helper: r.h, Ops: []Class{Xmm, RM128},
			Enc:  Enc{Pfx: 0x66, Map: Map0F38, Op: r.op, Ext: SlashR, W: WAny, L: LNone},
			Gate: aes})
	}
	add(Form{Mnemonic: "aeskeygenassist", Helper: "AeskeygenassistXmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc:  Enc{Pfx: 0x66, Map: Map0F3A, Op: 0xdf, Ext: SlashR, W: WAny, L: LNone},
		Gate: aes})

	// Bit 0 of imm8 picks the destination's quadword, bit 4 the source's.
	add(Form{Mnemonic: "pclmulqdq", Helper: "PclmulqdqXmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc:  Enc{Pfx: 0x66, Map: Map0F3A, Op: 0x44, Ext: SlashR, W: WAny, L: LNone},
		Gate: []feature.Feature{feature.PCLMULQDQ}})

	sha := []feature.Feature{feature.SHA}
	add(Form{Mnemonic: "sha1rnds4", Helper: "Sha1rnds4XmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc: Enc{Map: Map0F3A, Op: 0xcc, Ext: SlashR, W: WAny, L: LNone}, Gate: sha})
	for _, r := range []struct {
		n, h string
		op   byte
	}{
		{"sha1nexte", "Sha1nexteXmmRM128", 0xc8},
		{"sha1msg1", "Sha1msg1XmmRM128", 0xc9},
		{"sha1msg2", "Sha1msg2XmmRM128", 0xca},
		{"sha256msg1", "Sha256msg1XmmRM128", 0xcc},
		{"sha256msg2", "Sha256msg2XmmRM128", 0xcd},
	} {
		add(Form{Mnemonic: r.n, Helper: r.h, Ops: []Class{Xmm, RM128},
			Enc: Enc{Map: Map0F38, Op: r.op, Ext: SlashR, W: WAny, L: LNone}, Gate: sha})
	}
	// Two rounds; the third operand is the two words of W+K in XMM0's low
	// quadword, which the legacy encoding has no field for.
	add(Form{Mnemonic: "sha256rnds2", Helper: "Sha256rnds2XmmRM128Xmm0", Ops: []Class{Xmm, RM128, FixXMM0},
		Enc: Enc{Map: Map0F38, Op: 0xcb, Ext: SlashR, W: WAny, L: LNone}, Gate: sha})

	// CRC32 folds a byte, word, doubleword or quadword into the CRC-32C
	// (Castagnoli) remainder in the destination. The word form carries
	// both the operand-size prefix and the mandatory F2, in that order.
	sse42 := []feature.Feature{feature.SSE42}
	add(Form{Mnemonic: "crc32", Helper: "Crc32R32RM8", Ops: []Class{R32, RM8},
		Enc: Enc{Pfx: 0xf2, Map: Map0F38, Op: 0xf0, Ext: SlashR, W: WAny, L: LNone}, Gate: sse42})
	add(Form{Mnemonic: "crc32", Helper: "Crc32R32RM16", Ops: []Class{R32, RM16},
		Enc: Enc{Pfx: 0xf2, Map: Map0F38, Op: 0xf1, Ext: SlashR, W: WAny, L: LNone, Op16: true}, Gate: sse42})
	add(Form{Mnemonic: "crc32", Helper: "Crc32R32RM32", Ops: []Class{R32, RM32},
		Enc: Enc{Pfx: 0xf2, Map: Map0F38, Op: 0xf1, Ext: SlashR, W: WAny, L: LNone}, Gate: sse42})
	add(Form{Mnemonic: "crc32", Helper: "Crc32R64RM64", Ops: []Class{R64, RM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F38, Op: 0xf1, Ext: SlashR, W: 1, L: LNone}, Gate: sse42})
}
