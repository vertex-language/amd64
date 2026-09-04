package amd64

import "github.com/vertex-language/amd64/operand"

// Group 2: the shifts and rotates.
//
// Three count forms per width, and the split is in the encoding rather than
// in a convenience. D0/D1 is the by-one form and has no immediate field at
// all — the literal 1 is the opcode — so the count is in the helper's name
// and not in its parameters. D2/D3 takes the count in CL, which the form
// names, so again no parameter. C0/C1 is the one that carries a byte.
//
// The count is masked by the silicon to 5 bits, or 6 under REX.W. This
// package does not mask it for you and does not refuse a count of 40 on a
// 32-bit destination: the field is 8 bits wide and 40 fits it, so there is no
// range failure to report, and what the processor does with it is documented
// behaviour rather than a mistake this layer can identify.
//
// SAL is SHL's second documented spelling and emits SHL's bytes — same /4
// digit, same everything. Both exist as forms with an AliasOf so a listing
// says which name the caller wrote while the bytes say what the silicon does.

// ---- ROL ------------------------------------------------------------------

var (
	rolRM8One  = form("RolRM8One")
	rolRM16One = form("RolRM16One")
	rolRM32One = form("RolRM32One")
	rolRM64One = form("RolRM64One")

	rolRM8CL  = form("RolRM8CL")
	rolRM16CL = form("RolRM16CL")
	rolRM32CL = form("RolRM32CL")
	rolRM64CL = form("RolRM64CL")

	rolRM8Imm8  = form("RolRM8Imm8")
	rolRM16Imm8 = form("RolRM16Imm8")
	rolRM32Imm8 = form("RolRM32Imm8")
	rolRM64Imm8 = form("RolRM64Imm8")
)

func (s *Section) RolRM8One(dst operand.RM8)   { s.inst(rolRM8One, dst, fixOne) }
func (s *Section) RolRM16One(dst operand.RM16) { s.inst(rolRM16One, dst, fixOne) }
func (s *Section) RolRM32One(dst operand.RM32) { s.inst(rolRM32One, dst, fixOne) }
func (s *Section) RolRM64One(dst operand.RM64) { s.inst(rolRM64One, dst, fixOne) }

func (s *Section) RolRM8CL(dst operand.RM8)   { s.inst(rolRM8CL, dst, fixCL) }
func (s *Section) RolRM16CL(dst operand.RM16) { s.inst(rolRM16CL, dst, fixCL) }
func (s *Section) RolRM32CL(dst operand.RM32) { s.inst(rolRM32CL, dst, fixCL) }
func (s *Section) RolRM64CL(dst operand.RM64) { s.inst(rolRM64CL, dst, fixCL) }

func (s *Section) RolRM8Imm8(dst operand.RM8, n int64)   { s.inst(rolRM8Imm8, dst, imm(n)) }
func (s *Section) RolRM16Imm8(dst operand.RM16, n int64) { s.inst(rolRM16Imm8, dst, imm(n)) }
func (s *Section) RolRM32Imm8(dst operand.RM32, n int64) { s.inst(rolRM32Imm8, dst, imm(n)) }
func (s *Section) RolRM64Imm8(dst operand.RM64, n int64) { s.inst(rolRM64Imm8, dst, imm(n)) }

// ---- ROR ------------------------------------------------------------------

var (
	rorRM8One  = form("RorRM8One")
	rorRM16One = form("RorRM16One")
	rorRM32One = form("RorRM32One")
	rorRM64One = form("RorRM64One")

	rorRM8CL  = form("RorRM8CL")
	rorRM16CL = form("RorRM16CL")
	rorRM32CL = form("RorRM32CL")
	rorRM64CL = form("RorRM64CL")

	rorRM8Imm8  = form("RorRM8Imm8")
	rorRM16Imm8 = form("RorRM16Imm8")
	rorRM32Imm8 = form("RorRM32Imm8")
	rorRM64Imm8 = form("RorRM64Imm8")
)

