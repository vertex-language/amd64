package amd64

import (
	"encoding/binary"
	"testing"

	"github.com/vertex-language/amd64/obj"
)

// TestSymDelta: a relative pointer. Both ends in the field's section fold to
// the distance and leave the linker nothing; a minuend anywhere else becomes a
// PC-relative reference to it alone, with the distance from the field back to
// the subtrahend in the addend -- which is the same value by the Reference
// identity, and the one form COFF, ELF and Mach-O can all write.
func TestSymDelta(t *testing.T) {
	m := NewModule()
	m.Extern("far")

	r := m.SectionNamed(".rdata", ROData)
	r.Label("desc", Global, ObjectSym)
	r.Long(0)
	r.SymDelta("name", "desc", -4) // name - (desc+4): near
	r.SymDelta("far", "desc", -8)  // far - (desc+8): in another object
	r.Label("name", Local, ObjectSym)
	r.Asciz("x")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	s := o.SectionNamed(".rdata")
	b := s.Bytes()
	if got := int32(binary.LittleEndian.Uint32(b[4:])); got != 8 {
		t.Errorf("the near field holds %d, want 8: name is at 12, the field at 4", got)
	}
	refs := s.Refs()
	if len(refs) != 1 {
		t.Fatalf("%d references, want 1", len(refs))
	}
	want := obj.Reference{Offset: 8, Size: 4, PCRel: true, Sym: "far", Kind: obj.RefPC32, Addend: 0}
	if refs[0] != want {
		t.Errorf("reference is %+v, want %+v", refs[0], want)
	}
}

// TestSymDeltaRefusesAForeignSubtrahend: a subtrahend outside the field's
// section is a difference no PC-relative form expresses, and it is refused
// rather than written as something else.
func TestSymDeltaRefusesAForeignSubtrahend(t *testing.T) {
	m := NewModule()
	d := m.Section(Data)
	d.Label("a", Global, ObjectSym)
	d.Long(0)
	r := m.SectionNamed(".rdata", ROData)
	r.Label("b", Global, ObjectSym)
	r.SymDelta("b", "a", 0)
	if _, err := m.Finalize(); err == nil {
		t.Fatal("Finalize accepted a subtrahend in another section")
	}
}
