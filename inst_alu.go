package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// The arithmetic-logical block: ADD, OR, ADC, SBB, AND, SUB, XOR, CMP.
//
// Eight operations sharing one opcode pattern, which is why isa/table_base.go
// declares them in a loop and why this file reads as eight copies of one
// shape. Each has fifteen forms: the store and load directions at four
// widths, the four accumulator forms, and Group 1's sign-extended imm8 and
// full-width immediate rows.
//
// Two things are worth knowing before reading further, and neither repeats
// below. The accumulator forms take no register parameter because the form
// names the register and leaves no field to put another one in — AddRAXImm32
// is RAX or nothing, and that is in the name rather than in a signature that
// could be handed RCX. And the imm8 forms are the reason Emit's shortest-wins
// rule pays: add rax, 1 through the four-byte 83 /0 row rather than the
// seven-byte 81 /0 one. Naming a form directly, as these do, is stating which
// you want; a value that does not fit the field you named is ErrRange with
// the width and range in the notes.

// ---- ADD ------------------------------------------------------------------

var (
	addRM8R8     = form("AddRM8R8")
	addRM16R16   = form("AddRM16R16")
	addRM32R32   = form("AddRM32R32")
	addRM64R64   = form("AddRM64R64")
	addR8RM8     = form("AddR8RM8")
	addR16RM16   = form("AddR16RM16")
	addR32RM32   = form("AddR32RM32")
	addR64RM64   = form("AddR64RM64")
	addALImm8    = form("AddALImm8")
	addAXImm16   = form("AddAXImm16")
	addEAXImm32  = form("AddEAXImm32")
	addRAXImm32  = form("AddRAXImm32")
	addRM16Imm8  = form("AddRM16Imm8")
	addRM32Imm8  = form("AddRM32Imm8")
	addRM64Imm8  = form("AddRM64Imm8")
	addRM8Imm8   = form("AddRM8Imm8")
	addRM16Imm16 = form("AddRM16Imm16")
	addRM32Imm32 = form("AddRM32Imm32")
	addRM64Imm32 = form("AddRM64Imm32")
)

func (s *Section) AddRM8R8(dst operand.RM8, src reg.R8)     { s.inst(addRM8R8, dst, src) }
func (s *Section) AddRM16R16(dst operand.RM16, src reg.R16) { s.inst(addRM16R16, dst, src) }
func (s *Section) AddRM32R32(dst operand.RM32, src reg.R32) { s.inst(addRM32R32, dst, src) }
func (s *Section) AddRM64R64(dst operand.RM64, src reg.R64) { s.inst(addRM64R64, dst, src) }

func (s *Section) AddR8RM8(dst reg.R8, src operand.RM8)     { s.inst(addR8RM8, dst, src) }
func (s *Section) AddR16RM16(dst reg.R16, src operand.RM16) { s.inst(addR16RM16, dst, src) }
func (s *Section) AddR32RM32(dst reg.R32, src operand.RM32) { s.inst(addR32RM32, dst, src) }
func (s *Section) AddR64RM64(dst reg.R64, src operand.RM64) { s.inst(addR64RM64, dst, src) }

func (s *Section) AddALImm8(v int64)   { s.inst(addALImm8, fixAL, imm(v)) }
func (s *Section) AddAXImm16(v int64)  { s.inst(addAXImm16, fixAX, imm(v)) }
func (s *Section) AddEAXImm32(v int64) { s.inst(addEAXImm32, fixEAX, imm(v)) }
func (s *Section) AddRAXImm32(v int64) { s.inst(addRAXImm32, fixRAX, imm(v)) }

func (s *Section) AddRM8Imm8(dst operand.RM8, v int64)   { s.inst(addRM8Imm8, dst, imm(v)) }
func (s *Section) AddRM16Imm8(dst operand.RM16, v int64) { s.inst(addRM16Imm8, dst, imm(v)) }
func (s *Section) AddRM32Imm8(dst operand.RM32, v int64) { s.inst(addRM32Imm8, dst, imm(v)) }
func (s *Section) AddRM64Imm8(dst operand.RM64, v int64) { s.inst(addRM64Imm8, dst, imm(v)) }