func (s *Section) RorRM8One(dst operand.RM8)   { s.inst(rorRM8One, dst, fixOne) }
func (s *Section) RorRM16One(dst operand.RM16) { s.inst(rorRM16One, dst, fixOne) }
func (s *Section) RorRM32One(dst operand.RM32) { s.inst(rorRM32One, dst, fixOne) }
func (s *Section) RorRM64One(dst operand.RM64) { s.inst(rorRM64One, dst, fixOne) }

func (s *Section) RorRM8CL(dst operand.RM8)   { s.inst(rorRM8CL, dst, fixCL) }
func (s *Section) RorRM16CL(dst operand.RM16) { s.inst(rorRM16CL, dst, fixCL) }
func (s *Section) RorRM32CL(dst operand.RM32) { s.inst(rorRM32CL, dst, fixCL) }
func (s *Section) RorRM64CL(dst operand.RM64) { s.inst(rorRM64CL, dst, fixCL) }

func (s *Section) RorRM8Imm8(dst operand.RM8, n int64)   { s.inst(rorRM8Imm8, dst, imm(n)) }
func (s *Section) RorRM16Imm8(dst operand.RM16, n int64) { s.inst(rorRM16Imm8, dst, imm(n)) }
func (s *Section) RorRM32Imm8(dst operand.RM32, n int64) { s.inst(rorRM32Imm8, dst, imm(n)) }
func (s *Section) RorRM64Imm8(dst operand.RM64, n int64) { s.inst(rorRM64Imm8, dst, imm(n)) }

// ---- RCL ------------------------------------------------------------------
//
// The rotates through carry. They are the two instructions in this group that
// are microcoded and slow on every current processor, and they are here
// because multi-precision arithmetic has no other way to spell what they do.

var (
	rclRM8One  = form("RclRM8One")
	rclRM16One = form("RclRM16One")
	rclRM32One = form("RclRM32One")
	rclRM64One = form("RclRM64One")

	rclRM8CL  = form("RclRM8CL")
	rclRM16CL = form("RclRM16CL")
	rclRM32CL = form("RclRM32CL")
	rclRM64CL = form("RclRM64CL")

	rclRM8Imm8  = form("RclRM8Imm8")
	rclRM16Imm8 = form("RclRM16Imm8")
	rclRM32Imm8 = form("RclRM32Imm8")
	rclRM64Imm8 = form("RclRM64Imm8")
)

func (s *Section) RclRM8One(dst operand.RM8)   { s.inst(rclRM8One, dst, fixOne) }
func (s *Section) RclRM16One(dst operand.RM16) { s.inst(rclRM16One, dst, fixOne) }
func (s *Section) RclRM32One(dst operand.RM32) { s.inst(rclRM32One, dst, fixOne) }
func (s *Section) RclRM64One(dst operand.RM64) { s.inst(rclRM64One, dst, fixOne) }

func (s *Section) RclRM8CL(dst operand.RM8)   { s.inst(rclRM8CL, dst, fixCL) }
func (s *Section) RclRM16CL(dst operand.RM16) { s.inst(rclRM16CL, dst, fixCL) }
func (s *Section) RclRM32CL(dst operand.RM32) { s.inst(rclRM32CL, dst, fixCL) }
func (s *Section) RclRM64CL(dst operand.RM64) { s.inst(rclRM64CL, dst, fixCL) }

func (s *Section) RclRM8Imm8(dst operand.RM8, n int64)   { s.inst(rclRM8Imm8, dst, imm(n)) }
func (s *Section) RclRM16Imm8(dst operand.RM16, n int64) { s.inst(rclRM16Imm8, dst, imm(n)) }
func (s *Section) RclRM32Imm8(dst operand.RM32, n int64) { s.inst(rclRM32Imm8, dst, imm(n)) }
func (s *Section) RclRM64Imm8(dst operand.RM64, n int64) { s.inst(rclRM64Imm8, dst, imm(n)) }

