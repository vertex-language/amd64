package elf

import (
	"fmt"

	elfcore "github.com/vertex-language/elf"
	elfobj "github.com/vertex-language/elf/obj"

	"github.com/vertex-language/amd64/obj"
)

// The intent→R_X86_64_* table. This is the only place in the tree that knows
// what an ELF relocation number is.
//
// Every entry is a link-semantics decision the lowering already made, so
// nothing here chooses between PLT and direct or between GOT and absolute —
// it spells what was chosen. Two pairs are worth reading against the psABI:
//
//	RefGOTPCRELX     → R_X86_64_GOTPCRELX      relaxable, no REX prefix
//	RefRexGOTPCRELX  → R_X86_64_REX_GOTPCRELX  relaxable, REX prefix present
//
// They are two kinds for one semantics because a linker rewriting
// mov foo@GOTPCREL(%rip), %reg into lea foo(%rip), %reg has to know how many
// bytes back the instruction starts, and the answer depends on the prefix.
//
// Four kinds have no row and are refused: RefTLV is Mach-O's thread-local
// descriptor, and RefImageRel32, RefSecRel32 and RefSecIdx are COFF's. Each
// refusal is ErrRefKind at emission rather than at construction, because the
// same object is legal for the container the kind came from.
var relocTypes = map[obj.RefKind]elfcore.RelocX86_64{
	obj.RefAbs64:  elfcore.R_X86_64_64,
	obj.RefAbs32:  elfcore.R_X86_64_32,
	obj.RefAbs32S: elfcore.R_X86_64_32S,
	obj.RefAbs16:  elfcore.R_X86_64_16,
	obj.RefAbs8:   elfcore.R_X86_64_8,

	obj.RefPC64: elfcore.R_X86_64_PC64,
	obj.RefPC32: elfcore.R_X86_64_PC32,
	obj.RefPC16: elfcore.R_X86_64_PC16,
	obj.RefPC8:  elfcore.R_X86_64_PC8,

	obj.RefPLT32: elfcore.R_X86_64_PLT32,

	obj.RefGOT32:        elfcore.R_X86_64_GOT32,
	obj.RefGOTPCREL:     elfcore.R_X86_64_GOTPCREL,
	obj.RefGOTPCRELX:    elfcore.R_X86_64_GOTPCRELX,
	obj.RefRexGOTPCRELX: elfcore.R_X86_64_REX_GOTPCRELX,
	obj.RefGOTOFF64:     elfcore.R_X86_64_GOTOFF64,
	obj.RefGOTPC32:      elfcore.R_X86_64_GOTPC32,

	obj.RefSize32: elfcore.R_X86_64_SIZE32,
	obj.RefSize64: elfcore.R_X86_64_SIZE64,

	obj.RefTLSGD:    elfcore.R_X86_64_TLSGD,
	obj.RefTLSLD:    elfcore.R_X86_64_TLSLD,
	obj.RefDTPOFF32: elfcore.R_X86_64_DTPOFF32,
	obj.RefDTPOFF64: elfcore.R_X86_64_DTPOFF64,
	obj.RefGOTTPOFF: elfcore.R_X86_64_GOTTPOFF,
	obj.RefTPOFF32:  elfcore.R_X86_64_TPOFF32,
	obj.RefTPOFF64:  elfcore.R_X86_64_TPOFF64,
}

// writeRelocs translates one section's holes into RELA entries.
//
// The addend written is Addend + Adjust, and that sum is the whole contract:
//
//	value = target - (section offset of the field) + Adjust + Addend
//
// is what a Reference means, and R_X86_64_PC32 computes S + A - P. Equating
// the two makes A the sum. For an absolute relocation Adjust is zero and the
// sum degenerates to Addend. One expression, both cases, and no arithmetic
// anywhere upstream.
//
// Nothing is deposited into the section bytes. That is the RELA half of the
// bargain and the reason this writer can hand an object's own bytes straight
// through while the other two copy.
func writeRelocs(wr *elfobj.Writer, b *elfobj.SectionBuilder, s *obj.Section, syms map[string]elfobj.SymRef) error {
	for _, r := range s.Refs() {
		typ, ok := relocTypes[r.Kind]
		if !ok {
			return &obj.Error{
				Sentinel: obj.ErrRefKind,
				Arch:     obj.ArchAMD64,
				Section:  s.Name(),
				Offset:   r.Offset,
				Context:  fmt.Sprintf("%s reference to %q", r.Kind, r.Sym),
				Notes:    []string{"x86-64 ELF has no relocation for this kind"},
			}
		}
		if want := r.Kind.Size(); want != r.Size {
			return fmt.Errorf("elf: %s+%#x: %s reference to %q is a %d-byte field; %v writes %d",
				s.Name(), r.Offset, r.Kind, r.Sym, r.Size, r.Kind, want)
		}

		sym, ok := syms[r.Sym]
		if !ok {
			// Finalize refuses a reference to a name neither defined nor
			// declared Extern, so this is either a stripped symbol that
			// referencedNames should have kept or an object that did not come
			// from Finalize.
			return fmt.Errorf("elf: %s+%#x: reference to %q, which is not in the symbol table",
				s.Name(), r.Offset, r.Sym)
		}

		wr.Reloc(b, elfobj.RelocSpec{
			Offset: uint64(r.Offset),
			Sym:    sym,
			Type:   uint32(typ),
			Addend: r.Addend + r.Adjust,
		})
	}
	return wr.Err()
}
