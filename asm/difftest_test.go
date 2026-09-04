package asm_test

// Differential testing against a reference assembler, the twin of
// arm64/asm/difftest_test.go. Every line is assembled by this package and by
// clang's integrated x86-64 assembler, and the two must agree byte for byte.
//
// x86-64 makes this test do more work than AArch64's. Instructions are not one
// width, so the comparison is over a byte stream rather than a word array, and
// a line assembling to a different *length* than clang chose is a difference
// this test has to notice rather than silently realign past. So each line is
// assembled on its own and compared on its own.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/vertex-language/amd64/asm"
)

var lines = []string{
	// Moves.
	"movq %rax, %rbx",
	"movl %eax, %ebx",
	"movw %ax, %bx",
	"movb %al, %bl",
	"movq $1, %rax",
	"movl $42, %ecx",
	"movq %rsp, %rbp",
	"movq (%rax), %rbx",
	"movq 8(%rax), %rbx",
	"movq -16(%rbp), %rax",
	"movq %rax, (%rbx)",
	"movq %rax, 24(%rbx)",
	"movq (%rax,%rcx,8), %rdx",
	"movq (%rax,%rcx,1), %rdx",
	"movl 4(%rsp,%rdi,4), %eax",
	"movq $0, (%rsp)",
	"movb $1, (%rax)",

	// Arithmetic.
	"addq %rax, %rbx",
	"addq $8, %rsp",
	"subq $16, %rsp",
	"subl %eax, %ecx",
	"addq (%rax), %rbx",
	"addq %rax, (%rbx)",
	"imulq %rax, %rbx",
	"negq %rax",
	"incl %eax",
	"decq %rbx",
	"adcq %rax, %rbx",
	"sbbl %ecx, %edx",

	// Logic and compares.
	"andq %rax, %rbx",
	"orl %eax, %ebx",
	"xorq %rax, %rax",
	"notq %rcx",
	"cmpq %rax, %rbx",
	"cmpq $0, %rax",
	"testq %rax, %rax",
	"testb $1, %al",

	// Shifts.
	"shlq $3, %rax",
	"shrl $1, %eax",
	"sarq $63, %rbx",
	"rolq $8, %rax",

	// Stack.
	"pushq %rbp",
	"popq %rbp",
	"pushq %rax",

	// Control flow.
	"ret",
	"call foo",
	"jmp foo",
	"jmp *%rax",
	"call *%rbx",
	// The bit-test, double-shift, port and system instructions: everything
	// an instruction selector never emits and a C header writes by hand.
	"bt %ecx, %eax",
	"bt %rcx, %rax",
	"btl $1, %eax",
	"btq $1, %rax",
	"btsl $1, (%rdi)",
	"btrq $1, (%rdi)",
	"btcl %ecx, (%rdi)",
	"lock btsl $1, (%rdi)",
	"bsfl %eax, %ecx",
	"bsfq %rax, %rcx",
	"bsrl %eax, %ecx",
	"bsrq %rax, %rcx",
	"shldl $4, %eax, %ecx",
	"shldq $4, %rax, %rcx",
	"shldl %cl, %eax, %ecx",
	"shrdl $4, %eax, %ecx",
	"shrdq %cl, %rax, %rcx",
	"pushfq",
	"popfq",
	"inb $0x80, %al",
	"inw $0x80, %ax",
	"inl $0x80, %eax",
	"inb %dx, %al",
	"inl %dx, %eax",
	"outb %al, $0x80",
	"outl %eax, $0x80",
	"outb %al, %dx",
	"outl %eax, %dx",
	"int $0x80",
	"cld",
	"std",
	"swapgs",
	"iretq",
	"sysretq",
	"clac",
	"stac",
	"xgetbv",
	"xsetbv",
	"rdtscp",
	"lgdt (%rdi)",
	"lidt (%rdi)",
	"sgdt (%rdi)",
	"sidt (%rdi)",
	"invlpg (%rdi)",
	"ltr %ax",
	"str %ax",
	"lldt %ax",
	"sldt %ax",
	"clflush (%rdi)",
	"prefetcht0 (%rdi)",
	"prefetcht1 (%rdi)",
	"prefetcht2 (%rdi)",
	"prefetchnta (%rdi)",
	"fxsave (%rdi)",
	"fxrstor (%rdi)",
	"xsave (%rdi)",
	"xrstor (%rdi)",
	"movnti %eax, (%rdi)",
	"movnti %rax, (%rdi)",

	// A byte-wide immediate is a byte, and a sign-extended one is not: the
	// 83 group's imm8 becomes a wider operand, so 200 in it would mean -56
	// and the wider form has to win instead.
	"andb $0xff, %al",
	"orb $0x80, (%rdi)",
	"movb $0xff, %al",
	"cmpb $0xff, %al",
	"addl $200, %eax",
	"addl $100, %eax",
	"addq $-1, %rax",
	"pushq $200",
	"pushq $100",
	"imull $200, %eax, %ecx",
	"andl $0xff, %eax",
	"shll $200, %eax",

	// Mnemonics whose width letter is peeled off by asking the table, and
	// one that is decided by the register file instead: MOVQ is the vector
	// move and MOV with a `q` is the integer one, which is the only place
	// in this table where one spelling means two instructions.
	"bswapq %rax",
	"bswapl %eax",
	"movq %rax, %rbx",
	"movq %xmm0, %rax",
	"movq %rax, %xmm0",

	// The extending moves, whose mnemonic names two widths: the first
	// letter sizes the source and the last sizes the destination.
	"movzbl (%rdi), %eax",
	"movzbq (%rdi), %rax",
	"movzwl (%rdi), %eax",
	"movzwq (%rdi), %rax",
	"movsbl (%rdi), %eax",
	"movsbq (%rdi), %rax",
	"movswl (%rdi), %eax",
	"movswq (%rdi), %rax",
	"movslq (%rdi), %rax",
	"movzbl %cl, %eax",
	"movsbl %cl, %eax",
	"movswl %cx, %eax",

	"nop",
	"leave",
	"ud2",
	"hlt",
	"cli",
	"sti",
	"pause",
	"rdtsc",
	"rdmsr",
	"wrmsr",
	"wbinvd",

	// Addresses.
	"leaq 8(%rax), %rbx",
	"leaq (%rax,%rcx,4), %rdx",
	"leaq foo(%rip), %rax",
	"movq foo@GOTPCREL(%rip), %rax",

	// Extends and conversions.
	"cltq",
	"cqto",
	"cwtl",

	// SSE.
	"movsd %xmm0, %xmm1",
	"addsd %xmm0, %xmm1",
	"mulsd %xmm2, %xmm3",
	"subss %xmm0, %xmm1",
	"divss %xmm4, %xmm5",
	"ucomisd %xmm0, %xmm1",
	"movaps %xmm0, %xmm1",
	"xorps %xmm0, %xmm0",
	"cvtsi2sdq %rax, %xmm0",
	"cvttsd2siq %xmm0, %rax",

	// SSE2's packed integer instructions, which are what the _mm_*
	// intrinsics lower to. amd64/simd_test.go pins the same forms
	// against ml64 for the machines with no clang on them.
	"movdqa %xmm2, %xmm1",
	"movdqa %xmm1, (%rax)",
	"movdqu %xmm2, %xmm1",
	"movdqu %xmm1, (%rax)",
	"paddb %xmm2, %xmm1",
	"paddw %xmm2, %xmm1",
	"paddd %xmm2, %xmm1",
	"paddq %xmm2, %xmm1",
	"psubb %xmm2, %xmm1",
	"psubw %xmm2, %xmm1",
	"psubd %xmm2, %xmm1",
	"psubq %xmm2, %xmm1",
	"paddsb %xmm2, %xmm1",
	"paddsw %xmm2, %xmm1",
	"paddusb %xmm2, %xmm1",
	"paddusw %xmm2, %xmm1",
	"psubsb %xmm2, %xmm1",
	"psubsw %xmm2, %xmm1",
	"psubusb %xmm2, %xmm1",
	"psubusw %xmm2, %xmm1",
	"pmullw %xmm2, %xmm1",
	"pmulhw %xmm2, %xmm1",
	"pmulhuw %xmm2, %xmm1",
	"pmuludq %xmm2, %xmm1",
	"pavgb %xmm2, %xmm1",
	"pavgw %xmm2, %xmm1",
	"pminub %xmm2, %xmm1",
	"pmaxub %xmm2, %xmm1",
	"pminsw %xmm2, %xmm1",
	"pmaxsw %xmm2, %xmm1",
	"psadbw %xmm2, %xmm1",
	"pmaddwd %xmm2, %xmm1",
	"pand %xmm2, %xmm1",
	"pandn %xmm2, %xmm1",
	"por %xmm2, %xmm1",
	"pxor %xmm2, %xmm1",
	"pcmpeqb %xmm2, %xmm1",
	"pcmpeqw %xmm2, %xmm1",
	"pcmpeqd %xmm2, %xmm1",
	"pcmpgtb %xmm2, %xmm1",
	"pcmpgtw %xmm2, %xmm1",
	"pcmpgtd %xmm2, %xmm1",
	"punpcklbw %xmm2, %xmm1",
	"punpcklwd %xmm2, %xmm1",
	"punpckldq %xmm2, %xmm1",
	"punpcklqdq %xmm2, %xmm1",
	"punpckhbw %xmm2, %xmm1",
	"punpckhwd %xmm2, %xmm1",
	"punpckhdq %xmm2, %xmm1",
	"punpckhqdq %xmm2, %xmm1",
	"packsswb %xmm2, %xmm1",
	"packssdw %xmm2, %xmm1",
	"packuswb %xmm2, %xmm1",
	"psllw %xmm2, %xmm1",
	"psllw $3, %xmm1",
	"pslld %xmm2, %xmm1",
	"pslld $3, %xmm1",
	"psllq %xmm2, %xmm1",
	"psllq $3, %xmm1",
	"psrlw %xmm2, %xmm1",
	"psrlw $3, %xmm1",
	"psrld %xmm2, %xmm1",
	"psrld $3, %xmm1",
	"psrlq %xmm2, %xmm1",
	"psrlq $3, %xmm1",
	"psraw %xmm2, %xmm1",
	"psraw $3, %xmm1",
	"psrad %xmm2, %xmm1",
	"psrad $3, %xmm1",
	"pslldq $3, %xmm1",
	"psrldq $3, %xmm1",
	"pshufd $0x1b, %xmm2, %xmm1",
	"pshuflw $0x1b, %xmm2, %xmm1",
	"pshufhw $0x1b, %xmm2, %xmm1",
	"pmovmskb %xmm1, %eax",
	"pinsrw $3, %eax, %xmm1",
	"pextrw $5, %xmm1, %eax",

	// Atomics: the shape every C header's atomics are written in.
	"lock addq $1, (%rax)",
	"lock incq (%rax)",
	"lock cmpxchgq %rcx, (%rax)",
	"lock xaddq %rcx, (%rax)",
	"xchgq %rax, (%rbx)",
	"mfence",
	"lfence",
	"sfence",

	// Segment-relative, which is how thread-local storage is reached.
	"movq %fs:0, %rax",
}

