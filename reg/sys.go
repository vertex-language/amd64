package reg

// Cr is a control register. CR8 is the task priority register and is the one
// of these a 64-bit program is at all likely to name.
type Cr uint8

const (
	CR0 Cr = iota
	CR1
	CR2
	CR3
	CR4
	CR5
	CR6
	CR7
	CR8
	CR9
	CR10
	CR11
	CR12
	CR13
	CR14
	CR15
)

var namesCr = [...]string{
	"cr0", "cr1", "cr2", "cr3", "cr4", "cr5", "cr6", "cr7",
	"cr8", "cr9", "cr10", "cr11", "cr12", "cr13", "cr14", "cr15",
}

func (r Cr) operand()    {}
func (r Cr) Kind() Kind  { return KindCr }
func (r Cr) Num() uint8  { return uint8(r) }
func (r Cr) Ext() bool   { return r >= CR8 }
func (r Cr) Ext2() bool  { return false }
func (r Cr) Valid() bool { return int(r) < len(namesCr) }

func (r Cr) String() string {
	if r.Valid() {
		return namesCr[r]
	}
	return "cr?"
}

// Dr is a debug register. DR0 through DR7 are the architectural ones; DR8
// through DR15 are encodable via REX.R and are reserved, and naming one is
// the caller's business rather than this package's.
type Dr uint8

const (
	DR0 Dr = iota
	DR1
	DR2
	DR3
	DR4
	DR5
	DR6
	DR7
	DR8
	DR9
	DR10
	DR11
	DR12
	DR13
	DR14
	DR15
)

var namesDr = [...]string{
	"dr0", "dr1", "dr2", "dr3", "dr4", "dr5", "dr6", "dr7",
	"dr8", "dr9", "dr10", "dr11", "dr12", "dr13", "dr14", "dr15",
}

func (r Dr) operand()    {}
func (r Dr) Kind() Kind  { return KindDr }
func (r Dr) Num() uint8  { return uint8(r) }
func (r Dr) Ext() bool   { return r >= DR8 }
func (r Dr) Ext2() bool  { return false }
func (r Dr) Valid() bool { return int(r) < len(namesDr) }

func (r Dr) String() string {
	if r.Valid() {
		return namesDr[r]
	}
	return "dr?"
}

// Tmm is an AMX tile register. A tile has no fixed width — its rows and
// column count are configured by LDTILECFG — which is why Kind.Bits reports
// 0 for this class rather than guessing 8192.
type Tmm uint8

const (
	TMM0 Tmm = iota
	TMM1
	TMM2
	TMM3
	TMM4
	TMM5
	TMM6
	TMM7
)

var namesTmm = [...]string{
	"tmm0", "tmm1", "tmm2", "tmm3", "tmm4", "tmm5", "tmm6", "tmm7",
}

func (r Tmm) operand()    {}
func (r Tmm) Kind() Kind  { return KindTmm }
func (r Tmm) Num() uint8  { return uint8(r) }
func (r Tmm) Ext() bool   { return false }
func (r Tmm) Ext2() bool  { return false }
func (r Tmm) Valid() bool { return int(r) < len(namesTmm) }

func (r Tmm) String() string {
	if r.Valid() {
		return namesTmm[r]
	}
	return "tmm?"
}

// IP is the instruction pointer's type, and RIP is its only member.
//
// It is deliberately not an R64, not a Value, and not an Operand. No
// instruction takes it as an operand — it is a base for a RIP-relative
// address and nothing else — so an interface it satisfied would be a claim
// about a capability that does not exist. Package operand names the constant
// directly.
type IP struct{}

// RIP is the base of a RIP-relative memory operand.
var RIP IP

func (IP) String() string { return "rip" }
