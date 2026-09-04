package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// The bit-test, double-shift, port and system instructions, opposite
// internal/isa/table_system.go.
//
// Nothing in this tree selects any of these. They are here because inline
// assembly names them: a bit test is an AND with a mask by the time an IR has
// finished with it, and no instruction selector has ever emitted HLT.

var (
	btRM32R32   = form("BtRM32R32")
	btRM64R64   = form("BtRM64R64")
	btRM32Imm8  = form("BtRM32Imm8")
	btRM64Imm8  = form("BtRM64Imm8")
	btsRM32R32  = form("BtsRM32R32")
	btsRM64R64  = form("BtsRM64R64")
	btsRM32Imm8 = form("BtsRM32Imm8")
	btsRM64Imm8 = form("BtsRM64Imm8")
	btrRM32R32  = form("BtrRM32R32")
	btrRM64R64  = form("BtrRM64R64")
	btrRM32Imm8 = form("BtrRM32Imm8")
	btrRM64Imm8 = form("BtrRM64Imm8")
	btcRM32R32  = form("BtcRM32R32")
	btcRM64R64  = form("BtcRM64R64")
	btcRM32Imm8 = form("BtcRM32Imm8")
	btcRM64Imm8 = form("BtcRM64Imm8")
)

// BtRM32R32 emits BT r/m32, r32: copy the numbered bit into CF. The bit
// number is taken modulo the operand's width for a register destination and
// is *not* for a memory one, where it addresses beyond the four bytes named.
func (s *Section) BtRM32R32(dst operand.RM32, bit reg.R32) { s.inst(btRM32R32, dst, bit) }
func (s *Section) BtRM64R64(dst operand.RM64, bit reg.R64) { s.inst(btRM64R64, dst, bit) }
func (s *Section) BtRM32Imm8(dst operand.RM32, bit int64)  { s.inst(btRM32Imm8, dst, imm(bit)) }
func (s *Section) BtRM64Imm8(dst operand.RM64, bit int64)  { s.inst(btRM64Imm8, dst, imm(bit)) }

// BtsRM32R32 emits BTS: test and set. BTR clears, BTC complements.
func (s *Section) BtsRM32R32(dst operand.RM32, bit reg.R32) { s.inst(btsRM32R32, dst, bit) }
func (s *Section) BtsRM64R64(dst operand.RM64, bit reg.R64) { s.inst(btsRM64R64, dst, bit) }
func (s *Section) BtsRM32Imm8(dst operand.RM32, bit int64)  { s.inst(btsRM32Imm8, dst, imm(bit)) }
func (s *Section) BtsRM64Imm8(dst operand.RM64, bit int64)  { s.inst(btsRM64Imm8, dst, imm(bit)) }

func (s *Section) BtrRM32R32(dst operand.RM32, bit reg.R32) { s.inst(btrRM32R32, dst, bit) }
func (s *Section) BtrRM64R64(dst operand.RM64, bit reg.R64) { s.inst(btrRM64R64, dst, bit) }
func (s *Section) BtrRM32Imm8(dst operand.RM32, bit int64)  { s.inst(btrRM32Imm8, dst, imm(bit)) }
func (s *Section) BtrRM64Imm8(dst operand.RM64, bit int64)  { s.inst(btrRM64Imm8, dst, imm(bit)) }

func (s *Section) BtcRM32R32(dst operand.RM32, bit reg.R32) { s.inst(btcRM32R32, dst, bit) }
func (s *Section) BtcRM64R64(dst operand.RM64, bit reg.R64) { s.inst(btcRM64R64, dst, bit) }
func (s *Section) BtcRM32Imm8(dst operand.RM32, bit int64)  { s.inst(btcRM32Imm8, dst, imm(bit)) }
func (s *Section) BtcRM64Imm8(dst operand.RM64, bit int64)  { s.inst(btcRM64Imm8, dst, imm(bit)) }

var (
	bsfR32RM32 = form("BsfR32RM32")
	bsfR64RM64 = form("BsfR64RM64")
	bsrR32RM32 = form("BsrR32RM32")
	bsrR64RM64 = form("BsrR64RM64")
)

// BsfR32RM32 emits BSF: the index of the lowest set bit, with ZF set when
// the source is zero and the destination then undefined. TZCNT is the
// well-defined one and is gated behind BMI1; this is the one every processor
// has.
func (s *Section) BsfR32RM32(dst reg.R32, src operand.RM32) { s.inst(bsfR32RM32, dst, src) }
func (s *Section) BsfR64RM64(dst reg.R64, src operand.RM64) { s.inst(bsfR64RM64, dst, src) }

// BsrR32RM32 emits BSR: the index of the highest set bit.
func (s *Section) BsrR32RM32(dst reg.R32, src operand.RM32) { s.inst(bsrR32RM32, dst, src) }
func (s *Section) BsrR64RM64(dst reg.R64, src operand.RM64) { s.inst(bsrR64RM64, dst, src) }

