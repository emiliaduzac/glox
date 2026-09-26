package statements

import (
	"fmt"
	"go/token"
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

type VariableStmt struct {
	Name        token.Token
	Initializer expression.Expr
}

func NewVariableStmt(expr expression.Expr) *VariableStmt {
	return &VariableStmt{
		Initializer: expr,
	}
}

func (p *VariableStmt) Statement() {}

func String(s Stmt) string {
	switch s := s.(type) {
	case *PrintStmt:
		return fmt.Sprintf("(print %s)", expression.String(s.Expr))
	case *ExpressionStmt:
		return fmt.Sprintf("(expr %s)", expression.String(s.Expr))
	}
	return fmt.Sprintf("unknown statement: %T", s)
}
