package encode

import (
	"github.com/vertex-language/amd64/internal/isa"
	"github.com/vertex-language/amd64/obj"
)

// pp is the two-bit stand-in for a mandatory SIMD prefix.
func ppBits(pfx byte, op16 bool) (byte, error) {
	switch {
	case op16 || pfx == 0x66:
		return 1, nil
	case pfx == 0xf3:
		return 2, nil
	case pfx == 0xf2:
		return 3, nil
	case pfx == 0:
		return 0, nil
	}
	return 0, errf(obj.ErrForm, "not a SIMD prefix")
}

// mmmmm is the opcode map, which VEX carries as a field rather than as
// escape bytes.
func mapBits(m isa.Map) (byte, error) {
	switch m {
	case isa.Map0F:
		return 1, nil
	case isa.Map0F38:
		return 2, nil
	case isa.Map0F3A:
		return 3, nil
	}
	return 0, errf(obj.ErrForm, "a VEX form must live in the 0F, 0F38 or 0F3A map")
}

// emitVEX writes the two-byte form when every field it cannot carry is
// clear, and the three-byte form otherwise.
//
// The choice is not a form and is not in the table: an instruction that
// fits C5 is byte-for-byte the same instruction as its C4 spelling, one
// byte shorter, and there is no case where a caller wants the longer one.
func (b *builder) emitVEX() error {
	e := b.f.Enc

	pp, err := ppBits(e.Pfx, e.Op16)
	if err != nil {
		return err
	}
	mm, err := mapBits(e.Map)
	if err != nil {
		return err
	}

	var r, x, bb byte
	if b.hasReg && b.regExt {
		r = 1
	}
	switch {
	case b.rmIsReg:
		if b.rmReg.Ext() {
			bb = 1
		}
	case b.hasMem:
		if b.mem.HasBase && b.mem.Base.Ext() {
			bb = 1
		}
		if b.mem.HasIndex && b.mem.Index.Ext() {
			x = 1
		}
	}

	// vvvv is inverted, and a form that does not use it must write 1111 —
	// otherwise the instruction is #UD rather than merely odd.
	var vv byte = 0
	if b.hasVVVV {
		vv = b.vvvv & 0xf
	}

	var l byte
	if e.L == 1 {
		l = 1
	}

	w := byte(0)
	if e.W == 1 {
		w = 1
	}

	// The two-byte form carries R, vvvv, L and pp and nothing else. It is
	// available exactly when X, B and W are clear and the map is 0F.
	if x == 0 && bb == 0 && w == 0 && mm == 1 {
		b.buf = append(b.buf, 0xc5,
			(^r&1)<<7|(^vv&0xf)<<3|l<<2|pp)
		return nil
	}

	b.buf = append(b.buf, 0xc4,
		(^r&1)<<7|(^x&1)<<6|(^bb&1)<<5|mm,
		w<<7|(^vv&0xf)<<3|l<<2|pp)
	return nil
}
