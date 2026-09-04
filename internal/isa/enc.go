package isa

// Enc is how a form becomes bytes, minus everything that is arithmetic.
//
// A form names an opcode, an operand shape, and which encoding family it
// belongs to. Whether a legacy form gets a REX byte, and whether a VEX form
// gets the two-byte C5 or the three-byte C4, is computed from the operands
// at encode time, because it is a function of the register numbers and
// nothing a table row could know. Neither is a form and neither is visible
// here, because there is no case where a caller would want the longer one.
type Enc struct {
	Family Family
	Pfx    byte // mandatory prefix: 0, 0x66, 0xF2 or 0xF3
	Map    Map
	Op     byte // primary opcode

	// Ext is the /digit that goes in ModRM.reg, or SlashR when the reg
	// field holds an operand, or NoModRM when there is no ModRM byte.
	Ext int8

	// W is REX.W for legacy forms and VEX.W/EVEX.W otherwise. -1 means the
	// bit is not part of this form's identity and the encoder may leave it
	// clear.
	W int8

	// L is the VEX/EVEX vector length: 0 for 128, 1 for 256, 2 for 512.
	// -1 for legacy forms.
	L int8

	// OpReg is the +rb/+rw/+rd forms, where an operand register rides in
	// the opcode's low three bits instead of in a ModRM field.
	OpReg bool

	// Op16 is the 0x66 operand-size prefix that makes a form 16-bit. It is
	// separate from Pfx because a mandatory prefix and an operand-size
	// override are the same byte meaning two different things, and folding
	// them together would make a 16-bit SSE form unspellable.
	Op16 bool

	// Lock is the F0 prefix. It is part of a form's identity here rather
	// than a modifier on one, because a locking clone is its own row —
	// see table_lock.go — and a caller reaches it by naming it.
	Lock bool

	// FixedModRM is a ModRM byte the form states outright, for the
	// opcodes whose ModRM addresses nothing: MFENCE is 0F AE F0, and the
	// F0 there is not naming a register or a memory operand. Every other
	// row computes the byte from its operands, which is what
	// HasFixedModRM being false means.
	HasFixedModRM bool
	FixedModRM    byte
}

// Ext sentinels.
const (
	SlashR  int8 = -1 // ModRM.reg holds an operand
	NoModRM int8 = -2 // no ModRM byte at all
)

// W sentinel.
const WAny int8 = -1

// L sentinel, for forms with no vector length.
const LNone int8 = -1

type Family uint8

const (
	Legacy Family = iota
	VEX
	EVEX
)

func (f Family) String() string {
	switch f {
	case Legacy:
		return "legacy"
	case VEX:
		return "vex"
	case EVEX:
		return "evex"
	}
	return "family?"
}

// Map is the opcode map a form's primary opcode lives in.
type Map uint8

const (
	Map1  Map = iota // one-byte opcodes, no escape
	Map0F            // 0F escape
	Map0F38
	Map0F3A
)

func (m Map) String() string {
	switch m {
	case Map1:
		return ""
	case Map0F:
		return "0f"
	case Map0F38:
		return "0f38"
	case Map0F3A:
		return "0f3a"
	}
	return "map?"
}

// HasModRM reports whether the form emits a ModRM byte.
func (e Enc) HasModRM() bool { return e.Ext != NoModRM }
