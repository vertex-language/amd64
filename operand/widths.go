package operand

import "github.com/vertex-language/amd64/reg"

// Width is a memory operand's access width — how many bytes the instruction
// touches, not how wide its address or displacement is.
type Width uint16

const (
	// WNone is an address with no access width: LEA's operand, and
	// INVLPG's. An instruction taking one is computing or naming an
	// address, not reading through it.
	WNone Width = 0

	W8   Width = 8
	W16  Width = 16
	W32  Width = 32
	W64  Width = 64
	W128 Width = 128
	W256 Width = 256
	W512 Width = 512
)

func (w Width) Bits() int  { return int(w) }
func (w Width) Bytes() int { return int(w) / 8 }

func (w Width) String() string {
	switch w {
	case WNone:
		return "m"
	case W8:
		return "m8"
	case W16:
		return "m16"
	case W32:
		return "m32"
	case W64:
		return "m64"
	case W128:
		return "m128"
	case W256:
		return "m256"
	case W512:
		return "m512"
	}
	return "m?"
}

// Memory is any address, of any width or none. LeaR64M and InvlpgM take it
// rather than an RM class, because their operand is an address and a
// register in that slot would not be an instruction.
type Memory interface {
	Operand
	Width() Width
	Addr() Addr
	Err() error

	memory()
}

// The register-or-memory classes. Each is satisfied by exactly one register
// width and one memory width; the markers on the register side are declared
// in reg.
type (
	RM8 interface {
		Operand
		RM8()
	}
	RM16 interface {
		Operand
		RM16()
	}
	RM32 interface {
		Operand
		RM32()
	}
	RM64 interface {
		Operand
		RM64()
	}
	RM128 interface {
		Operand
		RM128()
	}
	RM256 interface {
		Operand
		RM256()
	}
	RM512 interface {
		Operand
		RM512()
	}

	// Scalar SSE reads and writes exactly its own operand's width out of a
	// vector register, so its r/m class is an xmm register or four or
	// eight bytes of memory — the memory half of RM32 and RM64, and the
	// register half of neither.
	XM32 interface {
		Operand
		XM32()
	}
	XM64 interface {
		Operand
		XM64()
	}
)

// Addr is a memory operand with no access width. It is what Rip returns for
// LEA, and what every width reports through Addr() when the encoder needs
// the address and not the width.
type Addr struct {
	reg.Seal
	a addr
}

func (m Addr) memory()        {}
func (m Addr) Width() Width   { return WNone }
func (m Addr) Addr() Addr     { return m }
func (m Addr) Err() error     { return m.a.err }
func (m Addr) String() string { return m.a.text(WNone) }

// The seven access widths. Each is a distinct type so that a width mismatch
// is a compile error, and each chain method returns its own type so a
// refined address still satisfies the parameter it was written into.
type (
	M8 struct {
		reg.Seal
		a addr
	}
	M16 struct {
		reg.Seal
		a addr
	}
	M32 struct {
		reg.Seal
		a addr
	}
	M64 struct {
		reg.Seal
		a addr
	}
	M128 struct {
		reg.Seal
		a addr
	}
	M256 struct {
		reg.Seal
		a addr
	}
	M512 struct {
		reg.Seal
		a addr
	}
)

func (m M8) memory()        {}
func (m M8) RM8()           {}
func (m M8) Width() Width   { return W8 }
func (m M8) Addr() Addr     { return Addr{a: m.a} }
func (m M8) Err() error     { return m.a.err }
func (m M8) String() string { return m.a.text(W8) }

func (m M8) Disp(d int32) M8                 { m.a = m.a.withDisp(d); return m }
func (m M8) Index(x reg.R64, scale uint8) M8 { m.a = m.a.withIndex(x, scale); return m }
func (m M8) Seg(s reg.Sreg) M8               { m.a = m.a.segment(s); return m }
func (m M8) Sym(r SymRef) M8                 { m.a = m.a.symbol(r); return m }

func (m M16) memory()        {}
func (m M16) RM16()          {}
func (m M16) Width() Width   { return W16 }
func (m M16) Addr() Addr     { return Addr{a: m.a} }
func (m M16) Err() error     { return m.a.err }
func (m M16) String() string { return m.a.text(W16) }

