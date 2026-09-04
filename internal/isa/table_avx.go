package isa

func buildAVX() {
	// FMA3 scalar float instructions
	// VFMADD231SS/SD xmm, xmm, xmm/m
	add(Form{Mnemonic: "vfmadd231ss", Helper: "Vfmadd231ssXmmXmmXM32",
		Ops: []Class{Xmm, Xmm, XM32},
		Enc: Enc{Family: VEX, Pfx: 0x66, Map: Map0F38, Op: 0xb9, Ext: SlashR, W: 0, L: 0}})
	add(Form{Mnemonic: "vfmadd231sd", Helper: "Vfmadd231sdXmmXmmXM64",
		Ops: []Class{Xmm, Xmm, XM64},
		Enc: Enc{Family: VEX, Pfx: 0x66, Map: Map0F38, Op: 0xb9, Ext: SlashR, W: 1, L: 0}})
}
