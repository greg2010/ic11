//nolint:govet
//nolint:structtag
package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "comment", Pattern: `//.*|/\*.*?\*/`},
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "Define", Pattern: "#define"},
		{Name: "Type", Pattern: `\b(int|float|string)\b`},
		{Name: "Device", Pattern: "d([0-6]|b)(:[0-9])?"},
		{Name: "Ident", Pattern: `\b([a-zA-Z_][a-zA-Z0-9_]*)\b`},
		{Name: "Punct", Pattern: `[-,()*/+%{};&\|!=:<>]|\[|\]`},
		{Name: "QuotedStr", Pattern: `"(.*?)"`},
		{Name: "Float", Pattern: `\d+(?:\.\d+)?`},
		{Name: "Int", Pattern: `\d+`},
	})

	// Build basicParser
	basicParser = participle.MustBuild[AST](
		participle.Unquote("QuotedStr"),
		participle.Lexer(lex),
		participle.UseLookahead(600))
	mappingParser = func(mapFunc participle.Mapper) *participle.Parser[AST] {
		return participle.MustBuild[AST](
			participle.Unquote("QuotedStr"),
			participle.Map(mapFunc),
			participle.Lexer(lex),
			participle.UseLookahead(600))
	}
)

// https://www.it.uu.se/katalog/aleji304/CompilersProject/uc.html

type AST struct {
	Pos lexer.Position

	TopDec []*TopDec `@@*`
}

func mergeAST(p1, p2 *AST) *AST {
	newTopDec := []*TopDec{}
	var p lexer.Position
	if p1 != nil {
		newTopDec = append(newTopDec, p1.TopDec...)
		p = p1.Pos
	}
	if p2 != nil {
		newTopDec = append(newTopDec, p2.TopDec...)
		p = p2.Pos
	}
	return &AST{
		Pos:    p,
		TopDec: newTopDec,
	}
}

type TopDec struct {
	Pos lexer.Position

	FunDec    *FunDec    `  @@`
	DefineDec *DefineDec `| @@`
	VarDec    *VarDec    `| @@ ";"`
}

type DefineDec struct {
	Pos    lexer.Position
	Name   string   `Define @Ident`
	Device string   `@Device`
	Value  *Literal `| @@`
}

type VarDec struct {
	Pos lexer.Position

	ScalarDec ScalarDec `@@`
}

type ScalarDec struct {
	Pos lexer.Position

	Name string `Type @Ident`
}

type ReturnStmt struct {
	Pos lexer.Position

	Result *Expr `"return" @@?`
}

type WhileStmt struct {
	Pos lexer.Position

	Condition *Expr `"while" "(" @@ ")"`
	Body      *Stmt `@@`
}

type IfStmt struct {
	Pos lexer.Position

	Condition *Expr `"if" "(" @@ ")"`
	Body      *Stmt `@@`
	Else      *Stmt `("else" @@)?`
}

type Stmts struct {
	Pos lexer.Position

	Stmts []*Stmt `@@*`
}

type Stmt struct {
	Pos lexer.Position

	IfStmt     *IfStmt     `  @@`
	ReturnStmt *ReturnStmt `| @@`
	WhileStmt  *WhileStmt  `| @@`
	Assignment *Assignment `| @@`
	CallFunc   *CallFunc   `| @@`
	Expr       *Expr       `| @@`
	Block      *Stmts      `| "{" @@ "}"`
	Empty      bool        `| @";"`
}

type FunBody struct {
	Pos lexer.Position

	Locals []*VarDec `(@@ ";")*`
	Stmts  *Stmts    `@@`
}

type FunDec struct {
	Pos lexer.Position

	ReturnType string       `@(Type | "void")`
	Name       string       `@Ident`
	Parameters []*Parameter `"(" ((@@ ("," @@)*) | "void") ")"`
	FunBody    *FunBody     `(";" | "{" @@ "}")`
}

type Parameter struct {
	Pos lexer.Position

	Scalar ScalarDec `@@`
}

type Assignment struct {
	Pos lexer.Position

	Left  string `@Ident "="`
	Right *Expr  `@@`
}

type Expr struct {
	Pos lexer.Position

	Binary  *Binary  ` @@`
	Primary *Primary `| @@`
	Unary   *Unary   `| @@`
}

type Binary struct {
	Pos lexer.Position

	LHS *Primary `@@`
	Op  string   `@( "|" "|" | "&" "&" | "!" "=" | ("!"|"<"|">") "="? | "=" "=" | "+" | "-" | "/" | "*" )`
	RHS *Primary `@@`
}

type Unary struct {
	Pos lexer.Position

	Op  string   `@( "-" | "!" )`
	RHS *Primary `@@`
}

type Primary struct {
	Pos lexer.Position

	CallFunc      *CallFunc `  @@`
	Literal       *Literal  `| @@`
	Ident         string    `| @Ident`
	SubExpression *Expr     `| "(" @@ ")" `
}

type Literal struct {
	Int    *int64   `  @('-'? Int)`
	Float  *float64 `| @('-'? Float)`
	String *string  `| @QuotedStr`
}

// Special function types
type HashConst struct {
	Pos lexer.Position

	Arg string `"hash" "(" @QuotedStr ")"`
}

type BuiltinArity0Func struct {
	Pos lexer.Position
	Op  string `@("yield" | "rand") "(" ")"`
}

type BuiltinArity1Func struct {
	Pos lexer.Position
	Op  string `@("sin" | "cos" |
	              "tan" | "abs" |
	              "acos" | "asin" |
				  "atan" | "ceil" |
				  "floor" | "log" |
				  "sqrt" | "round" |
				  "trunc" | "sleep" )`
	Arg *Expr `"(" @@ ")"`
}

type BuiltinArity2Func struct {
	Pos  lexer.Position
	Op   string `@("load" | "mod" | "xor" | "nor" | "max" | "min" )`
	Arg1 *Expr  `"(" @@ ","`
	Arg2 *Expr  `@@ ")"`
}

type BuiltinArity3Func struct {
	Pos  lexer.Position
	Op   string `@("store" | "store_batch" | "load_batch")`
	Arg1 *Expr  `"(" @@ ","`
	Arg2 *Expr  ` @@ ","`
	Arg3 *Expr  `@@ ")"`
}

type CallFunc struct {
	Pos lexer.Position

	Ident string  `@Ident`
	Index []*Expr `"(" (@@ ("," @@)*)? ")"`
}
