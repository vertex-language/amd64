package amd64_test

import (
	"bytes"
	"testing"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// An offset kind may share a displacement field with a base register, where
// an address kind may not: the difference is that an offset is the number
// the base wants added to it, and no linker has to know the base's run-time
// value to work out an addend.
//
// It is what a thread-local access is made of. This is COFF's, which clang
// writes as
//
//	movq  %gs:88, %rax
//	movl  _tls_index(%rip), %ecx
//	movq  (%rax,%rcx,8), %rax
//	leaq  v@SECREL32(%rax), %rax
//
// and the ELF forms are the same shape against a thread pointer.
func TestOffsetRefSharesTheDisplacement(t *testing.T) {
	m := amd64.NewModule()
	m.Extern("_tls_index")
	m.Extern("v")
	s := m.Section(amd64.Text)
	s.MovR64RM64(reg.RAX, operand.Addr64(0x58).Seg(reg.GS))
	s.MovR32RM32(reg.ECX, operand.Rip32(amd64.Ref("_tls_index", amd64.RefPC32)))
	s.MovR64RM64(reg.RAX, operand.Mem64(reg.RAX).Index(reg.RCX, 8))
	s.LeaR64M(reg.RAX, operand.Mem64(reg.RAX).Sym(amd64.Ref("v", amd64.RefSecRel32)))

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	var text *obj.Section
	for _, sec := range o.Sections() {
		if sec.Name() == ".text" {
			text = sec
		}
	}
	if text == nil {
		t.Fatal("no .text")
	}

	want := []byte{
		0x65, 0x48, 0x8b, 0x04, 0x25, 0x58, 0, 0, 0, // mov rax, gs:[0x58]
		0x8b, 0x0d, 0, 0, 0, 0, //                      mov ecx, [rip+_tls_index]
		0x48, 0x8b, 0x04, 0xc8, //                      mov rax, [rax+rcx*8]
		0x48, 0x8d, 0x80, 0, 0, 0, 0, //                lea rax, [rax+secrel32(v)]
	}
	if got := text.Bytes(); !bytes.Equal(got, want) {
		t.Errorf(".text = % x\n    want % x", got, want)
	}

	// The last instruction's displacement is a hole, and it is four bytes
	// even though the offset will be small: nothing here knows it, and a
	// byte-sized field is one the linker could not widen.
	var secrel bool
	for _, r := range text.Refs() {
		if r.Kind == amd64.RefSecRel32 && r.Sym == "v" {
			secrel = true
			if r.Offset != 22 {
				t.Errorf("secrel32 hole at %d, want 22", r.Offset)
			}
		}
	}
	if !secrel {
		t.Error("no secrel32 reference to v")
	}
}

// An address kind is still refused on a based address, which is the rule the
// offset kinds are the exception to.
func TestAddressRefOnABaseIsRefused(t *testing.T) {
	m := amd64.NewModule()
	m.Extern("v")
	s := m.Section(amd64.Text)
	s.LeaR64M(reg.RAX, operand.Mem64(reg.RAX).Sym(amd64.Ref("v", amd64.RefAbs32)))
	if _, err := m.Finalize(); err == nil {
		t.Fatal("an abs32 reference shared a displacement with a base register")
	}
}
