package statements

import (
	"tp-lox/internal/expression"
)

type Stmt interface {
	Statement()
}

// Expression Statements
type ExpressionStmt struct {
	Expr expression.Expr
}

func NewExpressionStmt(expr expression.Expr) *ExpressionStmt {
	return &ExpressionStmt{
		Expr: expr,
	}
}

func (e *ExpressionStmt) Statement() {}

// Print Statements
type PrintStmt struct {
	Expr expression.Expr
}

func NewPrintStmt(expr expression.Expr) *PrintStmt {
	return &PrintStmt{
		Expr: expr,
	}
}

func (p *PrintStmt) Statement() {}
