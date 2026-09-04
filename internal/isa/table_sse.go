package isa

// buildSSE declares scalar SSE and SSE2: the floating-point instructions a
// compiler needs to hold an f32 or an f64 in a register and do arithmetic
// on it.
//
// Nothing here is gated. SSE and SSE2 are both inside x86-64-v1, which is
// what feature.Default returns and the smallest set a module can have —
// the 64-bit ABI made them mandatory, and the x87 stack they replaced is
// not what a modern psABI passes a double in. So these rows are baseline
// in the same sense ADD is, and a Gate on them could never fail.
//
// What is here is scalar. ADDSS adds the low element of one register to
// the low element of another and leaves the upper three alone, which is
// what a `float` addition is; the packed forms that add all four are a
// different tranche and belong with the rest of the vector table.
//
// The four logical rows are the exception and they are packed, because
// there is no scalar form of them to have instead. ANDPS and XORPS are how
// a sign bit is cleared, set or flipped — fabs, fneg and copysign are a
// mask and one of these — and the operation being packed is invisible when
// the mask has zeroes everywhere but the low element.
//
// A note on the two prefix bytes, since they look interchangeable and are
// not. F3 makes an operation single-precision scalar and F2 makes it
// double; 66 makes it packed double, and no prefix at all makes it packed
// single. Those are four different instructions sharing one opcode, which
// is why the table spells the prefix on every row rather than deriving it.

func buildSSE() {
	buildSSERound()
	buildSSEMove()
	buildSSEArith()
	buildSSELogic()
	buildSSECompare()
	buildSSEConvert()
}

// ---- data movement --------------------------------------------------------

