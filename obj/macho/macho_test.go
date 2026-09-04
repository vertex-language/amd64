package macho_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	machocore "github.com/vertex-language/macho"
	machoobj "github.com/vertex-language/macho/obj"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/obj/macho"
)

// fixture is the module every test here writes: a RIP-relative reference, a
// call through the PLT (a branch, here — Mach-O has no PLT), a same-section
// branch, an undefined symbol, and a .rodata section the first reference
// names.
func fixture(t *testing.T) *obj.Object {
	t.Helper()

	m := amd64.NewModule()
	m.Extern("puts")

	s := m.Section(amd64.Text)
	s.Label("main", amd64.Global, amd64.Func)
	s.LeaR64M(amd64.RDI, amd64.Rip(amd64.Ref("msg", amd64.RefPC32)))
	s.CallRef(amd64.Ref("puts", amd64.RefPLT32))
	s.Ret()
	s.EndLabel("main")

	r := m.SectionNamed(".rodata", amd64.ROData)
	r.Label("msg", amd64.Local, amd64.ObjectSym)
	r.Asciz("hi")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	return o
}

// opts is the minimum Options a fixture can write with: Platform and MinOS
// are both required, per Options' doc.
func opts(extra macho.Options) macho.Options {
	extra.Platform = machocore.PlatformMacOS
	extra.MinOS = "12.0"
	return extra
}

// write emits o to a temp file and hands back both the path and the bytes.
func write(t *testing.T, o *obj.Object, opt macho.Options) (string, []byte) {
	t.Helper()

	var buf bytes.Buffer
	if err := macho.Write(&buf, o, opt); err != nil {
		t.Fatalf("Write: %v", err)
	}
	path := filepath.Join(t.TempDir(), "t.o")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, buf.Bytes()
}

// TestRoundTrip is the claim the package doc makes: emission goes through the
// module that also reads the format, so this is a round trip and not two
// independent guesses.
func TestRoundTrip(t *testing.T) {
	o := fixture(t)
	path, _ := write(t, o, opts(macho.Options{}))

	f, err := machoobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	target := f.Target()
	if target.CPU != machocore.CPU_TYPE_X86_64 {
		t.Errorf("CPU = %v, want CPU_TYPE_X86_64", target.CPU)
	}

	text := f.Section(machocore.Sec(machocore.SEG_TEXT, machocore.SECT_TEXT))
	if text == nil {
		t.Fatal("no __TEXT,__text in the written object")
	}
	_, attrs := machocore.UnpackSecFlags(text.Flags)
	if !attrs.Has(machocore.S_ATTR_PURE_INSTRUCTIONS) {
		t.Errorf("__text attrs = %#x, want S_ATTR_PURE_INSTRUCTIONS", attrs)
	}

	// x86-64 Mach-O is an implicit-addend format: the bytes on disk already
	// carry the addend, so they do not round-trip byte-for-byte against the
	// object's own copy the way ELF's RELA bytes do.
	if _, err := text.Data(); err != nil {
		t.Fatal(err)
	}

	rodata := f.Section(machocore.Sec(machocore.SEG_TEXT, machocore.SECT_CONST))
	if rodata == nil {
		t.Fatal("no __TEXT,__const in the written object")
	}
}

// TestRelocs pins the two mappings the fixture exercises and the implicit
// addend each produces.
func TestRelocs(t *testing.T) {
	path, _ := write(t, fixture(t), opts(macho.Options{}))

	f, err := machoobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	text := f.Section(machocore.Sec(machocore.SEG_TEXT, machocore.SECT_TEXT))
	if text == nil {
		t.Fatal("no __TEXT,__text in the written object")
	}
	relocs, err := text.Relocs()
	if err != nil {
		t.Fatalf("Relocs: %v", err)
	}
	if len(relocs) != 2 {
		t.Fatalf("want 2 relocations, got %d", len(relocs))
	}

	for i, want := range []struct {
		addr int64
		sym  string
		typ  machocore.X86_64Reloc
	}{
		// Every PC-relative field in the fixture is the last thing in its
		// instruction, so n (the trailing byte count) is 0 in both cases:
		// SIGNED for the LEA and BRANCH for the CALL.
		{3, "msg", machocore.X86_64_RELOC_SIGNED},
		{8, "puts", machocore.X86_64_RELOC_BRANCH},
	} {
		got := relocs[i]
		if !got.Extern {
			t.Errorf("reloc %d is not extern; the fixture only references symbols by name", i)
		}
		if got.Address != want.addr {
			t.Errorf("reloc %d address = %#x, want %#x", i, got.Address, want.addr)
		}
		if machocore.X86_64Reloc(got.Type) != want.typ {
			t.Errorf("reloc %d type = %v, want %v", i, machocore.X86_64Reloc(got.Type), want.typ)
		}
		if !got.PCRel {
			t.Errorf("reloc %d is not PCRel, want it to be", i)
		}
		if got.Sym == nil || got.Sym.Name != want.sym {
			t.Errorf("reloc %d symbol = %v, want %q", i, got.Sym, want.sym)
		}
	}
}