var (
	shldRM32R32Imm8 = form("ShldRM32R32Imm8")
	shldRM64R64Imm8 = form("ShldRM64R64Imm8")
	shldRM32R32CL   = form("ShldRM32R32CL")
	shldRM64R64CL   = form("ShldRM64R64CL")
	shrdRM32R32Imm8 = form("ShrdRM32R32Imm8")
	shrdRM64R64Imm8 = form("ShrdRM64R64Imm8")
	shrdRM32R32CL   = form("ShrdRM32R32CL")
	shrdRM64R64CL   = form("ShrdRM64R64CL")
)

// ShldRM32R32Imm8 emits SHLD r/m32, r32, imm8: shift the destination left,
// filling from the top of the source. It is how a 64-bit shift is written on
// a 32-bit machine and how a bitfield is moved across a word boundary.
func (s *Section) ShldRM32R32Imm8(dst operand.RM32, src reg.R32, n int64) {
	s.inst(shldRM32R32Imm8, dst, src, imm(n))
}

func (s *Section) ShldRM64R64Imm8(dst operand.RM64, src reg.R64, n int64) {
	s.inst(shldRM64R64Imm8, dst, src, imm(n))
}

// ShldRM32R32CL is the same with the count in CL, which the form names and
// this signature therefore does not.
func (s *Section) ShldRM32R32CL(dst operand.RM32, src reg.R32) {
	s.inst(shldRM32R32CL, dst, src, reg.CL)
}

func (s *Section) ShldRM64R64CL(dst operand.RM64, src reg.R64) {
	s.inst(shldRM64R64CL, dst, src, reg.CL)
}

func (s *Section) ShrdRM32R32Imm8(dst operand.RM32, src reg.R32, n int64) {
	s.inst(shrdRM32R32Imm8, dst, src, imm(n))
}

func (s *Section) ShrdRM64R64Imm8(dst operand.RM64, src reg.R64, n int64) {
	s.inst(shrdRM64R64Imm8, dst, src, imm(n))
}

func (s *Section) ShrdRM32R32CL(dst operand.RM32, src reg.R32) {
	s.inst(shrdRM32R32CL, dst, src, reg.CL)
}

func (s *Section) ShrdRM64R64CL(dst operand.RM64, src reg.R64) {
	s.inst(shrdRM64R64CL, dst, src, reg.CL)
}

var (
	pushfq = form("Pushfq")
	popfq  = form("Popfq")
)

// Pushfq pushes RFLAGS. In 64-bit mode the operand size is fixed at eight
// bytes, which is why there is no other spelling of it.
func (s *Section) Pushfq() { s.inst(pushfq) }
func (s *Section) Popfq()  { s.inst(popfq) }

var (
	inALImm8   = form("InALImm8")
	inAXImm8   = form("InAXImm8")
	inEAXImm8  = form("InEAXImm8")
	inALDX     = form("InALDX")
	inAXDX     = form("InAXDX")
	inEAXDX    = form("InEAXDX")
	outImm8AL  = form("OutImm8AL")
	outImm8AX  = form("OutImm8AX")
	outImm8EAX = form("OutImm8EAX")
	outDXAL    = form("OutDXAL")
	outDXAX    = form("OutDXAX")
	outDXEAX   = form("OutDXEAX")
)

// InALImm8 emits IN AL, imm8: read a byte from a port whose number is an
// immediate. Both registers are named by the form and neither is a parameter.
func (s *Section) InALImm8(port int64)  { s.inst(inALImm8, reg.AL, imm(port)) }
func (s *Section) InAXImm8(port int64)  { s.inst(inAXImm8, reg.AX, imm(port)) }
func (s *Section) InEAXImm8(port int64) { s.inst(inEAXImm8, reg.EAX, imm(port)) }

// InALDX reads from the port DX names, which is the only way to reach one
// above 255.
func (s *Section) InALDX()  { s.inst(inALDX, reg.AL, reg.DX) }
func (s *Section) InAXDX()  { s.inst(inAXDX, reg.AX, reg.DX) }
func (s *Section) InEAXDX() { s.inst(inEAXDX, reg.EAX, reg.DX) }

func (s *Section) OutImm8AL(port int64)  { s.inst(outImm8AL, imm(port), reg.AL) }
func (s *Section) OutImm8AX(port int64)  { s.inst(outImm8AX, imm(port), reg.AX) }
func (s *Section) OutImm8EAX(port int64) { s.inst(outImm8EAX, imm(port), reg.EAX) }
func (s *Section) OutDXAL()              { s.inst(outDXAL, reg.DX, reg.AL) }
func (s *Section) OutDXAX()              { s.inst(outDXAX, reg.DX, reg.AX) }
func (s *Section) OutDXEAX()             { s.inst(outDXEAX, reg.DX, reg.EAX) }

var (
	cld     = form("Cld")
	std     = form("Std")
	swapgs  = form("Swapgs")
	rdtscp  = form("Rdtscp")
	clac    = form("Clac")
	stac    = form("Stac")
	xgetbv  = form("Xgetbv")
	xsetbv  = form("Xsetbv")
	iretq   = form("Iretq")
	sysretq = form("Sysretq")
)

