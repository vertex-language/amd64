package obj

// RefKind states how a linker should resolve a reference.
//
// The set is the union of what the three containers can express, not the
// intersection. An intersection would be four kinds and would make every
// interesting object unbuildable; a per-container set would mean a lowering
// picks its container before it picks its instructions. So it is the union,
// every writer states what it cannot do, and ErrRefKind names the kind and
// the offset when a kind meets a container with no answer for it.
//
// It is declared once, here, and aliased upward. amd64.RefPLT32,
// operand.RefPLT32 and obj.RefPLT32 are the same constant and no conversion
// exists anywhere in the tree.
type RefKind uint8

const (
	RefNone RefKind = iota

	RefAbs64
	RefAbs32
	RefAbs32S
	RefAbs16
	RefAbs8

	RefPC64
	RefPC32
	RefPC16
	RefPC8

	RefPLT32

	RefGOT32
	RefGOTPCREL

	// Two kinds for one relocation semantics. A linker relaxing
	// mov foo@GOTPCREL(%rip), %reg into lea foo(%rip), %reg has to know how
	// many bytes back the instruction starts, and the answer depends on
	// whether a REX prefix is there. The encoder knows, but emitting the
	// relaxable kind is a statement that the addend is exactly the one the
	// psABI's transformation assumes, so the kind is the caller's to state
	// and the encoder refuses the wrong one.
	RefGOTPCRELX
	RefRexGOTPCRELX

	RefGOTOFF64
	RefGOTPC32

	RefSize32
	RefSize64

	RefTLSGD
	RefTLSLD
	RefDTPOFF32
	RefDTPOFF64
	RefGOTTPOFF
	RefTPOFF32
	RefTPOFF64

	// RefTLV is Mach-O's. A thread-local there is reached through a
	// descriptor and a call rather than a relocation model, which is why it
	// is one kind and the ELF models are seven.
	RefTLV

	RefImageRel32
	RefSecRel32
	RefSecIdx

	numRefKinds
)

var refNames = [numRefKinds]string{
	RefNone: "none",

	RefAbs64: "abs64", RefAbs32: "abs32", RefAbs32S: "abs32s",
	RefAbs16: "abs16", RefAbs8: "abs8",

	RefPC64: "pc64", RefPC32: "pc32", RefPC16: "pc16", RefPC8: "pc8",

	RefPLT32: "plt32",

	RefGOT32: "got32", RefGOTPCREL: "gotpcrel",
	RefGOTPCRELX: "gotpcrelx", RefRexGOTPCRELX: "rex-gotpcrelx",
	RefGOTOFF64: "gotoff64", RefGOTPC32: "gotpc32",

	RefSize32: "size32", RefSize64: "size64",

	RefTLSGD: "tlsgd", RefTLSLD: "tlsld",
	RefDTPOFF32: "dtpoff32", RefDTPOFF64: "dtpoff64",
	RefGOTTPOFF: "gottpoff",
	RefTPOFF32:  "tpoff32", RefTPOFF64: "tpoff64",

	RefTLV: "tlv",

	RefImageRel32: "imagerel32", RefSecRel32: "secrel32", RefSecIdx: "secidx",
}

// String is what ErrRefKind names when a writer refuses one.
func (k RefKind) String() string {
	if int(k) < len(refNames) {
		return refNames[k]
	}
	return "refkind?"
}

// Valid reports whether the value names a declared kind. RefNone is not one.
func (k RefKind) Valid() bool { return k > RefNone && k < numRefKinds }

