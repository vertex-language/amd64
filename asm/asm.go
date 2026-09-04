package asm

import (
	"strings"

	"github.com/vertex-language/amd64"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
	"github.com/vertex-language/asm/gas"
)

// target is the gas.Target half.
type target struct{ e *emitter }

func (t *target) Syntax() gas.Syntax { return gas.SysV }

// Inst parses one instruction's operands and emits it.
//
// Three things happen here that do not happen on AArch64, and all three are
// AT&T's doing.
//
// The operand order is reversed. `movq %rax, %rbx` writes rax into rbx, and
// the table is in Intel order, so the list is turned around before it is
// handed over. This is the single largest difference between the two
// dialects and it is one line.
//
// The mnemonic carries a width suffix that the table's mnemonics do not, so
// it is peeled off and used to size the memory operand — which is the only
// operand that needs it, since a register already states its own width.
//
// A prefix — `lock`, `rep` — is written as a separate word before the
// mnemonic rather than as part of it.
func (t *target) Inst(p *gas.Parser, mnem string) error {
	at := p.Peek()

	name, lock := splitPrefix(strings.ToLower(mnem))
	if name == "" {
		// A bare prefix: the instruction it applies to is the next word.
		n := p.Peek()
		if n.Kind != gas.Ident {
			return p.Errorf(n, "%s must be followed by an instruction, got %s", mnem, n)
		}
		p.Take()
		name = strings.ToLower(n.Text)
		lock = true
		at = p.Peek()
	}

	var vals []operandValue
	for !p.AtEnd() {
		v, err := t.parseOperand(p)
		if err != nil {
			return err
		}
		vals = append(vals, v)
		if !p.AcceptPunct(",") {
			break
		}
	}

	if alt, ok := attOnly[name]; ok {
		name = alt
	}
	// The extending moves name two widths and are sized by neither of the
	// usual rules; see extending.
	memWidth := 0
	if x, ok := extending[name]; ok {
		name, memWidth = x.mnem, x.src
	}
	base, suffix := splitSuffix(name, vals)
	if memWidth == 0 {
		memWidth = suffix
		if memWidth == 0 {
			memWidth = widthFromRegisters(vals)
		}
	}
	ops, err := t.materialise(p, base, memWidth, vals)
	if err != nil {
		return err
	}

	// AT&T writes source first; the table is in Intel order.
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}

	before := t.e.m.Err()
	if lock {
		t.e.sec.LockEmit(base, ops...)
	} else {
		t.e.sec.Emit(base, ops...)
	}
	if after := t.e.m.Err(); after != nil && before == nil {
		return p.Errorf(at, "%s: %v", mnem, after)
	}
	return nil
}

// splitPrefix peels a `lock` written as part of the mnemonic, and reports the
// bare-prefix case with an empty name.
func splitPrefix(mnem string) (name string, lock bool) {
	switch mnem {
	case "lock":
		return "", true
	}
	if rest, ok := strings.CutPrefix(mnem, "lock"); ok && rest != "" {
		return rest, true
	}
	return mnem, false
}

var suffixWidth = map[byte]int{'b': 1, 'w': 2, 'l': 4, 'q': 8}

// attOnly are the mnemonics AT&T spells differently from the table, which is
// Intel's.
//
// They are the sign-extending accumulator instructions, whose two names
// describe the same instruction from opposite ends — AT&T's cltq says "long
// to quad" and Intel's cdqe says "doubleword to quadword extend" — and there
// is no rule connecting them, only a list.
var attOnly = map[string]string{
	"cbtw": "cbw", "cwtl": "cwde", "cltq": "cdqe",
	"cwtd": "cwd", "cltd": "cdq", "cqto": "cqo",
}

