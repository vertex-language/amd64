package amd64

import "github.com/vertex-language/amd64/obj"

// SectionKind is how a section behaves at load time. It is obj's, so a
// lowering that never imports obj still spells it and a downstream consumer
// never converts: amd64.Text and obj.Text are the same value.
type SectionKind = obj.SectionKind

const (
	Text   = obj.Text   // ".text"
	Data   = obj.Data   // ".data"
	ROData = obj.ROData // ".rodata"
	BSS    = obj.BSS    // ".bss"

	// RelROData is read-only data a loader has to relocate first: a
	// table of pointers to other symbols. See obj's declaration.
	RelROData = obj.RelROData // ".data.rel.ro"
)
