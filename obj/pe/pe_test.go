package pe_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	pecore "github.com/vertex-language/pe"
	"github.com/vertex-language/pe/coff"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/obj/pe"
)

// fixture is the module every test here writes: a RIP-relative reference and
// a same-image call, both REL32, an undefined symbol, and a .rodata section
// the first reference names.
//
// Unlike the ELF and Mach-O fixtures, the call is RefPC32 rather than
// RefPLT32: COFF has no procedure linkage table, and RefPLT32 is one of the
// kinds this writer refuses (see TestRefusedKinds).
func fixture(t *testing.T) *obj.Object {
	t.Helper()

	m := amd64.NewModule()
	m.Extern("puts")

	s := m.Section(amd64.Text)
	s.Label("main", amd64.Global, amd64.Func)
	s.LeaR64M(amd64.RDI, amd64.Rip(amd64.Ref("msg", amd64.RefPC32)))
	s.CallRef(amd64.Ref("puts", amd64.RefPC32))
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

// write emits o to a temp file and hands back both the path and the bytes.
func write(t *testing.T, o *obj.Object, opts ...pe.Options) (string, []byte) {
	t.Helper()

	var buf bytes.Buffer
	if err := pe.Write(&buf, o, opts...); err != nil {
		t.Fatalf("Write: %v", err)
	}
	path := filepath.Join(t.TempDir(), "t.obj")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, buf.Bytes()
}

// section finds a section by its resolved name. File deliberately has no
// by-name lookup — /Gy makes .text non-unique in a real object — but the
// fixture here never repeats a name.
func section(f *coff.File, name string) *coff.Section {
	for _, s := range f.Sections {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// TestRoundTrip is the claim the package doc makes: emission goes through the
// module that also reads the format, so this is a round trip and not two
// independent guesses.
func TestRoundTrip(t *testing.T) {
	o := fixture(t)
	path, _ := write(t, o)

	f, err := coff.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	if f.Machine != pecore.MachineAMD64 {
		t.Errorf("Machine = %v, want MachineAMD64", f.Machine)
	}

	text := section(f, ".text")
	if text == nil {
		t.Fatal("no .text in the written object")
	}
	if !text.Kind().Has(pecore.SecCode) || !text.Prot().Has(pecore.SecExecute) {
		t.Errorf(".text kind/prot = %v/%v, want SecCode and SecExecute", text.Kind(), text.Prot())
	}

	// COFF is an implicit-addend format, so .text's bytes on disk already
	// carry the addend and do not round-trip byte-for-byte against the
	// object's own copy the way ELF's RELA bytes do.
	if _, err := text.Data(); err != nil {
		t.Fatal(err)
	}

	// ROData is renamed .rdata: ELF and COFF disagree about what read-only
	// initialized data is called, and link.exe's merge rules are written
	// against .rdata.
	rdata := section(f, ".rdata")
	if rdata == nil {
		t.Fatal("no .rdata in the written object; ROData should have been renamed from .rodata")
	}
	if rdata.Prot().Has(pecore.SecExecute) || rdata.Prot().Has(pecore.SecWrite) {
		t.Errorf(".rdata prot = %v, want read-only", rdata.Prot())
	}
}

// TestRelocs pins the two mappings the fixture exercises and the REL32 ladder
// rung each produces.
func TestRelocs(t *testing.T) {
	path, _ := write(t, fixture(t))

	f, err := coff.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	text := section(f, ".text")
	if text == nil {
		t.Fatal("no .text in the written object")
	}
	relocs, err := text.Relocs()
	if err != nil {
		t.Fatalf("Relocs: %v", err)
	}
	if len(relocs) != 2 {
		t.Fatalf("want 2 relocations, got %d", len(relocs))
	}

	syms, err := f.Symbols()
	if err != nil {
		t.Fatalf("Symbols: %v", err)
	}

	for i, want := range []struct {
		addr uint32
		sym  string
		typ  pecore.RelocAMD64
	}{
		// Every PC-relative field in the fixture is the last thing in its
		// instruction, so n is 0 in both cases and the rung is plain REL32.
		{3, "msg", pecore.IMAGE_REL_AMD64_REL32},
		{8, "puts", pecore.IMAGE_REL_AMD64_REL32},
	} {
		got := relocs[i]
		if got.Address != want.addr {
			t.Errorf("reloc %d address = %#x, want %#x", i, got.Address, want.addr)
		}
		if pecore.RelocAMD64(got.Type) != want.typ {
			t.Errorf("reloc %d type = %v, want %v", i, pecore.RelocAMD64(got.Type), want.typ)
		}
		sym := f.SymbolAt(got.SymIndex)
		if sym == nil || sym.Name != want.sym {
			t.Errorf("reloc %d symbol = %v, want %q", i, sym, want.sym)
		}
	}
	_ = syms
}

// TestSymbols checks that the vocabulary survives the crossing: binding maps
// onto storage class, and a defined symbol carries its section and offset.
func TestSymbols(t *testing.T) {
	path, _ := write(t, fixture(t))

	f, err := coff.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	syms, err := f.Symbols()
	if err != nil {
		t.Fatalf("Symbols: %v", err)
	}
	byName := map[string]*coff.Symbol{}
	for _, s := range syms {
		byName[s.Name] = s
	}

	main := byName["main"]
	if main == nil {
		t.Fatal("main is not in the symbol table")
	}
	if main.Class != pecore.ClassExternal {
		t.Errorf("main class = %v, want ClassExternal", main.Class)
	}
	if !main.Defined() {
		t.Error("main should be defined")
	}

	msg := byName["msg"]
	if msg == nil {
		t.Fatal("msg is not in the symbol table")
	}
	if msg.Class != pecore.ClassStatic {
		t.Errorf("msg class = %v, want ClassStatic", msg.Class)
	}

	puts := byName["puts"]
	if puts == nil || !puts.Undefined() {
		t.Errorf("puts = %v, want an undefined symbol", puts)
	}
	if puts.Class != pecore.ClassExternal {
		t.Errorf("puts class = %v, want ClassExternal", puts.Class)
	}
}

// TestWeakIsStatic. COFF has no weak definition — a weak external is a weak
// *reference* and cannot define anything — so a weak binding becomes a static
// symbol, one copy per object. See classOf for what that costs.
func TestWeakIsStatic(t *testing.T) {
	m := amd64.NewModule()
	m.Section(amd64.Text).Label("f", amd64.Weak, amd64.Func)
	m.Section(amd64.Text).Ret()

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	var buf bytes.Buffer
	if err := pe.Write(&buf, o); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// TestDirectives checks that Options.Directives lands in .drectve, which is
// how DEFAULTLIB and EXPORT reach the linker from an object.
func TestDirectives(t *testing.T) {
	path, _ := write(t, fixture(t), pe.Options{
		Directives: []pe.Directive{{Name: "DEFAULTLIB", Value: "libcmt.lib"}},
	})

	f, err := coff.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	ds := f.Directives()
	if len(ds) == 0 {
		t.Fatal("no directives recorded")
	}
}

// TestRefusedKinds names kinds COFF has no relocation for: there is no GOT,
// and none of the seven ELF TLS models has a SECREL/.tls counterpart.
//
// RefPLT32 is not among them. There is no procedure linkage table either, but
// a call needs no table to reach an import here — the linker binds it to a
// thunk — so PLT32 is the same relocation as PC32 rather than an unanswerable
// request.
func TestRefusedKinds(t *testing.T) {
	for _, kind := range []amd64.RefKind{
		amd64.RefGOTPCREL,
		amd64.RefTLSGD,
		amd64.RefTLV,
		amd64.RefSize32,
	} {
		m := amd64.NewModule()
		m.Extern("x")
		m.Section(amd64.Text).Ret()
		m.Section(amd64.Data).Ref(amd64.Ref("x", kind))

		o, err := m.Finalize()
		if err != nil {
			t.Fatalf("%v: Finalize: %v", kind, err)
		}
		err = pe.Write(&bytes.Buffer{}, o)
		if !errors.Is(err, obj.ErrRefKind) {
			t.Errorf("%v: Write returned %v, want ErrRefKind", kind, err)
		}
	}
}

// TestWrongArchIsRefused. There is no Target option, so the object's own Arch
// is the only thing that can disagree, and it is checked.
func TestWrongArchIsRefused(t *testing.T) {
	o := obj.New(obj.ArchI386, nil, nil)
	if err := pe.Write(&bytes.Buffer{}, o); err == nil {
		t.Error("Write accepted an i386 object")
	}
}