// extending is AT&T's sign- and zero-extending moves, which name both widths
// in the mnemonic: movzbl is byte-to-long, movswq is word-to-quad.
//
// They need their own table because neither sizing rule reaches them. The
// trailing letter is the destination's width and the register operand already
// states it, so treating it as a suffix would size the *source* — the memory
// operand — as four or eight bytes and resolve to a form that reads too much.
// What the memory operand needs is the width the first letter names, and that
// is what src carries.
var extending = map[string]struct {
	mnem string
	src  int
}{
	"movzbw": {"movzx", 1}, "movzbl": {"movzx", 1}, "movzbq": {"movzx", 1},
	"movzwl": {"movzx", 2}, "movzwq": {"movzx", 2},
	"movsbw": {"movsx", 1}, "movsbl": {"movsx", 1}, "movsbq": {"movsx", 1},
	"movswl": {"movsx", 2}, "movswq": {"movsx", 2},
	"movslq": {"movsxd", 4},
}

// splitSuffix peels the width letter off a mnemonic that takes one.
//
// The table decides, in both directions, and it has to. A name the table
// declares is a mnemonic and not a suffixed one: CMOVL is "move if less" and
// peeling its `l` would make it a CMOV that does not exist. A name it does
// not declare whose base it does is that base with a suffix, which is what
// makes MULL work without MULL being listed anywhere. The list this replaces
// could not have got the second case right: bswapq, popcntq, tzcntq and every
// cmov with a suffix were missing from it, and each absence was a mnemonic
// that would not assemble.
func splitSuffix(name string, vals []operandValue) (string, int) {
	if len(name) < 2 {
		return name, 0
	}
	w, ok := suffixWidth[name[len(name)-1]]
	if !ok {
		return name, 0
	}
	base := name[:len(name)-1]
	switch {
	case !amd64.HasMnemonic(base):
		// Nothing to peel down to.
		return name, 0

	case !amd64.HasMnemonic(name):
		// The base is a mnemonic and this is not, so this is that one with
		// a width letter: movl, bswapq, cmovnel.
		return base, w

	default:
		// Both are mnemonics, and one spelling means two instructions. It
		// happens once — MOVQ is the vector move and MOV with a `q` is the
		// integer one — and it can only happen this way round, because the
		// collision needs a vector mnemonic whose name ends in a width
		// letter. So the register file decides, which is what GNU as does
		// with the same line.
		if hasVectorOperand(vals) {
			return name, 0
		}
		return base, w
	}
}

// hasVectorOperand reports whether any operand names a register outside the
// general-purpose file.
func hasVectorOperand(vals []operandValue) bool {
	for _, v := range vals {
		if v.mem != nil || v.op == nil {
			continue
		}
		switch v.op.(type) {
		case reg.Xmm, reg.Ymm, reg.Zmm, reg.Mm, reg.K:
			return true
		}
	}
	return false
}

// materialise turns the parsed operands into the encoder's, sizing every
// memory operand to w.
//
// The caller decides w: the suffix when there is one, the registers
// otherwise, and the source width for an extending move, which states two.
// A memory operand with none of the three is refused rather than guessed,
// which is what GNU as does and for the same reason: `mov $1, (%rax)` writes
// one, two, four or eight bytes and the source says which only through the
// suffix nobody wrote.
func (t *target) materialise(p *gas.Parser, base string, w int,
	vals []operandValue) ([]operand.Operand, error) {

	// Neither the suffix nor a register said how wide the memory operand
	// is, so ask the table: some instructions have only one answer and no
	// way to write it. LEA's operand has no access width, CLFLUSH's is a
	// line, CMPXCHG16B's is sixteen bytes. A mnemonic whose forms disagree
	// — MOV and most of the rest — leaves w at zero and is refused below.
	fromTable := false
	if w == 0 {
		if n, ok := amd64.MemBytes(base); ok {
			w, fromTable = n, true
		}
	}

	ops := make([]operand.Operand, 0, len(vals))
	for _, v := range vals {
		if v.mem == nil {
			ops = append(ops, v.op)
			continue
		}
		// A bare symbol with no parentheses in a branch is a label, not an
		// address.
		if isBranch(base) && v.mem.hasSym && !v.mem.hasBase && !v.mem.hasIndex && !v.mem.rip {
			name := v.mem.sym.Sym
			t.e.refer(name)
			// A target this source defines is a distance; one it does not
			// is a symbol a linker has to find, and on x86-64 a call to
			// one goes through the PLT. The pre-scan is what lets the
			// parser tell them apart at the point of use rather than
			// guessing and repairing later.
			if p.Defines(name) {
				ops = append(ops, operand.NewLabel(name))
			} else {
				ops = append(ops, operand.Ref(name, operand.RefPLT32))
			}
			continue
		}
		if v.mem.hasSym {
			t.e.refer(v.mem.sym.Sym)
		}
		if w == 0 && !fromTable {
			return nil, p.Errorf(v.mem.tok,
				"%s: the width of this memory operand is not stated by any operand; "+
					"write a suffix, as in %sq", base, base)
		}
		m, err := v.mem.sized(w)
		if err != nil {
			return nil, p.Errorf(v.mem.tok, "%s: %v", base, err)
		}
		ops = append(ops, m)
	}
	return ops, nil
}

