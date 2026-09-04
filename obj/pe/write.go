// Package pe writes an *obj.Object as a COFF relocatable object.
//
// One function, pe.Write. The machine is derived from the object's Arch —
// ArchAMD64 is IMAGE_FILE_MACHINE_AMD64 — and everything else about the
// header follows from it.
//
// The same object is legal for all three containers, so what this package
// refuses it refuses at emission and not at construction: a RefKind COFF has
// no relocation for — RefPLT32, any GOT kind, any ELF TLS model — is
// ErrRefKind naming the kind, the symbol and the section offset. amd64 never
// learns that COFF exists; this is where it is learned.
//
// Emission goes through github.com/vertex-language/pe, the same module that
// reads these objects back, imported under an alias so the package a caller
// names is pe and the call site stays pe.Write.
package pe

import (
	"fmt"
	"io"

	pecore "github.com/vertex-language/pe"
	"github.com/vertex-language/pe/coff"

	"github.com/vertex-language/amd64/obj"
)

// Options are the things the assembler has no opinion about.
//
// There is no Target field, for the same reason there is none on the other
// two writers: the object names its architecture and a caller who could pass
// a different one could produce a file whose header disagrees with its bytes.
type Options struct {
	// ABI selects the toolchain convention the object is written for. It
	// changes nothing in the bytes today — the difference between MSVC and
	// MinGW shows up in section naming and .drectve spellings, both of which
	// this package leaves alone — but an object that does not state it makes
	// every consumer guess, and the guess is wrong for every MinGW build.
	//
	// The zero value is ABIMSVC, which is what an unstated object is assumed
	// to be everywhere else in that tree.
	ABI pecore.ABI

	// TimeDateStamp is written verbatim. Zero is the deterministic choice and
	// the default here; link.exe's /Brepro writes 0xffffffff.
	TimeDateStamp uint32

	// BigObj decides the header family. Auto promotes when the section count
	// requires it, which for a module with one section per kind it never
	// does; the field is here for the builds that compare output against a
	// reference and need the choice pinned.
	BigObj coff.BigObjMode

	// Characteristics is the COFF file header's flag field. Most of its bits
	// are image-only and meaningless in an object. A non-zero value forbids
	// promotion to bigobj, whose header has no such field.
	Characteristics pecore.FileChar

	// File is recorded as a .file symbol when non-empty: the source path a
	// debugger and a map file want.
	File string

	// Directives are linker options for the .drectve section — DEFAULTLIB and
	// EXPORT are the two anyone emits. The name's leading slash is supplied
	// by the writer; the value keeps its case.
	Directives []Directive
}

// Directive is one .drectve option.
type Directive struct {
	Name  string
	Value string
}

// ErrWeak names the thing COFF cannot express, and is kept for a caller
// matching on it.
//
// A COFF weak external is not a weaker binding; it is a name plus the
// alternate definition to use when nothing else defines it, and the alternate
// is a second symbol carried in an auxiliary record. It is a reference, not a
// definition, so it is not what a weak definition needs — and what a weak
// definition does need, a SELECT_ANY COMDAT, needs the symbol in a section of
// its own, which this writer does not produce. classOf says what happens
// instead.
var ErrWeak = fmt.Errorf("pe: a COFF weak external needs an alternate symbol")

// Write emits o to w as a relocatable COFF object.
//
// The order below is the one COFF's own structure imposes. Sections are
// created first because a symbol names a section number; symbols come next
// because a relocation names a symbol slot; contents come last because COFF
// is an implicit-addend format and a section's bytes are not final until
// every hole in them has been paid for.
func Write(w io.Writer, o *obj.Object, opts ...Options) error {
	if o == nil {
		return fmt.Errorf("pe: nil object")
	}
	if len(opts) > 1 {
		return fmt.Errorf("pe: Write takes at most one Options, got %d", len(opts))
	}
	var opt Options
	if len(opts) == 1 {
		opt = opts[0]
	}
	if o.Arch() != obj.ArchAMD64 {
		return fmt.Errorf("pe: object is %s, not %s", o.Arch(), obj.ArchAMD64)
	}

	abi := opt.ABI
	if abi == pecore.ABIUnknown {
		abi = pecore.ABIMSVC
	}

	wr := coff.NewWriter(w, coff.Options{
		Target: pecore.Target{
			Machine: pecore.MachineAMD64,
			ABI:     abi,
			OS:      pecore.OSWindows,
		},
		BigObj:          opt.BigObj,
		TimeDateStamp:   opt.TimeDateStamp,
		Characteristics: opt.Characteristics,
	})

	if opt.File != "" {
		wr.FileSymbol(opt.File)
	}
	for _, d := range opt.Directives {
		wr.Directive(d.Name, d.Value)
	}

	secs := o.Sections()
	builders := make([]*coff.SectionBuilder, len(secs))
	for i, s := range secs {
		builders[i] = newSection(wr, s)
	}

	syms, err := writeSymbols(wr, o, builders)
	if err != nil {
		return err
	}

	for i, s := range secs {
		if err := writeContents(wr, builders[i], s, syms); err != nil {
			return err
		}
	}

	return wr.Close()
}
