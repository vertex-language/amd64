// Package obj is the vocabulary the rest of this tree spells its types in,
// and the finished artifact the builder hands to a writer.
//
// It knows no instruction set and imports nothing from this tree. That is
// what puts it at the bottom of the import graph beside reg: operand needs
// RefKind and Error, internal/encode needs both, the root re-exports
// everything, and the three writers under obj/ take an *Object. A package
// this far down cannot import anything above it, and nothing here wants to.
//
// Everything in here is a constant, a plain data struct, or an accessor over
// one. Nothing acts. Past Finalize the artifact is inert: data with no
// methods that do anything but read, safe to write more than once, in more
// than one format, from more than one goroutine.
package obj

// Arch names the architecture an object was built for.
//
// It is a field of the object rather than of a writer's options because
// everything a container's header needs follows from it. A caller who could
// pass a different one could produce a file whose header disagrees with its
// bytes, which is why no writer here has a Target field.
//
// There are two tenants because there are two builders. The writers derive
// their target from this and always did.
type Arch uint8

const (
	ArchNone Arch = iota
	ArchI386
	ArchAMD64
)

var archNames = [...]string{
	ArchNone:  "",
	ArchI386:  "i386",
	ArchAMD64: "amd64",
}

// String is what a diagnostic leads with: "amd64 .text+0x11: ...".
func (a Arch) String() string {
	if int(a) < len(archNames) {
		return archNames[a]
	}
	return "arch?"
}

// Valid reports whether the value names a declared architecture. ArchNone is
// not one and reports false.
func (a Arch) Valid() bool { return a == ArchI386 || a == ArchAMD64 }

// Bits is the architecture's pointer width, or 0 for ArchNone.
func (a Arch) Bits() int {
	switch a {
	case ArchI386:
		return 32
	case ArchAMD64:
		return 64
	}
	return 0
}
