package asm

import (
	"fmt"
	"strings"

	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
	"github.com/vertex-language/asm/gas"
)

// AT&T operand grammar.
//
//	%rax                    a register
//	$42                     an immediate
//	sym                     a branch target or an absolute address
//	*%rax                   an indirect branch target
//	8(%rbx)                 base plus displacement
//	(%rbx, %rcx, 4)         base, index, scale
//	sym(%rip)               RIP-relative
//	sym@GOTPCREL(%rip)      through the GOT
//	%fs:0x10                a segment override
//
// The width of a memory operand is not written in it. It comes from the
// mnemonic's suffix, or from the register on the other side, and this file's
// job is to hold the pieces until whichever one answers.

// mem is a memory operand under construction, held apart from the sized types
// the parent package builds because the size is not known until the whole
// instruction has been read.
type mem struct {
	base     reg.R64
	hasBase  bool
	index    reg.R64
	hasIndex bool
	scale    uint8
	disp     int64
	sym      operand.SymRef
	hasSym   bool
	rip      bool
	seg      reg.Sreg
	hasSeg   bool
	tok      gas.Token
}

// operandValue is one parsed operand: exactly one field is set.
type operandValue struct {
	op       operand.Operand // a register or an immediate
	mem      *mem            // a memory operand, still unsized
	indirect bool            // *%rax or *sym: an indirect branch target
	tok      gas.Token
}

// parseOperand reads one comma-separated operand.
func (t *target) parseOperand(p *gas.Parser) (operandValue, error) {
	tok := p.Peek()

	// `*` marks an indirect branch target: `jmp *%rax`, `call *(%rbx)`.
	indirect := false
	if tok.IsPunct("*") {
		p.Take()
		indirect = true
		tok = p.Peek()
	}

	switch {
	case tok.IsPunct("%"):
		// A register, or a segment override introducing an address.
		r, err := t.parseReg(p)
		if err != nil {
			return operandValue{}, err
		}
		if s, ok := r.(reg.Sreg); ok && p.Peek().IsPunct(":") {
			p.Take()
			m, err := t.parseMem(p, tok)
			if err != nil {
				return operandValue{}, err
			}
			m.seg, m.hasSeg = s, true
			return operandValue{mem: m, indirect: indirect, tok: tok}, nil
		}
		v, ok := r.(operand.Operand)
		if !ok {
			return operandValue{}, p.Errorf(tok, "%%%s is not an operand", r)
		}
		return operandValue{op: v, indirect: indirect, tok: tok}, nil

	case tok.IsPunct("$"):
		p.Take()
		at := p.Peek()
		v, err := p.Expr()
		if err != nil {
			return operandValue{}, err
		}
		if !v.IsAbs() {
			// `$sym` is the address of a symbol as an immediate, which is
			// an absolute relocation into an imm32 field. The encoder has
			// no operand for that, so it is refused by name rather than
			// silently becoming zero.
			return operandValue{}, p.Errorf(at,
				"$%s: an immediate holding a symbol's address is not supported; "+
					"load it with lea %s(%%rip)", v, v.Sym)
		}
		return operandValue{op: operand.NewImm(v.Off), indirect: indirect, tok: tok}, nil
	}

	// Everything else opens a memory operand or a bare target: `(%rax)`,
	// `8(%rax)`, `sym(%rip)`, `sym`, `-16(%rbp)`.
	m, err := t.parseMem(p, tok)
	if err != nil {
		return operandValue{}, err
	}
	return operandValue{mem: m, indirect: indirect, tok: tok}, nil
}

// parseReg reads `%name`.
func (t *target) parseReg(p *gas.Parser) (fmt.Stringer, error) {
	p.Take() // '%'
	n := p.Peek()
	if n.Kind != gas.Ident {
		return nil, p.Errorf(n, "want a register name after %%, got %s", n)
	}
	p.Take()
	name := strings.ToLower(n.Text)
	if name == "rip" {
		// RIP is not a reg.Value: no instruction takes it as an operand, it
		// is a base and nothing else. The caller checks for this sentinel.
		return ripSentinel{}, nil
	}
	r, ok := reg.Lookup(name)
	if !ok {
		return nil, p.Errorf(n, "%%%s is not a register", n.Text)
	}
	return r, nil
}

// ripSentinel stands for %rip in the one place it may appear. It satisfies
// reg.Value's printing so a diagnostic can name it, and satisfies nothing
// else, which is the parent package's position on RIP restated.
type ripSentinel struct{}

func (ripSentinel) String() string { return "rip" }

