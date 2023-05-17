//nolint:govet
//nolint:structtag
package ic11

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "comment", Pattern: `//.*|/\*.*?\*/`},
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "Define", Pattern: "#define"},
		{Name: "Type", Pattern: `\bnum\b`},
		{Name: "Device", Pattern: "d([0-6]|b)(:[0-9])?"},
		{Name: "Ident", Pattern: `\b([a-zA-Z_][a-zA-Z0-9_]*)\b`},
		{Name: "Punct", Pattern: `[-,()*/+%{};&\|!=:<>]|\[|\]`},
		{Name: "QuotedStr", Pattern: `"(.*?)"`},
		{Name: "Float", Pattern: `\d+(?:\.\d+)?`},
		{Name: "Int", Pattern: `\d+`},
	})

	// Build parser
	parser = participle.MustBuild[Program](
		participle.Unquote("QuotedStr"),
		participle.Lexer(lex),
		participle.UseLookahead(600))
)

// https://www.it.uu.se/katalog/aleji304/CompilersProject/uc.html
//
// program         ::= topdec_list
// topdec_list     ::= /empty/ | topdec topdec_list
// topdec          ::= vardec ";"
//                  | funtype ident "(" formals ")" funbody
// vardec          ::= scalardec | arraydec
// scalardec       ::= typename ident
// arraydec        ::= typename ident "[" intconst "]"
// typename        ::= "int" | "char"
// funtype         ::= typename | "void"
// funbody         ::= "{" locals stmts "}" | ";"
// formals         ::= "void" | formal_list
// formal_list     ::= formaldec | formaldec "," formal_list
// formaldec       ::= scalardec | typename ident "[" "]"
// locals          ::= /empty/ | vardec ";" locals
// stmts           ::= /empty/ | stmt stmts
// stmt            ::= expr ";"
//                  | "return" expr ";" | "return" ";"
//                  | "while" condition stmt
//                  | "if" condition stmt else_part
//                  | "{" stmts "}"
//                  | ";"
// else_part       ::= /empty/ | "else" stmt
// condition       ::= "(" expr ")"
// expr            ::= intconst
//                  | ident | ident "[" expr "]"
//                  | unop expr
//                  | expr binop expr
//                  | ident "(" actuals ")"
//                  | "(" expr ")"
// unop            ::= "-" | "!"
// binop           ::= "+" | "-" | "*" | "/"
//                  | "<" | ">" | "<=" | ">=" | "!=" | "=="
//                  | "&&"
//                  | "="
// actuals         ::= /empty/ | expr_list
// expr_list       ::= expr | expr "," expr_list

type Program struct {
	Pos lexer.Position

	TopDec []*TopDec `@@*`
}

func mergeProgram(p1, p2 *Program) *Program {
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
	return &Program{
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
	Name   string  `Define @Ident`
	Device string  `@Device`
	Value  *Number `| @@`
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

	Op  string   `@( "-" )`
	RHS *Primary `@@`
}

type Primary struct {
	Pos lexer.Position

	HashConst         *HashConst         `  @@`
	Device            string             `| @Device`
	BuiltinArity3Func *BuiltinArity3Func `| @@`
	BuiltinArity2Func *BuiltinArity2Func `| @@`
	BuiltinArity1Func *BuiltinArity1Func `| @@`
	BuiltinArity0Func *BuiltinArity0Func `| @@`
	CallFunc          *CallFunc          `| @@`
	Number            *Number            `| @@`
	Ident             string             `| @Ident`
	StringValue       string             "| @QuotedStr"
	SubExpression     *Expr              `| "(" @@ ")" `
}

// Int | Float union type
type Number struct {
	Number float64 `@('-'? (Float | Int))`
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
				  "trunc" )`
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
	Op   string `@("store" | "store_batch")`
	Arg1 *Expr  `"(" @@ ","`
	Arg2 *Expr  ` @@ ","`
	Arg3 *Expr  `@@ ")"`
}

type CallFunc struct {
	Pos lexer.Position

	Ident string  `@Ident`
	Index []*Expr `"(" (@@ ("," @@)*)? ")"`
}
