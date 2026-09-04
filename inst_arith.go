package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// Groups 3, 4 and 5: the unary operations, TEST, INC and DEC, and the
// multi-operand IMUL forms that share a mnemonic with the unary one.
//
// The implicit accumulator is what shapes this file. MUL, IMUL, DIV and IDIV
// in their one-operand forms all read and write RAX and RDX — or their
// narrower halves — and none of that is in the signature, because none of it
// is in the encoding either. A lowering calling DivRM64 has to have put the
// dividend in RDX:RAX itself, and has to have executed Cqo to get RDX right;
// forgetting the second is how a positive-looking division traps with #DE.

// ---- NOT and NEG ----------------------------------------------------------

var (
	notRM8  = form("NotRM8")
	notRM16 = form("NotRM16")
	notRM32 = form("NotRM32")
	notRM64 = form("NotRM64")

	negRM8  = form("NegRM8")
	negRM16 = form("NegRM16")
	negRM32 = form("NegRM32")
	negRM64 = form("NegRM64")
)

func (s *Section) NotRM8(dst operand.RM8)   { s.inst(notRM8, dst) }
func (s *Section) NotRM16(dst operand.RM16) { s.inst(notRM16, dst) }
func (s *Section) NotRM32(dst operand.RM32) { s.inst(notRM32, dst) }
func (s *Section) NotRM64(dst operand.RM64) { s.inst(notRM64, dst) }

func (s *Section) NegRM8(dst operand.RM8)   { s.inst(negRM8, dst) }
func (s *Section) NegRM16(dst operand.RM16) { s.inst(negRM16, dst) }
func (s *Section) NegRM32(dst operand.RM32) { s.inst(negRM32, dst) }
func (s *Section) NegRM64(dst operand.RM64) { s.inst(negRM64, dst) }

// ---- MUL, IMUL, DIV, IDIV — the one-operand forms --------------------------
//
// MulRM64(x) computes RDX:RAX = RAX * x. DivRM64(x) computes RAX = RDX:RAX /
// x with the remainder in RDX. The width of the operand decides which
// register pair is involved and nothing else does.
//
// IDIV is signed and rounds toward zero, which is what makes it different
// from SarRM64Imm8 rather than a slower spelling of it.

var (
	mulRM8  = form("MulRM8")
	mulRM16 = form("MulRM16")
	mulRM32 = form("MulRM32")
	mulRM64 = form("MulRM64")

	imulRM8  = form("ImulRM8")
	imulRM16 = form("ImulRM16")
	imulRM32 = form("ImulRM32")
	imulRM64 = form("ImulRM64")

	divRM8  = form("DivRM8")
	divRM16 = form("DivRM16")
	divRM32 = form("DivRM32")
	divRM64 = form("DivRM64")

	idivRM8  = form("IdivRM8")
	idivRM16 = form("IdivRM16")
	idivRM32 = form("IdivRM32")
	idivRM64 = form("IdivRM64")
)

func (s *Section) MulRM8(src operand.RM8)   { s.inst(mulRM8, src) }
func (s *Section) MulRM16(src operand.RM16) { s.inst(mulRM16, src) }
func (s *Section) MulRM32(src operand.RM32) { s.inst(mulRM32, src) }
func (s *Section) MulRM64(src operand.RM64) { s.inst(mulRM64, src) }

func (s *Section) ImulRM8(src operand.RM8)   { s.inst(imulRM8, src) }
func (s *Section) ImulRM16(src operand.RM16) { s.inst(imulRM16, src) }
func (s *Section) ImulRM32(src operand.RM32) { s.inst(imulRM32, src) }
func (s *Section) ImulRM64(src operand.RM64) { s.inst(imulRM64, src) }

func (s *Section) DivRM8(src operand.RM8)   { s.inst(divRM8, src) }
func (s *Section) DivRM16(src operand.RM16) { s.inst(divRM16, src) }
func (s *Section) DivRM32(src operand.RM32) { s.inst(divRM32, src) }
func (s *Section) DivRM64(src operand.RM64) { s.inst(divRM64, src) }

func (s *Section) IdivRM8(src operand.RM8)   { s.inst(idivRM8, src) }
func (s *Section) IdivRM16(src operand.RM16) { s.inst(idivRM16, src) }
func (s *Section) IdivRM32(src operand.RM32) { s.inst(idivRM32, src) }
func (s *Section) IdivRM64(src operand.RM64) { s.inst(idivRM64, src) }

// ---- the multi-operand IMUL forms -----------------------------------------
//
// These are not unary at all and share a file with the group only because
// they share a mnemonic. Two operands is 0F AF and writes the destination
// register alone — no RDX, no double-width result — which is what a compiler
// emits for an ordinary integer multiply. Three operands multiplies the r/m
// by an immediate into the destination, and the imm8 row is four bytes to the
// imm32 row's seven, which is why both are declared.

var (
	imulR32RM32 = form("ImulR32RM32")
	imulR64RM64 = form("ImulR64RM64")

	imulR32RM32Imm8  = form("ImulR32RM32Imm8")
	imulR64RM64Imm8  = form("ImulR64RM64Imm8")
	imulR32RM32Imm32 = form("ImulR32RM32Imm32")
	imulR64RM64Imm32 = form("ImulR64RM64Imm32")
)

