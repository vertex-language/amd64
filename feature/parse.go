package feature

import "strings"

// Parse resolves a feature specification against a starting set.
//
// Two shapes, and which one you get is decided by the first term:
//
//	Parse(Default(), "x86-64-v3+aes")   // exact: a level, then extensions
//	Parse(base, "+avx512vnni,-avx512dq") // adjust: applied left to right
//
// A leading level name replaces the set outright; everything after it
// adjusts. A leading sign adjusts the set handed in. Terms are separated by
// commas, by the signs themselves, or by both, and spelling is
// case-insensitive. Every term is applied in order, so "-avx,+avx" and
// "+avx,-avx" are different specifications and neither is an error.
//
// The set that comes back is closed under requirements, so a term can bring
// or take more than it names.
func Parse(base Set, spec string) (Set, error) {
	spec = strings.TrimSpace(strings.ToLower(spec))
	if spec == "" {
		return base, &ParseError{Reason: "empty feature specification"}
	}

	terms, err := split(spec)
	if err != nil {
		return base, err
	}

	s := base
	for i, t := range terms {
		// A bare first term must name a level and replaces the set.
		if i == 0 && t.sign == 0 {
			l, ok := ParseLevel(t.name)
			if !ok {
				if err := reject(t.name); err != nil {
					return base, err
				}
				// It named a real feature, not a level. Adding it to the
				// base is the only reading, but saying so is better than
				// guessing silently.
				return base, &ParseError{
					Spelling: t.name,
					Reason:   "names a feature, not a level; write +" + t.name + " to add it to the current set",
				}
			}
			s = NewSet(l)
			continue
		}

		if l, ok := ParseLevel(t.name); ok {
			if t.sign == '-' {
				return base, &ParseError{
					Spelling: t.name,
					Reason:   "a level cannot be removed; name the level you want instead",
				}
			}
			// "+x86-64-v3" raises the floor without discarding extensions.
			s = s.Add(l.Features()...)
			continue
		}

		f, ok := byName[t.name]
		if !ok {
			if err := reject(t.name); err != nil {
				return base, err
			}
			return base, &ParseError{Spelling: t.name, Reason: "unknown feature"}
		}

		switch t.sign {
		case '-':
			if table[f].level == V1 {
				return base, &ParseError{
					Spelling: t.name,
					Reason:   "part of the x86-64 baseline and cannot be removed",
				}
			}
			s = s.Remove(f)
		default:
			s = s.Add(f)
		}
	}

	return s, nil
}

// term is one element of a specification: a sign (0 for none) and a name.
type term struct {
	sign byte
	name string
}