// Size is the width in bytes of the field the kind fills, or 0 for RefNone.
//
// It is a property of the kind rather than of an architecture — a RefAbs64 is
// eight bytes wherever it appears — which is why it is answered here and not
// in a writer. The data-side builders need it: Section.Ref places a hole with
// no encoder in the path to have sized it.
func (k RefKind) Size() int {
	switch k {
	case RefAbs64, RefPC64, RefGOTOFF64, RefSize64, RefDTPOFF64, RefTPOFF64:
		return 8
	case RefAbs32, RefAbs32S, RefPC32, RefPLT32,
		RefGOT32, RefGOTPCREL, RefGOTPCRELX, RefRexGOTPCRELX, RefGOTPC32,
		RefSize32, RefTLSGD, RefTLSLD, RefDTPOFF32, RefGOTTPOFF, RefTPOFF32,
		RefTLV, RefImageRel32, RefSecRel32:
		return 4
	case RefAbs16, RefPC16, RefSecIdx:
		return 2
	case RefAbs8, RefPC8:
		return 1
	}
	return 0
}

// PCRel reports whether the kind resolves against the position of its own
// field rather than against an absolute address.
//
// A writer uses it to sanity-check a Reference whose PCRel disagrees with its
// kind. The GOT kinds are here because the x86-64 ones are RIP-relative loads
// of a GOT slot; RefGOT32 is not, because it is an offset into the table.
func (k RefKind) PCRel() bool {
	switch k {
	case RefPC64, RefPC32, RefPC16, RefPC8, RefPLT32,
		RefGOTPCREL, RefGOTPCRELX, RefRexGOTPCRELX, RefGOTPC32,
		RefTLSGD, RefTLSLD, RefGOTTPOFF:
		return true
	}
	return false
}

// TLS reports whether the kind names a thread-local storage model. The
// writers each accept a different subset and refuse the rest, so asking the
// question in one place keeps three answers from drifting.
func (k RefKind) TLS() bool {
	switch k {
	case RefTLSGD, RefTLSLD, RefDTPOFF32, RefDTPOFF64,
		RefGOTTPOFF, RefTPOFF32, RefTPOFF64, RefTLV:
		return true
	}
	return false
}

// Reference is one hole a linker fills.
//
// The identity every consumer relies on is
//
//	value = target - (section offset of the field) + Adjust + Addend
//
// and Adjust is the reason a caller never writes -4. A PC-relative field
// resolves against the end of the instruction, and the field is not always
// the last thing in it; the encoder that placed the field is the only thing
// that knows how many bytes follow it. A disp32 with nothing after it carries
// Adjust == -4; one with a trailing imm32 carries -8.
//
// The three writers do three different things with that one integer. ELF
// writes Addend+Adjust into r_addend and leaves the bytes alone. COFF selects
// a relocation type from it — REL32 through REL32_5 — and folds the logical
// Addend into the section bytes. Mach-O selects a type too, and has no
// SIGNED_3, so one value of Adjust is refusable there and nowhere else.
//
// Nothing in this struct is AMD64-specific, on purpose. Size is 1, 2, 4 or 8
// and Addend is an int64 because R_X86_64_64 against a quadword is an
// ordinary thing to want. The rule that made the narrower i386 struct right
// is unchanged: the architecture validates the width, not the struct.
type Reference struct {
	Offset int // where the hole starts, section-relative
	Size   int // 1, 2, 4 or 8
	PCRel  bool
	Adjust int64 // field-position correction, already computed
	Sym    string
	Kind   RefKind
	Addend int64 // logical addend, never adjusted for the field
}

// Offset reports whether a kind's value is a displacement rather than an
// address: a number to be added to something the instruction already holds.
//
// It is what decides whether a reference may share a displacement field with
// a base register. An address may not — the linker would have to know the
// base's run-time value to work out an addend — but an offset is exactly
// what a base register wants added to it, and that is how a thread-local is
// reached on every platform that has one. clang writes the COFF form as
//
//	leaq sym@SECREL32(%rax), %rax
//
// where rax already holds the thread's block, and the ELF forms the same way
// against a thread pointer.
func (k RefKind) Offset() bool {
	switch k {
	case RefSecRel32, RefTPOFF32, RefTPOFF64, RefDTPOFF32, RefDTPOFF64:
		return true
	}
	return false
}
