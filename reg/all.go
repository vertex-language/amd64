package reg

// All returns every declared register, grouped by class in encoding order.
// It exists so a driver can print a register list, and so the integrity check
// below has something to walk. The returned slice is fresh on every call.
func All() []Value {
	out := make([]Value, len(all))
	copy(out, all)
	return out
}

// Lookup resolves a register's canonical lowercase spelling. Aliases are not
// accepted, because this package declares no aliases: r8 is the 64-bit
// register and r8b is the byte one, and there is no third name for either.
func Lookup(name string) (Value, bool) {
	v, ok := byName[name]
	return v, ok
}

var (
	all    = buildAll()
	byName = buildIndex(all)
)

func buildAll() []Value {
	v := make([]Value, 0, 200)

	for r := AL; r <= R15B; r++ {
		v = append(v, r)
	}
	for r := AH; r <= BH; r++ {
		v = append(v, r)
	}
	for r := AX; r <= R15W; r++ {
		v = append(v, r)
	}
	for r := EAX; r <= R15D; r++ {
		v = append(v, r)
	}
	for r := RAX; r <= R15Q; r++ {
		v = append(v, r)
	}
	for r := ES; r <= GS; r++ {
		v = append(v, r)
	}
	for r := ST0; r <= ST7; r++ {
		v = append(v, r)
	}
	for r := MM0; r <= MM7; r++ {
		v = append(v, r)
	}
	for r := XMM0; r <= XMM31; r++ {
		v = append(v, r)
	}
	for r := YMM0; r <= YMM31; r++ {
		v = append(v, r)
	}
	for r := ZMM0; r <= ZMM31; r++ {
		v = append(v, r)
	}
	for r := K0; r <= K7; r++ {
		v = append(v, r)
	}
	for r := CR0; r <= CR15; r++ {
		v = append(v, r)
	}
	for r := DR0; r <= DR15; r++ {
		v = append(v, r)
	}
	for r := TMM0; r <= TMM7; r++ {
		v = append(v, r)
	}

	return v
}

// buildIndex is also the integrity check. Two registers sharing a spelling
// would make Lookup answer one of them arbitrarily and make a diagnostic name
// the wrong register, and it is a question about this package's own data
// rather than about anything a caller did — so it fails at program start,
// whoever the caller is and whatever they were about to do.
func buildIndex(v []Value) map[string]Value {
	m := make(map[string]Value, len(v))
	for _, r := range v {
		n := r.String()
		if n == "" {
			panic("reg: register with an empty spelling in " + r.Kind().String())
		}
		if !r.Valid() {
			panic("reg: invalid register " + n + " in the declared set")
		}
		if _, dup := m[n]; dup {
			panic("reg: duplicate register spelling: " + n)
		}
		m[n] = r
	}
	return m
}
