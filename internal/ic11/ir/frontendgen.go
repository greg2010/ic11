package ir

import (
	"errors"

	"github.com/greg2010/ic11c/internal/ic11/parser"
)

// compile traverses the AST, calling corresponding compile* functions for each node type.
func (fr *Frontend) compile(ast *parser.AST) error {
	for _, top := range ast.TopDec {
		if top.FunDec != nil && top.FunDec.Name == "main" {
			if len(top.FunDec.Parameters) != 0 {
				return ErrMainFuncParameters
			}

			err := fr.compileFunDec(top.FunDec)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AST -> IR compile methods

func (fr *Frontend) compileFunDec(f *parser.FunDec) error {
	for _, stmt := range f.FunBody.Stmts.Stmts {
		err := fr.compileStmt(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fr *Frontend) compileStmt(s *parser.Stmt) error {
	if s.Empty {
		return nil
	}

	if s.Expr != nil {
		_, err := fr.compileExpr(s.Expr)
		if err != nil {
			return err
		}

		return nil
	}

	if s.IfStmt != nil {
		err := fr.compileIfStmt(s.IfStmt)
		if err != nil {
			return err
		}

		return nil
	}
	if s.WhileStmt != nil {
		err := fr.compileWhileStmt(s.WhileStmt)
		if err != nil {
			return err
		}

		return nil
	}

	if s.Assignment != nil {
		err := fr.compileAssignment(s.Assignment)
		if err != nil {
			return err
		}

		return nil
	}

	if s.CallFunc != nil {
		err := fr.compileVoidCallFunc(s.CallFunc)
		if err != nil {
			return err
		}
		return nil
	}

	if s.Block != nil {
		for _, subStmt := range s.Block.Stmts {
			err := fr.compileStmt(subStmt)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return errors.New("invalid stmt state")
}

// compilePrimary emits IR code if requfred and returns IRVar representing Primary node from AST
func (fr *Frontend) compilePrimary(p *parser.Primary) (*IRVar, error) {
	// If literal
	if p.Literal != nil {
		return fr.compileLiteral(p.Literal)
	}

	// If variable, just return it
	if p.Ident != "" {
		v := IRVar(p.Ident)
		return &v, nil
	}

	if p.SubExpression != nil {
		return fr.compileExpr(p.SubExpression)
	}

	if p.CallFunc != nil {
		return fr.compileRetCallFunc(p.CallFunc)
	}

	return nil, errors.New("invalid primary state")
}

func (fr *Frontend) compileLiteral(l *parser.Literal) (*IRVar, error) {
	lit, err := parserLiteralToIRLiteral(l)
	if err != nil {
		return nil, err
	}

	v := fr.newVar()
	fr.emit(IRAssignLiteral{Assignee: v, ValueVar: *lit})
	return &v, nil
}

func (fr *Frontend) compileExpr(e *parser.Expr) (*IRVar, error) {
	if e.Binary != nil {
		return fr.compileBinary(e.Binary)
	}

	if e.Unary != nil {
		return fr.compileUnary(e.Unary)
	}

	if e.Primary != nil {
		return fr.compilePrimary(e.Primary)
	}

	return nil, errors.New("invalid expr state")
}

func (fr *Frontend) compileBinary(b *parser.Binary) (*IRVar, error) {
	v := fr.newVar()

	l, err := fr.compilePrimary(b.LHS)
	if err != nil {
		return nil, err
	}

	r, err := fr.compilePrimary(b.RHS)
	if err != nil {
		return nil, err
	}

	// TODO convert binary operations to the known set of ops
	fr.emit(IRAssignBinary{Assignee: v, L: *l, R: *r, Op: b.Op})
	return &v, nil
}

func (fr *Frontend) compileUnary(u *parser.Unary) (*IRVar, error) {
	// TODO implement unary ops
	return nil, errors.New("invalid unary state")
}

func (fr *Frontend) compileAssignment(a *parser.Assignment) error {
	v, err := fr.compileExpr(a.Right)
	if err != nil {
		return err
	}

	fr.emit(IRAssignVar{Assignee: IRVar(a.Left), ValueVar: *v})
	return nil
}

func (fr *Frontend) compileIfStmt(i *parser.IfStmt) error {
	cond, err := fr.compileExpr(i.Condition)
	if err != nil {
		return err
	}
	if i.Else != nil {
		elseLbl := fr.newLabel()
		endLbl := fr.newLabel()
		fr.emit(IRIfZ{Cond: *cond, Label: elseLbl})
		err := fr.compileStmt(i.Body)
		if err != nil {
			return err
		}

		fr.emit(IRGoto{Label: endLbl})
		fr.emit(IRLabel{elseLbl})
		err = fr.compileStmt(i.Else)
		if err != nil {
			return err
		}

		fr.emit(IRLabel{endLbl})
	} else {
		endLbl := fr.newLabel()
		fr.emit(IRIfZ{Cond: *cond, Label: endLbl})
		err := fr.compileStmt(i.Body)
		if err != nil {
			return err
		}
		fr.emit(IRLabel{Label: endLbl})
	}

	return nil
}

func (fr *Frontend) compileWhileStmt(w *parser.WhileStmt) error {
	l1 := fr.newLabel()
	l2 := fr.newLabel()

	fr.emit(IRLabel{l1})

	cond, err := fr.compileExpr(w.Condition)
	if err != nil {
		return err
	}

	fr.emit(IRIfZ{Cond: *cond, Label: l2})
	err = fr.compileStmt(w.Body)
	if err != nil {
		return err
	}

	fr.emit(IRGoto{Label: l1})
	fr.emit(IRLabel{l2})

	return nil
}

func (fr *Frontend) compileVoidCallFunc(c *parser.CallFunc) error {
	switch c.Ident {
	case "store":
		fallthrough
	case "store_batch":
		return fr.compileBuiltinStore(c)
	default:
		var args []IRLiteralOrVar
		for _, arg := range c.Index {
			argV, err := fr.compileExpr(arg)
			if err != nil {
				return err
			}

			args = append(args, IRLiteralOrVar{v: argV})
		}

		instr := IRBuiltinCallVoid{BuiltinName: c.Ident, Params: args}
		fr.emit(instr)
		return nil
	}
}

func (fr *Frontend) compileRetCallFunc(c *parser.CallFunc) (*IRVar, error) {
	switch c.Ident {
	case "load":
		return fr.compileBuiltinLoadFunc(c)
	default:
		v := fr.newVar()

		var args []IRLiteralOrVar
		for _, arg := range c.Index {
			argV, err := fr.compileExpr(arg)
			if err != nil {
				return nil, err
			}

			args = append(args, IRLiteralOrVar{v: argV})
		}

		instr := IRBuiltinCallRet{BuiltinName: c.Ident, Params: args, Ret: v}
		fr.emit(instr)

		return &v, nil
	}
}

func (fr *Frontend) compileBuiltinLoadFunc(c *parser.CallFunc) (*IRVar, error) {
	if len(c.Index) < 2 {
		return nil, ErrInvalidFunctionCall
	}

	// First arg is device (passed as ident)
	device := c.Index[0]
	if device.Primary == nil || device.Primary.Ident == "" {
		return nil, ErrInvalidFunctionCall
	}
	arg0 := NewStringLiteral(device.Primary.Ident)

	// Second arg is device's Variable (passed as string)
	deviceVar := c.Index[1]
	if deviceVar.Primary == nil || deviceVar.Primary.Literal == nil {
		return nil, ErrInvalidFunctionCall
	}

	arg1, err := parserLiteralToIRLiteral(deviceVar.Primary.Literal)
	if err != nil {
		return nil, err
	}

	args := []IRLiteralOrVar{
		{
			lit: arg0,
		},
		{
			lit: arg1,
		},
	}

	v := fr.newVar()
	instr := IRBuiltinCallRet{c.Ident, args, v}
	fr.emit(instr)
	return &v, nil
}

// compileBuiltinStore compiles store and store_batch builtins into IR assembly
// this is required because these functions take special arguments that must be resolved literally
func (fr *Frontend) compileBuiltinStore(c *parser.CallFunc) error {
	if len(c.Index) < 3 {
		return ErrInvalidFunctionCall
	}

	// First arg is device (passed as ident)
	device := c.Index[0]
	if device.Primary == nil || device.Primary.Ident == "" {
		return ErrInvalidFunctionCall
	}
	arg0 := NewStringLiteral(device.Primary.Ident)

	// Second arg is device's Variable (passed as string)
	deviceVar := c.Index[1]
	if deviceVar.Primary == nil || deviceVar.Primary.Literal == nil {
		return ErrInvalidFunctionCall
	}

	arg1, err := parserLiteralToIRLiteral(deviceVar.Primary.Literal)
	if err != nil {
		return err
	}

	// Third arg is a register
	arg2, err := fr.compileExpr(c.Index[2])
	if err != nil {
		return err
	}

	args := []IRLiteralOrVar{
		{
			lit: arg0,
		},
		{
			lit: arg1,
		},
		{
			v: arg2,
		},
	}

	instr := IRBuiltinCallVoid{c.Ident, args}
	fr.emit(instr)
	return nil
}

func parserLiteralToIRLiteral(l *parser.Literal) (*IRLiteralType, error) {
	if l.Int != nil {
		return NewIntLiteral(*l.Int), nil
	}
	if l.Float != nil {
		return NewFloatLiteral(*l.Float), nil
	}

	if l.String != nil {
		return NewStringLiteral(*l.String), nil
	}

	return nil, ErrInvalidState
}
