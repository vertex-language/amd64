package isa

// buildSIMD declares SSE2's packed integer instructions: the ones that treat
// an XMM register as a vector of bytes, words, doublewords or quadwords
// rather than as one number.
//
// table_sse.go's note says the packed forms "belong with the rest of the
// vector table", and this is that table. The split is by what an operation
// means rather than by which extension introduced it: everything here reads
// every lane, everything there reads the low one.
//
// Nothing is gated, for table_sse.go's reason. SSE2 is inside x86-64-v1, so
// a module that can be built at all has these.
//
// Two encoding facts run through the file. Every packed integer form carries
// the 66 prefix — the same opcodes without it are the MMX instructions on the
// old register file, which are different instructions with a different
// destination. And the shift-by-immediate forms put their digit in ModRM.reg
// and the register being shifted in ModRM.rm, which is why they take a Slash
// rather than SlashR.
func buildSIMD() {
	buildSIMDMove()
	buildSIMDArith()
	buildSIMDLogic()
	buildSIMDCompare()
	buildSIMDShift()
	buildSIMDPack()
	buildSIMDMisc()
}

// ---- whole-register moves -------------------------------------------------

// MOVDQA and MOVDQU are the integer spellings of MOVAPS and MOVUPS. They move
// the same sixteen bytes; what differs is the processor's idea of what the
// bytes are for, which on some parts costs a forwarding penalty between an
// integer operation and a float one and never costs correctness.
func buildSIMDMove() {
	add(Form{Mnemonic: "movdqa", Helper: "MovdqaXmmRM128", Ops: []Class{Xmm, RM128},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x6f, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movdqa", Helper: "MovdqaRM128Xmm", Ops: []Class{RM128, Xmm},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x7f, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movdqu", Helper: "MovdquXmmRM128", Ops: []Class{Xmm, RM128},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x6f, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movdqu", Helper: "MovdquRM128Xmm", Ops: []Class{RM128, Xmm},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x7f, Ext: SlashR, W: WAny, L: LNone}})
}

// ---- arithmetic -----------------------------------------------------------

// simdArith is every packed add, subtract and multiply that is one opcode
// with one operand shape.
//
// The wrapping adds and subtracts come in all four lane widths; the
// saturating ones only in the two narrow widths, because saturation is a
// property of a narrow lane and a doubleword that overflows is one that
// wrapped.
var simdArith = []struct {
	name string
	op   byte
}{
	{"paddb", 0xfc}, {"paddw", 0xfd}, {"paddd", 0xfe}, {"paddq", 0xd4},
	{"psubb", 0xf8}, {"psubw", 0xf9}, {"psubd", 0xfa}, {"psubq", 0xfb},

	// Saturating: a result past the lane's range becomes the range's end
	// rather than wrapping into it, which is what makes these the right
	// instructions for pixels.
	{"paddsb", 0xec}, {"paddsw", 0xed},
	{"paddusb", 0xdc}, {"paddusw", 0xdd},
	{"psubsb", 0xe8}, {"psubsw", 0xe9},
	{"psubusb", 0xd8}, {"psubusw", 0xd9},

	// The multiplies keep half a product each: sixteen bits times sixteen
	// is thirty-two, and one instruction writes the low half and another
	// the high. PMULUDQ is the exception that keeps all of it, multiplying
	// two of the four doublewords into two quadwords.
	{"pmullw", 0xd5}, {"pmulhw", 0xe5}, {"pmulhuw", 0xe4}, {"pmuludq", 0xf4},

	// Averages, minima and maxima. The widths each comes in are the ones
	// SSE2 has rather than a pattern: unsigned bytes and signed words,
	// which is what the video codecs of the day needed.
	{"pavgb", 0xe0}, {"pavgw", 0xe3},
	{"pminub", 0xda}, {"pmaxub", 0xde},
	{"pminsw", 0xea}, {"pmaxsw", 0xee},

	// PSADBW sums the absolute differences of sixteen bytes into two
	// sixteen-bit totals, one per half of the register. PMADDWD multiplies
	// eight signed words pairwise and adds adjacent products into four
	// doublewords.
	{"psadbw", 0xf6}, {"pmaddwd", 0xf5},
}

func buildSIMDArith() {
	for _, a := range simdArith {
		add(Form{Mnemonic: a.name, Helper: title(a.name) + "XmmRM128",
			Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: 0x66, Map: Map0F, Op: a.op, Ext: SlashR, W: WAny, L: LNone}})
	}
}

// ---- bitwise --------------------------------------------------------------

// The integer spellings of ANDPS and its three companions, and the same four
// operations. PANDN is the one with an order to remember: it is (NOT dst) AND
// src, so the register inverted is the destination.
var simdLogic = []struct {
	name string
	op   byte
}{
	{"pand", 0xdb}, {"pandn", 0xdf}, {"por", 0xeb}, {"pxor", 0xef},
}

func buildSIMDLogic() {
	for _, l := range simdLogic {
		add(Form{Mnemonic: l.name, Helper: title(l.name) + "XmmRM128",
			Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: 0x66, Map: Map0F, Op: l.op, Ext: SlashR, W: WAny, L: LNone}})
	}
}

// ---- comparison -----------------------------------------------------------

// A packed compare writes a mask rather than flags: every lane becomes all
// ones or all zeroes, which is what makes the result an operand to the
// bitwise instructions above. That is the whole idiom — compare, mask,
// combine — and it is why there is no packed compare that sets EFLAGS.
//
// Equality and signed greater-than are all SSE2 has. Less-than is
// greater-than with the operands the other way round, and the unsigned
// comparisons are built from the minima and maxima and an equality test.
var simdCompare = []struct {
	name string
	op   byte
}{
	{"pcmpeqb", 0x74}, {"pcmpeqw", 0x75}, {"pcmpeqd", 0x76},
	{"pcmpgtb", 0x64}, {"pcmpgtw", 0x65}, {"pcmpgtd", 0x66},
}

