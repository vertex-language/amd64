package reg

// The vector classes run 0-31. Registers 16-31 exist only under AVX-512:
// their fifth encoding bit has no home in a legacy or VEX prefix, so a form
// with no EVEX encoding cannot name one. That is Ext2's whole job — the
// encoder asks, and a true answer on a non-EVEX form is a refusal rather
// than a silently truncated register number.

// Xmm is a 128-bit vector register.
type Xmm uint8

// Ymm is a 256-bit vector register.
type Ymm uint8

// Zmm is a 512-bit vector register.
type Zmm uint8

const (
	XMM0 Xmm = iota
	XMM1
	XMM2
	XMM3
	XMM4
	XMM5
	XMM6
	XMM7
	XMM8
	XMM9
	XMM10
	XMM11
	XMM12
	XMM13
	XMM14
	XMM15
	XMM16
	XMM17
	XMM18
	XMM19
	XMM20
	XMM21
	XMM22
	XMM23
	XMM24
	XMM25
	XMM26
	XMM27
	XMM28
	XMM29
	XMM30
	XMM31
)

const (
	YMM0 Ymm = iota
	YMM1
	YMM2
	YMM3
	YMM4
	YMM5
	YMM6
	YMM7
	YMM8
	YMM9
	YMM10
	YMM11
	YMM12
	YMM13
	YMM14
	YMM15
	YMM16
	YMM17
	YMM18
	YMM19
	YMM20
	YMM21
	YMM22
	YMM23
	YMM24
	YMM25
	YMM26
	YMM27
	YMM28
	YMM29
	YMM30
	YMM31
)

const (
	ZMM0 Zmm = iota
	ZMM1
	ZMM2
	ZMM3
	ZMM4
	ZMM5
	ZMM6
	ZMM7
	ZMM8
	ZMM9
	ZMM10
	ZMM11
	ZMM12
	ZMM13
	ZMM14
	ZMM15
	ZMM16
	ZMM17
	ZMM18
	ZMM19
	ZMM20
	ZMM21
	ZMM22
	ZMM23
	ZMM24
	ZMM25
	ZMM26
	ZMM27
	ZMM28
	ZMM29
	ZMM30
	ZMM31
)

// vecCount is the size of every vector class.
const vecCount = 32

// vecNames holds the numeric suffix for each register, shared by the three
// classes so the spellings cannot drift apart.
var vecNames = [vecCount]string{
	"0", "1", "2", "3", "4", "5", "6", "7",
	"8", "9", "10", "11", "12", "13", "14", "15",
	"16", "17", "18", "19", "20", "21", "22", "23",
	"24", "25", "26", "27", "28", "29", "30", "31",
}

func (r Xmm) operand()    {}
func (r Xmm) Kind() Kind  { return KindXmm }
func (r Xmm) Num() uint8  { return uint8(r) }
func (r Xmm) Ext() bool   { return r >= XMM8 }
func (r Xmm) Ext2() bool  { return r >= XMM16 }
func (r Xmm) Valid() bool { return int(r) < vecCount }

func (r Xmm) String() string {
	if r.Valid() {
		return "xmm" + vecNames[r]
	}
	return "xmm?"
}

func (r Ymm) operand()    {}
func (r Ymm) Kind() Kind  { return KindYmm }
func (r Ymm) Num() uint8  { return uint8(r) }
func (r Ymm) Ext() bool   { return r >= YMM8 }
func (r Ymm) Ext2() bool  { return r >= YMM16 }
func (r Ymm) Valid() bool { return int(r) < vecCount }

func (r Ymm) String() string {
	if r.Valid() {
		return "ymm" + vecNames[r]
	}
	return "ymm?"
}

func (r Zmm) operand()    {}
func (r Zmm) Kind() Kind  { return KindZmm }
func (r Zmm) Num() uint8  { return uint8(r) }
func (r Zmm) Ext() bool   { return r >= ZMM8 }
func (r Zmm) Ext2() bool  { return r >= ZMM16 }
func (r Zmm) Valid() bool { return int(r) < vecCount }

func (r Zmm) String() string {
	if r.Valid() {
		return "zmm" + vecNames[r]
	}
	return "zmm?"
}
