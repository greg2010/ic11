package assembler

import (
	"errors"
	"fmt"

	"github.com/greg2010/ic11c/internal/ic11/ir"
	"github.com/greg2010/ic11c/internal/ic11/regassign"
)

var (
	ErrUnknownIRInstruction          = errors.New("unknown IR instruction")
	ErrInvalidIRInstructionArguments = errors.New("invalid IR instruction arguments")
	ErrInvalidParameterType          = errors.New("invalid parameter type")
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

func (ma *MipsAssembler) String() string {
	return ma.program.String()
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
			err := ma.emitAsssignBinary(i)
			if err != nil {
				return err
			}
		case ir.IRLabel:
			ma.emitLabel(i)
		case ir.IRGoto:
			ma.emitGoto(i)
		case ir.IRIfZ:
			ma.emitIfZ(i)
		case ir.IRBuiltinCallVoid:
			err := ma.emitBuiltinCallVoid(i)
			if err != nil {
				return err
			}
		case ir.IRBuiltinCallRet:
			err := ma.emitBuiltinCallRet(i)
			if err != nil {
				return err
			}
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
	regNumber := ma.registerAssigner.GetRegister(irInstr.Assignee)
	regName := mipsRegisterName(regNumber)
	regNumber2 := ma.registerAssigner.GetRegister(irInstr.ValueVar)
	regName2 := mipsRegisterName(regNumber2)
	ma.program.Emit(newInstructionN(move, regName, regName2))
}

// emitAsssignBinary emits MIPS code that corresponds to IRAssignBinary
// example:
// t0 = a + b;
// ->
// add r1 r0 r2
func (ma *MipsAssembler) emitAsssignBinary(irInstr ir.IRAssignBinary) error {
	regNumber := ma.registerAssigner.GetRegister(irInstr.Assignee)
	regName := mipsRegisterName(regNumber)
	regNumber2 := ma.registerAssigner.GetRegister(irInstr.L)
	regName2 := mipsRegisterName(regNumber2)
	regNumber3 := ma.registerAssigner.GetRegister(irInstr.R)
	regName3 := mipsRegisterName(regNumber3)

	var opType string

	switch irInstr.Op {
	case "+":
		opType = add
	case "-":
		opType = sub
	case "/":
		opType = div
	case "*":
		opType = mul
	case "||":
		opType = or
	case "&&":
		opType = and
	case "==":
		opType = seq
	case "<":
		opType = slt
	default:
		return ErrInvalidIRInstructionArguments
	}
	ma.program.Emit(newInstructionN(opType, regName, regName2, regName3))
	return nil
}

// emitLabel emits MIPS code that corresponds to IRLabel
// example:
// _L1:
// ->
// _L1:
func (ma *MipsAssembler) emitLabel(irInstr ir.IRLabel) {
	labelName := string(irInstr.Label)
	ma.program.Emit(mipsLabel{labelName})
}

// emitGoto emits MIPS code that corresponds to IRGoto
// example:
// Goto _L0;
// ->
// j _L0
func (ma *MipsAssembler) emitGoto(irInstr ir.IRGoto) {
	labelName := string(irInstr.Label)
	ma.program.Emit(newInstructionN(j, labelName))
}

// emitIfZ emits MIPS code that corresponds to emitIfZ
// example:
// IfZ t0 Goto _L0;
// ->
// beqz r0 _L0
func (ma *MipsAssembler) emitIfZ(irInstr ir.IRIfZ) {
	labelName := string(irInstr.Label)
	regNumber := ma.registerAssigner.GetRegister(irInstr.Cond)
	regName := mipsRegisterName(regNumber)
	ma.program.Emit(newInstructionN(beqz, regName, labelName))
}

// emitBuiltinCallVoid emits MIPS code that corresponds to IRBuiltinCallVoid
// example:
// Bcall store d0 "Vertical" t0;
// ->
// beqz r0 _L0
func (ma *MipsAssembler) emitBuiltinCallVoid(irInstr ir.IRBuiltinCallVoid) error {
	paramSlice := ma.sliceParameters(irInstr.Params)
	var opType string
	switch irInstr.BuiltinName {
	case "store":
		opType = s
	case "store_batch":
		opType = sb
	default:
		return ErrInvalidIRInstructionArguments
	}
	ma.program.Emit(newInstructionN(opType, paramSlice...))
	return nil
}

// emitIfZ emits MIPS code that corresponds to IRBuiltinCallRet
// example:
// t0 = Bcall load d2 On;
// ->
// l r0 d2 On
func (ma *MipsAssembler) emitBuiltinCallRet(irInstr ir.IRBuiltinCallRet) error {
	regNumber := ma.registerAssigner.GetRegister(irInstr.Ret)
	regName := mipsRegisterName(regNumber)
	argSlice := []string{regName}
	paramSlice := ma.sliceParameters(irInstr.Params)

	argSlice = append(argSlice, paramSlice...)
	var opType string

	switch irInstr.BuiltinName {
	case "load":
		opType = l
	case "load_batch":
		opType = lb
	case "load_reagent":
		opType = lr
	case "rand":
		opType = rand
	default:
		return ErrInvalidIRInstructionArguments
	}
	ma.program.Emit(newInstructionN(opType, argSlice...))
	return nil
}

// Helpers

func mipsRegisterName(registerNumber int) string {
	return fmt.Sprintf("r%d", registerNumber)
}

func (ma *MipsAssembler) sliceParameters(params []ir.IRLiteralOrVar) []string {
	slice := []string{}

	for _, element := range params {
		if element.IsVar() {
			regNumber := ma.registerAssigner.GetRegister(element.GetVariable())
			regName := mipsRegisterName(regNumber)
			slice = append(slice, regName)
		} else {
			slice = append(slice, element.String())
		}
	}
	return slice
}
