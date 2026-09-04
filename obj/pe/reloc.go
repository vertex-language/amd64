package pe

import (
	"fmt"

	pecore "github.com/vertex-language/pe"
	"github.com/vertex-language/pe/coff"

	"github.com/vertex-language/amd64/obj"
)

// The intent→IMAGE_REL_AMD64_* table, and the kinds COFF has no answer for.
//
// Six rows, and the absences are the interesting part:
//
//   - RefPLT32 is refused. There is no procedure linkage table. An imported
//     call reaches its target through a thunk the linker synthesizes from the
//     import library, and the relocation on the call site is an ordinary
//     REL32 to the __imp_ symbol — which is a different symbol, so mapping
//     RefPLT32 onto REL32 here would relocate against a name that does not
//     mean what the lowering meant.
//   - The GOT kinds are refused for the same reason: there is no GOT, and
//     neither the GOTOFF/GOTPC dance nor the relaxable GOTPCRELX ladder has a
//     COFF counterpart.
//   - Every ELF TLS model is refused. COFF addresses thread-local data
//     through SECREL against the .tls section and the TLS directory, which is
//     not any of the seven models obj declares. RefTLV is Mach-O's and is
//     refused for the third reason again: it is another container's idea.
//   - RefSize32 and RefSize64 are refused. A symbol's extent is not a value
//     COFF can relocate against, because a COFF symbol has no size field to
//     take it from.
//   - The 8- and 16-bit widths are refused. x64 COFF has no ADDR16, no
//     ADDR8 and no REL16: its narrowest field is the 16-bit SECTION index,
//     which is a section ordinal and not an address, so there is nothing to
//     map a narrow absolute or displacement onto.
//   - RefPC64 is refused. COFF's PC-relative field is 32 bits.
//
// Each refusal is ErrRefKind at emission rather than at construction, because
// the same object is legal for the container the kind belongs to.
type relocForm struct {
	typ   pecore.RelocAMD64
	pcrel bool
}

var relocTypes = map[obj.RefKind]relocForm{
	// ADDR64 is the 64-bit virtual address. ADDR32 is the 32-bit one, which
	// needs the image based below 2 GB — the Windows default. RefAbs32 and
	// RefAbs32S share it: COFF has one type and the two kinds differ only in
	// whether a linker range-checks the result as signed or unsigned, a
	// distinction with no expression here and no consequence under 2 GB.
	obj.RefAbs64:  {pecore.IMAGE_REL_AMD64_ADDR64, false},
	obj.RefAbs32:  {pecore.IMAGE_REL_AMD64_ADDR32, false},
	obj.RefAbs32S: {pecore.IMAGE_REL_AMD64_ADDR32, false},

	// ADDR32NB is the 32-bit RVA — "no base" — which is what an unwind table
	// and a debug directory are built out of.
	obj.RefImageRel32: {pecore.IMAGE_REL_AMD64_ADDR32NB, false},

	// REL32 is the head of a ladder; pickRel32 chooses the rung.
	obj.RefPC32: {pecore.IMAGE_REL_AMD64_REL32, true},

	// PLT32 is REL32 here, and not by approximation: this platform has no
	// procedure linkage table, and a call to a symbol that turns out to
	// live in a DLL is bound by the linker to a thunk it synthesizes. The
	// two kinds name the same relocation, which is why every COFF
	// toolchain emits a call the same way whether or not the callee is
	// imported.
	obj.RefPLT32: {pecore.IMAGE_REL_AMD64_REL32, true},

	obj.RefSecRel32: {pecore.IMAGE_REL_AMD64_SECREL, false},
	obj.RefSecIdx:   {pecore.IMAGE_REL_AMD64_SECTION, false},
}

