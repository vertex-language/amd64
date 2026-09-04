package amd64

import (
	"errors"
	"testing"

	"github.com/vertex-language/amd64/feature"
	"github.com/vertex-language/amd64/internal/isa"
	"github.com/vertex-language/amd64/obj"
)

// TestTableBindsEveryHelper is the other half of the promise inst.go makes.
//
// form() panics at package initialisation on a helper with no row, so a
// missing row already fails whatever the caller was about to do — but only
// once something imports this package. This test is that something, and it
// also checks the direction form() cannot: a row nobody binds, which is a
// form no typed helper can reach.
func TestTableBindsEveryHelper(t *testing.T) {
	for _, f := range isa.All() {
		if f.Helper == "" {
			t.Fatalf("%s has no helper name", f)
		}
		if got := isa.ByHelper(f.Helper); got != f {
			t.Errorf("ByHelper(%q) did not return the row that declared it", f.Helper)
		}
	}
}

// TestEncode walks one module of every path a writer will meet: a RIP-relative
// reference, a call through the PLT, a based displacement, a same-section
// branch, nop padding, and an absolute reference placed on the data side.
func TestEncode(t *testing.T) {
	m := NewModule()
	m.Extern("puts")

	s := m.Section(Text)
	s.Label("main", Global, Func)
	s.LeaR64M(RDI, Rip(Ref("msg", RefPC32)))
	s.CallRef(Ref("puts", RefPLT32))
	s.Label("again")
	s.MovRM64Imm32(Mem64(RSP).Disp(8), 0)
	s.JmpLabel("again")
	s.Ret()
	s.EndLabel("main")

	r := m.SectionNamed(".rodata", ROData)
	r.Label("msg", Local, ObjectSym)
	r.Asciz("hi")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	text := o.SectionNamed(".text")
	want := []byte{
		0x48, 0x8d, 0x3d, 0, 0, 0, 0, // lea rdi, [rip+msg]
		0xe8, 0, 0, 0, 0, // call puts
		0x48, 0xc7, 0x44, 0x24, 0x08, 0, 0, 0, 0, // mov qword [rsp+8], 0
		0xe9, 0xf2, 0xff, 0xff, 0xff, // jmp again
		0xc3, // ret
	}
	if got := text.Bytes(); string(got) != string(want) {
		t.Errorf("text bytes:\n got % x\nwant % x", got, want)
	}

	refs := text.Refs()
	if len(refs) != 2 {
		t.Fatalf("want 2 references in .text, got %d", len(refs))
	}
	for i, w := range []struct {
		off  int
		sym  string
		kind RefKind
	}{{3, "msg", RefPC32}, {8, "puts", RefPLT32}} {
		r := refs[i]
		if r.Offset != w.off || r.Sym != w.sym || r.Kind != w.kind {
			t.Errorf("ref %d = %+v, want offset %d %s %v", i, r, w.off, w.sym, w.kind)
		}
		// Every PC-relative field here is the last thing in its instruction,
		// so the correction is exactly the field's own width.
		if !r.PCRel || r.Adjust != -4 {
			t.Errorf("ref %d: want PCRel with Adjust -4, got %v %d", i, r.PCRel, r.Adjust)
		}
	}

	if sym, ok := o.Symbol("main"); !ok || sym.Size != len(want) {
		t.Errorf("main: got %+v, want size %d", sym, len(want))
	}
	if sym, ok := o.Symbol("puts"); !ok || sym.Defined() {
		t.Errorf("puts: want an undefined symbol, got %+v", sym)
	}
}

// TestOpRegExtendsREXB is a regression test for a real encoding bug: every
// +rb/+rw/+rd form (the register rides in the opcode's own low three bits,
// not in a ModRM byte — MOV r*, imm and PUSH/POP r64 are this tree's only
// examples) extends that register through REX.B, because there is no
// ModRM.reg field for REX.R to extend instead. The encoder used to set
// REX.R unconditionally whenever the gathered register was extended,
// which is correct for an ordinary ModRM.reg operand but silently wrong
// here: R8-R15 through one of these forms encoded as R0-R7 with a
// pointless REX.R bit that a real CPU ignores for this opcode shape,
// producing a working instruction that named the wrong register. Nothing
// existing caught it — internal/encode had no tests of its own, and
// TestEncode above never happens to use an extended register with one of
// these forms.
func TestOpRegExtendsREXB(t *testing.T) {
	m := NewModule()
	s := m.Section(Text)
	s.Label("f", Global, Func)
	s.MovR32Imm32(R9D, 1)
	s.MovR64Imm64(R12, 2)
	s.MovR8Imm8(R8B, 3)
	s.PushR64(R13)
	s.PopR64(R14)
	s.Ret()
	s.EndLabel("f")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	want := []byte{
		0x41, 0xb9, 0x01, 0x00, 0x00, 0x00, // mov r9d, 1
		0x49, 0xbc, 2, 0, 0, 0, 0, 0, 0, 0, // mov r12, 2
		0x41, 0xb0, 0x03, // mov r8b, 3
		0x41, 0x55, // push r13
		0x41, 0x5e, // pop r14
		0xc3, // ret
	}
	if got := o.SectionNamed(".text").Bytes(); string(got) != string(want) {
		t.Errorf("text bytes:\n got % x\nwant % x", got, want)
	}
}

// TestFinalizeIsIdempotent is the claim Finalize's doc comment makes.
func TestFinalizeIsIdempotent(t *testing.T) {
	m := NewModule()
	m.Section(Text).Ret()

	first, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	second, err := m.Finalize()
	if err != nil || second != first {
		t.Errorf("second Finalize returned %p, %v; want %p, nil", second, err, first)
	}
}

