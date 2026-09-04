// Package macho writes an *obj.Object as a Mach-O relocatable object.
//
// One function, macho.Write. The CPU is derived from the object's Arch —
// ArchAMD64 is CPU_TYPE_X86_64 — and the header follows from it.
//
// The same object is legal for all three containers, so what this package
// refuses it refuses at emission and not at construction. Two things get
// refused here that nowhere else does: a section name with no segment to put
// it in (ErrSectionName), and a PC-relative reference whose field position
// has no Mach-O relocation type (ErrRefKind). Both are named with the section
// and the offset, because a lowering that produced one needs to find it.
//
// Emission goes through github.com/vertex-language/macho, the same module
// that reads these objects back, imported under an alias so the package a
// caller names is macho and the call site stays macho.Write.
package macho

import (
	"fmt"
	"io"

	machocore "github.com/vertex-language/macho"
	machoobj "github.com/vertex-language/macho/obj"

	"github.com/vertex-language/amd64/obj"
)

// SegSect is a Mach-O section identity: the segment it lives in and the name
// it carries there.
//
// It exists because Mach-O is the container that disagrees with the other two
// about naming. A section name is not enough to place a section — a segment
// is a load-time protection decision — so a custom name this package has
// never seen needs one supplied.
type SegSect struct {
	Segment string
	Section string
}

// Options are the things the assembler has no opinion about.
//
// There is no Target field. The object names its architecture; what is left
// is the platform, which the object cannot know and every linker needs.
type Options struct {
	// Platform is required. An object with no platform makes every linker
	// guess, and ld64 warns about exactly that, so an unset one is an error
	// rather than a default.
	Platform machocore.Platform

	// MinOS is the deployment target, e.g. "12.0". It is required for the
	// same reason Platform is: LC_BUILD_VERSION with a zero minos is the
	// thing the warning is about.
	MinOS string

	// SDK is the SDK version the object was built against. Optional; zero is
	// what an object built outside an SDK carries.
	SDK string

	// Subsections sets MH_SUBSECTIONS_VIA_SYMBOLS, which promises every
	// section can be cut at symbol boundaries. That is what lets the linker
	// dead-strip and reorder at function granularity, and it is a promise
	// about how the module was built, which is why it is yours to make.
	Subsections bool

	// Sections gives a segment and a name to a section this package has no
	// mapping for. The key is the section's name as the module spelled it.
	//
	// It is the escape hatch for ErrSectionName and it is one map entry:
	//
	//	Sections: map[string]SegSect{".mysec": {"__DATA", "__mysec"}},
	Sections map[string]SegSect
}

// Write emits o to w as an MH_OBJECT file.
//
// The order below is the one the format imposes. Sections are created first
// because a symbol names a section; symbols come next because a relocation
// names a symbol; contents come last because Mach-O is an implicit-addend
// format and a section's bytes are not final until every hole in them has
// been paid for.
func Write(w io.Writer, o *obj.Object, opts ...Options) error {
	if o == nil {
		return fmt.Errorf("macho: nil object")
	}
	if len(opts) > 1 {
		return fmt.Errorf("macho: Write takes at most one Options, got %d", len(opts))
	}
	var opt Options
	if len(opts) == 1 {
		opt = opts[0]
	}
	if o.Arch() != obj.ArchAMD64 {
		return fmt.Errorf("macho: object is %s, not %s", o.Arch(), obj.ArchAMD64)
	}

	target, err := targetFor(opt)
	if err != nil {
		return err
	}

	var flags machocore.Flags
	if opt.Subsections {
		flags |= machocore.MH_SUBSECTIONS_VIA_SYMBOLS
	}

	wr := machoobj.NewWriter(w, machoobj.Options{
		Target: target,
		Flags:  flags,
		// Build carries the same three values Target does. The duplication is
		// the underlying writer's contract: LC_BUILD_VERSION is required and
		// a zero value must fail loudly rather than be inferred.
		Build: target.Build(),
	})

	secs := o.Sections()
	builders := make([]*machoobj.SectionBuilder, len(secs))
	for i, s := range secs {
		b, err := newSection(wr, s, opt)
		if err != nil {
			return err
		}
		builders[i] = b
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

// targetFor builds the Mach-O target from the options and this package's one
// architecture.
//
// Platform and MinOS are checked here rather than left to the underlying
// writer because the failure they produce there is a zero LC_BUILD_VERSION,
// which is a file that links and then behaves oddly rather than one that
// fails.
func targetFor(opt Options) (machocore.Target, error) {
	if opt.Platform == machocore.PlatformUnknown {
		return machocore.Target{}, fmt.Errorf("macho: Options.Platform is required")
	}
	if opt.MinOS == "" {
		return machocore.Target{}, fmt.Errorf("macho: Options.MinOS is required, e.g. \"12.0\"")
	}
	minOS, err := machocore.ParseVersion(opt.MinOS)
	if err != nil {
		return machocore.Target{}, fmt.Errorf("macho: Options.MinOS: %w", err)
	}
	var sdk machocore.Version
	if opt.SDK != "" {
		if sdk, err = machocore.ParseVersion(opt.SDK); err != nil {
			return machocore.Target{}, fmt.Errorf("macho: Options.SDK: %w", err)
		}
	}
	return machocore.Target{
		CPU:      machocore.CPU_TYPE_X86_64,
		SubCPU:   machocore.CPU_SUBTYPE_X86_64_ALL,
		Platform: opt.Platform,
		MinOS:    minOS,
		SDK:      sdk,
		Endian:   machocore.LittleEndian,
	}, nil
}
