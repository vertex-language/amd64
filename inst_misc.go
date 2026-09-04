package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// Everything else in the baseline: the no-ops and traps, the accumulator
// sign-extensions, the two gated bit-counting instructions, and the atomics.

var (
	nop     = form("Nop")
	nopRM32 = form("NopRM32")

	int3    = form("Int3")
	intImm8 = form("IntImm8")
	ud2     = form("Ud2")
	syscall = form("Syscall")
	cpuid   = form("Cpuid")

	hlt    = form("Hlt")
	cli    = form("Cli")
	sti    = form("Sti")
	pause  = form("Pause")
	rdtsc  = form("Rdtsc")
	rdmsr  = form("Rdmsr")
	wrmsr  = form("Wrmsr")
	wbinvd = form("Wbinvd")
)

// Nop is the instruction. The multi-byte sequences Align pads a code section
// with are not forms and are not reachable from here: choosing a sequence for
// a given length is arithmetic over a table of byte strings, not form
// resolution, so it lives in encode.Nops.
func (s *Section) Nop() { s.inst(nop) }

// NopRM32 is 0F 1F /0, the long nop with an operand. It is what those padding
// sequences are made of, and it is exported because a caller aligning by hand
// — inside a section rather than at its edges — needs to spell one.
func (s *Section) NopRM32(dst operand.RM32) { s.inst(nopRM32, dst) }

func (s *Section) Int3()           { s.inst(int3) }
func (s *Section) IntImm8(v int64) { s.inst(intImm8, imm(v)) }
func (s *Section) Ud2()            { s.inst(ud2) }
func (s *Section) Syscall()        { s.inst(syscall) }
func (s *Section) Cpuid()          { s.inst(cpuid) }

// The privileged and serializing instructions with no operands. Nothing in
// this tree selects one; they are here because inline assembly names them,
// and a kernel's idle loop, critical section and spin loop are three of them.

// Hlt halts until the next interrupt.
func (s *Section) Hlt() { s.inst(hlt) }

// Cli clears the interrupt flag; Sti sets it.
func (s *Section) Cli() { s.inst(cli) }
func (s *Section) Sti() { s.inst(sti) }

// Pause is the spin-loop hint: F3 90, which is a NOP with a prefix on every
// processor that does not know it.
func (s *Section) Pause() { s.inst(pause) }

// Rdtsc reads the timestamp counter into EDX:EAX.
func (s *Section) Rdtsc() { s.inst(rdtsc) }

// Rdmsr and Wrmsr read and write the model-specific register ECX names.
func (s *Section) Rdmsr() { s.inst(rdmsr) }
func (s *Section) Wrmsr() { s.inst(wrmsr) }

// Wbinvd writes back and invalidates the caches.
func (s *Section) Wbinvd() { s.inst(wbinvd) }

// ---- accumulator sign extension -------------------------------------------
//
// Three mnemonics over two opcodes, separated only by operand size, which is
// exactly the case where the fixed operands belong in the name. CWDE widens
// AX into EAX and CDQE widens EAX into RAX, both 98. CDQ splits EAX across
// EDX:EAX and CQO splits RAX across RDX:RAX, both 99 — and the pair before
// IdivRM64 is what makes a signed division correct rather than trapping.

var (
	cwde = form("Cwde")
	cdqe = form("Cdqe")
	cdq  = form("Cdq")
	cqo  = form("Cqo")
)

func (s *Section) Cwde() { s.inst(cwde) }
func (s *Section) Cdqe() { s.inst(cdqe) }
func (s *Section) Cdq()  { s.inst(cdq) }
func (s *Section) Cqo()  { s.inst(cqo) }

// ---- gated bit counting ---------------------------------------------------
//
// POPCNT and LZCNT are ordinary integer instructions that happen to carry a
// CPUID bit, which is why they sit in the baseline file rather than a tranche
// of their own. LZCNT is the one with a trap: on a processor without it the
// same bytes decode as BSR, which computes something different and does not
// fault, so the gate is the only thing standing between a v1 module and
// silently wrong output on old silicon.

