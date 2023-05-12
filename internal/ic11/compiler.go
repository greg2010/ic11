package ic11

import (
	"errors"
	"os"

	"github.com/alecthomas/participle/v2"
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

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "comment", Pattern: `//.*|/\*.*?\*/`},
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "Define", Pattern: "#define"},
		{Name: "Type", Pattern: `\bnum\b`},
		{Name: "Ident", Pattern: `\b([a-zA-Z_][a-zA-Z0-9_]*)\b`},
		{Name: "Punct", Pattern: `[-,()*/+%{};&!=:<>]|\[|\]`},
		{Name: "Float", Pattern: `\d+(?:\.\d+)?`},
		{Name: "Int", Pattern: `\d+`},
	})
	parser = participle.MustBuild[Program](
		participle.Lexer(lex),
		participle.UseLookahead(600))
)

type Compiler struct {
	l   *zap.SugaredLogger
	fn  string
	ast *Program
	asm *asmprogram
}

func NewCompiler(l *zap.SugaredLogger, fn string) (*Compiler, error) {
	c := &Compiler{
		l:   l,
		fn:  fn,
		asm: newASMProgram(),
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
		if def.ConstDec != nil {
			vals[def.ConstDec.Name] = def.ConstDec.Value.Number
		}
	}

	return vals
}

func (comp *Compiler) Compile() (string, error) {
	consts := comp.getConsts()

	mainFunc := comp.getFunc("main")
	if mainFunc == nil {
		return "", ErrNoMain
	}

	mainComp, err := newFuncCompiler(comp.asm, mainFunc, consts)
	if err != nil {
		return "", err
	}

	err = mainComp.Compile()
	if err != nil {
		return "", err
	}

	outStr := comp.asm.print()

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
