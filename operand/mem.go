package operand

import "github.com/vertex-language/amd64/reg"

// addr is the one address representation. Every width wraps it; none of
// them adds a field.
//
// Three encoding facts are deliberately not here. That RSP or R12 as a base
// needs a SIB byte, that RBP or R13 as a base needs an explicit
// displacement even when it is zero, and that a plain numeric address needs
// the SIB no-base form because ModRM's disp32 encoding now means
// RIP-relative — all three are decided at encode time from these fields.
// They are the encoder's business, not the caller's and not this struct's.
type addr struct {
	base     reg.R64
	hasBase  bool
	index    reg.R64
	hasIndex bool
	scale    uint8

	disp int32

	rip bool // RIP-relative; base and index are meaningless
	abs bool // numeric absolute; forces the SIB no-base form

	ref    SymRef
	hasRef bool

	seg    reg.Sreg
	hasSeg bool

	bcst bool

	err error // sticky, first wins
}

// setErr records the first failure and keeps it. Later chain methods still
// run and still return an address; they simply cannot clear this.
func (a addr) setErr(err error) addr {
	if a.err == nil {
		a.err = err
	}
	return a
}

// The refinements are named withX rather than X because a struct cannot
// carry a field and a method under one name, and the fields are the
// address — they are what parts() hands the encoder.

func (a addr) withDisp(d int32) addr {
	// A displacement on a symbolic address folds into the symbol's addend,
	// so [rbx+8].Sym(x) and [rbx].Sym(x).Disp(8) are the same operand and
	// there is one place for the number whichever order the chain was
	// written in.
	if a.hasRef {
		a.ref.Addend += int64(d)
		return a
	}
	a.disp += d
	return a
}

func (a addr) withIndex(x reg.R64, scale uint8) addr {
	switch {
	case a.rip:
		return a.setErr(fail(
			"an index cannot be added to a RIP-relative address",
			"the RIP-relative form is [rip + disp32] and has no SIB byte",
		))
	case a.hasIndex:
		return a.setErr(fail(
			"the address already has an index register: " + a.index.String(),
		))
	case x.NoIndex():
		return a.setErr(fail(
			"rsp cannot be a SIB index register",
			"index 4 with REX.X clear is the encoding for no index at all",
			"r12 has the same low three bits and is a valid index, because REX.X is set",
		))
	case scale != 1 && scale != 2 && scale != 4 && scale != 8:
		return a.setErr(fail(
			"scale must be 1, 2, 4 or 8",
			"the SIB scale field is two bits and encodes those four values",
		))
	}
	a.index, a.hasIndex, a.scale = x, true, scale
	return a
}

func (a addr) segment(s reg.Sreg) addr {
	if !s.Valid() {
		return a.setErr(fail("not a segment register"))
	}
	if a.hasSeg {
		return a.setErr(fail(
			"the address already has a segment override: " + a.seg.String(),
		))
	}
	// A segment override is legal on every addressing form including
	// RIP-relative: the base applies after the effective address has been
	// computed, whatever computed it.
	a.seg, a.hasSeg = s, true
	return a
}

func (a addr) symbol(r SymRef) addr {
	if a.hasRef {
		return a.setErr(fail(
			"the address already names a symbol: " + a.ref.Sym,
		))
	}
	if r.Sym == "" {
		return a.setErr(fail("a symbol reference needs a name"))
	}
	// Any displacement already applied becomes the symbol's addend, so the
	// order the chain was written in does not matter.
	r.Addend += int64(a.disp)
	a.disp = 0
	a.ref, a.hasRef = r, true
	return a
}

func (a addr) broadcast() addr {
	if !a.hasBase && !a.hasIndex && !a.rip && !a.hasRef {
		return a.setErr(fail("broadcast needs a memory operand to broadcast from"))
	}
	a.bcst = true
	return a
}