func (m M16) Disp(d int32) M16                 { m.a = m.a.withDisp(d); return m }
func (m M16) Index(x reg.R64, scale uint8) M16 { m.a = m.a.withIndex(x, scale); return m }
func (m M16) Seg(s reg.Sreg) M16               { m.a = m.a.segment(s); return m }
func (m M16) Sym(r SymRef) M16                 { m.a = m.a.symbol(r); return m }

func (m M32) memory()        {}
func (m M32) RM32()          {}
func (m M32) Width() Width   { return W32 }
func (m M32) Addr() Addr     { return Addr{a: m.a} }
func (m M32) Err() error     { return m.a.err }
func (m M32) String() string { return m.a.text(W32) }

func (m M32) Disp(d int32) M32 { m.a = m.a.withDisp(d); return m }
func (m M32) XM32()            {}

func (m M32) Index(x reg.R64, scale uint8) M32 { m.a = m.a.withIndex(x, scale); return m }
func (m M32) Seg(s reg.Sreg) M32               { m.a = m.a.segment(s); return m }
func (m M32) Sym(r SymRef) M32                 { m.a = m.a.symbol(r); return m }

func (m M64) memory()        {}
func (m M64) RM64()          {}
func (m M64) Width() Width   { return W64 }
func (m M64) Addr() Addr     { return Addr{a: m.a} }
func (m M64) Err() error     { return m.a.err }
func (m M64) String() string { return m.a.text(W64) }

func (m M64) XM64() {}

func (m M64) Disp(d int32) M64                 { m.a = m.a.withDisp(d); return m }
func (m M64) Index(x reg.R64, scale uint8) M64 { m.a = m.a.withIndex(x, scale); return m }
func (m M64) Seg(s reg.Sreg) M64               { m.a = m.a.segment(s); return m }
func (m M64) Sym(r SymRef) M64                 { m.a = m.a.symbol(r); return m }

func (m M128) memory()        {}
func (m M128) RM128()         {}
func (m M128) Width() Width   { return W128 }
func (m M128) Addr() Addr     { return Addr{a: m.a} }
func (m M128) Err() error     { return m.a.err }
func (m M128) String() string { return m.a.text(W128) }

func (m M128) Disp(d int32) M128                 { m.a = m.a.withDisp(d); return m }
func (m M128) Index(x reg.R64, scale uint8) M128 { m.a = m.a.withIndex(x, scale); return m }
func (m M128) Seg(s reg.Sreg) M128               { m.a = m.a.segment(s); return m }
func (m M128) Sym(r SymRef) M128                 { m.a = m.a.symbol(r); return m }
func (m M128) Bcst() M128                        { m.a = m.a.broadcast(); return m }

func (m M256) memory()        {}
func (m M256) RM256()         {}
func (m M256) Width() Width   { return W256 }
func (m M256) Addr() Addr     { return Addr{a: m.a} }
func (m M256) Err() error     { return m.a.err }
func (m M256) String() string { return m.a.text(W256) }

func (m M256) Disp(d int32) M256                 { m.a = m.a.withDisp(d); return m }
func (m M256) Index(x reg.R64, scale uint8) M256 { m.a = m.a.withIndex(x, scale); return m }
func (m M256) Seg(s reg.Sreg) M256               { m.a = m.a.segment(s); return m }
func (m M256) Sym(r SymRef) M256                 { m.a = m.a.symbol(r); return m }
func (m M256) Bcst() M256                        { m.a = m.a.broadcast(); return m }

func (m M512) memory()        {}
func (m M512) RM512()         {}
func (m M512) Width() Width   { return W512 }
func (m M512) Addr() Addr     { return Addr{a: m.a} }
func (m M512) Err() error     { return m.a.err }
func (m M512) String() string { return m.a.text(W512) }

func (m M512) Disp(d int32) M512                 { m.a = m.a.withDisp(d); return m }
func (m M512) Index(x reg.R64, scale uint8) M512 { m.a = m.a.withIndex(x, scale); return m }
func (m M512) Seg(s reg.Sreg) M512               { m.a = m.a.segment(s); return m }
func (m M512) Sym(r SymRef) M512                 { m.a = m.a.symbol(r); return m }
func (m M512) Bcst() M512                        { m.a = m.a.broadcast(); return m }
