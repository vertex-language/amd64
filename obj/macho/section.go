package macho

import (
	"fmt"
	"strings"

	machocore "github.com/vertex-language/macho"
	machoobj "github.com/vertex-language/macho/obj"

	"github.com/vertex-language/amd64/obj"
)

// Mach-O is the container that disagrees with the other two about naming, and
// this file is mostly that disagreement.
//
// A name is not enough to place a section here. A segment is a load-time
// protection decision, so guessing __DATA for a name this package has never
// seen would produce a file that loads and misbehaves rather than one that
// fails. There are exactly three things a name can be: one of the four kinds
// under its conventional spelling, a DWARF section, or something the caller
// has to place with Options.Sections.

// The four kinds and their conventional pairs. ROData goes to (__TEXT,__const)
// rather than a __DATA section because read-only is a __TEXT property here.
func kindPlacement(k obj.SectionKind) (SegSect, machocore.SecType, machocore.SecAttrs) {
	switch k {
	case obj.Text:
		return SegSect{machocore.SEG_TEXT, machocore.SECT_TEXT}, machocore.S_REGULAR,
			machocore.S_ATTR_PURE_INSTRUCTIONS | machocore.S_ATTR_SOME_INSTRUCTIONS
	case obj.Data:
		return SegSect{machocore.SEG_DATA, machocore.SECT_DATA}, machocore.S_REGULAR, 0
	case obj.ROData:
		return SegSect{machocore.SEG_TEXT, machocore.SECT_CONST}, machocore.S_REGULAR, 0
	case obj.BSS:
		return SegSect{machocore.SEG_DATA, machocore.SECT_BSS}, machocore.S_ZEROFILL, 0
	}
	return SegSect{machocore.SEG_DATA, machocore.SECT_DATA}, machocore.S_REGULAR, 0
}

// dwarfNames is the ELF-to-Mach-O spelling of the DWARF sections.
//
// DWARF section names are the one custom family standardized across all three
// containers, so a caller writing debug info should not have to know that
// this container spells them differently — or that four of them do not fit
// the 16-byte name field and are abbreviated by convention rather than
// truncated.
var dwarfNames = map[string]string{
	".debug_abbrev":      "__debug_abbrev",
	".debug_addr":        "__debug_addr",
	".debug_aranges":     "__debug_aranges",
	".debug_cu_index":    "__debug_cu_index",
	".debug_frame":       "__debug_frame",
	".debug_info":        "__debug_info",
	".debug_line":        "__debug_line",
	".debug_line_str":    "__debug_line_str",
	".debug_loc":         "__debug_loc",
	".debug_loclists":    "__debug_loclists",
	".debug_macinfo":     "__debug_macinfo",
	".debug_macro":       "__debug_macro",
	".debug_names":       "__debug_names",
	".debug_pubnames":    "__debug_pubnames",
	".debug_pubtypes":    "__debug_pubtypes",
	".debug_ranges":      "__debug_ranges",
	".debug_rnglists":    "__debug_rnglists",
	".debug_str":         "__debug_str",
	".debug_str_offsets": "__debug_str_offs",
	".debug_tu_index":    "__debug_tu_index",
	".debug_types":       "__debug_types",
}

// ErrSectionName is a custom section name with no segment to put it in.
//
// It is this package's error and no other's: ELF places a section by name and
// COFF by name and flags, and only here is a segment a separate decision
// nothing in the name implies. The fix is one entry in Options.Sections, and
// the error names the section that needs it.
var ErrSectionName = obj.ErrSectionName

// placement resolves a section to a segment, a name, a type and its
// attributes.
//
// Options.Sections is consulted first, so a caller can override anything
// including a DWARF name — the map is an escape hatch, and an escape hatch
// that only works for names this package already refuses is one that fails
// exactly when the default is wrong.
func placement(s *obj.Section, opt Options) (SegSect, machocore.SecType, machocore.SecAttrs, error) {
	if ss, ok := opt.Sections[s.Name()]; ok {
		typ, attrs := shapeFor(s.Kind())
		return ss, typ, attrs, nil
	}

	// A section still carrying the conventional name for its kind is one of
	// the four. A name someone chose is a name someone meant, and falls
	// through to the custom rules below.
	if s.Name() == s.Kind().String() {
		ss, typ, attrs := kindPlacement(s.Kind())
		return ss, typ, attrs, nil
	}

	if name, ok := dwarfNames[s.Name()]; ok {
		return SegSect{machocore.SEG_DWARF, name}, machocore.S_REGULAR,
			machocore.S_ATTR_DEBUG, nil
	}
	if strings.HasPrefix(s.Name(), ".debug") {
		// A DWARF section this table does not know. Refused rather than
		// mechanically prefixed, because the abbreviations above show the
		// mechanical answer is not always the right one, and a debugger
		// reading the wrong name reports nothing rather than failing.
		return SegSect{}, 0, 0, &obj.Error{
			Sentinel: ErrSectionName,
			Arch:     obj.ArchAMD64,
			Section:  s.Name(),
			Context:  fmt.Sprintf("no __DWARF spelling for %q", s.Name()),
			Notes: []string{
				"give it one with Options.Sections: {\"" + s.Name() + "\": {\"__DWARF\", \"__...\"}}",
			},
		}
	}

	return SegSect{}, 0, 0, &obj.Error{
		Sentinel: ErrSectionName,
		Arch:     obj.ArchAMD64,
		Section:  s.Name(),
		Context:  fmt.Sprintf("no segment for %q", s.Name()),
		Notes: []string{
			"a segment is a load-time protection decision and this container will not guess one",
			"name it with Options.Sections: {\"" + s.Name() + "\": {\"__DATA\", \"__" + strings.TrimPrefix(s.Name(), ".") + "\"}}",
		},
	}
}

// shapeFor is the type and attributes a kind implies, without its name. It is
// what a caller-placed section gets: the placement is theirs, the load-time
// behaviour is still the kind's.
func shapeFor(k obj.SectionKind) (machocore.SecType, machocore.SecAttrs) {
	_, typ, attrs := kindPlacement(k)
	return typ, attrs
}

// newSection creates the section header and nothing else. Contents wait for
// writeContents, for the same reason they do in the COFF writer: this is an
// implicit-addend format and the bytes are not final until the symbol table
// exists.
func newSection(wr *machoobj.Writer, s *obj.Section, opt Options) (*machoobj.SectionBuilder, error) {
	ss, typ, attrs, err := placement(s, opt)
	if err != nil {
		return nil, err
	}
	return wr.Section(machoobj.SectionHeader{
		Segment: ss.Segment,
		Name:    ss.Section,
		Type:    typ,
		Attrs:   attrs,
		Align:   uint32(s.Align()),
	}), nil
}

// writeContents deposits every addend and hands the section its bytes.
//
// A zerofill section is Grow rather than Write: it has a size in memory and
// no bytes on disk. The zeros the builder accumulated in it are exactly that
// size.
func writeContents(wr *machoobj.Writer, b *machoobj.SectionBuilder, s *obj.Section, syms map[string]machoobj.SymRef) error {
	if b.Zerofill() {
		if len(s.Refs()) > 0 {
			return fmt.Errorf("macho: %s is zerofill and has no bytes to hold %d relocation addends",
				s.Name(), len(s.Refs()))
		}
		b.Grow(uint64(s.Size()))
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
