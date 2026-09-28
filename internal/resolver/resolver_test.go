package resolver

import (
	"strings"
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/interpreter"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

type UnsupportedStmt struct{}

func (u UnsupportedStmt) Statement() {}

type UnsupportedExpr struct{}

func (u UnsupportedExpr) Expression() {}

func SetUpResolver(t *testing.T) (*Resolver, *interpreter.Interpreter) {
	t.Helper()
	interp := interpreter.NewInterpreter()
	return NewResolver(interp), interp
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func AssertErrorMessage(t *testing.T, err error, expected string) {
	t.Helper()
	AssertError(t, err)
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %q", expected, err.Error())
	}
}

func TestResolveVariableInSameScope(t *testing.T) {
	r, interp := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
	varExpr := expression.NewVariable(tok)
	block := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewLiteral(1.0)),
		statements.NewPrintStmt(varExpr),
	})

	err := r.Resolve([]statements.Stmt{block})
	AssertNoError(t, err)

	distance, ok := interp.Locals[varExpr]
	if !ok {
		t.Fatalf("expected a resolved distance for the variable, got none")
	}
	if distance != 0 {
		t.Errorf("expected distance 0, got %d", distance)
	}
}

func TestResolveVariableOneScopeUp(t *testing.T) {
	r, interp := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
	varExpr := expression.NewVariable(tok)
	inner := statements.NewBlockStmt([]statements.Stmt{
		statements.NewPrintStmt(varExpr),
	})
	outer := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewLiteral(1.0)),
		inner,
	})

	err := r.Resolve([]statements.Stmt{outer})
	AssertNoError(t, err)

	distance, ok := interp.Locals[varExpr]
	if !ok {
		t.Fatalf("expected a resolved distance for the variable, got none")
	}
	if distance != 1 {
		t.Errorf("expected distance 1, got %d", distance)
	}
}

func TestResolveGlobalVariableIsNotResolved(t *testing.T) {
	r, interp := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "g", nil, 1)
	varExpr := expression.NewVariable(tok)
	global := statements.NewVarStmt(tok, expression.NewLiteral(1.0))
	block := statements.NewBlockStmt([]statements.Stmt{
		statements.NewPrintStmt(varExpr),
	})

	err := r.Resolve([]statements.Stmt{global, block})
	AssertNoError(t, err)

	if _, ok := interp.Locals[varExpr]; ok {
		t.Fatalf("expected no resolved distance for a global variable, but found one")
	}
}

func TestResolveAssignmentInSameScope(t *testing.T) {
	r, interp := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
	assignExpr := expression.NewAssign(tok, expression.NewLiteral(2.0))
	block := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewLiteral(1.0)),
		statements.NewExpressionStmt(assignExpr),
	})

	err := r.Resolve([]statements.Stmt{block})
	AssertNoError(t, err)

	distance, ok := interp.Locals[assignExpr]
	if !ok {
		t.Fatalf("expected a resolved distance for the assignment, got none")
	}
	if distance != 0 {
		t.Errorf("expected distance 0, got %d", distance)
	}
}

func TestResolveMultipleVariablesIndependently(t *testing.T) {
	r, interp := SetUpResolver(t)

	tokA := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
	tokB := *token.NewToken(token.IDENTIFIER, "b", nil, 1)
	varA := expression.NewVariable(tokA)
	varB := expression.NewVariable(tokB)
	block := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tokA, expression.NewLiteral(1.0)),
		statements.NewVarStmt(tokB, expression.NewLiteral(2.0)),
		statements.NewPrintStmt(varA),
		statements.NewPrintStmt(varB),
	})

	err := r.Resolve([]statements.Stmt{block})
	AssertNoError(t, err)

	distanceA, okA := interp.Locals[varA]
	if !okA || distanceA != 0 {
		t.Errorf("expected 'a' resolved at distance 0, got %d (found: %v)", distanceA, okA)
	}
	distanceB, okB := interp.Locals[varB]
	if !okB || distanceB != 0 {
		t.Errorf("expected 'b' resolved at distance 0, got %d (found: %v)", distanceB, okB)
	}
}

func TestResolveVariableInIfBranchWithoutBlock(t *testing.T) {
	r, interp := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
	varExpr := expression.NewVariable(tok)
	ifStmt := statements.NewIfStmt(expression.NewLiteral(true), statements.NewPrintStmt(varExpr), nil)
	block := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewLiteral(1.0)),
		ifStmt,
	})

	err := r.Resolve([]statements.Stmt{block})
	AssertNoError(t, err)

	distance, ok := interp.Locals[varExpr]
	if !ok {
		t.Fatalf("expected a resolved distance for the variable, got none")
	}
	if distance != 0 {
		t.Errorf("expected distance 0, got %d", distance)
	}
}

