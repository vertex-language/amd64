package obj

// Binding is a symbol's linkage.
type Binding uint8

const (
	Local Binding = iota
	Global
	Weak
)

func (b Binding) String() string {
	switch b {
	case Local:
		return "local"
	case Global:
		return "global"
	case Weak:
		return "weak"
	}
	return "binding?"
}

// SymbolType is what a symbol names.
//
// ObjectSym rather than Object because the artifact already claims that
// identifier. The root re-exports it under the same spelling, so a caller
// writes amd64.ObjectSym and amd64.Object stays the type.
type SymbolType uint8

const (
	NoType SymbolType = iota
	Func
	ObjectSym
	ThreadLocal
)

func (t SymbolType) String() string {
	switch t {
	case NoType:
		return "notype"
	case Func:
		return "func"
	case ObjectSym:
		return "object"
	case ThreadLocal:
		return "tls"
	}
	return "symtype?"
}

// Visibility is how far a defined symbol reaches.
//
// Only ELF carries all four. COFF drops the field entirely, because whether a
// name leaves a DLL is decided at link time, and Mach-O expresses roughly
// Hidden as N_PEXT. Both accept and discard rather than refusing: refusing
// Hidden would reject an object that links correctly.
type Visibility uint8

const (
	Default Visibility = iota
	Hidden
	Protected
	Internal
)

func (v Visibility) String() string {
	switch v {
	case Default:
		return "default"
	case Hidden:
		return "hidden"
	case Protected:
		return "protected"
	case Internal:
		return "internal"
	}
	return "visibility?"
}

// Symbol is one entry of the object's symbol table.
//
// Identity is the name, module-wide. There is one table and it is in
// definition order, which is what every object writer wants.
//
// Section is an index into Object.Sections, or -1 when the symbol is
// undefined — declared with Extern and left for the linker. Offset is within
// that section, and Size is the extent: closed by EndLabel if the builder was
// asked to, at the next symbol in the same section otherwise, and at the
// section end failing that.
type Symbol struct {
	Name    string
	Section int
	Offset  int
	Size    int

	Binding    Binding
	Type       SymbolType
	Visibility Visibility
}

// Defined reports whether this object provides the symbol rather than merely
// naming it. Section is -1 when it does not.
func (s Symbol) Defined() bool { return s.Section >= 0 }

func (s Symbol) String() string {
	if s.Name == "" {
		return "<unnamed symbol>"
	}
	if !s.Defined() {
		return s.Name + " [undefined]"
	}
	return s.Name
}
