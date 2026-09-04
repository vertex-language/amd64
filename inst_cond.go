package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// The condition-code family: thirty conditions, three instruction families,
// five helpers each.
//
// Every documented spelling gets its own helper because every one gets its
// own row. JeLabel and JzLabel are both 0x74 and are not two names for one
// method — they are two forms with an AliasOf, so a listing can report which
// name the caller wrote while the bytes report what the silicon does. The
// same holds for the twelve other synonym pairs below: Jc and Jb, Jnae and
// Jb, Jae and Jnb and Jnc, and so on down the table.
//
// The Jcc split follows the branch rule from inst_branch.go. Short pins rel8
// at two bytes, the plain name pins rel32 at six, and nothing relaxes between
// them: a short conditional to a far target is ErrRange at Finalize. There is
// no ShortRef, for the same reason JmpShortRef does not exist.
//
// Setcc writes a byte and Cmovcc moves conditionally. Neither has a rel form,
// so neither has a Short spelling, and Cmovcc has no 8-bit width because the
// architecture does not give it one.

// ---- Jcc, rel8 ------------------------------------------------------------

var (
	joShort   = form("JoShortLabel")
	jnoShort  = form("JnoShortLabel")
	jbShort   = form("JbShortLabel")
	jcShort   = form("JcShortLabel")
	jnaeShort = form("JnaeShortLabel")
	jaeShort  = form("JaeShortLabel")
	jnbShort  = form("JnbShortLabel")
	jncShort  = form("JncShortLabel")
	jeShort   = form("JeShortLabel")
	jzShort   = form("JzShortLabel")
	jneShort  = form("JneShortLabel")
	jnzShort  = form("JnzShortLabel")
	jbeShort  = form("JbeShortLabel")
	jnaShort  = form("JnaShortLabel")
	jaShort   = form("JaShortLabel")
	jnbeShort = form("JnbeShortLabel")
	jsShort   = form("JsShortLabel")
	jnsShort  = form("JnsShortLabel")
	jpShort   = form("JpShortLabel")
	jpeShort  = form("JpeShortLabel")
	jnpShort  = form("JnpShortLabel")
	jpoShort  = form("JpoShortLabel")
	jlShort   = form("JlShortLabel")
	jngeShort = form("JngeShortLabel")
	jgeShort  = form("JgeShortLabel")
	jnlShort  = form("JnlShortLabel")
	jleShort  = form("JleShortLabel")
	jngShort  = form("JngShortLabel")
	jgShort   = form("JgShortLabel")
	jnleShort = form("JnleShortLabel")
)

func (s *Section) JoShortLabel(n string)   { s.inst(joShort, label(n)) }
func (s *Section) JnoShortLabel(n string)  { s.inst(jnoShort, label(n)) }
func (s *Section) JbShortLabel(n string)   { s.inst(jbShort, label(n)) }
func (s *Section) JcShortLabel(n string)   { s.inst(jcShort, label(n)) }
func (s *Section) JnaeShortLabel(n string) { s.inst(jnaeShort, label(n)) }
func (s *Section) JaeShortLabel(n string)  { s.inst(jaeShort, label(n)) }
func (s *Section) JnbShortLabel(n string)  { s.inst(jnbShort, label(n)) }
func (s *Section) JncShortLabel(n string)  { s.inst(jncShort, label(n)) }
func (s *Section) JeShortLabel(n string)   { s.inst(jeShort, label(n)) }
func (s *Section) JzShortLabel(n string)   { s.inst(jzShort, label(n)) }
func (s *Section) JneShortLabel(n string)  { s.inst(jneShort, label(n)) }
func (s *Section) JnzShortLabel(n string)  { s.inst(jnzShort, label(n)) }
func (s *Section) JbeShortLabel(n string)  { s.inst(jbeShort, label(n)) }
func (s *Section) JnaShortLabel(n string)  { s.inst(jnaShort, label(n)) }
func (s *Section) JaShortLabel(n string)   { s.inst(jaShort, label(n)) }
func (s *Section) JnbeShortLabel(n string) { s.inst(jnbeShort, label(n)) }
func (s *Section) JsShortLabel(n string)   { s.inst(jsShort, label(n)) }
func (s *Section) JnsShortLabel(n string)  { s.inst(jnsShort, label(n)) }
func (s *Section) JpShortLabel(n string)   { s.inst(jpShort, label(n)) }
func (s *Section) JpeShortLabel(n string)  { s.inst(jpeShort, label(n)) }
func (s *Section) JnpShortLabel(n string)  { s.inst(jnpShort, label(n)) }
func (s *Section) JpoShortLabel(n string)  { s.inst(jpoShort, label(n)) }
func (s *Section) JlShortLabel(n string)   { s.inst(jlShort, label(n)) }
func (s *Section) JngeShortLabel(n string) { s.inst(jngeShort, label(n)) }
func (s *Section) JgeShortLabel(n string)  { s.inst(jgeShort, label(n)) }
func (s *Section) JnlShortLabel(n string)  { s.inst(jnlShort, label(n)) }
func (s *Section) JleShortLabel(n string)  { s.inst(jleShort, label(n)) }
func (s *Section) JngShortLabel(n string)  { s.inst(jngShort, label(n)) }
func (s *Section) JgShortLabel(n string)   { s.inst(jgShort, label(n)) }
func (s *Section) JnleShortLabel(n string) { s.inst(jnleShort, label(n)) }