func (s *Section) AddRM16Imm16(dst operand.RM16, v int64) { s.inst(addRM16Imm16, dst, imm(v)) }
func (s *Section) AddRM32Imm32(dst operand.RM32, v int64) { s.inst(addRM32Imm32, dst, imm(v)) }
func (s *Section) AddRM64Imm32(dst operand.RM64, v int64) { s.inst(addRM64Imm32, dst, imm(v)) }

// ---- OR -------------------------------------------------------------------

var (
	orRM8R8     = form("OrRM8R8")
	orRM16R16   = form("OrRM16R16")
	orRM32R32   = form("OrRM32R32")
	orRM64R64   = form("OrRM64R64")
	orR8RM8     = form("OrR8RM8")
	orR16RM16   = form("OrR16RM16")
	orR32RM32   = form("OrR32RM32")
	orR64RM64   = form("OrR64RM64")
	orALImm8    = form("OrALImm8")
	orAXImm16   = form("OrAXImm16")
	orEAXImm32  = form("OrEAXImm32")
	orRAXImm32  = form("OrRAXImm32")
	orRM16Imm8  = form("OrRM16Imm8")
	orRM32Imm8  = form("OrRM32Imm8")
	orRM64Imm8  = form("OrRM64Imm8")
	orRM8Imm8   = form("OrRM8Imm8")
	orRM16Imm16 = form("OrRM16Imm16")
	orRM32Imm32 = form("OrRM32Imm32")
	orRM64Imm32 = form("OrRM64Imm32")
)

func (s *Section) OrRM8R8(dst operand.RM8, src reg.R8)     { s.inst(orRM8R8, dst, src) }
func (s *Section) OrRM16R16(dst operand.RM16, src reg.R16) { s.inst(orRM16R16, dst, src) }
func (s *Section) OrRM32R32(dst operand.RM32, src reg.R32) { s.inst(orRM32R32, dst, src) }
func (s *Section) OrRM64R64(dst operand.RM64, src reg.R64) { s.inst(orRM64R64, dst, src) }

func (s *Section) OrR8RM8(dst reg.R8, src operand.RM8)     { s.inst(orR8RM8, dst, src) }
func (s *Section) OrR16RM16(dst reg.R16, src operand.RM16) { s.inst(orR16RM16, dst, src) }
func (s *Section) OrR32RM32(dst reg.R32, src operand.RM32) { s.inst(orR32RM32, dst, src) }
func (s *Section) OrR64RM64(dst reg.R64, src operand.RM64) { s.inst(orR64RM64, dst, src) }

func (s *Section) OrALImm8(v int64)   { s.inst(orALImm8, fixAL, imm(v)) }
func (s *Section) OrAXImm16(v int64)  { s.inst(orAXImm16, fixAX, imm(v)) }
func (s *Section) OrEAXImm32(v int64) { s.inst(orEAXImm32, fixEAX, imm(v)) }
func (s *Section) OrRAXImm32(v int64) { s.inst(orRAXImm32, fixRAX, imm(v)) }

func (s *Section) OrRM8Imm8(dst operand.RM8, v int64)   { s.inst(orRM8Imm8, dst, imm(v)) }
func (s *Section) OrRM16Imm8(dst operand.RM16, v int64) { s.inst(orRM16Imm8, dst, imm(v)) }
func (s *Section) OrRM32Imm8(dst operand.RM32, v int64) { s.inst(orRM32Imm8, dst, imm(v)) }
func (s *Section) OrRM64Imm8(dst operand.RM64, v int64) { s.inst(orRM64Imm8, dst, imm(v)) }

