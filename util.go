package amd64

// The small shared helpers. They are here rather than beside their callers
// because each has more than one, and a function copied into two files is two
// functions that can drift.

// hex renders an offset the way a diagnostic prints one, with no 0x prefix —
// the caller supplies that, because "at .text+0x11" and "0x11 bytes" want
// different framing around the same digits.
func hex(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	u := uint64(n)
	if neg {
		u = uint64(-n)
	}
	const digits = "0123456789abcdef"
	var b [17]byte
	i := len(b)
	for u > 0 {
		i--
		b[i] = digits[u&0xf]
		u >>= 4
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// decimal is strconv.FormatInt without the import. This package's diagnostics
// are the only caller and they want decimal and nothing else.
//
// It takes an int64 because an immediate and a displacement are int64 and a
// count is an int, and one function serving both is one thing that can be
// wrong rather than two.
func decimal(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	u := uint64(v)
	if neg {
		u = uint64(-v)
	}
	var b [21]byte
	i := len(b)
	for u > 0 {
		i--
		b[i] = byte('0' + u%10)
		u /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// itoa is decimal for the counts and widths a diagnostic quotes.
func itoa(n int) string { return decimal(int64(n)) }

// fitsField reports whether a patched value fits a field of the given width.
//
// The test is signed, and deliberately not the encoder's: encode.fits also
// accepts an unsigned bit pattern, because 0xff into an imm8 is a byte and
// refusing it would make masks unwritable. Nothing that reaches here is a
// mask. Every value this sees is a displacement or a same-section difference,
// both of which are signed quantities, and accepting 0xffffff80 as a rel8
// would turn a branch that does not reach into one that lands 128 bytes into
// the wrong instruction.
func fitsField(v int64, size int) bool {
	if size >= 8 {
		return true
	}
	bits := size * 8
	lo := int64(-1) << (bits - 1)
	hi := int64(1)<<(bits-1) - 1
	return v >= lo && v <= hi
}

// fieldRange is the reachable range of a signed field, for the ErrRange note.
// The numbers are spelled out rather than computed so the message reads the
// way the manual does.
func fieldRange(size int) string {
	switch size {
	case 1:
		return "-128..127"
	case 2:
		return "-32768..32767"
	case 4:
		return "-2147483648..2147483647"
	case 8:
		return "-9223372036854775808..9223372036854775807"
	}
	return "the field's width"
}
