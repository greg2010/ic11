package ic11

import (
	"errors"
	"math"
)

type funcCompiler struct {
	funAST         *FunDec
	varRegisterMap map[string]*register
	varValueMap    map[string]float64
	globalConsts   map[string]float64
	globalDevices  map[string]string
	registers      []*register
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

	varRegisterMap := make(map[string]*register)
	varValueMap := make(map[string]float64)
	if conf.PropagateVariables {
		for _, v := range funAST.FunBody.Locals {
			varValueMap[v.ScalarDec.Name] = 0
		}
	}

	regs := []*register{}
	for i := 0; i < 15; {
		regs = append(regs, newRegister(i))
		i++
	}

	funComp := &funcCompiler{
		funAST:         funAST,
		varRegisterMap: varRegisterMap,
		varValueMap:    varValueMap,
		globalConsts:   globalConsts,
		globalDevices:  globalDevices,
		registers:      regs,
		asmProgram:     asmProgram,
		conf:           conf,
	}

	for _, v := range funAST.FunBody.Locals {
		reg, err := funComp.allocatePermRegister()
		if err != nil {
			return nil, err
		}
		varRegisterMap[v.ScalarDec.Name] = reg
	}

	return funComp, nil
}

func (fc *funcCompiler) getVariableData(varName string) (*data, error) {
	if reg, found := fc.varRegisterMap[varName]; found {
		return newRegisterData(reg), nil
	}

	if val, found := fc.varValueMap[varName]; found {
		return newNumData(val), nil
	}

	return nil, ErrUnknownVar
}

func (fc *funcCompiler) allocateTempRegister() (*register, error) {
	for _, reg := range fc.registers {
		if !reg.allocated {
			reg.setAllocated(true)
			reg.setTemporary(true)
			return reg, nil
		}
	}

	return nil, CompilerError{Err: ErrOutOfTempRegisters}
}

func (fc *funcCompiler) allocatePermRegister() (*register, error) {
	for _, reg := range fc.registers {
		if !reg.allocated {
			reg.setAllocated(true)
			reg.setTemporary(false)
			return reg, nil
		}
	}

	return nil, CompilerError{Err: ErrOutOfTempRegisters}
}

// compileExpr compiles the Expr type
func (fc *funcCompiler) compileExpr(expr *Expr) (*data, error) {
	if expr.Primary != nil {
		return fc.compilePrimary(expr.Primary)
	}

	if expr.Unary != nil {
		return fc.compileUnary(expr.Unary)
	}
	if expr.Binary != nil {
		return fc.compileBinary(expr.Binary)
	}

	return nil, CompilerError{Err: ErrInvalidState, Pos: &expr.Pos}
}

func (fc *funcCompiler) compileUnary(unary *Unary) (*data, error) {
	target, err := fc.compilePrimary(unary.RHS)
	if err != nil {
		return nil, err
	}

	if fc.conf.PrecomputeExprs && target.isFloatValue() {
		if unary.Op == "-" {
			return newNumData(unary.RHS.Number.Number * -1), nil
		}
	}

	// maybe needs fixing

	tempReg, err := fc.allocateTempRegister()
	if err != nil {
		return nil, err
	}

	tempData := newRegisterData(tempReg)

	fc.asmProgram.emitMove(tempData, target)

	if unary.Op == "-" {
		fc.asmProgram.emitSub(tempData, newNumData(0), tempData)
		return tempData, nil
	}

	return nil, CompilerError{Err: ErrInvalidState, Pos: &unary.Pos}
}

