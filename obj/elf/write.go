// Package elf writes an *obj.Object as an ELF relocatable object.
//
// One function, elf.Write. The target is derived from the object's Arch —
// ArchAMD64 implies ELFCLASS64, little-endian, EM_X86_64 — because an object
// states what it was assembled for and a writer that took that as an argument
// would be a second place for the answer to live.
//
// Emission goes through github.com/vertex-language/elf, the same module that
// reads these objects back, so a round-trip test is a round trip and not two
// independent guesses about the format. That module is imported under an
// alias here so the package a caller names is elf and the call site stays
// elf.Write.
//
// This package is a leaf. Nothing in amd64/obj imports it, and a build that
// emits COFF never compiles a line of it.
package elf

import (
	"fmt"
	"io"

	elfcore "github.com/vertex-language/elf"
	elfobj "github.com/vertex-language/elf/obj"

	"github.com/vertex-language/amd64/obj"
)

// Options are the things the assembler has no opinion about.
//
// There is no Target field. The object names its architecture and everything
// an AMD64 ELF header needs follows from it; a caller who could pass a
// different one could produce a file whose header disagrees with its bytes.
type Options struct {
	// OSABI is e_ident[EI_OSABI]. Zero — ELFOSABI_NONE — is what most Linux
	// objects carry and is the right default; state it only when a consumer
	// needs it stated.
	//
	// It is mapped onto the target's OS, which is what the underlying writer
	// derives the byte from. Nothing here can emit STB_GNU_UNIQUE or
	// STT_GNU_IFUNC, so what is asked for is what is written.
	OSABI elfcore.OSABI

	// ABIVersion is e_ident[EI_ABIVERSION]. Zero for everything current.
	ABIVersion uint8

	// GNUStack selects the .note.GNU-stack declaration. The zero value emits
	// a non-executable one, which is the safe default: omitting the section
	// makes the kernel assume the stack should be executable.
	GNUStack elfcore.GNUStack

	// Comment is written to a .comment section when non-empty. It is the
	// producer string a toolchain stamps on its output.
	Comment string

	// StripLocals drops local symbols from .symtab.
	//
	// A local that a relocation names is kept regardless. Dropping it would
	// leave a hole pointing at nothing, and a stripped object that does not
	// link is worse than a slightly larger one.
	StripLocals bool
}

// Write emits o to w as an ET_REL object.
//
// Everything is decided here and nothing reaches w until the underlying
// writer closes, because the ELF header carries the section header table
// offset and every section header carries a file offset.
func Write(w io.Writer, o *obj.Object, opts ...Options) error {
	if o == nil {
		return fmt.Errorf("elf: nil object")
	}
	if len(opts) > 1 {
		return fmt.Errorf("elf: Write takes at most one Options, got %d", len(opts))
	}
	var opt Options
	if len(opts) == 1 {
		opt = opts[0]
	}
	if o.Arch() != obj.ArchAMD64 {
		return fmt.Errorf("elf: object is %s, not %s", o.Arch(), obj.ArchAMD64)
	}

	wr := elfobj.NewWriter(w, elfobj.Options{
		Target: elfcore.Target{
			Arch:   elfcore.ArchAMD64,
			Class:  elfcore.ELFCLASS64,
			Endian: elfcore.EndianLittle,
			OS:     osFor(opt.OSABI),
		},
		ABIVersion: opt.ABIVersion,
		GNUStack:   opt.GNUStack,

		// x86-64 is a RELA architecture: the addend lives in the relocation
		// entry and the section bytes are handed through untouched. This is
		// stated rather than left to RelocAuto because a psABI table that
		// ever said otherwise should break loudly here instead of quietly
		// writing an addend no linker will read. Nothing in this package
		// supplies a RELEncoder, and on this target nothing needs one.
		RelocFormat: elfcore.RelocRELA,
	})

	secs := o.Sections()
	builders := make([]*elfobj.SectionBuilder, len(secs))
	for i, s := range secs {
		b, err := writeSection(wr, s)
		if err != nil {
			return err
		}
		builders[i] = b
	}

	syms, err := writeSymbols(wr, o, builders, opt)
	if err != nil {
		return err
	}

	for i, s := range secs {
		if err := writeRelocs(wr, builders[i], s, syms); err != nil {
			return err
		}
	}

	if opt.Comment != "" {
		c := wr.Section(elfobj.SectionHeader{
			Name:      ".comment",
			Type:      elfcore.SHT_PROGBITS,
			Flags:     elfcore.SHF_MERGE | elfcore.SHF_STRINGS,
			Addralign: 1,
			Entsize:   1,
		})
		if _, err := c.WriteString(opt.Comment); err != nil {
			return err
		}
		if err := c.WriteByte(0); err != nil {
			return err
		}
	}

	return wr.Close()
}

// osFor maps a requested OSABI byte onto the target OS the underlying writer
// derives it from. It is a small table rather than a pass-through because a
// Target is the one place class, endianness, machine and OSABI are guaranteed
// not to contradict each other, and reaching around it would give that up.
//
// An OSABI with no OS in this vocabulary resolves to OSNone, which is
// ELFOSABI_NONE: unstated, which is what the overwhelming majority of objects
// carry and what every linker accepts.
func osFor(a elfcore.OSABI) elfcore.OS {
	switch a {
	// ELFOSABI_LINUX is a historical alias for ELFOSABI_GNU and the same
	// byte, so one case covers both spellings.
	case elfcore.ELFOSABI_GNU:
		return elfcore.OSLinux
	case elfcore.ELFOSABI_FREEBSD:
		return elfcore.OSFreeBSD
	case elfcore.ELFOSABI_NETBSD:
		return elfcore.OSNetBSD
	case elfcore.ELFOSABI_OPENBSD:
		return elfcore.OSOpenBSD
	case elfcore.ELFOSABI_SOLARIS:
		return elfcore.OSSolaris
	}
	return elfcore.OSNone
}
