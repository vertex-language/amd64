package amd64_test

import (
	"encoding/hex"
	"testing"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// SSE2's packed integer forms, each against the bytes ml64 produces for the
// same instruction.
//
// The wants here are not derived from the table they check. Every line was
// assembled by Microsoft's macro assembler from the MASM text in the second
// column and read back out with dumpbin /disasm, so a wrong opcode, a
// swapped ModRM role or a mandatory prefix in the wrong place is a
// disagreement with the assembler the platform ships rather than with a
// second reading of the manual by the same pair of eyes.
//
// asm/difftest_test.go does the same job continuously against clang for the
// whole instruction set. This is the fixed record for the machine that has
// no clang on it, which on Windows is most of them.
func TestSIMDMatchesMASM(t *testing.T) {
	x, y := reg.XMM1, reg.XMM2
	mem := operand.Mem128(reg.RAX)

	for _, c := range []struct {
		masm string
		emit func(s *amd64.Section)
		want string
	}{
		{"movdqa xmm1, xmm2", func(s *amd64.Section) { s.MovdqaXmmRM128(x, y) }, "660f6fca"},
		{"movdqa xmmword ptr [rax], xmm1", func(s *amd64.Section) { s.MovdqaRM128Xmm(mem, x) }, "660f7f08"},
		{"movdqu xmm1, xmm2", func(s *amd64.Section) { s.MovdquXmmRM128(x, y) }, "f30f6fca"},
		{"movdqu xmmword ptr [rax], xmm1", func(s *amd64.Section) { s.MovdquRM128Xmm(mem, x) }, "f30f7f08"},
		{"paddb xmm1, xmm2", func(s *amd64.Section) { s.PaddbXmmRM128(x, y) }, "660ffcca"},
		{"paddw xmm1, xmm2", func(s *amd64.Section) { s.PaddwXmmRM128(x, y) }, "660ffdca"},
		{"paddd xmm1, xmm2", func(s *amd64.Section) { s.PadddXmmRM128(x, y) }, "660ffeca"},
		{"paddq xmm1, xmm2", func(s *amd64.Section) { s.PaddqXmmRM128(x, y) }, "660fd4ca"},
		{"psubb xmm1, xmm2", func(s *amd64.Section) { s.PsubbXmmRM128(x, y) }, "660ff8ca"},
		{"psubw xmm1, xmm2", func(s *amd64.Section) { s.PsubwXmmRM128(x, y) }, "660ff9ca"},
		{"psubd xmm1, xmm2", func(s *amd64.Section) { s.PsubdXmmRM128(x, y) }, "660ffaca"},
		{"psubq xmm1, xmm2", func(s *amd64.Section) { s.PsubqXmmRM128(x, y) }, "660ffbca"},
		{"paddsb xmm1, xmm2", func(s *amd64.Section) { s.PaddsbXmmRM128(x, y) }, "660fecca"},
		{"paddsw xmm1, xmm2", func(s *amd64.Section) { s.PaddswXmmRM128(x, y) }, "660fedca"},
		{"paddusb xmm1, xmm2", func(s *amd64.Section) { s.PaddusbXmmRM128(x, y) }, "660fdcca"},
		{"paddusw xmm1, xmm2", func(s *amd64.Section) { s.PadduswXmmRM128(x, y) }, "660fddca"},
		{"psubsb xmm1, xmm2", func(s *amd64.Section) { s.PsubsbXmmRM128(x, y) }, "660fe8ca"},
		{"psubsw xmm1, xmm2", func(s *amd64.Section) { s.PsubswXmmRM128(x, y) }, "660fe9ca"},
		{"psubusb xmm1, xmm2", func(s *amd64.Section) { s.PsubusbXmmRM128(x, y) }, "660fd8ca"},
		{"psubusw xmm1, xmm2", func(s *amd64.Section) { s.PsubuswXmmRM128(x, y) }, "660fd9ca"},
		{"pmullw xmm1, xmm2", func(s *amd64.Section) { s.PmullwXmmRM128(x, y) }, "660fd5ca"},
		{"pmulhw xmm1, xmm2", func(s *amd64.Section) { s.PmulhwXmmRM128(x, y) }, "660fe5ca"},
		{"pmulhuw xmm1, xmm2", func(s *amd64.Section) { s.PmulhuwXmmRM128(x, y) }, "660fe4ca"},
		{"pmuludq xmm1, xmm2", func(s *amd64.Section) { s.PmuludqXmmRM128(x, y) }, "660ff4ca"},
		{"pavgb xmm1, xmm2", func(s *amd64.Section) { s.PavgbXmmRM128(x, y) }, "660fe0ca"},
		{"pavgw xmm1, xmm2", func(s *amd64.Section) { s.PavgwXmmRM128(x, y) }, "660fe3ca"},
		{"pminub xmm1, xmm2", func(s *amd64.Section) { s.PminubXmmRM128(x, y) }, "660fdaca"},
		{"pmaxub xmm1, xmm2", func(s *amd64.Section) { s.PmaxubXmmRM128(x, y) }, "660fdeca"},
		{"pminsw xmm1, xmm2", func(s *amd64.Section) { s.PminswXmmRM128(x, y) }, "660feaca"},
		{"pmaxsw xmm1, xmm2", func(s *amd64.Section) { s.PmaxswXmmRM128(x, y) }, "660feeca"},
		{"psadbw xmm1, xmm2", func(s *amd64.Section) { s.PsadbwXmmRM128(x, y) }, "660ff6ca"},
		{"pmaddwd xmm1, xmm2", func(s *amd64.Section) { s.PmaddwdXmmRM128(x, y) }, "660ff5ca"},
		{"pand xmm1, xmm2", func(s *amd64.Section) { s.PandXmmRM128(x, y) }, "660fdbca"},
		{"pandn xmm1, xmm2", func(s *amd64.Section) { s.PandnXmmRM128(x, y) }, "660fdfca"},
		{"por xmm1, xmm2", func(s *amd64.Section) { s.PorXmmRM128(x, y) }, "660febca"},
		{"pxor xmm1, xmm2", func(s *amd64.Section) { s.PxorXmmRM128(x, y) }, "660fefca"},
		{"pcmpeqb xmm1, xmm2", func(s *amd64.Section) { s.PcmpeqbXmmRM128(x, y) }, "660f74ca"},
		{"pcmpeqw xmm1, xmm2", func(s *amd64.Section) { s.PcmpeqwXmmRM128(x, y) }, "660f75ca"},
		{"pcmpeqd xmm1, xmm2", func(s *amd64.Section) { s.PcmpeqdXmmRM128(x, y) }, "660f76ca"},
		{"pcmpgtb xmm1, xmm2", func(s *amd64.Section) { s.PcmpgtbXmmRM128(x, y) }, "660f64ca"},
		{"pcmpgtw xmm1, xmm2", func(s *amd64.Section) { s.PcmpgtwXmmRM128(x, y) }, "660f65ca"},
		{"pcmpgtd xmm1, xmm2", func(s *amd64.Section) { s.PcmpgtdXmmRM128(x, y) }, "660f66ca"},
		{"punpcklbw xmm1, xmm2", func(s *amd64.Section) { s.PunpcklbwXmmRM128(x, y) }, "660f60ca"},
		{"punpcklwd xmm1, xmm2", func(s *amd64.Section) { s.PunpcklwdXmmRM128(x, y) }, "660f61ca"},
		{"punpckldq xmm1, xmm2", func(s *amd64.Section) { s.PunpckldqXmmRM128(x, y) }, "660f62ca"},
		{"punpcklqdq xmm1, xmm2", func(s *amd64.Section) { s.PunpcklqdqXmmRM128(x, y) }, "660f6cca"},
		{"punpckhbw xmm1, xmm2", func(s *amd64.Section) { s.PunpckhbwXmmRM128(x, y) }, "660f68ca"},
		{"punpckhwd xmm1, xmm2", func(s *amd64.Section) { s.PunpckhwdXmmRM128(x, y) }, "660f69ca"},
		{"punpckhdq xmm1, xmm2", func(s *amd64.Section) { s.PunpckhdqXmmRM128(x, y) }, "660f6aca"},
		{"punpckhqdq xmm1, xmm2", func(s *amd64.Section) { s.PunpckhqdqXmmRM128(x, y) }, "660f6dca"},
		{"packsswb xmm1, xmm2", func(s *amd64.Section) { s.PacksswbXmmRM128(x, y) }, "660f63ca"},
		{"packssdw xmm1, xmm2", func(s *amd64.Section) { s.PackssdwXmmRM128(x, y) }, "660f6bca"},
		{"packuswb xmm1, xmm2", func(s *amd64.Section) { s.PackuswbXmmRM128(x, y) }, "660f67ca"},
		{"psllw xmm1, xmm2", func(s *amd64.Section) { s.PsllwXmmRM128(x, y) }, "660ff1ca"},
		{"psllw xmm1, 3", func(s *amd64.Section) { s.PsllwXmmImm8(x, 3) }, "660f71f103"},
		{"pslld xmm1, xmm2", func(s *amd64.Section) { s.PslldXmmRM128(x, y) }, "660ff2ca"},
		{"pslld xmm1, 3", func(s *amd64.Section) { s.PslldXmmImm8(x, 3) }, "660f72f103"},
		{"psllq xmm1, xmm2", func(s *amd64.Section) { s.PsllqXmmRM128(x, y) }, "660ff3ca"},
		{"psllq xmm1, 3", func(s *amd64.Section) { s.PsllqXmmImm8(x, 3) }, "660f73f103"},
		{"psrlw xmm1, xmm2", func(s *amd64.Section) { s.PsrlwXmmRM128(x, y) }, "660fd1ca"},
		{"psrlw xmm1, 3", func(s *amd64.Section) { s.PsrlwXmmImm8(x, 3) }, "660f71d103"},
		{"psrld xmm1, xmm2", func(s *amd64.Section) { s.PsrldXmmRM128(x, y) }, "660fd2ca"},
		{"psrld xmm1, 3", func(s *amd64.Section) { s.PsrldXmmImm8(x, 3) }, "660f72d103"},
		{"psrlq xmm1, xmm2", func(s *amd64.Section) { s.PsrlqXmmRM128(x, y) }, "660fd3ca"},
		{"psrlq xmm1, 3", func(s *amd64.Section) { s.PsrlqXmmImm8(x, 3) }, "660f73d103"},
		{"psraw xmm1, xmm2", func(s *amd64.Section) { s.PsrawXmmRM128(x, y) }, "660fe1ca"},
		{"psraw xmm1, 3", func(s *amd64.Section) { s.PsrawXmmImm8(x, 3) }, "660f71e103"},
		{"psrad xmm1, xmm2", func(s *amd64.Section) { s.PsradXmmRM128(x, y) }, "660fe2ca"},
		{"psrad xmm1, 3", func(s *amd64.Section) { s.PsradXmmImm8(x, 3) }, "660f72e103"},
		{"pslldq xmm1, 3", func(s *amd64.Section) { s.PslldqXmmImm8(x, 3) }, "660f73f903"},
		{"psrldq xmm1, 3", func(s *amd64.Section) { s.PsrldqXmmImm8(x, 3) }, "660f73d903"},
		{"pshufd xmm1, xmm2, 1bh", func(s *amd64.Section) { s.PshufdXmmRM128Imm8(x, y, 0x1b) }, "660f70ca1b"},
		{"pshuflw xmm1, xmm2, 1bh", func(s *amd64.Section) { s.PshuflwXmmRM128Imm8(x, y, 0x1b) }, "f20f70ca1b"},
		{"pshufhw xmm1, xmm2, 1bh", func(s *amd64.Section) { s.PshufhwXmmRM128Imm8(x, y, 0x1b) }, "f30f70ca1b"},
		{"pmovmskb eax, xmm1", func(s *amd64.Section) { s.PmovmskbR32Xmm(reg.EAX, x) }, "660fd7c1"},
		{"pinsrw xmm1, eax, 3", func(s *amd64.Section) { s.PinsrwXmmRM32Imm8(x, reg.EAX, 3) }, "660fc4c803"},
		{"pextrw eax, xmm1, 5", func(s *amd64.Section) { s.PextrwR32XmmImm8(reg.EAX, x, 5) }, "660fc5c105"},
	} {
		t.Run(c.masm, func(t *testing.T) {
			m := amd64.NewModule()
			c.emit(m.Section(amd64.Text))
			o, err := m.Finalize()
			if err != nil {
				t.Fatalf("Finalize: %v", err)
			}
			var got []byte
			for _, sec := range o.Sections() {
				if sec.Name() == ".text" {
					got = sec.Bytes()
				}
			}
			if hex.EncodeToString(got) != c.want {
				t.Errorf("%s\n got %s\nwant %s", c.masm, hex.EncodeToString(got), c.want)
			}
		})
	}
}
