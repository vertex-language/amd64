package encode

import "github.com/vertex-language/amd64/obj"

// EVEX is four bytes: 62, then
//
//	P0  R X B R' 0 m m m
//	P1  W v v v v 1 p p
//	P2  z L'L b V' a a a
//
// R, X, B, R' and V' are inverted. R' and V' are the fifth bits that reach
// registers 16 through 31, which is why Ext2 exists on the vector classes
// and returns false everywhere else.
func (b *builder) emitEVEX() error {
	e := b.f.Enc

	pp, err := ppBits(e.Pfx, e.Op16)
	if err != nil {
		return err
	}
	mm, err := mapBits(e.Map)
	if err != nil {
		return err
	}

	var r, r2, x, bb byte
	if b.hasReg {
		if b.regExt {
			r = 1
		}
		if b.regExt2 {
			r2 = 1
		}
	}
	switch {
	case b.rmIsReg:
		if b.rmReg.Ext() {
			bb = 1
		}
		// A vector r/m register at 16 or above puts its fifth bit in X,
		// which is otherwise the SIB index extension and is free here
		// because a register r/m has no SIB byte.
		if b.rmReg.Ext2() {
			x = 1
		}
	case b.hasMem:
		if b.mem.HasBase && b.mem.Base.Ext() {
			bb = 1
		}
		if b.mem.HasIndex && b.mem.Index.Ext() {
			x = 1
		}
	}

	var vv, v2 byte
	if b.hasVVVV {
		vv = b.vvvv & 0xf
		if b.vvvvExt2 {
			v2 = 1
		}
	}

	w := byte(0)
	if e.W == 1 {
		w = 1
	}

	var ll byte
	switch e.L {
	case 0:
		ll = 0
	case 1:
		ll = 1
	case 2:
		ll = 2
	default:
		return errf(obj.ErrForm, "an EVEX form must state a vector length")
	}

	var bcst byte
	if b.hasMem && b.mem.Bcst {
		bcst = 1
	}

	b.buf = append(b.buf, 0x62,
		(^r&1)<<7|(^x&1)<<6|(^bb&1)<<5|(^r2&1)<<4|mm,
		w<<7|(^vv&0xf)<<3|1<<2|pp,
		b.mask.z<<7|ll<<5|bcst<<4|(^v2&1)<<3|b.mask.aaa)

	return nil
}

// maskState is the {k}{z} decoration. It is zero today: no EVEX row is
// declared, and masking rides on a register destination, which needs a
// wrapper type in operand that should land with the EVEX tranche rather
// than half-land now. The field is here so the byte layout above is the
// real one and not a sketch.
type maskState struct {
	aaa byte // opmask register number, 0 for none
	z   byte // zeroing rather than merging
}
