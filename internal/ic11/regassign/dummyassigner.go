package regassign

import "github.com/greg2010/ic11c/internal/ic11/ir"

type DummyAssigner struct {
	program       *ir.Program
	assignedSoFar map[ir.IRVar]int
	maxAssigned   int
}

func NewDummyAssigner(program *ir.Program) *DummyAssigner {
	return &DummyAssigner{program: program, assignedSoFar: make(map[ir.IRVar]int), maxAssigned: 0}
}

func (da *DummyAssigner) GetRegister(varName ir.IRVar) int {
	if register, found := da.assignedSoFar[varName]; found {
		return register
	}

	register := da.maxAssigned
	da.assignedSoFar[varName] = register
	da.maxAssigned = da.maxAssigned + 1
	return register
}
