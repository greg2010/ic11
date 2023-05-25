package ir

import (
	"fmt"
	"strings"
)

// A basic block is defined as a sequence of instructions that has only one entrypoint and one exit
type BasicBlock struct {
	prev    []*BasicBlock
	next    []*BasicBlock
	program *Program
	ID      int
}

func newBasicBlock(blockID int) *BasicBlock {
	return &BasicBlock{
		prev:    []*BasicBlock{},
		next:    []*BasicBlock{},
		program: NewProgram(),
		ID:      blockID,
	}
}

func (bb *BasicBlock) String() string {
	return fmt.Sprint(bb.ID)
}

func (bb *BasicBlock) Print() string {
	var b strings.Builder
	//fmt.Fprintf(&b, "%s -> (%d) -> %s\n", bb.prev, bb.ID, bb.next)
	b.WriteString(bb.program.String())
	return b.String()
}

func (bb *BasicBlock) addPrev(prev *BasicBlock) {
	bb.prev = append(bb.prev, prev)
}

func (bb *BasicBlock) addNext(next *BasicBlock) {
	bb.next = append(bb.next, next)
}

func (bb *BasicBlock) emit(instr IRInstruction) {
	bb.program.Emit(instr)
}
