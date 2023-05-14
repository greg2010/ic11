package ic11

import (
	"errors"
	"math"
)

type funcCompiler struct {
	funAST         *FunDec
	varRegisterMap map[string]*register
	globalConsts   map[string]float64
	globalDevices  map[string]string
	tempRegisters  []*register
	asmProgram     *asmprogram
	conf           CompilerOpts
}

func newFuncCompiler(
	asmProgram *asmprogram,
	conf CompilerOpts,
	funAST *FunDec,
	globalConsts map[string]float64,
	globalDevices map[string]string,
) (*funcCompiler, error) {
	regsNeeded := len(funAST.Parameters) + len(funAST.FunBody.Locals)

	if regsNeeded > 15 {
		return nil, CompilerError{Pos: &funAST.FunBody.Pos, Err: ErrTooManyVars}
	}

	varRegisterMap := make(map[string]*register)
	for i, v := range funAST.FunBody.Locals {
		varRegisterMap[v.ScalarDec.Name] = newRegister(i, false)
		varRegisterMap[v.ScalarDec.Name].allocate()
	}

	tempRegs := []*register{}
	for i := regsNeeded; i < 15; {
		tempRegs = append(tempRegs, newRegister(i, true))
		i++
	}

	return &funcCompiler{
		funAST:         funAST,
		varRegisterMap: varRegisterMap,
		globalConsts:   globalConsts,
		globalDevices:  globalDevices,
		tempRegisters:  tempRegs,
		asmProgram:     asmProgram,
		conf:           conf,
	}, nil
}

func (fc *funcCompiler) allocateTempRegister() (*register, error) {
	for _, reg := range fc.tempRegisters {
		if !reg.allocated {
			reg.allocate()
			return reg, nil
		}
	}

	return nil, CompilerError{Err: ErrOutOfTempRegisters}
}

// compileExpr compiles the Expr type. If out is specified, will try to use it instead of a temp register
//
//	note that data is not guaranteed to be put in out. the return of the function should be used instead
func (fc *funcCompiler) compileExpr(expr *Expr, out *register) (*data, error) {
	if expr.Primary != nil {
		return fc.compilePrimary(expr.Primary, out)
	}

	if expr.Unary != nil {
		return fc.compileUnary(expr.Unary, out)
	}
	if expr.Binary != nil {
		return fc.compileBinary(expr.Binary, out)
	}

	return nil, CompilerError{Err: ErrInvalidState, Pos: &expr.Pos}
}

func (fc *funcCompiler) compileUnary(unary *Unary, out *register) (*data, error) {
	target, err := fc.compilePrimary(unary.RHS, out)
	if err != nil {
		return nil, err
	}

	if fc.conf.PrecomputeExprs && target.isFloatValue() {
		if unary.Op == "-" {
			return newNumData(unary.RHS.Number.Number * -1), nil
		}
	}

	// maybe needs fixing

	var tempReg *register
	if out != nil {
		tempReg = out
	} else if !target.isTemporaryRegister() {
		tempReg, err = fc.allocateTempRegister()
		if err != nil {
			return nil, err
		}
	}

	tempData := newRegisterData(tempReg)

	fc.asmProgram.emitMove(tempData, target)

	if unary.Op == "-" {
		fc.asmProgram.emitSub(tempData, newNumData(0), tempData)
		return tempData, nil
	}

	return nil, CompilerError{Err: ErrInvalidState, Pos: &unary.Pos}
}

