package reg

// The markers below exist so package operand can declare register-or-memory
// interfaces that only the right width satisfies:
//
//	type RM64 interface {
//	    reg.Operand
//	    RM64()
//	}
//
// R64 and operand.M64 both have RM64(), M32 does not, and AddR64RM64 will
// not compile with an M32. The methods are exported because an unexported
// one would belong to this package and operand could not name it; they are
// no-ops and carry no information beyond their own existence.
//
// They live in one file rather than beside each type so that the width
// classes can be read as a set, and so nobody mistakes them for behaviour.

func (R8) RM8()    {}
func (R16) RM16()  {}
func (R32) RM32()  {}
func (R64) RM64()  {}
func (Xmm) RM128() {}
func (Ymm) RM256() {}
func (Zmm) RM512() {}

// Scalar SSE's classes. An xmm register or four bytes of memory is not the
// same class as a general-purpose register or four bytes of memory, and
// only the memory half is shared — so these are markers of their own rather
// than M32 and M64 being reused, and an R32 will not compile into a
// MovssXmmXM32 no matter that both read four bytes.
func (Xmm) XM32() {}
func (Xmm) XM64() {}

// Mm has no marker. MMX's register-or-memory class is Mm-or-M64, which is a
// different class from R64-or-M64 under the same name, and declaring the
// wrong one now is worse than declaring none. It arrives with the MMX
// tranche.
