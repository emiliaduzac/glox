package parser

import (
	"strings"
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

// HELPERS

var printToken = *token.NewToken(token.PRINT, "print", nil, 1)
var semicolonToken = *token.NewToken(token.SEMICOLON, ";", nil, 1)
var varToken = *token.NewToken(token.VAR, "var", nil, 1)
var ifToken = *token.NewToken(token.IF, "if", nil, 1)
var elseToken = *token.NewToken(token.ELSE, "else", nil, 1)
var leftParen = *token.NewToken(token.LEFT_PAREN, "(", nil, 1)
var rightParen = *token.NewToken(token.RIGHT_PAREN, ")", nil, 1)
var trueToken = *token.NewToken(token.TRUE, "true", nil, 1)
var falseToken = *token.NewToken(token.FALSE, "false", nil, 1)

func setUpParser(t *testing.T, tokens []token.Token) []statements.Stmt {
	t.Helper()
	tokens = append(tokens, *token.NewToken(token.EOF, "", nil, 1))
	s := NewParser(tokens)
	stmts, err := s.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return stmts
}

func setUpParserExpectError(t *testing.T, tokens []token.Token, expectedErrorContent string) {
	t.Helper()
	tokens = append(tokens, *token.NewToken(token.EOF, "", nil, 1))
	s := NewParser(tokens)
	_, err := s.Parse()
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), expectedErrorContent) {
		t.Errorf("expected error to contain %q, got %q", expectedErrorContent, err.Error())
	}
}

func assertExpressionStatement(t *testing.T, stmts []statements.Stmt) expression.Expr {
	t.Helper()
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	stmtExpr, ok := stmts[0].(*statements.ExpressionStmt)
	if !ok {
		t.Fatalf("expected an ExpressionStmt, got %T", stmts[0])
	}
	if stmtExpr.Expr == nil {
		t.Fatalf("expected a non nil expression in ExpressionStmt")
	}
	return stmtExpr.Expr
}

func assertPrintStatement(t *testing.T, stmts []statements.Stmt) expression.Expr {
	t.Helper()
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	stmtExpr, ok := stmts[0].(*statements.PrintStmt)
	if !ok {
		t.Fatalf("expected a PrintStmt, got %T", stmts[0])
	}
	if stmtExpr.Expr == nil {
		t.Fatalf("expected a non nil expression in PrintStmt")
	}
	return stmtExpr.Expr
}

func assertVariableStatement(t *testing.T, stmts []statements.Stmt, expectedName string) expression.Expr {
	t.Helper()
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	stmtVar, ok := stmts[0].(*statements.VarStmt)
	if !ok {
		t.Fatalf("expected a VariableStmt, got %T", stmts[0])
	}
	if stmtVar.Name.Lexeme != expectedName {
		t.Errorf("expected variable name %q, got %q", expectedName, stmtVar.Name.Lexeme)
	}
	return stmtVar.Initializer
}

func assertLiteral(t *testing.T, expr expression.Expr, expectedValue any) {
	t.Helper()

	literal, ok := expr.(*expression.Literal)
	if !ok {
		t.Fatalf("expected a Literal expression, got %T", expr)
	}
	if literal.Value != expectedValue {
		t.Errorf("expected literal value %#v, got %#v", expectedValue, literal.Value)
	}
}

func assertUnary(t *testing.T, expr expression.Expr, expectedOperator token.TokenType, expectedRight any) {
	t.Helper()

	unary, ok := expr.(*expression.Unary)
	if !ok {
		t.Fatalf("expected a Unary expression, got %T", expr)
	}
	if unary.Operator.TokenType != expectedOperator {
		t.Errorf("expected operator %s, got %s", expectedOperator, unary.Operator.TokenType)
	}
	assertLiteral(t, unary.Right, expectedRight)
}

func assertBinary(t *testing.T, expr expression.Expr, expectedOperator token.TokenType, expectedLeft any, expectedRight any) {
	t.Helper()

	binary, ok := expr.(*expression.Binary)
	if !ok {
		t.Fatalf("expected a Binary expression, got %T", expr)
	}
	if binary.Operator.TokenType != expectedOperator {
		t.Errorf("expected operator %s, got %s", expectedOperator, binary.Operator.TokenType)
	}
	assertLiteral(t, binary.Left, expectedLeft)
	assertLiteral(t, binary.Right, expectedRight)
}

func assertString(t *testing.T, expr expression.Expr, expected string) {
	t.Helper()

	if got := expression.String(expr); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func assertStatementString(t *testing.T, stmts []statements.Stmt, expected string) {
	t.Helper()

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	if got := statements.String(stmts[0]); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