// ---- Jcc, rel32 -----------------------------------------------------------
//
// Each of these takes both a Label and a Ref, binding one row, for the reason
// given in inst_branch.go: Rel32 accepts either operand and a row has one
// helper name. A conditional branch that leaves the section is rare — it is
// usually a tail call in disguise — but it is expressible and refusing it
// here would be arbitrary.

var (
	jo   = form("JoLabel")
	jno  = form("JnoLabel")
	jb   = form("JbLabel")
	jc   = form("JcLabel")
	jnae = form("JnaeLabel")
	jae  = form("JaeLabel")
	jnb  = form("JnbLabel")
	jnc  = form("JncLabel")
	je   = form("JeLabel")
	jz   = form("JzLabel")
	jne  = form("JneLabel")
	jnz  = form("JnzLabel")
	jbe  = form("JbeLabel")
	jna  = form("JnaLabel")
	ja   = form("JaLabel")
	jnbe = form("JnbeLabel")
	js   = form("JsLabel")
	jns  = form("JnsLabel")
	jp   = form("JpLabel")
	jpe  = form("JpeLabel")
	jnp  = form("JnpLabel")
	jpo  = form("JpoLabel")
	jl   = form("JlLabel")
	jnge = form("JngeLabel")
	jge  = form("JgeLabel")
	jnl  = form("JnlLabel")
	jle  = form("JleLabel")
	jng  = form("JngLabel")
	jg   = form("JgLabel")
	jnle = form("JnleLabel")
)