func (fc *funcCompiler) compileBinary(binary *Binary) (*data, error) {
	leftData, err := fc.compilePrimary(binary.LHS)
	if err != nil {
		return nil, err
	}

	rightData, err := fc.compilePrimary(binary.RHS)
	if err != nil {
		return nil, err
	}

	if fc.conf.PrecomputeExprs && leftData.isFloatValue() && rightData.isFloatValue() {
		v, err := computeBinop(leftData.floatValue, rightData.floatValue, binary.Op)
		if err == nil && !math.IsNaN(v) {
			return newNumData(v), nil
		}
	}

	var targetReg *register
	if leftData.isTemporaryRegister() {
		targetReg = leftData.register
	} else if rightData.isTemporaryRegister() {
		targetReg = rightData.register
	} else {
		var err error
		targetReg, err = fc.allocateTempRegister()
		if err != nil {
			return nil, err
		}
	}

	if leftData.isTemporaryRegister() && leftData.register.id != targetReg.id {
		leftData.register.release()
	}

	if rightData.isTemporaryRegister() && rightData.register.id != targetReg.id {
		rightData.register.release()
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
	case "&&":
		fc.asmProgram.emitAnd(targetData, leftData, rightData)
		return targetData, nil
	case "||":
		fc.asmProgram.emitOr(targetData, leftData, rightData)
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

	if fc.conf.PropagateVariables {
		// search value map
		if value, found := fc.varValueMap[symbol]; found {
			return newNumData(value)
		}
	}

	return nil
}

func (fc *funcCompiler) compilePrimary(primary *Primary) (*data, error) {
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
		return fc.compileBuiltinArity0Func(primary.BuiltinArity0Func)
	}

	if primary.BuiltinArity1Func != nil {
		return fc.compileBuiltinArity1Func(primary.BuiltinArity1Func)
	}

	if primary.BuiltinArity2Func != nil {
		return fc.compileBuiltinArity2Func(primary.BuiltinArity2Func)
	}

	if primary.BuiltinArity3Func != nil {
		return fc.compileBuiltinArity3Func(primary.BuiltinArity3Func)
	}

	if primary.SubExpression != nil {
		return fc.compileExpr(primary.SubExpression)
	}

	if primary.HashConst != nil {
		if AllCompilerOpts().PrecomputeHashes {
			return newNumData(float64(ComputeHash(primary.HashConst.Arg))), nil

		} else {
			return newHashData(primary.HashConst.Arg), nil
		}
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
func (fc *funcCompiler) compileBuiltinArity0Func(fun *BuiltinArity0Func) (*data, error) {
	switch fun.Op {
	case "yield":
		fc.asmProgram.emitYield()
		return nil, nil
	case "rand":
		targetReg, err := fc.allocateTempRegister()
		if err != nil {
			return nil, err
		}
		targetData := newRegisterData(targetReg)
		fc.asmProgram.emitRand(targetData)

		return targetData, nil
	}

	return nil, CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
}

func (fc *funcCompiler) compileBuiltinArity1Func(fun *BuiltinArity1Func) (*data, error) {
	arg, err := fc.compileExpr(fun.Arg)
	if err != nil {
		return nil, err
	}

	// handling sleep separately because it does not return a value
	if fun.Op == "sleep" {
		fc.asmProgram.emitSleep(arg)
		return nil, nil

	}

	if fc.conf.PrecomputeExprs && arg.isFloatValue() {
		precomp, err := computeBuiltinArity1(arg.floatValue, fun.Op)
		if err == nil && !math.IsNaN(precomp) {
			return newNumData(precomp), nil
		}
	}

	targetReg, err := fc.allocateTempRegister()
	if err != nil {
		return nil, err
	}

	targetData := newRegisterData(targetReg)

	switch fun.Op {
	case "sin":
		fc.asmProgram.emitSin(targetData, arg)
	case "cos":
		fc.asmProgram.emitCos(targetData, arg)
	case "tan":
		fc.asmProgram.emitTan(targetData, arg)
	case "abs":
		fc.asmProgram.emitAbs(targetData, arg)
	case "acos":
		fc.asmProgram.emitAcos(targetData, arg)
	case "asin":
		fc.asmProgram.emitAsin(targetData, arg)
	case "atan":
		fc.asmProgram.emitAtan(targetData, arg)
	case "ceil":
		fc.asmProgram.emitCeil(targetData, arg)
	case "floor":
		fc.asmProgram.emitFloor(targetData, arg)
	case "log":
		fc.asmProgram.emitLog(targetData, arg)
	case "sqrt":
		fc.asmProgram.emitSqrt(targetData, arg)
	case "round":
		fc.asmProgram.emitRound(targetData, arg)
	case "trunc":
		fc.asmProgram.emitTrunc(targetData, arg)
	default:
		return nil, CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
	}

	return targetData, nil
}

func (fc *funcCompiler) compileBuiltinArity2Func(fun *BuiltinArity2Func) (*data, error) {
	arg1, err := fc.compileExpr(fun.Arg1)
	if err != nil {
		return nil, err
	}

	arg2, err := fc.compileExpr(fun.Arg2)
	if err != nil {
		return nil, err
	}

	if fc.conf.PrecomputeExprs && arg1.isFloatValue() && arg2.isFloatValue() {
		precomp, err := computeBuiltinArity2(arg1.floatValue, arg2.floatValue, fun.Op)
		if err == nil && !math.IsNaN(precomp) {
			return newNumData(precomp), nil
		}
	}

	targetReg, err := fc.allocateTempRegister()
	if err != nil {
		return nil, err
	}

	targetData := newRegisterData(targetReg)

	switch fun.Op {
	case "load":
		// Assert arg1 is device and arg2 is int or hash
		fc.asmProgram.emitL(targetData, arg1, arg2)
	case "mod":
		fc.asmProgram.emitMod(targetData, arg1, arg2)
	case "xor":
		fc.asmProgram.emitXor(targetData, arg1, arg2)
	case "nor":
		fc.asmProgram.emitNor(targetData, arg1, arg2)
	case "max":
		fc.asmProgram.emitMax(targetData, arg1, arg2)
	case "min":
		fc.asmProgram.emitMin(targetData, arg1, arg2)
	default:
		return nil, CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
	}

	return targetData, nil
}

func (fc *funcCompiler) compileBuiltinArity3Func(fun *BuiltinArity3Func) (*data, error) {
	arg1, err := fc.compileExpr(fun.Arg1)
	if err != nil {
		return nil, err
	}

	arg2, err := fc.compileExpr(fun.Arg2)
	if err != nil {
		return nil, err
	}

	arg3, err := fc.compileExpr(fun.Arg3)
	if err != nil {
		return nil, err
	}

	// These do not return anything, processing them first
	switch fun.Op {
	case "store":
		// Assert arg1 is device and arg2 is int or hash
		fc.asmProgram.emitS(arg1, arg2, arg3)
		return nil, nil
	case "store_batch":
		// Assert arg1 is device and arg2 is int or hash
		fc.asmProgram.emitSb(arg1, arg2, arg3)
		return nil, nil
	}

	targetReg, err := fc.allocateTempRegister()
	if err != nil {
		return nil, err
	}

	targetData := newRegisterData(targetReg)
	switch fun.Op {
	case "load_batch":
		// Assert arg1 is device and arg2 is int or hash
		fc.asmProgram.emitLb(targetData, arg1, arg2, arg3)
		return targetData, nil
	default:
		return nil, CompilerError{Err: ErrInvalidState, Pos: &fun.Pos}
	}
}

func (fc *funcCompiler) compileAssignment(assignment *Assignment) error {
	data, err := fc.getVariableData(assignment.Left)
	if err != nil {
		return err
	}

	targetReg := data.register

	value, err := fc.compileExpr(assignment.Right)
	if err != nil {
		return err
	}

	if value == nil {
		return ErrInvalidState
	}

	if value.isFloatValue() {
		// if value is raw data and opt is enabled, update the value map
		if fc.conf.PropagateVariables {
			fc.varValueMap[assignment.Left] = value.floatValue
			return nil
		}

		// haven't allocated a register yet
		if targetReg == nil {
			targetReg, err = fc.allocatePermRegister()
			if err != nil {
				return err
			}
		}
		// move value to register, add to map
		fc.asmProgram.emitMove(newRegisterData(targetReg), value)
		fc.varRegisterMap[assignment.Left] = targetReg
		return nil
	}

	if value.isRegister() {
		if value.register.temporary && targetReg == nil {
			// promote register to permanent and add to map
			value.register.setTemporary(false)
			fc.varRegisterMap[assignment.Left] = value.register
			return nil
		}

		// haven't allocated a register yet
		if targetReg == nil {
			targetReg, err = fc.allocatePermRegister()
			if err != nil {
				return err
			}
		}
		// move value to register if needed, add to map
		if targetReg.id != value.register.id {
			fc.asmProgram.emitMove(newRegisterData(targetReg), value)
		}
		fc.varRegisterMap[assignment.Left] = targetReg
		return nil
	}

	return ErrInvalidState
}

func (fc *funcCompiler) compileJump(condData *data, jumpTo *data, invertCondition bool) {
	if fc.conf.OptimizeJumps && condData.isFloatValue() {
		if floatToBool(condData.floatValue) != invertCondition {
			fc.asmProgram.emitJ(jumpTo)
		}

		return
	}

	if invertCondition {
		fc.asmProgram.emitBeqz(condData, jumpTo)
	} else {
		fc.asmProgram.emitBnez(condData, jumpTo)
	}
}

func (fc *funcCompiler) compileCondJump(op string, l, r, jumpTo *data, invertCondition bool) error {
	if invertCondition {
		switch op {
		case ">":
			op = "<="
		case ">=":
			op = "<"
		case "<":
			op = ">="
		case "<=":
			op = ">"
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

// compileCondition compiles conditional branch if cond != invert condition
func (fc *funcCompiler) compileCondition(cond *Expr, jumpTo *data, invertCondition bool) error {
	if fc.conf.OptimizeJumps && cond.Binary != nil && opHasCondJump(cond.Binary.Op) {
		l, err := fc.compilePrimary(cond.Binary.LHS)
		if err != nil {
			return err
		}

		r, err := fc.compilePrimary(cond.Binary.RHS)
		if err != nil {
			return err
		}

		// If result of binary expression can be computed, compute and issue unconditional jump
		if fc.conf.PrecomputeExprs && l.isFloatValue() && r.isFloatValue() {
			v, err := computeBinop(l.floatValue, r.floatValue, cond.Binary.Op)
			if err != nil {
				return err
			}

			condVal := newNumData(v)
			fc.compileJump(condVal, jumpTo, invertCondition)
			return nil
		}

		// Issue binary conditional jump
		return fc.compileCondJump(cond.Binary.Op, l, r, jumpTo, invertCondition)

	}

	condVal, err := fc.compileExpr(cond)
	if err != nil {
		return err
	}

	fc.compileJump(condVal, jumpTo, invertCondition)
	return nil

}

func (fc *funcCompiler) compileIfStmt(ifStmt *IfStmt) error {
	hasElse := ifStmt.Else != nil
	ifLbl := newLabelData(fc.asmProgram.getUniqueLabel())
	endLbl := newLabelData(fc.asmProgram.getUniqueLabel())

	// If else condition exists, we issue jump to normal ifLbl. Otherwise, we jump straight to endLbl and invert the condition
	if !hasElse {
		err := fc.compileCondition(ifStmt.Condition, endLbl, true)
		if err != nil {
			return err
		}
	} else {
		err := fc.compileCondition(ifStmt.Condition, ifLbl, false)
		if err != nil {
			return err
		}
		err = fc.compileStmt(ifStmt.Else)
		if err != nil {
			return err
		}
		fc.asmProgram.emitJ(endLbl)
		fc.asmProgram.emitLabel(ifLbl.label)
	}

	err := fc.compileStmt(ifStmt.Body)
	if err != nil {
		return err
	}
	fc.asmProgram.emitLabel(endLbl.label)

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
		_, err := fc.compileExpr(stmt.Expr)
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
	case "abs":
		return math.Abs(l), nil
	case "acos":
		return math.Acos(l), nil
	case "asin":
		return math.Asin(l), nil
	case "atan":
		return math.Atan(l), nil
	case "ceil":
		return math.Ceil(l), nil
	case "floor":
		return math.Floor(l), nil
	case "log":
		return math.Log(l), nil
	case "sqrt":
		return math.Sqrt(l), nil
	case "round":
		return math.Round(l), nil
	case "trunc":
		return math.Trunc(l), nil
	default:
		return 0, CompilerError{Err: ErrInvalidState}
	}
}

func computeBuiltinArity2(l, r float64, op string) (float64, error) {
	switch op {
	case "mod":
		return math.Mod(l, r), nil
	case "xor":
		return boolToFloat(floatToBool(l) != floatToBool(r)), nil
	case "nor":
		return boolToFloat(!floatToBool(l) && !floatToBool(r)), nil
	case "max":
		return math.Max(l, r), nil
	case "min":
		return math.Min(l, r), nil
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
		return boolToFloat(l == r), nil
	case "!=":
		return boolToFloat(l != r), nil
	case ">=":
		return boolToFloat(l >= r), nil
	case ">":
		return boolToFloat(l > r), nil
	case "<=":
		return boolToFloat(l <= r), nil
	case "<":
		return boolToFloat(l < r), nil
	case "&&":
		return boolToFloat(floatToBool(l) && floatToBool(r)), nil
	case "||":
		return boolToFloat(floatToBool(l) || floatToBool(r)), nil
	default:
		return 0, CompilerError{Err: ErrInvalidState}
	}
}

// returns true if op has a special conditional jump instruction in the MIPS instruction set
func opHasCondJump(op string) bool {
	return op == "==" || op == "!=" || op == ">=" || op == ">" || op == "<" || op == "<="
}

func floatToBool(f float64) bool {
	return f != 0
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	} else {
		return 0
	}
}
