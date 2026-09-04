package isa

import "github.com/vertex-language/amd64/feature"

// buildBase declares the x86-64 baseline: everything a module gets with no
// feature set at all, plus the handful of baseline-adjacent rows that carry
// a gate of their own.

func buildBase() {
	buildALU()
	buildShift()
	buildUnary()
	buildMov()
	buildStack()
	buildBranch()
	buildCond()
	buildMisc()
}

// ---- the arithmetic-logical block ----------------------------------------
//
// Eight operations sharing one opcode pattern. For operation index i the
// regular forms are at i*8+0 through i*8+5, and the immediate forms are
// Group 1 — 80, 81 and 83 — with i in ModRM.reg. Writing this as a loop is
// not a shortcut; it is what the block is.

var aluOps = []struct {
	name     string
	i        byte
	lockable bool
}{
	{"add", 0, true},
	{"or", 1, true},
	{"adc", 2, true},
	{"sbb", 3, true},
	{"and", 4, true},
	{"sub", 5, true},
	{"xor", 6, true},
	{"cmp", 7, false}, // CMP writes nothing, so LOCK on it is #UD
}

func buildALU() {
	for _, op := range aluOps {
		n := op.name
		N := title(n)
		base := op.i * 8
		lk := op.lockable

		// r/m, r — the store direction.
		add(Form{Mnemonic: n, Helper: N + "RM8R8", Ops: []Class{RM8, R8},
			Enc: Enc{Op: base + 0, Ext: SlashR, W: WAny, L: LNone}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM16R16", Ops: []Class{RM16, R16},
			Enc: Enc{Op: base + 1, Ext: SlashR, W: WAny, L: LNone, Op16: true}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM32R32", Ops: []Class{RM32, R32},
			Enc: Enc{Op: base + 1, Ext: SlashR, W: WAny, L: LNone}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM64R64", Ops: []Class{RM64, R64},
			Enc: Enc{Op: base + 1, Ext: SlashR, W: 1, L: LNone}, Lockable: lk})

		// r, r/m — the load direction.
		add(Form{Mnemonic: n, Helper: N + "R8RM8", Ops: []Class{R8, RM8},
			Enc: Enc{Op: base + 2, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "R16RM16", Ops: []Class{R16, RM16},
			Enc: Enc{Op: base + 3, Ext: SlashR, W: WAny, L: LNone, Op16: true}})
		add(Form{Mnemonic: n, Helper: N + "R32RM32", Ops: []Class{R32, RM32},
			Enc: Enc{Op: base + 3, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "R64RM64", Ops: []Class{R64, RM64},
			Enc: Enc{Op: base + 3, Ext: SlashR, W: 1, L: LNone}})

		// The accumulator forms. Shorter than the Group 1 equivalents
		// because they carry no ModRM byte, which is the only reason they
		// exist and the only reason Emit ever picks one.
		add(Form{Mnemonic: n, Helper: N + "ALImm8", Ops: []Class{FixAL, Imm8},
			Enc: Enc{Op: base + 4, Ext: NoModRM, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "AXImm16", Ops: []Class{FixAX, Imm16},
			Enc: Enc{Op: base + 5, Ext: NoModRM, W: WAny, L: LNone, Op16: true}})
		add(Form{Mnemonic: n, Helper: N + "EAXImm32", Ops: []Class{FixEAX, Imm32},
			Enc: Enc{Op: base + 5, Ext: NoModRM, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "RAXImm32", Ops: []Class{FixRAX, Imm32},
			Enc: Enc{Op: base + 5, Ext: NoModRM, W: 1, L: LNone}})

		// Group 1, sign-extended imm8 first. Order here only breaks ties;
		// the four-byte 83 form wins over the seven-byte 81 form on length,
		// computed by encoding both.
		//
		// Imm8S and not Imm8: the field is one byte and the operand is two,
		// four or eight, so 200 in this form is -56 in that operand. It has
		// to fail to match here, or shortest-wins would pick it over the
		// wider form that means what the caller wrote.
		ext := int8(op.i)
		add(Form{Mnemonic: n, Helper: N + "RM16Imm8", Ops: []Class{RM16, Imm8S},
			Enc: Enc{Op: 0x83, Ext: ext, W: WAny, L: LNone, Op16: true}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM32Imm8", Ops: []Class{RM32, Imm8S},
			Enc: Enc{Op: 0x83, Ext: ext, W: WAny, L: LNone}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM64Imm8", Ops: []Class{RM64, Imm8S},
			Enc: Enc{Op: 0x83, Ext: ext, W: 1, L: LNone}, Lockable: lk})

		add(Form{Mnemonic: n, Helper: N + "RM8Imm8", Ops: []Class{RM8, Imm8},
			Enc: Enc{Op: 0x80, Ext: ext, W: WAny, L: LNone}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM16Imm16", Ops: []Class{RM16, Imm16},
			Enc: Enc{Op: 0x81, Ext: ext, W: WAny, L: LNone, Op16: true}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM32Imm32", Ops: []Class{RM32, Imm32},
			Enc: Enc{Op: 0x81, Ext: ext, W: WAny, L: LNone}, Lockable: lk})
		add(Form{Mnemonic: n, Helper: N + "RM64Imm32", Ops: []Class{RM64, Imm32},
			Enc: Enc{Op: 0x81, Ext: ext, W: 1, L: LNone}, Lockable: lk})
	}
}

// ---- Group 2, the shifts and rotates --------------------------------------

var shiftOps = []struct {
	name  string
	ext   int8
	alias string
}{
	{"rol", 0, ""},
	{"ror", 1, ""},
	{"rcl", 2, ""},
	{"rcr", 3, ""},
	{"shl", 4, ""},
	{"shr", 5, ""},
	{"sal", 4, "shl"}, // documented second spelling, same bytes
	{"sar", 7, ""},
}

func buildShift() {
	for _, op := range shiftOps {
		n, N, ext, al := op.name, title(op.name), op.ext, op.alias

		// The by-one forms. The literal 1 is in the name, not a parameter:
		// D1 /4 has no immediate field to put another number in.
		add(Form{Mnemonic: n, Helper: N + "RM8One", Ops: []Class{RM8, FixOne}, AliasOf: al,
			Enc: Enc{Op: 0xd0, Ext: ext, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "RM16One", Ops: []Class{RM16, FixOne}, AliasOf: al,
			Enc: Enc{Op: 0xd1, Ext: ext, W: WAny, L: LNone, Op16: true}})
		add(Form{Mnemonic: n, Helper: N + "RM32One", Ops: []Class{RM32, FixOne}, AliasOf: al,
			Enc: Enc{Op: 0xd1, Ext: ext, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "RM64One", Ops: []Class{RM64, FixOne}, AliasOf: al,
			Enc: Enc{Op: 0xd1, Ext: ext, W: 1, L: LNone}})

		// By CL. Same reasoning: the form names the register.
		add(Form{Mnemonic: n, Helper: N + "RM8CL", Ops: []Class{RM8, FixCL}, AliasOf: al,
			Enc: Enc{Op: 0xd2, Ext: ext, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "RM16CL", Ops: []Class{RM16, FixCL}, AliasOf: al,
			Enc: Enc{Op: 0xd3, Ext: ext, W: WAny, L: LNone, Op16: true}})
		add(Form{Mnemonic: n, Helper: N + "RM32CL", Ops: []Class{RM32, FixCL}, AliasOf: al,
			Enc: Enc{Op: 0xd3, Ext: ext, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "RM64CL", Ops: []Class{RM64, FixCL}, AliasOf: al,
			Enc: Enc{Op: 0xd3, Ext: ext, W: 1, L: LNone}})

		// By immediate. The count field is 8 bits and the silicon masks it
		// to 5 or 6; this package does not mask it for you.
		add(Form{Mnemonic: n, Helper: N + "RM8Imm8", Ops: []Class{RM8, Imm8}, AliasOf: al,
			Enc: Enc{Op: 0xc0, Ext: ext, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "RM16Imm8", Ops: []Class{RM16, Imm8}, AliasOf: al,
			Enc: Enc{Op: 0xc1, Ext: ext, W: WAny, L: LNone, Op16: true}})
		add(Form{Mnemonic: n, Helper: N + "RM32Imm8", Ops: []Class{RM32, Imm8}, AliasOf: al,
			Enc: Enc{Op: 0xc1, Ext: ext, W: WAny, L: LNone}})
		add(Form{Mnemonic: n, Helper: N + "RM64Imm8", Ops: []Class{RM64, Imm8}, AliasOf: al,
			Enc: Enc{Op: 0xc1, Ext: ext, W: 1, L: LNone}})
	}
}

// ---- Groups 3, 4 and 5: the unary operations ------------------------------

func buildUnary() {
	// Group 3: F6/F7. TEST is here rather than with the ALU block because
	// its immediate forms are unary-group opcodes, not Group 1 ones.
	unary := []struct {
		name     string
		ext      int8
		lockable bool
	}{
		{"not", 2, true},
		{"neg", 3, true},
		{"mul", 4, false},
		{"imul", 5, false},
		{"div", 6, false},
		{"idiv", 7, false},
	}
	for _, u := range unary {
		n, N := u.name, title(u.name)
		add(Form{Mnemonic: n, Helper: N + "RM8", Ops: []Class{RM8},
			Enc: Enc{Op: 0xf6, Ext: u.ext, W: WAny, L: LNone}, Lockable: u.lockable})
		add(Form{Mnemonic: n, Helper: N + "RM16", Ops: []Class{RM16},
			Enc: Enc{Op: 0xf7, Ext: u.ext, W: WAny, L: LNone, Op16: true}, Lockable: u.lockable})
		add(Form{Mnemonic: n, Helper: N + "RM32", Ops: []Class{RM32},
			Enc: Enc{Op: 0xf7, Ext: u.ext, W: WAny, L: LNone}, Lockable: u.lockable})
		add(Form{Mnemonic: n, Helper: N + "RM64", Ops: []Class{RM64},
			Enc: Enc{Op: 0xf7, Ext: u.ext, W: 1, L: LNone}, Lockable: u.lockable})
	}

	// TEST: the regular forms, the accumulator forms, and Group 3 /0.
	add(Form{Mnemonic: "test", Helper: "TestRM8R8", Ops: []Class{RM8, R8},
		Enc: Enc{Op: 0x84, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestRM16R16", Ops: []Class{RM16, R16},
		Enc: Enc{Op: 0x85, Ext: SlashR, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "test", Helper: "TestRM32R32", Ops: []Class{RM32, R32},
		Enc: Enc{Op: 0x85, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestRM64R64", Ops: []Class{RM64, R64},
		Enc: Enc{Op: 0x85, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestALImm8", Ops: []Class{FixAL, Imm8},
		Enc: Enc{Op: 0xa8, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestEAXImm32", Ops: []Class{FixEAX, Imm32},
		Enc: Enc{Op: 0xa9, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestRAXImm32", Ops: []Class{FixRAX, Imm32},
		Enc: Enc{Op: 0xa9, Ext: NoModRM, W: 1, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestRM8Imm8", Ops: []Class{RM8, Imm8},
		Enc: Enc{Op: 0xf6, Ext: 0, W: WAny, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestRM32Imm32", Ops: []Class{RM32, Imm32},
		Enc: Enc{Op: 0xf7, Ext: 0, W: WAny, L: LNone}})
	add(Form{Mnemonic: "test", Helper: "TestRM64Imm32", Ops: []Class{RM64, Imm32},
		Enc: Enc{Op: 0xf7, Ext: 0, W: 1, L: LNone}})

	// Groups 4 and 5: INC and DEC. The single-byte 40+r forms are gone in
	// long mode — that opcode range is REX now — so these are the only
	// encodings, and there is no shorter one for Emit to find.
	for _, u := range []struct {
		name string
		ext  int8
	}{{"inc", 0}, {"dec", 1}} {
		n, N := u.name, title(u.name)
		add(Form{Mnemonic: n, Helper: N + "RM8", Ops: []Class{RM8},
			Enc: Enc{Op: 0xfe, Ext: u.ext, W: WAny, L: LNone}, Lockable: true})
		add(Form{Mnemonic: n, Helper: N + "RM16", Ops: []Class{RM16},
			Enc: Enc{Op: 0xff, Ext: u.ext, W: WAny, L: LNone, Op16: true}, Lockable: true})
		add(Form{Mnemonic: n, Helper: N + "RM32", Ops: []Class{RM32},
			Enc: Enc{Op: 0xff, Ext: u.ext, W: WAny, L: LNone}, Lockable: true})
		add(Form{Mnemonic: n, Helper: N + "RM64", Ops: []Class{RM64},
			Enc: Enc{Op: 0xff, Ext: u.ext, W: 1, L: LNone}, Lockable: true})
	}

	// The three-operand IMUL forms, which are not unary at all and live
	// here only because they share a mnemonic with the Group 3 one.
	add(Form{Mnemonic: "imul", Helper: "ImulR32RM32", Ops: []Class{R32, RM32},
		Enc: Enc{Map: Map0F, Op: 0xaf, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "imul", Helper: "ImulR64RM64", Ops: []Class{R64, RM64},
		Enc: Enc{Map: Map0F, Op: 0xaf, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "imul", Helper: "ImulR32RM32Imm8", Ops: []Class{R32, RM32, Imm8S},
		Enc: Enc{Op: 0x6b, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "imul", Helper: "ImulR64RM64Imm8", Ops: []Class{R64, RM64, Imm8S},
		Enc: Enc{Op: 0x6b, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "imul", Helper: "ImulR32RM32Imm32", Ops: []Class{R32, RM32, Imm32},
		Enc: Enc{Op: 0x69, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "imul", Helper: "ImulR64RM64Imm32", Ops: []Class{R64, RM64, Imm32},
		Enc: Enc{Op: 0x69, Ext: SlashR, W: 1, L: LNone}})
}

// ---- data movement --------------------------------------------------------

func buildMov() {
	add(Form{Mnemonic: "mov", Helper: "MovRM8R8", Ops: []Class{RM8, R8},
		Enc: Enc{Op: 0x88, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "mov", Helper: "MovRM16R16", Ops: []Class{RM16, R16},
		Enc: Enc{Op: 0x89, Ext: SlashR, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "mov", Helper: "MovRM32R32", Ops: []Class{RM32, R32},
		Enc: Enc{Op: 0x89, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "mov", Helper: "MovRM64R64", Ops: []Class{RM64, R64},
		Enc: Enc{Op: 0x89, Ext: SlashR, W: 1, L: LNone}})

	add(Form{Mnemonic: "mov", Helper: "MovR8RM8", Ops: []Class{R8, RM8},
		Enc: Enc{Op: 0x8a, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "mov", Helper: "MovR16RM16", Ops: []Class{R16, RM16},
		Enc: Enc{Op: 0x8b, Ext: SlashR, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "mov", Helper: "MovR32RM32", Ops: []Class{R32, RM32},
		Enc: Enc{Op: 0x8b, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "mov", Helper: "MovR64RM64", Ops: []Class{R64, RM64},
		Enc: Enc{Op: 0x8b, Ext: SlashR, W: 1, L: LNone}})

	// The +rd immediate forms. MovR32Imm32 is five bytes and zeroes the
	// upper half of the destination; MovRM64Imm32 is seven and
	// sign-extends. They are different instructions with different
	// destination classes, which is why Emit will never substitute one for
	// the other however much shorter it is.
	add(Form{Mnemonic: "mov", Helper: "MovR8Imm8", Ops: []Class{R8, Imm8},
		Enc: Enc{Op: 0xb0, Ext: NoModRM, W: WAny, L: LNone, OpReg: true}})
	add(Form{Mnemonic: "mov", Helper: "MovR16Imm16", Ops: []Class{R16, Imm16},
		Enc: Enc{Op: 0xb8, Ext: NoModRM, W: WAny, L: LNone, OpReg: true, Op16: true}})
	add(Form{Mnemonic: "mov", Helper: "MovR32Imm32", Ops: []Class{R32, Imm32},
		Enc: Enc{Op: 0xb8, Ext: NoModRM, W: WAny, L: LNone, OpReg: true}})
	add(Form{Mnemonic: "mov", Helper: "MovR64Imm64", Ops: []Class{R64, Imm64},
		Enc: Enc{Op: 0xb8, Ext: NoModRM, W: 1, L: LNone, OpReg: true}})

	add(Form{Mnemonic: "mov", Helper: "MovRM8Imm8", Ops: []Class{RM8, Imm8},
		Enc: Enc{Op: 0xc6, Ext: 0, W: WAny, L: LNone}})
	add(Form{Mnemonic: "mov", Helper: "MovRM16Imm16", Ops: []Class{RM16, Imm16},
		Enc: Enc{Op: 0xc7, Ext: 0, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "mov", Helper: "MovRM32Imm32", Ops: []Class{RM32, Imm32},
		Enc: Enc{Op: 0xc7, Ext: 0, W: WAny, L: LNone}})
	add(Form{Mnemonic: "mov", Helper: "MovRM64Imm32", Ops: []Class{RM64, Imm32},
		Enc: Enc{Op: 0xc7, Ext: 0, W: 1, L: LNone}})

	// Width extension. MOVSXD is the one that has no MOVZXD twin, because
	// a 32-bit write already zeroes the upper half and the zero-extending
	// instruction would be MOV.
	add(Form{Mnemonic: "movzx", Helper: "MovzxR32RM8", Ops: []Class{R32, RM8},
		Enc: Enc{Map: Map0F, Op: 0xb6, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movzx", Helper: "MovzxR64RM8", Ops: []Class{R64, RM8},
		Enc: Enc{Map: Map0F, Op: 0xb6, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "movzx", Helper: "MovzxR32RM16", Ops: []Class{R32, RM16},
		Enc: Enc{Map: Map0F, Op: 0xb7, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movzx", Helper: "MovzxR64RM16", Ops: []Class{R64, RM16},
		Enc: Enc{Map: Map0F, Op: 0xb7, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "movsx", Helper: "MovsxR32RM8", Ops: []Class{R32, RM8},
		Enc: Enc{Map: Map0F, Op: 0xbe, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movsx", Helper: "MovsxR64RM8", Ops: []Class{R64, RM8},
		Enc: Enc{Map: Map0F, Op: 0xbe, Ext: SlashR, W: 1, L: LNone}})
	// MOVSX r32, r/m16 — movswl, and the one row of this family that was
	// missing. Its zero-extending twin is here and so are both 64-bit
	// destinations; nothing selected the 32-bit signed one, and an
	// assembler parsing ordinary AT&T reached it immediately.
	add(Form{Mnemonic: "movsx", Helper: "MovsxR32RM16", Ops: []Class{R32, RM16},
		Enc: Enc{Map: Map0F, Op: 0xbf, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movsx", Helper: "MovsxR64RM16", Ops: []Class{R64, RM16},
		Enc: Enc{Map: Map0F, Op: 0xbf, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "movsxd", Helper: "MovsxdR64RM32", Ops: []Class{R64, RM32},
		Enc: Enc{Op: 0x63, Ext: SlashR, W: 1, L: LNone}})

	// LEA takes Memory, not RM64: its operand is an address and has no
	// access width. A register in that slot would not be an instruction.
	add(Form{Mnemonic: "lea", Helper: "LeaR64M", Ops: []Class{R64, M},
		Enc: Enc{Op: 0x8d, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "lea", Helper: "LeaR32M", Ops: []Class{R32, M},
		Enc: Enc{Op: 0x8d, Ext: SlashR, W: WAny, L: LNone}})

	add(Form{Mnemonic: "xchg", Helper: "XchgRM8R8", Ops: []Class{RM8, R8},
		Enc: Enc{Op: 0x86, Ext: SlashR, W: WAny, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "xchg", Helper: "XchgRM16R16", Ops: []Class{RM16, R16},
		Enc: Enc{Op: 0x87, Ext: SlashR, W: WAny, L: LNone, Op16: true}, Lockable: true})
	add(Form{Mnemonic: "xchg", Helper: "XchgRM32R32", Ops: []Class{RM32, R32},
		Enc: Enc{Op: 0x87, Ext: SlashR, W: WAny, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "xchg", Helper: "XchgRM64R64", Ops: []Class{RM64, R64},
		Enc: Enc{Op: 0x87, Ext: SlashR, W: 1, L: LNone}, Lockable: true})

	add(Form{Mnemonic: "movbe", Helper: "MovbeR32RM32", Ops: []Class{R32, RM32},
		Enc:  Enc{Map: Map0F38, Op: 0xf0, Ext: SlashR, W: WAny, L: LNone},
		Gate: []feature.Feature{feature.MOVBE}})
	add(Form{Mnemonic: "movbe", Helper: "MovbeR64RM64", Ops: []Class{R64, RM64},
		Enc:  Enc{Map: Map0F38, Op: 0xf0, Ext: SlashR, W: 1, L: LNone},
		Gate: []feature.Feature{feature.MOVBE}})
}

// ---- the stack ------------------------------------------------------------
//
// PUSH and POP default to a 64-bit operand size in long mode. There is no
// REX.W on any of these and no 32-bit form at all; the width is not a
// choice the encoding offers.

func buildStack() {
	add(Form{Mnemonic: "push", Helper: "PushR64", Ops: []Class{R64},
		Enc: Enc{Op: 0x50, Ext: NoModRM, W: WAny, L: LNone, OpReg: true}})
	add(Form{Mnemonic: "push", Helper: "PushRM64", Ops: []Class{RM64},
		Enc: Enc{Op: 0xff, Ext: 6, W: WAny, L: LNone}})
	add(Form{Mnemonic: "push", Helper: "PushImm8", Ops: []Class{Imm8S},
		Enc: Enc{Op: 0x6a, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "push", Helper: "PushImm32", Ops: []Class{Imm32},
		Enc: Enc{Op: 0x68, Ext: NoModRM, W: WAny, L: LNone}})

	add(Form{Mnemonic: "pop", Helper: "PopR64", Ops: []Class{R64},
		Enc: Enc{Op: 0x58, Ext: NoModRM, W: WAny, L: LNone, OpReg: true}})
	add(Form{Mnemonic: "pop", Helper: "PopRM64", Ops: []Class{RM64},
		Enc: Enc{Op: 0x8f, Ext: 0, W: WAny, L: LNone}})

	add(Form{Mnemonic: "leave", Helper: "Leave", Ops: nil,
		Enc: Enc{Op: 0xc9, Ext: NoModRM, W: WAny, L: LNone}})
}

// ---- branches -------------------------------------------------------------
//
// Two rules live in the naming rather than the table. Short pins rel8 and
// the plain name pins rel32, with no relaxation between them: a short
// branch to a far target is a range failure at Finalize, not a silently
// widened instruction. And Label and Ref split by where they resolve —
// same-section and patched, or crossing and surviving into Refs().

func buildBranch() {
	add(Form{Mnemonic: "jmp", Helper: "JmpShortLabel", Ops: []Class{Rel8},
		Enc: Enc{Op: 0xeb, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "jmp", Helper: "JmpLabel", Ops: []Class{Rel32},
		Enc: Enc{Op: 0xe9, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "jmp", Helper: "JmpRM64", Ops: []Class{RM64},
		Enc: Enc{Op: 0xff, Ext: 4, W: WAny, L: LNone}})

	add(Form{Mnemonic: "call", Helper: "CallLabel", Ops: []Class{Rel32},
		Enc: Enc{Op: 0xe8, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "call", Helper: "CallRM64", Ops: []Class{RM64},
		Enc: Enc{Op: 0xff, Ext: 2, W: WAny, L: LNone}})

	add(Form{Mnemonic: "ret", Helper: "Ret", Ops: nil,
		Enc: Enc{Op: 0xc3, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "ret", Helper: "RetImm16", Ops: []Class{Imm16},
		Enc: Enc{Op: 0xc2, Ext: NoModRM, W: WAny, L: LNone}})

	// Mnemonics with only a rel8 form take the plain name — there is no
	// wider encoding for Short to distinguish them from.
	add(Form{Mnemonic: "loop", Helper: "LoopLabel", Ops: []Class{Rel8},
		Enc: Enc{Op: 0xe2, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "loope", Helper: "LoopeLabel", Ops: []Class{Rel8},
		Enc: Enc{Op: 0xe1, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "loopne", Helper: "LoopneLabel", Ops: []Class{Rel8},
		Enc: Enc{Op: 0xe0, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "jrcxz", Helper: "JrcxzLabel", Ops: []Class{Rel8},
		Enc: Enc{Op: 0xe3, Ext: NoModRM, W: WAny, L: LNone}})
}

// ---- the condition-code family --------------------------------------------
//
// One table, three instruction families. Every documented spelling gets its
// own row with an AliasOf, so a listing says which name the caller used and
// the bytes say what the silicon does: JeLabel and JzLabel are both 0x74.

var conds = []struct {
	name  string
	cc    byte
	alias string
}{
	{"o", 0x0, ""}, {"no", 0x1, ""},
	{"b", 0x2, ""}, {"c", 0x2, "b"}, {"nae", 0x2, "b"},
	{"ae", 0x3, ""}, {"nb", 0x3, "ae"}, {"nc", 0x3, "ae"},
	{"e", 0x4, ""}, {"z", 0x4, "e"},
	{"ne", 0x5, ""}, {"nz", 0x5, "ne"},
	{"be", 0x6, ""}, {"na", 0x6, "be"},
	{"a", 0x7, ""}, {"nbe", 0x7, "a"},
	{"s", 0x8, ""}, {"ns", 0x9, ""},
	{"p", 0xa, ""}, {"pe", 0xa, "p"},
	{"np", 0xb, ""}, {"po", 0xb, "np"},
	{"l", 0xc, ""}, {"nge", 0xc, "l"},
	{"ge", 0xd, ""}, {"nl", 0xd, "ge"},
	{"le", 0xe, ""}, {"ng", 0xe, "le"},
	{"g", 0xf, ""}, {"nle", 0xf, "g"},
}

func buildCond() {
	for _, c := range conds {
		// The condition suffix is lower-case in the helper name too: JneLabel,
		// not JNeLabel. Only the leading word is capitalised.
		j, J := "j"+c.name, "J"+c.name
		aliasJ := ""
		if c.alias != "" {
			aliasJ = "j" + c.alias
		}

		add(Form{Mnemonic: j, Helper: J + "ShortLabel", Ops: []Class{Rel8}, AliasOf: aliasJ,
			Enc: Enc{Op: 0x70 + c.cc, Ext: NoModRM, W: WAny, L: LNone}})
		add(Form{Mnemonic: j, Helper: J + "Label", Ops: []Class{Rel32}, AliasOf: aliasJ,
			Enc: Enc{Map: Map0F, Op: 0x80 + c.cc, Ext: NoModRM, W: WAny, L: LNone}})

		s, S := "set"+c.name, "Set"+c.name
		aliasS := ""
		if c.alias != "" {
			aliasS = "set" + c.alias
		}
		add(Form{Mnemonic: s, Helper: S + "RM8", Ops: []Class{RM8}, AliasOf: aliasS,
			Enc: Enc{Map: Map0F, Op: 0x90 + c.cc, Ext: SlashR, W: WAny, L: LNone}})

		m, M2 := "cmov"+c.name, "Cmov"+c.name
		aliasM := ""
		if c.alias != "" {
			aliasM = "cmov" + c.alias
		}
		add(Form{Mnemonic: m, Helper: M2 + "R32RM32", Ops: []Class{R32, RM32}, AliasOf: aliasM,
			Enc: Enc{Map: Map0F, Op: 0x40 + c.cc, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: m, Helper: M2 + "R64RM64", Ops: []Class{R64, RM64}, AliasOf: aliasM,
			Enc: Enc{Map: Map0F, Op: 0x40 + c.cc, Ext: SlashR, W: 1, L: LNone}})
	}
}

// ---- everything else in the baseline --------------------------------------

func buildMisc() {
	// NOP the instruction. The nop sequences Align pads with are not forms
	// and are not here: they are Nops in encode, because choosing a
	// sequence for a given length is arithmetic over a table of byte
	// strings, not form resolution.
	add(Form{Mnemonic: "nop", Helper: "Nop", Ops: nil,
		Enc: Enc{Op: 0x90, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "nop", Helper: "NopRM32", Ops: []Class{RM32},
		Enc: Enc{Map: Map0F, Op: 0x1f, Ext: 0, W: WAny, L: LNone}})

	add(Form{Mnemonic: "int3", Helper: "Int3", Ops: nil,
		Enc: Enc{Op: 0xcc, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "int", Helper: "IntImm8", Ops: []Class{Imm8},
		Enc: Enc{Op: 0xcd, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "ud2", Helper: "Ud2", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x0b, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "syscall", Helper: "Syscall", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x05, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cpuid", Helper: "Cpuid", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0xa2, Ext: NoModRM, W: WAny, L: LNone}})

	// The privileged and serializing instructions with no operands.
	//
	// These are here because inline assembly reaches them and nothing else
	// does: a kernel's idle loop is HLT, its critical sections are CLI and
	// STI, and its spin loops are PAUSE. Nothing this tree selects emits
	// one, which is why they were absent — a row nothing selects is a row
	// nothing tests, and the assembler is what made them worth having.
	add(Form{Mnemonic: "hlt", Helper: "Hlt", Ops: nil,
		Enc: Enc{Op: 0xf4, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cli", Helper: "Cli", Ops: nil,
		Enc: Enc{Op: 0xfa, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "sti", Helper: "Sti", Ops: nil,
		Enc: Enc{Op: 0xfb, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "pause", Helper: "Pause", Ops: nil,
		Enc: Enc{Pfx: 0xf3, Op: 0x90, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "rdtsc", Helper: "Rdtsc", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x31, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "rdmsr", Helper: "Rdmsr", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x32, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "wrmsr", Helper: "Wrmsr", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x30, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "wbinvd", Helper: "Wbinvd", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x09, Ext: NoModRM, W: WAny, L: LNone}})

	// Sign extension of the accumulator. Three mnemonics, one opcode,
	// separated only by operand size — which is exactly the case where a
	// fixed operand belongs in the name.
	add(Form{Mnemonic: "cwde", Helper: "Cwde", Ops: nil,
		Enc: Enc{Op: 0x98, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cdqe", Helper: "Cdqe", Ops: nil,
		Enc: Enc{Op: 0x98, Ext: NoModRM, W: 1, L: LNone}})
	add(Form{Mnemonic: "cdq", Helper: "Cdq", Ops: nil,
		Enc: Enc{Op: 0x99, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "cqo", Helper: "Cqo", Ops: nil,
		Enc: Enc{Op: 0x99, Ext: NoModRM, W: 1, L: LNone}})

	// Gated baseline-adjacent rows. These are here rather than in a
	// tranche of their own because they are ordinary integer instructions
	// that happen to carry a CPUID bit.
	add(Form{Mnemonic: "popcnt", Helper: "PopcntR32RM32", Ops: []Class{R32, RM32},
		Enc:  Enc{Pfx: 0xf3, Map: Map0F, Op: 0xb8, Ext: SlashR, W: WAny, L: LNone},
		Gate: []feature.Feature{feature.POPCNT}})
	add(Form{Mnemonic: "popcnt", Helper: "PopcntR64RM64", Ops: []Class{R64, RM64},
		Enc:  Enc{Pfx: 0xf3, Map: Map0F, Op: 0xb8, Ext: SlashR, W: 1, L: LNone},
		Gate: []feature.Feature{feature.POPCNT}})
	add(Form{Mnemonic: "lzcnt", Helper: "LzcntR32RM32", Ops: []Class{R32, RM32},
		Enc:  Enc{Pfx: 0xf3, Map: Map0F, Op: 0xbd, Ext: SlashR, W: WAny, L: LNone},
		Gate: []feature.Feature{feature.LZCNT}})
	add(Form{Mnemonic: "lzcnt", Helper: "LzcntR64RM64", Ops: []Class{R64, RM64},
		Enc:  Enc{Pfx: 0xf3, Map: Map0F, Op: 0xbd, Ext: SlashR, W: 1, L: LNone},
		Gate: []feature.Feature{feature.LZCNT}})

	// TZCNT has LZCNT's trap in the other direction: without BMI1 the
	// same bytes decode as BSF, which computes the same answer for every
	// non-zero operand and a different one for zero — BSF leaves the
	// destination alone, TZCNT writes the operand size. So the gate is
	// what stands between a v1 module and code that is right until
	// something counts the zeros in a zero.
	add(Form{Mnemonic: "tzcnt", Helper: "TzcntR32RM32", Ops: []Class{R32, RM32},
		Enc:  Enc{Pfx: 0xf3, Map: Map0F, Op: 0xbc, Ext: SlashR, W: WAny, L: LNone},
		Gate: []feature.Feature{feature.BMI1}})
	add(Form{Mnemonic: "tzcnt", Helper: "TzcntR64RM64", Ops: []Class{R64, RM64},
		Enc:  Enc{Pfx: 0xf3, Map: Map0F, Op: 0xbc, Ext: SlashR, W: 1, L: LNone},
		Gate: []feature.Feature{feature.BMI1}})

	// BSWAP is baseline and takes its register in the opcode, which is
	// why it has no ModRM byte and no memory form. There is a 16-bit
	// encoding and it is not here: Intel documents the result as
	// undefined, and a row that assembles to something undefined is
	// worse than no row.
	add(Form{Mnemonic: "bswap", Helper: "BswapR32", Ops: []Class{R32},
		Enc: Enc{Map: Map0F, Op: 0xc8, Ext: NoModRM, W: WAny, L: LNone, OpReg: true}})
	add(Form{Mnemonic: "bswap", Helper: "BswapR64", Ops: []Class{R64},
		Enc: Enc{Map: Map0F, Op: 0xc8, Ext: NoModRM, W: 1, L: LNone, OpReg: true}})

	add(Form{Mnemonic: "cmpxchg", Helper: "CmpxchgRM8R8", Ops: []Class{RM8, R8},
		Enc: Enc{Map: Map0F, Op: 0xb0, Ext: SlashR, W: WAny, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "cmpxchg", Helper: "CmpxchgRM16R16", Ops: []Class{RM16, R16},
		Enc: Enc{Map: Map0F, Op: 0xb1, Ext: SlashR, W: WAny, L: LNone, Op16: true}, Lockable: true})
	add(Form{Mnemonic: "cmpxchg", Helper: "CmpxchgRM32R32", Ops: []Class{RM32, R32},
		Enc: Enc{Map: Map0F, Op: 0xb1, Ext: SlashR, W: WAny, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "cmpxchg", Helper: "CmpxchgRM64R64", Ops: []Class{RM64, R64},
		Enc: Enc{Map: Map0F, Op: 0xb1, Ext: SlashR, W: 1, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "xadd", Helper: "XaddRM8R8", Ops: []Class{RM8, R8},
		Enc: Enc{Map: Map0F, Op: 0xc0, Ext: SlashR, W: WAny, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "xadd", Helper: "XaddRM16R16", Ops: []Class{RM16, R16},
		Enc: Enc{Map: Map0F, Op: 0xc1, Ext: SlashR, W: WAny, L: LNone, Op16: true}, Lockable: true})
	add(Form{Mnemonic: "xadd", Helper: "XaddRM32R32", Ops: []Class{RM32, R32},
		Enc: Enc{Map: Map0F, Op: 0xc1, Ext: SlashR, W: WAny, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "xadd", Helper: "XaddRM64R64", Ops: []Class{RM64, R64},
		Enc: Enc{Map: Map0F, Op: 0xc1, Ext: SlashR, W: 1, L: LNone}, Lockable: true})
	add(Form{Mnemonic: "cmpxchg16b", Helper: "Cmpxchg16bM128", Ops: []Class{RM128},
		Enc: Enc{Map: Map0F, Op: 0xc7, Ext: 1, W: 1, L: LNone}, Lockable: true,
		Gate: []feature.Feature{feature.CX16}})
}

// title upper-cases the first byte of an ASCII mnemonic fragment. Helper
// names are built from it so the table and inst_*.go cannot disagree about
// capitalisation.
func title(s string) string {
	if s == "" {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
