package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// The stack.
//
// PUSH and POP default to a 64-bit operand size in long mode, and a 32-bit
// one is not encodable at all. So there is no REX.W on any of these, no
// 32-bit helper to leave out, and no width parameter: the size is not a
// choice the encoding offers, which is why these are the shortest signatures
// in the tree.
//
// What is not here is a frame. Nothing in this package knows who set up the
// stack or who tears it down, because that is a calling convention and this
// module has an instruction set instead. SysV's 128-byte red zone and
// Microsoft x64's 32 bytes of shadow space are both facts about your target.

var (
	pushR64   = form("PushR64")
	pushRM64  = form("PushRM64")
	pushImm8  = form("PushImm8")
	pushImm32 = form("PushImm32")

	popR64  = form("PopR64")
	popRM64 = form("PopRM64")

	leave = form("Leave")
)

// PushR64 is the one-byte 50+rd form. PushRM64 is FF /6 and reaches memory;
// handed a register it is three bytes to PushR64's one, so a lowering that
// knows it has a register should say so.
func (s *Section) PushR64(src reg.R64)       { s.inst(pushR64, src) }
func (s *Section) PushRM64(src operand.RM64) { s.inst(pushRM64, src) }

// PushImm8 sign-extends its byte to the full 64-bit push. PushImm32 does the
// same with a doubleword, and there is no imm64 push: a 64-bit constant
// reaches the stack through MovR64Imm64 and PushR64.
func (s *Section) PushImm8(v int64)  { s.inst(pushImm8, imm(v)) }
func (s *Section) PushImm32(v int64) { s.inst(pushImm32, imm(v)) }

func (s *Section) PopR64(dst reg.R64)       { s.inst(popR64, dst) }
func (s *Section) PopRM64(dst operand.RM64) { s.inst(popRM64, dst) }

// Leave is the one-byte teardown: mov rsp, rbp then pop rbp. Its counterpart
// ENTER is not in the table — it is slow, nothing emits it, and a frame setup
// is two instructions a lowering already knows how to write.
func (s *Section) Leave() { s.inst(leave) }
