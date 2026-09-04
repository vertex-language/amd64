package amd64_test

import (
	"testing"

	"github.com/vertex-language/amd64"
)

// MemBytes answers for the mnemonics that have one answer, and declines for
// the ones that do not. It is what an AT&T front end asks when neither a
// suffix nor a register says how wide a memory operand is.
func TestMemBytes(t *testing.T) {
	for _, tc := range []struct {
		mnem string
		n    int
		ok   bool
	}{
		{"lea", 0, true},         // an address with no access width
		{"cmpxchg16b", 16, true}, // one size and no way to write it
		{"mov", 0, false},        // every width; the suffix has to say
		{"movzx", 0, false},      // r/m8 and r/m16 both
		{"nosuchthing", 0, false},
	} {
		n, ok := amd64.MemBytes(tc.mnem)
		if n != tc.n || ok != tc.ok {
			t.Errorf("MemBytes(%q) = %d, %v; want %d, %v", tc.mnem, n, ok, tc.n, tc.ok)
		}
	}
}
