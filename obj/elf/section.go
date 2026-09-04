package elf

import (
	"strings"

	elfcore "github.com/vertex-language/elf"
	elfobj "github.com/vertex-language/elf/obj"

	"github.com/vertex-language/amd64/obj"
)

// Section kinds map onto a type and a flag word. The kind is a load-time
// property, which is exactly what SHF_ALLOC, SHF_WRITE and SHF_EXECINSTR
// describe, so this is a translation and not a decision.
//
// One rule is not derivable from the kind, and it is the one heuristic in
// this package: a section whose name marks it as debug or comment data is not
// allocated at run time, whatever kind it was created with. DWARF is built
// with SectionNamed(".debug_line", Data) because Data is the load-time
// behaviour of initialized bytes — but SHF_ALLOC on .debug_line puts the
// debug info in the image, which is wrong in a way nothing downstream
// reports. The rule is stated here rather than pushed onto the caller because
// there is no name under it that a caller would want allocated.
func sectionShape(s *obj.Section) (elfcore.SHType, uint64) {
	if s.Kind() == obj.BSS {
		if nonAlloc(s.Name()) {
			// A non-allocated NOBITS section occupies neither file nor
			// memory, which is a section that does not exist. PROGBITS with
			// no flags at least round-trips.
			return elfcore.SHT_PROGBITS, 0
		}
		return elfcore.SHT_NOBITS, elfcore.SHF_ALLOC | elfcore.SHF_WRITE
	}

	if nonAlloc(s.Name()) {
		return elfcore.SHT_PROGBITS, 0
	}

	switch s.Kind() {
	case obj.Text:
		return elfcore.SHT_PROGBITS, elfcore.SHF_ALLOC | elfcore.SHF_EXECINSTR
	case obj.Data:
		return elfcore.SHT_PROGBITS, elfcore.SHF_ALLOC | elfcore.SHF_WRITE
	case obj.ROData:
		return elfcore.SHT_PROGBITS, elfcore.SHF_ALLOC
	}
	return elfcore.SHT_PROGBITS, elfcore.SHF_ALLOC
}

// nonAlloc names the sections ELF itself defines as not present at run time.
// The list is prefixes because the families are open — .debug_line,
// .debug_info, .debug_str — and closed at the top: nothing outside these
// names is guessed about.
func nonAlloc(name string) bool {
	switch {
	case strings.HasPrefix(name, ".debug"):
		return true
	case name == ".comment":
		return true
	}
	return false
}

// writeSection creates the section and fills it.
//
// A BSS section arrives with a run of zero bytes in it, because the builder's
// Zero appends real bytes and Offset has to keep meaning what it says. Here
// that run becomes sh_size and no file content, which is the whole point of
// SHT_NOBITS.
func writeSection(wr *elfobj.Writer, s *obj.Section) (*elfobj.SectionBuilder, error) {
	typ, flags := sectionShape(s)

	hdr := elfobj.SectionHeader{
		Name:      s.Name(),
		Type:      typ,
		Flags:     flags,
		Addralign: uint64(s.Align()),
	}
	if typ == elfcore.SHT_NOBITS {
		hdr.Size = uint64(s.Size())
	}

	b := wr.Section(hdr)
	if typ == elfcore.SHT_NOBITS {
		return b, nil
	}
	// The bytes go through untouched. On a REL architecture this is where an
	// addend would be folded in; x86-64 is RELA and the addend rides in the
	// entry, so the object's own copy is what lands in the file.
	if _, err := b.Write(s.Bytes()); err != nil {
		return nil, err
	}
	return b, nil
}
