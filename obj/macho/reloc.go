package macho

import (
	"fmt"

	machocore "github.com/vertex-language/macho"
	machoobj "github.com/vertex-language/macho/obj"

	"github.com/vertex-language/amd64/obj"
)

// The intent→X86_64_RELOC_* table, and the kinds Mach-O has no answer for.
//
// The mappings that carry a judgement:
//
//   - RefPLT32 is X86_64_RELOC_BRANCH, and unlike COFF this is a faithful
//     mapping: the target is the symbol the lowering named, and ld64
//     synthesizes a stub if it needs one. Nothing is relocated against a
//     different name.
//   - RefGOTPCREL is GOT and the two relaxable kinds are GOT_LOAD, which is
//     ld64's equivalent claim — the instruction is a movq load of a GOT entry
//     and may be rewritten to a leaq. The REX/no-REX split ELF needs does not
//     exist here, so both relaxable kinds land on the one type.
//   - RefTLV is the only thread-local kind accepted. Mach-O reaches a
//     thread-local through a descriptor and a call, not a relocation model,
//     so the seven ELF models are refused and the descriptor sequence is an
//     instruction sequence a lowering emits.
//   - RefSize32, RefSize64, RefImageRel32, RefSecRel32 and RefSecIdx are
//     refused. The first two are an ELF idea and the last three a COFF one.
//   - The 8- and 16-bit widths and RefPC64 are refused: every x86-64 Mach-O
//     relocation writes four bytes, or eight for UNSIGNED.
type relocForm struct {
	typ    machocore.X86_64Reloc
	pcrel  bool
	ladder bool // selects a SIGNED rung from Adjust
}

var relocTypes = map[obj.RefKind]relocForm{
	// UNSIGNED is the absolute address, at either width. RefAbs32S shares it
	// with RefAbs32: Mach-O has one absolute type and the two kinds differ
	// only in how a linker range-checks the result, which is a distinction
	// this container does not draw.
	obj.RefAbs64:  {machocore.X86_64_RELOC_UNSIGNED, false, false},
	obj.RefAbs32:  {machocore.X86_64_RELOC_UNSIGNED, false, false},
	obj.RefAbs32S: {machocore.X86_64_RELOC_UNSIGNED, false, false},

	obj.RefPC32: {machocore.X86_64_RELOC_SIGNED, true, true},

	obj.RefPLT32: {machocore.X86_64_RELOC_BRANCH, true, false},

	obj.RefGOTPCREL:     {machocore.X86_64_RELOC_GOT, true, false},
	obj.RefGOTPCRELX:    {machocore.X86_64_RELOC_GOT_LOAD, true, false},
	obj.RefRexGOTPCRELX: {machocore.X86_64_RELOC_GOT_LOAD, true, false},

	obj.RefTLV: {machocore.X86_64_RELOC_TLV, true, false},
}

// signedLadder maps trailing-byte counts onto relocation types.
//
// The gap at 3 is the format's, not this package's. ld64 declares SIGNED,
// SIGNED_1, SIGNED_2 and SIGNED_4 and no SIGNED_3, so a displacement followed
// by exactly three bytes is the one field position this container cannot
// name. It is refused rather than folded into a plain SIGNED with a biased
// addend, because ld64 reads that bias as an offset into the target atom and
// would attribute the reference three bytes into the wrong place.
var signedLadder = map[int64]machocore.X86_64Reloc{
	0: machocore.X86_64_RELOC_SIGNED,
	1: machocore.X86_64_RELOC_SIGNED_1,
	2: machocore.X86_64_RELOC_SIGNED_2,
	4: machocore.X86_64_RELOC_SIGNED_4,
}