// Cld clears the direction flag; Std sets it. A string instruction reads it,
// and an ABI that does not require it clear is rare enough that clearing it
// before a REP is ordinary.
func (s *Section) Cld() { s.inst(cld) }
func (s *Section) Std() { s.inst(std) }

// Swapgs exchanges GS.base with the kernel's saved value, which is the first
// instruction of every SYSCALL entry path.
func (s *Section) Swapgs() { s.inst(swapgs) }

// Rdtscp is RDTSC with the processor id, and serializing on the read side.
func (s *Section) Rdtscp() { s.inst(rdtscp) }

// Clac and Stac clear and set the alignment-check-for-user-access flag.
func (s *Section) Clac() { s.inst(clac) }
func (s *Section) Stac() { s.inst(stac) }

// Xgetbv and Xsetbv read and write the extended control register ECX names.
func (s *Section) Xgetbv() { s.inst(xgetbv) }
func (s *Section) Xsetbv() { s.inst(xsetbv) }

// Iretq returns from an interrupt to 64-bit code; Sysretq from a syscall.
// The q in each is REX.W and nothing else.
func (s *Section) Iretq()   { s.inst(iretq) }
func (s *Section) Sysretq() { s.inst(sysretq) }

var (
	sldtRM16 = form("SldtRM16")
	strRM16  = form("StrRM16")
	lldtRM16 = form("LldtRM16")
	ltrRM16  = form("LtrRM16")
)

// LtrRM16 loads the task register; StrRM16 stores it. LldtRM16 and SldtRM16
// are the same pair for the local descriptor table.
func (s *Section) LtrRM16(src operand.RM16)  { s.inst(ltrRM16, src) }
func (s *Section) StrRM16(dst operand.RM16)  { s.inst(strRM16, dst) }
func (s *Section) LldtRM16(src operand.RM16) { s.inst(lldtRM16, src) }
func (s *Section) SldtRM16(dst operand.RM16) { s.inst(sldtRM16, dst) }

var (
	sgdtM        = form("SgdtM")
	sidtM        = form("SidtM")
	lgdtM        = form("LgdtM")
	lidtM        = form("LidtM")
	invlpgM      = form("InvlpgM")
	clflushM     = form("ClflushM")
	fxsaveM      = form("FxsaveM")
	fxrstorM     = form("FxrstorM")
	xsaveM       = form("XsaveM")
	xrstorM      = form("XrstorM")
	prefetchntaM = form("PrefetchntaM")
	prefetcht0M  = form("Prefetcht0M")
	prefetcht1M  = form("Prefetcht1M")
	prefetcht2M  = form("Prefetcht2M")
)

// LgdtM loads the global descriptor table register from the address given.
// The operand takes a Memory rather than an RM: there is no register form,
// and it has no access width of its own — how much LGDT reads is a fact
// about the mode, not about the operand.
func (s *Section) LgdtM(src operand.Memory) { s.inst(lgdtM, src) }
func (s *Section) LidtM(src operand.Memory) { s.inst(lidtM, src) }
func (s *Section) SgdtM(dst operand.Memory) { s.inst(sgdtM, dst) }
func (s *Section) SidtM(dst operand.Memory) { s.inst(sidtM, dst) }

// InvlpgM invalidates the TLB entry for one page.
func (s *Section) InvlpgM(addr operand.Memory) { s.inst(invlpgM, addr) }

// ClflushM writes back and invalidates the cache line containing the address.
func (s *Section) ClflushM(addr operand.Memory) { s.inst(clflushM, addr) }

// FxsaveM and the three beside it save and restore the floating-point and
// vector state. Their operands are areas, not values, which is why the width
// is not in the name.
func (s *Section) FxsaveM(dst operand.Memory)  { s.inst(fxsaveM, dst) }
func (s *Section) FxrstorM(src operand.Memory) { s.inst(fxrstorM, src) }
func (s *Section) XsaveM(dst operand.Memory)   { s.inst(xsaveM, dst) }
func (s *Section) XrstorM(src operand.Memory)  { s.inst(xrstorM, src) }

// The prefetch hints, which differ only in which cache level they ask for.
func (s *Section) PrefetchntaM(addr operand.Memory) { s.inst(prefetchntaM, addr) }
func (s *Section) Prefetcht0M(addr operand.Memory)  { s.inst(prefetcht0M, addr) }
func (s *Section) Prefetcht1M(addr operand.Memory)  { s.inst(prefetcht1M, addr) }
func (s *Section) Prefetcht2M(addr operand.Memory)  { s.inst(prefetcht2M, addr) }

var (
	movntiMR32 = form("MovntiMR32")
	movntiMR64 = form("MovntiMR64")
)

// MovntiMR32 emits MOVNTI: a store that does not allocate a cache line. Its
// destination is memory and only memory — missing the cache is the point,
// and a register has none.
func (s *Section) MovntiMR32(dst operand.Memory, src reg.R32) { s.inst(movntiMR32, dst, src) }
func (s *Section) MovntiMR64(dst operand.Memory, src reg.R64) { s.inst(movntiMR64, dst, src) }