func (fc *funcCompiler) compileBinary(binary *Binary, out *register) (*data, error) {
	// Only passing out to left side if right side is a number
	var leftOut *register
	if binary.RHS.Number != nil {
		leftOut = out
	}
	leftData, err := fc.compilePrimary(binary.LHS, leftOut)
	if err != nil {
		return nil, err
	}

	rightData, err := fc.compilePrimary(binary.RHS, out)
	if err != nil {
		return nil, err
	}

	if fc.conf.PrecomputeExprs && leftData.isFloatValue() && rightData.isFloatValue() {
		v, err := computeBinop(leftData.floatValue, rightData.floatValue, binary.Op)
		if err != nil {
			return nil, err
		}

		return newNumData(v), nil
	}

	var targetReg *register
	if out != nil {
		targetReg = out
	} else if leftData.isTemporaryRegister() {
		targetReg = leftData.register
	} else if rightData.isTemporaryRegister() {
		targetReg = rightData.register
	}

	// unsure if this is correct
	if out != nil && leftData.isTemporaryRegister() {
		leftData.register.deallocate()
	}

	if out != nil && rightData.isTemporaryRegister() && rightData.register.id != out.id {
		rightData.register.deallocate()
	}

	if out == nil && leftData.isTemporaryRegister() && rightData.isTemporaryRegister() {
		rightData.register.deallocate()
	}

	if targetReg == nil {
		targetReg, err = fc.allocateTempRegister()
		if err != nil {
			return nil, CompilerError{Err: err, Pos: &binary.Pos}
		}
	}

	targetData := newRegisterData(targetReg)
	switch binary.Op {
	case "+":
		fc.asmProgram.emitAdd(targetData, leftData, rightData)
		return targetData, nil
	case "-":
		fc.asmProgram.emitSub(targetData, leftData, rightData)
		return targetData, nil
	case "*":
		fc.asmProgram.emitMul(targetData, leftData, rightData)
		return targetData, nil
	case "/":
		fc.asmProgram.emitDiv(targetData, leftData, rightData)
		return targetData, nil
	case ">":
		fc.asmProgram.emitSgt(targetData, leftData, rightData)
		return targetData, nil
	case ">=":
		fc.asmProgram.emitSge(targetData, leftData, rightData)
		return targetData, nil
	case "<":
		fc.asmProgram.emitSlt(targetData, leftData, rightData)
		return targetData, nil
	case "<=":
		fc.asmProgram.emitSle(targetData, leftData, rightData)
		return targetData, nil
	case "==":
		fc.asmProgram.emitSeq(targetData, leftData, rightData)
		return targetData, nil
	case "!=":
		fc.asmProgram.emitSne(targetData, leftData, rightData)
		return targetData, nil
	default:
		return nil, CompilerError{Err: ErrInvalidState, Pos: &binary.Pos}
	}
}

func (fc *funcCompiler) getSymbolData(symbol string) *data {
	// first, search global consts
	if num, found := fc.globalConsts[symbol]; found {
		return newNumData(num)
	}

	if dev, found := fc.globalDevices[symbol]; found {
		return newDeviceData(dev)
	}

	// search assigned registers
	if reg, found := fc.varRegisterMap[symbol]; found {
		return newRegisterData(reg)
	}

	return nil
}

func (fc *funcCompiler) compilePrimary(primary *Primary, out *register) (*data, error) {
	if primary.Number != nil {
		return newNumData(primary.Number.Number), nil
	}

	if primary.Device != "" {
		return newDeviceData(primary.Device), nil
	}

	if primary.Ident != "" {
		data := fc.getSymbolData(primary.Ident)
		if data == nil {
			return nil, CompilerError{Err: ErrUnknownVar, Pos: &primary.Pos}
		}

		return data, nil
	}

	if primary.CallFunc != nil {
		return nil, errors.New("unimplemented")
		//fc.compileCallFunc(primary.CallFunc, out)
	}
	if primary.BuiltinArity0Func != nil {
		return nil, fc.compileBuiltinArity0Func(primary.BuiltinArity0Func)
	}

	if primary.BuiltinArity1Func != nil {
		return fc.compileBuiltinArity1Func(primary.BuiltinArity1Func, out)
	}

	if primary.BuiltinArity2Func != nil {
		return fc.compileBuiltinArity2Func(primary.BuiltinArity2Func, out)
	}

	if primary.BuiltinArity3Func != nil {
		return nil, fc.compileBuiltinArity3Func(primary.BuiltinArity3Func)
	}

	if primary.SubExpression != nil {
		return fc.compileExpr(primary.SubExpression, out)
	}

	if primary.HashConst != nil {
		return newHashData(primary.HashConst.Arg), nil
	}

	if primary.StringValue != "" {
		return newStringData(primary.StringValue), nil
	}

	return nil, CompilerError{Err: ErrInvalidState, Pos: &primary.Pos}
}