func (s *Section) OrRM16Imm16(dst operand.RM16, v int64) { s.inst(orRM16Imm16, dst, imm(v)) }
func (s *Section) OrRM32Imm32(dst operand.RM32, v int64) { s.inst(orRM32Imm32, dst, imm(v)) }
func (s *Section) OrRM64Imm32(dst operand.RM64, v int64) { s.inst(orRM64Imm32, dst, imm(v)) }

// ---- ADC ------------------------------------------------------------------

var (
	adcRM8R8     = form("AdcRM8R8")
	adcRM16R16   = form("AdcRM16R16")
	adcRM32R32   = form("AdcRM32R32")
	adcRM64R64   = form("AdcRM64R64")
	adcR8RM8     = form("AdcR8RM8")
	adcR16RM16   = form("AdcR16RM16")
	adcR32RM32   = form("AdcR32RM32")
	adcR64RM64   = form("AdcR64RM64")
	adcALImm8    = form("AdcALImm8")
	adcAXImm16   = form("AdcAXImm16")
	adcEAXImm32  = form("AdcEAXImm32")
	adcRAXImm32  = form("AdcRAXImm32")
	adcRM16Imm8  = form("AdcRM16Imm8")
	adcRM32Imm8  = form("AdcRM32Imm8")
	adcRM64Imm8  = form("AdcRM64Imm8")
	adcRM8Imm8   = form("AdcRM8Imm8")
	adcRM16Imm16 = form("AdcRM16Imm16")
	adcRM32Imm32 = form("AdcRM32Imm32")
	adcRM64Imm32 = form("AdcRM64Imm32")
)

func (s *Section) AdcRM8R8(dst operand.RM8, src reg.R8)     { s.inst(adcRM8R8, dst, src) }
func (s *Section) AdcRM16R16(dst operand.RM16, src reg.R16) { s.inst(adcRM16R16, dst, src) }
func (s *Section) AdcRM32R32(dst operand.RM32, src reg.R32) { s.inst(adcRM32R32, dst, src) }
func (s *Section) AdcRM64R64(dst operand.RM64, src reg.R64) { s.inst(adcRM64R64, dst, src) }

func (s *Section) AdcR8RM8(dst reg.R8, src operand.RM8)     { s.inst(adcR8RM8, dst, src) }
func (s *Section) AdcR16RM16(dst reg.R16, src operand.RM16) { s.inst(adcR16RM16, dst, src) }
func (s *Section) AdcR32RM32(dst reg.R32, src operand.RM32) { s.inst(adcR32RM32, dst, src) }
func (s *Section) AdcR64RM64(dst reg.R64, src operand.RM64) { s.inst(adcR64RM64, dst, src) }

func (s *Section) AdcALImm8(v int64)   { s.inst(adcALImm8, fixAL, imm(v)) }
func (s *Section) AdcAXImm16(v int64)  { s.inst(adcAXImm16, fixAX, imm(v)) }
func (s *Section) AdcEAXImm32(v int64) { s.inst(adcEAXImm32, fixEAX, imm(v)) }
func (s *Section) AdcRAXImm32(v int64) { s.inst(adcRAXImm32, fixRAX, imm(v)) }

func (s *Section) AdcRM8Imm8(dst operand.RM8, v int64)   { s.inst(adcRM8Imm8, dst, imm(v)) }
func (s *Section) AdcRM16Imm8(dst operand.RM16, v int64) { s.inst(adcRM16Imm8, dst, imm(v)) }
func (s *Section) AdcRM32Imm8(dst operand.RM32, v int64) { s.inst(adcRM32Imm8, dst, imm(v)) }
func (s *Section) AdcRM64Imm8(dst operand.RM64, v int64) { s.inst(adcRM64Imm8, dst, imm(v)) }

