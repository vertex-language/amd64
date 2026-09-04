package amd64

import (
	"strings"

	"github.com/vertex-language/amd64/internal/encode"
	"github.com/vertex-language/amd64/internal/isa"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
)

// Emit resolves a form at run time from a mnemonic and operands.
//
// It is the escape hatch for table-driven emission, where the mnemonic is
// data. If you know the instruction at compile time, the typed helper is
// the surface: it pins its form, and a width or class mismatch is a
// compile error rather than an ErrForm at run time.
//
// Two selection rules, and the second is a correction to the obvious one:
//
//   - Among forms whose length is decidable now, shortest wins, ties
//     broken by table order. Length is computed by encoding every
//     candidate, so there is no size estimator to disagree with the
//     encoder.
//   - Among rel forms, widest wins. A displacement is not known until
//     Finalize, so "shortest" would mean "always rel8, and fail later if
//     the target is far" — correct for immediates and a landmine for
//     branches. A caller who wants the two-byte encoding is making a claim
//     about distance, and claims belong in the typed surface:
//     JmpShortLabel.
//
// Emit picks an encoding of the instruction you named and never a
// different instruction. Emit("mov", RAX, NewImm(1)) is seven bytes, not
// the five of mov r32, imm32, because that is a different form with a
// different destination class. Getting those five bytes means knowing that
// a 32-bit write zeroes the upper half, which is a claim about the
// architecture, so it goes where you state it: MovR32Imm32(EAX, 1).
func (s *Section) Emit(mnemonic string, ops ...operand.Operand) {
	if !s.ok() {
		return
	}
	f := s.resolve(mnemonic, ops)
	if f == nil {
		return // resolve recorded the failure
	}
	s.inst(f, ops...)
}

// LockEmit is Emit for the LOCK-prefixed form of an instruction.
//
// It is separate from Emit rather than a mnemonic Emit accepts, because LOCK
// is a prefix and "lock add" is not an instruction name the way "add" is —
// the same position the typed surface takes by giving the locking forms their
// own helpers. What makes it necessary anyway is that assembly source spells
// it as a word: `lock cmpxchgq %rax, (%rbx)` is how every atomic in a C
// header is written, and an assembler cannot send its user to a Go method.
func (s *Section) LockEmit(mnemonic string, ops ...operand.Operand) {
	if !s.ok() {
		return
	}
	f := s.resolveLock(mnemonic, ops)
	if f == nil {
		return
	}
	s.inst(f, ops...)
}

// resolveLock is resolve for the locking clone of a mnemonic.
func (s *Section) resolveLock(mnemonic string, ops []operand.Operand) *isa.Form {
	name := strings.ToLower(strings.TrimSpace(mnemonic))
	name = strings.TrimSpace(strings.TrimPrefix(name, "lock"))
	if name == "" {
		s.m.fail(s.errorAt(obj.ErrForm, "LOCK with no instruction after it"))
		return nil
	}
	if !isa.Known("lock " + name) {
		if isa.Known(name) {
			s.m.fail(s.errorAt(obj.ErrForm,
				"LOCK is not architecturally legal on "+name,
				"the prefix is only defined on the read-modify-write forms"))
			return nil
		}
		s.m.fail(s.errorAt(obj.ErrForm,
			"no instruction named "+name,
			"the mnemonic is not declared in this table"))
		return nil
	}
	return s.pick("lock "+name, ops)
}

// resolve picks the form, or records why it could not and returns nil.
//
// The winning candidate is encoded twice: once here to measure it, once in
// inst to emit it. That is deliberate. The alternative is a size estimator
// or a cached buffer, and a second answer to "how long is this" is a
// second thing that can be wrong — wrong in exactly the cases that matter.
// Emit is the run-time path; the typed helpers, which are the hot one, do
// not come through here at all.
func (s *Section) resolve(mnemonic string, ops []operand.Operand) *isa.Form {
	name := strings.ToLower(strings.TrimSpace(mnemonic))

	// LOCK is not part of a mnemonic. Naming one is what the typed surface
	// is for, and isa.Resolve already skips the locking clones — this is
	// here so the diagnostic says which surface to use rather than "no
	// instruction named lock add".
	if name == "lock" || strings.HasPrefix(name, "lock ") {
		s.m.fail(s.errorAt(obj.ErrForm,
			"Emit does not take a LOCK prefix",
			"LOCK is not part of a mnemonic; the typed helpers spell it, as in "+
				"LockAddRM64R64, and LockEmit takes it as data"))
		return nil
	}
	return s.pick(name, ops)
}

