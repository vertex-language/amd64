package obj

import (
	"strconv"
	"strings"
)

// sentinel is the type behind every declared failure.
//
// It is a string type rather than errors.New so each sentinel is a comparable
// value with no allocation and no init order to think about. errors.Is finds
// them by equality, which is the only question anyone asks of them.
type sentinel string

func (s sentinel) Error() string { return string(s) }

// The sentinels. Each names one failure and only one, which is what makes
// errors.Is against them worth writing.
var (
	// ErrFeature: the form needs a feature the module's set does not hold.
	ErrFeature error = sentinel("feature not enabled")

	// ErrForm: the operands are the wrong kinds for the form, or no form of
	// that name exists.
	ErrForm error = sentinel("no such instruction form")

	// ErrOperand: the operands were the right kinds and one of them was
	// built wrong — RSP as an index, a scale of 3, AH in a REX-prefixed
	// form, a register handed to a locking form.
	//
	// It is its own sentinel rather than a flavour of ErrForm because "no
	// matching form" sends a caller hunting through the ISA table for a row
	// that is already there.
	ErrOperand error = sentinel("invalid operand")

	// ErrDuplicate: a name was defined twice, or asked for two ways.
	ErrDuplicate error = sentinel("duplicate definition")

	// ErrUndefined: a reference, an alias or a label names nothing.
	ErrUndefined error = sentinel("undefined symbol")

	// ErrRange: a value does not fit a field. Both the immediate at the call
	// site and the label displacement at Finalize.
	ErrRange error = sentinel("value out of range")

	// ErrAlign: an alignment is not a power of two.
	ErrAlign error = sentinel("invalid alignment")

	// ErrFinalized: the module is spent.
	ErrFinalized error = sentinel("module is finalized")

	// ErrRefKind: the writer has no relocation for a kind, or no encoding
	// for this Adjust. Only ever comes from a writer.
	ErrRefKind error = sentinel("unsupported relocation kind")

	// ErrSectionName: the writer cannot place a custom section. Only ever
	// comes from obj/macho.
	ErrSectionName error = sentinel("unsupported section name")
)

// Error is the one error type in this tree — the builder, all three writers,
// and both architectures — so tooling that formats a diagnostic is written
// once:
//
//	amd64 .text+0x11: 300 does not fit ADD r/m64, imm8: value out of range
//	  note: the immediate field of ADD r/m64, imm8 is 1 bytes; the range is -128..127
type Error struct {
	Arch Arch

	// Section and Offset are the position. They are empty and zero on an
	// error an operand built for itself, because a position did not exist
	// until the operand reached an instruction; the section fills them in.
	Section string
	Offset  int

	Context string

	// Sentinel is which declared failure this is.
	Sentinel error

	// Cause is the underlying error where one exists — an encoder failure,
	// most often. Its type may be internal to the package that raised it,
	// and anything a caller needs from it is restated in Notes as text.
	Cause error

	Notes []string
}

func (e *Error) Error() string {
	var b strings.Builder
	if e.Arch.Valid() {
		b.WriteString(e.Arch.String())
		b.WriteByte(' ')
	}
	if e.Section != "" {
		b.WriteString(e.Section)
		b.WriteString("+0x")
		b.WriteString(strconv.FormatInt(int64(e.Offset), 16))
		b.WriteString(": ")
	}
	b.WriteString(e.Context)
	if e.Sentinel != nil {
		b.WriteString(": ")
		b.WriteString(e.Sentinel.Error())
	}
	for _, n := range e.Notes {
		b.WriteString("\n  note: ")
		b.WriteString(n)
	}
	return b.String()
}

// Unwrap returns the sentinel and the cause, both of them.
//
// It is the multi-error form on purpose. A single Unwrap would make a caller
// choose which of the two questions they are allowed to ask, and the two are
// different questions: the sentinel is which failure this is, the cause is
// what the layer below said about it.
func (e *Error) Unwrap() []error {
	switch {
	case e.Sentinel != nil && e.Cause != nil:
		return []error{e.Sentinel, e.Cause}
	case e.Sentinel != nil:
		return []error{e.Sentinel}
	case e.Cause != nil:
		return []error{e.Cause}
	}
	return nil
}

// Positioned reports whether the error names a place in a section.
func (e *Error) Positioned() bool { return e.Section != "" }
