package amd64

import (
	"errors"
	"testing"

	"github.com/vertex-language/amd64/feature"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// One prefix byte in front, and everything else unchanged. That is what a
// locking clone is, and generating them from the table rather than
// writing them out is what keeps the two copies from disagreeing.
func TestLockPrefix(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.LockAddRM64R64(operand.Mem64(reg.RDI), reg.RSI)
		s.AddRM64R64(operand.Mem64(reg.RDI), reg.RSI)
		s.LockAddRM32Imm8(operand.Mem32(reg.RDI), 1)
		s.LockIncRM64(operand.Mem64(reg.RDI))
		s.LockAndRM8Imm8(operand.Mem8(reg.RDI), 3)
	})
	checkBytes(t, got, []byte{
		0xf0, 0x48, 0x01, 0x37, // lock add [rdi], rsi
		0x48, 0x01, 0x37, // add [rdi], rsi        — the same three bytes
		0xf0, 0x83, 0x07, 0x01, // lock add dword [rdi], 1
		0xf0, 0x48, 0xff, 0x07, // lock inc qword [rdi]
		0xf0, 0x80, 0x27, 0x03, // lock and byte [rdi], 3
	})
}

// LOCK comes before every other prefix, which is what a decoder expects:
// LOCK is group 1 and the segment override is group 2, and REX is always
// last before the opcode.
func TestLockPrefixOrder(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.LockCmpxchgRM64R64(operand.Mem64(reg.R12Q).Disp(8), reg.RCX)
		s.LockXaddRM64R64(operand.Mem64(reg.RDI), reg.RSI)
	})
	checkBytes(t, got, []byte{
		0xf0, 0x49, 0x0f, 0xb1, 0x4c, 0x24, 0x08, // lock cmpxchg [r12+8], rcx
		0xf0, 0x48, 0x0f, 0xc1, 0x37, // lock xadd [rdi], rsi
	})
}

// The three fences, whose ModRM byte addresses nothing: MFENCE is 0F AE
// F0 and the F0 is not naming a register.
func TestFences(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.Mfence()
		s.Lfence()
		s.Sfence()
	})
	checkBytes(t, got, []byte{
		0x0f, 0xae, 0xf0,
		0x0f, 0xae, 0xe8,
		0x0f, 0xae, 0xf8,
	})
}

// A register destination is refused at the call. The operand kinds were
// right — the class is r/m64 either way — so it is the value that has no
// encoding, which is ErrOperand and not ErrForm.
func TestLockRefusesARegisterDestination(t *testing.T) {
	m := NewModule()
	s := m.Section(Text)
	s.Label("f", Global, Func)
	s.LockAddRM64R64(reg.RDI, reg.RSI)
	s.Ret()
	s.EndLabel("f")

	_, err := m.Finalize()
	if !errors.Is(err, obj.ErrOperand) {
		t.Fatalf("Finalize = %v, want ErrOperand", err)
	}
	// The note has to name the form the caller can actually use, which
	// is the base row and not this one.
	if got := err.Error(); !contains(got, "the unlocked form is ADD r/m64, r64") {
		t.Errorf("error = %q, want it to name the unlocked form", got)
	}
}

// A clone inherits its base row's gate, because it is the same
// instruction: CMPXCHG16B needs CX16 whether or not the prefix is there.
func TestLockClonesInheritTheirGate(t *testing.T) {
	build := func(m *Module) error {
		s := m.Section(Text)
		s.Label("f", Global, Func)
		s.LockCmpxchg16bM128(operand.Mem128(reg.RDI))
		s.Ret()
		s.EndLabel("f")
		_, err := m.Finalize()
		return err
	}

	if err := build(NewModule()); !errors.Is(err, obj.ErrFeature) {
		t.Errorf("baseline module = %v, want ErrFeature", err)
	}
	with := NewModule(WithFeatures(feature.Default().Add(feature.CX16)))
	if err := build(with); err != nil {
		t.Errorf("module with cx16 = %v, want it to assemble", err)
	}
}

// The narrow atomics, which the compare-and-swap and fetch-and-add rows
// were missing: a lowering of an eight- or sixteen-bit atomic has these
// and does not have to widen to reach one.
func TestNarrowAtomics(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.LockXaddRM8R8(operand.Mem8(reg.RDI), reg.CL)
		s.LockXaddRM16R16(operand.Mem16(reg.RDI), reg.CX)
		s.LockCmpxchgRM16R16(operand.Mem16(reg.RDI), reg.CX)
		s.XchgRM16R16(operand.Mem16(reg.RDI), reg.CX)
	})
	checkBytes(t, got, []byte{
		0xf0, 0x0f, 0xc0, 0x0f, // lock xadd byte [rdi], cl
		0xf0, 0x66, 0x0f, 0xc1, 0x0f, // lock xadd word [rdi], cx
		0xf0, 0x66, 0x0f, 0xb1, 0x0f, // lock cmpxchg word [rdi], cx
		0x66, 0x87, 0x0f, // xchg word [rdi], cx
	})
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
