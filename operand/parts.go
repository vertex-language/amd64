package operand

import "github.com/vertex-language/amd64/reg"

// Parts is a read-only view of an address, for the encoder.
//
// It exists because internal/encode is a different package and cannot reach
// addr's fields. Exposing them as a value rather than as a dozen accessors
// keeps the surface one name wide, and nothing here is settable: a Parts
// handed back does not become an address again.
type Parts struct {
	Base    reg.R64
	HasBase bool

	Index    reg.R64
	HasIndex bool
	Scale    uint8

	Disp int32

	// RIP and Abs are the two baseless forms. RIP is [rip + disp32]; Abs is
	// the SIB no-base form, which is how a plain numeric or symbolic
	// address is spelled now that ModRM's disp32 encoding means RIP.
	RIP bool
	Abs bool

	Ref    SymRef
	HasRef bool

	Seg    reg.Sreg
	HasSeg bool

	Bcst bool
}

// Parts returns the address's fields.
func (m Addr) Parts() Parts { return m.a.parts() }

func (a addr) parts() Parts {
	return Parts{
		Base: a.base, HasBase: a.hasBase,
		Index: a.index, HasIndex: a.hasIndex, Scale: a.scale,
		Disp: a.disp,
		RIP:  a.rip, Abs: a.abs,
		Ref: a.ref, HasRef: a.hasRef,
		Seg: a.seg, HasSeg: a.hasSeg,
		Bcst: a.bcst,
	}
}
