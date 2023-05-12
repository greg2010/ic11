package ic11

import (
	"github.com/alecthomas/participle/v2/lexer"
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

type TopDec struct {
	Pos lexer.Position

	FunDec   *FunDec   `  @@`
	ConstDec *ConstDec `| @@`
	VarDec   *VarDec   `| @@ ";"`
}

type ConstDec struct {
	Pos   lexer.Position
	Name  string  `Define @Ident`
	Value *Number `@@`
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
	Block      *Stmts      `| "{" @@ "}"`
	Assignment *Assignment `| @@`
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

	Lhs *Primary `@@`
	Op  string   `@( "|" "|" | "&" "&" | "!" "=" | ("!"|"="|"<"|">") "="? | "+" | "-" | "/" | "*" )`
	Rhs *Primary `@@`
}

type Unary struct {
	Pos lexer.Position

	Op  string   `@( "-" )`
	Rhs *Primary `@@`
}

type Primary struct {
	Pos lexer.Position

	Number *Number `  @@`
	CallFunc      *CallFunc `| @@`
	Ident         string `| @Ident`
	SubExpression *Expr  `| "(" @@ ")" `
}

// Int | Float union type
type Number struct {
	Number float64 `@('-'? (Float | Int))`
}


type CallFunc struct {
	Pos lexer.Position

	Ident string  `@Ident`
	Index []*Expr `"(" (@@ ("," @@)*)? ")"`
}

const sample = `
/* This is an example uC program. */
void putint(int i);

int fac(int n)
{
    if (n < 2)
        return n;
    return n * fac(n - 1);
}

int sum(int n, int a[])
{
    int i;
    int s;

    i = 0;
    s = 0;
    while (i <= n) {
        s = s + a[i];
        i = i + 1;
    }
    return s;
}

int main(void)
{
    int a[2];

    a[0] = fac(5);
    a[1] = 27;
    putint(sum(2, a)); // prints 147
    return 0;
}
`
