package obj

// Object is a finished relocatable object: immutable, pure data, safe to
// write more than once, in more than one format, from more than one
// goroutine.
//
// Nothing on it mutates and every slice accessor returns a copy, which is
// what lets three writers hold the same object without any of them being able
// to see another's work. Two of the three fold addends into the section
// bytes; on this architecture ELF does not, because x86-64 is RELA and the
// addend lives in the relocation entry.
type Object struct {
	arch Arch

	sections []*Section
	symbols  []Symbol

	byName map[string]*Section
	symAt  map[string]int
}

// Section is one finished section of an object.
type Section struct {
	o *Object

	name  string
	kind  SectionKind
	index int
	align int

	bytes []byte
	refs  []Reference
}

// SectionData is one section handed to New.
//
// It exists because Object's fields are unexported and the builder that fills
// them is a different package. A struct of plain data is a narrower seam than
// exporting the fields would be: New copies everything it is given, so the
// caller cannot retain a slice and reach back into a finished object.
type SectionData struct {
	Name  string
	Kind  SectionKind
	Align int
	Bytes []byte
	Refs  []Reference
}

// New assembles a finished object.
//
// It is the one way to make one, and it copies: the bytes, the references and
// the symbol table are all duplicated out of the builder's storage. That copy
// is the moment the artifact becomes inert, and paying for it once here is
// what makes every accessor below able to hand out copies cheaply — and what
// makes it true that finalizing a module twice returns the same object rather
// than a second pass over live state.
//
// An Align below 1 becomes 1. A section's index is its position in secs, and
// that index is what a symbol's Section names.
func New(arch Arch, secs []SectionData, syms []Symbol) *Object {
	o := &Object{
		arch:     arch,
		sections: make([]*Section, 0, len(secs)),
		symbols:  make([]Symbol, len(syms)),
		byName:   make(map[string]*Section, len(secs)),
		symAt:    make(map[string]int, len(syms)),
	}
	copy(o.symbols, syms)

	for i, sd := range secs {
		align := sd.Align
		if align < 1 {
			align = 1
		}
		s := &Section{
			o:     o,
			name:  sd.Name,
			kind:  sd.Kind,
			index: i,
			align: align,
			bytes: append([]byte(nil), sd.Bytes...),
			refs:  append([]Reference(nil), sd.Refs...),
		}
		o.sections = append(o.sections, s)
		// A duplicate name cannot happen: the builder refuses one with
		// ErrDuplicate long before Finalize. First wins if it ever does,
		// rather than silently rebinding the lookup to the later section.
		if _, dup := o.byName[s.name]; !dup {
			o.byName[s.name] = s
		}
	}
	for i, sym := range o.symbols {
		if _, dup := o.symAt[sym.Name]; !dup {
			o.symAt[sym.Name] = i
		}
	}
	return o
}

// Arch is the architecture this object was built for. Every writer derives
// its target from it, and none of them takes one as an option.
func (o *Object) Arch() Arch { return o.arch }

// Sections returns the object's sections in creation order.
func (o *Object) Sections() []*Section {
	out := make([]*Section, len(o.sections))
	copy(out, o.sections)
	return out
}

// SectionAt returns the section with the given index, or nil. The index is
// what a symbol's Section field names.
func (o *Object) SectionAt(i int) *Section {
	if i < 0 || i >= len(o.sections) {
		return nil
	}
	return o.sections[i]
}

// SectionNamed returns the section with the given name, or nil.
func (o *Object) SectionNamed(name string) *Section {
	return o.byName[name]
}

// Symbols returns the object's one symbol table, in definition order.
func (o *Object) Symbols() []Symbol {
	out := make([]Symbol, len(o.symbols))
	copy(out, o.symbols)
	return out
}

// Symbol returns the symbol with the given name. Identity is the name,
// module-wide, so there is at most one.
func (o *Object) Symbol(name string) (Symbol, bool) {
	i, ok := o.symAt[name]
	if !ok {
		return Symbol{}, false
	}
	return o.symbols[i], true
}

func (s *Section) Name() string      { return s.name }
func (s *Section) Kind() SectionKind { return s.kind }

// Index is the section's position, which is what a symbol's Section names.
func (s *Section) Index() int { return s.index }

// Align is the largest alignment the builder asked for, at least 1. It is
// what the object writer stamps on the section header.
func (s *Section) Align() int { return s.align }

// Size is the section's length in bytes. A BSS section has a size and no
// bytes, so this is the field a writer sizes the gap from.
func (s *Section) Size() int { return len(s.bytes) }

// Bytes returns the finished contents, same-section labels already patched.
// It is a copy, and so is every other slice accessor here.
func (s *Section) Bytes() []byte {
	out := make([]byte, len(s.bytes))
	copy(out, s.bytes)
	return out
}

// Refs returns the holes a linker fills, in the order the builder placed
// them. Order is preserved rather than sorted: a writer that emits pairs
// relies on adjacency, and nothing in a Reference names its partner.
func (s *Section) Refs() []Reference {
	out := make([]Reference, len(s.refs))
	copy(out, s.refs)
	return out
}

// Symbols returns the symbols defined in this section, in table order. It is
// a filtered view over the object's one table, not a table of its own.
func (s *Section) Symbols() []Symbol {
	var out []Symbol
	for _, sym := range s.o.symbols {
		if sym.Section == s.index {
			out = append(out, sym)
		}
	}
	return out
}

func (s *Section) String() string { return s.name }