func TestResolveVariableInWhileBodyWithoutBlock(t *testing.T) {
	r, interp := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
	varExpr := expression.NewVariable(tok)
	whileStmt := statements.NewWhileStmt(expression.NewLiteral(true), statements.NewPrintStmt(varExpr))
	block := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewLiteral(1.0)),
		whileStmt,
	})

	err := r.Resolve([]statements.Stmt{block})
	AssertNoError(t, err)

	distance, ok := interp.Locals[varExpr]
	if !ok {
		t.Fatalf("expected a resolved distance for the variable, got none")
	}
	if distance != 0 {
		t.Errorf("expected distance 0, got %d", distance)
	}
}

func TestResolveVariableNestedInsideBinaryExpression(t *testing.T) {
	r, interp := SetUpResolver(t)

	tokX := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
	tokY := *token.NewToken(token.IDENTIFIER, "y", nil, 1)
	varX := expression.NewVariable(tokX)
	initializer := expression.NewBinary(varX, *token.NewToken(token.PLUS, "+", nil, 1), expression.NewLiteral(1.0))
	inner := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tokY, initializer),
	})
	outer := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tokX, expression.NewLiteral(1.0)),
		inner,
	})

	err := r.Resolve([]statements.Stmt{outer})
	AssertNoError(t, err)

	distance, ok := interp.Locals[varX]
	if !ok {
		t.Fatalf("expected a resolved distance for the variable inside the Binary, got none")
	}
	if distance != 1 {
		t.Errorf("expected distance 1, got %d", distance)
	}
}

func TestResolveVariableUsedInOwnInitializer(t *testing.T) {
	r, _ := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
	block := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewVariable(tok)),
	})

	err := r.Resolve([]statements.Stmt{block})
	AssertErrorMessage(t, err, "Cannot read local variable in its own initializer")
}

func TestResolveUnsupportedStatement(t *testing.T) {
	r, _ := SetUpResolver(t)

	err := r.Resolve([]statements.Stmt{UnsupportedStmt{}})
	AssertError(t, err)
}

func TestResolveUnsupportedExpression(t *testing.T) {
	r, _ := SetUpResolver(t)

	stmt := statements.NewExpressionStmt(UnsupportedExpr{})
	err := r.Resolve([]statements.Stmt{stmt})
	AssertError(t, err)
}

func TestResolveVariableAtTwoScopesUp(t *testing.T) {
	r, interp := SetUpResolver(t)

	tokX := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
	tokY := *token.NewToken(token.IDENTIFIER, "y", nil, 1)
	varX := expression.NewVariable(tokX)
	varY := expression.NewVariable(tokY)

	innermost := statements.NewBlockStmt([]statements.Stmt{
		statements.NewPrintStmt(varX),
		statements.NewPrintStmt(varY),
	})
	middle := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tokY, expression.NewLiteral(2.0)),
		innermost,
	})
	outer := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tokX, expression.NewLiteral(1.0)),
		middle,
	})

	err := r.Resolve([]statements.Stmt{outer})
	AssertNoError(t, err)

	distanceX, okX := interp.Locals[varX]
	if !okX || distanceX != 2 {
		t.Errorf("expected 'x' resolved at distance 2, got %d (found: %v)", distanceX, okX)
	}
	distanceY, okY := interp.Locals[varY]
	if !okY || distanceY != 1 {
		t.Errorf("expected 'y' resolved at distance 1, got %d (found: %v)", distanceY, okY)
	}
}

func TestResolveShadowedVariableResolvesToNearestDeclaration(t *testing.T) {
	r, interp := SetUpResolver(t)

	tok := *token.NewToken(token.IDENTIFIER, "i", nil, 1)
	innerVar := expression.NewVariable(tok)
	outerVar := expression.NewVariable(tok)

	inner := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewLiteral(10.0)),
		statements.NewPrintStmt(innerVar),
	})
	outer := statements.NewBlockStmt([]statements.Stmt{
		statements.NewVarStmt(tok, expression.NewLiteral(0.0)),
		inner,
		statements.NewPrintStmt(outerVar),
	})

	err := r.Resolve([]statements.Stmt{outer})
	AssertNoError(t, err)

	innerDistance, innerOk := interp.Locals[innerVar]
	if !innerOk || innerDistance != 0 {
		t.Errorf("expected inner 'i' resolved at distance 0, got %d (found: %v)", innerDistance, innerOk)
	}
	outerDistance, outerOk := interp.Locals[outerVar]
	if !outerOk || outerDistance != 0 {
		t.Errorf("expected outer 'i' resolved at distance 0, got %d (found: %v)", outerDistance, outerOk)
	}
}
