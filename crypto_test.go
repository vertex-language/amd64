package amd64_test

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/feature"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// cryptoFeatures is every gate the rows below carry.
var cryptoFeatures = feature.Default().Add(feature.SSSE3, feature.SSE42,
	feature.AES, feature.PCLMULQDQ, feature.SHA)

// SSSE3's permutes, SHUFPS, AES-NI, PCLMULQDQ, SHA and CRC32, each against
// the bytes clang's integrated assembler produces for the AT&T line in the
// first column (clang -c -target x86_64-linux-gnu, read back by objdump -d).
// The extended registers are there for REX: SHA's forms have no mandatory
// prefix, so REX follows nothing, and CRC32's sits between F2 and 0F.
func TestCryptoMatchesClang(t *testing.T) {
	x, y := reg.XMM1, reg.XMM2

	for _, c := range []struct {
		att  string
		emit func(s *amd64.Section)
		want string
	}{
		{"pshufb %xmm2, %xmm1", func(s *amd64.Section) { s.PshufbXmmRM128(x, y) }, "660f3800ca"},
		{"pshufb (%rax), %xmm9", func(s *amd64.Section) { s.PshufbXmmRM128(reg.XMM9, operand.Mem128(reg.RAX)) }, "66440f380008"},
		{"palignr $4, %xmm2, %xmm1", func(s *amd64.Section) { s.PalignrXmmRM128Imm8(x, y, 4) }, "660f3a0fca04"},
		{"shufps $0xbb, %xmm2, %xmm1", func(s *amd64.Section) { s.ShufpsXmmRM128Imm8(x, y, 0xbb) }, "0fc6cabb"},
		{"shufps $0x11, %xmm10, %xmm1", func(s *amd64.Section) { s.ShufpsXmmRM128Imm8(x, reg.XMM10, 0x11) }, "410fc6ca11"},
		{"aesenc %xmm2, %xmm1", func(s *amd64.Section) { s.AesencXmmRM128(x, y) }, "660f38dcca"},
		{"aesenclast %xmm2, %xmm1", func(s *amd64.Section) { s.AesenclastXmmRM128(x, y) }, "660f38ddca"},
		{"aesdec %xmm2, %xmm1", func(s *amd64.Section) { s.AesdecXmmRM128(x, y) }, "660f38deca"},
		{"aesdeclast %xmm2, %xmm1", func(s *amd64.Section) { s.AesdeclastXmmRM128(x, y) }, "660f38dfca"},
		{"aesimc %xmm2, %xmm1", func(s *amd64.Section) { s.AesimcXmmRM128(x, y) }, "660f38dbca"},
		{"aeskeygenassist $1, %xmm2, %xmm1", func(s *amd64.Section) { s.AeskeygenassistXmmRM128Imm8(x, y, 1) }, "660f3adfca01"},
		{"pclmulqdq $0x11, %xmm2, %xmm1", func(s *amd64.Section) { s.PclmulqdqXmmRM128Imm8(x, y, 0x11) }, "660f3a44ca11"},
		{"pclmulqdq $0x10, %xmm12, %xmm11", func(s *amd64.Section) { s.PclmulqdqXmmRM128Imm8(reg.XMM11, reg.XMM12, 0x10) }, "66450f3a44dc10"},
		{"sha1rnds4 $3, %xmm2, %xmm1", func(s *amd64.Section) { s.Sha1rnds4XmmRM128Imm8(x, y, 3) }, "0f3accca03"},
		{"sha1nexte %xmm2, %xmm1", func(s *amd64.Section) { s.Sha1nexteXmmRM128(x, y) }, "0f38c8ca"},
		{"sha1msg1 %xmm2, %xmm1", func(s *amd64.Section) { s.Sha1msg1XmmRM128(x, y) }, "0f38c9ca"},
		{"sha1msg2 %xmm2, %xmm1", func(s *amd64.Section) { s.Sha1msg2XmmRM128(x, y) }, "0f38caca"},
		{"sha256rnds2 %xmm0, %xmm2, %xmm1", func(s *amd64.Section) { s.Sha256rnds2XmmRM128Xmm0(x, y) }, "0f38cbca"},
		{"sha256rnds2 %xmm0, %xmm10, %xmm9", func(s *amd64.Section) { s.Sha256rnds2XmmRM128Xmm0(reg.XMM9, reg.XMM10) }, "450f38cbca"},
		{"sha256msg1 %xmm2, %xmm1", func(s *amd64.Section) { s.Sha256msg1XmmRM128(x, y) }, "0f38ccca"},
		{"sha256msg2 %xmm2, %xmm1", func(s *amd64.Section) { s.Sha256msg2XmmRM128(x, y) }, "0f38cdca"},
		{"crc32b %cl, %eax", func(s *amd64.Section) { s.Crc32R32RM8(reg.EAX, reg.CL) }, "f20f38f0c1"},
		{"crc32b %sil, %eax", func(s *amd64.Section) { s.Crc32R32RM8(reg.EAX, reg.SIL) }, "f2400f38f0c6"},
		{"crc32w %cx, %eax", func(s *amd64.Section) { s.Crc32R32RM16(reg.EAX, reg.CX) }, "66f20f38f1c1"},
		{"crc32l %ecx, %eax", func(s *amd64.Section) { s.Crc32R32RM32(reg.EAX, reg.ECX) }, "f20f38f1c1"},
		{"crc32l %r9d, %r10d", func(s *amd64.Section) { s.Crc32R32RM32(reg.R10D, reg.R9D) }, "f2450f38f1d1"},
		{"crc32q %rcx, %rax", func(s *amd64.Section) { s.Crc32R64RM64(reg.RAX, reg.RCX) }, "f2480f38f1c1"},
		{"crc32q (%rdi), %rax", func(s *amd64.Section) { s.Crc32R64RM64(reg.RAX, operand.Mem64(reg.RDI)) }, "f2480f38f107"},
	} {
		t.Run(c.att, func(t *testing.T) {
			m := amd64.NewModule(amd64.WithFeatures(cryptoFeatures))
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
				t.Errorf("%s\n got %s\nwant %s", c.att, hex.EncodeToString(got), c.want)
			}
		})
	}
}

// The gates are real: each family is refused on the baseline, and SHUFPS,
// which is SSE, is not.
func TestCryptoGates(t *testing.T) {
	x, y := reg.XMM1, reg.XMM2
	for name, emit := range map[string]func(s *amd64.Section){
		"pshufb":      func(s *amd64.Section) { s.PshufbXmmRM128(x, y) },
		"aesenc":      func(s *amd64.Section) { s.AesencXmmRM128(x, y) },
		"pclmulqdq":   func(s *amd64.Section) { s.PclmulqdqXmmRM128Imm8(x, y, 0) },
		"sha256rnds2": func(s *amd64.Section) { s.Sha256rnds2XmmRM128Xmm0(x, y) },
		"crc32":       func(s *amd64.Section) { s.Crc32R32RM32(reg.EAX, reg.ECX) },
	} {
		m := amd64.NewModule()
		emit(m.Section(amd64.Text))
		if _, err := m.Finalize(); !errors.Is(err, obj.ErrFeature) {
			t.Errorf("baseline %s = %v, want ErrFeature", name, err)
		}
	}
	m := amd64.NewModule()
	m.Section(amd64.Text).ShufpsXmmRM128Imm8(x, y, 0)
	if _, err := m.Finalize(); err != nil {
		t.Errorf("baseline shufps: %v", err)
	}
}