// parseMem reads a memory operand or a bare symbol.
func (t *target) parseMem(p *gas.Parser, at gas.Token) (*mem, error) {
	m := &mem{tok: at}

	// An optional displacement or symbol before the parenthesis.
	if !p.Peek().IsPunct("(") {
		e, err := p.Expr()
		if err != nil {
			return nil, err
		}
		switch {
		case e.IsDiff():
			return nil, p.Errorf(at, "%s: a label difference is not an address", e)
		case e.Sym != "":
			m.sym, m.hasSym = operand.Ref(e.Sym, 0).Add(e.Off), true
			// `sym@GOTPCREL` and friends.
			if p.Peek().IsPunct("@") {
				p.Take()
				mod := p.Peek()
				if mod.Kind != gas.Ident {
					return nil, p.Errorf(mod, "want a relocation modifier after @, got %s", mod)
				}
				p.Take()
				kind, ok := modifiers[strings.ToUpper(mod.Text)]
				if !ok {
					return nil, p.Errorf(mod, "unknown relocation modifier @%s", mod.Text)
				}
				m.sym = operand.Ref(e.Sym, kind).Add(e.Off)
			}
		default:
			m.disp = e.Off
		}
	}

	if !p.AcceptPunct("(") {
		// A bare symbol or number: a branch target, or an absolute address.
		return m, nil
	}

	// (%base), (%base, %index, scale), (, %index, scale), (%rip)
	if !p.Peek().IsPunct(",") {
		r, err := t.parseReg(p)
		if err != nil {
			return nil, err
		}
		if _, isRip := r.(ripSentinel); isRip {
			m.rip = true
		} else {
			b, ok := r.(reg.R64)
			if !ok {
				return nil, p.Errorf(at, "%s is not a 64-bit base register", r)
			}
			m.base, m.hasBase = b, true
		}
	}

	if p.AcceptPunct(",") {
		if !p.Peek().IsPunct(")") {
			r, err := t.parseReg(p)
			if err != nil {
				return nil, err
			}
			x, ok := r.(reg.R64)
			if !ok {
				return nil, p.Errorf(at, "%s is not a 64-bit index register", r)
			}
			m.index, m.hasIndex = x, true
			m.scale = 1
		}
		if p.AcceptPunct(",") {
			st := p.Peek()
			v, err := p.Expr()
			if err != nil {
				return nil, err
			}
			switch v.Off {
			case 1, 2, 4, 8:
				m.scale = uint8(v.Off)
			default:
				return nil, p.Errorf(st, "a scale is 1, 2, 4 or 8, not %s", v)
			}
		}
	}

	if err := p.ExpectPunct(")"); err != nil {
		return nil, err
	}
	return m, nil
}

// modifiers maps GNU as's @-suffixes onto relocation kinds.
//
// This is the mapping gas cannot make, and the reason it is here rather than
// there is the same one that puts :lo12: in arm64/asm: the spellings are the
// architecture's, and a shared table of both would be a table of two unrelated
// things.
var modifiers = map[string]operand.RefKind{
	"GOTPCREL": operand.RefGOTPCREL,
	"PLT":      operand.RefPLT32,
	"GOTOFF":   operand.RefGOTOFF64,
}

// sized turns the collected pieces into the parent package's sized address
// type. Seven near-identical branches, because the width is in the Go type
// there and that is what makes a width mismatch a compile error at every
// typed call site.
func (m *mem) sized(w int) (operand.Operand, error) {
	switch {
	case m.rip:
		if !m.hasSym {
			return nil, fmt.Errorf("a RIP-relative address needs a symbol")
		}
		switch w {
		case 0:
			return operand.Rip(m.sym), nil
		case 1:
			return operand.Rip8(m.sym), nil
		case 2:
			return operand.Rip16(m.sym), nil
		case 4:
			return operand.Rip32(m.sym), nil
		case 8:
			return operand.Rip64(m.sym), nil
		case 16:
			return operand.Rip128(m.sym), nil
		case 32:
			return operand.Rip256(m.sym), nil
		case 64:
			return operand.Rip512(m.sym), nil
		}
		return nil, fmt.Errorf("no %d-byte RIP-relative address", w)

	case !m.hasBase && !m.hasIndex && m.hasSym:
		switch w {
		case 1:
			return operand.Abs8(m.sym), nil
		case 2:
			return operand.Abs16(m.sym), nil
		case 4:
			return operand.Abs32(m.sym), nil
		case 8:
			return operand.Abs64(m.sym), nil
		case 16:
			return operand.Abs128(m.sym), nil
		case 32:
			return operand.Abs256(m.sym), nil
		case 64:
			return operand.Abs512(m.sym), nil
		}
		return nil, fmt.Errorf("no %d-byte absolute address", w)
	}

	if !m.hasBase && !m.hasIndex {
		// A bare number: an absolute address, which is how a segment-
		// relative operand like %fs:0 is written.
		if m.disp < 0 || m.disp > 0xffffffff {
			return nil, fmt.Errorf("absolute address %d does not fit 32 unsigned bits", m.disp)
		}
		n := uint32(m.disp)
		var a operand.Operand
		switch w {
		case 1:
			x := operand.Addr8(n)
			if m.hasSeg {
				x = x.Seg(m.seg)
			}
			a = x
		case 2:
			x := operand.Addr16(n)
			if m.hasSeg {
				x = x.Seg(m.seg)
			}
			a = x
		case 4:
			x := operand.Addr32(n)
			if m.hasSeg {
				x = x.Seg(m.seg)
			}
			a = x
		case 8:
			x := operand.Addr64(n)
			if m.hasSeg {
				x = x.Seg(m.seg)
			}
			a = x
		default:
			return nil, fmt.Errorf("no %d-byte absolute address", w)
		}
		return a, nil
	}
	if !m.hasBase {
		return nil, fmt.Errorf("an address with an index and no base is not supported")
	}
	if m.disp > 0x7fffffff || m.disp < -0x80000000 {
		return nil, fmt.Errorf("displacement %d does not fit 32 signed bits", m.disp)
	}
	d := int32(m.disp)

	if w == 0 {
		// LEA computes an address and reads no memory, so its operand is
		// the address without a width. Building it at one and dropping the
		// width is what Addr() is for.
		x := operand.Mem64(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x.Addr(), nil
	}

	switch w {
	case 1:
		x := operand.Mem8(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x, nil
	case 2:
		x := operand.Mem16(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x, nil
	case 4:
		x := operand.Mem32(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x, nil
	case 8:
		x := operand.Mem64(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x, nil
	case 16:
		x := operand.Mem128(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x, nil
	case 32:
		x := operand.Mem256(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x, nil
	case 64:
		x := operand.Mem512(m.base).Disp(d)
		if m.hasIndex {
			x = x.Index(m.index, m.scale)
		}
		if m.hasSeg {
			x = x.Seg(m.seg)
		}
		return x, nil
	}
	return nil, fmt.Errorf("no %d-byte address", w)
}