// writeRelocs translates one section's holes into relocation entries and
// deposits their addends into content.
//
// Mach-O is an implicit-addend format, and for a PC-relative entry the value
// deposited is
//
//	stored = Addend + Adjust + 4
//
// which falls straight out of the two identities. A Reference means
//
//	value = target - (section offset of the field) + Adjust + Addend
//
// and ld64 resolving SIGNED_n reads the field as an addend biased by n, then
// computes target + (stored + n) - (field + 4 + n) — the n cancels, and what
// is left fixes stored. Choosing n as the number of trailing bytes,
// -(Adjust+4), makes stored equal Addend - n, which is exactly what clang
// emits for the same instruction.
func writeRelocs(wr *machoobj.Writer, b *machoobj.SectionBuilder, s *obj.Section, content []byte, syms map[string]machoobj.SymRef) error {
	for _, r := range s.Refs() {
		form, ok := relocTypes[r.Kind]
		if !ok {
			return refKindError(s, r, "Mach-O has no relocation for this kind")
		}
		if want := r.Kind.Size(); want != r.Size {
			return fmt.Errorf("macho: %s+%#x: %s reference to %q is a %d-byte field; %v writes %d",
				s.Name(), r.Offset, r.Kind, r.Sym, r.Size, r.Kind, want)
		}
		length, ok := lengthOf(r.Size)
		if !ok {
			return refKindError(s, r, fmt.Sprintf("a %d-byte field has no r_length", r.Size))
		}

		sym, ok := syms[r.Sym]
		if !ok {
			return fmt.Errorf("macho: %s+%#x: reference to %q, which is not in the symbol table",
				s.Name(), r.Offset, r.Sym)
		}

		typ, stored := form.typ, r.Addend
		if form.pcrel {
			// n is the number of bytes trailing the field. Every accepted
			// type here fixes it: the SIGNED ladder names it outright, and
			// BRANCH, GOT, GOT_LOAD and TLV all mean "nothing trails", so a
			// reference carrying anything else is a field position this
			// container cannot express.
			n := -(r.Adjust + 4)
			if form.ladder {
				rung, ok := signedLadder[n]
				if !ok {
					return refKindError(s, r, fmt.Sprintf(
						"Adjust %d puts %d bytes after the field; Mach-O has SIGNED, SIGNED_1, SIGNED_2 and SIGNED_4 and no other rung",
						r.Adjust, n))
				}
				typ = rung
			} else if n != 0 {
				return refKindError(s, r, fmt.Sprintf(
					"%v is only encodable with nothing after the field, and Adjust %d puts %d bytes there",
					typ, r.Adjust, n))
			}
			stored = r.Addend + r.Adjust + 4
		}

		if err := deposit(content, r.Offset, r.Size, stored); err != nil {
			return fmt.Errorf("macho: %s+%#x: %w", s.Name(), r.Offset, err)
		}

		wr.Reloc(b, machoobj.RelocSpec{
			Address: uint64(r.Offset),
			Sym:     sym,
			Type:    uint8(typ),
			PCRel:   form.pcrel,
			Length:  length,
		})
	}
	return wr.Err()
}

// lengthOf is r_length, which is a log2 byte width rather than a byte count.
func lengthOf(size int) (machocore.RelocLength, bool) {
	switch size {
	case 1:
		return machocore.RelocByte, true
	case 2:
		return machocore.RelocWord, true
	case 4:
		return machocore.RelocLong, true
	case 8:
		return machocore.RelocQuad, true
	}
	return 0, false
}

// deposit adds an implicit addend into the section's bytes.
//
// It adds rather than overwrites, because "the implicit addend is the value
// in the field" is what an implicit-addend format means. The encoder
// zero-fills every hole it leaves, so the two are the same today.
func deposit(content []byte, off, size int, addend int64) error {
	if off < 0 || size < 0 || off+size > len(content) {
		return fmt.Errorf("a %d-byte field at %#x does not fit %d bytes of section",
			size, off, len(content))
	}

	var cur uint64
	for i := 0; i < size; i++ {
		cur |= uint64(content[off+i]) << (8 * i)
	}
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

// refKindError is the one diagnostic this package raises for a reference
// Mach-O cannot express.
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