func (s *Section) AdcRM16Imm16(dst operand.RM16, v int64) { s.inst(adcRM16Imm16, dst, imm(v)) }
func (s *Section) AdcRM32Imm32(dst operand.RM32, v int64) { s.inst(adcRM32Imm32, dst, imm(v)) }
func (s *Section) AdcRM64Imm32(dst operand.RM64, v int64) { s.inst(adcRM64Imm32, dst, imm(v)) }

// ---- SBB ------------------------------------------------------------------

var (
	sbbRM8R8     = form("SbbRM8R8")
	sbbRM16R16   = form("SbbRM16R16")
	sbbRM32R32   = form("SbbRM32R32")
	sbbRM64R64   = form("SbbRM64R64")
	sbbR8RM8     = form("SbbR8RM8")
	sbbR16RM16   = form("SbbR16RM16")
	sbbR32RM32   = form("SbbR32RM32")
	sbbR64RM64   = form("SbbR64RM64")
	sbbALImm8    = form("SbbALImm8")
	sbbAXImm16   = form("SbbAXImm16")
	sbbEAXImm32  = form("SbbEAXImm32")
	sbbRAXImm32  = form("SbbRAXImm32")
	sbbRM16Imm8  = form("SbbRM16Imm8")
	sbbRM32Imm8  = form("SbbRM32Imm8")
	sbbRM64Imm8  = form("SbbRM64Imm8")
	sbbRM8Imm8   = form("SbbRM8Imm8")
	sbbRM16Imm16 = form("SbbRM16Imm16")
	sbbRM32Imm32 = form("SbbRM32Imm32")
	sbbRM64Imm32 = form("SbbRM64Imm32")
)

func (s *Section) SbbRM8R8(dst operand.RM8, src reg.R8)     { s.inst(sbbRM8R8, dst, src) }
func (s *Section) SbbRM16R16(dst operand.RM16, src reg.R16) { s.inst(sbbRM16R16, dst, src) }
func (s *Section) SbbRM32R32(dst operand.RM32, src reg.R32) { s.inst(sbbRM32R32, dst, src) }
func (s *Section) SbbRM64R64(dst operand.RM64, src reg.R64) { s.inst(sbbRM64R64, dst, src) }

func (s *Section) SbbR8RM8(dst reg.R8, src operand.RM8)     { s.inst(sbbR8RM8, dst, src) }
func (s *Section) SbbR16RM16(dst reg.R16, src operand.RM16) { s.inst(sbbR16RM16, dst, src) }
func (s *Section) SbbR32RM32(dst reg.R32, src operand.RM32) { s.inst(sbbR32RM32, dst, src) }
func (s *Section) SbbR64RM64(dst reg.R64, src operand.RM64) { s.inst(sbbR64RM64, dst, src) }

func (s *Section) SbbALImm8(v int64)   { s.inst(sbbALImm8, fixAL, imm(v)) }
func (s *Section) SbbAXImm16(v int64)  { s.inst(sbbAXImm16, fixAX, imm(v)) }
func (s *Section) SbbEAXImm32(v int64) { s.inst(sbbEAXImm32, fixEAX, imm(v)) }
func (s *Section) SbbRAXImm32(v int64) { s.inst(sbbRAXImm32, fixRAX, imm(v)) }

func (s *Section) SbbRM8Imm8(dst operand.RM8, v int64)   { s.inst(sbbRM8Imm8, dst, imm(v)) }
func (s *Section) SbbRM16Imm8(dst operand.RM16, v int64) { s.inst(sbbRM16Imm8, dst, imm(v)) }
func (s *Section) SbbRM32Imm8(dst operand.RM32, v int64) { s.inst(sbbRM32Imm8, dst, imm(v)) }
func (s *Section) SbbRM64Imm8(dst operand.RM64, v int64) { s.inst(sbbRM64Imm8, dst, imm(v)) }

