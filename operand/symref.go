package operand

import "github.com/vertex-language/amd64/reg"

// SymRef names a symbol and states how the linker should resolve it.
//
// The kind is stated at construction because it is a decision — PLT or
// direct, GOT or absolute, relaxable or not — and this package does not make
// decisions. call puts@plt and call puts are byte-identical, e8 either way,
// so the kind rides beside the bytes as data rather than being inferred from
// them.
type SymRef struct {
	reg.Seal

	Sym    string
	Kind   RefKind
	Addend int64
}

// Ref names a symbol with the given link semantics.
func Ref(sym string, kind RefKind) SymRef {
	return SymRef{Sym: sym, Kind: kind}
}

// Add returns the reference with n added to its addend. A displacement
// applied to an address that names this symbol folds in here too, so there is
// one place for the number.
func (r SymRef) Add(n int64) SymRef {
	r.Addend += n
	return r
}

func (r SymRef) String() string {
	if r.Sym == "" {
		return "<unnamed reference>"
	}
	return r.Sym
}