func (s *Section) JoLabel(n string)   { s.inst(jo, label(n)) }
func (s *Section) JnoLabel(n string)  { s.inst(jno, label(n)) }
func (s *Section) JbLabel(n string)   { s.inst(jb, label(n)) }
func (s *Section) JcLabel(n string)   { s.inst(jc, label(n)) }
func (s *Section) JnaeLabel(n string) { s.inst(jnae, label(n)) }
func (s *Section) JaeLabel(n string)  { s.inst(jae, label(n)) }
func (s *Section) JnbLabel(n string)  { s.inst(jnb, label(n)) }
func (s *Section) JncLabel(n string)  { s.inst(jnc, label(n)) }
func (s *Section) JeLabel(n string)   { s.inst(je, label(n)) }
func (s *Section) JzLabel(n string)   { s.inst(jz, label(n)) }
func (s *Section) JneLabel(n string)  { s.inst(jne, label(n)) }
func (s *Section) JnzLabel(n string)  { s.inst(jnz, label(n)) }
func (s *Section) JbeLabel(n string)  { s.inst(jbe, label(n)) }
func (s *Section) JnaLabel(n string)  { s.inst(jna, label(n)) }
func (s *Section) JaLabel(n string)   { s.inst(ja, label(n)) }
func (s *Section) JnbeLabel(n string) { s.inst(jnbe, label(n)) }
func (s *Section) JsLabel(n string)   { s.inst(js, label(n)) }
func (s *Section) JnsLabel(n string)  { s.inst(jns, label(n)) }
func (s *Section) JpLabel(n string)   { s.inst(jp, label(n)) }
func (s *Section) JpeLabel(n string)  { s.inst(jpe, label(n)) }
func (s *Section) JnpLabel(n string)  { s.inst(jnp, label(n)) }
func (s *Section) JpoLabel(n string)  { s.inst(jpo, label(n)) }
func (s *Section) JlLabel(n string)   { s.inst(jl, label(n)) }
func (s *Section) JngeLabel(n string) { s.inst(jnge, label(n)) }
func (s *Section) JgeLabel(n string)  { s.inst(jge, label(n)) }
func (s *Section) JnlLabel(n string)  { s.inst(jnl, label(n)) }
func (s *Section) JleLabel(n string)  { s.inst(jle, label(n)) }
func (s *Section) JngLabel(n string)  { s.inst(jng, label(n)) }
func (s *Section) JgLabel(n string)   { s.inst(jg, label(n)) }
func (s *Section) JnleLabel(n string) { s.inst(jnle, label(n)) }

func (s *Section) JoRef(r SymRef)   { s.inst(jo, r) }
func (s *Section) JnoRef(r SymRef)  { s.inst(jno, r) }
func (s *Section) JbRef(r SymRef)   { s.inst(jb, r) }
func (s *Section) JcRef(r SymRef)   { s.inst(jc, r) }
func (s *Section) JnaeRef(r SymRef) { s.inst(jnae, r) }
func (s *Section) JaeRef(r SymRef)  { s.inst(jae, r) }
func (s *Section) JnbRef(r SymRef)  { s.inst(jnb, r) }
func (s *Section) JncRef(r SymRef)  { s.inst(jnc, r) }
func (s *Section) JeRef(r SymRef)   { s.inst(je, r) }
func (s *Section) JzRef(r SymRef)   { s.inst(jz, r) }
func (s *Section) JneRef(r SymRef)  { s.inst(jne, r) }
func (s *Section) JnzRef(r SymRef)  { s.inst(jnz, r) }
func (s *Section) JbeRef(r SymRef)  { s.inst(jbe, r) }
func (s *Section) JnaRef(r SymRef)  { s.inst(jna, r) }
func (s *Section) JaRef(r SymRef)   { s.inst(ja, r) }
func (s *Section) JnbeRef(r SymRef) { s.inst(jnbe, r) }
func (s *Section) JsRef(r SymRef)   { s.inst(js, r) }
func (s *Section) JnsRef(r SymRef)  { s.inst(jns, r) }
func (s *Section) JpRef(r SymRef)   { s.inst(jp, r) }
func (s *Section) JpeRef(r SymRef)  { s.inst(jpe, r) }
func (s *Section) JnpRef(r SymRef)  { s.inst(jnp, r) }
func (s *Section) JpoRef(r SymRef)  { s.inst(jpo, r) }
func (s *Section) JlRef(r SymRef)   { s.inst(jl, r) }
func (s *Section) JngeRef(r SymRef) { s.inst(jnge, r) }
func (s *Section) JgeRef(r SymRef)  { s.inst(jge, r) }
func (s *Section) JnlRef(r SymRef)  { s.inst(jnl, r) }
func (s *Section) JleRef(r SymRef)  { s.inst(jle, r) }
func (s *Section) JngRef(r SymRef)  { s.inst(jng, r) }
func (s *Section) JgRef(r SymRef)   { s.inst(jg, r) }
func (s *Section) JnleRef(r SymRef) { s.inst(jnle, r) }

// ---- SETcc ----------------------------------------------------------------
//
// Writes 1 or 0 into a byte. The destination is r/m8, which means the
// high-byte rule applies: SetzRM8 handed AH is fine on its own and ErrOperand
// the moment anything in the instruction forces a REX prefix.