// writeRelocs translates one section's holes into COFF relocation records and
// deposits their addends into content.
//
// COFF is an implicit-addend format, so the addend goes in the section bytes —
// but COFF and the Reference identity measure a PC-relative field from
// different places, and that difference is this function's whole reason for
// existing.
//
// The identity every consumer of a Reference relies on is
//
//	value = target - (section offset of the field) + Adjust + Addend
//
// while IMAGE_REL_AMD64_REL32_n resolves to
//
//	target - (address of the byte after the field, plus n) + stored
//
// so the bytes after the field have to be paid for:
//
//	stored = Addend + Adjust + 4 + n
//
// and n is free. That freedom is the ladder: choosing n = -(Adjust+4) makes
// stored exactly Addend, which is the form every Microsoft tool emits and
// expects, and it makes the rung state the truth about the instruction — how
// many bytes trail the displacement. A disp32 with nothing after it is REL32;
// one with a trailing imm32 is REL32_4.
func writeRelocs(wr *coff.Writer, b *coff.SectionBuilder, s *obj.Section, content []byte, syms map[string]*coff.SymbolRef) error {
	for _, r := range s.Refs() {
		form, ok := relocTypes[r.Kind]
		if !ok {
			return refKindError(s, r, "COFF has no relocation for this kind")
		}
		if want := r.Kind.Size(); want != r.Size {
			return fmt.Errorf("pe: %s+%#x: %s reference to %q is a %d-byte field; %v writes %d",
				s.Name(), r.Offset, r.Kind, r.Sym, r.Size, r.Kind, want)
		}

		sym, ok := syms[r.Sym]
		if !ok {
			return fmt.Errorf("pe: %s+%#x: reference to %q, which is not in the symbol table",
				s.Name(), r.Offset, r.Sym)
		}

		typ, stored := form.typ, r.Addend
		if form.pcrel {
			rung, extra, err := pickRel32(r.Adjust)
			if err != nil {
				return refKindError(s, r, err.Error())
			}
			typ, stored = rung, r.Addend+extra
		}

		if err := deposit(content, r.Offset, r.Size, stored); err != nil {
			return fmt.Errorf("pe: %s+%#x: %w", s.Name(), r.Offset, err)
		}

		wr.Reloc(b, coff.RelocSpec{
			Address: uint32(r.Offset),
			Sym:     sym,
			Type:    uint16(typ),
		})
	}
	return wr.Err()
}

// pickRel32 chooses a rung of the REL32 ladder and returns the leftover that
// has to be folded into the stored addend.
//
// The rung is the number of bytes trailing the field, which is -(Adjust+4).
// Two ends of that range are special:
//
//   - Above the top rung — Adjust below -9 — is ErrRefKind. REL32_5 is the
//     last one the format declares. Nothing this encoder emits reaches it;
//     the check is here so that a future form which does fails at the writer
//     rather than producing a REL32 off by six.
//   - Below the bottom rung — Adjust above -4, which is what a data-side Ref
//     of a PC-relative kind carries, since no encoder sized its field — there
//     is no rung to name. REL32 is used and the difference is folded into the
//     stored addend, which is arithmetically the same relocation and the only
//     way to express one at all.
func pickRel32(adjust int64) (pecore.RelocAMD64, int64, error) {
	n := -(adjust + 4)
	switch {
	case n > 5:
		return 0, 0, fmt.Errorf("Adjust %d needs REL32_%d; the ladder ends at REL32_5", adjust, n)
	case n < 0:
		// stored = Addend + Adjust + 4 + 0
		return pecore.IMAGE_REL_AMD64_REL32, adjust + 4, nil
	}
	// stored = Addend + Adjust + 4 + n, and n was chosen to make that Addend.
	return pecore.IMAGE_REL_AMD64_REL32 + pecore.RelocAMD64(n), 0, nil
}

// deposit adds an implicit addend into the section's bytes.
//
// It adds rather than overwrites, because "the implicit addend is the value
// in the field" is what an implicit-addend format means. The encoder
// zero-fills every hole it leaves, so the two are the same today; they would
// not be if a caller ever built a field by hand.
func deposit(content []byte, off, size int, addend int64) error {
	if off < 0 || size < 0 || off+size > len(content) {
		return fmt.Errorf("a %d-byte field at %#x does not fit %d bytes of section",
			size, off, len(content))
	}

	var cur uint64
	for i := 0; i < size; i++ {
		cur |= uint64(content[off+i]) << (8 * i)
	}
	// Sign-extend what is there so a negative implicit addend already in
	// place stays negative through the addition.
	switch size {
	case 1:
		cur = uint64(int64(int8(cur)))
	case 2:
		cur = uint64(int64(int16(cur)))
	case 4:
		cur = uint64(int64(int32(cur)))
	}

	v := uint64(int64(cur) + addend)
	for i := 0; i < size; i++ {
		content[off+i] = byte(v >> (8 * i))
	}
	return nil
}

// refKindError is the one diagnostic this package raises for a kind COFF
// cannot express. It names the kind, the symbol and the section offset,
// because a lowering that produced one needs all three to find it.
func refKindError(s *obj.Section, r obj.Reference, note string) error {
	return &obj.Error{
		Sentinel: obj.ErrRefKind,
		Arch:     obj.ArchAMD64,
		Section:  s.Name(),
		Offset:   r.Offset,
		Context:  fmt.Sprintf("%s reference to %q", r.Kind, r.Sym),
		Notes:    []string{note, "the object is still legal for the container this kind belongs to"},
	}
}