func (s *Section) ImulR32RM32(dst reg.R32, src operand.RM32) { s.inst(imulR32RM32, dst, src) }
func (s *Section) ImulR64RM64(dst reg.R64, src operand.RM64) { s.inst(imulR64RM64, dst, src) }

func (s *Section) ImulR32RM32Imm8(dst reg.R32, src operand.RM32, v int64) {
	s.inst(imulR32RM32Imm8, dst, src, imm(v))
}

func (s *Section) ImulR64RM64Imm8(dst reg.R64, src operand.RM64, v int64) {
	s.inst(imulR64RM64Imm8, dst, src, imm(v))
}

func (s *Section) ImulR32RM32Imm32(dst reg.R32, src operand.RM32, v int64) {
	s.inst(imulR32RM32Imm32, dst, src, imm(v))
}

func (s *Section) ImulR64RM64Imm32(dst reg.R64, src operand.RM64, v int64) {
	s.inst(imulR64RM64Imm32, dst, src, imm(v))
}

// ---- TEST -----------------------------------------------------------------
//
// TEST is in this file rather than with the ALU block because its immediate
// forms are group-3 opcodes — F6 /0 and F7 /0 — and not Group 1 ones. It is
// an AND that writes only flags, and TestRM64R64(RAX, RAX) is the two-byte
// way to ask whether a register is zero.
//
// Two rows the table does not declare and a caller may expect: there is no
// TestAXImm16 and no TestRM16Imm16. Both exist in the architecture. Their
// absence is a gap in table_base.go rather than a decision, and adding them
// is two lines there and two helpers here.

var (
	testRM8R8   = form("TestRM8R8")
	testRM16R16 = form("TestRM16R16")
	testRM32R32 = form("TestRM32R32")
	testRM64R64 = form("TestRM64R64")

	testALImm8   = form("TestALImm8")
	testEAXImm32 = form("TestEAXImm32")
	testRAXImm32 = form("TestRAXImm32")

	testRM8Imm8   = form("TestRM8Imm8")
	testRM32Imm32 = form("TestRM32Imm32")
	testRM64Imm32 = form("TestRM64Imm32")
)

func (s *Section) TestRM8R8(a operand.RM8, b reg.R8)     { s.inst(testRM8R8, a, b) }
func (s *Section) TestRM16R16(a operand.RM16, b reg.R16) { s.inst(testRM16R16, a, b) }
func (s *Section) TestRM32R32(a operand.RM32, b reg.R32) { s.inst(testRM32R32, a, b) }
func (s *Section) TestRM64R64(a operand.RM64, b reg.R64) { s.inst(testRM64R64, a, b) }

func (s *Section) TestALImm8(v int64)   { s.inst(testALImm8, fixAL, imm(v)) }
func (s *Section) TestEAXImm32(v int64) { s.inst(testEAXImm32, fixEAX, imm(v)) }
func (s *Section) TestRAXImm32(v int64) { s.inst(testRAXImm32, fixRAX, imm(v)) }

func (s *Section) TestRM8Imm8(dst operand.RM8, v int64)    { s.inst(testRM8Imm8, dst, imm(v)) }
func (s *Section) TestRM32Imm32(dst operand.RM32, v int64) { s.inst(testRM32Imm32, dst, imm(v)) }
func (s *Section) TestRM64Imm32(dst operand.RM64, v int64) { s.inst(testRM64Imm32, dst, imm(v)) }

// ---- INC and DEC ----------------------------------------------------------
//
// Groups 4 and 5, and the only encodings these have. The single-byte 40+r
// forms are gone in long mode — that opcode range is REX now — so there is no
// shorter encoding for Emit to find and no reason for a lowering to prefer
// INC over ADD on size grounds. It still has the partial-flag behaviour that
// makes it different: INC leaves CF alone, ADD does not.

var (
	incRM8  = form("IncRM8")
	incRM16 = form("IncRM16")
	incRM32 = form("IncRM32")
	incRM64 = form("IncRM64")

	decRM8  = form("DecRM8")
	decRM16 = form("DecRM16")
	decRM32 = form("DecRM32")
	decRM64 = form("DecRM64")
)

func (s *Section) IncRM8(dst operand.RM8)   { s.inst(incRM8, dst) }
func (s *Section) IncRM16(dst operand.RM16) { s.inst(incRM16, dst) }
func (s *Section) IncRM32(dst operand.RM32) { s.inst(incRM32, dst) }
func (s *Section) IncRM64(dst operand.RM64) { s.inst(incRM64, dst) }

func (s *Section) DecRM8(dst operand.RM8)   { s.inst(decRM8, dst) }
func (s *Section) DecRM16(dst operand.RM16) { s.inst(decRM16, dst) }
func (s *Section) DecRM32(dst operand.RM32) { s.inst(decRM32, dst) }
func (s *Section) DecRM64(dst operand.RM64) { s.inst(decRM64, dst) }
