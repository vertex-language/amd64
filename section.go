package amd64

import (
	"github.com/vertex-language/amd64/internal/encode"
	"github.com/vertex-language/amd64/internal/isa"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
)

// Section is one section under construction. The typed helpers are methods
// on this type, which is what makes a width mismatch a compile error;
// an interface or a generic builder would erase exactly the checking the
// surface exists for.
type Section struct {
	m *Module

	kind  SectionKind
	name  string
	index int
	align int

	buf  []byte
	refs []obj.Reference

	labels map[string]int
	holes  []labelHole

	// dead marks the spent handle. Every call on it returns immediately.
	dead bool
}

// labelHole is a field Finalize fills from a same-section label. It leaves
// no relocation and a linker never sees it.
type labelHole struct {
	off    int
	size   int
	pcrel  bool
	adjust int64
	to     string
	from   string // set for a LabelDiff; the subtrahend
	form   string // for the diagnostic when the displacement does not fit
}

// Module is the module this section belongs to, for a caller holding a
// section and needing the module-level operations that go with it.
func (s *Section) Module() *Module { return s.m }

func (s *Section) Kind() SectionKind { return s.kind }
func (s *Section) Name() string      { return s.name }
func (s *Section) Index() int        { return s.index }

// Offset is the current end of the section: the offset the next byte will
// land at, and the value a Label placed now would name.
//
// It is exported because vtables, jump tables and literal pools are yours
// to build, and building them requires knowing where you are.
func (s *Section) Offset() int { return len(s.buf) }

// ok reports whether the section should do anything. It is the single
// gate every method passes through, which is what makes "sticky and
// first-wins" one rule rather than a rule repeated forty times.
func (s *Section) ok() bool {
	return !s.dead && s.m.err == nil && !s.m.done
}

// ---- instructions ---------------------------------------------------------

// inst is what every typed helper funnels into.
//
// A helper pins its form — exactly the encoding you named — and the
// diagnostics follow from that. The helper has already checked operand
// kinds by being a Go signature; what is left is the gate, the encoder, and
// positioning whatever comes back.
func (s *Section) inst(f *isa.Form, ops ...operand.Operand) {
	if !s.ok() {
		return
	}

	if !f.Permitted(s.m.features) {
		missing := f.Missing(s.m.features)
		notes := make([]string, 0, len(missing)+1)
		notes = append(notes, "the module's feature set is "+s.m.features.String())
		for _, g := range missing {
			notes = append(notes, "add "+g.String()+" to emit "+f.String())
		}
		s.m.fail(s.errorAt(obj.ErrFeature,
			f.String()+" needs a feature the module does not have", notes...))
		return
	}

	in, err := encode.Encode(f, ops)
	if err != nil {
		s.m.fail(s.lift(err))
		return
	}

	base := len(s.buf)
	s.buf = append(s.buf, in.Bytes...)

	for _, h := range in.Refs {
		s.refs = append(s.refs, obj.Reference{
			Offset: base + h.Offset,
			Size:   h.Size,
			PCRel:  h.PCRel,
			Adjust: h.Adjust,
			Sym:    h.Sym,
			Kind:   h.Kind,
			Addend: h.Addend,
		})
	}
	for _, h := range in.Labels {
		s.holes = append(s.holes, labelHole{
			off: base + h.Offset, size: h.Size,
			pcrel: h.PCRel, adjust: h.Adjust,
			to: h.Label, form: f.String(),
		})
	}
}

// ---- labels and symbols ---------------------------------------------------

// Label names an offset in this section.
//
// A bare Label is not a symbol. It gets patched at Finalize and leaves no
// trace in the symbol table. Any attribute — a Binding, a SymbolType, a
// Visibility — promotes it into Symbols().
//
// That rule is why nothing here mangles and nothing invents .L prefixes: a
// block label is a bare label, a local symbol you want emitted is Local,
// and whether locals reach the file at all is the writer's StripLocals.
func (s *Section) Label(name string, attrs ...any) {
	if !s.ok() {
		return
	}
	if prev, dup := s.labels[name]; dup {
		s.m.fail(s.errorAt(obj.ErrDuplicate,
			"label "+name+" is already defined in "+s.name,
			"the first definition is at "+s.name+"+"+hex(prev)))
		return
	}
	s.labels[name] = len(s.buf)

	if len(attrs) == 0 {
		return
	}
	s.m.define(s, name, len(s.buf), attrs...)
}

