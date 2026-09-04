package amd64

import (
	"errors"

	"github.com/vertex-language/amd64/internal/encode"
	"github.com/vertex-language/amd64/obj"
)

// fail records the first error and keeps it.
//
// Errors are sticky and first-wins: every builder call after a failure is a
// no-op, and Finalize surfaces the first one, positioned. That is what lets
// an instruction-selection loop run without a single error check in it.
func (m *Module) fail(err error) {
	if m.err == nil {
		m.err = err
	}
}

// errorAt builds a positioned diagnostic.
//
//	amd64 .text+0x11: 300 does not fit ADD r/m64, imm8: value out of range
//	  note: the immediate field of ADD r/m64, imm8 is 1 bytes; the range is -128..127
func (s *Section) errorAt(sentinel error, context string, notes ...string) *obj.Error {
	return &obj.Error{
		Arch:     obj.ArchAMD64,
		Section:  s.name,
		Offset:   len(s.buf),
		Context:  context,
		Sentinel: sentinel,
		Notes:    notes,
	}
}

// lift turns an error from the encoder or from an operand into a positioned
// one.
//
// An encoder failure reaches a caller through *obj.Error as a sentinel, a
// message and notes — data, not internal types. The internal error joins
// the chain as the cause, so Unwrap returns both it and the sentinel and
// errors.Is works against either, but nothing a caller needs is only
// reachable by asserting on it.
func (s *Section) lift(err error) *obj.Error {
	if err == nil {
		return nil
	}

	// An operand built this itself, positionless, because a position did
	// not exist until now.
	var oe *obj.Error
	if errors.As(err, &oe) && oe.Section == "" {
		positioned := *oe
		positioned.Section = s.name
		positioned.Offset = len(s.buf)
		return &positioned
	}
	if oe != nil {
		return oe
	}

	sentinel := encode.Sentinel(err)
	if sentinel == nil {
		sentinel = obj.ErrForm
	}
	return &obj.Error{
		Arch:     obj.ArchAMD64,
		Section:  s.name,
		Offset:   len(s.buf),
		Context:  err.Error(),
		Sentinel: sentinel,
		Cause:    err,
		Notes:    encode.Notes(err),
	}
}

// Err returns the first error, or nil. It is the same error Finalize will
// return, and it is there so a long build can bail out early rather than
// running to completion over a module that is already spent.
func (m *Module) Err() error { return m.err }