func buildSSEMove() {
	// The scalar loads and stores. Each moves exactly its own width: MOVSS
	// from memory writes the low four bytes and zeroes the rest of the
	// register, and MOVSS to memory writes four bytes and no more.
	//
	// Register to register is the one case where they differ from that
	// description — MOVSS xmm1, xmm2 merges rather than zeroing — which is
	// why a lowering that wants a whole-register copy reaches for MOVAPS
	// below and not for these.
	add(Form{Mnemonic: "movss", Helper: "MovssXmmXM32", Ops: []Class{Xmm, XM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x10, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movss", Helper: "MovssXM32Xmm", Ops: []Class{XM32, Xmm},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x11, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movsd", Helper: "MovsdXmmXM64", Ops: []Class{Xmm, XM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x10, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movsd", Helper: "MovsdXM64Xmm", Ops: []Class{XM64, Xmm},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x11, Ext: SlashR, W: WAny, L: LNone}})

	// The whole-register moves, aligned and unaligned. MOVAPS is the
	// register-to-register copy every scalar lowering actually emits: one
	// byte shorter than MOVSS's register form, and with no false
	// dependency on the destination's upper bits, because it writes all of
	// them.
	//
	// Aligned means a memory operand must be 16-byte aligned or the
	// instruction faults. That is a fact about the address and not about
	// the form, so it is the caller's to know; MOVUPS is the row for an
	// address that has not been proven.
	for _, m := range []struct {
		name string
		pfx  byte
	}{{"movaps", 0}, {"movapd", 0x66}} {
		N := title(m.name)
		add(Form{Mnemonic: m.name, Helper: N + "XmmRM128", Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: m.pfx, Map: Map0F, Op: 0x28, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: m.name, Helper: N + "RM128Xmm", Ops: []Class{RM128, Xmm},
			Enc: Enc{Pfx: m.pfx, Map: Map0F, Op: 0x29, Ext: SlashR, W: WAny, L: LNone}})
	}
	for _, m := range []struct {
		name string
		pfx  byte
	}{{"movups", 0}, {"movupd", 0x66}} {
		N := title(m.name)
		add(Form{Mnemonic: m.name, Helper: N + "XmmRM128", Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: m.pfx, Map: Map0F, Op: 0x10, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: m.name, Helper: N + "RM128Xmm", Ops: []Class{RM128, Xmm},
			Enc: Enc{Pfx: m.pfx, Map: Map0F, Op: 0x11, Ext: SlashR, W: WAny, L: LNone}})
	}

	// Between the two register files. This is how a float constant reaches
	// a vector register without a constant pool: build the bit pattern in
	// a general-purpose register and move it across.
	//
	// MOVD and MOVQ are one opcode pair distinguished by REX.W, which is
	// the same thing that distinguishes them in their own names.
	add(Form{Mnemonic: "movd", Helper: "MovdXmmRM32", Ops: []Class{Xmm, RM32},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x6e, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movd", Helper: "MovdRM32Xmm", Ops: []Class{RM32, Xmm},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x7e, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movq", Helper: "MovqXmmRM64", Ops: []Class{Xmm, RM64},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x6e, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "movq", Helper: "MovqRM64Xmm", Ops: []Class{RM64, Xmm},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x7e, Ext: SlashR, W: 1, L: LNone}})
}

// ---- scalar arithmetic ----------------------------------------------------

// sseArith is the block of scalar operations that share an opcode between
// single and double precision and differ only in the prefix byte.
//
// MIN and MAX are in it and are not the minimum and the maximum. Both
// return the second operand when either is NaN and when both are zero,
// which makes them neither commutative nor NaN-propagating — IEEE's
// minNum and maxNum are a longer sequence, and a lowering that needs
// those cannot reach for these.
var sseArith = []struct {
	name string
	op   byte
}{
	{"add", 0x58},
	{"mul", 0x59},
	{"sub", 0x5c},
	{"min", 0x5d},
	{"div", 0x5e},
	{"max", 0x5f},
}

func buildSSEArith() {
	for _, a := range sseArith {
		add(Form{Mnemonic: a.name + "ss", Helper: title(a.name) + "ssXmmXM32",
			Ops: []Class{Xmm, XM32},
			Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: a.op, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: a.name + "sd", Helper: title(a.name) + "sdXmmXM64",
			Ops: []Class{Xmm, XM64},
			Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: a.op, Ext: SlashR, W: WAny, L: LNone}})
	}

	// SQRT is the same shape and not in the loop above, because it is a
	// unary operation whose destination is not also a source. Two operands
	// either way, which is why it looks identical here and why the
	// difference matters to whoever emits it and not to the encoder.
	add(Form{Mnemonic: "sqrtss", Helper: "SqrtssXmmXM32", Ops: []Class{Xmm, XM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x51, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "sqrtsd", Helper: "SqrtsdXmmXM64", Ops: []Class{Xmm, XM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x51, Ext: SlashR, W: WAny, L: LNone}})
}

// ---- the logical rows -----------------------------------------------------

var sseLogic = []struct {
	name string
	op   byte
}{
	{"and", 0x54},
	{"andn", 0x55},
	{"or", 0x56},
	{"xor", 0x57},
}

func buildSSELogic() {
	for _, l := range sseLogic {
		add(Form{Mnemonic: l.name + "ps", Helper: title(l.name) + "psXmmRM128",
			Ops: []Class{Xmm, RM128},
			Enc: Enc{Map: Map0F, Op: l.op, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: l.name + "pd", Helper: title(l.name) + "pdXmmRM128",
			Ops: []Class{Xmm, RM128},
			Enc: Enc{Pfx: 0x66, Map: Map0F, Op: l.op, Ext: SlashR, W: WAny, L: LNone}})
	}
}

// ---- comparison -----------------------------------------------------------

// The scalar compares write EFLAGS rather than a register, which is what
// lets a float comparison feed the same Jcc and SETcc an integer one does.
//
// They set ZF, PF and CF and clear OF, SF and AF, so the conditions that
// read them are the unsigned ones: above, below, equal. There is no signed
// reading of a float compare and no JG that means anything after one.
//
// UCOMIS raises an exception only on a signalling NaN; COMIS raises on any
// NaN. A language whose `<` on NaN is false without trapping wants UCOMIS,
// and PF is how it finds out that is what happened.
func buildSSECompare() {
	add(Form{Mnemonic: "ucomiss", Helper: "UcomissXmmXM32", Ops: []Class{Xmm, XM32},
		Enc: Enc{Map: Map0F, Op: 0x2e, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "ucomisd", Helper: "UcomisdXmmXM64", Ops: []Class{Xmm, XM64},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x2e, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "comiss", Helper: "ComissXmmXM32", Ops: []Class{Xmm, XM32},
		Enc: Enc{Map: Map0F, Op: 0x2f, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "comisd", Helper: "ComisdXmmXM64", Ops: []Class{Xmm, XM64},
		Enc: Enc{Pfx: 0x66, Map: Map0F, Op: 0x2f, Ext: SlashR, W: WAny, L: LNone}})
}

// ---- conversion -----------------------------------------------------------

// Every conversion between the two register files, and the two between the
// float widths.
//
// CVTSI2SS and CVTSI2SD read a general-purpose register and REX.W says
// whether it is 32 or 64 bits of signed integer. There is no unsigned row
// and the silicon has none before AVX-512: an unsigned conversion is a
// sequence, and which sequence depends on the width.
//
// The truncating and the rounding forms are different instructions and not
// a rounding mode. CVTTSD2SI truncates toward zero, which is what a C cast
// means; CVTSD2SI rounds by MXCSR, which is what a `lrint` means. The
// double T is the whole difference in the name and it is easy to misread,
// which is why both rows are here and spelled out.
func buildSSEConvert() {
	// Integer to float.
	add(Form{Mnemonic: "cvtsi2ss", Helper: "Cvtsi2ssXmmRM32", Ops: []Class{Xmm, RM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x2a, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cvtsi2ss", Helper: "Cvtsi2ssXmmRM64", Ops: []Class{Xmm, RM64},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x2a, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "cvtsi2sd", Helper: "Cvtsi2sdXmmRM32", Ops: []Class{Xmm, RM32},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x2a, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cvtsi2sd", Helper: "Cvtsi2sdXmmRM64", Ops: []Class{Xmm, RM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x2a, Ext: SlashR, W: 1, L: LNone}})

	// Float to integer, truncating.
	add(Form{Mnemonic: "cvttss2si", Helper: "Cvttss2siR32XM32", Ops: []Class{R32, XM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x2c, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cvttss2si", Helper: "Cvttss2siR64XM32", Ops: []Class{R64, XM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x2c, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "cvttsd2si", Helper: "Cvttsd2siR32XM64", Ops: []Class{R32, XM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x2c, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cvttsd2si", Helper: "Cvttsd2siR64XM64", Ops: []Class{R64, XM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x2c, Ext: SlashR, W: 1, L: LNone}})

	// Float to integer, rounding by MXCSR.
	add(Form{Mnemonic: "cvtss2si", Helper: "Cvtss2siR32XM32", Ops: []Class{R32, XM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x2d, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cvtss2si", Helper: "Cvtss2siR64XM32", Ops: []Class{R64, XM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x2d, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "cvtsd2si", Helper: "Cvtsd2siR32XM64", Ops: []Class{R32, XM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x2d, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cvtsd2si", Helper: "Cvtsd2siR64XM64", Ops: []Class{R64, XM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x2d, Ext: SlashR, W: 1, L: LNone}})

	// Between the float widths.
	add(Form{Mnemonic: "cvtss2sd", Helper: "Cvtss2sdXmmXM32", Ops: []Class{Xmm, XM32},
		Enc: Enc{Pfx: 0xf3, Map: Map0F, Op: 0x5a, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cvtsd2ss", Helper: "Cvtsd2ssXmmXM64", Ops: []Class{Xmm, XM64},
		Enc: Enc{Pfx: 0xf2, Map: Map0F, Op: 0x5a, Ext: SlashR, W: WAny, L: LNone}})
}

// ---- SSE4.1 Rounding ------------------------------------------------------
func init() {
	// Wait, we can't use init() here if we want it controlled by forms.go.
}

// ---- rounding -------------------------------------------------------------

func buildSSERound() {
	add(Form{Mnemonic: "roundss", Helper: "RoundssXmmXM32Imm8", Ops: []Class{Xmm, XM32, Imm8},
		Enc: Enc{Pfx: 0x66, Map: Map0F3A, Op: 0x0a, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "roundsd", Helper: "RoundsdXmmXM64Imm8", Ops: []Class{Xmm, XM64, Imm8},
		Enc: Enc{Pfx: 0x66, Map: Map0F3A, Op: 0x0b, Ext: SlashR, W: WAny, L: LNone}})
}
