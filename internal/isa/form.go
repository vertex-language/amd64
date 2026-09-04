package isa

import (
	"strings"

	"github.com/vertex-language/amd64/feature"
	"github.com/vertex-language/amd64/operand"
)

// Form is one declared encoding of one instruction.
//
// Helper is the method name on *amd64.Section that pins this form. It is
// the join between the table and inst_*.go, and it is a name rather than an
// index on purpose: appending rows breaks nothing, and a removed or renamed
// form panics at program start naming the missing form rather than silently
// binding to the wrong row.
type Form struct {
	Mnemonic string
	Helper   string
	Ops      []Class
	Enc      Enc

	// Gate is the features a module must hold to emit this form. Empty
	// means the baseline, which every module this package can build has.
	Gate []feature.Feature

	// AliasOf names the form this one is a second spelling of. JeLabel and
	// JzLabel are both 0x74; SalRM64One emits SHL's bytes. They are
	// separate forms so a listing can say which name the caller used while
	// the bytes say what the silicon does.
	AliasOf string

	// Lockable reports whether LOCK is architecturally legal on this form.
	// table_lock.go clones the rows that say yes; a register destination on
	// the clone is refused at the call, because LOCK on a register is #UD
	// and that is a fact about the value rather than its class.
	Lockable bool

	index int
}

// Index is the form's position in the table. It exists for diagnostics and
// for stable tie-breaking in Emit, and nothing binds to it.
func (f *Form) Index() int { return f.index }

// String is the Intel spelling: "ADD r/m64, imm8". It is what an error
// message names, so it has to read the way the manual reads.
func (f *Form) String() string {
	var b strings.Builder
	b.WriteString(strings.ToUpper(f.Mnemonic))
	for i, c := range f.Ops {
		if i == 0 {
			b.WriteByte(' ')
		} else {
			b.WriteString(", ")
		}
		b.WriteString(c.String())
	}
	return b.String()
}

// Gated reports whether the form needs anything past the baseline.
func (f *Form) Gated() bool { return len(f.Gate) > 0 }

// Permitted reports whether the given feature set allows this form.
func (f *Form) Permitted(s feature.Set) bool {
	for _, g := range f.Gate {
		if !s.Has(g) {
			return false
		}
	}
	return true
}

// Missing returns the gates the set does not hold, for the error message.
func (f *Form) Missing(s feature.Set) []feature.Feature {
	var out []feature.Feature
	for _, g := range f.Gate {
		if !s.Has(g) {
			out = append(out, g)
		}
	}
	return out
}

// Accepts reports whether the operands fill this form's positions.
func (f *Form) Accepts(ops []operand.Operand) bool {
	if len(ops) != len(f.Ops) {
		return false
	}
	for i, c := range f.Ops {
		if !c.Accepts(ops[i]) {
			return false
		}
	}
	return true
}

// HasRel reports whether any operand is a branch displacement. Emit needs
// this: among rel forms the widest wins, because a displacement is not
// known until Finalize and "shortest" would mean "always rel8, and fail
// later if the target is far".
func (f *Form) HasRel() bool {
	for _, c := range f.Ops {
		if c == Rel8 || c == Rel32 {
			return true
		}
	}
	return false
}

// RelWidth is the displacement field's width in bytes, or 0.
func (f *Form) RelWidth() int {
	for _, c := range f.Ops {
		switch c {
		case Rel8:
			return 1
		case Rel32:
			return 4
		}
	}
	return 0
}
