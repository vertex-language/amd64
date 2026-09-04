package isa

import (
	"github.com/vertex-language/amd64/operand"
)

// The table. Each tranche has its own file and its own build function,
// called from the one init below; each has an inst_*.go opposite it in the
// root package. Nothing about an existing row changes when a tranche lands.
var (
	forms      []*Form
	byHelper   map[string]*Form
	byMnemonic map[string][]*Form
)

// buildLock runs last of the tranches, and that is an ordering
// requirement rather than a habit: it clones every row that admits LOCK,
// so a lockable row declared after it would have no clone.
func init() {
	buildBase()
	buildSystem()
	buildSSE()
	buildSIMD()
	buildAVX()
	buildLock()
	index()
}

// add appends a row. It is the only way into the table.
func add(f Form) {
	f.index = len(forms)
	forms = append(forms, &f)
}

// index builds the lookups, and is also this package's integrity check.
//
// Two forms sharing a helper name would be a duplicate Go identifier in
// inst_*.go and so a compile error there — but the table is the thing that
// makes the promise, so it checks itself, and it fails at program start
// whatever the caller was about to do.
func index() {
	byHelper = make(map[string]*Form, len(forms))
	byMnemonic = make(map[string][]*Form, 256)

	for _, f := range forms {
		if f.Mnemonic == "" {
			panic("isa: form with no mnemonic at index " + itoa(f.index))
		}
		if f.Helper == "" {
			panic("isa: " + f.String() + " has no helper name")
		}
		if prev, dup := byHelper[f.Helper]; dup {
			panic("isa: duplicate helper name " + f.Helper +
				": " + prev.String() + " and " + f.String())
		}
		byHelper[f.Helper] = f
		byMnemonic[f.Mnemonic] = append(byMnemonic[f.Mnemonic], f)
	}
}

// All returns every declared form in table order.
func All() []*Form {
	out := make([]*Form, len(forms))
	copy(out, forms)
	return out
}

// ByHelper returns the form a typed helper pins, or nil.
//
// inst_*.go calls this once per helper at package initialisation and panics
// on nil naming the missing form. That is the earliest a renamed row can
// fail, and it beats failing halfway through someone's code generation.
func ByHelper(name string) *Form { return byHelper[name] }

// ByMnemonic returns every form of a mnemonic, in table order.
func ByMnemonic(name string) []*Form { return byMnemonic[name] }

// Known reports whether the mnemonic exists at all. Emit uses it to tell
// "no such instruction" apart from "no form of it takes these operands",
// which are different errors and send a caller to different places.
func Known(name string) bool { return len(byMnemonic[name]) > 0 }

// Resolve returns every form of the mnemonic whose operand classes accept
// the operands, in table order.
//
// It does not choose. Choosing needs lengths, lengths need the encoder, and
// there is deliberately no size estimator in this tree that could disagree
// with it. Emit encodes every candidate and takes the shortest — except
// among rel forms, where it takes the widest.
//
// An empty result with Known(name) true means the operands matched nothing:
// that is ErrForm naming the operands. False means ErrForm naming the
// mnemonic.
func Resolve(mnemonic string, ops []operand.Operand) []*Form {
	cands := byMnemonic[mnemonic]
	if len(cands) == 0 {
		return nil
	}
	// A locking clone is reachable only by asking for it by its own name,
	// "lock add". byMnemonic keys each row by its own mnemonic, so the two
	// families never share a key and this is a second lock on the same door:
	// Emit resolving "add" cannot reach a clone however the table changes.
	wantLock := len(mnemonic) > 5 && mnemonic[:5] == "lock "
	out := make([]*Form, 0, 4)
	for _, f := range cands {
		if f.Lock() != wantLock {
			continue
		}
		if f.Accepts(ops) {
			out = append(out, f)
		}
	}
	return out
}

// Lock reports whether this row is a locking clone rather than a base form.
func (f *Form) Lock() bool { return len(f.Mnemonic) > 4 && f.Mnemonic[:5] == "lock " }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