// ---- RCR ------------------------------------------------------------------

var (
	rcrRM8One  = form("RcrRM8One")
	rcrRM16One = form("RcrRM16One")
	rcrRM32One = form("RcrRM32One")
	rcrRM64One = form("RcrRM64One")

	rcrRM8CL  = form("RcrRM8CL")
	rcrRM16CL = form("RcrRM16CL")
	rcrRM32CL = form("RcrRM32CL")
	rcrRM64CL = form("RcrRM64CL")

	rcrRM8Imm8  = form("RcrRM8Imm8")
	rcrRM16Imm8 = form("RcrRM16Imm8")
	rcrRM32Imm8 = form("RcrRM32Imm8")
	rcrRM64Imm8 = form("RcrRM64Imm8")
)

func (s *Section) RcrRM8One(dst operand.RM8)   { s.inst(rcrRM8One, dst, fixOne) }
func (s *Section) RcrRM16One(dst operand.RM16) { s.inst(rcrRM16One, dst, fixOne) }
func (s *Section) RcrRM32One(dst operand.RM32) { s.inst(rcrRM32One, dst, fixOne) }
func (s *Section) RcrRM64One(dst operand.RM64) { s.inst(rcrRM64One, dst, fixOne) }

func (s *Section) RcrRM8CL(dst operand.RM8)   { s.inst(rcrRM8CL, dst, fixCL) }
func (s *Section) RcrRM16CL(dst operand.RM16) { s.inst(rcrRM16CL, dst, fixCL) }
func (s *Section) RcrRM32CL(dst operand.RM32) { s.inst(rcrRM32CL, dst, fixCL) }
func (s *Section) RcrRM64CL(dst operand.RM64) { s.inst(rcrRM64CL, dst, fixCL) }

func (s *Section) RcrRM8Imm8(dst operand.RM8, n int64)   { s.inst(rcrRM8Imm8, dst, imm(n)) }
func (s *Section) RcrRM16Imm8(dst operand.RM16, n int64) { s.inst(rcrRM16Imm8, dst, imm(n)) }
func (s *Section) RcrRM32Imm8(dst operand.RM32, n int64) { s.inst(rcrRM32Imm8, dst, imm(n)) }
func (s *Section) RcrRM64Imm8(dst operand.RM64, n int64) { s.inst(rcrRM64Imm8, dst, imm(n)) }

// ---- SHL ------------------------------------------------------------------

var (
	shlRM8One  = form("ShlRM8One")
	shlRM16One = form("ShlRM16One")
	shlRM32One = form("ShlRM32One")
	shlRM64One = form("ShlRM64One")

	shlRM8CL  = form("ShlRM8CL")
	shlRM16CL = form("ShlRM16CL")
	shlRM32CL = form("ShlRM32CL")
	shlRM64CL = form("ShlRM64CL")

	shlRM8Imm8  = form("ShlRM8Imm8")
	shlRM16Imm8 = form("ShlRM16Imm8")
	shlRM32Imm8 = form("ShlRM32Imm8")
	shlRM64Imm8 = form("ShlRM64Imm8")
)

func (s *Section) ShlRM8One(dst operand.RM8)   { s.inst(shlRM8One, dst, fixOne) }
func (s *Section) ShlRM16One(dst operand.RM16) { s.inst(shlRM16One, dst, fixOne) }
func (s *Section) ShlRM32One(dst operand.RM32) { s.inst(shlRM32One, dst, fixOne) }
func (s *Section) ShlRM64One(dst operand.RM64) { s.inst(shlRM64One, dst, fixOne) }

func (s *Section) ShlRM8CL(dst operand.RM8)   { s.inst(shlRM8CL, dst, fixCL) }
func (s *Section) ShlRM16CL(dst operand.RM16) { s.inst(shlRM16CL, dst, fixCL) }
func (s *Section) ShlRM32CL(dst operand.RM32) { s.inst(shlRM32CL, dst, fixCL) }
func (s *Section) ShlRM64CL(dst operand.RM64) { s.inst(shlRM64CL, dst, fixCL) }