func (s *Section) SbbRM16Imm16(dst operand.RM16, v int64) { s.inst(sbbRM16Imm16, dst, imm(v)) }
func (s *Section) SbbRM32Imm32(dst operand.RM32, v int64) { s.inst(sbbRM32Imm32, dst, imm(v)) }
func (s *Section) SbbRM64Imm32(dst operand.RM64, v int64) { s.inst(sbbRM64Imm32, dst, imm(v)) }

// ---- AND ------------------------------------------------------------------

var (
	andRM8R8     = form("AndRM8R8")
	andRM16R16   = form("AndRM16R16")
	andRM32R32   = form("AndRM32R32")
	andRM64R64   = form("AndRM64R64")
	andR8RM8     = form("AndR8RM8")
	andR16RM16   = form("AndR16RM16")
	andR32RM32   = form("AndR32RM32")
	andR64RM64   = form("AndR64RM64")
	andALImm8    = form("AndALImm8")
	andAXImm16   = form("AndAXImm16")
	andEAXImm32  = form("AndEAXImm32")
	andRAXImm32  = form("AndRAXImm32")
	andRM16Imm8  = form("AndRM16Imm8")
	andRM32Imm8  = form("AndRM32Imm8")
	andRM64Imm8  = form("AndRM64Imm8")
	andRM8Imm8   = form("AndRM8Imm8")
	andRM16Imm16 = form("AndRM16Imm16")
	andRM32Imm32 = form("AndRM32Imm32")
	andRM64Imm32 = form("AndRM64Imm32")
)

func (s *Section) AndRM8R8(dst operand.RM8, src reg.R8)     { s.inst(andRM8R8, dst, src) }
func (s *Section) AndRM16R16(dst operand.RM16, src reg.R16) { s.inst(andRM16R16, dst, src) }
func (s *Section) AndRM32R32(dst operand.RM32, src reg.R32) { s.inst(andRM32R32, dst, src) }
func (s *Section) AndRM64R64(dst operand.RM64, src reg.R64) { s.inst(andRM64R64, dst, src) }

func (s *Section) AndR8RM8(dst reg.R8, src operand.RM8)     { s.inst(andR8RM8, dst, src) }
func (s *Section) AndR16RM16(dst reg.R16, src operand.RM16) { s.inst(andR16RM16, dst, src) }
func (s *Section) AndR32RM32(dst reg.R32, src operand.RM32) { s.inst(andR32RM32, dst, src) }
func (s *Section) AndR64RM64(dst reg.R64, src operand.RM64) { s.inst(andR64RM64, dst, src) }

func (s *Section) AndALImm8(v int64)   { s.inst(andALImm8, fixAL, imm(v)) }
func (s *Section) AndAXImm16(v int64)  { s.inst(andAXImm16, fixAX, imm(v)) }
func (s *Section) AndEAXImm32(v int64) { s.inst(andEAXImm32, fixEAX, imm(v)) }
func (s *Section) AndRAXImm32(v int64) { s.inst(andRAXImm32, fixRAX, imm(v)) }

func (s *Section) AndRM8Imm8(dst operand.RM8, v int64)   { s.inst(andRM8Imm8, dst, imm(v)) }
func (s *Section) AndRM16Imm8(dst operand.RM16, v int64) { s.inst(andRM16Imm8, dst, imm(v)) }
func (s *Section) AndRM32Imm8(dst operand.RM32, v int64) { s.inst(andRM32Imm8, dst, imm(v)) }
func (s *Section) AndRM64Imm8(dst operand.RM64, v int64) { s.inst(andRM64Imm8, dst, imm(v)) }

func (s *Section) AndRM16Imm16(dst operand.RM16, v int64) { s.inst(andRM16Imm16, dst, imm(v)) }
func (s *Section) AndRM32Imm32(dst operand.RM32, v int64) { s.inst(andRM32Imm32, dst, imm(v)) }
func (s *Section) AndRM64Imm32(dst operand.RM64, v int64) { s.inst(andRM64Imm32, dst, imm(v)) }

