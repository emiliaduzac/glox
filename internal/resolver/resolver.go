package resolver

import (
	"tp-lox/internal/interpreter"
	"tp-lox/internal/statements"
)

type Resolver struct {
	Interpreter *interpreter.Interpreter
	scopes      []map[string]bool
}

func NewResolver(interpreter *interpreter.Interpreter) *Resolver {
	return &Resolver{
		Interpreter: interpreter,
		scopes:      []map[string]bool{},
	}
}

func (r *Resolver) Resolve(statements []statements.Stmt) {
	for _, stmt := range statements {
		r.resolveStmt(stmt)
	}
}

func (r *Resolver) resolveStmt(stmt statements.Stmt) {
	switch s := stmt.(type) {
	case *statements.BlockStmt:
		r.beginScope()
		r.Resolve(s.Statements)
		r.endScope()
	default:
		// agregar caso default
	}
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}
