package macho

import (
	"fmt"

	machocore "github.com/vertex-language/macho"
	machoobj "github.com/vertex-language/macho/obj"

	"github.com/vertex-language/amd64/obj"
)

// Mach-O spells linkage as two bits on the type byte, and it has two of the
// four things obj carries.
//
//   - Binding is N_EXT and the weak bits. Local is neither; Global is N_EXT;
//     Weak is N_EXT plus N_WEAK_DEF where the module defines the symbol and
//     N_WEAK_REF where it only names one. Unlike COFF nothing is refused: a
//     Mach-O weak definition is a weaker binding and needs no alternate.
//   - Size has no field. A Mach-O symbol's extent is the distance to the next
//     one, which is what MH_SUBSECTIONS_VIA_SYMBOLS is a promise about, so
//     what EndLabel recorded is dropped here.
//   - SymbolType has no field either. Func, ObjectSym and ThreadLocal are all
//     N_SECT; whether a symbol is code is a property of the section it lands
//     in. Accepted and dropped rather than refused, for the same reason COFF
//     drops Visibility: refusing would reject an object that links correctly.
//   - Visibility survives only as N_PEXT, the private external, which is
//     roughly Hidden. Protected and Internal have no spelling and are
//     dropped.
//
// Nothing here mangles. A Mach-O symbol conventionally leads with an
// underscore, and that convention belongs to the frontend that chose the
// names — a writer that added one would make Symbols() a lie about the object
// it came from.
func writeSymbols(wr *machoobj.Writer, o *obj.Object, builders []*machoobj.SectionBuilder) (map[string]machoobj.SymRef, error) {
	syms := o.Symbols()
	out := make(map[string]machoobj.SymRef, len(syms))

	for _, sym := range syms {
		def := machoobj.SymbolDef{
			Name: sym.Name,
			Type: machocore.N_UNDF,
			Ext:  sym.Binding != obj.Local,
			Pext: sym.Binding != obj.Local && sym.Visibility == obj.Hidden,
		}

		if sym.Defined() {
			if sym.Section < 0 || sym.Section >= len(builders) {
				return nil, fmt.Errorf("macho: symbol %q names section %d of %d",
					sym.Name, sym.Section, len(builders))
			}
			def.Type = machocore.N_SECT
			def.Section = builders[sym.Section]
			def.Value = uint64(sym.Offset)
		}

		if sym.Binding == obj.Weak {
			// A weak definition and a weak reference are different bits, and
			// which one this is follows from whether the module defined it.
			if sym.Defined() {
				def.WeakDef = true
			} else {
				def.WeakRef = true
			}
		}

		out[sym.Name] = wr.Symbol(def)
	}
	return out, wr.Err()
}
