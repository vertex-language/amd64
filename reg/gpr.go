package reg

// R64 is a 64-bit general-purpose register.
type R64 uint8

const (
	RAX R64 = iota
	RCX
	RDX
	RBX
	RSP
	RBP
	RSI
	RDI
	R8Q
	R9Q
	R10Q
	R11Q
	R12Q
	R13Q
	R14Q
	R15Q
)

var names64 = [...]string{
	"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi",
	"r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15",
}

func (r R64) operand()    {}
func (r R64) Kind() Kind  { return KindR64 }
func (r R64) Num() uint8  { return uint8(r) }
func (r R64) Ext() bool   { return r >= R8Q }
func (r R64) Ext2() bool  { return false }
func (r R64) Valid() bool { return int(r) < len(names64) }
func (r R64) String() string {
	if r.Valid() {
		return names64[r]
	}
	return "r64?"
}

// NoIndex reports whether the register cannot appear in the SIB index field.
// Only RSP: the index field has no encoding for it, because index 4 with
// REX.X clear is what "no index" means. R12 has the same low three bits and
// is a perfectly good index, because REX.X is set and that is what tells the
// two apart.
func (r R64) NoIndex() bool { return r == RSP }

// NeedsSIB reports whether the register as a base forces a SIB byte: its low
// three bits are 4, which is the ModRM escape into SIB. RSP and R12.
func (r R64) NeedsSIB() bool { return r&7 == 4 }

// NeedsDisp reports whether the register as a base forces an explicit
// displacement even when the displacement is zero: its low three bits are 5,
// which with mod=00 means RIP-relative rather than [base]. RBP and R13.
func (r R64) NeedsDisp() bool { return r&7 == 5 }

// R32 is a 32-bit general-purpose register.
//
// A write to one of these zeroes the upper half of the containing 64-bit
// register. That is the ISA, not a convenience, and nothing in this tree
// hides it.
type R32 uint8

const (
	EAX R32 = iota
	ECX
	EDX
	EBX
	ESP
	EBP
	ESI
	EDI
	R8D
	R9D
	R10D
	R11D
	R12D
	R13D
	R14D
	R15D
)

var names32 = [...]string{
	"eax", "ecx", "edx", "ebx", "esp", "ebp", "esi", "edi",
	"r8d", "r9d", "r10d", "r11d", "r12d", "r13d", "r14d", "r15d",
}

func (r R32) operand()    {}
func (r R32) Kind() Kind  { return KindR32 }
func (r R32) Num() uint8  { return uint8(r) }
func (r R32) Ext() bool   { return r >= R8D }
func (r R32) Ext2() bool  { return false }
func (r R32) Valid() bool { return int(r) < len(names32) }
func (r R32) String() string {
	if r.Valid() {
		return names32[r]
	}
	return "r32?"
}

// R16 is a 16-bit general-purpose register.
//
// A write to one of these leaves the upper 48 bits alone.
type R16 uint8

const (
	AX R16 = iota
	CX
	DX
	BX
	SP
	BP
	SI
	DI
	R8W
	R9W
	R10W
	R11W
	R12W
	R13W
	R14W
	R15W
)

var names16 = [...]string{
	"ax", "cx", "dx", "bx", "sp", "bp", "si", "di",
	"r8w", "r9w", "r10w", "r11w", "r12w", "r13w", "r14w", "r15w",
}

func (r R16) operand()    {}
func (r R16) Kind() Kind  { return KindR16 }
func (r R16) Num() uint8  { return uint8(r) }
func (r R16) Ext() bool   { return r >= R8W }
func (r R16) Ext2() bool  { return false }
func (r R16) Valid() bool { return int(r) < len(names16) }
func (r R16) String() string {
	if r.Valid() {
		return names16[r]
	}
	return "r16?"
}

// R8 is an 8-bit general-purpose register.
//
// This class has twenty members, not sixteen, because encodings 4 through 7
// mean two different things depending on whether a REX prefix is present:
// AH, CH, DH and BH without one, SPL, BPL, SIL and DIL with one. The two
// groups are separate constants here and Num folds them back onto the same
// four numbers, so the encoder is told the number and asked the question
// separately.
type R8 uint8

const (
	AL   R8 = iota // 0
	CL             // 1
	DL             // 2
	BL             // 3
	SPL            // 4 — requires REX
	BPL            // 5 — requires REX
	SIL            // 6 — requires REX
	DIL            // 7 — requires REX
	R8B            // 8
	R9B            // 9
	R10B           // 10
	R11B           // 11
	R12B           // 12
	R13B           // 13
	R14B           // 14
	R15B           // 15

	// The high byte registers. Their stored values are 16-19; they encode as
	// 4-7 and are unreachable from any instruction carrying REX.
	AH R8 = 16
	CH R8 = 17
	DH R8 = 18
	BH R8 = 19
)

var names8 = [...]string{
	"al", "cl", "dl", "bl", "spl", "bpl", "sil", "dil",
	"r8b", "r9b", "r10b", "r11b", "r12b", "r13b", "r14b", "r15b",
	"ah", "ch", "dh", "bh",
}

func (r R8) operand()   {}
func (r R8) Kind() Kind { return KindR8 }

// Num is the encoding number. AH, CH, DH and BH report 4, 5, 6 and 7 — the
// numbers they share with SPL, BPL, SIL and DIL.
func (r R8) Num() uint8 {
	if r.High() {
		return uint8(r) - uint8(AH) + 4
	}
	return uint8(r)
}

func (r R8) Ext() bool   { return !r.High() && r >= R8B }
func (r R8) Ext2() bool  { return false }
func (r R8) Valid() bool { return int(r) < len(names8) }

func (r R8) String() string {
	if r.Valid() {
		return names8[r]
	}
	return "r8?"
}

// High reports whether the register is AH, CH, DH or BH — reachable only
// when no REX prefix is present. An instruction that names one of these and
// anything requiring REX has no encoding, and the encoder refuses it as
// ErrOperand naming both registers.
func (r R8) High() bool { return r >= AH }

// NeedsREX reports whether the register is reachable only with a REX prefix
// present: SPL, BPL, SIL, DIL, and R8B through R15B. The first four need a
// bare 0x40 with no bits set, which is the one case where REX carries no
// payload and is emitted anyway.
func (r R8) NeedsREX() bool { return !r.High() && r >= SPL }
