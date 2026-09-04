package amd64

import (
	"testing"

	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// sseText assembles one section and returns its bytes.
func sseText(t *testing.T, build func(s *Section)) []byte {
	t.Helper()

	m := NewModule()
	s := m.Section(Text)
	s.Label("f", Global, Func)
	build(s)
	s.EndLabel("f")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	return o.SectionNamed(".text").Bytes()
}

func checkBytes(t *testing.T, got, want []byte) {
	t.Helper()
	if string(got) != string(want) {
		t.Errorf("bytes:\n got % x\nwant % x", got, want)
	}
}

// The prefix byte is what makes these four different instructions out of
// one opcode: F3 is scalar single, F2 is scalar double, 66 is packed
// double, and none at all is packed single. Nothing else in this test is
// as easy to get wrong.
func TestSSEPrefixes(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.AddssXmmXM32(XMM0, XMM1)
		s.AddsdXmmXM64(XMM0, XMM1)
		s.AndpsXmmRM128(XMM0, XMM1)
		s.AndpdXmmRM128(XMM0, XMM1)
	})
	checkBytes(t, got, []byte{
		0xf3, 0x0f, 0x58, 0xc1, // addss xmm0, xmm1
		0xf2, 0x0f, 0x58, 0xc1, // addsd xmm0, xmm1
		0x0f, 0x54, 0xc1, // andps xmm0, xmm1
		0x66, 0x0f, 0x54, 0xc1, // andpd xmm0, xmm1
	})
}

// The scalar arithmetic block, both widths of each. One opcode per
// operation and the prefix says the width, which is what makes the table a
// loop rather than twelve rows.
func TestSSEScalarArithmetic(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.AddssXmmXM32(XMM0, XMM1)
		s.SubssXmmXM32(XMM0, XMM1)
		s.MulssXmmXM32(XMM0, XMM1)
		s.DivssXmmXM32(XMM0, XMM1)
		s.MinssXmmXM32(XMM0, XMM1)
		s.MaxssXmmXM32(XMM0, XMM1)
		s.SqrtssXmmXM32(XMM0, XMM1)

		s.AddsdXmmXM64(XMM0, XMM1)
		s.SubsdXmmXM64(XMM0, XMM1)
		s.MulsdXmmXM64(XMM0, XMM1)
		s.DivsdXmmXM64(XMM0, XMM1)
		s.MinsdXmmXM64(XMM0, XMM1)
		s.MaxsdXmmXM64(XMM0, XMM1)
		s.SqrtsdXmmXM64(XMM0, XMM1)
	})
	checkBytes(t, got, []byte{
		0xf3, 0x0f, 0x58, 0xc1,
		0xf3, 0x0f, 0x5c, 0xc1,
		0xf3, 0x0f, 0x59, 0xc1,
		0xf3, 0x0f, 0x5e, 0xc1,
		0xf3, 0x0f, 0x5d, 0xc1,
		0xf3, 0x0f, 0x5f, 0xc1,
		0xf3, 0x0f, 0x51, 0xc1,

		0xf2, 0x0f, 0x58, 0xc1,
		0xf2, 0x0f, 0x5c, 0xc1,
		0xf2, 0x0f, 0x59, 0xc1,
		0xf2, 0x0f, 0x5e, 0xc1,
		0xf2, 0x0f, 0x5d, 0xc1,
		0xf2, 0x0f, 0x5f, 0xc1,
		0xf2, 0x0f, 0x51, 0xc1,
	})
}

// Memory operands at scalar widths, which is the whole reason XM32 and XM64
// exist: MOVSS's memory operand is four bytes, not the sixteen an
// xmm/m128 row would have accepted.
func TestSSEMemoryOperands(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.MovssXmmXM32(XMM0, operand.Mem32(reg.RDI))
		s.MovssXM32Xmm(operand.Mem32(reg.RDI).Disp(8), XMM3)
		s.MovsdXmmXM64(XMM1, operand.Mem64(reg.RSP))
		s.MovupsXmmRM128(XMM4, operand.Mem128(reg.RAX))
	})
	checkBytes(t, got, []byte{
		0xf3, 0x0f, 0x10, 0x07, // movss xmm0, [rdi]
		0xf3, 0x0f, 0x11, 0x5f, 0x08, // movss [rdi+8], xmm3
		0xf2, 0x0f, 0x10, 0x0c, 0x24, // movsd xmm1, [rsp]
		0x0f, 0x10, 0x20, // movups xmm4, [rax]
	})
}

// The extended vector registers, which reach the encoder the same way the
// extended general-purpose ones do: through REX, and through the same two
// bits, because ModRM does not care which register file it is naming.
func TestSSEExtendedRegisters(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.MovsdXmmXM64(XMM9, operand.Mem64(reg.R12Q))
		s.MovsdXM64Xmm(operand.Mem64(reg.RSP), XMM11)
		s.XorpdXmmRM128(XMM15, XMM15)
		s.MovqRM64Xmm(reg.RAX, XMM8)
	})
	checkBytes(t, got, []byte{
		0xf2, 0x45, 0x0f, 0x10, 0x0c, 0x24, // movsd xmm9, [r12]
		0xf2, 0x44, 0x0f, 0x11, 0x1c, 0x24, // movsd [rsp], xmm11
		0x66, 0x45, 0x0f, 0x57, 0xff, // xorpd xmm15, xmm15
		0x66, 0x4c, 0x0f, 0x7e, 0xc0, // movq rax, xmm8
	})
}

