package interpreter

import (
	"fmt"
	"tp-lox/internal/environment"
	"tp-lox/internal/statements"
)

type Function struct {
	declaration *statements.FunctionStmt
}

func NewFunction(declaration *statements.FunctionStmt) *Function {
	return &Function{
		declaration: declaration,
	}
}

func (f *Function) Arity() int {
	return len(f.declaration.Params)
}

func (f *Function) Call(in *Interpreter, arguments []any) (any, error) {
	env := environment.NewEnclosedEnvironment(in.globals)
	for i, param := range f.declaration.Params {
		env.Define(param.Lexeme, arguments[i])
	}

	err := in.ExecuteBlock(f.declaration.Body, env)
	if err != nil {
		if returnValue, ok := err.(ReturnValue); ok {
			return returnValue.Value, nil
		}
		return nil, err
	}

	return nil, nil
}

func (f *Function) String() string {
	return fmt.Sprintf("<fn %s>", f.declaration.Name.Lexeme)
}
