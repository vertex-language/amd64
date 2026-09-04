package reg

// Kind names a register class. It is what a diagnostic prints when it has to
// say which class an operand belonged to.
type Kind uint8

const (
	KindInvalid Kind = iota
	KindR8
	KindR16
	KindR32
	KindR64
	KindSreg
	KindSt
	KindMm
	KindXmm
	KindYmm
	KindZmm
	KindK
	KindCr
	KindDr
	KindTmm
)

var kindNames = [...]string{
	KindInvalid: "invalid",
	KindR8:      "r8",
	KindR16:     "r16",
	KindR32:     "r32",
	KindR64:     "r64",
	KindSreg:    "sreg",
	KindSt:      "st",
	KindMm:      "mm",
	KindXmm:     "xmm",
	KindYmm:     "ymm",
	KindZmm:     "zmm",
	KindK:       "k",
	KindCr:      "cr",
	KindDr:      "dr",
	KindTmm:     "tmm",
}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind?"
}

// Bits is the architectural width of the class, or 0 where the class has no
// single width (KindTmm — a tile's shape is configured by LDTILECFG).
func (k Kind) Bits() int {
	switch k {
	case KindR8:
		return 8
	case KindR16, KindSreg:
		return 16
	case KindR32:
		return 32
	case KindMm, KindK, KindCr, KindDr, KindR64:
		return 64
	case KindSt:
		return 80
	case KindXmm:
		return 128
	case KindYmm:
		return 256
	case KindZmm:
		return 512
	}
	return 0
}

// Value is what every register type has in common. It is closed: Operand's
// method is unexported and declared here, so nothing outside this package
// can be a Value.
//
// Num, Ext and Ext2 are the three things the encoder asks a register. Num is
// the low three bits of the ModRM, SIB or opcode field; Ext is the fourth bit
// (REX.R/X/B, or the inverted VEX equivalent); Ext2 is the fifth, which only
// EVEX can carry.
type Value interface {
	Operand

	Kind() Kind

	// Num is the register's encoding number, 0-31. For the high byte
	// registers this is the number they share with SPL/BPL/SIL/DIL: AH is 4,
	// not 16.
	Num() uint8

	// Ext reports whether Num is 8 or above, so the field needs its fourth bit.
	Ext() bool

	// Ext2 reports whether Num is 16 or above, so the field needs its fifth
	// bit and the instruction must be EVEX-encoded. Only the vector classes
	// can return true.
	Ext2() bool

	// Valid reports whether the value names a declared register. It is false
	// only for a value produced by an out-of-range conversion.
	Valid() bool

	String() string
}