func (s *Section) ShlRM8Imm8(dst operand.RM8, n int64)   { s.inst(shlRM8Imm8, dst, imm(n)) }
func (s *Section) ShlRM16Imm8(dst operand.RM16, n int64) { s.inst(shlRM16Imm8, dst, imm(n)) }
func (s *Section) ShlRM32Imm8(dst operand.RM32, n int64) { s.inst(shlRM32Imm8, dst, imm(n)) }
func (s *Section) ShlRM64Imm8(dst operand.RM64, n int64) { s.inst(shlRM64Imm8, dst, imm(n)) }

// ---- SHR ------------------------------------------------------------------

var (
	shrRM8One  = form("ShrRM8One")
	shrRM16One = form("ShrRM16One")
	shrRM32One = form("ShrRM32One")
	shrRM64One = form("ShrRM64One")

	shrRM8CL  = form("ShrRM8CL")
	shrRM16CL = form("ShrRM16CL")
	shrRM32CL = form("ShrRM32CL")
	shrRM64CL = form("ShrRM64CL")

	shrRM8Imm8  = form("ShrRM8Imm8")
	shrRM16Imm8 = form("ShrRM16Imm8")
	shrRM32Imm8 = form("ShrRM32Imm8")
	shrRM64Imm8 = form("ShrRM64Imm8")
)

func (s *Section) ShrRM8One(dst operand.RM8)   { s.inst(shrRM8One, dst, fixOne) }
func (s *Section) ShrRM16One(dst operand.RM16) { s.inst(shrRM16One, dst, fixOne) }
func (s *Section) ShrRM32One(dst operand.RM32) { s.inst(shrRM32One, dst, fixOne) }
func (s *Section) ShrRM64One(dst operand.RM64) { s.inst(shrRM64One, dst, fixOne) }

func (s *Section) ShrRM8CL(dst operand.RM8)   { s.inst(shrRM8CL, dst, fixCL) }
func (s *Section) ShrRM16CL(dst operand.RM16) { s.inst(shrRM16CL, dst, fixCL) }
func (s *Section) ShrRM32CL(dst operand.RM32) { s.inst(shrRM32CL, dst, fixCL) }
func (s *Section) ShrRM64CL(dst operand.RM64) { s.inst(shrRM64CL, dst, fixCL) }

func (s *Section) ShrRM8Imm8(dst operand.RM8, n int64)   { s.inst(shrRM8Imm8, dst, imm(n)) }
func (s *Section) ShrRM16Imm8(dst operand.RM16, n int64) { s.inst(shrRM16Imm8, dst, imm(n)) }
func (s *Section) ShrRM32Imm8(dst operand.RM32, n int64) { s.inst(shrRM32Imm8, dst, imm(n)) }
func (s *Section) ShrRM64Imm8(dst operand.RM64, n int64) { s.inst(shrRM64Imm8, dst, imm(n)) }

// ---- SAL ------------------------------------------------------------------
//
// SHL's other name. Same /4 digit, same bytes, its own rows so a listing can
// report the spelling the caller used.

var (
	salRM8One  = form("SalRM8One")
	salRM16One = form("SalRM16One")
	salRM32One = form("SalRM32One")
	salRM64One = form("SalRM64One")

	salRM8CL  = form("SalRM8CL")
	salRM16CL = form("SalRM16CL")
	salRM32CL = form("SalRM32CL")
	salRM64CL = form("SalRM64CL")

	salRM8Imm8  = form("SalRM8Imm8")
	salRM16Imm8 = form("SalRM16Imm8")
	salRM32Imm8 = form("SalRM32Imm8")
	salRM64Imm8 = form("SalRM64Imm8")
)