func split(spec string) ([]term, error) {
	var (
		out  []term
		cur  strings.Builder
		sign byte
		open bool
	)

	flush := func() error {
		if !open {
			return nil
		}
		name := strings.TrimSpace(cur.String())
		if name == "" {
			return &ParseError{Reason: "empty term in feature specification"}
		}
		out = append(out, term{sign: sign, name: name})
		cur.Reset()
		sign = 0
		open = false
		return nil
	}

	for i := 0; i < len(spec); i++ {
		c := spec[i]
		switch c {
		case '+', '-':
			// A hyphen inside a name is not a separator, and deciding which
			// is which is longest match rather than a prefix test. Both
			// "x86-64" and "x86-64-v3" are spellings this table holds, so a
			// rule that split as soon as the accumulated text was complete
			// would read "x86-64-v3" as v1 minus a feature called v3 — and
			// would do the same to "aes-ni", whose prefix is also a feature.
			//
			// So the hyphen stays part of the name while extending it still
			// names something, and while the prefix is not a name yet. It
			// splits only when the prefix is complete and the extension is
			// not: "x86-64-v3-avx" is v3 with AVX removed.
			if c == '-' && open {
				acc := cur.String()
				if known(acc+"-"+peekName(spec, i+1)) || !known(acc) {
					cur.WriteByte(c)
					continue
				}
			}
			if err := flush(); err != nil {
				return nil, err
			}
			sign = c
			open = true
		case ',', ' ', '\t':
			if err := flush(); err != nil {
				return nil, err
			}
		default:
			open = true
			cur.WriteByte(c)
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, &ParseError{Reason: "empty feature specification"}
	}
	return out, nil
}

// known reports whether a spelling names something this table can resolve:
// a feature, a level, or a name it refuses with a reason of its own. The
// refused ones count, because "+amx-tile" should reach its explanation
// rather than being split into "amx" and a removal of "tile".
func known(s string) bool {
	if _, ok := byName[s]; ok {
		return true
	}
	if _, ok := rejected[s]; ok {
		return true
	}
	_, ok := ParseLevel(s)
	return ok
}

// peekName is the run of characters at i that would extend the term being
// accumulated: everything up to the next separator. It is the lookahead the
// longest-match rule in split needs.
func peekName(spec string, i int) string {
	for j := i; j < len(spec); j++ {
		switch spec[j] {
		case '+', '-', ',', ' ', '\t':
			return spec[i:j]
		}
	}
	return spec[i:]
}

// ParseError says which spelling failed and why. The distinction it exists
// to draw is between a typo and a name that means something real which this
// table does not hold — the second deserves a reason, not "unknown".
type ParseError struct {
	Spelling string
	Reason   string
}

func (e *ParseError) Error() string {
	if e.Spelling == "" {
		return "feature: " + e.Reason
	}
	return "feature: " + e.Spelling + ": " + e.Reason
}

// rejected holds spellings that name something real and are still not
// members of this table. Each gets the reason rather than being failed as a
// typo, because a caller who wrote "sse4a" knows what they meant and needs
// to be told why it is not what they got.
var rejected = map[string]string{
	"3dnow":  "removed from the architecture; no current AMD64 processor implements it",
	"3dnowa": "removed from the architecture; no current AMD64 processor implements it",
	"sse4a":  "an AMD extension, not a spelling of sse4.2, and not declared in this table",
	"xop":    "an AMD extension retired with Bulldozer, not declared in this table",
	"fma4":   "an AMD extension retired with Bulldozer; the Intel three-operand form is fma",
	"tbm":    "an AMD extension retired with Bulldozer, not declared in this table",
	"mpx":    "removed from the architecture; the registers and instructions are gone",

	"apx":      "declared by Intel but not in this table yet: REX2, the EVEX-promoted legacy forms and r16-r31 are not in reg or the ISA table",
	"egpr":     "part of APX, which is not in this table yet",
	"rex2":     "part of APX, which is not in this table yet",
	"avx10":    "declared by Intel but not in this table yet; no EVEX row is declared",
	"avx10.1":  "declared by Intel but not in this table yet; no EVEX row is declared",
	"avx10.2":  "declared by Intel but not in this table yet; no EVEX row is declared",
	"amx-tile": "the tile registers are in reg but no AMX row is declared in the ISA table yet",
	"amx-int8": "the tile registers are in reg but no AMX row is declared in the ISA table yet",
	"amx-bf16": "the tile registers are in reg but no AMX row is declared in the ISA table yet",

	// The i386 module's ladder. These are levels over a 32-bit baseline and
	// have no 64-bit member; naming one here is a target confusion rather
	// than a typo.
	"i386":       "an i386 level; it is defined over a 32-bit baseline and has no AMD64 member",
	"i486":       "an i386 level; it is defined over a 32-bit baseline and has no AMD64 member",
	"i586":       "an i386 level; it is defined over a 32-bit baseline and has no AMD64 member",
	"i686":       "an i386 level; it is defined over a 32-bit baseline and has no AMD64 member",
	"pentium":    "an i386 level; it is defined over a 32-bit baseline and has no AMD64 member",
	"pentiumpro": "an i386 level; it is defined over a 32-bit baseline and has no AMD64 member",
}

func reject(name string) error {
	if reason, ok := rejected[name]; ok {
		return &ParseError{Spelling: name, Reason: reason}
	}
	return nil
}