// EndLabel closes a symbol's size range at the current offset.
//
// Size closes here if you call it, at the next symbol in the same section
// if you do not, and at the section end otherwise. The next-symbol fallback
// is a guess, so prefer EndLabel for anything you care about.
func (s *Section) EndLabel(name string) {
	if !s.ok() {
		return
	}
	i, ok := s.m.symAt[name]
	if !ok || !s.m.symbols[i].defined {
		s.m.fail(s.errorAt(obj.ErrUndefined,
			"EndLabel("+name+") names no symbol defined in this module"))
		return
	}
	sym := &s.m.symbols[i]
	sym.size = len(s.buf) - sym.off
	sym.sizeClosed = true
}

// ---- alignment and data ---------------------------------------------------

// Align pads to a power-of-two boundary: a code section with the multi-byte
// nop sequences, a data section with zeros.
//
// Unlike i386 there is no gate on this. 0F 1F is in the x86-64 baseline, so
// every module this package can build can execute the long nops.
//
// The largest n a section sees also becomes its Align(), which is what the
// object writer stamps on the section header.
func (s *Section) Align(n int) {
	if !s.ok() {
		return
	}
	if n <= 0 || n&(n-1) != 0 {
		s.m.fail(s.errorAt(obj.ErrAlign,
			"alignment must be a power of two",
			"got "+itoa(n)))
		return
	}
	if n > s.align {
		s.align = n
	}
	pad := (n - len(s.buf)%n) % n
	if pad == 0 {
		return
	}
	if s.kind == Text {
		s.buf = append(s.buf, encode.Nops(pad)...)
		return
	}
	s.buf = append(s.buf, make([]byte, pad)...)
}

// The data builders. The raw-bytes one is Data because Bytes() is the read
// side of the contract, and there is no Word, because a name whose width
// depends on the package you are in is not a name.

func (s *Section) Byte(v byte) {
	if s.ok() {
		s.buf = append(s.buf, v)
	}
}

func (s *Section) Long(v uint32) { s.le(uint64(v), 4) }
func (s *Section) Quad(v uint64) { s.le(v, 8) }

func (s *Section) le(v uint64, n int) {
	if !s.ok() {
		return
	}
	for i := 0; i < n; i++ {
		s.buf = append(s.buf, byte(v>>(8*i)))
	}
}

func (s *Section) Ascii(str string) {
	if s.ok() {
		s.buf = append(s.buf, str...)
	}
}

func (s *Section) Asciz(str string) {
	if s.ok() {
		s.buf = append(s.buf, str...)
		s.buf = append(s.buf, 0)
	}
}

func (s *Section) Zero(n int) {
	if s.ok() && n > 0 {
		s.buf = append(s.buf, make([]byte, n)...)
	}
}

func (s *Section) Data(b []byte) {
	if s.ok() {
		s.buf = append(s.buf, b...)
	}
}

// Ref places a hole and a relocation.
//
// The width is the kind's, because a kind knows how wide the field it fills
// is — a RefAbs64 is eight bytes wherever it appears — and there is no
// encoder in this path to have sized it. An undeclared kind falls back to
// eight.
//
// This package refuses to build your vtables, jump tables and literal
// pools, so it must not make them unbuildable. Ref, LabelRef and LabelDiff
// are the data-side twins of what an instruction operand does.
func (s *Section) Ref(r SymRef) {
	if !s.ok() {
		return
	}
	size := 8
	if sz := r.Kind.Size(); sz != 0 {
		size = sz
	}
	s.refs = append(s.refs, obj.Reference{
		Offset: len(s.buf), Size: size,
		Sym: r.Sym, Kind: r.Kind, Addend: r.Addend,
	})
	s.buf = append(s.buf, make([]byte, size)...)
}

// LabelRef places an eight-byte hole patched at Finalize from a
// same-section label. No relocation.
func (s *Section) LabelRef(name string) {
	if !s.ok() {
		return
	}
	s.holes = append(s.holes, labelHole{
		off: len(s.buf), size: 8, to: name,
		form: "an 8-byte label reference",
	})
	s.buf = append(s.buf, make([]byte, 8)...)
}

// LabelDiff places a four-byte hole holding to - from, both labels in this
// section, resolved exactly at Finalize.
//
// This is the one that earns its place on this architecture. A jump table
// of 32-bit offsets from the table's own base is how you avoid eight bytes
// per entry and a relocation per entry, and the subtraction leaves nothing
// for a linker to do.
func (s *Section) LabelDiff(to, from string) {
	if !s.ok() {
		return
	}
	s.holes = append(s.holes, labelHole{
		off: len(s.buf), size: 4, to: to, from: from,
		form: "a 4-byte label difference",
	})
	s.buf = append(s.buf, make([]byte, 4)...)
}
