package assembler

import (
	"fmt"
	"strings"
)

type mipsInstruction interface {
	String() string
}

type mipsLabel struct {
	label string
}

func (l mipsLabel) String() string {
	return fmt.Sprintf("%s:", l.label)
}

type mipsInstructionN struct {
	cmd  string
	args []string
}

func newInstructionN(cmd string, args ...string) *mipsInstructionN {
	return &mipsInstructionN{cmd, args}
}

func (in mipsInstructionN) String() string {
	if len(in.args) == 0 {
		return in.cmd
	}

	return fmt.Sprintf("%s %s", in.cmd, strings.Join(in.args, " "))
}
