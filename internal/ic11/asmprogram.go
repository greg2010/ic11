package ic11

import (
	"fmt"
	"strings"
)

type asmprogram struct {
	program    []instruction
	labelCount int
}

func newASMProgram() *asmprogram {
	return &asmprogram{program: []instruction{}, labelCount: 0}
}

func (asm *asmprogram) print() string {
	strArr := []string{}
	for _, instr := range asm.program {
		strArr = append(strArr, instr.ToString())
	}

	return strings.Join(strArr, "\n")
}

func (asm *asmprogram) getUniqueLabel() string {

	lbl := fmt.Sprintf("L_%d", asm.labelCount)
	asm.labelCount = asm.labelCount + 1

	return lbl
}

func (asm *asmprogram) emitLabel(lbl string) {
	instruction := &label{label: lbl}
	asm.program = append(asm.program, instruction)
}

func (asm *asmprogram) emitArity1(cmd string, a *data) {
	instruction := &instruction1{cmd: cmd, a: a}
	asm.program = append(asm.program, instruction)
}

func (asm *asmprogram) emitArity2(cmd string, a, b *data) {
	instruction := &instruction2{cmd: cmd, a: a, b: b}
	asm.program = append(asm.program, instruction)
}

func (asm *asmprogram) emitArity3(cmd string, a, b, c *data) {
	instruction := &instruction3{cmd: cmd, a: a, b: b, c: c}
	asm.program = append(asm.program, instruction)
}
func (asm *asmprogram) emitArity4(cmd string, a, b, c, d *data) {
	instruction := &instruction4{cmd: cmd, a: a, b: b, c: c, d: d}
	asm.program = append(asm.program, instruction)
}

func (asm *asmprogram) emitMove(a, b *data) {
	asm.emitArity2(move, a, b)
}

func (asm *asmprogram) emitAdd(a, b, c *data) {
	asm.emitArity3(add, a, b, c)
}

func (asm *asmprogram) emitSub(a, b, c *data) {
	asm.emitArity3(sub, a, b, c)
}

func (asm *asmprogram) emitMul(a, b, c *data) {
	asm.emitArity3(mul, a, b, c)
}

func (asm *asmprogram) emitDiv(a, b, c *data) {
	asm.emitArity3(div, a, b, c)
}

func (asm *asmprogram) emitSge(a, b, c *data) {
	asm.emitArity3(sge, a, b, c)
}

func (asm *asmprogram) emitSgt(a, b, c *data) {
	asm.emitArity3(sgt, a, b, c)
}

func (asm *asmprogram) emitSle(a, b, c *data) {
	asm.emitArity3(sle, a, b, c)
}

func (asm *asmprogram) emitSlt(a, b, c *data) {
	asm.emitArity3(slt, a, b, c)
}

func (asm *asmprogram) emitSeq(a, b, c *data) {
	asm.emitArity3(seq, a, b, c)
}

func (asm *asmprogram) emitSne(a, b, c *data) {
	asm.emitArity3(sne, a, b, c)
}

func (asm *asmprogram) emitJ(a *data) {
	asm.emitArity1(j, a)
}

func (asm *asmprogram) emitBnez(a, b *data) {
	asm.emitArity2(bnez, a, b)
}

func (asm *asmprogram) emitBeqz(a, b *data) {
	asm.emitArity2(beqz, a, b)
}

func (asm *asmprogram) emitSin(a, b *data) {
	asm.emitArity2(sin, a, b)
}

func (asm *asmprogram) emitCos(a, b *data) {
	asm.emitArity2(cos, a, b)
}

func (asm *asmprogram) emitTan(a, b *data) {
	asm.emitArity2(tan, a, b)
}
