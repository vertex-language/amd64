package amd64

import (
	"github.com/vertex-language/amd64/operand"
	"github.com/vertex-language/amd64/reg"
)

var (
	vfmadd231ssXmmXmmXM32 = form("Vfmadd231ssXmmXmmXM32")
	vfmadd231sdXmmXmmXM64 = form("Vfmadd231sdXmmXmmXM64")
)

func (s *Section) Vfmadd231ssXmmXmmXM32(dst reg.Xmm, src1 reg.Xmm, src2 operand.XM32) {
	s.inst(vfmadd231ssXmmXmmXM32, dst, src1, src2)
}

func (s *Section) Vfmadd231sdXmmXmmXM64(dst reg.Xmm, src1 reg.Xmm, src2 operand.XM64) {
	s.inst(vfmadd231sdXmmXmmXM64, dst, src1, src2)
}
