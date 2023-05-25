package ir

import (
	"fmt"
	"strings"
)

// An IR Program is a list of IR instructions
type Program struct {
	instructions []IRInstruction
}

func NewProgram() *Program {
	return &Program{instructions: []IRInstruction{}}
}

func (p *Program) Emit(i IRInstruction) {
	p.instructions = append(p.instructions, i)
}

func (p *Program) Get() []IRInstruction {
	return p.instructions
}

func (p *Program) String() string {
	var b strings.Builder
	for _, instruction := range p.instructions {
		fmt.Fprintln(&b, instruction.String())
	}

	return b.String()
}
