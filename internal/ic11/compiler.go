package ic11

import (
	"errors"
	"fmt"
	"io"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/greg2010/ic11/internal/printer"
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
	OptimizeLabels     bool
	PrecomputeExprs    bool
	OptimizeJumps      bool
	PropagateVariables bool
	EmitDeviceAliases  bool
	PrecomputeHashes   bool
}

func AllCompilerOpts() CompilerOpts {
	return CompilerOpts{
		OptimizeLabels:  true,
		PrecomputeExprs: true,
		OptimizeJumps:   true,
		// Bug: PropagateVariables is bugged, disabling by default
		PropagateVariables: false,
		EmitDeviceAliases:  true,
		PrecomputeHashes:   true,
	}
}

func NoCompilerOpts() CompilerOpts {
	return CompilerOpts{
		OptimizeLabels:     false,
		PrecomputeExprs:    false,
		OptimizeJumps:      false,
		PropagateVariables: false,
		EmitDeviceAliases:  true,
		PrecomputeHashes:   false,
	}

}

type Compiler struct {
	printer printer.Printer
	ast     *Program
	asm     *asmprogram
	conf    CompilerOpts
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

func NewCompiler(files []io.Reader, conf CompilerOpts, printer printer.Printer) (*Compiler, error) {
	c := &Compiler{
		printer: printer,
		asm:     newASMProgram(conf.OptimizeLabels),
		conf:    conf,
	}
	// Parse microC into an AST
	err := c.parse(files)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (comp *Compiler) parse(files []io.Reader) error {
	for _, reader := range files {
		ast, err := parser.Parse("", reader)
		if err != nil {
			return err
		}
		comp.ast = mergeProgram(comp.ast, ast)
	}

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

// processDeviceAliases generates device aliases if flag is set and returns a map of aliases
func (comp *Compiler) processDeviceAliases() map[string]string {
	vals := make(map[string]string)

	for _, def := range comp.ast.TopDec {
		if def.DefineDec != nil && def.DefineDec.Device != "" {
			vals[def.DefineDec.Name] = def.DefineDec.Device
			if comp.conf.EmitDeviceAliases {
				comp.asm.emitAlias(newStringData(def.DefineDec.Name), newDeviceData(def.DefineDec.Device))
			}
		}
	}

	return vals

}

func (comp *Compiler) Compile() (string, error) {
	consts := comp.getConsts()
	deviceAliases := comp.processDeviceAliases()

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
