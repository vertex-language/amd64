// Package operand declares everything an instruction can take that is not a
// register: immediates, memory addresses, branch labels and symbol
// references.
//
// It imports reg, for the registers an address is built from and for the seal
// on the operand interface, and obj, for RefKind and the error vocabulary — a
// reference's link semantics are a fact about the artifact, so they belong in
// the artifact's vocabulary and are declared once, there.
//
// Errors here are sticky. A malformed address does not fail at the chain
// method that malformed it; it carries the failure and surfaces at the
// instruction that uses it, positioned by the section. A builder chain is not
// followed by a run of error checks.
package operand

import (
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/reg"
)

// Operand is the sealed interface every instruction operand satisfies.
//
// The seal is reg's, and it has to be: the method is unexported and declared
// there, so only a type in reg can implement it directly and only a type
// embedding reg.Seal can implement it indirectly. A seal declared in obj —
// which imports nothing from this tree — could never be satisfied by a
// register, and a register is the operand this tree is built out of.
//
// What the seal buys is that Emit's variadic cannot be handed a bare Go
// integer. NewImm(n) is the spelling, and that is what keeps an arbitrary
// type out of a mnemonic-as-data path.
type Operand = reg.Operand

// RefKind and the reference kinds are obj's, aliased here so a lowering that
// imports operand and not obj still spells them. Same constant, no conversion
// anywhere in the tree.
type RefKind = obj.RefKind

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

	RefGOT32        = obj.RefGOT32
	RefGOTPCREL     = obj.RefGOTPCREL
	RefGOTPCRELX    = obj.RefGOTPCRELX
	RefRexGOTPCRELX = obj.RefRexGOTPCRELX
	RefGOTOFF64     = obj.RefGOTOFF64
	RefGOTPC32      = obj.RefGOTPC32

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

// fail builds a positionless ErrOperand. The section fills in Section and
// Offset when the operand reaches an instruction, which is the only point at
// which a position exists.
func fail(context string, notes ...string) error {
	return &obj.Error{
		Arch:     obj.ArchAMD64,
		Sentinel: obj.ErrOperand,
		Context:  context,
		Notes:    notes,
	}
}
