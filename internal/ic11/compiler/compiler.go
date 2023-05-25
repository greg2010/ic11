package compiler

import (
	"fmt"
	"io"
	"strings"

	"github.com/greg2010/ic11c/internal/ic11/ir"
	"github.com/greg2010/ic11c/internal/ic11/parser"
)

type Compiler struct {
	ast *parser.AST
	ir  *ir.Frontend
}

func New(files []io.Reader) (*Compiler, error) {
	ast, err := parser.Parse(files)
	if err != nil {
		return nil, err
	}

	ir, err := ir.NewFrontend(ast)
	if err != nil {
		return nil, err
	}

	return &Compiler{
		ast: ast,
		ir:  ir,
	}, nil
}

func (c *Compiler) Compile() (string, error) {
	var b strings.Builder
	fmt.Fprintln(&b, "raw IR:")
	b.WriteString(c.ir.String())
	//fmt.Fprintln(&b, "Block view:")
	//blockProg := ir.NewBlockProgram(c.ir.Get())
	//b.WriteString(blockProg.String())
	return b.String(), nil
}
