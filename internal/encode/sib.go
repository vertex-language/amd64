package encode

import "github.com/vertex-language/amd64/obj"

// SIB: scale[7:6] index[5:3] base[2:0].
func sib(scale, index, base byte) byte {
	return scale<<6 | (index&7)<<3 | base&7
}

const (
	// noIndex is index = 100 with REX.X clear, which is the encoding that
	// means no index at all. It is also RSP's number, which is why RSP
	// cannot be an index and R12 — same low bits, REX.X set — can.
	noIndex byte = 4

	// sibNoBase is base = 101, which with mod = 00 means the base term is
	// null and a disp32 follows.
	sibNoBase byte = 5
)

// scaleBits encodes 1, 2, 4 or 8 as the two-bit field.
func scaleBits(s uint8) (byte, error) {
	switch s {
	case 1:
		return 0, nil
	case 2:
		return 1, nil
	case 4:
		return 2, nil
	case 8:
		return 3, nil
	}
	return 0, errf(obj.ErrOperand,
		"scale must be 1, 2, 4 or 8",
		"the SIB scale field is two bits and encodes exactly those four values")
}
