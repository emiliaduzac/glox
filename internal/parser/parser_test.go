package parser

import (
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

// HELPERS

func setUpParser(t *testing.T, tokens []token.Token) expression.Expr {
	t.Helper()
	tokens = append(tokens, *token.NewToken(token.EOF, "", nil, 1))
	s := NewParser(tokens)
	expr, err := s.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return expr
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

// TESTS -- Casos base válidos

func TestParseLiterals(t *testing.T) {
	tests := []struct {
		name     string
		token    *token.Token
		expected any
	}{
		{"numeros", token.NewToken(token.NUMBER, "123", 123.0, 1), 123.0},
		{"cadenas", token.NewToken(token.STRING, "hola", "hola", 1), "hola"},
		{"true", token.NewToken(token.TRUE, "true", nil, 1), true},
		{"false", token.NewToken(token.FALSE, "false", nil, 1), false},
		{"nil", token.NewToken(token.NIL, "nil", nil, 1), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr := setUpParser(t, []token.Token{*test.token})
			assertLiteral(t, expr, test.expected)
		})
	}
}

func TestParseUnary(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		operator token.TokenType
		expected any
	}{
		{
			"exclamacion",
			[]token.Token{*token.NewToken(token.BANG, "!", nil, 1), *token.NewToken(token.NUMBER, "123", 123.0, 1)},
			token.BANG,
			123.0,
		},
		{
			"menos",
			[]token.Token{*token.NewToken(token.MINUS, "-", nil, 1), *token.NewToken(token.NUMBER, "123", 123.0, 1)},
			token.MINUS,
			123.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr := setUpParser(t, test.tokens)
			assertUnary(t, expr, test.operator, test.expected)
		})
	}
}

func TestParseBinary(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		operator token.TokenType
		left     any
		right    any
	}{
		{
			"suma",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.PLUS, "+", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.PLUS,
			1.0,
			2.0,
		},
		{
			"resta",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.MINUS, "-", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.MINUS,
			1.0,
			2.0,
		},
		{
			"multiplicacion",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.STAR, "*", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.STAR,
			1.0,
			2.0,
		},
		{
			"division",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.SLASH, "/", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.SLASH,
			1.0,
			2.0,
		},
		{
			"modulo",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.PERCENT, "%", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.PERCENT,
			1.0,
			2.0,
		},
		{
			"igualdad",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.EQUAL_EQUAL, "==", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.EQUAL_EQUAL,
			1.0,
			2.0,
		},
		{
			"desigualdad",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.BANG_EQUAL, "!=", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.BANG_EQUAL,
			1.0,
			2.0,
		},
		{
			"mayor",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.GREATER, ">", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.GREATER,
			1.0,
			2.0,
		},
		{
			"mayor_o_igual",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.GREATER_EQUAL, ">=", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.GREATER_EQUAL,
			1.0,
			2.0,
		},
		{
			"menor",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.LESS, "<", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.LESS,
			1.0,
			2.0,
		},
		{
			"menor_o_igual",
			[]token.Token{*token.NewToken(token.NUMBER, "1", 1.0, 1), *token.NewToken(token.LESS_EQUAL, "<=", nil, 1), *token.NewToken(token.NUMBER, "2", 2.0, 1)},
			token.LESS_EQUAL,
			1.0,
			2.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr := setUpParser(t, test.tokens)
			assertBinary(t, expr, test.operator, test.left, test.right)
		})
	}
}

func TestParseGrouping(t *testing.T) {
	tokens := []token.Token{
		*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
		*token.NewToken(token.NUMBER, "123", 123.0, 1),
		*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
	}

	expr := setUpParser(t, tokens)

	grouping, ok := expr.(*expression.Grouping)
	if !ok {
		t.Fatalf("expected a Grouping expression, got %T", expr)
	}
	assertLiteral(t, grouping.Expr, 123.0)
}

// TESTS -- Casos de asociatividad y precedencia de operadores
func TestPrecedence(t *testing.T) {

}
