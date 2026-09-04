package amd64

import (
	"github.com/vertex-language/amd64/operand"
)

// Branches and calls.
//
// Two things split these helpers, and neither is a width.
//
// Where the target resolves. A Label is same-section: it is patched at
// Finalize, leaves no relocation, and a linker never sees it. A Ref crosses,
// survives into Refs(), and carries the link semantics the caller stated.
// JmpLabel and JmpRef are the same table row — Rel32 accepts either operand —
// so they bind one form and differ only in what they wrap. That is why form()
// is a lookup rather than an ownership claim, and why helpers_test.go asks
// that every form have a method rather than exactly one.
//
// How wide the displacement is. Short pins rel8 and the plain name pins
// rel32, with no relaxation between them: a short branch to a far target is
// ErrRange at Finalize, loud, rather than two bytes silently becoming five.
// In the small code model rel32 reaches everything, so the absence costs
// nothing until you are past two gigabytes of text.

var (
	jmpShortRel = form("JmpShortLabel")
	jmpRel      = form("JmpLabel")
	jmpRM64     = form("JmpRM64")

	callRel  = form("CallLabel")
	callRM64 = form("CallRM64")

	ret      = form("Ret")
	retImm16 = form("RetImm16")

	loopRel   = form("LoopLabel")
	loopeRel  = form("LoopeLabel")
	loopneRel = form("LoopneLabel")
	jrcxzRel  = form("JrcxzLabel")
)

func (s *Section) JmpLabel(name string) { s.inst(jmpRel, label(name)) }
func (s *Section) JmpRef(r SymRef)      { s.inst(jmpRel, r) }

// JmpShortLabel is the two-byte form, and taking it is a claim about
// distance. There is no JmpShortRef: a rel8 that crosses a section would need
// a one-byte relocation, and of the three containers only ELF has one — so
// the object would build and then be refused by two writers out of three,
// which is a worse place to find out than here.
func (s *Section) JmpShortLabel(name string) { s.inst(jmpShortRel, label(name)) }

// JmpRM64 is the indirect jump: a jump table's tail, or the far half of the
// large code model.
func (s *Section) JmpRM64(target operand.RM64) { s.inst(jmpRM64, target) }

// CallLabel calls a function compiled into this module. CallRef is the one
// that crosses, and the kind rides beside the bytes: call puts@plt and call
// puts are byte-identical, e8 either way, so RefPLT32 versus RefPC32 is a
// decision the lowering states rather than one this package infers.
func (s *Section) CallLabel(name string) { s.inst(callRel, label(name)) }
func (s *Section) CallRef(r SymRef)      { s.inst(callRel, r) }

func (s *Section) CallRM64(target operand.RM64) { s.inst(callRM64, target) }

// Ret emits c3 and asks no questions about who set up the frame. RetImm16 is
// the form that pops arguments on the way out — stdcall's, and nothing SysV
// or Microsoft x64 emits.
func (s *Section) Ret()             { s.inst(ret) }
func (s *Section) RetImm16(v int64) { s.inst(retImm16, imm(v)) }

// The rel8-only mnemonics take the plain name, because there is no wider
// encoding for Short to distinguish them from. All four are two bytes and all
// four fail with ErrRange at Finalize if the target is more than 127 bytes
// away, which for a loop body is the common case sooner than you would like.
func (s *Section) LoopLabel(name string)   { s.inst(loopRel, label(name)) }
func (s *Section) LoopeLabel(name string)  { s.inst(loopeRel, label(name)) }
func (s *Section) LoopneLabel(name string) { s.inst(loopneRel, label(name)) }

// JrcxzLabel branches on RCX being zero without touching flags, which is the
// only reason to reach for it over TestRM64R64 and JzLabel.
func (s *Section) JrcxzLabel(name string) { s.inst(jrcxzRel, label(name)) }
