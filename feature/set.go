package feature

import "strings"

// Set is a feature set: what a module may emit. It is a value, comparable
// with ==, and every operation returns a new one.
//
// A Set is always closed under requirements in both directions. Adding AVX2
// brings AVX, and everything AVX brings. Removing AVX drops AVX2, F16C and
// FMA, because a set holding AVX2 but not AVX describes no silicon and
// gating against it would produce diagnostics no processor could falsify.
//
// The baseline is always present. Nothing removes it.
type Set struct {
	bits [2]uint64
}

// NewSet returns the cumulative feature set of a level, with no extensions.
// An invalid level yields the baseline.
func NewSet(l Level) Set {
	if !l.Valid() {
		l = V1
	}
	return levelSets[l]
}

// Default is x86-64-v1 with no extensions: the baseline every AMD64
// processor implements, and what a Module gets when nobody says otherwise.
func Default() Set { return levelSets[V1] }

// Has reports whether the set contains the feature.
func (s Set) Has(f Feature) bool {
	if !f.Valid() {
		return false
	}
	return s.bits[f>>6]&(1<<(f&63)) != 0
}

// Add returns the set with the feature and everything it requires, added
// transitively.
func (s Set) Add(fs ...Feature) Set {
	for _, f := range fs {
		if !f.Valid() {
			continue
		}
		s = s.addClosed(f)
	}
	return s
}

func (s Set) addClosed(f Feature) Set {
	if s.Has(f) {
		return s
	}
	s.bits[f>>6] |= 1 << (f & 63)
	for _, req := range table[f].requires {
		s = s.addClosed(req)
	}
	return s
}

// Remove returns the set without the feature and without everything that
// requires it, transitively.
//
// Baseline features are not removable and Remove leaves them in place. That
// is silent here because a Set is a value being built; Parse reports the
// same attempt as an error, because there someone wrote it down.
func (s Set) Remove(fs ...Feature) Set {
	for _, f := range fs {
		if !f.Valid() || table[f].level == V1 {
			continue
		}
		s = s.removeClosed(f)
	}
	return s
}

func (s Set) removeClosed(f Feature) Set {
	if !s.Has(f) {
		return s
	}
	s.bits[f>>6] &^= 1 << (f & 63)
	for _, dep := range reverse[f] {
		s = s.removeClosed(dep)
	}
	return s
}

// Level is the highest rung the set fully satisfies. Removing a feature a
// level requires demotes it: there is no way to hold a set that reports V3
// and does not hold AVX.
func (s Set) Level() Level {
	best := V1
	for l := V1; l <= V4; l++ {
		if s.contains(levelSets[l]) {
			best = l
		}
	}
	return best
}

func (s Set) contains(o Set) bool {
	return s.bits[0]&o.bits[0] == o.bits[0] && s.bits[1]&o.bits[1] == o.bits[1]
}

// Features returns every feature in the set, in declaration order.
func (s Set) Features() []Feature {
	out := make([]Feature, 0, numFeatures)
	for f := Feature(0); f < numFeatures; f++ {
		if s.Has(f) {
			out = append(out, f)
		}
	}
	return out
}

// Extensions returns the features the set holds beyond its level — exactly
// what String prints after the level name.
func (s Set) Extensions() []Feature {
	lv := levelSets[s.Level()]
	out := make([]Feature, 0, 8)
	for f := Feature(0); f < numFeatures; f++ {
		if s.Has(f) && !lv.Has(f) {
			out = append(out, f)
		}
	}
	return out
}

// Equal reports whether two sets hold the same features.
func (s Set) Equal(o Set) bool { return s == o }

// String is the canonical spelling: the highest level the set satisfies,
// then every feature past it, in declaration order.
//
//	x86-64-v3+aes+sha
//
// Parse accepts what String prints, and the two round-trip. A set that lost
// a level's feature prints as the lower level plus whatever survived, which
// is the same thing said honestly.
func (s Set) String() string {
	var b strings.Builder
	b.WriteString(s.Level().String())
	for _, f := range s.Extensions() {
		b.WriteByte('+')
		b.WriteString(f.String())
	}
	return b.String()
}
