package operand

// Rip returns a RIP-relative address with no access width — the operand
// LeaR64M takes.
//
// It takes a SymRef and nothing else. There is no Rip(disp) overload,
// because a RIP-relative displacement to a numeric address is a claim about
// where the instruction will be linked, which is not knowable here and not
// something a caller should be able to spell.
func Rip(r SymRef) Addr { return Addr{a: ripRel(r)} }

func Rip8(r SymRef) M8     { return M8{a: ripRel(r)} }
func Rip16(r SymRef) M16   { return M16{a: ripRel(r)} }
func Rip32(r SymRef) M32   { return M32{a: ripRel(r)} }
func Rip64(r SymRef) M64   { return M64{a: ripRel(r)} }
func Rip128(r SymRef) M128 { return M128{a: ripRel(r)} }
func Rip256(r SymRef) M256 { return M256{a: ripRel(r)} }
func Rip512(r SymRef) M512 { return M512{a: ripRel(r)} }

func ripRel(r SymRef) addr {
	a := addr{rip: true}
	return a.symbol(r)
}