// ---- SUB ------------------------------------------------------------------

var (
	subRM8R8     = form("SubRM8R8")
	subRM16R16   = form("SubRM16R16")
	subRM32R32   = form("SubRM32R32")
	subRM64R64   = form("SubRM64R64")
	subR8RM8     = form("SubR8RM8")
	subR16RM16   = form("SubR16RM16")
	subR32RM32   = form("SubR32RM32")
	subR64RM64   = form("SubR64RM64")
	subALImm8    = form("SubALImm8")
	subAXImm16   = form("SubAXImm16")
	subEAXImm32  = form("SubEAXImm32")
	subRAXImm32  = form("SubRAXImm32")
	subRM16Imm8  = form("SubRM16Imm8")
	subRM32Imm8  = form("SubRM32Imm8")
	subRM64Imm8  = form("SubRM64Imm8")
	subRM8Imm8   = form("SubRM8Imm8")
	subRM16Imm16 = form("SubRM16Imm16")
	subRM32Imm32 = form("SubRM32Imm32")
	subRM64Imm32 = form("SubRM64Imm32")
)

func (s *Section) SubRM8R8(dst operand.RM8, src reg.R8)     { s.inst(subRM8R8, dst, src) }
func (s *Section) SubRM16R16(dst operand.RM16, src reg.R16) { s.inst(subRM16R16, dst, src) }
func (s *Section) SubRM32R32(dst operand.RM32, src reg.R32) { s.inst(subRM32R32, dst, src) }
func (s *Section) SubRM64R64(dst operand.RM64, src reg.R64) { s.inst(subRM64R64, dst, src) }

func (s *Section) SubR8RM8(dst reg.R8, src operand.RM8)     { s.inst(subR8RM8, dst, src) }
func (s *Section) SubR16RM16(dst reg.R16, src operand.RM16) { s.inst(subR16RM16, dst, src) }
func (s *Section) SubR32RM32(dst reg.R32, src operand.RM32) { s.inst(subR32RM32, dst, src) }
func (s *Section) SubR64RM64(dst reg.R64, src operand.RM64) { s.inst(subR64RM64, dst, src) }

func (s *Section) SubALImm8(v int64)   { s.inst(subALImm8, fixAL, imm(v)) }
func (s *Section) SubAXImm16(v int64)  { s.inst(subAXImm16, fixAX, imm(v)) }
func (s *Section) SubEAXImm32(v int64) { s.inst(subEAXImm32, fixEAX, imm(v)) }
func (s *Section) SubRAXImm32(v int64) { s.inst(subRAXImm32, fixRAX, imm(v)) }

func (s *Section) SubRM8Imm8(dst operand.RM8, v int64)   { s.inst(subRM8Imm8, dst, imm(v)) }
func (s *Section) SubRM16Imm8(dst operand.RM16, v int64) { s.inst(subRM16Imm8, dst, imm(v)) }
func (s *Section) SubRM32Imm8(dst operand.RM32, v int64) { s.inst(subRM32Imm8, dst, imm(v)) }
func (s *Section) SubRM64Imm8(dst operand.RM64, v int64) { s.inst(subRM64Imm8, dst, imm(v)) }

func (s *Section) SubRM16Imm16(dst operand.RM16, v int64) { s.inst(subRM16Imm16, dst, imm(v)) }
func (s *Section) SubRM32Imm32(dst operand.RM32, v int64) { s.inst(subRM32Imm32, dst, imm(v)) }
func (s *Section) SubRM64Imm32(dst operand.RM64, v int64) { s.inst(subRM64Imm32, dst, imm(v)) }

// ---- XOR ------------------------------------------------------------------
//
// XorR32RM32(EAX, EAX) is the two-byte idiom that clears all sixty-four bits
// of RAX, because a 32-bit destination zeroes the upper half. XorR16RM16(AX,
// AX) clears sixteen and leaves the rest. Nothing here promotes the second
// into the first: the surprise is the ISA, and a lowering has to know it
// either way.

