package encode

import (
	"github.com/vertex-language/amd64/internal/isa"
	"github.com/vertex-language/amd64/obj"
)

// ModRM: mod[7:6] reg[5:3] r/m[2:0].
func modrm(mod, reg, rm byte) byte {
	return mod<<6 | (reg&7)<<3 | rm&7
}

const (
	modIndirect byte = 0 // [r/m], or the two baseless special cases
	modDisp8    byte = 1
	modDisp32   byte = 2
	modReg      byte = 3

	rmSIB byte = 4 // r/m = 100: a SIB byte follows
	rmRIP byte = 5 // r/m = 101 with mod = 00: [rip + disp32]
)

// regField is what goes in ModRM.reg: an operand register, or the form's
// /digit when the reg field is an opcode extension instead.
func (b *builder) regField() byte {
	if b.f.Enc.Ext >= 0 {
		return byte(b.f.Enc.Ext)
	}
	return b.regNum
}

func (b *builder) emitModRM() error {
	if b.rmIsReg {
		b.buf = append(b.buf, modrm(modReg, b.regField(), b.rmReg.Num()))
		return nil
	}
	if !b.hasMem {
		// A form with a ModRM byte and no r/m operand: the reg field is a
		// /digit and there is nothing to address. Nothing in the table does
		// this today, and if something does it is a table bug, not a
		// caller's.
		return errf(obj.ErrForm, b.f.String()+" has a ModRM byte but no r/m operand")
	}
	return b.emitMem()
}

func (b *builder) emitMem() error {
	m := b.mem
	rf := b.regField()

	switch {
	// [rip + disp32]. The r/m encoding 101 with mod 00 is not affected by
	// REX — selecting R13 with REX.B set and mod 00 still means RIP — so
	// this form has no SIB byte and no base to disambiguate.
	case m.RIP:
		b.buf = append(b.buf, modrm(modIndirect, rf, rmRIP))
		return b.emitDisp32Hole(true)

	// A baseless address: numeric or symbolic, with or without an index.
	// ModRM's disp32 encoding was taken over by RIP, so this has to go
	// through SIB with base = 101 and mod = 00, which is the encoding that
	// means "no base, a disp32 follows". It costs a byte more than the
	// i386 form for exactly that reason.
	case m.Abs || (!m.HasBase && !m.HasIndex):
		b.buf = append(b.buf, modrm(modIndirect, rf, rmSIB))
		idx, scale := noIndex, byte(0)
		if m.HasIndex {
			idx = m.Index.Num()
			var err error
			if scale, err = scaleBits(m.Scale); err != nil {
				return err
			}
		}
		b.buf = append(b.buf, sib(scale, idx, sibNoBase))
		return b.emitDisp32Hole(false)

	// An index with no base is the same SIB no-base form.
	case !m.HasBase:
		scale, err := scaleBits(m.Scale)
		if err != nil {
			return err
		}
		b.buf = append(b.buf, modrm(modIndirect, rf, rmSIB))
		b.buf = append(b.buf, sib(scale, m.Index.Num(), sibNoBase))
		return b.emitDisp32Hole(false)
	}

	// A based address. It may name a symbol only where the symbol's value
	// is an offset rather than an address: a displacement field that both
	// adds to a base register and carries an *address* has no consistent
	// addend, the linker having to know the base's run-time value to work
	// one out. An offset has no such problem, being the number the base
	// wants added to it — which is how a thread-local is reached
	// everywhere. See obj.RefKind.Offset.
	//
	// The check is here rather than after the bytes because an encoder that
	// appends and then fails leaves a partial instruction in the buffer. The
	// error stops the build either way, but a section holding three bytes of
	// a refused instruction is a worse thing to hand a debugger than one
	// holding none.
	if m.HasRef && !m.Ref.Kind.Offset() {
		return errf(obj.ErrOperand,
			"a symbol reference needs a baseless or RIP-relative address",
			"a displacement field that both adds to a base register and carries a relocation has no consistent addend",
			"an offset kind — secrel32, tpoff, dtpoff — may share the field, being a number rather than an address")
	}

	// Two facts decide the shape, and both are properties of the base
	// register's low three bits:
	//
	//   RSP and R12 (low bits 100) need a SIB byte, because 100 in r/m is
	//   the SIB escape and there is no way to say "just this register".
	//
	//   RBP and R13 (low bits 101) need an explicit displacement even when
	//   it is zero, because 101 with mod 00 means RIP-relative.
	mod := modIndirect
	switch {
	case m.HasRef:
		// The displacement is a relocation, so the field is four bytes
		// whatever the offset turns out to be. Nothing here knows it, and
		// a byte-sized guess is one the linker could not widen.
		mod = modDisp32
	case m.Disp == 0 && !m.Base.NeedsDisp():
		mod = modIndirect
	case fitsInt8(m.Disp):
		mod = modDisp8
	default:
		mod = modDisp32
	}

	if m.HasIndex || m.Base.NeedsSIB() {
		idx, scale := noIndex, byte(0)
		if m.HasIndex {
			idx = m.Index.Num()
			var err error
			if scale, err = scaleBits(m.Scale); err != nil {
				return err
			}
		}
		b.buf = append(b.buf, modrm(mod, rf, rmSIB))
		b.buf = append(b.buf, sib(scale, idx, m.Base.Num()))
	} else {
		b.buf = append(b.buf, modrm(mod, rf, m.Base.Num()))
	}

	switch mod {
	case modDisp8:
		b.buf = append(b.buf, byte(m.Disp))
	case modDisp32:
		// The hole writer handles both: a plain displacement is four
		// bytes of it, and a reference is four bytes and an entry.
		return b.emitDisp32Hole(false)
	}
	return nil
}

// emitDisp32Hole writes four zero bytes and records what fills them.
func (b *builder) emitDisp32Hole(pcrel bool) error {
	off := len(b.buf)
	m := b.mem

	if !m.HasRef {
		b.appendInt(int64(m.Disp), 4)
		return nil
	}

	b.appendInt(0, 4)

	// RefGOTPCRELX states that the addend is exactly the one the psABI's
	// relaxation assumes, and the relaxation has to know how many bytes
	// back the instruction starts — which depends on whether a REX prefix
	// is there. The encoder knows, because it emitted the prefix or did
	// not, but it is not the encoder's call to make. So the kind is the
	// caller's to state and this is where the wrong one is refused.
	if m.Ref.Kind == obj.RefGOTPCRELX && b.carriesREX() {
		return errf(obj.ErrRefKind,
			"RefGOTPCRELX on a form that carries a REX prefix",
			"the REX-prefixed relaxation is RefRexGOTPCRELX; the two differ only in how far back the linker looks")
	}

	b.refs = append(b.refs, RefHole{
		Offset: off, Size: 4, PCRel: pcrel,
		Sym: m.Ref.Sym, Kind: m.Ref.Kind, Addend: m.Ref.Addend,
	})
	return nil
}

// carriesREX reports whether a REX prefix was emitted for this instruction.
func (b *builder) carriesREX() bool {
	if b.f.Enc.Family != isa.Legacy {
		return false
	}
	return b.rexBits() != 0 || b.rexForcedByByteReg()
}

func fitsInt8(d int32) bool { return d >= -128 && d <= 127 }
