package resolver

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/interpreter"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

type Resolver struct {
	Interpreter *interpreter.Interpreter
	ScopesStack []map[string]bool
}

func NewResolver(interpreter *interpreter.Interpreter) *Resolver {
	return &Resolver{
		Interpreter: interpreter,
		ScopesStack: []map[string]bool{},
	}
}

func (r *Resolver) Resolve(statements []statements.Stmt) error {
	for _, stmt := range statements {
		err := r.resolveStmt(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) resolveStmt(stmt statements.Stmt) error {
	switch s := stmt.(type) {
	case *statements.BlockStmt:
		r.beginScope()
		err := r.Resolve(s.Statements)
		if err != nil {
			return err
		}
		r.endScope()

	case *statements.VarStmt:
		r.declare(s.Name.Lexeme)
		if s.Initializer != nil {
			err := r.resolveExpr(s.Initializer)
			if err != nil {
				return err
			}
		}
		r.define(s.Name.Lexeme)

	case *statements.ExpressionStmt:
		err := r.resolveExpr(s.Expr)
		if err != nil {
			return err
		}

	case *statements.PrintStmt:
		err := r.resolveExpr(s.Expr)
		if err != nil {
			return err
		}

	case *statements.IfStmt:
		err := r.resolveExpr(s.Condition)
		if err != nil {
			return err
		}
		err = r.resolveStmt(s.ThenBranch)
		if err != nil {
			return err
		}
		if s.ElseBranch != nil {
			err = r.resolveStmt(s.ElseBranch)
			if err != nil {
				return err
			}
		}

	case *statements.WhileStmt:
		err := r.resolveExpr(s.Condition)
		if err != nil {
			return err
		}
		err = r.resolveStmt(s.Body)
		if err != nil {
			return err
		}

	case *statements.FunctionStmt:
		r.declare(s.Name.Lexeme)
		r.define(s.Name.Lexeme)
		err := r.resolveFunction(s)
		if err != nil {
			return err
		}

	case *statements.ReturnStmt:
		if s.Value != nil {
			err := r.resolveExpr(s.Value)
			if err != nil {
				return err
			}
		}

	default:
		return errors.NewLoxError(0, fmt.Sprintf("unknown statement: %T", stmt))
	}
	return nil
}

func (r *Resolver) beginScope() {
	r.ScopesStack = append(r.ScopesStack, make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.ScopesStack = r.ScopesStack[:len(r.ScopesStack)-1]
}

func (r *Resolver) declare(name string) {
	if len(r.ScopesStack) == 0 {
		return
	}
	scope := r.ScopesStack[len(r.ScopesStack)-1]
	scope[name] = false
}

func (r *Resolver) define(name string) {
	if len(r.ScopesStack) == 0 {
		return
	}
	scope := r.ScopesStack[len(r.ScopesStack)-1]
	scope[name] = true
}

func (r *Resolver) resolveExpr(expr expression.Expr) error {
	switch e := expr.(type) {
	case *expression.Variable:
		if len(r.ScopesStack) > 0 {
			scope := r.ScopesStack[len(r.ScopesStack)-1]
			if defined, ok := scope[e.Name.Lexeme]; ok && !defined {
				return errors.NewLoxError(e.Name.Line, "Cannot read local variable in its own initializer.")
			}
		}
		r.resolveLocal(e, e.Name)

	case *expression.Assign:
		err := r.resolveExpr(e.Value)
		if err != nil {
			return err
		}
		r.resolveLocal(e, e.Name)

	case *expression.Binary:
		err := r.resolveExpr(e.Left)
		if err != nil {
			return err
		}
		err = r.resolveExpr(e.Right)
		if err != nil {
			return err
		}

	case *expression.Grouping:
		err := r.resolveExpr(e.Expr)
		if err != nil {
			return err
		}

	case *expression.Literal:
		// nada por hacer: los literales no mencionan variables ni contienen expresiones dentro

	case *expression.Logical:
		err := r.resolveExpr(e.Left)
		if err != nil {
			return err
		}
		err = r.resolveExpr(e.Right)
		if err != nil {
			return err
		}

	case *expression.Unary:
		err := r.resolveExpr(e.Right)
		if err != nil {
			return err
		}

	case *expression.Call:
		err := r.resolveExpr(e.Callee)
		if err != nil {
			return err
		}
		for _, arg := range e.Arguments {
			err = r.resolveExpr(arg)
			if err != nil {
				return err
			}
		}

	default:
		return errors.NewLoxError(0, fmt.Sprintf("unknown expression: %T", expr))
	}
	return nil
}

func (r *Resolver) resolveLocal(expr expression.Expr, name token.Token) {
	for i := len(r.ScopesStack) - 1; i >= 0; i-- {
		if _, ok := r.ScopesStack[i][name.Lexeme]; ok {
			r.Interpreter.Resolve(expr, len(r.ScopesStack)-1-i)
			return
		}
	}
}

func (r *Resolver) resolveFunction(function *statements.FunctionStmt) error {
	r.beginScope()
	for _, param := range function.Params {
		r.declare(param.Lexeme)
		r.define(param.Lexeme)
	}

	err := r.Resolve(function.Body)
	if err != nil {
		return err
	}

	r.endScope()
	return nil
}
