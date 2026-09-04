// Package feature declares the AMD64 feature vocabulary: the x86-64-v1
// through v4 microarchitecture levels, the orthogonal extensions that sit
// outside them, and the requirement graph that keeps a set describing real
// silicon.
//
// The package stands alone. It imports nothing from this tree.
//
// Two different things live here and are deliberately not flattened together.
// A Level is a point on the cumulative ladder the x86-64 psABI defines, and
// it is a bundle rather than an instruction group. A Feature is an extension
// with a CPUID bit of its own that no level requires — the psABI's own rule
// is that extensions not concerned with general-purpose computation stay out
// of the levels, which is why AES and RDRAND are features and SSE4.2 is not.
//
// This package holds no CPUID leaf or bit numbers. It describes what a
// module may emit, not what a processor reports; a runtime detector is a
// different program with a different table.
package feature

// Feature names one extension.
type Feature uint8

const (
	// x86-64-v1, the baseline. Always present, never removable.
	CMOV Feature = iota
	CX8
	FPU
	FXSR
	MMX
	OSFXSR
	SCE
	SSE
	SSE2

	// x86-64-v2.
	CX16
	LAHFSAHF
	POPCNT
	SSE3
	SSSE3
	SSE41
	SSE42

	// x86-64-v3.
	AVX
	AVX2
	BMI1
	BMI2
	F16C
	FMA
	LZCNT
	MOVBE
	XSAVE

	// x86-64-v4.
	AVX512F
	AVX512BW
	AVX512CD
	AVX512DQ
	AVX512VL

	// Orthogonal extensions. No level requires any of these.
	AES
	PCLMULQDQ
	SHA
	RDRAND
	RDSEED
	ADX
	GFNI
	VAES
	VPCLMULQDQ
	AVX512IFMA
	AVX512VBMI
	AVX512VBMI2
	AVX512VNNI
	AVX512BITALG
	AVX512VPOPCNTDQ
	AVX512VP2INTERSECT
	AVX512BF16
	AVX512FP16
	AVX5124VNNIW
	AVX5124FMAPS

	numFeatures
)

// info is one row of the table. The canonical spelling is what String
// prints and what Set.String composes; aliases are accepted by Parse and
// never printed.
type info struct {
	name     string
	aliases  []string
	level    Level // the level that requires it, or LevelNone if orthogonal
	requires []Feature
}

var table = [numFeatures]info{
	// ---- x86-64-v1 -------------------------------------------------------
	CMOV:   {name: "cmov", level: V1},
	CX8:    {name: "cx8", aliases: []string{"cmpxchg8b"}, level: V1},
	FPU:    {name: "fpu", aliases: []string{"x87"}, level: V1},
	FXSR:   {name: "fxsr", level: V1},
	MMX:    {name: "mmx", level: V1},
	OSFXSR: {name: "osfxsr", level: V1, requires: []Feature{FXSR}},
	SCE:    {name: "sce", aliases: []string{"syscall"}, level: V1},
	SSE:    {name: "sse", level: V1},
	SSE2:   {name: "sse2", level: V1, requires: []Feature{SSE}},

	// ---- x86-64-v2 -------------------------------------------------------
	CX16:     {name: "cmpxchg16b", aliases: []string{"cx16"}, level: V2},
	LAHFSAHF: {name: "lahf-sahf", aliases: []string{"lahfsahf", "lahf_lm"}, level: V2},
	POPCNT:   {name: "popcnt", level: V2},
	SSE3:     {name: "sse3", level: V2, requires: []Feature{SSE2}},
	SSSE3:    {name: "ssse3", level: V2, requires: []Feature{SSE3}},
	SSE41:    {name: "sse4.1", aliases: []string{"sse4_1", "sse41"}, level: V2, requires: []Feature{SSSE3}},
	SSE42:    {name: "sse4.2", aliases: []string{"sse4_2", "sse42"}, level: V2, requires: []Feature{SSE41}},

	// ---- x86-64-v3 -------------------------------------------------------
	// XSAVE is the psABI's OSXSAVE: the level's requirement is that the OS
	// has enabled the state, not merely that the instruction decodes. Both
	// spellings parse; this one prints.
	XSAVE: {name: "xsave", aliases: []string{"osxsave"}, level: V3},
	AVX:   {name: "avx", level: V3, requires: []Feature{SSE42, XSAVE}},
	AVX2:  {name: "avx2", level: V3, requires: []Feature{AVX}},
	BMI1:  {name: "bmi", aliases: []string{"bmi1"}, level: V3},
	BMI2:  {name: "bmi2", level: V3, requires: []Feature{BMI1}},
	F16C:  {name: "f16c", level: V3, requires: []Feature{AVX}},
	FMA:   {name: "fma", level: V3, requires: []Feature{AVX}},
	LZCNT: {name: "lzcnt", aliases: []string{"abm"}, level: V3},
	MOVBE: {name: "movbe", level: V3},

	// ---- x86-64-v4 -------------------------------------------------------
	AVX512F:  {name: "avx512f", level: V4, requires: []Feature{AVX2}},
	AVX512BW: {name: "avx512bw", level: V4, requires: []Feature{AVX512F}},
	AVX512CD: {name: "avx512cd", level: V4, requires: []Feature{AVX512F}},
	AVX512DQ: {name: "avx512dq", level: V4, requires: []Feature{AVX512F}},
	AVX512VL: {name: "avx512vl", level: V4, requires: []Feature{AVX512F}},

	// ---- orthogonal ------------------------------------------------------
	AES:        {name: "aes", aliases: []string{"aesni", "aes-ni"}, requires: []Feature{SSE2}},
	PCLMULQDQ:  {name: "pclmulqdq", aliases: []string{"pclmul"}, requires: []Feature{SSE2}},
	SHA:        {name: "sha", aliases: []string{"sha_ni", "sha-ni"}, requires: []Feature{SSE2}},
	RDRAND:     {name: "rdrand", aliases: []string{"rdrnd"}},
	RDSEED:     {name: "rdseed"},
	ADX:        {name: "adx"},
	GFNI:       {name: "gfni", requires: []Feature{SSE2}},
	VAES:       {name: "vaes", requires: []Feature{AES, AVX}},
	VPCLMULQDQ: {name: "vpclmulqdq", requires: []Feature{PCLMULQDQ, AVX}},

	AVX512IFMA:         {name: "avx512ifma", requires: []Feature{AVX512F}},
	AVX512VBMI:         {name: "avx512vbmi", requires: []Feature{AVX512BW}},
	AVX512VBMI2:        {name: "avx512vbmi2", requires: []Feature{AVX512BW}},
	AVX512VNNI:         {name: "avx512vnni", requires: []Feature{AVX512F}},
	AVX512BITALG:       {name: "avx512bitalg", requires: []Feature{AVX512BW}},
	AVX512VPOPCNTDQ:    {name: "avx512vpopcntdq", requires: []Feature{AVX512F}},
	AVX512VP2INTERSECT: {name: "avx512vp2intersect", requires: []Feature{AVX512F}},
	AVX512BF16:         {name: "avx512bf16", requires: []Feature{AVX512BW, AVX512VL}},
	AVX512FP16:         {name: "avx512fp16", requires: []Feature{AVX512BW, AVX512DQ, AVX512VL}},
	AVX5124VNNIW:       {name: "avx5124vnniw", requires: []Feature{AVX512F}},
	AVX5124FMAPS:       {name: "avx5124fmaps", requires: []Feature{AVX512F}},
}

