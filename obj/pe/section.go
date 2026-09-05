package pe

import (
	"fmt"

	pecore "github.com/vertex-language/pe"
	"github.com/vertex-language/pe/coff"

	"github.com/vertex-language/amd64/obj"
)

// Section kinds map onto COFF's content and memory flags. One of the four
// also changes name.
//
// ELF and COFF disagree about what read-only initialized data is called:
// .rodata on one side, .rdata on the other, and link.exe's default merge
// rules are written against .rdata. The rename applies only to a section
// still carrying the conventional name for its kind — Section(ROData) — and
// never to one a caller named itself with SectionNamed, because a name
// someone chose is a name someone meant.
func coffName(s *obj.Section) string {
	if s.Name() != s.Kind().String() {
		return s.Name()
	}
	if s.Kind() == obj.ROData || s.Kind() == obj.RelROData {
		return ".rdata"
	}
	return s.Name()
}

func sectionShape(k obj.SectionKind) (pecore.SecKind, pecore.SecProt) {
	switch k {
	case obj.Text:
		return pecore.SecCode, pecore.SecExecute | pecore.SecRead
	case obj.Data:
		return pecore.SecInitData, pecore.SecRead | pecore.SecWrite
	case obj.ROData:
		return pecore.SecInitData, pecore.SecRead
	case obj.RelROData:
		// COFF has no relro and needs none: base relocations are
		// applied before page protections are set, so read-only is
		// both correct and what MSVC emits for the same declaration.
		return pecore.SecInitData, pecore.SecRead
	case obj.BSS:
		return pecore.SecUninitData, pecore.SecRead | pecore.SecWrite
	}
	return pecore.SecInitData, pecore.SecRead
}

// newSection creates the section header and nothing else.
//
// Contents are deliberately not written here. COFF is implicit-addend, so a
// section's bytes are not final until every reference in it has deposited its
// addend — and a reference cannot be resolved before the symbol table exists,
// which cannot exist before every section does. Two passes, in that order,
// and writeContents is the second.
func newSection(wr *coff.Writer, s *obj.Section) *coff.SectionBuilder {
	kind, prot := sectionShape(s.Kind())
	return wr.Section(coff.SectionHeader{
		Name:  coffName(s),
		Kind:  kind,
		Prot:  prot,
		Align: s.Align(),
	})
}

// writeContents deposits every addend and hands the section its bytes.
//
// A BSS section is Reserve rather than Write: it has a size and no bytes, and
// the underlying writer refuses contents in a section that by definition has
// none. The zeros the builder accumulated in it are exactly that size, which
// is why the count comes from the section and not from a separate field.
func writeContents(wr *coff.Writer, b *coff.SectionBuilder, s *obj.Section, syms map[string]*coff.SymbolRef) error {
	if s.Kind() == obj.BSS {
		if len(s.Refs()) > 0 {
			return fmt.Errorf("pe: %s is uninitialized data and has no bytes to hold %d relocation addends",
				s.Name(), len(s.Refs()))
		}
		b.Reserve(uint32(s.Size()))
		return wr.Err()
	}

	// Bytes() handed back a copy, and every addend goes into that copy. The
	// object is untouched, which is what lets three writers run over one
	// object without any of them seeing another's work.
	content := s.Bytes()
	if err := writeRelocs(wr, b, s, content, syms); err != nil {
		return err
	}
	if _, err := b.Write(content); err != nil {
		return err
	}
	return wr.Err()
}
