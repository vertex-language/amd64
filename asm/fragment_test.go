package asm

import (
	"strings"
	"testing"

	"github.com/vertex-language/amd64"
)

// A fragment names a symbol the module around it already defines.
//
// The implicit extern is a guess about a name nothing defines — the rule
// that makes `call puts` work with no `.extern puts` in front of it — and it
// has to stop at the module's own definitions. An asm goto is where this
// shows up: the block it branches to is a label in the same function, and
// which of the two the emitter reaches first is a question about block
// layout that the assembler has no business depending on.
func TestFragmentDoesNotExternWhatTheModuleDefines(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before bool
	}{
		{"defined before the fragment", true},
		{"defined after it", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := amd64.NewModule()
			text := m.Section(amd64.Text)
			text.Label("f", amd64.Global, amd64.Func)

			define := func() {
				text.Label("target", amd64.Local)
				text.Ret()
			}
			if tc.before {
				define()
			}
			if err := AssembleFragment(text, "jmp target", Options{File: "t.s"}); err != nil {
				t.Fatalf("AssembleFragment: %v", err)
			}
			if !tc.before {
				define()
			}
			text.EndLabel("f")

			o, err := m.Finalize()
			if err != nil {
				t.Fatalf("Finalize: %v", err)
			}
			for _, sy := range o.Symbols() {
				if sy.Name == "target" && sy.Section < 0 {
					t.Error("target is undefined in the object; the module defines it")
				}
			}
		})
	}
}

// The rule it must not break: a name nothing in the module defines is still
// an import, which is what a call to a libc function relies on.
func TestFragmentStillExternsWhatNothingDefines(t *testing.T) {
	m := amd64.NewModule()
	text := m.Section(amd64.Text)
	text.Label("f", amd64.Global, amd64.Func)
	if err := AssembleFragment(text, "call puts", Options{File: "t.s"}); err != nil {
		t.Fatalf("AssembleFragment: %v", err)
	}
	text.Ret()
	text.EndLabel("f")

	o, err := m.Finalize()
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	var names []string
	for _, sy := range o.Symbols() {
		if sy.Section < 0 {
			names = append(names, sy.Name)
		}
	}
	if strings.Join(names, ",") != "puts" {
		t.Errorf("undefined symbols = %v, want just puts", names)
	}
}