func (s *Section) SalRM8One(dst operand.RM8)   { s.inst(salRM8One, dst, fixOne) }
func (s *Section) SalRM16One(dst operand.RM16) { s.inst(salRM16One, dst, fixOne) }
func (s *Section) SalRM32One(dst operand.RM32) { s.inst(salRM32One, dst, fixOne) }
func (s *Section) SalRM64One(dst operand.RM64) { s.inst(salRM64One, dst, fixOne) }

func (s *Section) SalRM8CL(dst operand.RM8)   { s.inst(salRM8CL, dst, fixCL) }
func (s *Section) SalRM16CL(dst operand.RM16) { s.inst(salRM16CL, dst, fixCL) }
func (s *Section) SalRM32CL(dst operand.RM32) { s.inst(salRM32CL, dst, fixCL) }
func (s *Section) SalRM64CL(dst operand.RM64) { s.inst(salRM64CL, dst, fixCL) }

func (s *Section) SalRM8Imm8(dst operand.RM8, n int64)   { s.inst(salRM8Imm8, dst, imm(n)) }
func (s *Section) SalRM16Imm8(dst operand.RM16, n int64) { s.inst(salRM16Imm8, dst, imm(n)) }
func (s *Section) SalRM32Imm8(dst operand.RM32, n int64) { s.inst(salRM32Imm8, dst, imm(n)) }
func (s *Section) SalRM64Imm8(dst operand.RM64, n int64) { s.inst(salRM64Imm8, dst, imm(n)) }

// ---- SAR ------------------------------------------------------------------
//
// The arithmetic right shift, /7, and the one of the four that is not a
// division: SAR rounds toward negative infinity, so a signed divide by a
// power of two needs a correction term first. IdivRM64 is the instruction
// that rounds toward zero.

var (
	sarRM8One  = form("SarRM8One")
	sarRM16One = form("SarRM16One")
	sarRM32One = form("SarRM32One")
	sarRM64One = form("SarRM64One")

	sarRM8CL  = form("SarRM8CL")
	sarRM16CL = form("SarRM16CL")
	sarRM32CL = form("SarRM32CL")
	sarRM64CL = form("SarRM64CL")

	sarRM8Imm8  = form("SarRM8Imm8")
	sarRM16Imm8 = form("SarRM16Imm8")
	sarRM32Imm8 = form("SarRM32Imm8")
	sarRM64Imm8 = form("SarRM64Imm8")
)

func (s *Section) SarRM8One(dst operand.RM8)   { s.inst(sarRM8One, dst, fixOne) }
func (s *Section) SarRM16One(dst operand.RM16) { s.inst(sarRM16One, dst, fixOne) }
func (s *Section) SarRM32One(dst operand.RM32) { s.inst(sarRM32One, dst, fixOne) }
func (s *Section) SarRM64One(dst operand.RM64) { s.inst(sarRM64One, dst, fixOne) }

func (s *Section) SarRM8CL(dst operand.RM8)   { s.inst(sarRM8CL, dst, fixCL) }
func (s *Section) SarRM16CL(dst operand.RM16) { s.inst(sarRM16CL, dst, fixCL) }
func (s *Section) SarRM32CL(dst operand.RM32) { s.inst(sarRM32CL, dst, fixCL) }
func (s *Section) SarRM64CL(dst operand.RM64) { s.inst(sarRM64CL, dst, fixCL) }

func (s *Section) SarRM8Imm8(dst operand.RM8, n int64)   { s.inst(sarRM8Imm8, dst, imm(n)) }
func (s *Section) SarRM16Imm8(dst operand.RM16, n int64) { s.inst(sarRM16Imm8, dst, imm(n)) }
func (s *Section) SarRM32Imm8(dst operand.RM32, n int64) { s.inst(sarRM32Imm8, dst, imm(n)) }
func (s *Section) SarRM64Imm8(dst operand.RM64, n int64) { s.inst(sarRM64Imm8, dst, imm(n)) }