func buildSIMDCompare() {
	for _, c := range simdCompare {
		add(Form{Mnemonic: c.name, Helper: title(c.name) + "XmmRM128",
			Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: 0x66, Map: Map0F, Op: c.op, Ext: SlashR, W: WAny, L: LNone}})
	}
}

// ---- shifts ---------------------------------------------------------------

// Each shift comes in two forms: by a count in an immediate, and by a count
// in the low quadword of another register. Both shift every lane by the same
// amount — a per-lane count is AVX2's variable shift and not here.
//
// A count past the lane's width produces zero rather than wrapping, which is
// the opposite of what the scalar shifts do and is the useful answer.
var simdShift = []struct {
	name  string
	byReg byte // the shift-by-xmm opcode
	imm   byte // the group opcode of the shift-by-imm8 form
	digit int8 // and its digit
}{
	{"psllw", 0xf1, 0x71, 6}, {"pslld", 0xf2, 0x72, 6}, {"psllq", 0xf3, 0x73, 6},
	{"psrlw", 0xd1, 0x71, 2}, {"psrld", 0xd2, 0x72, 2}, {"psrlq", 0xd3, 0x73, 2},

	// The arithmetic right shifts smear the sign bit. SSE2 has no PSRAQ —
	// a signed quadword shift arrived with AVX-512 — which is why this row
	// is shorter than the two above it.
	{"psraw", 0xe1, 0x71, 4}, {"psrad", 0xe2, 0x72, 4},
}

func buildSIMDShift() {
	for _, s := range simdShift {
		add(Form{Mnemonic: s.name, Helper: title(s.name) + "XmmRM128",
			Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: 0x66, Map: Map0F, Op: s.byReg, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: s.name, Helper: title(s.name) + "XmmImm8",
			Ops: []Class{Xmm, Imm8},
			Enc: Enc{Pfx: 0x66, Map: Map0F, Op: s.imm, Ext: s.digit, W: WAny, L: LNone}})
	}

	// The two byte-wise shifts of the whole register, which have no
	// by-register form at all: the count is in bytes and is always
	// immediate. They are how a lane reaches another position without a
	// shuffle.
	add(Form{Mnemonic: "pslldq", Helper: "PslldqXmmImm8", Ops: []Class{Xmm, Imm8},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x73, Ext: 7, W: WAny, L: LNone}})
	add(Form{Mnemonic: "psrldq", Helper: "PsrldqXmmImm8", Ops: []Class{Xmm, Imm8},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x73, Ext: 3, W: WAny, L: LNone}})
}

// ---- interleave and pack --------------------------------------------------

// Unpacking interleaves the lanes of two registers, taking them from one
// half: PUNPCKLBW makes eight words out of the low eight bytes of each
// operand, alternating. Widening a vector is unpacking against zero, and
// unpacking against itself is how a lane is duplicated.
//
// Packing is the other direction and saturates: two registers of words
// become one of bytes, and a word outside the byte's range becomes the
// range's end. PACKUSWB saturates to unsigned bytes, which is what a pixel
// pipeline finishes with.
var simdPack = []struct {
	name string
	op   byte
}{
	{"punpcklbw", 0x60}, {"punpcklwd", 0x61}, {"punpckldq", 0x62}, {"punpcklqdq", 0x6c},
	{"punpckhbw", 0x68}, {"punpckhwd", 0x69}, {"punpckhdq", 0x6a}, {"punpckhqdq", 0x6d},
	{"packsswb", 0x63}, {"packssdw", 0x6b}, {"packuswb", 0x67},
}

func buildSIMDPack() {
	for _, p := range simdPack {
		add(Form{Mnemonic: p.name, Helper: title(p.name) + "XmmRM128",
			Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: 0x66, Map: Map0F, Op: p.op, Ext: SlashR, W: WAny, L: LNone}})
	}
}

// ---- shuffles, masks and lane access --------------------------------------

func buildSIMDMisc() {
	// The three shuffles share opcode 0F 70 and differ only in the prefix,
	// which is the trick the scalar table's four ADDs play. PSHUFD permutes
	// four doublewords by an immediate of four two-bit fields; PSHUFLW and
	// PSHUFHW permute the four words of one half and copy the other half
	// through unchanged.
	add(Form{Mnemonic: "pshufd", Helper: "PshufdXmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x70, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "pshuflw", Helper: "PshuflwXmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x70, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "pshufhw", Helper: "PshufhwXmmRM128Imm8", Ops: []Class{Xmm, RM128, Imm8},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x70, Ext: SlashR, W: WAny, L: LNone}})

	// PMOVMSKB gathers the top bit of each of sixteen bytes into the low
	// sixteen bits of a general register. It is how a packed compare's mask
	// becomes something a branch can read.
	add(Form{Mnemonic: "pmovmskb", Helper: "PmovmskbR32Xmm", Ops: []Class{R32, Xmm},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0xd7, Ext: SlashR, W: WAny, L: LNone}})

	// One word in and one word out, by index. SSE2 has these two and no
	// byte or doubleword counterpart; those arrived with SSE4.1.
	add(Form{Mnemonic: "pinsrw", Helper: "PinsrwXmmRM32Imm8", Ops: []Class{Xmm, RM32, Imm8},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0xc4, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "pextrw", Helper: "PextrwR32XmmImm8", Ops: []Class{R32, Xmm, Imm8},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0xc5, Ext: SlashR, W: WAny, L: LNone}})
}