func (f Feature) String() string {
	if f.Valid() {
		return table[f].name
	}
	return "feature?"
}

// Valid reports whether the value names a declared feature.
func (f Feature) Valid() bool { return int(f) < int(numFeatures) }

// Level returns the level that requires the feature, or LevelNone if no
// level does. It is the answer to "is this a bundle member or an extension".
func (f Feature) Level() Level {
	if f.Valid() {
		return table[f].level
	}
	return LevelNone
}

// All returns every declared feature in declaration order, which is the
// order Set.String prints them in.
func All() []Feature {
	out := make([]Feature, 0, numFeatures)
	for f := Feature(0); f < numFeatures; f++ {
		out = append(out, f)
	}
	return out
}

// Requires returns the features the given one directly requires. It is not
// the transitive closure; walk it if you want that, or add the feature to a
// Set and read back what arrived.
func Requires(f Feature) []Feature {
	if !f.Valid() {
		return nil
	}
	out := make([]Feature, len(table[f].requires))
	copy(out, table[f].requires)
	return out
}

// RequiredBy returns the features that directly require the given one — the
// reverse edge, which is what Set.Remove walks. Also not transitive.
func RequiredBy(f Feature) []Feature {
	if !f.Valid() {
		return nil
	}
	out := make([]Feature, len(reverse[f]))
	copy(out, reverse[f])
	return out
}

// reverse is the requirement graph inverted, built once from table.
var reverse = buildReverse()

func buildReverse() [numFeatures][]Feature {
	var r [numFeatures][]Feature
	for f := Feature(0); f < numFeatures; f++ {
		for _, req := range table[f].requires {
			r[req] = append(r[req], f)
		}
	}
	return r
}

// byName resolves every canonical spelling and every alias. Building it is
// also this package's integrity check: a spelling claimed twice would make
// Parse answer arbitrarily, and that is a question about this table rather
// than about anything a caller did, so it fails at program start.
var byName = buildIndex()

func buildIndex() map[string]Feature {
	m := make(map[string]Feature, numFeatures*2)
	add := func(name string, f Feature) {
		if name == "" {
			panic("feature: empty spelling for " + f.String())
		}
		if _, dup := m[name]; dup {
			panic("feature: duplicate spelling: " + name)
		}
		m[name] = f
	}
	for f := Feature(0); f < numFeatures; f++ {
		add(table[f].name, f)
		for _, a := range table[f].aliases {
			add(a, f)
		}
	}
	return m
}
