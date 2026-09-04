// Package amd64 builds AMD64 objects: registers, the ISA surface, and an
// in-memory object you can hand straight to an ELF, COFF or Mach-O writer.
//
// This package never learns what a linker is. Nothing here imports link,
// backend or image from any sibling module.
package amd64

import (
	"github.com/vertex-language/amd64/feature"
	"github.com/vertex-language/amd64/obj"
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

// One declaration per concept, aliased upward. A value crossing a package
// line in this tree is never converted, only renamed: amd64.RefPLT32,
// operand.RefPLT32 and obj.RefPLT32 are the same constant and no conversion
// exists anywhere.

type (
	Object    = obj.Object
	Reference = obj.Reference
	Error     = obj.Error
	Symbol    = obj.Symbol

	Binding    = obj.Binding
	SymbolType = obj.SymbolType
	Visibility = obj.Visibility
)

const (
	Local  = obj.Local
	Global = obj.Global
	Weak   = obj.Weak

	NoType      = obj.NoType
	Func        = obj.Func
	ObjectSym   = obj.ObjectSym
	ThreadLocal = obj.ThreadLocal

	Default   = obj.Default
	Hidden    = obj.Hidden
	Protected = obj.Protected
	Internal  = obj.Internal
)

// The error sentinels. errors.Is works against every one of them, and the
// concrete type behind all of them is *obj.Error — one error type across
// the builder, all three writers and both architectures in the tree, so
// tooling that formats a diagnostic is written once.
var (
	ErrFeature     = obj.ErrFeature
	ErrForm        = obj.ErrForm
	ErrOperand     = obj.ErrOperand
	ErrDuplicate   = obj.ErrDuplicate
	ErrUndefined   = obj.ErrUndefined
	ErrRange       = obj.ErrRange
	ErrAlign       = obj.ErrAlign
	ErrFinalized   = obj.ErrFinalized
	ErrRefKind     = obj.ErrRefKind
	ErrSectionName = obj.ErrSectionName
)

// ---- registers ------------------------------------------------------------
//
// The numbered registers drop their width suffix here. reg spells the
// 64-bit r8 as R8Q because the type R8 already claims that identifier;
// callers write R8, which is what the ISA calls it.

const (
	RAX, RCX, RDX, RBX = reg.RAX, reg.RCX, reg.RDX, reg.RBX
	RSP, RBP, RSI, RDI = reg.RSP, reg.RBP, reg.RSI, reg.RDI

	R8, R9, R10, R11   = reg.R8Q, reg.R9Q, reg.R10Q, reg.R11Q
	R12, R13, R14, R15 = reg.R12Q, reg.R13Q, reg.R14Q, reg.R15Q

	EAX, ECX, EDX, EBX = reg.EAX, reg.ECX, reg.EDX, reg.EBX
	ESP, EBP, ESI, EDI = reg.ESP, reg.EBP, reg.ESI, reg.EDI

	R8D, R9D, R10D, R11D   = reg.R8D, reg.R9D, reg.R10D, reg.R11D
	R12D, R13D, R14D, R15D = reg.R12D, reg.R13D, reg.R14D, reg.R15D

	AX, CX, DX, BX = reg.AX, reg.CX, reg.DX, reg.BX
	SP, BP, SI, DI = reg.SP, reg.BP, reg.SI, reg.DI

	R8W, R9W, R10W, R11W   = reg.R8W, reg.R9W, reg.R10W, reg.R11W
	R12W, R13W, R14W, R15W = reg.R12W, reg.R13W, reg.R14W, reg.R15W

	AL, CL, DL, BL     = reg.AL, reg.CL, reg.DL, reg.BL
	SPL, BPL, SIL, DIL = reg.SPL, reg.BPL, reg.SIL, reg.DIL
	AH, CH, DH, BH     = reg.AH, reg.CH, reg.DH, reg.BH

	R8B, R9B, R10B, R11B   = reg.R8B, reg.R9B, reg.R10B, reg.R11B
	R12B, R13B, R14B, R15B = reg.R12B, reg.R13B, reg.R14B, reg.R15B

	ES, CS, SS, DS, FS, GS = reg.ES, reg.CS, reg.SS, reg.DS, reg.FS, reg.GS
)

// RIP is a base for Rip() and nothing else. It is not an R64 and satisfies
// no operand interface, because no instruction takes it as an operand.
var RIP = reg.RIP

// The vector and mask registers are declared in reg and re-exported by the
// generated block in alias_vec.go, because thirty-two of each in three
// widths is a list nobody should maintain by hand.

// ---- operands -------------------------------------------------------------

type (
	Imm    = operand.Imm
	Label  = operand.Label
	SymRef = operand.SymRef
	Memory = operand.Memory
	Addr   = operand.Addr
)

// The access widths and the register-or-memory classes, re-exported so a
// lowering that writes its own wrappers around these helpers spells its
// parameter types at the root and never imports operand for them.
type (
	M8   = operand.M8
	M16  = operand.M16
	M32  = operand.M32
	M64  = operand.M64
	M128 = operand.M128
	M256 = operand.M256
	M512 = operand.M512

	RM8   = operand.RM8
	RM16  = operand.RM16
	RM32  = operand.RM32
	RM64  = operand.RM64
	RM128 = operand.RM128
	RM256 = operand.RM256
	RM512 = operand.RM512
)

// The memory constructors. Four kinds of address, one question each, and
// the number is always the access width.
var (
	Mem8, Mem16, Mem32, Mem64 = operand.Mem8, operand.Mem16, operand.Mem32, operand.Mem64
	Mem128, Mem256, Mem512    = operand.Mem128, operand.Mem256, operand.Mem512

	Rip                = operand.Rip
	Rip8, Rip16, Rip32 = operand.Rip8, operand.Rip16, operand.Rip32
	Rip64, Rip128      = operand.Rip64, operand.Rip128
	Rip256, Rip512     = operand.Rip256, operand.Rip512

	Abs8, Abs16, Abs32, Abs64 = operand.Abs8, operand.Abs16, operand.Abs32, operand.Abs64
	Abs128, Abs256, Abs512    = operand.Abs128, operand.Abs256, operand.Abs512

	Addr8, Addr16, Addr32, Addr64 = operand.Addr8, operand.Addr16, operand.Addr32, operand.Addr64
	Addr128, Addr256, Addr512     = operand.Addr128, operand.Addr256, operand.Addr512

	NewImm   = operand.NewImm
	NewImmU  = operand.NewImmU
	NewLabel = operand.NewLabel
)

// ---- features -------------------------------------------------------------

type (
	Feature    = feature.Feature
	FeatureSet = feature.Set
	Level      = feature.Level
)

const (
	V1, V2, V3, V4 = feature.V1, feature.V2, feature.V3, feature.V4

	AES, PCLMULQDQ, SHA = feature.AES, feature.PCLMULQDQ, feature.SHA
	RDRAND, RDSEED, ADX = feature.RDRAND, feature.RDSEED, feature.ADX
	AVX, AVX2, FMA      = feature.AVX, feature.AVX2, feature.FMA
	AVX512F, AVX512BW   = feature.AVX512F, feature.AVX512BW
	AVX512CD, AVX512DQ  = feature.AVX512CD, feature.AVX512DQ
	AVX512VL            = feature.AVX512VL
)

var (
	NewFeatureSet   = feature.NewSet
	DefaultFeatures = feature.Default
	ParseFeatures   = feature.Parse
)
