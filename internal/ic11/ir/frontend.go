package ir

import (
	"errors"
	"fmt"

	"github.com/greg2010/ic11c/internal/ic11/parser"
)

var ErrInvalidFunctionCall = errors.New("invalid function call")
var ErrInvalidState = errors.New("parser produced invalid state")
var ErrMainFuncParameters = errors.New("main function cannot have parameters")

type Frontend struct {
	varCount   int
	labelCount int
	program    *Program
}

func NewFrontend(ast *parser.AST) (*Frontend, error) {
	ir := Frontend{
		varCount:   0,
		labelCount: 0,
		program:    NewProgram(),
	}
	err := ir.compile(ast)
	if err != nil {
		return nil, err
	}

	return &ir, nil
}

func (ir *Frontend) Get() *Program {
	return ir.program
}

func (ir *Frontend) String() string {
	return ir.program.String()
}

func (ir *Frontend) newVar() IRVar {
	str := fmt.Sprintf("t%d", ir.varCount)
	ir.varCount = ir.varCount + 1
	return IRVar(str)
}

func (ir *Frontend) newLabel() IRLabelType {
	str := fmt.Sprintf("_L%d", ir.labelCount)
	ir.labelCount = ir.labelCount + 1
	return IRLabelType(str)
}

func (ir *Frontend) emit(instr IRInstruction) {
	ir.program.Emit(instr)
}