var (
	seto   = form("SetoRM8")
	setno  = form("SetnoRM8")
	setb   = form("SetbRM8")
	setc   = form("SetcRM8")
	setnae = form("SetnaeRM8")
	setae  = form("SetaeRM8")
	setnb  = form("SetnbRM8")
	setnc  = form("SetncRM8")
	sete   = form("SeteRM8")
	setz   = form("SetzRM8")
	setne  = form("SetneRM8")
	setnz  = form("SetnzRM8")
	setbe  = form("SetbeRM8")
	setna  = form("SetnaRM8")
	seta   = form("SetaRM8")
	setnbe = form("SetnbeRM8")
	sets   = form("SetsRM8")
	setns  = form("SetnsRM8")
	setp   = form("SetpRM8")
	setpe  = form("SetpeRM8")
	setnp  = form("SetnpRM8")
	setpo  = form("SetpoRM8")
	setl   = form("SetlRM8")
	setnge = form("SetngeRM8")
	setge  = form("SetgeRM8")
	setnl  = form("SetnlRM8")
	setle  = form("SetleRM8")
	setng  = form("SetngRM8")
	setg   = form("SetgRM8")
	setnle = form("SetnleRM8")
)

func (s *Section) SetoRM8(d operand.RM8)   { s.inst(seto, d) }
func (s *Section) SetnoRM8(d operand.RM8)  { s.inst(setno, d) }
func (s *Section) SetbRM8(d operand.RM8)   { s.inst(setb, d) }
func (s *Section) SetcRM8(d operand.RM8)   { s.inst(setc, d) }
func (s *Section) SetnaeRM8(d operand.RM8) { s.inst(setnae, d) }
func (s *Section) SetaeRM8(d operand.RM8)  { s.inst(setae, d) }
func (s *Section) SetnbRM8(d operand.RM8)  { s.inst(setnb, d) }
func (s *Section) SetncRM8(d operand.RM8)  { s.inst(setnc, d) }
func (s *Section) SeteRM8(d operand.RM8)   { s.inst(sete, d) }
func (s *Section) SetzRM8(d operand.RM8)   { s.inst(setz, d) }
func (s *Section) SetneRM8(d operand.RM8)  { s.inst(setne, d) }
func (s *Section) SetnzRM8(d operand.RM8)  { s.inst(setnz, d) }
func (s *Section) SetbeRM8(d operand.RM8)  { s.inst(setbe, d) }
func (s *Section) SetnaRM8(d operand.RM8)  { s.inst(setna, d) }
func (s *Section) SetaRM8(d operand.RM8)   { s.inst(seta, d) }
func (s *Section) SetnbeRM8(d operand.RM8) { s.inst(setnbe, d) }
func (s *Section) SetsRM8(d operand.RM8)   { s.inst(sets, d) }
func (s *Section) SetnsRM8(d operand.RM8)  { s.inst(setns, d) }
func (s *Section) SetpRM8(d operand.RM8)   { s.inst(setp, d) }
func (s *Section) SetpeRM8(d operand.RM8)  { s.inst(setpe, d) }
func (s *Section) SetnpRM8(d operand.RM8)  { s.inst(setnp, d) }
func (s *Section) SetpoRM8(d operand.RM8)  { s.inst(setpo, d) }
func (s *Section) SetlRM8(d operand.RM8)   { s.inst(setl, d) }
func (s *Section) SetngeRM8(d operand.RM8) { s.inst(setnge, d) }
func (s *Section) SetgeRM8(d operand.RM8)  { s.inst(setge, d) }
func (s *Section) SetnlRM8(d operand.RM8)  { s.inst(setnl, d) }
func (s *Section) SetleRM8(d operand.RM8)  { s.inst(setle, d) }
func (s *Section) SetngRM8(d operand.RM8)  { s.inst(setng, d) }
func (s *Section) SetgRM8(d operand.RM8)   { s.inst(setg, d) }
func (s *Section) SetnleRM8(d operand.RM8) { s.inst(setnle, d) }

