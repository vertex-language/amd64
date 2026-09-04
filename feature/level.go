package feature

// Level is a point on the cumulative x86-64-v1..v4 ladder.
//
// The levels are a naming convention over feature sets that AMD, Intel and
// the distributions agreed on in 2020 — not a CPU ladder. Every feature in
// every level has a CPUID bit, which is what makes them different in kind
// from the i386 module's i486..i686 rungs, where the early ones predate
// CPUID entirely.
type Level uint8

const (
	// LevelNone is the zero value and names no level. A Feature reports it
	// when no level requires the feature.
	LevelNone Level = iota
	V1
	V2
	V3
	V4
)

var levelNames = [...]string{
	LevelNone: "",
	V1:        "x86-64-v1",
	V2:        "x86-64-v2",
	V3:        "x86-64-v3",
	V4:        "x86-64-v4",
}

// levelAliases are the other spellings the world writes. GCC and Clang
// accept plain "x86-64" for the baseline, and it is the one people type.
var levelAliases = map[string]Level{
	"x86-64":    V1,
	"x86_64":    V1,
	"baseline":  V1,
	"x86-64-v1": V1,
	"x86_64_v2": V2,
	"x86_64_v3": V3,
	"x86_64_v4": V4,
}

func (l Level) String() string {
	if int(l) < len(levelNames) {
		return levelNames[l]
	}
	return "level?"
}

// Valid reports whether the value names a declared level. LevelNone is not
// a level and reports false.
func (l Level) Valid() bool { return l >= V1 && l <= V4 }

// Levels returns the ladder in order, lowest first.
func Levels() []Level { return []Level{V1, V2, V3, V4} }

// Adds returns exactly the features this rung introduces — not the ones the
// rungs below it already required. It is here so a driver can print the
// level rather than paraphrase it.
func (l Level) Adds() []Feature {
	if !l.Valid() {
		return nil
	}
	out := make([]Feature, 0, 16)
	for f := Feature(0); f < numFeatures; f++ {
		if table[f].level == l {
			out = append(out, f)
		}
	}
	return out
}

// Features returns everything the level requires, cumulatively: its own
// rung and every rung below it.
func (l Level) Features() []Feature {
	if !l.Valid() {
		return nil
	}
	out := make([]Feature, 0, 32)
	for f := Feature(0); f < numFeatures; f++ {
		lv := table[f].level
		if lv.Valid() && lv <= l {
			out = append(out, f)
		}
	}
	return out
}

// set is the cumulative feature set of the level, as bits.
func (l Level) set() Set {
	var s Set
	for _, f := range l.Features() {
		s.bits[f>>6] |= 1 << (f & 63)
	}
	return s
}

// levelSets is the four cumulative sets, computed once. Set.Level compares
// against these rather than walking the table on every call.
var levelSets = [...]Set{
	V1: V1.set(),
	V2: V2.set(),
	V3: V3.set(),
	V4: V4.set(),
}

// ParseLevel resolves a level spelling. It accepts the canonical names, the
// underscore variants, and plain "x86-64" for the baseline.
func ParseLevel(s string) (Level, bool) {
	for l := V1; l <= V4; l++ {
		if levelNames[l] == s {
			return l, true
		}
	}
	l, ok := levelAliases[s]
	return l, ok
}
