package assembler

import (
	"fmt"
	"strings"
)

// A MIPS MIPSProgram is a list of MIPS instructions
type MIPSProgram struct {
	instructions []mipsInstruction
}

func NewMipsProgram() *MIPSProgram {
	return &MIPSProgram{instructions: []mipsInstruction{}}
}

func (p *MIPSProgram) Emit(i mipsInstruction) {
	p.instructions = append(p.instructions, i)
}

func (p *MIPSProgram) Get() []mipsInstruction {
	return p.instructions
}

func (p *MIPSProgram) String() string {
	var b strings.Builder
	for _, instruction := range p.instructions {
		fmt.Fprintln(&b, instruction.String())
	}

	return b.String()
}
