package ic11

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
	"go.uber.org/zap"
)

var ErrNoMain = errors.New("main function is missing")
var ErrTooManyVars = errors.New("maximum number of variables supported is 15")
var ErrUnknownVar = errors.New("variable is not known")
var ErrInvalidState = errors.New("parser produced invalid state")
var ErrOutOfTempRegisters = errors.New("compiler ran out of temporary registers")
var ErrDiv0 = errors.New("division by 0")
var ErrInvalidFuncCall = errors.New("invalid function call")
var ErrUnknownLabel = errors.New("unknown label")

type CompilerOpts struct {
	OptimizeLabels  bool
	PrecomputeExprs bool
	OptimizeJumps   bool
}

type Compiler struct {
	l    *zap.SugaredLogger
	fn   string
	ast  *Program
	asm  *asmprogram
	conf CompilerOpts
}

type CompilerError struct {
	Pos      *lexer.Position
	Err      error
	CausedBy *CompilerError
}

func (ce CompilerError) Error() string {
	locHeader := ""

	if ce.Pos != nil {
		locHeader = fmt.Sprintf("%s: %d:%d ", ce.Pos.Filename, ce.Pos.Line, ce.Pos.Column)
	}

	errStr := fmt.Sprintf("%s%s", locHeader, ce.Err.Error())

	if ce.CausedBy != nil {
		errStr = fmt.Sprintf("%s\nCaused by: %s", errStr, ce.CausedBy.Error())
	}

	return errStr
}

func NewCompiler(l *zap.SugaredLogger, fn string, conf CompilerOpts) (*Compiler, error) {
	c := &Compiler{
		l:    l,
		fn:   fn,
		asm:  newASMProgram(conf.OptimizeLabels),
		conf: conf,
	}

	// Parse microC into an AST
	err := c.parse()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (comp *Compiler) parse() error {
	file, err := os.Open(comp.fn)
	if err != nil {
		comp.l.Errorf("error while opening file: %v", err)
		return err
	}

	defer file.Close()

	ast, err := parser.Parse("", file)
	repr.Println(ast)
	if err != nil {
		comp.l.Errorf("error while parsing: %v", err)
		return err
	}

	comp.ast = ast
	return nil
}

func (comp *Compiler) getConsts() map[string]float64 {
	vals := make(map[string]float64)

	for _, def := range comp.ast.TopDec {
		if def.DefineDec != nil && def.DefineDec.Value != nil {
			vals[def.DefineDec.Name] = def.DefineDec.Value.Number
		}
	}

	return vals
}

func (comp *Compiler) getDeviceAliases() map[string]string {
	vals := make(map[string]string)

	for _, def := range comp.ast.TopDec {
		if def.DefineDec != nil && def.DefineDec.Device != "" {
			vals[def.DefineDec.Name] = def.DefineDec.Device
		}
	}

	return vals

}

func (comp *Compiler) Compile() (string, error) {
	consts := comp.getConsts()
	deviceAliases := comp.getDeviceAliases()

	mainFunc := comp.getFunc("main")
	if mainFunc == nil {
		return "", &CompilerError{Err: ErrNoMain}
	}

	mainComp, err := newFuncCompiler(comp.asm, comp.conf, mainFunc, consts, deviceAliases)
	if err != nil {
		return "", err
	}

	err = mainComp.Compile()
	if err != nil {
		return "", err
	}

	outStr, err := comp.asm.print()
	if err != nil {
		return "", err
	}

	return outStr, nil
}

func (comp *Compiler) getFunc(name string) *FunDec {
	for _, top := range comp.ast.TopDec {
		if top.FunDec != nil && top.FunDec.Name == name {
			return top.FunDec
		}
	}

	return nil
}
