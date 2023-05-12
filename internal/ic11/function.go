package ic11

type funcCompiler struct {
	funAST         *FunDec
	varRegisterMap map[string]*register
	globalConsts   map[string]float64
	emitted        []string
	tempRegisters  []*register
	asmProgram     *asmprogram
}

func newFuncCompiler(asmProgram *asmprogram, funAST *FunDec, globalConsts map[string]float64) (*funcCompiler, error) {
	regsNeeded := len(funAST.Parameters) + len(funAST.FunBody.Locals)

	if regsNeeded > 15 {
		return nil, ErrTooManyVars
	}

	varRegisterMap := make(map[string]*register)
	for i, v := range funAST.FunBody.Locals {
		varRegisterMap[v.ScalarDec.Name] = newRegister(i, false)
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
		tempRegisters:  tempRegs,
		asmProgram:     asmProgram,
	}, nil
}

func (fc *funcCompiler) allocateTempRegister() (*register, error) {
	for _, reg := range fc.tempRegisters {
		if !reg.allocated {
			reg.allocate()
			return reg, nil
		}
	}

	return nil, ErrOutOfTempRegisters
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

	return nil, ErrInvalidState
}

func (fc *funcCompiler) compileUnary(unary *Unary, out *register) (*data, error) {
	target, err := fc.compilePrimary(unary.Rhs, out)
	if err != nil {
		return nil, err
	}

	// opt
	if target.isValue() {
		if unary.Op == "-" {
			return newNumData(unary.Rhs.Number.Number * -1), nil
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

	return nil, ErrInvalidState
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
		} else {
			return float64(0), nil
		}
	case "!=":
		if l != r {
			return float64(1), nil
		} else {
			return float64(0), nil
		}
	case ">=":
		if l >= r {
			return float64(1), nil
		} else {
			return float64(0), nil
		}
	case ">":
		if l > r {
			return float64(1), nil
		} else {
			return float64(0), nil
		}
	case "<=":
		if l <= r {
			return float64(1), nil
		} else {
			return float64(0), nil
		}
	case "<":
		if l < r {
			return float64(1), nil
		} else {
			return float64(0), nil
		}
	default:
		return 0, ErrInvalidState
	}
}

func (fc *funcCompiler) compileBinary(binary *Binary, out *register) (*data, error) {
	// Only passing out to right side (not left side) because if out is used on the right side,
	//  and left side mutated the value, computation will be invalid
	leftData, err := fc.compilePrimary(binary.Lhs, nil)
	if err != nil {
		return nil, err
	}

	rightData, err := fc.compilePrimary(binary.Rhs, out)

	//opt
	if leftData.isValue() && rightData.isValue() {
		v, err := computeBinop(leftData.value, rightData.value, binary.Op)
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
	if out != nil && rightData.isTemporaryRegister() && rightData.register.i != out.i {
		rightData.register.deallocate()
	}

	if out == nil && leftData.isTemporaryRegister() && rightData.isTemporaryRegister() {
		rightData.register.deallocate()
	}

	if targetReg == nil {
		targetReg, err = fc.allocateTempRegister()
		if err != nil {
			return nil, err
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
		return nil, ErrInvalidState
	}
}

func (fc *funcCompiler) getSymbolData(symbol string) *data {
	// first, search global consts
	if num, found := fc.globalConsts[symbol]; found {
		return newNumData(num)
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

	if primary.Ident != "" {
		data := fc.getSymbolData(primary.Ident)
		if data == nil {
			return nil, ErrUnknownVar
		}

		return data, nil
	}

	if primary.SubExpression != nil {
		return fc.compileExpr(primary.SubExpression, out)
	}

	if primary.CallFunc != nil {
		return fc.compileCallFunc(primary.CallFunc, out)
	}

	return nil, ErrInvalidState
}

func (fc *funcCompiler) compileCallFunc(callFunc *CallFunc, out *register) (*data, error) {
	targetReg := out
	if targetReg == nil {
		var err error
		targetReg, err = fc.allocateTempRegister()
		if err != nil {
			return nil, err
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

	return targetData, nil
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
	if !value.isRegister() || value.register.i != targetReg.i {
		fc.asmProgram.emitMove(newRegisterData(targetReg), value)
		// if value is a temp register it's up to us here to free it
		if value.register != nil && value.register.temporary {
			value.register.deallocate()
		}
	}

	return nil
}

func (fc *funcCompiler) compileIfStmt(ifStmt *IfStmt) error {
	cond, err := fc.compileExpr(ifStmt.Condition, nil)
	if err != nil {
		return err
	}

	elseLbl := newLabelData(fc.asmProgram.getUniqueLabel())
	fc.asmProgram.emitBeqz(cond, elseLbl)

	err = fc.compileStmt(ifStmt.Body)
	if err != nil {
		return err
	}

	fc.asmProgram.emitLabel(elseLbl.label)
	if ifStmt.Else != nil {
		endLbl := newLabelData(fc.asmProgram.getUniqueLabel())
		fc.asmProgram.emitBnez(cond, endLbl)
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
	cond, err := fc.compileExpr(whileStmt.Condition, nil)
	if err != nil {
		return err
	}

	fc.asmProgram.emitBnez(cond, loopStartLbl)

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

	return ErrInvalidState
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