// widthFromRegisters is the width the register operands agree on, or zero.
func widthFromRegisters(vals []operandValue) int {
	for _, v := range vals {
		if v.mem != nil || v.op == nil {
			continue
		}
		if r, ok := v.op.(reg.Value); ok {
			if bits := regBits(r); bits > 0 {
				return bits / 8
			}
		}
	}
	return 0
}

func regBits(r reg.Value) int {
	switch r.(type) {
	case reg.R8:
		return 8
	case reg.R16:
		return 16
	case reg.R32:
		return 32
	case reg.R64:
		return 64
	case reg.Xmm:
		return 128
	case reg.Ymm:
		return 256
	case reg.Zmm:
		return 512
	}
	return 0
}

func isBranch(mnem string) bool {
	switch mnem {
	case "call", "jmp":
		return true
	}
	return strings.HasPrefix(mnem, "j")
}

// Options configure an assembly.
type Options struct {
	File        string
	Features    amd64.FeatureSet
	LabelPrefix string
}

// Assemble assembles src into a finished object.
func Assemble(src string, opts Options) (*obj.Object, error) {
	m, err := AssembleInto(nil, src, opts)
	if err != nil {
		return nil, err
	}
	return m.Finalize()
}

// AssembleInto assembles src into m, creating a module when m is nil.
func AssembleInto(m *amd64.Module, src string, opts Options) (*amd64.Module, error) {
	if m == nil {
		if opts.Features != (amd64.FeatureSet{}) {
			m = amd64.NewModule(amd64.WithFeatures(opts.Features))
		} else {
			m = amd64.NewModule()
		}
	}
	e := newEmitter(m)
	if err := gas.Parse(src, &target{e: e}, e, gas.Options{
		File: opts.File, LocalPrefix: opts.LabelPrefix,
	}); err != nil {
		return m, err
	}
	e.declareExterns()
	return m, m.Err()
}

// AssembleFragment assembles src into a section that is already open, at the
// offset it has already reached. See arm64/asm's for what it is for.
func AssembleFragment(sec *amd64.Section, src string, opts Options) error {
	m := sec.Module()
	e := newEmitter(m)
	e.sec = sec
	err := gas.Parse(src, &target{e: e}, e, gas.Options{
		File:        opts.File,
		LocalPrefix: opts.LabelPrefix,
		Section:     sec.Name(),
		Kind:        gasKind(sec.Kind()),
	})
	if err != nil {
		return err
	}
	e.declareExterns()
	return m.Err()
}

func gasKind(k amd64.SectionKind) gas.SectionKind {
	switch k {
	case amd64.Text:
		return gas.Text
	case amd64.ROData:
		return gas.ROData
	case amd64.BSS:
		return gas.BSS
	}
	return gas.Data
}
