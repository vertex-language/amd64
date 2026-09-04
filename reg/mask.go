package reg

// K is an AVX-512 opmask register.
type K uint8

const (
	K0 K = iota
	K1
	K2
	K3
	K4
	K5
	K6
	K7
)

var namesK = [...]string{"k0", "k1", "k2", "k3", "k4", "k5", "k6", "k7"}

func (r K) operand()    {}
func (r K) Kind() Kind  { return KindK }
func (r K) Num() uint8  { return uint8(r) }
func (r K) Ext() bool   { return false }
func (r K) Ext2() bool  { return false }
func (r K) Valid() bool { return int(r) < len(namesK) }

func (r K) String() string {
	if r.Valid() {
		return namesK[r]
	}
	return "k?"
}

// Writable reports whether the register can serve as an EVEX write mask.
// K0 cannot: that encoding in the aaa field means "no masking", so an
// operand asking for .Mask(K0) is asking for something it will not get, and
// the operand layer refuses it rather than emitting an unmasked instruction
// that looks masked in the source.
//
// K0 is an ordinary source and destination for the mask instructions
// themselves; it is only the write-mask slot it cannot fill.
func (r K) Writable() bool { return r != K0 && r.Valid() }
