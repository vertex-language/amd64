package elf

import (
	"fmt"

	elfcore "github.com/vertex-language/elf"
	elfobj "github.com/vertex-language/elf/obj"

	"github.com/vertex-language/amd64/obj"
)

// The symbol vocabulary is a straight translation: obj's three bindings, four
// types and four visibilities are ELF's, under names that do not stutter with
// obj.Object. Nothing here is inferred — a symbol is what the builder said it
// was, and a writer that improved on that would make Symbols() a lie.
//
// ELF is the container that carries all of it. The other two drop something,
// which is why the vocabulary is spelled this way and translated outward.

func bindingOf(b obj.Binding) elfcore.SymBind {
	switch b {
	case obj.Global:
		return elfcore.STB_GLOBAL
	case obj.Weak:
		return elfcore.STB_WEAK
	}
	return elfcore.STB_LOCAL
}

func typeOf(t obj.SymbolType) elfcore.SymType {
	switch t {
	case obj.Func:
		return elfcore.STT_FUNC
	case obj.ObjectSym:
		return elfcore.STT_OBJECT
	case obj.ThreadLocal:
		return elfcore.STT_TLS
	}
	return elfcore.STT_NOTYPE
}

func visibilityOf(v obj.Visibility) elfcore.SymVisibility {
	switch v {
	case obj.Hidden:
		return elfcore.STV_HIDDEN
	case obj.Protected:
		return elfcore.STV_PROTECTED
	case obj.Internal:
		return elfcore.STV_INTERNAL
	}
	return elfcore.STV_DEFAULT
}

// writeSymbols declares every symbol and returns the handles relocations
// resolve through.
//
// The map is by name because symbol identity is the name, module-wide — the
// same rule the builder enforces at the second Label — so a reference's Sym
// field is a key and not an index that has to survive a reordering.
func writeSymbols(wr *elfobj.Writer, o *obj.Object, builders []*elfobj.SectionBuilder, opt Options) (map[string]elfobj.SymRef, error) {
	syms := o.Symbols()
	out := make(map[string]elfobj.SymRef, len(syms))

	var referenced map[string]bool
	if opt.StripLocals {
		referenced = referencedNames(o)
	}

	for _, sym := range syms {
		if opt.StripLocals && sym.Binding == obj.Local && !referenced[sym.Name] {
			continue
		}

		def := elfobj.SymbolDef{
			Name:  sym.Name,
			Bind:  bindingOf(sym.Binding),
			Type:  typeOf(sym.Type),
			Other: elfcore.WithVisibility(0, visibilityOf(sym.Visibility)),
		}

		if sym.Defined() {
			if sym.Section < 0 || sym.Section >= len(builders) {
				// Finalize guarantees this, so reaching here is a bug in this
				// tree rather than anything a caller did. It is still checked,
				// because the alternative is an index panic inside another
				// module.
				return nil, fmt.Errorf("elf: symbol %q names section %d of %d",
					sym.Name, sym.Section, len(builders))
			}
			def.Where = elfobj.SymInSection
			def.Section = builders[sym.Section]
			def.Value = uint64(sym.Offset)
			def.Size = uint64(sym.Size)
		}

		out[sym.Name] = wr.Symbol(def)
	}
	return out, nil
}

// referencedNames is every symbol a relocation names. It exists so StripLocals
// can drop what nothing points at and keep what does.
func referencedNames(o *obj.Object) map[string]bool {
	out := map[string]bool{}
	for _, s := range o.Sections() {
		for _, r := range s.Refs() {
			out[r.Sym] = true
		}
	}
	return out
}
