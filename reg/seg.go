package reg

// Sreg is a segment register.
//
// In long mode CS, SS, DS and ES have a base fixed at zero and an override
// naming one of them changes nothing. FS and GS are the two that still do
// something, and FS is where the thread pointer lives on SysV.
type Sreg uint8

const (
	ES Sreg = iota // 0
	CS             // 1
	SS             // 2
	DS             // 3
	FS             // 4
	GS             // 5
)

var namesSreg = [...]string{"es", "cs", "ss", "ds", "fs", "gs"}

// The legacy prefix byte each register is spelled as when it appears as a
// segment override on a memory operand. The numbering above is the sreg3
// field used by MOV Sreg / PUSH Sreg; the bytes below are unrelated to it,
// which is why both live here rather than being computed from each other.
var overrides = [...]byte{
	ES: 0x26,
	CS: 0x2e,
	SS: 0x36,
	DS: 0x3e,
	FS: 0x64,
	GS: 0x65,
}

func (r Sreg) operand()    {}
func (r Sreg) Kind() Kind  { return KindSreg }
func (r Sreg) Num() uint8  { return uint8(r) }
func (r Sreg) Ext() bool   { return false }
func (r Sreg) Ext2() bool  { return false }
func (r Sreg) Valid() bool { return int(r) < len(namesSreg) }

func (r Sreg) String() string {
	if r.Valid() {
		return namesSreg[r]
	}
	return "sreg?"
}

// Override is the legacy prefix byte for a segment override, or 0 if the
// register is not a declared one.
func (r Sreg) Override() byte {
	if r.Valid() {
		return overrides[r]
	}
	return 0
}
