package encode

import "github.com/vertex-language/amd64/obj"

// encErr is the encoder's failure. Its type is internal on purpose:
// anything a caller might need from it is restated in the section's
// *obj.Error as text, and the section joins this into the chain as the
// cause.
type encErr struct {
	sentinel error
	msg      string
	notes    []string
}

func errf(sentinel error, msg string, notes ...string) error {
	return &encErr{sentinel: sentinel, msg: msg, notes: notes}
}

func (e *encErr) Error() string { return e.msg }
func (e *encErr) Unwrap() error { return e.sentinel }

// Sentinel and Notes are how the section lifts this into an *obj.Error
// without exporting the type.
func (e *encErr) Sentinel() error { return e.sentinel }
func (e *encErr) Notes() []string { return e.notes }

// Sentinel reports the obj sentinel behind an encoder error, or nil.
func Sentinel(err error) error {
	if e, ok := err.(*encErr); ok {
		return e.sentinel
	}
	return nil
}

// Notes reports an encoder error's notes, or nil.
func Notes(err error) []string {
	if e, ok := err.(*encErr); ok {
		return e.notes
	}
	return nil
}

var _ = obj.ErrForm // the sentinels this package raises all live in obj

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
