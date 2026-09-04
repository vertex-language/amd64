package amd64

import (
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
)

// RefKind states how a linker should resolve a reference. It is declared
// once, in obj, and aliased upward.
type RefKind = obj.RefKind

// The kind set is the union of what the three containers can express, not
// the intersection. An intersection would be four kinds and would make
// every interesting object unbuildable; a per-container set would mean a
// lowering picks its container before it picks its instructions. So the set
// is the union, every writer states what it cannot do, and ErrRefKind names
// the kind and the offset when a kind meets a container with no answer.
const (
	RefAbs64  = obj.RefAbs64
	RefAbs32  = obj.RefAbs32
	RefAbs32S = obj.RefAbs32S
	RefAbs16  = obj.RefAbs16
	RefAbs8   = obj.RefAbs8

	RefPC64 = obj.RefPC64
	RefPC32 = obj.RefPC32
	RefPC16 = obj.RefPC16
	RefPC8  = obj.RefPC8

	RefPLT32 = obj.RefPLT32

	RefGOT32    = obj.RefGOT32
	RefGOTPCREL = obj.RefGOTPCREL
	RefGOTOFF64 = obj.RefGOTOFF64
	RefGOTPC32  = obj.RefGOTPC32

	// Two kinds for one relocation semantics, because a linker relaxing
	// mov foo@GOTPCREL(%rip), %reg into lea foo(%rip), %reg has to know how
	// many bytes back the instruction starts, and the answer depends on
	// whether a REX prefix is there. The encoder knows — it emitted the
	// prefix or did not — but emitting the relaxable kind is a statement
	// that the addend is exactly the one the psABI's transformation
	// assumes, so the kind is yours to state and the encoder refuses the
	// wrong one.
	RefGOTPCRELX    = obj.RefGOTPCRELX
	RefRexGOTPCRELX = obj.RefRexGOTPCRELX

	RefSize32 = obj.RefSize32
	RefSize64 = obj.RefSize64

	RefTLSGD    = obj.RefTLSGD
	RefTLSLD    = obj.RefTLSLD
	RefDTPOFF32 = obj.RefDTPOFF32
	RefDTPOFF64 = obj.RefDTPOFF64
	RefGOTTPOFF = obj.RefGOTTPOFF
	RefTPOFF32  = obj.RefTPOFF32
	RefTPOFF64  = obj.RefTPOFF64

	RefTLV = obj.RefTLV

	RefImageRel32 = obj.RefImageRel32
	RefSecRel32   = obj.RefSecRel32
	RefSecIdx     = obj.RefSecIdx
)

// Ref names a symbol and states its link semantics.
//
// The kind is stated at construction because it is a decision — PLT or
// direct, GOT or absolute, relaxable or not — and this package does not
// make decisions. call puts@plt and call puts are byte-identical, e8 either
// way, so the kind rides beside the bytes as data.
func Ref(sym string, kind RefKind) SymRef { return operand.Ref(sym, kind) }