// widerThanClang are lines this package assembles correctly and longer.
//
// It is not a defect and it is not fixable here. The parent package picks the
// *widest* rel form on purpose: a displacement is not known until Finalize, so
// choosing rel8 would mean choosing it before knowing whether the target is
// within reach, and being wrong about that is a failure at the end of a build
// rather than a slightly larger function. clang runs a relaxation pass and can
// therefore start short. A caller who knows the distance says so by naming the
// short form, which is what the typed surface's JmpShortLabel is for.
var widerThanClang = []string{
	"je 1f", "jne 1f", "jl 1f", "jge 1f", "jmp 1f",
}

// TestWiderThanClangStillAssembles keeps that list from hiding a real
// failure: these have to assemble, and to be longer for the stated reason
// rather than for any other.
func TestWiderThanClangStillAssembles(t *testing.T) {
	for _, l := range widerThanClang {
		o, err := asm.Assemble(l+"\n1:\n", asm.Options{File: "w.s"})
		if err != nil {
			t.Errorf("%s: %v", l, err)
			continue
		}
		if n := len(o.SectionAt(0).Bytes()); n < 5 {
			t.Errorf("%s assembled to %d bytes; the near form is being picked "+
				"after all, and this list is stale", l, n)
		}
	}
}

// tableGaps are lines this package parses correctly and the parent package's
// table cannot encode.
//
// Empty, and kept for the next one. Its single entry was HLT, and what it
// actually named was a category: the privileged and serializing instructions
// with no operands, which nothing in this tree selects and which every kernel
// writes by hand. They are rows now, and the test below is what will say so
// when the next parser finds the next gap.
var tableGaps = []string{}

