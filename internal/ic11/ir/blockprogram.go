package ir

import (
	"strings"

	"github.com/greg2010/ic11c/internal/stack"
)

// An IR BlockProgram is a chain of basic blocks.
// Each basic block can contain at most one label (at the entry point),
// and at most one control flow instruction (as the last instruction in the block).
type BlockProgram struct {
	blockCount  int
	entrypoint  *BasicBlock
	current     *BasicBlock
	blockLabels map[IRLabelType]*BasicBlock
	blocks      []*BasicBlock
}

func NewBlockProgram(program *Program) *BlockProgram {
	ir := BlockProgram{
		blockCount:  0,
		entrypoint:  nil,
		current:     nil,
		blockLabels: make(map[IRLabelType]*BasicBlock),
		blocks:      []*BasicBlock{},
	}
	ir.process(program)

	return &ir
}

func (bp *BlockProgram) String() string {
	var b strings.Builder
	preorder := bp.FifoSort()
	for _, block := range preorder {
		b.WriteString(block.Print())
	}

	return b.String()
}

func (bp *BlockProgram) FifoSort() []*BasicBlock {
	arr := []*BasicBlock{}
	c := make(chan *BasicBlock, len(bp.blocks))

	seen := make(map[int]bool)
	c <- bp.entrypoint

	for len(c) > 0 {
		cur := <-c
		if _, found := seen[cur.ID]; found {
			continue
		}
		arr = append(arr, cur)
		seen[cur.ID] = true
		for _, elem := range cur.next {
			c <- elem
		}
	}

	return arr
}

func (bp *BlockProgram) PreorderBlockSort() []*BasicBlock {
	seen := make(map[int]bool)
	arr := []*BasicBlock{}
	s := stack.New(bp.entrypoint)
	for s.Len() > 0 {
		cur := s.Pop()
		if _, found := seen[cur.ID]; found {
			continue
		}
		arr = append(arr, cur)
		s.PushReverse(cur.next)
		seen[cur.ID] = true
	}

	return arr
}

func TraverseBasicBlockPreorder(bb *BasicBlock) []*BasicBlock {
	soFar := []*BasicBlock{bb}
	for _, child := range bb.next {
		soFar = append(soFar, TraverseBasicBlockPreorder(child)...)
	}

	return soFar
}

func (bp *BlockProgram) process(p *Program) {
	for _, i := range p.Get() {
		bp.emit(i)
	}
}

func (bp *BlockProgram) newBasicBlock(label *IRLabelType) *BasicBlock {
	bb := newBasicBlock(bp.blockCount)
	bp.blockCount = bp.blockCount + 1
	if label != nil {
		bp.blockLabels[*label] = bb
	}
	return bb
}

func (bp *BlockProgram) setCurrent(bb *BasicBlock) {
	bp.current = bb
	bp.blocks = append(bp.blocks, bb)
}

func (bp *BlockProgram) linkToCurrent(bb *BasicBlock) {
	if bp.current != nil {
		bb.addPrev(bp.current)
		bp.current.addNext(bb)
	}
}

func (bp *BlockProgram) findOrCreateBlockWithLabel(label IRLabelType) *BasicBlock {
	if bb, found := bp.blockLabels[label]; found {
		return bb
	}

	return bp.newBasicBlock(&label)
}

func (bp *BlockProgram) emitToCurrentBlock(instr IRInstruction) {
	cur := bp.current
	if cur == nil {
		cur = bp.newBasicBlock(nil)
		bp.setCurrent(cur)
	}
	cur.emit(instr)
	if bp.entrypoint == nil {
		bp.entrypoint = cur
	}
}

func (bp *BlockProgram) emit(instr IRInstruction) {
	switch i := instr.(type) {
	case IRLabel:
		// IRLabel indicates a new block started
		block := bp.findOrCreateBlockWithLabel(i.Label)
		bp.linkToCurrent(block)
		// Emit label instruction and set current to the block
		bp.setCurrent(block)
		block.emit(i)
	case IRGoto:
		// IRGoto terminates current block
		bp.emitToCurrentBlock(instr)
		next := bp.findOrCreateBlockWithLabel(i.Label)
		bp.linkToCurrent(next)
		// We just emitted goto, next non-label instruction won't belong to a block
		bp.current = nil
	case IRIfZ:
		bp.emitToCurrentBlock(i)
		// We just emitted ifZ, our next blocks are wherever goto leads and next sequential block
		// that we create here without a label
		nextSeq := bp.newBasicBlock(nil)
		nextGoto := bp.findOrCreateBlockWithLabel(i.Label)
		bp.linkToCurrent(nextSeq)
		bp.linkToCurrent(nextGoto)
		bp.setCurrent(nextSeq)
	default:
		bp.emitToCurrentBlock(instr)
	}
}
