//nolint:revive
package ir

import (
	"fmt"
	"strconv"
	"strings"
)

// IC11 IR is a TAC: Three Address Code

type IRVar string
type IRStringConst string
type IRIntConst int64
type IRFloatConst float64
type IRLabelType string

// Compound helper types

type IRLiteralType struct {
	// Only one of these can be set
	valueInt    *IRIntConst
	valueString *IRStringConst
	valueFloat  *IRFloatConst
	valueLabel  *IRLabelType
}

func NewStringLiteral(s string) *IRLiteralType {
	sc := IRStringConst(s)
	return &IRLiteralType{valueString: &sc}
}

func NewIntLiteral(i int64) *IRLiteralType {
	ic := IRIntConst(i)
	return &IRLiteralType{valueInt: &ic}
}

func NewFloatLiteral(f float64) *IRLiteralType {
	fc := IRFloatConst(f)
	return &IRLiteralType{valueFloat: &fc}
}
func NewLabelLiteral(l string) *IRLiteralType {
	lc := IRLabelType(l)
	return &IRLiteralType{valueLabel: &lc}
}

func (lit IRLiteralType) String() string {
	if lit.valueInt != nil {
		return fmt.Sprintf("%d", *lit.valueInt)
	}

	if lit.valueFloat != nil {
		return strconv.FormatFloat(float64(*lit.valueFloat), 'f', -1, 64)
	}

	if lit.valueString != nil {
		return string(*lit.valueString)
	}

	if lit.valueLabel != nil {
		return string(*lit.valueLabel)
	}

	panic("empty IRLiteralType")
}

type IRLiteralOrVar struct {
	// Only one of these can be set
	lit *IRLiteralType
	v   *IRVar
}

// keyword (variable-assigned-to-type type) function-name (arguments-passed) return-type {implementation}
func NewLiteralOrVarLiteral(lit IRLiteralType) IRLiteralOrVar {
	return IRLiteralOrVar{lit: &lit}
}

func NewLiteralOrVarVar(v IRVar) IRLiteralOrVar {
	return IRLiteralOrVar{v: &v}
}

func (litOrVar IRLiteralOrVar) IsLiteral() bool {
	return litOrVar.lit != nil
}

func (litOrVar IRLiteralOrVar) IsVar() bool {
	return litOrVar.v != nil
}

func (litOrVar IRLiteralOrVar) String() string {
	if litOrVar.lit != nil {
		return litOrVar.lit.String()
	}

	if litOrVar.v != nil {
		return string(*litOrVar.v)
	}

	panic("empty irliteralOrVar")
}

func (litOrVar IRLiteralOrVar) GetVariable() IRVar {
	if litOrVar.v != nil {
		return *litOrVar.v
	}
	panic("literal is not variable")
}

// All instructions must implement the following interface
type IRInstruction interface {
	String() string
}

// Instructions

type IRAssignLiteral struct {
	Assignee IRVar
	ValueVar IRLiteralType
}

type IRAssignVar struct {
	Assignee IRVar
	ValueVar IRVar
}

type IRAssignBinary struct {
	Assignee IRVar
	L        IRVar
	R        IRVar
	// Op can be one of '+', '-', '*', '/', '==', '<', '&&', '||'
	Op string
}

type IRLabel struct {
	Label IRLabelType
}
type IRGoto struct {
	Label IRLabelType
}

type IRIfZ struct {
	Cond  IRVar
	Label IRLabelType
}

type IRBuiltinCallVoid struct {
	BuiltinName string
	Params      []IRLiteralOrVar
}

type IRBuiltinCallRet struct {
	BuiltinName string
	Params      []IRLiteralOrVar
	Ret         IRVar
}

// All of the IR instructions implement String() to assist with debugging

func (ir IRAssignBinary) String() string {
	return fmt.Sprintf("%s = %s %s %s;", ir.Assignee, ir.L, ir.Op, ir.R)
}

func (ir IRLabel) String() string {
	return fmt.Sprintf("%s:", ir.Label)
}

func (ir IRAssignLiteral) String() string {
	return fmt.Sprintf("%v = %v;", ir.Assignee, ir.ValueVar)
}

func (ir IRAssignVar) String() string {
	return fmt.Sprintf("%v = %v;", ir.Assignee, ir.ValueVar)
}

func (ir IRGoto) String() string {
	return fmt.Sprintf("Goto %s;", ir.Label)
}

func (ir IRIfZ) String() string {
	return fmt.Sprintf("IfZ %s Goto %s;", ir.Cond, ir.Label)
}

func (ir IRBuiltinCallVoid) String() string {
	strParams := []string{}
	for _, param := range ir.Params {
		strParams = append(strParams, param.String())
	}
	return fmt.Sprintf("Bcall %s %s;", ir.BuiltinName, strings.Join(strParams, " "))
}

func (ir IRBuiltinCallRet) String() string {
	strParams := []string{}
	for _, param := range ir.Params {
		strParams = append(strParams, param.String())
	}
	return fmt.Sprintf("%s = Bcall %s %s;", ir.Ret, ir.BuiltinName, strings.Join(strParams, " "))
}
