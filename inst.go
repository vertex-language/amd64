package amd64

// inst.go is the join between the ISA table and the typed helpers. The
// helpers themselves live one tranche per file — inst_alu.go opposite
// isa/table_base.go's buildALU, inst_avx.go opposite table_avx.go's
// buildAVX — and every one of them is three lines:
//
//	var addR64RM64 = form("AddR64RM64")
//
//	// AddR64RM64 emits ADD r64, r/m64.
//	func (s *Section) AddR64RM64(dst reg.R64, src operand.RM64) {
//		s.inst(addR64RM64, dst, src)
//	}
//
// The binding is by name, not by table index. Appending rows breaks
// nothing; a removed or renamed row panics at program start naming the
// missing form, rather than silently binding to the wrong row or failing
// halfway through someone's code generation. Two helpers binding one name
// is a duplicate Go identifier and so a compile error, which is the
// earliest any of these can fail — and isa.index() checks the same thing
// from the table's side, because the table is what makes the promise.

import (
	"github.com/vertex-language/amd64/internal/isa"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// form binds a helper to its row.
//
// It panics rather than returning an error because a missing row is a
// question about this tree's own data, not about anything a caller did. It
// should fail whoever the caller is and whatever they were about to do,
// which is what package initialisation means.
func form(helper string) *isa.Form {
	f := isa.ByHelper(helper)
	if f == nil {
		panic("amd64: the ISA table declares no form for helper " + helper +
			"; the row was removed or renamed")
	}
	return f
}

// ---- operand shorthands ---------------------------------------------------
//
// A form's fixed operands are in the helper's name and not its parameters,
// but the encoder is still handed one operand per declared class: plan
// gives a fixed position roleNone and emits nothing for it, and gather
// still indexes it. So the helper supplies the value the form already
// named, from here.

// imm wraps a Go integer for a helper whose form pins an immediate field.
// The caller does not write this — MovRM64Imm32(RAX, 60) takes an int64 and
// the helper wraps it — because naming the form is what states the field,
// and a value that does not fit is ErrRange from the encoder with the
// width and range in the notes.
func imm(v int64) operand.Imm { return operand.NewImm(v) }

// immU is the same for a caller holding a bit pattern rather than a
// number, which is what MovR64Imm64 usually gets.
func immU(v uint64) operand.Imm { return operand.NewImmU(v) }

// label wraps a same-section branch target. JmpLabel takes a string and
// this makes it an operand; JmpRef takes a SymRef and needs no wrapper,
// because a reference that leaves the section is already one.
func label(name string) operand.Label { return operand.NewLabel(name) }

// The fixed operands, one value per Fix class in isa. They are values
// rather than calls because a fixed operand is a constant of the form.
var (
	fixAL  = reg.AL
	fixCL  = reg.CL
	fixAX  = reg.AX
	fixEAX = reg.EAX
	fixRAX = reg.RAX

	// fixOne is the literal 1 of SHL r/m64, 1. D1 /4 has no immediate
	// field to put another number in, which is why the helper is
	// ShlRM64One and not ShlRM64Imm8.
	fixOne = operand.NewImm(1)
)

// ---- the locking clones ---------------------------------------------------

// lockInst is what every Lock* helper funnels into.
//
// LOCK on a register destination is #UD. That is a fact about the value
// rather than its class — the class is RM64 either way and the operand is
// perfectly well formed — so it is ErrOperand at the call rather than
// ErrForm, and it is checked here rather than in the encoder, which sees a
// locking clone and its base row as the same shape.
func (s *Section) lockInst(f *isa.Form, ops ...operand.Operand) {
	if !s.ok() {
		return
	}
	if len(ops) > 0 {
		if _, isMem := ops[0].(operand.Memory); !isMem {
			s.m.fail(s.errorAt(obj.ErrOperand,
				"LOCK needs a memory destination",
				describe(ops[0])+" is a register, and LOCK on a register destination is #UD",
				"the unlocked form is "+unlockedName(f)))
			return
		}
	}
	s.inst(f, ops...)
}

// unlockedName is the Intel spelling of a locking clone's base form —
// the clone's own, less the prefix. A note telling a caller that the
// unlocked form is "LOCK ADD r/m64, r64" would be telling them nothing.
func unlockedName(f *isa.Form) string {
	name := f.String()
	if len(name) > 5 && name[:5] == "LOCK " {
		return name[5:]
	}
	return name
}