// MOVD and MOVQ are one opcode pair told apart by REX.W, which is the same
// thing that tells their names apart. Both directions, because moving a bit
// pattern into a vector register and reading one back out are different
// opcodes and not a swapped operand order.
func TestSSEMoveAcrossRegisterFiles(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.MovdXmmRM32(XMM0, reg.EDI)
		s.MovdRM32Xmm(reg.EAX, XMM1)
		s.MovqXmmRM64(XMM0, reg.RDI)
		s.MovqRM64Xmm(reg.RAX, XMM1)
	})
	checkBytes(t, got, []byte{
		0x66, 0x0f, 0x6e, 0xc7, // movd xmm0, edi
		0x66, 0x0f, 0x7e, 0xc8, // movd eax, xmm1
		0x66, 0x48, 0x0f, 0x6e, 0xc7, // movq xmm0, rdi
		0x66, 0x48, 0x0f, 0x7e, 0xc8, // movq rax, xmm1
	})
}

// Every conversion, and the two pairs that are easy to confuse. CVTSI2SD
// and CVTSI2SS differ only in the prefix; CVTTSD2SI and CVTSD2SI differ
// only in one letter of the name and in whether they truncate or round.
func TestSSEConversions(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.Cvtsi2ssXmmRM32(XMM0, reg.EDI)
		s.Cvtsi2ssXmmRM64(XMM0, reg.RDI)
		s.Cvtsi2sdXmmRM32(XMM0, reg.EDI)
		s.Cvtsi2sdXmmRM64(XMM0, reg.RDI)

		s.Cvttss2siR32XM32(reg.EAX, XMM0)
		s.Cvttss2siR64XM32(reg.RAX, XMM0)
		s.Cvttsd2siR32XM64(reg.EAX, XMM0)
		s.Cvttsd2siR64XM64(reg.RAX, XMM0)

		s.Cvtss2siR32XM32(reg.EAX, XMM0)
		s.Cvtss2siR64XM32(reg.RAX, XMM0)
		s.Cvtsd2siR32XM64(reg.EAX, XMM0)
		s.Cvtsd2siR64XM64(reg.RAX, XMM0)

		s.Cvtss2sdXmmXM32(XMM0, XMM1)
		s.Cvtsd2ssXmmXM64(XMM0, XMM1)
	})
	checkBytes(t, got, []byte{
		0xf3, 0x0f, 0x2a, 0xc7, // cvtsi2ss xmm0, edi
		0xf3, 0x48, 0x0f, 0x2a, 0xc7, // cvtsi2ss xmm0, rdi
		0xf2, 0x0f, 0x2a, 0xc7, // cvtsi2sd xmm0, edi
		0xf2, 0x48, 0x0f, 0x2a, 0xc7, // cvtsi2sd xmm0, rdi

		0xf3, 0x0f, 0x2c, 0xc0, // cvttss2si eax, xmm0
		0xf3, 0x48, 0x0f, 0x2c, 0xc0, // cvttss2si rax, xmm0
		0xf2, 0x0f, 0x2c, 0xc0, // cvttsd2si eax, xmm0
		0xf2, 0x48, 0x0f, 0x2c, 0xc0, // cvttsd2si rax, xmm0

		0xf3, 0x0f, 0x2d, 0xc0, // cvtss2si eax, xmm0
		0xf3, 0x48, 0x0f, 0x2d, 0xc0, // cvtss2si rax, xmm0
		0xf2, 0x0f, 0x2d, 0xc0, // cvtsd2si eax, xmm0
		0xf2, 0x48, 0x0f, 0x2d, 0xc0, // cvtsd2si rax, xmm0

		0xf3, 0x0f, 0x5a, 0xc1, // cvtss2sd xmm0, xmm1
		0xf2, 0x0f, 0x5a, 0xc1, // cvtsd2ss xmm0, xmm1
	})
}

// The compares and the whole-register moves.
func TestSSECompareAndMove(t *testing.T) {
	got := sseText(t, func(s *Section) {
		s.UcomissXmmXM32(XMM0, XMM1)
		s.UcomisdXmmXM64(XMM0, XMM1)
		s.ComissXmmXM32(XMM0, XMM1)
		s.ComisdXmmXM64(XMM0, XMM1)
		s.MovapsXmmRM128(XMM4, XMM5)
		s.MovapdXmmRM128(XMM4, XMM5)
	})
	checkBytes(t, got, []byte{
		0x0f, 0x2e, 0xc1, // ucomiss xmm0, xmm1
		0x66, 0x0f, 0x2e, 0xc1, // ucomisd xmm0, xmm1
		0x0f, 0x2f, 0xc1, // comiss xmm0, xmm1
		0x66, 0x0f, 0x2f, 0xc1, // comisd xmm0, xmm1
		0x0f, 0x28, 0xe5, // movaps xmm4, xmm5
		0x66, 0x0f, 0x28, 0xe5, // movapd xmm4, xmm5
	})
}