// TestErrorsAreSticky checks that a failure stops the build and surfaces
// through Finalize positioned, rather than being overwritten by a later one.
func TestErrorsAreSticky(t *testing.T) {
	m := NewModule()
	s := m.Section(Text)
	s.Align(3) // not a power of two
	s.Align(5) // must not replace the first error
	s.Ret()

	err := m.Err()
	if !errors.Is(err, ErrAlign) {
		t.Fatalf("Err() = %v, want ErrAlign", err)
	}
	var e *Error
	if !errors.As(err, &e) || !e.Positioned() {
		t.Fatalf("want a positioned *amd64.Error, got %#v", err)
	}
	if _, ferr := m.Finalize(); ferr != err {
		t.Errorf("Finalize returned %v, want the first error", ferr)
	}
}

// TestParseLevels pins the canonical level spellings. The hyphen in
// "x86-64-v3" is part of the name and not a removal, and so is the one in
// "aes-ni" — both prefixes are themselves declared spellings, which is what
// makes longest match the rule rather than a prefix test.
func TestParseLevels(t *testing.T) {
	for _, tc := range []struct {
		spec string
		want Level
	}{
		{"x86-64", V1},
		{"x86-64-v2", V2},
		{"x86-64-v3", V3},
		{"x86-64-v4", V4},
		{"x86_64_v3", V3},
		{"x86-64-v3+aes", V3},
	} {
		got, err := ParseFeatures(DefaultFeatures(), tc.spec)
		if err != nil {
			t.Errorf("ParseFeatures(%q): %v", tc.spec, err)
			continue
		}
		if got.Level() != tc.want {
			t.Errorf("ParseFeatures(%q).Level() = %v, want %v", tc.spec, got.Level(), tc.want)
		}
	}

	// A hyphen that extends no known name is a removal.
	got, err := ParseFeatures(DefaultFeatures(), "x86-64-v3-avx2")
	if err != nil {
		t.Fatalf("ParseFeatures: %v", err)
	}
	if got.Has(AVX2) {
		t.Error("x86-64-v3-avx2 kept AVX2")
	}
	if !got.Has(feature.SSE42) {
		t.Error("x86-64-v3-avx2 dropped more than it named")
	}

	// And a hyphenated feature name survives being one.
	if _, err := ParseFeatures(DefaultFeatures(), "+aes-ni"); err != nil {
		t.Errorf("ParseFeatures(\"+aes-ni\"): %v", err)
	}
}

// The bit-counting and byte-swap rows, and the two facts about them that
// are easy to get wrong.
//
// BSWAP takes its register in the opcode, so an extended register
// extends through REX.B and not REX.R — bswapq %r12 is 49 0F CC, and
// getting that bit wrong would name %r4. And TZCNT is gated because
// without BMI1 the same bytes decode as BSF, which agrees with it on
// every non-zero operand and disagrees on zero.
func TestBitCountingAndByteSwap(t *testing.T) {
	m := NewModule(WithFeatures(feature.NewSet(feature.V3)))
	s := m.Section(Text)
	s.Label("f", Global, Func)
	s.PopcntR32RM32(EAX, ECX)
	s.LzcntR64RM64(RAX, RCX)
	s.TzcntR32RM32(EAX, ECX)
	s.BswapR32(EAX)
	s.BswapR64(R12)
	s.BswapR32(R9D)
	s.Ret()
	s.EndLabel("f")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	want := []byte{
		0xf3, 0x0f, 0xb8, 0xc1, // popcnt eax, ecx
		0xf3, 0x48, 0x0f, 0xbd, 0xc1, // lzcnt rax, rcx
		0xf3, 0x0f, 0xbc, 0xc1, // tzcnt eax, ecx
		0x0f, 0xc8, // bswap eax
		0x49, 0x0f, 0xcc, // bswap r12
		0x41, 0x0f, 0xc9, // bswap r9d
		0xc3, // ret
	}
	if got := o.SectionNamed(".text").Bytes(); string(got) != string(want) {
		t.Errorf("text bytes:\n got % x\nwant % x", got, want)
	}

	// The gate is real: the same TZCNT on a baseline module is refused.
	base := NewModule()
	b := base.Section(Text)
	b.Label("g", Global, Func)
	b.TzcntR32RM32(EAX, ECX)
	b.Ret()
	b.EndLabel("g")
	if _, err := base.Finalize(); !errors.Is(err, obj.ErrFeature) {
		t.Errorf("baseline tzcnt = %v, want ErrFeature", err)
	}
}

// The privileged and serializing instructions, whose bytes are clang's for
// the same eight lines. Nothing in this tree selects one — they are here for
// inline assembly — so this is the only place the typed door onto them is
// walked through.
func TestPrivilegedNoOperandInstructions(t *testing.T) {
	m := NewModule()
	s := m.Section(Text)
	s.Label("f", Global, Func)
	s.Hlt()
	s.Cli()
	s.Sti()
	s.Pause()
	s.Rdtsc()
	s.Rdmsr()
	s.Wrmsr()
	s.Wbinvd()
	s.Ret()
	s.EndLabel("f")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	want := []byte{
		0xf4,       // hlt
		0xfa,       // cli
		0xfb,       // sti
		0xf3, 0x90, // pause
		0x0f, 0x31, // rdtsc
		0x0f, 0x32, // rdmsr
		0x0f, 0x30, // wrmsr
		0x0f, 0x09, // wbinvd
		0xc3, // ret
	}
	if got := o.SectionNamed(".text").Bytes(); string(got) != string(want) {
		t.Errorf("text bytes:\n got % x\nwant % x", got, want)
	}
}
