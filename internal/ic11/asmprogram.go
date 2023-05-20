package ic11

import (
	"fmt"
	"strings"
)

type asmprogram struct {
	program           []instruction
	labelOptimization bool
	labelMap          map[string]int
	labelCount        int
}

func newASMProgram(labelOptimization bool) *asmprogram {
	return &asmprogram{program: []instruction{}, labelCount: 0, labelOptimization: labelOptimization, labelMap: make(map[string]int)}
}

func (asm *asmprogram) print() (string, error) {
	strArr := []string{}
	for _, instr := range asm.program {
		if asm.labelOptimization {
			err := instr.ReplaceLabels(asm.labelMap)
			if err != nil {
				return "", err
			}
		}
		strArr = append(strArr, instr.ToString())
	}

	return strings.Join(strArr, "\n"), nil
}

func (asm *asmprogram) getUniqueLabel() string {
	lbl := fmt.Sprintf("L_%d", asm.labelCount)
	asm.labelCount = asm.labelCount + 1

	return lbl
}

func (asm *asmprogram) emitLabel(lbl string) {
	if asm.labelOptimization {
		curLine := len(asm.program)
		asm.labelMap[lbl] = curLine
	} else {
		instruction := &label{label: lbl}
		asm.program = append(asm.program, instruction)
	}
}

func (asm *asmprogram) emitArityN(cmd string, args ...*data) {
	instruction := newInstructionN(cmd, args...)
	asm.program = append(asm.program, instruction)
}

func (asm *asmprogram) emitMove(a, b *data) {
	asm.emitArityN(move, a, b)
}

func (asm *asmprogram) emitAdd(a, b, c *data) {
	asm.emitArityN(add, a, b, c)
}

func (asm *asmprogram) emitSub(a, b, c *data) {
	asm.emitArityN(sub, a, b, c)
}

func (asm *asmprogram) emitMul(a, b, c *data) {
	asm.emitArityN(mul, a, b, c)
}

func (asm *asmprogram) emitDiv(a, b, c *data) {
	asm.emitArityN(div, a, b, c)
}

func (asm *asmprogram) emitSge(a, b, c *data) {
	asm.emitArityN(sge, a, b, c)
}

func (asm *asmprogram) emitSgt(a, b, c *data) {
	asm.emitArityN(sgt, a, b, c)
}

func (asm *asmprogram) emitSle(a, b, c *data) {
	asm.emitArityN(sle, a, b, c)
}

func (asm *asmprogram) emitSlt(a, b, c *data) {
	asm.emitArityN(slt, a, b, c)
}

func (asm *asmprogram) emitSeq(a, b, c *data) {
	asm.emitArityN(seq, a, b, c)
}

func (asm *asmprogram) emitSne(a, b, c *data) {
	asm.emitArityN(sne, a, b, c)
}

func (asm *asmprogram) emitJ(a *data) {
	asm.emitArityN(j, a)
}

func (asm *asmprogram) emitBnez(a, b *data) {
	asm.emitArityN(bnez, a, b)
}

func (asm *asmprogram) emitBeqz(a, b *data) {
	asm.emitArityN(beqz, a, b)
}

func (asm *asmprogram) emitSin(a, b *data) {
	asm.emitArityN(sin, a, b)
}

func (asm *asmprogram) emitCos(a, b *data) {
	asm.emitArityN(cos, a, b)
}

func (asm *asmprogram) emitTan(a, b *data) {
	asm.emitArityN(tan, a, b)
}

func (asm *asmprogram) emitMod(a, b, c *data) {
	asm.emitArityN(mod, a, b, c)
}

func (asm *asmprogram) emitL(a, b, c *data) {
	asm.emitArityN(l, a, b, c)
}

func (asm *asmprogram) emitLb(a, b, c, d *data) {
	asm.emitArityN(lb, a, b, c, d)
}

func (asm *asmprogram) emitS(a, b, c *data) {
	asm.emitArityN(s, a, b, c)
}

func (asm *asmprogram) emitSb(a, b, c *data) {
	asm.emitArityN(sb, a, b, c)
}

func (asm *asmprogram) emitYield() {
	asm.emitArityN(yield)
}

func (asm *asmprogram) emitBge(a, b, c *data) {
	asm.emitArityN(bge, a, b, c)
}

func (asm *asmprogram) emitBgt(a, b, c *data) {
	asm.emitArityN(bgt, a, b, c)
}

func (asm *asmprogram) emitBle(a, b, c *data) {
	asm.emitArityN(ble, a, b, c)
}

func (asm *asmprogram) emitBlt(a, b, c *data) {
	asm.emitArityN(blt, a, b, c)
}

func (asm *asmprogram) emitBeq(a, b, c *data) {
	asm.emitArityN(beq, a, b, c)
}

func (asm *asmprogram) emitBne(a, b, c *data) {
	asm.emitArityN(bne, a, b, c)
}

func (asm *asmprogram) emitAnd(a, b, c *data) {
	asm.emitArityN(and, a, b, c)
}

func (asm *asmprogram) emitOr(a, b, c *data) {
	asm.emitArityN(or, a, b, c)
}

func (asm *asmprogram) emitXor(a, b, c *data) {
	asm.emitArityN(xor, a, b, c)
}

func (asm *asmprogram) emitNor(a, b, c *data) {
	asm.emitArityN(nor, a, b, c)
}

func (asm *asmprogram) emitAbs(a, b *data) {
	asm.emitArityN(abs, a, b)
}

func (asm *asmprogram) emitAcos(a, b *data) {
	asm.emitArityN(acos, a, b)
}

func (asm *asmprogram) emitAsin(a, b *data) {
	asm.emitArityN(asin, a, b)
}

func (asm *asmprogram) emitAtan(a, b *data) {
	asm.emitArityN(atan, a, b)
}

func (asm *asmprogram) emitCeil(a, b *data) {
	asm.emitArityN(ceil, a, b)
}

func (asm *asmprogram) emitExp(a, b *data) {
	asm.emitArityN(exp, a, b)
}

func (asm *asmprogram) emitFloor(a, b *data) {
	asm.emitArityN(floor, a, b)
}

func (asm *asmprogram) emitLog(a, b *data) {
	asm.emitArityN(log, a, b)
}

func (asm *asmprogram) emitMax(a, b, c *data) {
	asm.emitArityN(max, a, b, c)
}

func (asm *asmprogram) emitMin(a, b, c *data) {
	asm.emitArityN(min, a, b, c)
}

func (asm *asmprogram) emitSqrt(a, b *data) {
	asm.emitArityN(sqrt, a, b)
}

func (asm *asmprogram) emitRound(a, b *data) {
	asm.emitArityN(round, a, b)
}

func (asm *asmprogram) emitTrunc(a, b *data) {
	asm.emitArityN(trunc, a, b)
}

func (asm *asmprogram) emitRand(a *data) {
	asm.emitArityN(rand, a)
}

func (asm *asmprogram) emitSleep(a *data) {
	asm.emitArityN(sleep, a)
}

func (asm *asmprogram) emitAlias(a, b *data) {
	asm.emitArityN(alias, a, b)
}
