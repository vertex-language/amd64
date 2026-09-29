package pe

import (
	"fmt"

	pecore "github.com/vertex-language/pe"
	"github.com/vertex-language/pe/coff"

	"github.com/vertex-language/amd64/obj"
)

// COFF spells linkage as a storage class, and it has two of the four things
// obj carries and not the other two.
//
//   - Binding is the storage class: Global is EXTERNAL, Local is STATIC.
//     A weak definition becomes STATIC, because COFF has no weak
//     definition; see classOf. A weak *reference* -- an undefined weak
//     symbol -- is what COFF does have: a WEAK_EXTERNAL; see weakRef.
//   - SymbolType survives only as the derived-function bit, which is what
//     tells a debugger and /OPT:REF that a symbol is code. ObjectSym and
//     ThreadLocal have no COFF spelling and become NULL: TLS in COFF is a
//     property of the .tls section a variable lands in, not of its symbol.
//   - Size has no field at all. A COFF symbol's extent is the distance to the
//     next one, which is why EndLabel exists on the builder and why nothing
//     here can carry what it recorded.
//   - Visibility has no field either. Whether a name leaves a DLL is decided
//     by /EXPORT or a .def file at link time, not by the object, so all four
//     values are accepted and dropped rather than refused — refusing Hidden
//     would reject an object that links correctly.
func classOf(sym obj.Symbol) (pecore.StorageClass, error) {
	switch sym.Binding {
	case obj.Global:
		return pecore.ClassExternal, nil
	case obj.Local:
		return pecore.ClassStatic, nil
	case obj.Weak:
		// A weak *definition* has no direct spelling in COFF.
		// IMAGE_SYM_CLASS_WEAK_EXTERNAL is a weak reference — an
		// undefined symbol with an alternate — and cannot be a
		// definition at all; what MSVC emits for an inline function is a
		// COMDAT section with SELECT_ANY keyed on the symbol, which needs
		// the function to be in a section of its own and this writer
		// puts every function in one .text.
		//
		// So it becomes static, and the difference is real: each object
		// keeps its own copy rather than the linker electing one, and
		// two translation units that both emit an inline function
		// disagree about its address. That is the cost, and it is paid
		// rather than refused because refusing is refusing to compile
		// against the Windows SDK, whose <stdio.h> defines printf as an
		// inline function and provides no other.
		return pecore.ClassStatic, nil
	}
	return pecore.ClassExternal, nil
}

func symTypeOf(t obj.SymbolType) pecore.SymType {
	if t == obj.Func {
		return pecore.PackSymType(pecore.BaseNull, pecore.DerivedFunction)
	}
	return pecore.PackSymType(pecore.BaseNull, pecore.DerivedNull)
}

// writeSymbols declares every symbol and returns the handles relocations
// resolve through.
//
// An undefined symbol — a name declared Extern — is an external with no
// section and a zero Value. A non-zero Value there would be a common-block
// request, which is a different thing entirely and one this builder cannot
// express, so the zero is load-bearing.
func writeSymbols(wr *coff.Writer, o *obj.Object, builders []*coff.SectionBuilder) (map[string]*coff.SymbolRef, error) {
	syms := o.Symbols()
	out := make(map[string]*coff.SymbolRef, len(syms))
	suffix := weakSuffix(syms)

	for _, sym := range syms {
		if !sym.Defined() && sym.Binding == obj.Weak {
			out[sym.Name] = weakRef(wr, sym.Name, suffix)
			continue
		}
		class, err := classOf(sym)
		if err != nil {
			return nil, err
		}

		def := coff.SymbolDef{
			Name:  sym.Name,
			Class: class,
			Type:  symTypeOf(sym.Type),
		}
		if sym.Defined() {
			if sym.Section < 0 || sym.Section >= len(builders) {
				return nil, fmt.Errorf("pe: symbol %q names section %d of %d",
					sym.Name, sym.Section, len(builders))
			}
			def.Section = builders[sym.Section]
			def.Value = uint32(sym.Offset)
		}

		out[sym.Name] = wr.Symbol(def)
	}
	return out, wr.Err()
}

// weakRef declares an undefined weak symbol: a reference that binds to a
// definition if the link has one and to zero if it does not.
//
// That is a WEAK_EXTERNAL, whose auxiliary record names the alternate the
// linker falls back to -- here an absolute zero, which is what a weak
// reference to nothing reads as on every other container. The alternate is
// looked up by name, so it is external, and so its name has to be unique to
// this object or two objects referencing one weak name would each define
// it: the suffix is LLVM's, the object's first external definition.
//
// The search kind is ALIAS, measured: clang's x86_64-windows-msvc object
// for `extern int x __attribute__((weak))` is exactly this -- x a
// WEAK_EXTERNAL, Search: Alias, linked to .weak.x.default.f, an EXTERNAL
// absolute zero.
func weakRef(wr *coff.Writer, name, suffix string) *coff.SymbolRef {
	alt := wr.Symbol(coff.SymbolDef{
		Name:     ".weak." + name + ".default" + suffix,
		Class:    pecore.ClassExternal,
		Absolute: true,
	})
	return wr.WeakSymbol(name, alt, pecore.WeakAlias)
}

// weakSuffix is what makes this object's weak alternates' names its own:
// a dot and its first external definition, or nothing if it has none.
func weakSuffix(syms []obj.Symbol) string {
	for _, s := range syms {
		if s.Defined() && s.Binding == obj.Global {
			return "." + s.Name
		}
	}
	return ""
}
