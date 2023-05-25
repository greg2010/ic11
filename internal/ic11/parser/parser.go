package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func Parse(files []io.Reader) (*AST, error) {
	readFiles := []string{}
	for _, file := range files {
		bytes, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		readFiles = append(readFiles, (string(bytes)))
	}

	initAST, err := parse(readersFromStrings(readFiles), basicParser)
	if err != nil {
		return nil, err
	}

	defMap := genPreprocessorMapping(initAST)
	mapFunc := mappingFunc(defMap)

	retAST, err := parse(readersFromStrings(readFiles), mappingParser(mapFunc))
	if err != nil {
		return nil, err
	}

	return retAST, nil
}

// mappingFunc returns a token -> token mapping function based on a go map
func mappingFunc(m map[string]string) func(lexer.Token) (lexer.Token, error) {
	return func(t lexer.Token) (lexer.Token, error) {
		if replaceVal, found := m[t.String()]; found {
			return lexer.Token{Type: t.Type, Pos: t.Pos, Value: replaceVal}, nil
		}
		return t, nil
	}
}

// genPreprocessorMapping generates a preprocessor mapping from #define statements in AST
func genPreprocessorMapping(ast *AST) map[string]string {
	defMap := make(map[string]string)
	for _, top := range ast.TopDec {
		if top.DefineDec != nil {
			if top.DefineDec.Device != "" {
				defMap[top.DefineDec.Name] = top.DefineDec.Device
				continue
			}

			if top.DefineDec.Value != nil {
				if top.DefineDec.Value.String != nil {
					defMap[top.DefineDec.Name] = *top.DefineDec.Value.String
					continue
				}
				if top.DefineDec.Value.Int != nil {
					defMap[top.DefineDec.Name] = fmt.Sprintf("%d", *top.DefineDec.Value.Int)
					continue
				}
				if top.DefineDec.Value.Float != nil {
					defMap[top.DefineDec.Name] = fmt.Sprintf("%f", *top.DefineDec.Value.Float)
					continue
				}
			}
		}
	}

	return defMap
}

func parse(files []io.Reader, parser *participle.Parser[AST]) (*AST, error) {
	var ast *AST
	for _, reader := range files {
		astSoFar, err := parser.Parse("", reader)
		if err != nil {
			return nil, err
		}
		ast = mergeAST(ast, astSoFar)
	}

	return ast, nil
}

func readersFromStrings(s []string) []io.Reader {
	readers := []io.Reader{}
	for _, str := range s {
		readers = append(readers, strings.NewReader(str))
	}

	return readers
}
