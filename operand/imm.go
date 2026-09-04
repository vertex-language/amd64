package operand

import "github.com/vertex-language/amd64/reg"

// Imm is an immediate. It carries a value and no width: the form the caller
// named pins the field, and a value that does not fit it is ErrRange from
// the encoder with the field width and range in the notes.
//
// It exists as a type rather than a bare int64 because Emit's variadic
// takes Operand, which is sealed. A bare Go integer does not coerce, and
// NewImm(n) is the spelling — that seal is what keeps an arbitrary type out
// of a mnemonic-as-data path.
type Imm struct {
	reg.Seal
	v int64
}

// NewImm returns an immediate holding n.
func NewImm(n int64) Imm { return Imm{v: n} }

// NewImmU returns an immediate holding n, reinterpreted as signed. It is
// for the cases where a caller has a bit pattern rather than a number —
// 0xdeadbeefcafef00d into MovR64Imm64, which is a perfectly ordinary thing
// to want and an awkward int64 literal to write.
func NewImmU(n uint64) Imm { return Imm{v: int64(n)} }

// Int64 is the value.
func (i Imm) Int64() int64 { return i.v }

// Uint64 is the value's bit pattern.
func (i Imm) Uint64() uint64 { return uint64(i.v) }

// FitsSigned reports whether the value fits a signed field of the given
// width in bits. The encoder asks this; a caller generally should not have
// to, because naming a form is what states the field.
func (i Imm) FitsSigned(bits int) bool {
	if bits >= 64 {
		return true
	}
	lo := int64(-1) << (bits - 1)
	hi := int64(1)<<(bits-1) - 1
	return i.v >= lo && i.v <= hi
}

// FitsUnsigned reports whether the value's bit pattern fits an unsigned
// field of the given width in bits.
func (i Imm) FitsUnsigned(bits int) bool {
	if bits >= 64 {
		return true
	}
	return uint64(i.v) < uint64(1)<<bits
}
