# amd64

The AMD64 instruction builder: registers, the ISA table, the encoder, and an in-memory object you can hand straight to an ELF, COFF, or Mach-O writer.

```go
m := amd64.NewModule()

t := m.Section(amd64.Text)
t.Label("main", amd64.Global, amd64.Func)
t.SubRM64Imm8(amd64.RSP, 8)                           // 48 83 ec 08 — decided now
t.LeaR64M(amd64.RDI, amd64.Rip(amd64.Ref("msg", amd64.RefPC32)))
t.CallRef(amd64.Ref("puts", amd64.RefPLT32))          // e8 + 4-byte hole + reference
t.XorR32RM32(amd64.EAX, amd64.EAX)                    // 31 c0 — zeroes all of RAX
t.AddRM64Imm8(amd64.RSP, 8)
t.Ret()
t.EndLabel("main")

m.Extern("puts")

o, err := m.Finalize()
if err != nil {
    log.Fatal(err)
}
elf.Write(f, o)
```

## Install

```sh
go get github.com/vertex-language/amd64
```

Container writers are leaf packages, so a build only pays for the format it emits:

```go
import "github.com/vertex-language/amd64/obj/elf"   // *obj.Object → ELF ET_REL
import "github.com/vertex-language/amd64/obj/pe"    // *obj.Object → COFF .obj
import "github.com/vertex-language/amd64/obj/macho" // *obj.Object → Mach-O MH_OBJECT
```

