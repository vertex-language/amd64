package encode

// REX bits, in the order the byte carries them: 0100 W R X B.
const (
	rexBase = 0x40
	rexW    = 0x08
	rexR    = 0x04
	rexX    = 0x02
	rexB    = 0x01
)

// rexBits computes the payload from the operands. It is arithmetic over
// register numbers and nothing a table row could know, which is why the
// prefix is not a form.
func (b *builder) rexBits() byte {
	var v byte

	if b.f.Enc.W == 1 {
		v |= rexW
	}
	if b.hasReg && b.regExt {
		if b.regIsOpReg {
			v |= rexB
		} else {
			v |= rexR
		}
	}

	switch {
	case b.rmIsReg:
		if b.rmReg.Ext() {
			v |= rexB
		}
	case b.hasMem:
		if b.mem.HasBase && b.mem.Base.Ext() {
			v |= rexB
		}
		if b.mem.HasIndex && b.mem.Index.Ext() {
			v |= rexX
		}
	}
	return v
}

// emitREX writes the prefix when a field needs it, and omits it when none
// does — with one exception that is not an exception at all: SPL, BPL, SIL
// and DIL are reachable only when REX is present, so a bare 0x40 with no
// bits set is emitted for them. That is the only case where the prefix
// carries no payload and is still required.
func (b *builder) emitREX() error {
	v := b.rexBits()
	if v == 0 && !b.rexForcedByByteReg() {
		return nil
	}
	b.buf = append(b.buf, rexBase|v)
	return nil
}

func (b *builder) rexForcedByByteReg() bool {
	if !b.rexForced {
		return false
	}
	r8, ok := b.rexForcer.(interface{ NeedsREX() bool })
	return ok && r8.NeedsREX()
}
