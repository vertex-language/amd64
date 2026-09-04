// Package encode turns a form and its operands into bytes and holes.
//
// It is where every prefix decision lives, and none of those decisions is a
// choice. REX is emitted when a field needs it and omitted when no field
// does. A VEX form takes the two-byte C5 when the fields fit and the
// three-byte C4 when they do not. Neither is a form, neither is in the
// table, and there is no case where a caller would want the longer one.
package encode

import (
	"github.com/vertex-language/amd64/internal/isa"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// Inst is one encoded instruction: its bytes, and the holes a later stage
// fills. Offsets are relative to the first byte of the instruction; the
// section adds its own.
type Inst struct {
	Bytes  []byte
	Refs   []RefHole
	Labels []LabelHole
}

// RefHole is a field the linker fills. It becomes an obj.Reference once the
// section knows where the instruction landed.
type RefHole struct {
	Offset int
	Size   int
	PCRel  bool
	Adjust int64
	Sym    string
	Kind   obj.RefKind
	Addend int64
}

// LabelHole is a field Finalize fills from a same-section label. It never
// reaches a linker and leaves no relocation.
type LabelHole struct {
	Offset int
	Size   int
	PCRel  bool
	Adjust int64
	Label  string
}

// role is where an operand goes.
type role uint8

const (
	roleNone  role = iota // a fixed operand; the form named it, nothing is emitted
	roleReg               // ModRM.reg
	roleRM                // ModRM.r/m, with SIB and displacement as needed
	roleVVVV              // VEX/EVEX.vvvv
	roleOpReg             // the opcode's low three bits
	roleImm
	roleRel
)

// plan assigns each operand position a role from its class.
//
// The rule is positional and falls out of how the ISA is shaped: an r/m
// class goes to r/m, an immediate to the immediate, and a register class
// goes to reg the first time and to vvvv the second. That last clause is
// the whole of three-operand VEX — vaddps ymm0, ymm1, ymm2 puts ymm0 in
// reg, ymm1 in vvvv and ymm2 in r/m, in that order, because that is the
// order the mnemonic writes them.
//
// A legacy form has one more place to put a register, and some of them use
// it: r/m addresses a register as readily as memory, so an operand that can
// only ever be a register still lands there when the form has no r/m class
// of its own. PMOVMSKB r32, xmm is the shape — the xmm is in r/m and m128
// is not encodable — and PSLLW xmm, imm8 is the same thing with the reg
// field spent on a /digit, leaving r/m as the only field the register can
// be in. Spelling those classes xmm/m128 instead would be a lie a caller
// could take up on.
//
// The clause is legacy-only on purpose. A VEX form always has vvvv, so a
// register with no r/m class of its own belongs there, and which of the two
// a /digit form's operands means is a question this rule would answer
// wrongly.
func plan(f *isa.Form) []role {
	roles := make([]role, len(f.Ops))
	regTaken := false
	rmFree := f.Enc.Family == isa.Legacy && f.Enc.Ext != isa.NoModRM && !f.Enc.OpReg
	for _, c := range f.Ops {
		if isRM(c) {
			rmFree = false
		}
	}

	for i, c := range f.Ops {
		switch {
		case c.Fixed():
			roles[i] = roleNone

		case c == isa.Imm8, c == isa.Imm8S, c == isa.Imm16, c == isa.Imm32, c == isa.Imm64:
			roles[i] = roleImm

		case c == isa.Rel8, c == isa.Rel32:
			roles[i] = roleRel

		case isRM(c):
			roles[i] = roleRM

		default: // a register class
			switch {
			case f.Enc.OpReg && !regTaken:
				roles[i] = roleOpReg
				regTaken = true
			case !regTaken && f.Enc.Ext == isa.SlashR:
				roles[i] = roleReg
				regTaken = true
			case rmFree:
				roles[i] = roleRM
				rmFree = false
			default:
				roles[i] = roleVVVV
			}
		}
	}
	return roles
}

// isRM is the half of plan's rule that depends on the class alone: whether
// an operand position addresses memory, and so has to be the r/m one.
func isRM(c isa.Class) bool {
	switch c {
	case isa.RM8, isa.RM16, isa.RM32, isa.RM64,
		isa.RM128, isa.RM256, isa.RM512,
		isa.XM32, isa.XM64, isa.M:
		return true
	}
	return false
}

// Encode produces the instruction's bytes and holes.
//
// Feature gating is not here. A gate is a fact about the module, the module
// is the section's, and a gated form reaching this function has already
// been permitted.
func Encode(f *isa.Form, ops []operand.Operand) (*Inst, error) {
	b := &builder{f: f, ops: ops, roles: plan(f)}

	if err := b.gather(); err != nil {
		return nil, err
	}
	if err := b.checkHighByte(); err != nil {
		return nil, err
	}
	if err := b.emit(); err != nil {
		return nil, err
	}

	// Adjust is the correction that makes the identity hold:
	//
	//   value = target - (section offset of the field) + Adjust + Addend
	//
	// A PC-relative field resolves against the end of the instruction, and
	// the field is not always the last thing in it. The encoder placed the
	// field, so it is the only thing that knows how many bytes follow —
	// which is exactly what this subtraction is.
	total := len(b.buf)
	for i := range b.refs {
		if b.refs[i].PCRel {
			b.refs[i].Adjust = -int64(total - b.refs[i].Offset)
		}
	}
	for i := range b.labels {
		if b.labels[i].PCRel {
			b.labels[i].Adjust = -int64(total - b.labels[i].Offset)
		}
	}

	return &Inst{Bytes: b.buf, Refs: b.refs, Labels: b.labels}, nil
}

// Length encodes and reports the byte count, discarding the result.
//
// Emit's shortest-wins rule calls this on every candidate. There is
// deliberately no size estimator anywhere in this tree, because a second
// implementation of "how long is this" is a second thing that can be wrong,
// and it would be wrong in exactly the cases that matter.
func Length(f *isa.Form, ops []operand.Operand) (int, error) {
	in, err := Encode(f, ops)
	if err != nil {
		return 0, err
	}
	return len(in.Bytes), nil
}

type builder struct {
	f     *isa.Form
	ops   []operand.Operand
	roles []role

	// Gathered operands.
	regNum  uint8 // ModRM.reg or the opcode register
	regExt  bool
	regExt2 bool
	hasReg  bool

	// regIsOpReg distinguishes which of the two roleReg/roleOpReg wrote
	// regNum/regExt: a real ModRM.reg extends via REX.R, but the register
	// riding in the opcode's own low three bits (+rb/+rw/+rd) extends via
	// REX.B instead — the same bit ModRM.rm would use, which is exactly why
	// this form never has a ModRM byte to also need it. rexBits reads this
	// to route the bit correctly.
	regIsOpReg bool

	// vvvv is a four-bit field, so Num carries the bit Ext reports and
	// there is no vvvvExt. vvvvExt2 is the fifth bit, which only EVEX's V'
	// can hold.
	vvvv     uint8
	vvvvExt2 bool
	hasVVVV  bool

	// mask is the EVEX {k}{z} decoration. It is zero until an EVEX row is
	// declared; the field is here so evex.go's byte layout is the real one.
	mask maskState

	rmReg   reg.Value // set when r/m is a register
	rmIsReg bool
	mem     operand.Parts
	hasMem  bool

	imm       int64
	immSize   int
	immSigned bool
	hasImm    bool

	rel     operand.Operand
	relSize int
	hasRel  bool

	// High-byte tracking, for the diagnostic.
	highByte  reg.R8
	hasHigh   bool
	rexForcer reg.Value
	rexForced bool

	buf    []byte
	refs   []RefHole
	labels []LabelHole
}

func (b *builder) gather() error {
	for i, r := range b.roles {
		op := b.ops[i]

		switch r {
		case roleNone:
			// The form named it. Nothing to emit, but a fixed AL still has
			// to be seen by the high-byte check below, and it cannot be
			// high, so there is nothing to do here either.

		case roleReg, roleOpReg:
			v, ok := op.(reg.Value)
			if !ok {
				return errf(obj.ErrForm, "operand "+itoa(i)+" is not a register")
			}
			b.regNum, b.regExt, b.regExt2, b.hasReg = v.Num(), v.Ext(), v.Ext2(), true
			b.regIsOpReg = r == roleOpReg
			b.noteByte(op, v)

		case roleVVVV:
			v, ok := op.(reg.Value)
			if !ok {
				return errf(obj.ErrForm, "operand "+itoa(i)+" is not a register")
			}
			b.vvvv, b.vvvvExt2, b.hasVVVV = v.Num(), v.Ext2(), true

		case roleRM:
			switch m := op.(type) {
			case operand.Memory:
				if err := m.Err(); err != nil {
					return err
				}
				b.mem, b.hasMem = m.Addr().Parts(), true
				if b.mem.HasIndex && b.mem.Index.NoIndex() {
					return errf(obj.ErrOperand, "rsp cannot be a SIB index register")
				}
			default:
				v, ok := op.(reg.Value)
				if !ok {
					return errf(obj.ErrForm, "operand "+itoa(i)+" is neither a register nor an address")
				}
				b.rmReg, b.rmIsReg = v, true
				b.noteByte(op, v)
			}

		case roleImm:
			im, ok := op.(operand.Imm)
			if !ok {
				return errf(obj.ErrForm, "operand "+itoa(i)+" is not an immediate")
			}
			b.imm, b.hasImm = im.Int64(), true
			switch b.f.Ops[i] {
			case isa.Imm8:
				b.immSize = 1
			case isa.Imm8S:
				b.immSize, b.immSigned = 1, true
			case isa.Imm16:
				b.immSize = 2
			case isa.Imm32:
				b.immSize = 4
			case isa.Imm64:
				b.immSize = 8
			}
			if !fits(b.imm, b.immSize, b.immSigned) {
				return errf(obj.ErrRange,
					"does not fit "+b.f.String(),
					"the immediate field of "+b.f.String()+" is "+itoa(b.immSize)+" bytes; the range is "+rangeText(b.immSize, b.immSigned))
			}

		case roleRel:
			b.rel, b.hasRel = op, true
			b.relSize = b.f.RelWidth()
		}
	}
	return nil
}

// noteByte records the two byte-register facts the REX decision depends on.
func (b *builder) noteByte(op operand.Operand, v reg.Value) {
	r8, ok := op.(reg.R8)
	if !ok {
		if v.Ext() {
			b.rexForcer, b.rexForced = v, true
		}
		return
	}
	switch {
	case r8.High():
		b.highByte, b.hasHigh = r8, true
	case r8.NeedsREX():
		b.rexForcer, b.rexForced = r8, true
	}
}

// checkHighByte refuses the one combination the silicon cannot encode.
//
// AH, CH, DH and BH are reachable only when no REX prefix is present,
// because those encodings are how SPL, BPL, SIL and DIL are reached once
// REX is there. It is ErrOperand rather than ErrForm because the operand
// kinds were right; the combination is what has no encoding.
func (b *builder) checkHighByte() error {
	if !b.hasHigh {
		return nil
	}
	forcer, forced := b.rexForcer, b.rexForced
	if !forced && b.hasMem {
		switch {
		case b.mem.HasBase && b.mem.Base.Ext():
			forcer, forced = b.mem.Base, true
		case b.mem.HasIndex && b.mem.Index.Ext():
			forcer, forced = b.mem.Index, true
		}
	}
	if !forced && b.f.Enc.W == 1 {
		return errf(obj.ErrOperand,
			b.highByte.String()+" cannot be used in a REX-prefixed instruction",
			b.f.String()+" carries REX.W, and REX makes encoding "+itoa(int(b.highByte.Num()))+" mean "+lowByteName(b.highByte)+" instead")
	}
	if !forced {
		return nil
	}
	return errf(obj.ErrOperand,
		b.highByte.String()+" and "+forcer.String()+" cannot appear in the same instruction",
		forcer.String()+" forces a REX prefix, and REX makes encoding "+itoa(int(b.highByte.Num()))+" mean "+lowByteName(b.highByte)+" instead of "+b.highByte.String())
}

func lowByteName(h reg.R8) string {
	switch h {
	case reg.AH:
		return "spl"
	case reg.CH:
		return "bpl"
	case reg.DH:
		return "sil"
	case reg.BH:
		return "dil"
	}
	return "?"
}

func (b *builder) emit() error {
	e := b.f.Enc

	// Legacy prefixes, in the order a decoder expects them: LOCK, then
	// the segment override, then the operand-size or mandatory SIMD
	// prefix, then REX immediately before the opcode. LOCK is group 1 and
	// the segment override is group 2, which is the order.
	if e.Lock {
		b.buf = append(b.buf, 0xf0)
	}
	if b.hasMem && b.mem.HasSeg {
		b.buf = append(b.buf, b.mem.Seg.Override())
	}

	switch e.Family {
	case isa.Legacy:
		if e.Op16 {
			b.buf = append(b.buf, 0x66)
		}
		if e.Pfx != 0 {
			b.buf = append(b.buf, e.Pfx)
		}
		if err := b.emitREX(); err != nil {
			return err
		}
		b.emitMap(e.Map)

	case isa.VEX:
		if err := b.emitVEX(); err != nil {
			return err
		}

	case isa.EVEX:
		if err := b.emitEVEX(); err != nil {
			return err
		}
	}

	op := e.Op
	if e.OpReg {
		op |= b.regNum & 7
	}
	b.buf = append(b.buf, op)

	if e.HasModRM() {
		if e.HasFixedModRM {
			b.buf = append(b.buf, e.FixedModRM)
		} else if err := b.emitModRM(); err != nil {
			return err
		}
	}

	if b.hasImm {
		b.appendInt(b.imm, b.immSize)
	}
	if b.hasRel {
		b.emitRel()
	}
	return nil
}

// emitMap writes the escape bytes for a legacy form's opcode map.
func (b *builder) emitMap(m isa.Map) {
	switch m {
	case isa.Map0F:
		b.buf = append(b.buf, 0x0f)
	case isa.Map0F38:
		b.buf = append(b.buf, 0x0f, 0x38)
	case isa.Map0F3A:
		b.buf = append(b.buf, 0x0f, 0x3a)
	}
}

func (b *builder) emitRel() {
	off := len(b.buf)
	b.appendInt(0, b.relSize)

	switch t := b.rel.(type) {
	case operand.Label:
		b.labels = append(b.labels, LabelHole{
			Offset: off, Size: b.relSize, PCRel: true, Label: t.Name(),
		})
	case operand.SymRef:
		b.refs = append(b.refs, RefHole{
			Offset: off, Size: b.relSize, PCRel: true,
			Sym: t.Sym, Kind: t.Kind, Addend: t.Addend,
		})
	}
}

func (b *builder) appendInt(v int64, size int) {
	for i := 0; i < size; i++ {
		b.buf = append(b.buf, byte(v>>(8*i)))
	}
}

// fits reports whether v fits the immediate field.
//
// signed is the sign-extending case — the 83 group's imm8, PUSH imm8, IMUL's
// — where the field is narrower than the operand it becomes, so an unsigned
// bit pattern is not the value the instruction will compute with. Everywhere
// else the field is the operand's own width and a byte is a byte: 0xff into
// an imm8 is a mask, and refusing it would make masks unwritable.
func fits(v int64, size int, signed bool) bool {
	if size >= 8 {
		return true
	}
	bits := size * 8
	lo := int64(-1) << (bits - 1)
	hi := int64(1)<<(bits-1) - 1
	if v >= lo && v <= hi {
		return true
	}
	return !signed && uint64(v) < uint64(1)<<bits
}

func rangeText(size int, signed bool) string {
	switch size {
	case 1:
		if signed {
			return "-128..127"
		}
		return "-128..255"
	case 2:
		if signed {
			return "-32768..32767"
		}
		return "-32768..65535"
	case 4:
		if signed {
			return "-2147483648..2147483647"
		}
		return "-2147483648..4294967295"
	}
	return "the field's width"
}