/*
	func (fc *funcCompiler) compileCallFunc(callFunc *CallFunc, out *register) (*data, error) {
		targetReg := out
		if targetReg == nil {
			var err error
			targetReg, err = fc.allocateTempRegister()
			if err != nil {
				return nil, CompilerError{Err: err, Pos: &callFunc.Pos}
			}
		}

		args := []*data{}
		for i, expr := range callFunc.Index {
			// When computing first arg, we are permitted to use the target register
			var tempOut *register
			if i == 0 {
				tempOut = targetReg
			}

			arg, err := fc.compileExpr(expr, tempOut)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}

		targetData := newRegisterData(targetReg)

		switch callFunc.Ident {
		case "sin":
			if len(args) < 1 {
				return nil, ErrInvalidFuncCall
			}
			fc.asmProgram.emitSin(targetData, args[0])
		case "cos":
			if len(args) < 1 {
				return nil, ErrInvalidFuncCall
			}
			fc.asmProgram.emitCos(targetData, args[0])
		case "tan":
			if len(args) < 1 {
				return nil, ErrInvalidFuncCall
			}
			fc.asmProgram.emitTan(targetData, args[0])
		default:
			return nil, ErrInvalidState
		}

		return nil, errors.New("not implemented")
	}
*/
func (fc *funcCompiler) compileBuiltinArity0Func(fun *BuiltinArity0Func) error {
	switch fun.Op {
	case "yield":
		fc.asmProgram.emitYield()
		return nil
	}

	return CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
}

func (fc *funcCompiler) compileBuiltinArity1Func(fun *BuiltinArity1Func, out *register) (*data, error) {

	arg, err := fc.compileExpr(fun.Arg, out)
	if err != nil {
		return nil, err
	}

	if fc.conf.PrecomputeExprs && arg.isFloatValue() {
		precomp, err := computeBuiltinArity1(arg.floatValue, fun.Op)
		if err != nil {
			return nil, err
		}

		return newNumData(precomp), nil
	}
	targetReg := out
	if targetReg == nil {
		var err error
		targetReg, err = fc.allocateTempRegister()
		if err != nil {
			return nil, err
		}
	}

	targetData := newRegisterData(targetReg)

	switch fun.Op {
	case "sin":
		fc.asmProgram.emitSin(targetData, arg)
	case "cos":
		fc.asmProgram.emitCos(targetData, arg)
	case "tan":
		fc.asmProgram.emitTan(targetData, arg)
	default:
		return nil, CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
	}

	return targetData, nil
}

func (fc *funcCompiler) compileBuiltinArity2Func(fun *BuiltinArity2Func, out *register) (*data, error) {
	targetReg := out
	if targetReg == nil {
		var err error
		targetReg, err = fc.allocateTempRegister()
		if err != nil {
			return nil, err
		}
	}

	arg1, err := fc.compileExpr(fun.Arg1, nil)
	if err != nil {
		return nil, err
	}

	arg2, err := fc.compileExpr(fun.Arg2, targetReg)
	if err != nil {
		return nil, err
	}

	targetData := newRegisterData(targetReg)

	switch fun.Op {
	case "load":
		// Assert arg1 is device and arg2 is int or hash
		fc.asmProgram.emitL(targetData, arg1, arg2)
	default:
		return nil, CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
	}

	return targetData, nil
}

func (fc *funcCompiler) compileBuiltinArity3Func(fun *BuiltinArity3Func) error {

	arg1, err := fc.compileExpr(fun.Arg1, nil)
	if err != nil {
		return err
	}

	arg2, err := fc.compileExpr(fun.Arg2, nil)
	if err != nil {
		return err
	}

	arg3, err := fc.compileExpr(fun.Arg3, nil)
	if err != nil {
		return err
	}

	switch fun.Op {
	case "store":
		// Assert arg1 is device and arg2 is int or hash
		fc.asmProgram.emitS(arg1, arg2, arg3)
	case "store_batch":
		// Assert arg1 is device and arg2 is int or hash
		fc.asmProgram.emitSb(arg1, arg2, arg3)
	default:
		return CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
	}

	return nil
}