The three writers are the next tranche and are not in the tree yet. The builder,
the encoder and `obj` are, and [Writers](#writers) is the contract the three are
being built to. Everything else in this file describes code that exists.

## Targets

One architecture in, three containers out:

| | Container | Writer | Links in-tree? |
|---|---|---|---|
| **Linux, BSD, and other SysV** | ELF `ET_REL`, `ELFCLASS64`, `EM_X86_64` | `amd64/obj/elf` | Yes — `elf/x86_64` |
| **Windows** | COFF `.obj`, `IMAGE_FILE_MACHINE_AMD64` | `amd64/obj/pe` | Yes — `pe/x64` |
| **Apple** | Mach-O `MH_OBJECT`, `CPU_TYPE_X86_64` | `amd64/obj/macho` | Yes — `macho/x86_64` |

The right-hand column is why this module reads differently from its i386 sibling. There, every row said *not yet*, because the sibling linkers had registered only 64-bit backends and i386 was the odd size out. This is that size. `elf/link`, `pe/link` and `macho/link` each resolve their machine through a registered backend, and for this architecture the backend already exists and already ships: blank-import `elf/x86_64`, `pe/x64`, or `macho/x86_64` and the object this module produces links in-tree, end to end, without leaving the module family.

It also links out of tree, which is the more important claim and the older one. `ld`, `lld`, `link.exe /MACHINE:X64` and `ld64` all take an object out of these writers, because an object is an object and not a dialect. In-tree linking is a convenience; container correctness is the contract.

This module never learns what a linker is, and the fact that three linkers next door accept its output does not change that. Nothing here imports `link`, `backend`, or `image` from any of them.

## Contents

- [Targets](#targets)
- [Package map](#package-map)
- [Where `obj` lives](#where-obj-lives)
- [Quick start](#quick-start)
  - [Build a module](#build-a-module)
  - [Emit an object](#emit-an-object)
  - [Read the finished object](#read-the-finished-object)
  - [Lower from an IR](#lower-from-an-ir)
- [Instructions](#instructions)
- [Registers and widths](#registers-and-widths)
- [Sections and data](#sections-and-data)
- [Memory operands](#memory-operands)
- [Symbols](#symbols)
- [References](#references)
- [Features](#features)
- [Errors](#errors)
- [Writers](#writers)
- [How it's put together](#how-its-put-together)
- [Known limitations](#known-limitations)
- [License](#license)

## Package map

| Package | Purpose |
|---|---|
| `amd64` | The builder. `Module`, `Section`, typed helpers, `Emit`. This is what a frontend imports. |
| `amd64/reg` | `R8`, `R16`, `R32`, `R64`, `Sreg`, `St`, `Mm`, `Xmm`, `Ymm`, `Zmm`, `K`, `Cr`, `Dr`, `Tmm`, and `RIP`. Imports nothing. |
| `amd64/operand` | `Imm`, `M8`..`M512`, `Label`, `SymRef`, the memory constructors. Imports `reg` and `obj`. |
| `amd64/feature` | The `x86-64-v1`..`v4` ladder, the orthogonal extensions, `Parse`. Stands alone. |
| `amd64/obj` | The vocabulary and the finished artifact: `Object`, `Section`, `Symbol`, `Reference`, `RefKind`, `SectionKind`, `Arch`, `Error`. Imports nothing. |
| `amd64/obj/elf` | `*obj.Object` → ELF relocatable object, via `github.com/vertex-language/elf`. |
| `amd64/obj/pe` | `*obj.Object` → COFF relocatable object, via `github.com/vertex-language/pe`. |
| `amd64/obj/macho` | `*obj.Object` → Mach-O relocatable object, via `github.com/vertex-language/macho`. |
| `amd64/internal/isa` | The form table: classes, opcodes, encodings, gates, `Resolve`. |
| `amd64/internal/encode` | form + operands → bytes + fixups; REX/VEX/EVEX emission; `Nops` for `Align`. |

Two directions, no cycles. `reg` imports nothing. `operand` sits on `reg` and on `amd64/obj`: it needs `reg` for the registers an address is built from, and `obj` for `RefKind`, because a reference's link semantics are a fact about the artifact and belong in the artifact's vocabulary. **`RefKind` is declared once, in `obj`, and aliased upward.** `amd64.RefPLT32`, `operand.RefPLT32` and `obj.RefPLT32` are the same constant, and no conversion exists anywhere in the tree.

`isa` and `encode` are `internal/` because they are implementation: the typed helpers and `Emit` are the instruction surface, and nothing a caller writes holds an `isa` or `encode` type. Encoder failures still reach you — through `*obj.Error` as a sentinel, a message, and notes. Data, not internal types.

### File layout

Each public package is one concept per file, and the instruction surface is split by tranche so `inst_avx.go` and `isa/table_avx.go`'s `buildAVX` sit opposite each other:

```
amd64/
  module.go section.go section_kind.go symbol.go error.go ref.go
  alias.go alias_vec.go build.go util.go emit.go inst.go
  inst_alu.go inst_mov.go inst_stack.go inst_arith.go inst_shift.go
  inst_branch.go inst_cond.go inst_misc.go inst_sse.go inst_lock.go
  reg/       reg.go value.go gpr.go seg.go x87.go vec.go mask.go sys.go
             rm.go all.go
  operand/   operand.go imm.go label.go mem.go rip.go widths.go symref.go
             parts.go
  feature/   feature.go level.go set.go parse.go
  obj/       arch.go object.go section_kind.go symbol.go ref.go error.go
  internal/
    isa/     class.go form.go forms.go enc.go table_base.go table_sse.go
             table_lock.go
    encode/  encode.go rex.go vex.go evex.go modrm.go sib.go nop.go error.go
```

Not landed yet, and named here because the shape is decided: `inst_bit.go`,
`inst_system.go` and `inst_avx.go` opposite `isa/table_avx.go`; and the
three writers under `obj/`, each `write.go section.go symbol.go reloc.go`.

The three writers carry the same four filenames in the same order, so the trio of `reloc.go` files sit opposite each other. Between them they hold every fact this tree knows about relocation numbering.

## Where `obj` lives

`obj` is a package of this module, at `github.com/vertex-language/amd64/obj`. It
imports nothing — not `reg`, not `operand`, nothing outside the standard library
— which is what puts it at the bottom of the import graph beside `reg`, and what
lets the three writers sit *under* it as `obj/elf`, `obj/pe` and `obj/macho`
without a cycle. The i386 module has the same shape at `i386/obj`, and the two
are siblings rather than one shared dependency: there is no
`github.com/vertex-language/obj` module and nothing here imports one.

That is a duplicated vocabulary, and it is duplicated on purpose for now. Lifting
it into a module both builders import is a real option and the field set was
chosen so it stays open — but a shared module is a shared release cadence, and
neither builder is finished enough to want one. Two things make the lift cheap
whenever it is worth doing.

**`Arch` was in the struct set from the first day.** i386's writers derived their
target from `ArchI386` rather than hardcoding it, and this one's `Arch` is
`ArchI386` or `ArchAMD64` over the same code path. Nothing has to learn a second
way to ask what an object was built for.

**`Reference` is wider here, and that is not a walk-back.** i386's version said
`Size` tops out at 4 and `Addend` is an `int32`, because no field in that
architecture is wider than a doubleword and a struct that cannot describe an i386
relocation is not more general but wrong. Here `Size` is 1, 2, 4 or 8 and
`Addend` is an `int64`, because `R_X86_64_64` against a quadword is an ordinary
thing to want. The rule that made the narrow struct right is unchanged: the
*architecture* validates the width, not the struct. An 8 arriving at a writer on
an `ArchI386` object is `ErrRange` naming the field. A wider alphabet, not a
looser grammar.

## Quick start

### Build a module

```go
m := amd64.NewModule()

t := m.Section(amd64.Text)
t.Align(16)
t.Label("main", amd64.Global, amd64.Func)
t.PushR64(amd64.RBX)
t.LeaR64M(amd64.RDI, amd64.Rip(amd64.Ref("msg", amd64.RefPC32)))
t.CallRef(amd64.Ref("puts", amd64.RefPLT32))
t.XorR32RM32(amd64.EAX, amd64.EAX)
t.PopR64(amd64.RBX)
t.Ret()
t.EndLabel("main")

r := m.Section(amd64.ROData)
r.Label("msg", amd64.Local, amd64.ObjectSym)
r.Asciz("hello\n")

m.Extern("puts")

o, err := m.Finalize()
if err != nil {
    log.Fatal(err)
}
```

`Finalize` patches same-section labels, closes symbol sizes, resolves aliases and visibility, verifies that every surviving reference names something, and returns an `*obj.Object` — immutable, pure data, safe to write more than once, in more than one format, from more than one goroutine.

The builder is spent afterwards. Every call on `m` is a no-op recording `ErrFinalized`, and asking for a section that was never created is refused the same way: you get a usable handle whose every call is a no-op, and it is not added to the module, so `Sections()` is as frozen as the bytes are. `Finalize` is idempotent — call it again and you get the same object and the same error, not a second pass.

Section kinds and symbol attributes are `obj`'s constants, re-exported at the root. `amd64.Text` and `obj.Text` are the same value, so a lowering that never imports `obj` still spells them and a downstream consumer never converts. Every example here uses the root spelling.

### Emit an object

Once the writers land — see [Writers](#writers) — this is the whole of it:

```go
f, err := os.Create("hello.o")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

if err := elf.Write(f, o); err != nil {
    log.Fatal(err)
}
```

`Options` is there for the things the assembler has no opinion about, and is variadic so the common call takes none:

```go
elf.Write(f, o, elf.Options{
    OSABI:   elfconst.ELFOSABI_LINUX,
    Comment: "vertex 0.4",
})
```

Same object, different container, one import changed:

```go
pe.Write(f, o, pe.Options{
    ABI:  pecoff.ABIMSVC,
    File: "hello.c",
    Directives: []pe.Directive{{Name: "DEFAULTLIB", Value: "msvcrt"}},
})

macho.Write(f, o, macho.Options{
    Platform:  machoconst.PlatformMacOS,
    MinOS:     "12.0",
    Subsections: true, // MH_SUBSECTIONS_VIA_SYMBOLS
})
```

There is no `Target` field on any of them. The object names its architecture and everything an AMD64 header needs follows from it; a caller who could pass a different one could produce a file whose header disagrees with its bytes.

A `RefKind` the container has no relocation for — `RefPLT32` or any ELF TLS model into COFF, `RefSecRel32` into anything but COFF — is `ErrRefKind` from the writer, naming the kind, the symbol, and the section offset. It is refused there rather than at construction because the same object is legal for a different container, and the whole point of one object and three writers is that the object does not have to pick.

Three writers, and for once that is all of them. There is no fourth container this architecture ships into.

### Read the finished object

```go
for _, s := range o.Sections() {
    s.Index()     // its position, which is what a symbol's Section names
    s.Kind()      // amd64.Text, amd64.Data, amd64.ROData, amd64.BSS, amd64.RelROData
    s.Name()      // ".text", or whatever SectionNamed was given
    s.Align()     // the largest alignment the builder asked for, at least 1
    s.Size()      // length in bytes
    s.Bytes()     // finished machine code, same-section labels patched
    s.Refs()      // []obj.Reference — the holes a linker fills
    s.Symbols()   // filtered view over the object's one symbol table
}

for _, sym := range o.Symbols() {
    sym.Name, sym.Section, sym.Offset, sym.Size
    sym.Binding, sym.Type, sym.Visibility
    sym.Defined()  // Section is -1 when undefined
}

o.Symbol("main")        // by name; identity is the name, module-wide
o.SectionNamed(".text")
o.SectionAt(0)
```

`Bytes()` returns a copy, and so does every other slice accessor.

Immutability is cheaper here than it was on i386. That module's ELF writer had to fold addends into the section bytes before writing `SHT_REL` entries, and bought immutability with a copy. x86-64 is a **RELA** architecture: the addend lives in the relocation entry, so `amd64/obj/elf` never touches the section bytes at all and hands them straight through. The COFF and Mach-O writers are still implicit-addend and still copy. Three writers over one object cannot see each other's work, whichever of them is copying.

The symbol table is module-level and in definition order, which is what every object writer wants.

### Lower from an IR

The builder is the only AMD64-specific type a frontend touches. Write your instruction selection against `*amd64.Section`, and everything downstream — object emission, linking, testing — is `obj.Object`:

```go
func (l *Lowering) block(t *amd64.Section, b *ir.Block) {
    t.Label(b.Name)                       // bare: a branch target, not a symbol
    for _, in := range b.Insts {
        switch in.Op {
        case ir.Add:
            t.AddR64RM64(l.reg(in.Dst), l.rm(in.Src))
        case ir.Load:
            t.MovR64RM64(l.reg(in.Dst), l.mem(in.Src))
        case ir.Call:
            t.CallRef(amd64.Ref(in.Sym, amd64.RefPLT32))
        case ir.Br:
            t.JmpLabel(in.Target)
        }
    }
}
```

No error checking in that loop, and that is the point: errors are sticky and first-wins, so the first failure positions itself and every later call is a no-op. Check once at `Finalize`, or call `m.Err()` to bail out of a long build early.

Calling conventions are not in this package. SysV puts the first six integer arguments in RDI, RSI, RDX, RCX, R8, R9 and gives you a 128-byte red zone; Microsoft x64 uses RCX, RDX, R8, R9, requires 32 bytes of shadow space, and has no red zone. Which one you are lowering for is a fact about your target, and this module does not have one — it has an instruction set. `Ret` emits `c3` and asks no questions about who set up the frame.

## Instructions

### Typed helpers — the primary surface

One method per declared form, named `MnemonicClassClass`:

```go
t.MovRM64Imm32(amd64.RAX, 60)                         // sign-extended into 64 bits
t.MovR64Imm64(amd64.RAX, 0xdeadbeefcafef00d)          // movabs, the only imm64
t.MovRM64R64(amd64.Mem64(amd64.RBX).Disp(8), amd64.RSI)
t.AddR64RM64(amd64.RAX, amd64.Mem64(amd64.RSP))
t.ShlRM64Imm8(amd64.RDX, 3)
t.JzLabel("done")
```

An `RM` class takes a register or an address, so `MovRM64Imm32(amd64.RAX, 60)`
and `MovRM64Imm32(amd64.Mem64(amd64.RBX), 60)` are the same helper. Scalar SSE
adds two classes of its own, `XM32` and `XM64` — an xmm register or four or
eight bytes of memory, which is the memory half of `RM32` and `RM64` and the
register half of neither, so `MovssXmmXM32(amd64.XMM0, amd64.EAX)` does not
compile. The AVX tranches are not in the table yet, so `VaddpsYmmYmmRM256` is a name
this convention will produce rather than a method that exists today — see
[Known limitations](#known-limitations).

The operand classes are the parameter types, so a width or class mismatch is a compile error: `MovR64Imm64(amd64.EAX, …)` does not build, because `EAX` is a `reg.R32`. An isel bug is a red squiggle rather than a runtime `ErrForm`.

A helper pins its form — exactly the encoding you named — and the diagnostics follow from that. A helper checks operand *kinds* only and hands values to the encoder: the wrong kind of operand is `ErrForm`; a value that does not fit the field the form pins is `ErrRange`, with the field width and range in the error's notes and the encoder's error as the cause. You named the form, so a too-big constant is a value problem, not a missing-form problem. Gated helpers still gate: a gated helper on a module without the level or extension fails at the call with `ErrFeature`, naming the gate.

`inst_*.go` is written by hand against the internal table and binds each helper to its form **by name lookup, not by table index**. Appending rows breaks nothing, and a removed or renamed form panics at program start naming the missing form, rather than silently binding to the wrong row or failing halfway through someone's code generation. Two helpers binding the same name is a duplicate Go identifier and so a compile error, which is the earliest any of these can fail. `amd64_test.go` closes the loop from the other side, so a row nobody binds fails a test rather than sitting unreachable.

Naming conventions beyond the class spelling:

- **Fixed operands are in the name, not the parameters.** `AddRAXImm32(60)` — the form names RAX and leaves no field to put another register in. Same for `AL`, `CL`, `DX`, `EAX`, and the literal `1` of `ShlRM64One`.
- **Branch targets split by where they resolve.** `JmpLabel("loop")` is same-section, patched at `Finalize`, no relocation. `JmpRef(amd64.Ref("f", amd64.RefPC32))` leaves the section and survives into `Refs()`. `CallLabel` is the bare-label call to a function compiled into this module; `CallRef` is the one that crosses.
- **`Short` pins rel8.** `JmpShortLabel` and `JzShortLabel` are the 2-byte forms; the plain names pin rel32. No relaxation between them: a short branch to a far target is `ErrRange` at `Finalize`. Mnemonics with only a rel8 form (`LoopLabel`, `JrcxzLabel`) take the plain name.
- **`Lock` is a prefix in the name.** `LockAddRM64R64` is `ADD r/m64, r64` plus `f0`, cloned from that row so the bytes cannot drift. A register handed to one is `ErrOperand` at the call, because `LOCK` on a register destination is `#UD` and that is a fact about the value rather than its class.
- **`V` is part of the mnemonic, not a modifier.** `VaddpsXmmXmmRM128` and `AddpsXmmRM128` are different forms with different operand counts, because the VEX encoding gave the instruction a non-destructive third operand. They are not two spellings of one thing and the surface does not pretend otherwise.
- **Masking and broadcast ride on the operand.** `.Mask(amd64.K1)`, `.Z()` and `.Bcst()` are chain methods on the EVEX-capable operands, so a form that has no EVEX encoding cannot be given a mask — it is a compile error, not a gate.
- **Both documented spellings exist and both emit the same bytes.** `JeLabel` and `JzLabel` are both `0x74`; `SalRM64One` emits `SHL`'s bytes. They are separate forms with an `AliasOf`, so a listing says which name the caller used and the bytes say what the silicon does.

### Emit — the escape hatch

```go
t.Emit("mov", amd64.RAX, amd64.NewImm(60))
t.Emit("add", amd64.RAX, amd64.NewImm(1))     // picks the 4-byte imm8 form
t.Emit("jmp", amd64.NewLabel("done"))         // pins rel32
```

Runtime form resolution for table-driven emission, where the mnemonic is data. If you know the instruction at compile time, use the typed helper.

Two selection rules, and the second one is a correction to the obvious design:

- **Among forms whose length is decidable now, shortest wins,** ties broken by table order. `add rax, 1` gets the four-byte sign-extended imm8 form over the seven-byte imm32 form, which is the whole reason those rows exist. Length is computed by encoding every candidate, so there is no size estimator to disagree with the encoder.
- **Among `rel` forms, widest wins.** A branch displacement is not known until `Finalize`, so "shortest" would mean "always rel8, and fail later if the target is far" — a rule that is correct for immediates and a landmine for branches. `Emit("jmp", …)` pins rel32 and always reaches. A caller who wants the two-byte encoding is making a claim about distance, and claims belong in the typed surface: `JmpShortLabel`.

There is a third thing that looks like a rule and is not one. `Emit("mov", RAX, NewImm(1))` produces seven bytes (`48 c7 c0 01 00 00 00`), not five, even though five bytes of `b8 01 00 00 00` would leave RAX holding 1. That is because `mov r32, imm32` is a *different form* with a different destination class, and `Emit` picks an encoding of the instruction you named, never a different instruction. Getting the five-byte version means knowing that a write to a 32-bit register zeroes the upper half, which is a claim about the architecture, so it goes in the typed surface where you state it: `MovR32Imm32(EAX, 1)`.

Prefix selection is the encoder's and is not a choice at all. REX is emitted when a field needs it and omitted when no field does; a VEX form takes the two-byte `c5` when the fields fit and the three-byte `c4` when they don't. Neither is a form and neither is visible in the table, because there is no case where a caller would want the longer one.

`Emit` never reaches a locking form: `LOCK` is not part of a mnemonic, so naming one is what the typed surface is for. Its variadic takes `operand.Operand`, which is sealed — bare Go integers do not coerce, and `amd64.NewImm(n)` is the spelling. The seal is what keeps an arbitrary type out of a mnemonic-as-data path.

A mnemonic the table does not have is `ErrForm` saying so by name. Operands no form accepts is `ErrForm` naming the operands. Operands that matched a form the feature set does not permit is `ErrFeature` naming the gates that would allow it.

## Registers and widths

```go
amd64.RAX  amd64.EAX  amd64.AX   amd64.AL   amd64.AH
amd64.R8   amd64.R8D  amd64.R8W  amd64.R8B
amd64.XMM0 amd64.YMM0 amd64.ZMM0 amd64.K1
```

Sixteen general-purpose registers in four widths each, thirty-two vector registers, eight mask registers. Each width is a distinct Go type, which is what makes `AddR64RM64(EAX, …)` fail to compile.

Two facts about this architecture leak through the surface on purpose, because hiding either produces wrong code.

**A 32-bit destination zeroes the upper half. A 16- or 8-bit destination does not.** `XorR32RM32(EAX, EAX)` clears all sixty-four bits of RAX in two bytes; `XorR16RM16(AX, AX)` clears sixteen and leaves the rest. This package will not silently promote your 16-bit operation to a 32-bit one to make that surprise go away, because the surprise is the ISA and your lowering has to know it either way. What it will do is make the shorter idiom easy to spell.

**`AH`, `CH`, `DH` and `BH` cannot appear in an instruction that carries a REX prefix**, because those encodings are how `SPL`, `BPL`, `SIL` and `DIL` are reached once REX is present. A high-byte register in a form the encoder must REX-prefix — anything naming a 64-bit register, R8 through R15, or any of the new byte registers — is `ErrOperand` at the call, naming both registers and saying which one forced the prefix. It is `ErrOperand` rather than `ErrForm` because the kinds were right; the combination is what the silicon cannot encode.

`reg.RIP` exists and is not an `R64`. It is a base for `Rip()` and nothing else, because no instruction takes it as an operand, and an interface it satisfied would be a claim about a capability that does not exist.

## Sections and data

There is one section per kind under the conventional name, and `Section(kind)` creates it on first use:

```go
t := m.Section(amd64.Text)        // ".text"
```

For anything outside those four names — debug info, unwind tables, a custom note section — `SectionNamed` takes the name and the load-time kind it behaves as:

```go
d := m.SectionNamed(".debug_line", amd64.Data)
d.Data(dwarfBytes)
```

`Section(k)` is exactly `SectionNamed(k.String(), k)`, so the two cannot disagree about `.text`. A name asked for twice returns the same section; a name asked for with a different kind than it was created with is `ErrDuplicate`, because a section that is code on Tuesday and data on Wednesday is a bug with a delayed fuse.

Names are spelled the ELF way and translated by the writers, because a name has to be spelled *some* way and one of the three containers has to be the one the vocabulary borrows from. `.rodata` becomes `.rdata` for COFF and `(__TEXT,__const)` for Mach-O; see [Writers](#writers) for the rules and their edges.

```go
r.Byte(0x90)
r.Long(0xdeadbeef)         // little-endian, 4 bytes
r.Quad(1)                  // little-endian, 8 bytes
r.Ascii("no terminator")
r.Asciz("terminated")
r.Zero(64)
r.Data(blob)               // raw bytes
r.Ref(amd64.Ref("f", amd64.RefAbs64)) // 8-byte hole + a relocation
r.LabelRef("case_3")                  // 8-byte hole, patched at Finalize
r.LabelDiff("case_3", "table")        // 4-byte hole, same-section difference
```

The raw-bytes builder is `Data` because `Bytes()` is the read side of the contract; there is no `Word`, because a name whose width depends on the package you're in is not a name.

`Ref`, `LabelRef` and `LabelDiff` are the data-side twins of what an instruction operand does. They exist because this package refuses to build your vtables, jump tables and literal pools and must not make them unbuildable. `LabelDiff` is the one that earns its place on this architecture: a jump table of 32-bit offsets from the table's own base is how you avoid eight bytes per entry and a relocation per entry, both operands are same-section labels, and the subtraction is exact at `Finalize` with nothing left for a linker to do.

`Align(n)` pads a code section with the multi-byte nop sequences and a data section with zeros. Unlike i386, there is no gate on this: `0F 1F` is in the x86-64 baseline, so every module this package can build can execute the long nops. `n` must be a power of two; anything else is `ErrAlign`. The largest `n` a section sees also becomes `Section.Align()`, which is what the object writer stamps on the section header.

`Offset()` is the current end of the section: the offset the next byte will land at, the value a `Label` placed now would name. It is exported because those tables are yours to build, and building them requires knowing where you are.

## Memory operands

Four constructors per access width, one question each:

```go
amd64.Mem64(amd64.RBX)                        // based:      [rbx]
amd64.Rip(amd64.Ref("msg", amd64.RefPC32))    // RIP-relative: [rip + disp32]
amd64.Abs64(amd64.Ref("msg", amd64.RefAbs64)) // symbolic:   [msg], a 64-bit relocation
amd64.Addr32(0xb8000)                         // direct:     [0xb8000], no relocation
```

`Rip` is the one to reach for and the one this architecture was given. On i386, loading a symbol's address from `.text` meant an absolute relocation in a non-PIC build or an EIP thunk and the GOTOFF dance in a PIC one, and the README there had to explain which instruction sequence you were on the hook for. Here `lea rdi, [rip + msg]` is four bytes, position-independent, and needs no thunk, no GOT, and no decision. PIC is the default because the addressing mode is.

`Rip` takes a `SymRef` and nothing else — there is no `Rip(disp)` overload — because a RIP-relative displacement to a numeric address is a claim about where the instruction will be linked, which is not knowable here and not something a caller should be able to spell.

Chain methods refine the address and keep the width's type, so the result still satisfies the helper parameter it is written into:

```go
amd64.Mem64(amd64.RBX).Disp(8)
amd64.Mem64(amd64.RBP).Index(amd64.RSI, 4).Disp(-12)
amd64.Addr32(0).Index(amd64.RDI, 8)           // index-only
amd64.Mem64(amd64.RBX).Seg(amd64.FS)          // the TLS base
```

A displacement applied to a symbolic address folds into the symbol's addend, so `[rbx+8].Sym(x)` and `[rbx].Sym(x).Disp(8)` are the same operand and there is one place for the number whichever order the chain was written in.

Three encoding facts the constructors handle so you never see them: RSP or R12 as a base needs a SIB byte; RBP or R13 as a base needs an explicit displacement even when it is zero; and a plain 32-bit absolute address needs a SIB byte with no base and no index, because the encoding that meant *disp32* on i386 means *RIP-relative* here. `Addr32(0xb8000)` is a byte longer than its i386 twin for that reason, and it is the encoder's business, not yours.

RSP cannot be an index, and R12 can — the SIB index field has no encoding for RSP, and REX.X is what tells the two apart. That asymmetry is checked, and getting it wrong is `ErrOperand` naming the register.

Widths run `M8` through `M512`. `LeaR64M` and `InvlpgM` take `Memory` rather than an `RM64`, because their operand is an address and has no access width — a register in that slot would not be an instruction.

Construction errors — RSP as an index, a scale that is not 1/2/4/8, an index on a RIP-relative address — are sticky on the operand and surface at the instruction that uses it, positioned, under `ErrOperand`. A builder chain is not followed by a run of error checks, and the diagnostic still points at the instruction rather than at the encoder.

## Symbols

```go
t.Label("loop")                                // bare: a branch target, this section only
t.Label("main", amd64.Global, amd64.Func)      // attributed: a symbol, size range opens
t.EndLabel("main")                             // closes it
m.Extern("puts")                               // undefined, Section == -1
m.Alias("_main", "main")                       // second name, same offset
m.SetVisibility("main", amd64.Hidden)
```

**A bare `Label` is not a symbol.** It names an offset in this section's namespace, gets patched at `Finalize`, and leaves no trace in the symbol table. Any attribute — a `Binding`, a `SymbolType`, a `Visibility` — promotes it into `Symbols()`. That rule is why nothing here mangles and nothing invents `.L` prefixes: a block label is a bare label, a local symbol you want emitted is `amd64.Local`, and whether locals reach the file at all is the writer's `StripLocals`.

Mach-O is the container that makes this rule load-bearing rather than tidy. `MH_SUBSECTIONS_VIA_SYMBOLS` tells `ld64` it may split a section into atoms at symbol boundaries and dead-strip them independently, and a section whose only symbol is at offset zero is one atom no matter how many functions are in it. If you set `macho.Options.Subsections`, give every function an attributed `Label`.

`Size` closes at `EndLabel` if you call it, at the next symbol in the same section if you don't, and at the section end otherwise. It is stated rather than guessed because `STT_FUNC` symbols with a zero size defeat GC in every linker that does it, and the next-symbol fallback is a guess, so prefer `EndLabel` for anything you care about.

`Alias` and `SetVisibility` resolve at `Finalize`, so the order of the call and the `Label` it names does not matter.

Symbol identity is the name, module-wide. A duplicate definition is `ErrDuplicate` at the second `Label`, naming the first one's section and offset. Nothing here adds a leading underscore; if you are targeting Mach-O and want `_main`, spell it.

## References

```go
amd64.Ref("puts", amd64.RefPLT32)         // call through the PLT
amd64.Ref("msg", amd64.RefPC32)           // RIP-relative, the common case
amd64.Ref("p", amd64.RefRexGOTPCRELX)     // GOT load the linker may relax away
amd64.Ref("tls_var", amd64.RefGOTTPOFF)   // initial-exec TLS
```

```
RefAbs64  RefAbs32  RefAbs32S  RefAbs16  RefAbs8      absolute, that many bytes
RefPC64   RefPC32   RefPC16    RefPC8                 PC-relative
RefPLT32                                              call via procedure linkage table
RefGOT32  RefGOTPCREL  RefGOTPCRELX  RefRexGOTPCRELX
RefGOTOFF64  RefGOTPC32                               global offset table forms
RefSize32  RefSize64                                  symbol size, not address
RefTLSGD  RefTLSLD  RefDTPOFF32  RefDTPOFF64
RefGOTTPOFF  RefTPOFF32  RefTPOFF64                   ELF thread-local models
RefTLV                                                Mach-O thread-local variable
RefImageRel32  RefSecRel32  RefSecIdx                 COFF image- and section-relative
```

The kind is stated at construction because it is a decision — PLT or direct, GOT or absolute, relaxable or not — and this package does not make decisions. `call puts@plt` and `call puts` are byte-identical, `e8` either way, so the kind rides beside the bytes as data.

That list is the *union* of what the three containers can express, not the intersection, and the union is deliberate. An intersection would be four kinds and would make every interesting object unbuildable; a per-container kind set would mean a lowering picks its container before it picks its instructions. So the set is the union, every writer states what it cannot do, and `ErrRefKind` names the kind and the offset when a kind meets a container that has no answer for it.

`RefGOTPCRELX` and `RefRexGOTPCRELX` are two kinds for one relocation semantics because a linker relaxing `mov foo@GOTPCREL(%rip), %reg` into `lea foo(%rip), %reg` has to know how many bytes back the instruction starts, and the answer depends on whether a REX prefix is there. The encoder knows — it emitted the prefix or didn't — but it is not the encoder's call to make, because emitting the relaxable kind is a statement that the addend is exactly the one the psABI's transformation assumes. So the kind is yours to state and the encoder refuses the wrong one: `RefGOTPCRELX` on a form that carries REX is `ErrRefKind` at the call, naming the prefix.

### Adjust, and the one identity three formats disagree about

```go
type Reference struct {
    Offset int      // where the hole starts, section-relative
    Size   int      // 1, 2, 4 or 8
    PCRel  bool
    Adjust int64    // field-position correction, already computed
    Sym    string
    Kind   RefKind
    Addend int64    // logical addend, never adjusted for the field
}
```

`Adjust` is the reason you never write `-4`. A PC-relative field is resolved against the end of the instruction, and the field is not always the last thing in it; the encoder that placed the field knows how many bytes follow it. The identity every consumer relies on is

```
value = target - (section offset of the field) + Adjust + Addend
```

so a `disp32` with nothing after it carries `Adjust == -4`, and one with a trailing `imm32` carries `Adjust == -8`.

On i386 that number went two places: `patchLabels` computed with it directly, and each writer folded it into its own arithmetic. Here it goes three, and two of the three do something the i386 module never had to.

**ELF wants it as an addend.** x86-64 is RELA, so `Addend + Adjust` is written into `r_addend` and the section bytes are left alone. The `-4` that every hand-written x86-64 assembly listing carries on a `PC32` is this, computed.

**COFF wants it as a relocation type.** `IMAGE_REL_AMD64_REL32` measures from the byte after the four-byte field, and `REL32_1` through `REL32_5` measure from one to five bytes further along. That is `Adjust` and nothing else: `-4` selects `REL32`, `-8` selects `REL32_4`, and the logical `Addend` — with the `Adjust` term removed, because the relocation type is now carrying it — folds into the section bytes.

**Mach-O wants it as a relocation type too, with a hole in the middle.** `X86_64_RELOC_SIGNED` is the field-ends-the-instruction case and `SIGNED_1`, `SIGNED_2` and `SIGNED_4` cover one, two and four trailing bytes. There is no `SIGNED_3`. An instruction with three bytes after its displacement has no Mach-O encoding, so it is `ErrRefKind` from that writer alone, naming the instruction offset — and the same object still writes cleanly to ELF and COFF, which is exactly why the refusal lives in the writer.

One struct, one identity, and three different answers derived from the same integer by the only three packages that know which format they are.

`Size` reaches 8 and `Addend` is an `int64` because `R_X86_64_64` against a quadword is an ordinary thing to want. Nothing in the struct is AMD64-specific, and that is on purpose: `obj` is a vocabulary an architecture fills in, and it is written so a second one could without the struct changing shape.

Symbol *differences* across sections are not in the kind set. The same-section case is `LabelDiff` and needs no relocation, which covers the jump tables and DWARF line programs that actually motivate it. The cross-section case is Mach-O's `SUBTRACTOR`, which is a relocation *pair* naming two symbols against one hole, and a two-symbol `Reference` is a change to the shared vocabulary that should be made once, for a reason, with all three writers ready for it. See [Known limitations](#known-limitations).

## Features

```go
m := amd64.NewModule(amd64.WithFeatures(
    amd64.NewFeatureSet(amd64.V3).Add(amd64.AES),
))
```

Default is `x86-64-v1`, the baseline, with no extensions. The feature set is fixed at construction and nothing about it is configurable after: a gate that changed mid-module would make already-emitted diagnostics unfalsifiable.

Two different things live in the vocabulary and are deliberately not flattened together. A **level** is a point on the cumulative `x86-64-v1`..`v4` ladder the psABI defines, and a level is a bundle rather than an instruction group: `v2` is CMPXCHG16B, LAHF-SAHF, POPCNT, SSE3, SSSE3, SSE4.1 and SSE4.2 together; `v3` adds AVX, AVX2, BMI1, BMI2, F16C, FMA, LZCNT, MOVBE and XSAVE; `v4` adds AVX512F, AVX512BW, AVX512CD, AVX512DQ and AVX512VL. A **feature** is an orthogonal extension with a CPUID bit of its own and no level that requires it — AES, PCLMULQDQ, SHA, RDRAND, RDSEED, ADX, and the AVX-512 extensions past the v4 five.

The i386 module's levels were a CPU ladder where the early rungs predated CPUID. These are not that: every feature in every level has a CPUID bit, and the levels are a naming convention over feature sets that distributions and loaders agreed on. `Level.Adds` gives the exact set, so a driver can print it rather than paraphrase it.

Sets are closed under requirements in both directions: adding AVX2 brings AVX, removing AVX drops AVX2 and F16C and FMA, because a set holding AVX2 but not AVX describes no silicon.

`ParseFeatures` resolves the spellings the world writes, against a starting set:

```go
amd64.ParseFeatures(amd64.DefaultFeatures(), "x86-64-v3+aes")  // exact: level plus extensions
amd64.ParseFeatures(base, "+avx512vnni,-avx512dq")             // adjust: applied left to right
```

`Set.String()` prints the canonical spelling — `x86-64-v3+aes+sha` — and `Parse` accepts it back, so the two round-trip. Removing a feature a level requires demotes the level and prints as the lower level plus what survives; there is no way to hold a set that says `v3` and not `avx`.

Spellings that name something real but wrong get the reason, not "unknown": `sse2` and `cmov` are part of the baseline and cannot be removed; `3dnow` was removed from the architecture; `sse4a` is AMD's and is not a spelling of SSE4.2; `apx` and `avx10` are not declared in this table yet and say so rather than failing as typos; i386's `i486`..`i686` levels are defined over a 32-bit baseline and have no 64-bit member.

`feature.Levels`, `feature.All`, `Level.Adds`, `feature.Requires` and `feature.RequiredBy` are there so a driver can print its own `--help`. The data is the library's; the formatting is the consumer's.

## Errors

```
ErrFeature  ErrForm  ErrOperand  ErrDuplicate
ErrUndefined  ErrRange  ErrAlign  ErrFinalized  ErrRefKind  ErrSectionName
```

Each names one failure and only one:

- `ErrForm` — the operands are the wrong *kinds* for the form, or no form of that name exists.
- `ErrOperand` — the operands are the right kinds and one of them was built wrong. RSP as an index, a scale of 3, `AH` in a REX-prefixed form, a register handed to a locking form. It is its own sentinel because "no matching form" sends a caller hunting through the ISA table for a row that exists.
- `ErrRange` — a value does not fit a field. Both the immediate at the call site and the label displacement at `Finalize`, both carrying the field width and reachable range in the notes. There is no branch relaxation and no silent form substitution: the failure is loud instead of the bytes being different.
- `ErrRefKind` — the writer has no relocation for a kind, or no encoding for this `Adjust`. Only ever comes from a writer.
- `ErrSectionName` — the writer cannot place a custom section. Only ever comes from `amd64/obj/macho`.

The concrete type is `*obj.Error`, with `Arch`, `Section`, `Offset`, `Context`, `Sentinel`, `Cause` and `Notes`. One error type across the builder and all three writers, and across both architectures in the tree, so tooling that formats a diagnostic is written once:

```
amd64 .text+0x11: 300 does not fit ADD r/m64, imm8: value out of range
  note: the immediate field of ADD r/m64, imm8 is 1 bytes; the range is -128..127
```

Errors are sticky and first-wins: every builder call after a failure is a no-op, and `Finalize` surfaces the first one, positioned. `Module.Err()` returns the same error `Finalize` will.

`errors.Is` works against every sentinel. Where an encoder or operand error is the cause, it joins the chain — `Unwrap` returns both the sentinel and the cause — but its type is internal; anything a caller might need from it is restated in `Notes` as text.

## Writers

> **Not in the tree yet.** `obj/elf`, `obj/pe` and `obj/macho` are the next
> tranche. What follows is the contract they are being built to — the signature,
> the options, and every kind each container refuses — written down first because
> the refusals are the design and they have to agree with `RefKind` before a line
> of any of them is written. Everything above this section is in the tree and
> works today.

All three leaves are one exported function and one options struct. Everything else in them is the translation table for their format.

### `amd64/obj/elf`

```go
func Write(w io.Writer, o *obj.Object, opts ...Options) error

type Options struct {
    OSABI       elf.OSABI     // e_ident[EI_OSABI]; zero is ELFOSABI_NONE
    ABIVersion  uint8
    GNUStack    elf.GNUStack  // zero declares a non-executable stack
    Comment     string        // written to .comment when non-empty
    StripLocals bool          // drops locals no relocation names
}
```

x86-64 is a RELA architecture, so the addend lives in the relocation entry and the section bytes are written through untouched. The writer states `RelocRELA` rather than letting the underlying module infer it, because a psABI table that ever said otherwise should break loudly instead of quietly writing an addend the linker will not read.

`StripLocals` keeps a local that a relocation names regardless of the flag. Dropping it would leave a hole pointing at nothing, and a stripped object that does not link is worse than a slightly larger one.

The one heuristic in the package is in `section.go`: a section whose name begins `.debug` or is `.comment` is not `SHF_ALLOC`, whatever kind it was created with. There is no name under that rule a caller would want allocated.

### `amd64/obj/pe`

```go
func Write(w io.Writer, o *obj.Object, opts ...Options) error

type Options struct {
    ABI             pe.ABI          // zero is ABIMSVC
    TimeDateStamp   uint32          // zero is the deterministic choice
    BigObj          coff.BigObjMode
    Characteristics pe.FileChar
    File            string          // recorded as a .file symbol
    Directives      []Directive     // .drectve options
}
```

COFF carries less than ELF does, and the writer drops what it must rather than inventing what it can't:

- **`Weak` is refused,** as `pe.ErrWeak`. A COFF weak external is a name plus the alternate definition to use when nothing else defines it, and an object stating `Weak` and nothing else has not said what that alternate is.
- **`Size` and `Visibility` are dropped.** A COFF symbol's extent is the distance to the next one, and whether a name leaves a DLL is decided at link time. Both are accepted and discarded rather than refused, because refusing `Hidden` would reject an object that links correctly.
- **`ThreadLocal` becomes `NULL`.** TLS in COFF is a property of the `.tls` section a variable lands in, not of its symbol.
- **`ROData` under its conventional name becomes `.rdata`,** because that is what link.exe's default merge rules are written against. A section a caller named itself with `SectionNamed` keeps its name.
- **`RefPLT32`, the GOT kinds and every ELF TLS model are `ErrRefKind`.** There is no PLT and no GOT; an imported call reaches its target through a linker-synthesized thunk and an ordinary `REL32` to the `__imp_` symbol, which is a *different symbol*, so mapping `RefPLT32` onto `REL32` would relocate against a name the lowering never meant.
- **`RefPC64` and the 8-bit widths are `ErrRefKind`.** COFF's PC-relative field is 32 bits and its smallest field is 16.
- **`Adjust` past `-9` is `ErrRefKind`**, because `REL32_5` is the last of that ladder. Nothing this encoder emits reaches it; the check is there so that a future form which does fails at the writer rather than producing a `REL32` off by six.

### `amd64/obj/macho`

```go
func Write(w io.Writer, o *obj.Object, opts ...Options) error

type Options struct {
    Platform    macho.Platform      // required; an object with no platform makes every linker guess
    MinOS       string              // e.g. "12.0"
    SDK         string
    Subsections bool                // MH_SUBSECTIONS_VIA_SYMBOLS
    Sections    map[string]SegSect  // segment/section for custom names
}
```

Mach-O is the container that disagrees with the other two about naming, and the writer's job is mostly that disagreement:

- **The four kinds map to their conventional pairs**: `Text` to `(__TEXT,__text)` with `S_ATTR_PURE_INSTRUCTIONS`, `Data` to `(__DATA,__data)`, `ROData` to `(__TEXT,__const)`, `BSS` to `(__DATA,__bss)` as `S_ZEROFILL`.
- **DWARF names are known**: `.debug_line` and its siblings become `(__DWARF,__debug_line)` with `S_ATTR_DEBUG`, because DWARF section names are the one custom family standardized across all three containers, and a caller writing debug info should not have to know that this container spells them differently.
- **Any other custom name is `ErrSectionName`** unless `Options.Sections` gives it a segment. There is no default that would be right: a segment is a load-time protection decision and guessing `__DATA` for a name the writer has never seen would produce a file that loads and misbehaves rather than one that fails. The escape hatch is one map entry and the error names the section that needs it.
- **`RefPLT32` becomes `X86_64_RELOC_BRANCH`.** Unlike COFF, this is a faithful mapping: the target is the symbol the lowering named, and `ld64` synthesizes a stub if it needs one.
- **`RefGOTPCREL` becomes `GOT`, and the relaxable kinds become `GOT_LOAD`,** which is `ld64`'s equivalent claim — the instruction is a `movq` load of a GOT entry and may be rewritten to a `leaq`.
- **`RefTLV` is the only thread-local kind accepted.** Mach-O reaches a thread-local through a descriptor and a call, not a relocation model; the ELF TLS kinds are `ErrRefKind` and the descriptor sequence is an *instruction sequence* your lowering emits.
- **`RefSize32`, `RefSize64`, `RefImageRel32`, `RefSecRel32` and `RefSecIdx` are `ErrRefKind`.** The first two are an ELF idea, the last three are a COFF one.

## How it's put together

- **The builder is concrete, the artifact is inert.** Typed helpers are methods on `*amd64.Section` because that is what makes a width mismatch a compile error; an interface or a generic builder would erase exactly the checking the surface exists for. Everything past `Finalize` is `obj.Object`: data with no methods that do anything but read.
- **`obj` imports nothing and knows no ISA.** It is the vocabulary the rest of the tree spells its types in, and the finished artifact the builder hands to a writer. That it imports nothing is what lets the writers live under it — `obj/elf`, `obj/pe`, `obj/macho` — while `operand` and `internal/encode` import it from above, with no cycle either way.
- **One declaration per concept, aliased upward.** `RefKind` is `obj`'s, aliased by `operand` and the root. `SectionKind` and the symbol attributes are `obj`'s, re-exported at the root. A value crossing a package line is never converted, only renamed.
- **The operand interface is sealed, and the seal lives in `reg`.** `operand` imports `reg`, so a seal declared above could never be satisfied by a register — and a register is the operand this tree is built out of.
- **Encoding is not in the table.** A form names an opcode, an operand shape, and which encoding *family* it belongs to — legacy, VEX or EVEX. Whether a legacy form gets a REX byte, and whether a VEX form gets two or three, is computed from the operands at encode time, because it is a function of the register numbers and nothing a table row could know. That is why `inst_*.go` is a few hundred hand-written lines against a few hundred rows rather than a generator: the rows are the ISA and the prefixes are arithmetic.
- **Writers are called, not registered.** The sibling format modules register backends because a linker resolves an architecture it cannot name at compile time. Here the caller names the format, so `elf.Write` is a function; the import graph already gives "only pay for what you use" without an `init` in sight.
- **Relocation mapping lives in the writer.** `amd64` never learns that ELF exists. Each leaf holds one small intent-to-format table and refuses the kinds its format has no answer for.
- **Emission reuses the format modules, not a private copy.** Each leaf builds a writer from the sibling module that also *reads* the format, so a round-trip test is a round trip and not two independent guesses. Each leaf aliases that import inside its own `write.go`, so the package a caller names is `elf`, `pe` or `macho` and the call site stays `elf.Write`. One line of aliasing per leaf is cheaper than an `out` suffix at every call site in every frontend.
- **Two table-integrity checks run before any caller can be wrong.** `reg` panics at first use on a duplicate register spelling; `isa` panics on two forms sharing a helper name. Both are questions about this tree's own data and should fail whatever the caller is.

## Known limitations

- **The table is base integer, scalar SSE, the system tranche and the locking forms: 786 forms over 265 mnemonics.** `table_base.go` declares the x86-64 baseline — the ALU block, the shifts and rotates, the unary group, `mov`/`movsx`/`movzx`, the stack, the branches, the full condition-code family as `j`/`set`/`cmov`, and the baseline-adjacent rows that carry a gate of their own — POPCNT, LZCNT, TZCNT, CMPXCHG16B and MOVBE, and the operand-less privileged and serializing instructions inline assembly reaches — HLT, CLI, STI, PAUSE, RDTSC, RDMSR, WRMSR and WBINVD, none of which anything here selects. `table_system.go` adds what an instruction selector never emits and a C header writes by hand: the bit-test group, the double-precision shifts, port I/O, the flags on the stack, the descriptor-table and TLB instructions, the cache and state-management ones, and the privileged instructions with no operands. Nothing in this tree selects any of it — the assembler is the door it comes through, which is also why it was absent until there was one. `table_sse.go` adds scalar SSE and SSE2: the moves, the six scalar arithmetic operations in both widths, `sqrt`, the four logical rows, the compares, and every conversion between the two register files. It is ungated, because SSE and SSE2 are inside `x86-64-v1` and so are baseline in the same sense `add` is. **Packed arithmetic is not in it** — `addps` adds four floats and `addss` adds one, and only the scalar half is declared. `table_lock.go` adds the memory-ordering tranche: the locking clone of every row that admits LOCK, generated from the table rather than written out, plus the three fences. `table_avx.go` is named in `forms.go`'s `init` and commented out, because it is not written yet: AVX and AVX-512 parse, print and round-trip as feature sets and gate nothing, because no VEX or EVEX row is declared, and the VEX and EVEX prefix emitters in `internal/encode` are written and unreached. Extension tranches are additional files added the same way, each with its own `build*` called from the one `init` and its own `inst_*.go` opposite it. Nothing about an existing row changes when they land.
- **No unwind-table generation.** `.eh_frame` on SysV, `.pdata`/`.xdata` on Windows, `__compact_unwind` on Apple. All three are bytes: build the sections yourself with `SectionNamed` and `Data`. The Windows one is not optional in the way the others are — table-based exception handling requires an entry for every function that allocates stack or calls another, and an image without one unwinds unreliably rather than degrading — so a frontend targeting Windows needs a plan for it before it needs anything else on this list.
- **No cross-section symbol differences.** `LabelDiff` covers the same-section case. Mach-O's `SUBTRACTOR` needs a `Reference` naming two symbols against one hole, which is a shared-vocabulary change and should be made once with all three writers ready — ELF and COFF would have to synthesize a pair or refuse, and refusing after the vocabulary grew is worse than not growing it yet.
- **No COMDAT / ELF section groups,** so no inline-function or template deduplication. `.text` is one section per module, which matters more on this architecture than it did on i386.
- **No large or medium code model support beyond the instructions.** `MovR64Imm64` and an indirect `call r64` are in the table and are all the large model is; arranging them, and knowing when you need to, is your lowering's business.
- **No branch relaxation, by design.** A rel8 that does not reach is an error, not a silently widened instruction. In the small code model rel32 reaches everything, so unlike on a 32-bit target the absence costs nothing until you are past two gigabytes of text.
- **No APX.** REX2, EVEX-promoted legacy instructions, and r16–r31 are not in `reg` or the table, and `RefCode4GOTPCRELX` is not a kind. The relocation number is allocated and the ladder it belongs to is understood; the registers are the work.
- **No 16-bit or real-mode forms,** and no system forms that only exist outside long mode.
- **No disassembler.** This module writes bytes. Reading them back is `elf/obj`, `pe/coff` and `macho/obj`, which is where a round-trip test gets its other half.

## License

MIT