var (
	xorRM8R8     = form("XorRM8R8")
	xorRM16R16   = form("XorRM16R16")
	xorRM32R32   = form("XorRM32R32")
	xorRM64R64   = form("XorRM64R64")
	xorR8RM8     = form("XorR8RM8")
	xorR16RM16   = form("XorR16RM16")
	xorR32RM32   = form("XorR32RM32")
	xorR64RM64   = form("XorR64RM64")
	xorALImm8    = form("XorALImm8")
	xorAXImm16   = form("XorAXImm16")
	xorEAXImm32  = form("XorEAXImm32")
	xorRAXImm32  = form("XorRAXImm32")
	xorRM16Imm8  = form("XorRM16Imm8")
	xorRM32Imm8  = form("XorRM32Imm8")
	xorRM64Imm8  = form("XorRM64Imm8")
	xorRM8Imm8   = form("XorRM8Imm8")
	xorRM16Imm16 = form("XorRM16Imm16")
	xorRM32Imm32 = form("XorRM32Imm32")
	xorRM64Imm32 = form("XorRM64Imm32")
)

func (s *Section) XorRM8R8(dst operand.RM8, src reg.R8)     { s.inst(xorRM8R8, dst, src) }
func (s *Section) XorRM16R16(dst operand.RM16, src reg.R16) { s.inst(xorRM16R16, dst, src) }
func (s *Section) XorRM32R32(dst operand.RM32, src reg.R32) { s.inst(xorRM32R32, dst, src) }
func (s *Section) XorRM64R64(dst operand.RM64, src reg.R64) { s.inst(xorRM64R64, dst, src) }

func (s *Section) XorR8RM8(dst reg.R8, src operand.RM8)     { s.inst(xorR8RM8, dst, src) }
func (s *Section) XorR16RM16(dst reg.R16, src operand.RM16) { s.inst(xorR16RM16, dst, src) }
func (s *Section) XorR32RM32(dst reg.R32, src operand.RM32) { s.inst(xorR32RM32, dst, src) }
func (s *Section) XorR64RM64(dst reg.R64, src operand.RM64) { s.inst(xorR64RM64, dst, src) }

func (s *Section) XorALImm8(v int64)   { s.inst(xorALImm8, fixAL, imm(v)) }
func (s *Section) XorAXImm16(v int64)  { s.inst(xorAXImm16, fixAX, imm(v)) }
func (s *Section) XorEAXImm32(v int64) { s.inst(xorEAXImm32, fixEAX, imm(v)) }
func (s *Section) XorRAXImm32(v int64) { s.inst(xorRAXImm32, fixRAX, imm(v)) }

func (s *Section) XorRM8Imm8(dst operand.RM8, v int64)   { s.inst(xorRM8Imm8, dst, imm(v)) }
func (s *Section) XorRM16Imm8(dst operand.RM16, v int64) { s.inst(xorRM16Imm8, dst, imm(v)) }
func (s *Section) XorRM32Imm8(dst operand.RM32, v int64) { s.inst(xorRM32Imm8, dst, imm(v)) }
func (s *Section) XorRM64Imm8(dst operand.RM64, v int64) { s.inst(xorRM64Imm8, dst, imm(v)) }

func (s *Section) XorRM16Imm16(dst operand.RM16, v int64) { s.inst(xorRM16Imm16, dst, imm(v)) }
func (s *Section) XorRM32Imm32(dst operand.RM32, v int64) { s.inst(xorRM32Imm32, dst, imm(v)) }
func (s *Section) XorRM64Imm32(dst operand.RM64, v int64) { s.inst(xorRM64Imm32, dst, imm(v)) }

// ---- CMP ------------------------------------------------------------------
//
// CMP is the one of the eight with no locking clone. It writes nothing, so
// LOCK on it is #UD, and isa/table_base.go's aluOps says so with a false.