func (fc *funcCompiler) compileAssignment(assignment *Assignment) error {
	targetReg, found := fc.varRegisterMap[assignment.Left]
	if !found {
		return ErrUnknownVar
	}

	value, err := fc.compileExpr(assignment.Right, targetReg)
	if err != nil {
		return err
	}

	// if value is raw data (not a register) or if registers do not match, do the move
	if value != nil && (!value.isRegister() || value.register.id != targetReg.id) {
		fc.asmProgram.emitMove(newRegisterData(targetReg), value)
		// if value is a temp register it's up to us here to free it
		if value.register != nil && value.register.temporary {
			value.register.deallocate()
		}
	}

	return nil
}

func (fc *funcCompiler) compileJump(condData *data, jumpTo *data, inverse bool) {
	if fc.conf.OptimizeJumps && condData.isFloatValue() {
		if condData.floatValue != 0 || (condData.floatValue == 0 && inverse) {
			fc.asmProgram.emitJ(jumpTo)
			return
		}
	}

	if inverse {
		fc.asmProgram.emitBnez(condData, jumpTo)
	} else {
		fc.asmProgram.emitBeqz(condData, jumpTo)
	}
	return
}

func (fc *funcCompiler) compileCondJump(op string, l, r, jumpTo *data, inverse bool) error {
	if inverse {
		switch op {
		case ">":
			op = "<"
		case ">=":
			op = "<="
		case "<":
			op = ">"
		case "<=":
			op = ">="
		case "==":
			op = "!="
		case "!=":
			op = "=="
		default:
			return CompilerError{Err: ErrInvalidState}
		}
	}

	switch op {
	case ">":
		fc.asmProgram.emitBgt(l, r, jumpTo)
	case ">=":
		fc.asmProgram.emitBge(l, r, jumpTo)
	case "<":
		fc.asmProgram.emitBlt(l, r, jumpTo)
	case "<=":
		fc.asmProgram.emitBle(l, r, jumpTo)
	case "==":
		fc.asmProgram.emitBeq(l, r, jumpTo)
	case "!=":
		fc.asmProgram.emitBne(l, r, jumpTo)
	default:
		return CompilerError{Err: ErrInvalidState}
	}

	return nil
}

func (fc *funcCompiler) compileCondition(cond *Expr, jumpTo *data, inverse bool) error {
	if fc.conf.OptimizeJumps && cond.Binary != nil && opHasCondJump(cond.Binary.Op) {
		l, err := fc.compilePrimary(cond.Binary.LHS, nil)
		if err != nil {
			return err
		}

		r, err := fc.compilePrimary(cond.Binary.RHS, nil)
		if err != nil {
			return err
		}

		if fc.conf.PrecomputeExprs && l.isFloatValue() && r.isFloatValue() {
			v, err := computeBinop(l.floatValue, r.floatValue, cond.Binary.Op)
			if err != nil {
				return err
			}

			condVal := newNumData(v)
			fc.compileJump(condVal, jumpTo, inverse)
			return nil
		}

		return fc.compileCondJump(cond.Binary.Op, l, r, jumpTo, inverse)

	} else {
		cond, err := fc.compileExpr(cond, nil)
		if err != nil {
			return err
		}

		fc.compileJump(cond, jumpTo, inverse)
	}

	return nil
}