// text renders the address the way a diagnostic should print it. It is not
// an assembler syntax and nothing parses it back.
func (a addr) text(w Width) string {
	s := w.String() + " ["
	wrote := false
	switch {
	case a.rip:
		s += "rip"
		wrote = true
	case a.hasBase:
		s += a.base.String()
		wrote = true
	}
	if a.hasIndex {
		if wrote {
			s += "+"
		}
		s += a.index.String() + "*" + scaleText(a.scale)
		wrote = true
	}
	if a.hasRef {
		if wrote {
			s += "+"
		}
		s += a.ref.Sym
	}
	return s + "]"
}

// scaleText prints a scale. A zero here means the index was refused before
// the scale was recorded, and the address is already carrying that error;
// printing a digit that was never valid would name a scale nobody wrote.
func scaleText(s uint8) string {
	switch s {
	case 1:
		return "1"
	case 2:
		return "2"
	case 4:
		return "4"
	case 8:
		return "8"
	}
	return "?"
}

// ---- constructors ---------------------------------------------------------
//
// Four kinds of address, one question each, and the number in the name is
// always the access width:
//
//	MemN(base)  based        [rbx]
//	RipN(ref)   RIP-relative [rip + msg]     — see rip.go
//	AbsN(ref)   symbolic     [msg]
//	AddrN(n)    direct       [0xb8000]
//
// Rip is the one to reach for. On this architecture lea rdi, [rip + msg] is
// four bytes, position-independent, and needs no thunk, no GOT and no
// decision, so PIC is the default because the addressing mode is.

func based(b reg.R64) addr {
	if !b.Valid() {
		return addr{err: fail("not a 64-bit general-purpose register")}
	}
	return addr{base: b, hasBase: true}
}

func Mem8(b reg.R64) M8     { return M8{a: based(b)} }
func Mem16(b reg.R64) M16   { return M16{a: based(b)} }
func Mem32(b reg.R64) M32   { return M32{a: based(b)} }
func Mem64(b reg.R64) M64   { return M64{a: based(b)} }
func Mem128(b reg.R64) M128 { return M128{a: based(b)} }
func Mem256(b reg.R64) M256 { return M256{a: based(b)} }
func Mem512(b reg.R64) M512 { return M512{a: based(b)} }

// symbolic is an absolute address named by a symbol: no base, no index, a
// displacement-sized hole and a relocation. The hole is four bytes, which
// is what a ModRM displacement field is; a kind wider than that is ErrRange
// from the encoder naming the field.
func symbolic(r SymRef) addr {
	a := addr{abs: true}
	return a.symbol(r)
}

func Abs8(r SymRef) M8     { return M8{a: symbolic(r)} }
func Abs16(r SymRef) M16   { return M16{a: symbolic(r)} }
func Abs32(r SymRef) M32   { return M32{a: symbolic(r)} }
func Abs64(r SymRef) M64   { return M64{a: symbolic(r)} }
func Abs128(r SymRef) M128 { return M128{a: symbolic(r)} }
func Abs256(r SymRef) M256 { return M256{a: symbolic(r)} }
func Abs512(r SymRef) M512 { return M512{a: symbolic(r)} }

// direct is a numeric absolute address with no relocation. It costs a byte
// more than its i386 twin, because the encoding that meant disp32 there
// means RIP-relative here and the address has to go through the SIB no-base
// form instead. That is the encoder's business, not yours.
func direct(n uint32) addr {
	return addr{abs: true, disp: int32(n)}
}

func Addr8(n uint32) M8     { return M8{a: direct(n)} }
func Addr16(n uint32) M16   { return M16{a: direct(n)} }
func Addr32(n uint32) M32   { return M32{a: direct(n)} }
func Addr64(n uint32) M64   { return M64{a: direct(n)} }
func Addr128(n uint32) M128 { return M128{a: direct(n)} }
func Addr256(n uint32) M256 { return M256{a: direct(n)} }
func Addr512(n uint32) M512 { return M512{a: direct(n)} }
