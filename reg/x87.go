package reg

// St is an x87 stack register. ST0 is the top of the stack at the moment the
// instruction executes, not a fixed physical register.
type St uint8

const (
	ST0 St = iota
	ST1
	ST2
	ST3
	ST4
	ST5
	ST6
	ST7
)

var namesSt = [...]string{"st0", "st1", "st2", "st3", "st4", "st5", "st6", "st7"}

func (r St) operand()    {}
func (r St) Kind() Kind  { return KindSt }
func (r St) Num() uint8  { return uint8(r) }
func (r St) Ext() bool   { return false }
func (r St) Ext2() bool  { return false }
func (r St) Valid() bool { return int(r) < len(namesSt) }

func (r St) String() string {
	if r.Valid() {
		return namesSt[r]
	}
	return "st?"
}

// Mm is an MMX register. These alias the x87 stack's mantissa fields, so
// mixing MMX and x87 without EMMS is a correctness problem for the caller;
// nothing in this package can detect it.
type Mm uint8

const (
	MM0 Mm = iota
	MM1
	MM2
	MM3
	MM4
	MM5
	MM6
	MM7
)

var namesMm = [...]string{"mm0", "mm1", "mm2", "mm3", "mm4", "mm5", "mm6", "mm7"}

func (r Mm) operand()    {}
func (r Mm) Kind() Kind  { return KindMm }
func (r Mm) Num() uint8  { return uint8(r) }
func (r Mm) Ext() bool   { return false }
func (r Mm) Ext2() bool  { return false }
func (r Mm) Valid() bool { return int(r) < len(namesMm) }

func (r Mm) String() string {
	if r.Valid() {
		return namesMm[r]
	}
	return "mm?"
}
