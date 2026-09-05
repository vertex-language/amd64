package obj

// SectionKind is how a section behaves at load time.
//
// Names are spelled the ELF way and translated by the writers. A name has to
// be spelled some way, and one of the three containers had to be the one the
// vocabulary borrows from.
type SectionKind uint8

const (
	Text SectionKind = iota
	Data
	ROData
	BSS

	// RelROData is read-only data that a loader has to write first.
	//
	// A constant whose value is another symbol's address -- a vtable,
	// a table of function pointers, a const pointer -- is not known
	// until something has relocated it, so its bytes cannot live
	// somewhere never written. Nor do they belong in .data: once the
	// relocation is applied the bytes are constant, and leaving them
	// writable for the life of the process gives an attacker a table
	// of function pointers to edit.
	//
	// The three containers answer this differently, which is why it
	// is a kind rather than a name. ELF has .data.rel.ro, allocated
	// and writable, which the loader mprotects read-only after
	// relocating. Mach-O has (__DATA,__const), which ld moves into
	// the __DATA_CONST segment. COFF has no such concept and does not
	// need one: base relocations are applied before page protections
	// are set, so read-only .rdata is both correct and already what
	// this wants.
	RelROData
)

var sectionKindNames = [...]string{
	Text:   ".text",
	Data:   ".data",
	ROData: ".rodata",
	BSS:    ".bss",

	RelROData: ".data.rel.ro",
}

// String is the conventional name, and is exactly what Module.Section hands
// to SectionNamed, so the two cannot disagree about what ".text" is.
func (k SectionKind) String() string {
	if int(k) < len(sectionKindNames) {
		return sectionKindNames[k]
	}
	return ".section?"
}

// Valid reports whether the value names a declared kind.
func (k SectionKind) Valid() bool { return int(k) < len(sectionKindNames) }

// NoBits reports whether the section has a size in memory and no bytes in the
// file. All three writers ask this, and asking it here is why they cannot
// disagree about whether .bss is written out.
func (k SectionKind) NoBits() bool { return k == BSS }

// Exec reports whether the section holds instructions. It is also what
// decides whether Align pads with nops or with zeros.
func (k SectionKind) Exec() bool { return k == Text }

// Writable reports whether the section is writable at run time.
func (k SectionKind) Writable() bool { return k == Data || k == BSS || k == RelROData }
