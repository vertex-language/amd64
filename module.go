package amd64

import (
	"github.com/vertex-language/amd64/feature"
	"github.com/vertex-language/amd64/obj"
)

// Module is the builder. It is spent once Finalize returns.
type Module struct {
	features feature.Set

	sections []*Section
	byName   map[string]*Section

	symbols []symbol
	symAt   map[string]int

	aliases []aliasReq
	visReqs []visReq

	err error

	done     bool
	final    *obj.Object
	finalErr error

	// spent is the handle every call returns once the module is finalized,
	// and the handle a refused SectionNamed returns. Every call on it is a
	// no-op, and it is never added to sections, so Sections() is as frozen
	// as the bytes are.
	spent *Section
}

// Option configures a module at construction.
type Option func(*Module)

// WithFeatures fixes the feature set.
//
// It is fixed at construction and nothing about it is configurable after: a
// gate that changed mid-module would make an already-emitted diagnostic
// unfalsifiable, because the same call would have succeeded or failed
// depending on when it ran.
func WithFeatures(s FeatureSet) Option {
	return func(m *Module) { m.features = s }
}

// NewModule returns an empty module. The default feature set is x86-64-v1,
// the baseline, with no extensions.
func NewModule(opts ...Option) *Module {
	m := &Module{
		features: feature.Default(),
		byName:   make(map[string]*Section, 4),
		symAt:    make(map[string]int, 16),
	}
	for _, o := range opts {
		o(m)
	}
	m.spent = &Section{m: m, dead: true, name: "<spent>"}
	return m
}

// Features is the set the module was built with.
func (m *Module) Features() FeatureSet { return m.features }

// Section returns the section of the given kind under its conventional
// name, creating it on first use.
//
// It is exactly SectionNamed(k.String(), k), so the two cannot disagree
// about what ".text" is.
func (m *Module) Section(k SectionKind) *Section {
	return m.SectionNamed(k.String(), k)
}

// SectionNamed returns a section by name, creating it with the given
// load-time kind on first use.
//
// A name asked for twice returns the same section. A name asked for with a
// different kind than it was created with is ErrDuplicate, because a
// section that is code on Tuesday and data on Wednesday is a bug with a
// delayed fuse.
//
// Names are spelled the ELF way and translated by the writers: .rodata
// becomes .rdata for COFF and (__TEXT,__const) for Mach-O. One of the three
// containers had to be the one the vocabulary borrows from.
func (m *Module) SectionNamed(name string, k SectionKind) *Section {
	if m.done {
		m.fail(&obj.Error{
			Arch: obj.ArchAMD64, Sentinel: obj.ErrFinalized,
			Context: "SectionNamed(" + name + ") after Finalize",
		})
		return m.spent
	}
	if s, ok := m.byName[name]; ok {
		if s.kind != k {
			m.fail(&obj.Error{
				Arch: obj.ArchAMD64, Section: name, Sentinel: obj.ErrDuplicate,
				Context: "section " + name + " already exists with a different kind",
				Notes: []string{
					"it was created as " + s.kind.String() + " and is now asked for as " + k.String(),
				},
			})
			return m.spent
		}
		return s
	}
	if m.err != nil {
		// The module is already failed. Hand back something usable so the
		// caller's build runs to its own end, and do not add it: a section
		// created after the first failure would change Sections() depending
		// on where the failure was.
		return m.spent
	}

	s := &Section{
		m:      m,
		kind:   k,
		name:   name,
		index:  len(m.sections),
		align:  1,
		labels: make(map[string]int, 8),
	}
	m.sections = append(m.sections, s)
	m.byName[name] = s
	return s
}

// Sections returns the module's sections in creation order.
func (m *Module) Sections() []*Section {
	out := make([]*Section, len(m.sections))
	copy(out, m.sections)
	return out
}

// Defines reports whether this module defines name.
//
// It is the question a caller assembling text into a module that is already
// partly built has to ask before declaring an implicit extern. GNU as reads
// a name it never sees defined as one the object imports, and that reading
// is right for a whole .s file and wrong for a fragment: a name the fragment
// does not define may still be defined by the module around it — another
// function, or a block label the emitter has already reached.
func (m *Module) Defines(name string) bool {
	i, ok := m.symAt[name]
	return ok && m.symbols[i].defined
}

// Extern declares an undefined symbol. Its Section is -1 in the finished
// object, and a reference naming it survives for the linker to resolve.
func (m *Module) Extern(name string) {
	if m.guard("Extern") {
		return
	}
	if i, ok := m.symAt[name]; ok {
		if m.symbols[i].defined {
			// Declaring something extern that this module defines is a
			// contradiction rather than a duplicate, and the definition is
			// the one that is right.
			m.fail(&obj.Error{
				Arch: obj.ArchAMD64, Sentinel: obj.ErrDuplicate,
				Context: "Extern(" + name + ") names a symbol this module defines",
			})
		}
		return
	}
	m.symAt[name] = len(m.symbols)
	m.symbols = append(m.symbols, symbol{
		name:    name,
		sec:     -1,
		binding: Global,
	})
}

// Alias gives an existing symbol a second name at the same offset.
//
// It resolves at Finalize, so the order of this call and the Label it names
// does not matter.
func (m *Module) Alias(name, of string) {
	if m.guard("Alias") {
		return
	}
	m.aliases = append(m.aliases, aliasReq{name: name, of: of})
}

// SetVisibility sets a defined symbol's visibility. It resolves at
// Finalize, like Alias.
func (m *Module) SetVisibility(name string, v Visibility) {
	if m.guard("SetVisibility") {
		return
	}
	m.visReqs = append(m.visReqs, visReq{name: name, vis: v})
}

// guard reports whether a call should be a no-op, recording ErrFinalized if
// the module is spent.
func (m *Module) guard(what string) bool {
	if m.done {
		m.fail(&obj.Error{
			Arch: obj.ArchAMD64, Sentinel: obj.ErrFinalized,
			Context: what + " after Finalize",
		})
		return true
	}
	return m.err != nil
}

// Finalize patches same-section labels, closes symbol sizes, resolves
// aliases and visibility, verifies that every surviving reference names
// something, and returns the object.
//
// What comes back is immutable, pure data: safe to write more than once, in
// more than one format, from more than one goroutine.
//
// Finalize is idempotent. Call it again and you get the same object and the
// same error, not a second pass.
func (m *Module) Finalize() (*obj.Object, error) {
	if m.done {
		return m.final, m.finalErr
	}
	m.done = true

	if m.err != nil {
		m.finalErr = m.err
		return nil, m.finalErr
	}

	for _, step := range []func() error{
		m.patchLabels,
		m.closeSizes,
		m.resolveAliases,
		m.resolveVisibility,
		m.verifyRefs,
	} {
		if err := step(); err != nil {
			m.err, m.finalErr = err, err
			return nil, err
		}
	}

	m.final = m.build()
	return m.final, nil
}
