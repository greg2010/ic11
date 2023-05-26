package compiler

import (
	"fmt"
	"io"
	"strings"

	"github.com/greg2010/ic11c/internal/ic11/assembler"
	"github.com/greg2010/ic11c/internal/ic11/ir"
	"github.com/greg2010/ic11c/internal/ic11/parser"
	"github.com/greg2010/ic11c/internal/ic11/regassign"
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
	asm, err := assembler.New(c.ir.Get(), regassign.NewDummyAssigner(c.ir.Get()))
	if err != nil {
		return "", err
	}
	fmt.Fprintln(&b, "MIPS:")
	b.WriteString(asm.String())
	//fmt.Fprintln(&b, "Block view:")
	//blockProg := ir.NewBlockProgram(c.ir.Get())
	//b.WriteString(blockProg.String())
	return b.String(), nil
}