var (
	popcntR32RM32 = form("PopcntR32RM32")
	popcntR64RM64 = form("PopcntR64RM64")
	lzcntR32RM32  = form("LzcntR32RM32")
	lzcntR64RM64  = form("LzcntR64RM64")
	tzcntR32RM32  = form("TzcntR32RM32")
	tzcntR64RM64  = form("TzcntR64RM64")

	bswapR32 = form("BswapR32")
	bswapR64 = form("BswapR64")
)

func (s *Section) PopcntR32RM32(dst reg.R32, src operand.RM32) { s.inst(popcntR32RM32, dst, src) }
func (s *Section) PopcntR64RM64(dst reg.R64, src operand.RM64) { s.inst(popcntR64RM64, dst, src) }
func (s *Section) LzcntR32RM32(dst reg.R32, src operand.RM32)  { s.inst(lzcntR32RM32, dst, src) }
func (s *Section) LzcntR64RM64(dst reg.R64, src operand.RM64)  { s.inst(lzcntR64RM64, dst, src) }
func (s *Section) TzcntR32RM32(dst reg.R32, src operand.RM32)  { s.inst(tzcntR32RM32, dst, src) }
func (s *Section) TzcntR64RM64(dst reg.R64, src operand.RM64)  { s.inst(tzcntR64RM64, dst, src) }

// BswapR32 emits BSWAP r32, which reverses the register's four bytes.
// The register rides in the opcode, so there is no memory form.
func (s *Section) BswapR32(dst reg.R32) { s.inst(bswapR32, dst) }

// BswapR64 emits BSWAP r64.
func (s *Section) BswapR64(dst reg.R64) { s.inst(bswapR64, dst) }

// ---- atomics --------------------------------------------------------------
//
// The unlocked forms. Every one of these has a locking clone in
// inst_lock.go, and on a multiprocessor the unlocked form of CMPXCHG or XADD
// is almost always a bug — but "almost always" is not "always", and a
// single-threaded compare-and-swap on thread-local state is a real thing to
// want. So both spellings exist and the caller says which.

var (
	cmpxchgRM8R8   = form("CmpxchgRM8R8")
	cmpxchgRM16R16 = form("CmpxchgRM16R16")
	cmpxchgRM32R32 = form("CmpxchgRM32R32")
	cmpxchgRM64R64 = form("CmpxchgRM64R64")

	xaddRM8R8   = form("XaddRM8R8")
	xaddRM16R16 = form("XaddRM16R16")
	xaddRM32R32 = form("XaddRM32R32")
	xaddRM64R64 = form("XaddRM64R64")

	cmpxchg16bM128 = form("Cmpxchg16bM128")
)

func (s *Section) CmpxchgRM8R8(dst operand.RM8, src reg.R8)     { s.inst(cmpxchgRM8R8, dst, src) }
func (s *Section) CmpxchgRM16R16(dst operand.RM16, src reg.R16) { s.inst(cmpxchgRM16R16, dst, src) }
func (s *Section) CmpxchgRM32R32(dst operand.RM32, src reg.R32) { s.inst(cmpxchgRM32R32, dst, src) }
func (s *Section) CmpxchgRM64R64(dst operand.RM64, src reg.R64) { s.inst(cmpxchgRM64R64, dst, src) }

func (s *Section) XaddRM8R8(dst operand.RM8, src reg.R8)     { s.inst(xaddRM8R8, dst, src) }
func (s *Section) XaddRM16R16(dst operand.RM16, src reg.R16) { s.inst(xaddRM16R16, dst, src) }
func (s *Section) XaddRM32R32(dst operand.RM32, src reg.R32) { s.inst(xaddRM32R32, dst, src) }
func (s *Section) XaddRM64R64(dst operand.RM64, src reg.R64) { s.inst(xaddRM64R64, dst, src) }

// Cmpxchg16bM128 takes M128 and not RM128, which is narrower than the row it
// binds. The register form of CMPXCHG16B is #UD — mod must not be 11 — so a
// signature that accepted an Xmm would accept something with no encoding, and
// the mismatch would surface as ErrForm from the encoder rather than as a
// compile error here.
//
// The table row says RM128 because the class vocabulary has no memory-only
// 128-bit class yet. Narrowing it there rather than here would be the better
// fix and belongs with whatever adds the first m-only vector row.
func (s *Section) Cmpxchg16bM128(dst operand.M128) { s.inst(cmpxchg16bM128, dst) }