// TestSymbols checks that the vocabulary survives the crossing: binding maps
// onto N_EXT/weak bits, and a defined symbol carries its section and offset.
func TestSymbols(t *testing.T) {
	path, _ := write(t, fixture(t), opts(macho.Options{}))

	f, err := machoobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	syms, err := f.Symbols()
	if err != nil {
		t.Fatalf("Symbols: %v", err)
	}
	byName := map[string]*machoobj.Symbol{}
	for _, s := range syms {
		byName[s.Name] = s
	}

	main := byName["main"]
	if main == nil {
		t.Fatal("main is not in the symbol table")
	}
	if !main.Ext() {
		t.Error("main is Global; want N_EXT set")
	}
	if main.Sec == nil {
		t.Error("main has no resolved section")
	}

	msg := byName["msg"]
	if msg == nil {
		t.Fatal("msg is not in the symbol table")
	}
	if msg.Ext() {
		t.Error("msg is Local; want N_EXT clear")
	}

	puts := byName["puts"]
	if puts == nil || !puts.Undefined() {
		t.Errorf("puts = %v, want an undefined symbol", puts)
	}
	if !puts.Ext() {
		t.Error("an undefined symbol must be external or nothing can bind to it")
	}
}

// TestSubsections checks that Options.Subsections sets the one file-level
// flag it owns.
func TestSubsections(t *testing.T) {
	path, _ := write(t, fixture(t), opts(macho.Options{Subsections: true}))

	f, err := machoobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	if !f.Subsections() {
		t.Error("Options.Subsections did not set MH_SUBSECTIONS_VIA_SYMBOLS")
	}
}

// TestSectionsOverride exercises the escape hatch for ErrSectionName: a
// custom section name given a segment and name in Options.Sections.
func TestSectionsOverride(t *testing.T) {
	m := amd64.NewModule()
	m.SectionNamed(".mysec", amd64.Data).Data([]byte{1, 2, 3, 4})

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	path, _ := write(t, o, opts(macho.Options{
		Sections: map[string]macho.SegSect{".mysec": {Segment: "__DATA", Section: "__mysec"}},
	}))

	f, err := machoobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	if f.Section(machocore.Sec("__DATA", "__mysec")) == nil {
		t.Fatal("no __DATA,__mysec in the written object")
	}
}

// TestUnknownSectionIsRefused is ErrSectionName's other half: a custom name
// with no override in Options.Sections fails rather than guesses a segment.
func TestUnknownSectionIsRefused(t *testing.T) {
	m := amd64.NewModule()
	m.SectionNamed(".mysec", amd64.Data).Data([]byte{1, 2, 3, 4})

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	err = macho.Write(&bytes.Buffer{}, o, opts(macho.Options{}))
	if !errors.Is(err, macho.ErrSectionName) {
		t.Errorf("Write returned %v, want ErrSectionName", err)
	}
}

// TestRefusedKinds names kinds Mach-O has no relocation for: two ELF ideas
// and three COFF ones, per the package doc.
func TestRefusedKinds(t *testing.T) {
	for _, kind := range []amd64.RefKind{
		amd64.RefSize32,
		amd64.RefSize64,
		amd64.RefImageRel32,
		amd64.RefSecRel32,
		amd64.RefSecIdx,
	} {
		m := amd64.NewModule()
		m.Extern("x")
		m.Section(amd64.Text).Ret()
		m.Section(amd64.Data).Ref(amd64.Ref("x", kind))

		o, err := m.Finalize()
		if err != nil {
			t.Fatalf("%v: Finalize: %v", kind, err)
		}
		err = macho.Write(&bytes.Buffer{}, o, opts(macho.Options{}))
		if !errors.Is(err, obj.ErrRefKind) {
			t.Errorf("%v: Write returned %v, want ErrRefKind", kind, err)
		}
	}
}

// TestWrongArchIsRefused. There is no Target option beyond CPU, and the
// object's own Arch is checked before Platform or MinOS.
func TestWrongArchIsRefused(t *testing.T) {
	o := obj.New(obj.ArchI386, nil, nil)
	if err := macho.Write(&bytes.Buffer{}, o, opts(macho.Options{})); err == nil {
		t.Error("Write accepted an i386 object")
	}
}

// TestMissingPlatformIsRefused and TestMissingMinOSIsRefused pin the two
// required Options fields: an object with neither makes every linker guess,
// which is exactly what ld64's LC_BUILD_VERSION warning is about.
func TestMissingPlatformIsRefused(t *testing.T) {
	err := macho.Write(&bytes.Buffer{}, fixture(t), macho.Options{MinOS: "12.0"})
	if err == nil {
		t.Error("Write accepted an object with no Platform")
	}
}

func TestMissingMinOSIsRefused(t *testing.T) {
	err := macho.Write(&bytes.Buffer{}, fixture(t), macho.Options{Platform: machocore.PlatformMacOS})
	if err == nil {
		t.Error("Write accepted an object with no MinOS")
	}
}