func TestTableGapsStillGap(t *testing.T) {
	for _, l := range tableGaps {
		if _, err := asm.Assemble(l+"\n1:\n", asm.Options{File: "g.s"}); err == nil {
			t.Errorf("%q now assembles; move it into lines", l)
		}
	}
}

func TestDifferentialAgainstClang(t *testing.T) {
	clangPath, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang not on PATH; skipping differential test against the reference assembler")
	}
	if _, err := exec.LookPath("objdump"); err != nil {
		t.Skip("objdump not on PATH")
	}

	dir := t.TempDir()
	ours := make([][]byte, len(lines))
	var refused []string
	for i, l := range lines {
		o, err := asm.Assemble(l+"\n1:\n", asm.Options{File: "l.s"})
		if err != nil {
			refused = append(refused, fmt.Sprintf("%-32s %s", l, firstLine(err.Error())))
			continue
		}
		ours[i] = o.SectionAt(0).Bytes()
	}

	// clang assembles each line into its own object too, so a length
	// difference on one line cannot shift every line after it.
	for i, l := range lines {
		if ours[i] == nil {
			continue
		}
		want, err := clangBytes(t, clangPath, dir, i, l)
		if err != nil {
			t.Errorf("%-32s clang: %v", l, err)
			continue
		}
		if !bytes.Equal(ours[i], want) {
			t.Errorf("%-32s got % x, want % x (clang)", l, ours[i], want)
		}
	}

	if len(refused) > 0 {
		t.Errorf("%d of %d lines were refused:\n  %s",
			len(refused), len(lines), strings.Join(refused, "\n  "))
	}
	if !t.Failed() {
		t.Logf("%d lines agree with clang", len(lines))
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// clangBytes assembles one line and returns its .text bytes.
func clangBytes(t *testing.T, clangPath, dir string, i int, line string) ([]byte, error) {
	t.Helper()
	src := filepath.Join(dir, fmt.Sprintf("l%d.s", i))
	obj := filepath.Join(dir, fmt.Sprintf("l%d.o", i))
	if err := os.WriteFile(src, []byte(line+"\n1:\n"), 0o644); err != nil {
		return nil, err
	}
	cmd := exec.Command(clangPath, "-target", "x86_64-linux-gnu", "-c", "-o", obj, src)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return textBytes(obj)
}

// textBytes disassembles .text and returns its bytes in address order.
var byteRe = regexp.MustCompile(`(?m)^\s+[0-9a-f]+:\s+((?:[0-9a-f]{2} )+)`)

func textBytes(objPath string) ([]byte, error) {
	out, err := exec.Command("objdump", "-d", "-j", ".text", objPath).CombinedOutput()
	if err != nil {
		return nil, err
	}
	var b []byte
	for _, m := range byteRe.FindAllStringSubmatch(string(out), -1) {
		for _, f := range strings.Fields(m[1]) {
			v, err := strconv.ParseUint(f, 16, 8)
			if err != nil {
				return nil, err
			}
			b = append(b, byte(v))
		}
	}
	return b, nil
}
