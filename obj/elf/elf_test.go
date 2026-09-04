package elf_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	elfcore "github.com/vertex-language/elf"
	elfobj "github.com/vertex-language/elf/obj"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/obj/elf"
)

// fixture is the module every test here writes: a RIP-relative reference, a
// call through the PLT, a same-section branch, an undefined symbol, and a
// .rodata section the first reference names.
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

// write emits o to a temp file and hands back both the path and the bytes.
func write(t *testing.T, o *obj.Object, opts ...elf.Options) (string, []byte) {
	t.Helper()

	var buf bytes.Buffer
	if err := elf.Write(&buf, o, opts...); err != nil {
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
	path, _ := write(t, o)

	f, err := elfobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	if got := f.Target(); got.Arch != elfcore.ArchAMD64 || got.Class != elfcore.ELFCLASS64 {
		t.Errorf("target = %v, want x86-64/ELFCLASS64", got)
	}

	text := f.Section(".text")
	if text == nil {
		t.Fatal("no .text in the written object")
	}
	if !text.Alloc() || !text.Executable() || text.Writable() {
		t.Errorf(".text flags = %#x, want alloc+exec and not writable", text.Flags)
	}

	// RELA means the bytes go through untouched. This is the half of the
	// bargain the other two writers do not get.
	data, err := text.Data()
	if err != nil {
		t.Fatal(err)
	}
	if want := o.SectionNamed(".text").Bytes(); !bytes.Equal(data, want) {
		t.Errorf(".text bytes:\n got % x\nwant % x", data, want)
	}

	rodata := f.Section(".rodata")
	if rodata == nil {
		t.Fatal("no .rodata in the written object")
	}
	if rodata.Executable() || rodata.Writable() {
		t.Errorf(".rodata flags = %#x, want alloc only", rodata.Flags)
	}
}

// TestRelocs pins the two mappings the fixture exercises and the addend rule
// that produces them.
func TestRelocs(t *testing.T) {
	path, _ := write(t, fixture(t))

	f, err := elfobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	relocs, err := f.Section(".text").Relocs()
	if err != nil {
		t.Fatalf("Relocs: %v", err)
	}
	if len(relocs) != 2 {
		t.Fatalf("want 2 relocations, got %d", len(relocs))
	}

	for i, want := range []struct {
		off    uint64
		sym    string
		typ    elfcore.RelocX86_64
		addend int64
	}{
		// Every PC-relative field in the fixture is the last thing in its
		// instruction, so Adjust is -4 and the written addend is -4.
		{3, "msg", elfcore.R_X86_64_PC32, -4},
		{8, "puts", elfcore.R_X86_64_PLT32, -4},
	} {
		got := relocs[i]
		if !got.Explicit {
			t.Errorf("reloc %d is not RELA; x86-64 carries its addend in the entry", i)
		}
		if got.Offset != want.off {
			t.Errorf("reloc %d offset = %#x, want %#x", i, got.Offset, want.off)
		}
		if elfcore.RelocX86_64(got.Type) != want.typ {
			t.Errorf("reloc %d type = %v, want %v", i, elfcore.RelocX86_64(got.Type), want.typ)
		}
		if got.Addend != want.addend {
			t.Errorf("reloc %d addend = %d, want %d", i, got.Addend, want.addend)
		}
		if got.Sym == nil || got.Sym.Name != want.sym {
			t.Errorf("reloc %d symbol = %v, want %q", i, got.Sym, want.sym)
		}
	}
}

// TestSymbols checks that the vocabulary survives the crossing. ELF is the
// container that carries all of it, so nothing here should be dropped.
func TestSymbols(t *testing.T) {
	path, _ := write(t, fixture(t))

	f, err := elfobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	syms, err := f.Symbols()
	if err != nil {
		t.Fatalf("Symbols: %v", err)
	}
	byName := map[string]*elfobj.Symbol{}
	for _, s := range syms {
		byName[s.Name] = s
	}

	main := byName["main"]
	if main == nil {
		t.Fatal("main is not in the symbol table")
	}
	if main.Bind != elfcore.STB_GLOBAL || main.Type != elfcore.STT_FUNC {
		t.Errorf("main = %v/%v, want STB_GLOBAL/STT_FUNC", main.Bind, main.Type)
	}
	if main.Size == 0 {
		t.Error("main has no size; EndLabel should have closed it")
	}

	msg := byName["msg"]
	if msg == nil {
		t.Fatal("msg is not in the symbol table")
	}
	if msg.Bind != elfcore.STB_LOCAL || msg.Type != elfcore.STT_OBJECT {
		t.Errorf("msg = %v/%v, want STB_LOCAL/STT_OBJECT", msg.Bind, msg.Type)
	}

	puts := byName["puts"]
	if puts == nil || !puts.Undefined() {
		t.Errorf("puts = %v, want an undefined symbol", puts)
	}
}

// TestStripLocalsKeepsReferenced is the one rule that flag has: a local a
// relocation names is kept regardless, because a stripped object that does
// not link is worse than a slightly larger one.
func TestStripLocalsKeepsReferenced(t *testing.T) {
	m := amd64.NewModule()
	s := m.Section(amd64.Text)
	s.Label("kept", amd64.Local)
	s.LeaR64M(amd64.RDI, amd64.Rip(amd64.Ref("kept", amd64.RefPC32)))
	s.Label("dropped", amd64.Local)
	s.Ret()

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	path, _ := write(t, o, elf.Options{StripLocals: true})

	f, err := elfobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	var kept, dropped bool
	for _, s := range syms {
		switch s.Name {
		case "kept":
			kept = true
		case "dropped":
			dropped = true
		}
	}
	if !kept {
		t.Error("StripLocals dropped a local a relocation names")
	}
	if dropped {
		t.Error("StripLocals kept a local nothing names")
	}
}

// TestComment is the producer string, and it lands in a section that is not
// allocated at run time.
func TestComment(t *testing.T) {
	path, _ := write(t, fixture(t), elf.Options{Comment: "vertex 0.4"})

	f, err := elfobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	c := f.Section(".comment")
	if c == nil {
		t.Fatal("no .comment section")
	}
	if c.Alloc() {
		t.Error(".comment is SHF_ALLOC; it does not belong in the image")
	}
	data, err := c.Data()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "vertex 0.4\x00" {
		t.Errorf(".comment = %q, want %q", data, "vertex 0.4\x00")
	}
}

// TestDebugSectionIsNotAllocated is the one heuristic in the package.
func TestDebugSectionIsNotAllocated(t *testing.T) {
	m := amd64.NewModule()
	m.Section(amd64.Text).Ret()
	m.SectionNamed(".debug_line", amd64.Data).Data([]byte{1, 2, 3, 4})

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	path, _ := write(t, o)

	f, err := elfobj.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	d := f.Section(".debug_line")
	if d == nil {
		t.Fatal("no .debug_line section")
	}
	if d.Alloc() {
		t.Error(".debug_line is SHF_ALLOC; debug info does not belong in the image")
	}
}

// TestRefusedKinds names the four kinds ELF has no relocation for. They are
// the other two containers' ideas, and refusing them here is what lets the
// object stay legal for those.
func TestRefusedKinds(t *testing.T) {
	for _, kind := range []amd64.RefKind{
		amd64.RefTLV,
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
		err = elf.Write(&bytes.Buffer{}, o)
		if !errors.Is(err, obj.ErrRefKind) {
			t.Errorf("%v: Write returned %v, want ErrRefKind", kind, err)
		}
	}
}

// TestWrongArchIsRefused. There is no Target option, so the object's own Arch
// is the only thing that can disagree, and it is checked.
func TestWrongArchIsRefused(t *testing.T) {
	o := obj.New(obj.ArchI386, nil, nil)
	if err := elf.Write(&bytes.Buffer{}, o); err == nil {
		t.Error("Write accepted an i386 object")
	}
}