// pick is the form choice, over a mnemonic that has already been normalised
// and checked for the LOCK prefix. Emit and LockEmit differ in how they get
// here and in nothing after it.
func (s *Section) pick(name string, ops []operand.Operand) *isa.Form {
	cands := isa.Resolve(name, ops)
	if len(cands) == 0 {
		// Two different failures that send a caller to two different
		// places: the table has no such mnemonic, or it has the mnemonic
		// and no form of it takes these operands.
		if !isa.Known(name) {
			s.m.fail(s.errorAt(obj.ErrForm,
				"no instruction named "+name,
				"the mnemonic is not declared in this table"))
			return nil
		}
		s.m.fail(s.errorAt(obj.ErrForm,
			"no form of "+name+" takes "+describeOps(ops),
			declaredForms(name)...))
		return nil
	}

	permitted := make([]*isa.Form, 0, len(cands))
	for _, f := range cands {
		if f.Permitted(s.m.features) {
			permitted = append(permitted, f)
		}
	}
	if len(permitted) == 0 {
		return s.failGated(name, ops, cands)
	}

	// The rel rule. A mnemonic can have both rel and non-rel forms — jmp
	// has rel8, rel32 and r/m64 — but the operands have already decided
	// which family this is, because a Label satisfies no r/m class and a
	// register satisfies no rel class.
	var widest *isa.Form
	for _, f := range permitted {
		if !f.HasRel() {
			continue
		}
		if widest == nil || f.RelWidth() > widest.RelWidth() {
			widest = f
		}
	}
	if widest != nil {
		return widest
	}

	// Shortest wins, ties by table order. Resolve returns table order and
	// the comparison is strict, so the first of equal length keeps it.
	var (
		best     *isa.Form
		bestLen  int
		firstErr error
	)
	for _, f := range permitted {
		n, err := encode.Length(f, ops)
		if err != nil {
			// A candidate whose classes accepted the operands and whose
			// encoder refused them: a symbol reference on a based address,
			// AH beside a REX forcer, a bad scale. Keep the first reason
			// in case every candidate fails.
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if best == nil || n < bestLen {
			best, bestLen = f, n
		}
	}
	if best == nil {
		if firstErr == nil {
			// Unreachable: permitted is non-empty, so every iteration set
			// best or firstErr.
			firstErr = obj.ErrForm
		}
		s.m.fail(s.lift(firstErr))
		return nil
	}
	return best
}

// failGated records the ErrFeature for operands that matched only forms
// this module may not emit.
//
// The gates named are the nearest candidate's — the one needing the fewest
// features — because listing the union across every candidate would tell a
// caller to enable AVX-512 to emit an SSE instruction.
func (s *Section) failGated(name string, ops []operand.Operand, cands []*isa.Form) *isa.Form {
	nearest := cands[0]
	fewest := len(nearest.Missing(s.m.features))
	for _, f := range cands[1:] {
		if n := len(f.Missing(s.m.features)); n < fewest {
			nearest, fewest = f, n
		}
	}

	missing := nearest.Missing(s.m.features)
	notes := make([]string, 0, len(missing)+1)
	notes = append(notes, "the module's feature set is "+s.m.features.String())
	for _, g := range missing {
		notes = append(notes, "add "+g.String()+" to emit "+nearest.String())
	}

	s.m.fail(s.errorAt(obj.ErrFeature,
		name+" "+describeOps(ops)+" matches only forms this feature set does not permit",
		notes...))
	return nil
}

// declaredForms lists what the mnemonic does take, for the ErrForm notes.
// It is capped because a caller who wrote the wrong operands for cmov does
// not need thirty rows read back at them.
func declaredForms(name string) []string {
	all := isa.ByMnemonic(name)
	const cap = 6

	n := len(all)
	if n > cap {
		n = cap
	}
	notes := make([]string, 0, n+1)
	for _, f := range all[:n] {
		notes = append(notes, "declared: "+f.String())
	}
	if len(all) > cap {
		notes = append(notes, "and "+decimal(int64(len(all)-cap))+" more forms of "+name)
	}
	return notes
}

// describeOps renders an operand list the way a diagnostic should read.
func describeOps(ops []operand.Operand) string {
	if len(ops) == 0 {
		return "no operands"
	}
	var b strings.Builder
	for i, op := range ops {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(describe(op))
	}
	return b.String()
}

// describe names one operand. Imm is first because it deliberately has no
// String method: an immediate carries a value and no width, and printing
// one is a diagnostic's job rather than the type's.
func describe(op operand.Operand) string {
	switch v := op.(type) {
	case operand.Imm:
		return decimal(v.Int64())
	case interface{ String() string }:
		return v.String()
	}
	return "?"
}