// ---- CMOVcc ---------------------------------------------------------------
//
// Two widths and no 8-bit one, because the architecture has none. The 32-bit
// form zeroes the upper half of its destination whether or not the condition
// holds, which is the usual reason a lowering reaches for the 64-bit form
// even when the values are narrow.

var (
	cmovo32, cmovo64     = form("CmovoR32RM32"), form("CmovoR64RM64")
	cmovno32, cmovno64   = form("CmovnoR32RM32"), form("CmovnoR64RM64")
	cmovb32, cmovb64     = form("CmovbR32RM32"), form("CmovbR64RM64")
	cmovc32, cmovc64     = form("CmovcR32RM32"), form("CmovcR64RM64")
	cmovnae32, cmovnae64 = form("CmovnaeR32RM32"), form("CmovnaeR64RM64")
	cmovae32, cmovae64   = form("CmovaeR32RM32"), form("CmovaeR64RM64")
	cmovnb32, cmovnb64   = form("CmovnbR32RM32"), form("CmovnbR64RM64")
	cmovnc32, cmovnc64   = form("CmovncR32RM32"), form("CmovncR64RM64")
	cmove32, cmove64     = form("CmoveR32RM32"), form("CmoveR64RM64")
	cmovz32, cmovz64     = form("CmovzR32RM32"), form("CmovzR64RM64")
	cmovne32, cmovne64   = form("CmovneR32RM32"), form("CmovneR64RM64")
	cmovnz32, cmovnz64   = form("CmovnzR32RM32"), form("CmovnzR64RM64")
	cmovbe32, cmovbe64   = form("CmovbeR32RM32"), form("CmovbeR64RM64")
	cmovna32, cmovna64   = form("CmovnaR32RM32"), form("CmovnaR64RM64")
	cmova32, cmova64     = form("CmovaR32RM32"), form("CmovaR64RM64")
	cmovnbe32, cmovnbe64 = form("CmovnbeR32RM32"), form("CmovnbeR64RM64")
	cmovs32, cmovs64     = form("CmovsR32RM32"), form("CmovsR64RM64")
	cmovns32, cmovns64   = form("CmovnsR32RM32"), form("CmovnsR64RM64")
	cmovp32, cmovp64     = form("CmovpR32RM32"), form("CmovpR64RM64")
	cmovpe32, cmovpe64   = form("CmovpeR32RM32"), form("CmovpeR64RM64")
	cmovnp32, cmovnp64   = form("CmovnpR32RM32"), form("CmovnpR64RM64")
	cmovpo32, cmovpo64   = form("CmovpoR32RM32"), form("CmovpoR64RM64")
	cmovl32, cmovl64     = form("CmovlR32RM32"), form("CmovlR64RM64")
	cmovnge32, cmovnge64 = form("CmovngeR32RM32"), form("CmovngeR64RM64")
	cmovge32, cmovge64   = form("CmovgeR32RM32"), form("CmovgeR64RM64")
	cmovnl32, cmovnl64   = form("CmovnlR32RM32"), form("CmovnlR64RM64")
	cmovle32, cmovle64   = form("CmovleR32RM32"), form("CmovleR64RM64")
	cmovng32, cmovng64   = form("CmovngR32RM32"), form("CmovngR64RM64")
	cmovg32, cmovg64     = form("CmovgR32RM32"), form("CmovgR64RM64")
	cmovnle32, cmovnle64 = form("CmovnleR32RM32"), form("CmovnleR64RM64")
)

