package isa

// buildLock declares the memory-ordering tranche: the locking clone of
// every row that admits LOCK, and the three fences.
//
// The clones are generated from the table rather than written out,
// because that is what they are. A locking form is its base form with
// one prefix byte in front — same opcode, same operands, same encoding
// — so writing them by hand would be writing the same forty rows twice
// and inviting the two copies to disagree. Lockable is the base row's
// own answer to whether LOCK is architecturally legal on it, which is a
// fact about the instruction and belongs where the instruction is
// declared.
//
// A clone is not reachable from Emit. LOCK is not part of a mnemonic —
// it is a prefix, and "lock add" is not an instruction name the way
// "add" is — so a caller reaches one by naming its typed helper, and
// Resolve skips every row whose mnemonic starts with "lock ". That is
// also what keeps Emit's shortest-wins rule from ever picking a locking
// form for an unlocked "add".
//
// This has to run after every tranche that declares a lockable row.
// forms.go's init is where that order is stated.

func buildLock() {
	// Range over a copy of the length, not the slice: add appends to
	// forms as this loop runs, and a clone is never itself lockable.
	n := len(forms)
	for i := 0; i < n; i++ {
		f := forms[i]
		if !f.Lockable {
			continue
		}
		enc := f.Enc
		enc.Lock = true
		add(Form{
			Mnemonic: "lock " + f.Mnemonic,
			Helper:   "Lock" + f.Helper,
			Ops:      f.Ops,
			Enc:      enc,
			Gate:     f.Gate,
			AliasOf:  f.AliasOf,
		})
	}

	buildFence()
}

// buildFence declares the three fences.
//
// They are here rather than with the ordinary integer rows because they
// are the other half of what LOCK is for: LOCK orders one access against
// everything, and a fence orders everything before it against everything
// after. A module that needs one needs the other.
//
// All three are 0F AE with a ModRM byte that addresses nothing — MFENCE
// is 0F AE F0, and the F0 is not naming a register. FixedModRM is how a
// row says that; every other row in this table computes the byte from
// its operands.
func buildFence() {
	add(Form{Mnemonic: "lfence", Helper: "Lfence", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xe8}})
	add(Form{Mnemonic: "mfence", Helper: "Mfence", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xf0}})
	add(Form{Mnemonic: "sfence", Helper: "Sfence", Ops: nil,
		Enc: Enc{Map: Map0F, Op: 0xae, Ext: SlashR, W: WAny, L: LNone,
			HasFixedModRM: true, FixedModRM: 0xf8}})
}
