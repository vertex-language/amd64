package asm

import (
	"fmt"
	"sort"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/asm/gas"
)

// emitter adapts a Module to gas.Emitter. It is arm64/asm's twin, and differs
// only where the two builder APIs do.
type emitter struct {
	m   *amd64.Module
	sec *amd64.Section

	defined  map[string]bool
	referred map[string]bool
}

func newEmitter(m *amd64.Module) *emitter {
	return &emitter{m: m, defined: map[string]bool{}, referred: map[string]bool{}}
}

func (e *emitter) refer(name string) { e.referred[name] = true }

// declareExterns declares everything referred to and never defined, which is
// what makes `call puts` work without an `.extern puts` nobody writes.
func (e *emitter) declareExterns() {
	names := make([]string, 0, len(e.referred))
	for n := range e.referred {
		if !e.defined[n] {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for _, n := range names {
		if e.m.Defines(n) {
			// Defined by the module this fragment is being assembled
			// into, rather than by the fragment. An implicit extern is a
			// guess about a name nothing defines, and this name is
			// defined — a block label an asm goto branches to, or another
			// function in the same object.
			continue
		}
		e.m.Extern(n)
	}
}

func (e *emitter) Section(name string, k gas.SectionKind) error {
	e.sec = e.m.SectionNamed(name, sectionKind(k))
	return e.m.Err()
}

func sectionKind(k gas.SectionKind) amd64.SectionKind {
	switch k {
	case gas.Text:
		return amd64.Text
	case gas.ROData:
		return amd64.ROData
	case gas.BSS:
		return amd64.BSS
	}
	return amd64.Data
}

func (e *emitter) Offset() int { return e.sec.Offset() }

// Label names the current offset, and every label becomes a symbol. The
// reasoning is arm64/asm's: a parser cannot know which labels are
// address-taken when it emits them, so it promotes all of them and lets the
// linker drop what it does not need.
func (e *emitter) Label(name string, a gas.Attrs) error {
	e.defined[name] = true
	attrs := []any{binding(a.Binding)}
	switch a.Type {
	case gas.TypeFunc:
		attrs = append(attrs, amd64.Func)
	case gas.TypeObject:
		attrs = append(attrs, amd64.ObjectSym)
	}
	switch a.Vis {
	case gas.VisHidden:
		attrs = append(attrs, amd64.Hidden)
	case gas.VisProtected:
		attrs = append(attrs, amd64.Protected)
	}
	e.sec.Label(name, attrs...)
	return e.m.Err()
}

func binding(b gas.Binding) amd64.Binding {
	switch b {
	case gas.BindGlobal:
		return amd64.Global
	case gas.BindWeak:
		return amd64.Weak
	}
	return amd64.Local
}

func (e *emitter) EndLabel(name string) error {
	e.sec.EndLabel(name)
	return e.m.Err()
}

func (e *emitter) Extern(name string) {
	e.defined[name] = true
	e.m.Extern(name)
}

func (e *emitter) Alias(name, of string) {
	e.defined[name] = true
	e.refer(of)
	e.m.Alias(name, of)
}

func (e *emitter) Data(b []byte) { e.sec.Data(b) }

func (e *emitter) Int(v int64, size int) error {
	switch size {
	case 1:
		e.sec.Byte(byte(v))
	case 2:
		e.sec.Byte(byte(v))
		e.sec.Byte(byte(v >> 8))
	case 4:
		e.sec.Long(uint32(v))
	case 8:
		e.sec.Quad(uint64(v))
	default:
		return fmt.Errorf("no %d-byte integer directive", size)
	}
	return e.m.Err()
}

func (e *emitter) Zero(n int) { e.sec.Zero(n) }

func (e *emitter) Align(n int) error {
	e.sec.Align(n)
	return e.m.Err()
}

func (e *emitter) SymRef(sym string, addend int64, size int) error {
	var kind obj.RefKind
	switch size {
	case 8:
		kind = operand.RefAbs64
	case 4:
		kind = operand.RefAbs32
	case 2:
		kind = operand.RefAbs16
	case 1:
		kind = operand.RefAbs8
	default:
		return fmt.Errorf("a %d-byte symbol reference has no relocation", size)
	}
	e.refer(sym)
	e.sec.Ref(operand.Ref(sym, kind).Add(addend))
	return e.m.Err()
}

func (e *emitter) SymDiff(to, from string, addend int64, size int) error {
	if size != 4 {
		return fmt.Errorf("a label difference is four bytes, not %d", size)
	}
	if addend != 0 {
		return fmt.Errorf("%s-%s%+d: an addend on a label difference is not supported yet",
			to, from, addend)
	}
	e.refer(to)
	e.refer(from)
	e.sec.LabelDiff(to, from)
	return e.m.Err()
}