func (fc *funcCompiler) compileIfStmt(ifStmt *IfStmt) error {
	elseLbl := newLabelData(fc.asmProgram.getUniqueLabel())
	err := fc.compileCondition(ifStmt.Condition, elseLbl, false)
	if err != nil {
		return err
	}

	err = fc.compileStmt(ifStmt.Body)
	if err != nil {
		return err
	}

	fc.asmProgram.emitLabel(elseLbl.label)
	if ifStmt.Else != nil {
		endLbl := newLabelData(fc.asmProgram.getUniqueLabel())
		err := fc.compileCondition(ifStmt.Condition, endLbl, true)
		err = fc.compileStmt(ifStmt.Else)
		if err != nil {
			return err
		}
		fc.asmProgram.emitLabel(endLbl.label)
	}

	return nil
}

func (fc *funcCompiler) compileWhileStmt(whileStmt *WhileStmt) error {
	loopStartLbl := newLabelData(fc.asmProgram.getUniqueLabel())
	loopEndLbl := newLabelData(fc.asmProgram.getUniqueLabel())
	fc.asmProgram.emitJ(loopEndLbl)
	fc.asmProgram.emitLabel(loopStartLbl.label)

	err := fc.compileStmt(whileStmt.Body)
	if err != nil {
		return err
	}

	fc.asmProgram.emitLabel(loopEndLbl.label)
	err = fc.compileCondition(whileStmt.Condition, loopStartLbl, false)
	return nil
}

func (fc *funcCompiler) compileStmt(stmt *Stmt) error {
	if stmt.Empty {
		return nil
	}

	if stmt.Block != nil {
		err := fc.compileStmts(stmt.Block)
		if err != nil {
			return err
		}

		return nil
	}

	if stmt.Assignment != nil {
		err := fc.compileAssignment(stmt.Assignment)
		if err != nil {
			return err
		}

		return nil
	}

	if stmt.IfStmt != nil {
		err := fc.compileIfStmt(stmt.IfStmt)
		if err != nil {
			return err
		}

		return nil
	}

	if stmt.WhileStmt != nil {
		err := fc.compileWhileStmt(stmt.WhileStmt)
		if err != nil {
			return err
		}

		return nil
	}

	if stmt.Expr != nil {
		_, err := fc.compileExpr(stmt.Expr, nil)
		if err != nil {
			return err
		}

		return nil
	}

	return CompilerError{Err: ErrInvalidState, Pos: &stmt.Pos}
}

func (fc *funcCompiler) compileStmts(stmts *Stmts) error {
	for _, stmt := range stmts.Stmts {
		err := fc.compileStmt(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fc *funcCompiler) compileFunc() error {
	err := fc.compileStmts(fc.funAST.FunBody.Stmts)
	if err != nil {
		return err
	}

	return nil
}

func (fc *funcCompiler) Compile() error {
	err := fc.compileFunc()
	if err != nil {
		return err
	}

	return nil
}

func computeBuiltinArity1(l float64, op string) (float64, error) {
	switch op {
	case "sin":
		return math.Sin(l), nil
	case "cos":
		return math.Cos(l), nil
	case "tan":
		return math.Tan(l), nil
	default:
		return 0, CompilerError{Err: ErrInvalidState}
	}
}

func computeBinop(l, r float64, op string) (float64, error) {
	switch op {
	case "+":
		return l + r, nil
	case "-":
		return l - r, nil
	case "*":
		return l * r, nil
	case "/":
		if r == 0 {
			return 0, ErrDiv0
		}
		return l / r, nil
	case "==":
		if l == r {
			return float64(1), nil
		}
		return float64(0), nil
	case "!=":
		if l != r {
			return float64(1), nil
		}
		return float64(0), nil
	case ">=":
		if l >= r {
			return float64(1), nil
		}
		return float64(0), nil
	case ">":
		if l > r {
			return float64(1), nil
		}
		return float64(0), nil
	case "<=":
		if l <= r {
			return float64(1), nil
		}
		return float64(0), nil
	case "<":
		if l < r {
			return float64(1), nil
		}
		return float64(0), nil
	default:
		return 0, CompilerError{Err: ErrInvalidState}
	}
}

// returns true if op has a special conditional jump instruction in the MIPS instruction set
func opHasCondJump(op string) bool {
	return op == "==" || op == "!=" || op == ">=" || op == ">" || op == "<" || op == "<="
}
