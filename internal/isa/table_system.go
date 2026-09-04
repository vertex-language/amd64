package isa

// The bit-test, double-shift, port and system instructions.
//
// A tranche of its own because of what it has in common: nothing in this
// tree selects any of it. An instruction selector emits arithmetic, loads,
// branches and calls; it does not emit HLT, and it does not emit BT either,
// because a bit test is an AND with a mask by the time an IR has finished
// with it. These are the instructions a C header writes by hand, and the
// assembler is the door they come through.
//
// That is also why they were absent. A row nothing selects is a row nothing
// tests — and the way this list was found was by pointing the assembler at
// the inline assembly of a libc and a kernel and reading what it refused.
func buildSystem() {
	// ---- Bit test and set -------------------------------------------------
	//
	// Two encodings each: a bit number in a register, and a bit number as
	// an immediate under one shared opcode with the operation in /digit.
	for _, op := range []struct {
		mnem string
		N    string
		reg  byte // the 0F xx register-form opcode
		ext  int8 // the /digit of the 0F BA immediate form
		lock bool
	}{
		{"bt", "Bt", 0xa3, 4, false},
		{"bts", "Bts", 0xab, 5, true},
		{"btr", "Btr", 0xb3, 6, true},
		{"btc", "Btc", 0xbb, 7, true},
	} {
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM32R32", Ops: []Class{RM32, R32},
			Enc: Enc{Map: Map0F, Op: op.reg, Ext: SlashR, W: WAny, L: LNone}, Lockable: op.lock})
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM64R64", Ops: []Class{RM64, R64},
			Enc: Enc{Map: Map0F, Op: op.reg, Ext: SlashR, W: 1, L: LNone}, Lockable: op.lock})
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM32Imm8", Ops: []Class{RM32, Imm8},
			Enc: Enc{Map: Map0F, Op: 0xba, Ext: op.ext, W: WAny, L: LNone}, Lockable: op.lock})
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM64Imm8", Ops: []Class{RM64, Imm8},
			Enc: Enc{Map: Map0F, Op: 0xba, Ext: op.ext, W: 1, L: LNone}, Lockable: op.lock})
	}

	// ---- Bit scan ---------------------------------------------------------
	add(Form{Mnemonic: "bsf", Helper: "BsfR32RM32", Ops: []Class{R32, RM32},
		Enc: Enc{Map: Map0F, Op: 0xbc, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "bsf", Helper: "BsfR64RM64", Ops: []Class{R64, RM64},
		Enc: Enc{Map: Map0F, Op: 0xbc, Ext: SlashR, W: 1, L: LNone}})
	add(Form{Mnemonic: "bsr", Helper: "BsrR32RM32", Ops: []Class{R32, RM32},
		Enc: Enc{Map: Map0F, Op: 0xbd, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "bsr", Helper: "BsrR64RM64", Ops: []Class{R64, RM64},
		Enc: Enc{Map: Map0F, Op: 0xbd, Ext: SlashR, W: 1, L: LNone}})

	// ---- Double-precision shift -------------------------------------------
	//
	// Three operands, and the count is either an immediate or CL. The count
	// is a byte and not a sign-extended one: it is masked to 5 or 6 bits by
	// the silicon and means nothing as a negative number.
	for _, op := range []struct {
		mnem, N string
		imm, cl byte
	}{
		{"shld", "Shld", 0xa4, 0xa5},
		{"shrd", "Shrd", 0xac, 0xad},
	} {
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM32R32Imm8", Ops: []Class{RM32, R32, Imm8},
			Enc: Enc{Map: Map0F, Op: op.imm, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM64R64Imm8", Ops: []Class{RM64, R64, Imm8},
			Enc: Enc{Map: Map0F, Op: op.imm, Ext: SlashR, W: 1, L: LNone}})
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM32R32CL", Ops: []Class{RM32, R32, FixCL},
			Enc: Enc{Map: Map0F, Op: op.cl, Ext: SlashR, W: WAny, L: LNone}})
		add(Form{Mnemonic: op.mnem, Helper: op.N + "RM64R64CL", Ops: []Class{RM64, R64, FixCL},
			Enc: Enc{Map: Map0F, Op: op.cl, Ext: SlashR, W: 1, L: LNone}})
	}

	// ---- The flags on the stack -------------------------------------------
	//
	// Spelled with the q: in 64-bit mode the operand size is fixed at eight
	// bytes and PUSHFQ is the only name the instruction has.
	add(Form{Mnemonic: "pushfq", Helper: "Pushfq", Ops: nil,
		Enc: Enc{Op: 0x9c, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "popfq", Helper: "Popfq", Ops: nil,
		Enc: Enc{Op: 0x9d, Ext: NoModRM, W: WAny, L: LNone}})

	// ---- Port I/O ---------------------------------------------------------
	//
	// The port is either an immediate byte or DX, and the data register is
	// fixed by the width. Both halves are named in the form, which is what
	// makes these eight rows rather than two with an operand each.
	add(Form{Mnemonic: "in", Helper: "InALImm8", Ops: []Class{FixAL, Imm8},
		Enc: Enc{Op: 0xe4, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "in", Helper: "InAXImm8", Ops: []Class{FixAX, Imm8},
		Enc: Enc{Op: 0xe5, Ext: NoModRM, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "in", Helper: "InEAXImm8", Ops: []Class{FixEAX, Imm8},
		Enc: Enc{Op: 0xe5, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "in", Helper: "InALDX", Ops: []Class{FixAL, FixDX},
		Enc: Enc{Op: 0xec, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "in", Helper: "InAXDX", Ops: []Class{FixAX, FixDX},
		Enc: Enc{Op: 0xed, Ext: NoModRM, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "in", Helper: "InEAXDX", Ops: []Class{FixEAX, FixDX},
		Enc: Enc{Op: 0xed, Ext: NoModRM, W: WAny, L: LNone}})

	add(Form{Mnemonic: "out", Helper: "OutImm8AL", Ops: []Class{Imm8, FixAL},
		Enc: Enc{Op: 0xe6, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "out", Helper: "OutImm8AX", Ops: []Class{Imm8, FixAX},
		Enc: Enc{Op: 0xe7, Ext: NoModRM, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "out", Helper: "OutImm8EAX", Ops: []Class{Imm8, FixEAX},
		Enc: Enc{Op: 0xe7, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "out", Helper: "OutDXAL", Ops: []Class{FixDX, FixAL},
		Enc: Enc{Op: 0xee, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "out", Helper: "OutDXAX", Ops: []Class{FixDX, FixAX},
		Enc: Enc{Op: 0xef, Ext: NoModRM, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "out", Helper: "OutDXEAX", Ops: []Class{FixDX, FixEAX},
		Enc: Enc{Op: 0xef, Ext: NoModRM, W: WAny, L: LNone}})

	// ---- Direction flag ---------------------------------------------------
	add(Form{Mnemonic: "cld", Helper: "Cld", Ops: nil,
		Enc: Enc{Op: 0xfc, Ext: NoModRM, W: WAny, L: LNone}})
	add(Form{Mnemonic: "std", Helper: "Std", Ops: nil,
		Enc: Enc{Op: 0xfd, Ext: NoModRM, W: WAny, L: LNone}})

	// ---- Privileged, no operands ------------------------------------------
	add(Form{Mnemonic: "swapgs", Helper: "Swapgs", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xf8}})
	add(Form{Mnemonic: "rdtscp", Helper: "Rdtscp", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xf9}})
	add(Form{Mnemonic: "clac", Helper: "Clac", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xca}})
	add(Form{Mnemonic: "stac", Helper: "Stac", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xcb}})
	add(Form{Mnemonic: "xgetbv", Helper: "Xgetbv", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xd0}})
	add(Form{Mnemonic: "xsetbv", Helper: "Xsetbv", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xd1}})

	// IRETQ and SYSRETQ carry REX.W, which is the whole of what the q says:
	// the same opcodes without it return to 32-bit code.
	add(Form{Mnemonic: "iretq", Helper: "Iretq", Ops: nil,
		Enc: Enc{Op: 0xcf, Ext: NoModRM, W: 1, L: LNone}})
	add(Form{Mnemonic: "sysretq", Helper: "Sysretq", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0x07, Ext: NoModRM, W: 1, L: LNone}})

	// ---- The task and descriptor-table registers --------------------------
	//
	// One opcode, four /digits, and a 16-bit operand throughout. LTR and
	// LLDT read exactly sixteen bits, so no operand-size prefix is needed
	// or emitted; STR and SLDT write a register and take the prefix that
	// makes the write sixteen bits wide rather than thirty-two.
	add(Form{Mnemonic: "sldt", Helper: "SldtRM16", Ops: []Class{RM16},
		Enc: Enc{Map: Map0F, Op: 0x00, Ext: 0, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "str", Helper: "StrRM16", Ops: []Class{RM16},
		Enc: Enc{Map: Map0F, Op: 0x00, Ext: 1, W: WAny, L: LNone, Op16: true}})
	add(Form{Mnemonic: "lldt", Helper: "LldtRM16", Ops: []Class{RM16},
		Enc: Enc{Map: Map0F, Op: 0x00, Ext: 2, W: WAny, L: LNone}})
	add(Form{Mnemonic: "ltr", Helper: "LtrRM16", Ops: []Class{RM16},
		Enc: Enc{Map: Map0F, Op: 0x00, Ext: 3, W: WAny, L: LNone}})

	// ---- Descriptor tables and the TLB ------------------------------------
	//
	// The operand is an address and has no access width: LGDT reads six or
	// ten bytes depending on the mode, which is not something an operand
	// says. Class M is that, and it is the reason MemBytes answers zero.
	add(Form{Mnemonic: "sgdt", Helper: "SgdtM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: 0, W: WAny, L: LNone}})
	add(Form{Mnemonic: "sidt", Helper: "SidtM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: 1, W: WAny, L: LNone}})
	add(Form{Mnemonic: "lgdt", Helper: "LgdtM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: 2, W: WAny, L: LNone}})
	add(Form{Mnemonic: "lidt", Helper: "LidtM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: 3, W: WAny, L: LNone}})
	add(Form{Mnemonic: "invlpg", Helper: "InvlpgM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x01, Ext: 7, W: WAny, L: LNone}})

	// ---- Cache and state management ---------------------------------------
	add(Form{Mnemonic: "clflush", Helper: "ClflushM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: 7, W: WAny, L: LNone}})
	add(Form{Mnemonic: "fxsave", Helper: "FxsaveM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: 0, W: WAny, L: LNone}})
	add(Form{Mnemonic: "fxrstor", Helper: "FxrstorM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: 1, W: WAny, L: LNone}})
	add(Form{Mnemonic: "xsave", Helper: "XsaveM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: 4, W: WAny, L: LNone}})
	add(Form{Mnemonic: "xrstor", Helper: "XrstorM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: 5, W: WAny, L: LNone}})

	// The four prefetch hints, which differ only in /digit.
	add(Form{Mnemonic: "prefetchnta", Helper: "PrefetchntaM", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x18, Ext: 0, W: WAny, L: LNone}})
	add(Form{Mnemonic: "prefetcht0", Helper: "Prefetcht0M", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x18, Ext: 1, W: WAny, L: LNone}})
	add(Form{Mnemonic: "prefetcht1", Helper: "Prefetcht1M", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x18, Ext: 2, W: WAny, L: LNone}})
	add(Form{Mnemonic: "prefetcht2", Helper: "Prefetcht2M", Ops: []Class{M},
		Enc: Enc{Map: Map0F, Op: 0x18, Ext: 3, W: WAny, L: LNone}})

	// A non-temporal store, whose destination is memory and only memory:
	// the point of it is to miss the cache, and a register has none.
	add(Form{Mnemonic: "movnti", Helper: "MovntiMR32", Ops: []Class{M, R32},
		Enc: Enc{Map: Map0F, Op: 0xc3, Ext: SlashR, W: WAny, L: LNone}})
	add(Form{Mnemonic: "movnti", Helper: "MovntiMR64", Ops: []Class{M, R64},
		Enc: Enc{Map: Map0F, Op: 0xc3, Ext: SlashR, W: 1, L: LNone}})
}
