package encode

// The recommended multi-byte NOP sequences, indexed by length. Intel
// specifies up to nine bytes; AMD documents longer ones, and GNU as carries
// legacy sequences for processors this architecture predates. Nine is
// enough — a longer run is a sequence of nines.
//
// There is no gate on any of this. 0F 1F is in the x86-64 baseline, so
// every module this package can build can execute the long nops, which is
// the one place this architecture is simpler than its i386 sibling.
var nops = [10][]byte{
	1: {0x90},
	2: {0x66, 0x90},
	3: {0x0f, 0x1f, 0x00},
	4: {0x0f, 0x1f, 0x40, 0x00},
	5: {0x0f, 0x1f, 0x44, 0x00, 0x00},
	6: {0x66, 0x0f, 0x1f, 0x44, 0x00, 0x00},
	7: {0x0f, 0x1f, 0x80, 0x00, 0x00, 0x00, 0x00},
	8: {0x0f, 0x1f, 0x84, 0x00, 0x00, 0x00, 0x00, 0x00},
	9: {0x66, 0x0f, 0x1f, 0x84, 0x00, 0x00, 0x00, 0x00, 0x00},
}

// Nops returns exactly n bytes of no-operation, as few instructions as the
// table allows. Align pads a code section with this; a data section gets
// zeros, which is a different question with a different answer.
func Nops(n int) []byte {
	if n <= 0 {
		return nil
	}
	out := make([]byte, 0, n)
	for n > 9 {
		out = append(out, nops[9]...)
		n -= 9
	}
	return append(out, nops[n]...)
}
