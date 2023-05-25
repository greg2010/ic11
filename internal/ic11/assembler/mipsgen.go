package assembler

import (
	"errors"
	"fmt"

	"github.com/greg2010/ic11c/internal/ic11/ir"
	"github.com/greg2010/ic11c/internal/ic11/regassign"
)

var (
	ErrUnknownIRInstruction          = errors.New("unknown IR instruction")
	ErrInvalidIRInstructionArguments = errors.New("invalid IR instruction argumetns")
)

type MipsAssembler struct {
	registerAssigner regassign.RegisterAssigner
	program          *MIPSProgram
}

func New(program *ir.Program, reg regassign.RegisterAssigner) (*MipsAssembler, error) {
	mp := NewMipsProgram()
	assembler := &MipsAssembler{registerAssigner: reg, program: mp}
	err := assembler.compile(program)
	if err != nil {
		return nil, err
	}
	return assembler, nil
}

// compile iterates over IR program and emits corresponding MIPS instructions to MipsProgram
func (ma *MipsAssembler) compile(irProgram *ir.Program) error {
	for _, irInstr := range irProgram.Get() {
		switch i := irInstr.(type) {
		case ir.IRAssignLiteral:
			ma.emitAssignLiteral(i)
		case ir.IRAssignVar:
			ma.emitAsssignVar(i)
		case ir.IRAssignBinary:
			ma.emitAsssignBinary(i)
		case ir.IRLabel:
			ma.emitLabel(i)
		case ir.IRGoto:
			ma.emitGoto(i)
		case ir.IRIfZ:
			ma.emitIfZ(i)
		case ir.IRBuiltinCallVoid:
			ma.emitBuiltinCallVoid(i)
		case ir.IRBuiltinCallRet:
			ma.emitBuiltinCallRet(i)
		default:
			return ErrUnknownIRInstruction
		}
	}

	return nil
}

// emitAssignLiteral emits MIPS code that corresponds to IRAssignLiteral
// example:
// t0 = 0;
// ->
// move r0 0
func (ma *MipsAssembler) emitAssignLiteral(irInstr ir.IRAssignLiteral) {
	regNumber := ma.registerAssigner.GetRegister(irInstr.Assignee)
	regName := mipsRegisterName(regNumber)
	ma.program.Emit(newInstructionN(move, regName, irInstr.ValueVar.String()))
}

// emitAsssignVar emits MIPS code that corresponds to IRAssignVar
// example:
// t0 = a;
// ->
// move r1 r0
func (ma *MipsAssembler) emitAsssignVar(irInstr ir.IRAssignVar) {
	panic("unimplmented")
}

// emitAsssignBinary emits MIPS code that corresponds to IRAssignBinary
// example:
// t0 = a + 1;
// ->
// add r1 r0 1
func (ma *MipsAssembler) emitAsssignBinary(irInstr ir.IRAssignBinary) error {
	switch irInstr.Op {
	case "+":
		return nil
	case "-":
		return nil
	case "/":
		return nil
	case "*":
		return nil
	case "||":
		return nil
	case "&&":
		return nil
	case "==":
		return nil
	case "<":
		return nil
	default:
		return ErrInvalidIRInstructionArguments
	}
}

// emitLabel emits MIPS code that corresponds to IRLabel
// example:
// _L1:
// ->
// _L1:
func (ma *MipsAssembler) emitLabel(irInstr ir.IRLabel) {
	panic("unimplmented")
}

// emitGoto emits MIPS code that corresponds to IRGoto
// example:
// Goto _L0;
// ->
// j _L0
func (ma *MipsAssembler) emitGoto(irInstr ir.IRGoto) {
	panic("unimplmented")
}

// emitIfZ emits MIPS code that corresponds to emitIfZ
// example:
// IfZ t0 Goto _L0;
// ->
// beqz r0 _L0
func (ma *MipsAssembler) emitIfZ(irInstr ir.IRIfZ) {
	panic("unimplmented")
}

// emitBuiltinCallVoid emits MIPS code that corresponds to IRBuiltinCallVoid
// example:
// Bcall store d0 "Vertical" t0;
// ->
// beqz r0 _L0
func (ma *MipsAssembler) emitBuiltinCallVoid(irInstr ir.IRBuiltinCallVoid) {
	panic("unimplmented")
}

// emitIfZ emits MIPS code that corresponds to IRBuiltinCallRet
// example:
// t0 = Bcall load d2 On;
// ->
// l r0 d2 On
func (ma *MipsAssembler) emitBuiltinCallRet(irInstr ir.IRBuiltinCallRet) {
	panic("unimplmented")
}

// Helpers

func mipsRegisterName(registerNumber int) string {
	return fmt.Sprintf("r%d", registerNumber)
}