func (s *Section) CmovoR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovo32, d, x) }
func (s *Section) CmovnoR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovno32, d, x) }
func (s *Section) CmovbR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovb32, d, x) }
func (s *Section) CmovcR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovc32, d, x) }
func (s *Section) CmovnaeR32RM32(d reg.R32, x operand.RM32) { s.inst(cmovnae32, d, x) }
func (s *Section) CmovaeR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovae32, d, x) }
func (s *Section) CmovnbR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovnb32, d, x) }
func (s *Section) CmovncR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovnc32, d, x) }
func (s *Section) CmoveR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmove32, d, x) }
func (s *Section) CmovzR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovz32, d, x) }
func (s *Section) CmovneR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovne32, d, x) }
func (s *Section) CmovnzR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovnz32, d, x) }
func (s *Section) CmovbeR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovbe32, d, x) }
func (s *Section) CmovnaR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovna32, d, x) }
func (s *Section) CmovaR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmova32, d, x) }
func (s *Section) CmovnbeR32RM32(d reg.R32, x operand.RM32) { s.inst(cmovnbe32, d, x) }
func (s *Section) CmovsR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovs32, d, x) }
func (s *Section) CmovnsR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovns32, d, x) }
func (s *Section) CmovpR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovp32, d, x) }
func (s *Section) CmovpeR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovpe32, d, x) }
func (s *Section) CmovnpR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovnp32, d, x) }
func (s *Section) CmovpoR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovpo32, d, x) }
func (s *Section) CmovlR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovl32, d, x) }
func (s *Section) CmovngeR32RM32(d reg.R32, x operand.RM32) { s.inst(cmovnge32, d, x) }
func (s *Section) CmovgeR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovge32, d, x) }
func (s *Section) CmovnlR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovnl32, d, x) }
func (s *Section) CmovleR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovle32, d, x) }
func (s *Section) CmovngR32RM32(d reg.R32, x operand.RM32)  { s.inst(cmovng32, d, x) }
func (s *Section) CmovgR32RM32(d reg.R32, x operand.RM32)   { s.inst(cmovg32, d, x) }
func (s *Section) CmovnleR32RM32(d reg.R32, x operand.RM32) { s.inst(cmovnle32, d, x) }

func (s *Section) CmovoR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovo64, d, x) }
func (s *Section) CmovnoR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovno64, d, x) }
func (s *Section) CmovbR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovb64, d, x) }
func (s *Section) CmovcR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovc64, d, x) }
func (s *Section) CmovnaeR64RM64(d reg.R64, x operand.RM64) { s.inst(cmovnae64, d, x) }
func (s *Section) CmovaeR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovae64, d, x) }
func (s *Section) CmovnbR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovnb64, d, x) }
func (s *Section) CmovncR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovnc64, d, x) }
func (s *Section) CmoveR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmove64, d, x) }
func (s *Section) CmovzR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovz64, d, x) }
func (s *Section) CmovneR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovne64, d, x) }
func (s *Section) CmovnzR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovnz64, d, x) }
func (s *Section) CmovbeR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovbe64, d, x) }
func (s *Section) CmovnaR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovna64, d, x) }
func (s *Section) CmovaR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmova64, d, x) }
func (s *Section) CmovnbeR64RM64(d reg.R64, x operand.RM64) { s.inst(cmovnbe64, d, x) }
func (s *Section) CmovsR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovs64, d, x) }
func (s *Section) CmovnsR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovns64, d, x) }
func (s *Section) CmovpR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovp64, d, x) }
func (s *Section) CmovpeR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovpe64, d, x) }
func (s *Section) CmovnpR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovnp64, d, x) }
func (s *Section) CmovpoR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovpo64, d, x) }
func (s *Section) CmovlR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovl64, d, x) }
func (s *Section) CmovngeR64RM64(d reg.R64, x operand.RM64) { s.inst(cmovnge64, d, x) }
func (s *Section) CmovgeR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovge64, d, x) }
func (s *Section) CmovnlR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovnl64, d, x) }
func (s *Section) CmovleR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovle64, d, x) }
func (s *Section) CmovngR64RM64(d reg.R64, x operand.RM64)  { s.inst(cmovng64, d, x) }
func (s *Section) CmovgR64RM64(d reg.R64, x operand.RM64)   { s.inst(cmovg64, d, x) }
func (s *Section) CmovnleR64RM64(d reg.R64, x operand.RM64) { s.inst(cmovnle64, d, x) }
