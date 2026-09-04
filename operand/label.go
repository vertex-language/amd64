package operand

import "github.com/vertex-language/amd64/reg"

// Label names an offset in the enclosing section.
//
// It is not a symbol. It is patched at Finalize, leaves no trace in the
// symbol table, and cannot cross a section boundary — a branch that leaves
// the section is a SymRef and survives into Refs(). That split is the one
// the helper names carry: JmpLabel resolves here, JmpRef does not.
type Label struct {
	reg.Seal
	name string
}

// NewLabel names a label.
func NewLabel(name string) Label { return Label{name: name} }

// Name is the label's spelling. Nothing here mangles it and nothing invents
// a .L prefix.
func (l Label) Name() string { return l.name }

func (l Label) String() string {
	if l.name == "" {
		return "<unnamed label>"
	}
	return l.name
}

// There is no LabelDiff operand type. A label difference is a data-side
// construct — a jump table of 32-bit offsets from the table's own base — and
// it is placed by Section.LabelDiff, which names both labels directly. No
// instruction takes one, so an operand that could satisfy no class would be a
// type a caller could hold and never use.
