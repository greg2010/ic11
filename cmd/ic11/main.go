package main

import (
	"fmt"
	"os"

	"github.com/greg2010/ic11/internal/ic11"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	l := logger.Sugar()

	args := os.Args[1:]

	if len(args) < 1 {
		l.Fatal("Filename not provided")
	}

	fname := args[0]

	compiler, err := ic11.NewCompiler(l, fname)
	if err != nil {
		l.Fatal(err)
	}

	out, err := compiler.Compile()
	if err != nil {
		l.Fatal(err)
	}

	fmt.Print(out)
}
