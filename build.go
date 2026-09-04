package amd64

import "github.com/vertex-language/amd64/obj"

// build turns the finished module into the artifact.
//
// It runs after every Finalize step has passed: labels patched, sizes closed,
// aliases and visibility resolved, every reference verified. So there is
// nothing to check here and nothing that can fail — this is a translation,
// and the only reason it is a separate function is that the builder's storage
// and the artifact's are different shapes.
//
// The copy that makes the object inert happens inside obj.New, not here. What
// this hands over aliases the builder's slices, and obj.New duplicates them;
// putting the copy there rather than here is what makes "the artifact is
// immutable" a property of the artifact rather than a promise this function
// makes.
func (m *Module) build() *obj.Object {
	secs := make([]obj.SectionData, 0, len(m.sections))
	for _, s := range m.sections {
		secs = append(secs, obj.SectionData{
			Name:  s.name,
			Kind:  s.kind,
			Align: s.align,
			Bytes: s.buf,
			Refs:  s.refs,
		})
	}

	syms := make([]obj.Symbol, 0, len(m.symbols))
	for _, sym := range m.symbols {
		syms = append(syms, obj.Symbol{
			Name: sym.name,
			// sec is -1 for an undefined symbol and stays -1. It is the
			// same sentinel on both sides, so Defined() means here exactly
			// what defined meant in the builder.
			Section:    sym.sec,
			Offset:     sym.off,
			Size:       sym.size,
			Binding:    sym.binding,
			Type:       sym.typ,
			Visibility: sym.vis,
		})
	}

	return obj.New(obj.ArchAMD64, secs, syms)
}