var (
	cmpRM8R8     = form("CmpRM8R8")
	cmpRM16R16   = form("CmpRM16R16")
	cmpRM32R32   = form("CmpRM32R32")
	cmpRM64R64   = form("CmpRM64R64")
	cmpR8RM8     = form("CmpR8RM8")
	cmpR16RM16   = form("CmpR16RM16")
	cmpR32RM32   = form("CmpR32RM32")
	cmpR64RM64   = form("CmpR64RM64")
	cmpALImm8    = form("CmpALImm8")
	cmpAXImm16   = form("CmpAXImm16")
	cmpEAXImm32  = form("CmpEAXImm32")
	cmpRAXImm32  = form("CmpRAXImm32")
	cmpRM16Imm8  = form("CmpRM16Imm8")
	cmpRM32Imm8  = form("CmpRM32Imm8")
	cmpRM64Imm8  = form("CmpRM64Imm8")
	cmpRM8Imm8   = form("CmpRM8Imm8")
	cmpRM16Imm16 = form("CmpRM16Imm16")
	cmpRM32Imm32 = form("CmpRM32Imm32")
	cmpRM64Imm32 = form("CmpRM64Imm32")
)

func (s *Section) CmpRM8R8(dst operand.RM8, src reg.R8)     { s.inst(cmpRM8R8, dst, src) }
func (s *Section) CmpRM16R16(dst operand.RM16, src reg.R16) { s.inst(cmpRM16R16, dst, src) }
func (s *Section) CmpRM32R32(dst operand.RM32, src reg.R32) { s.inst(cmpRM32R32, dst, src) }
func (s *Section) CmpRM64R64(dst operand.RM64, src reg.R64) { s.inst(cmpRM64R64, dst, src) }

func (s *Section) CmpR8RM8(dst reg.R8, src operand.RM8)     { s.inst(cmpR8RM8, dst, src) }
func (s *Section) CmpR16RM16(dst reg.R16, src operand.RM16) { s.inst(cmpR16RM16, dst, src) }
func (s *Section) CmpR32RM32(dst reg.R32, src operand.RM32) { s.inst(cmpR32RM32, dst, src) }
func (s *Section) CmpR64RM64(dst reg.R64, src operand.RM64) { s.inst(cmpR64RM64, dst, src) }

func (s *Section) CmpALImm8(v int64)   { s.inst(cmpALImm8, fixAL, imm(v)) }
func (s *Section) CmpAXImm16(v int64)  { s.inst(cmpAXImm16, fixAX, imm(v)) }
func (s *Section) CmpEAXImm32(v int64) { s.inst(cmpEAXImm32, fixEAX, imm(v)) }
func (s *Section) CmpRAXImm32(v int64) { s.inst(cmpRAXImm32, fixRAX, imm(v)) }

func (s *Section) CmpRM8Imm8(dst operand.RM8, v int64)   { s.inst(cmpRM8Imm8, dst, imm(v)) }
func (s *Section) CmpRM16Imm8(dst operand.RM16, v int64) { s.inst(cmpRM16Imm8, dst, imm(v)) }
func (s *Section) CmpRM32Imm8(dst operand.RM32, v int64) { s.inst(cmpRM32Imm8, dst, imm(v)) }
func (s *Section) CmpRM64Imm8(dst operand.RM64, v int64) { s.inst(cmpRM64Imm8, dst, imm(v)) }

func (s *Section) CmpRM16Imm16(dst operand.RM16, v int64) { s.inst(cmpRM16Imm16, dst, imm(v)) }
func (s *Section) CmpRM32Imm32(dst operand.RM32, v int64) { s.inst(cmpRM32Imm32, dst, imm(v)) }
func (s *Section) CmpRM64Imm32(dst operand.RM64, v int64) { s.inst(cmpRM64Imm32, dst, imm(v)) }
