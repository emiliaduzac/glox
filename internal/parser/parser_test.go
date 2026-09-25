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

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name+"_expr", func(t *testing.T) {
			stmts := setUpParser(t, []token.Token{*test.token, semicolonToken})
			expr := assertExpressionStatement(t, stmts)
			assertLiteral(t, expr, test.expected)
		})
	}

	// Como print statements
	for _, test := range tests {
		t.Run(test.name+"_prt", func(t *testing.T) {
			stmts := setUpParser(t, []token.Token{printToken, *test.token, semicolonToken})
			expr := assertPrintStatement(t, stmts)
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

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name+"_expr", func(t *testing.T) {
			tokens := append(test.tokens, semicolonToken)
			stmts := setUpParser(t, tokens)
			expr := assertExpressionStatement(t, stmts)
			assertUnary(t, expr, test.operator, test.expected)
		})
	}

	// Como print statements
	for _, test := range tests {
		t.Run(test.name+"_prt", func(t *testing.T) {
			tokens := append([]token.Token{printToken}, test.tokens...)
			tokens = append(tokens, semicolonToken)
			stmts := setUpParser(t, tokens)
			expr := assertPrintStatement(t, stmts)
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

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name+"_expr", func(t *testing.T) {
			tokens := append(test.tokens, semicolonToken)
			stmts := setUpParser(t, tokens)
			expr := assertExpressionStatement(t, stmts)
			assertBinary(t, expr, test.operator, test.left, test.right)
		})
	}

	// Como print statements
	for _, test := range tests {
		t.Run(test.name+"_prt", func(t *testing.T) {
			tokens := append([]token.Token{printToken}, test.tokens...)
			tokens = append(tokens, semicolonToken)
			stmts := setUpParser(t, tokens)
			expr := assertPrintStatement(t, stmts)
			assertBinary(t, expr, test.operator, test.left, test.right)
		})
	}
}

func TestParseGrouping(t *testing.T) {
	tokens := []token.Token{
		*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
		*token.NewToken(token.NUMBER, "123", 123.0, 1),
		*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
		semicolonToken,
	}

	// Como expression statements
	t.Run("grouping_expr", func(t *testing.T) {
		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		grouping, ok := expr.(*expression.Grouping)
		if !ok {
			t.Fatalf("expected a Grouping expression, got %T", expr)
		}
		assertLiteral(t, grouping.Expr, 123.0)
	})

	// Como print statements
	t.Run("grouping_prt", func(t *testing.T) {
		stmts := setUpParser(t, append([]token.Token{printToken}, tokens...))
		expr := assertPrintStatement(t, stmts)
		grouping, ok := expr.(*expression.Grouping)
		if !ok {
			t.Fatalf("expected a Grouping expression, got %T", expr)
		}
		assertLiteral(t, grouping.Expr, 123.0)
	})
}

// TESTS -- Casos de asociatividad y precedencia de operadores

func TestPrecedenceEquality(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		expected string
	}{
		{
			"comparison es prioritario que equality",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.LESS, "<", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				*token.NewToken(token.LESS, "<", nil, 1),
				*token.NewToken(token.NUMBER, "4", 4.0, 1),
				semicolonToken,
			},
			"((<1> LESS <2>) EQUAL_EQUAL (<3> LESS <4>))",
		},
		{
			"asociatividad izquierda",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"((<1> EQUAL_EQUAL <2>) EQUAL_EQUAL <3>)",
		},
	}

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmts := setUpParser(t, test.tokens)
			expr := assertExpressionStatement(t, stmts)
			assertString(t, expr, test.expected)
		})
	}
}

func TestPrecedenceComparison(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		expected string
	}{
		{
			"term es mas prioritario que comparison",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.PLUS, "+", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.LESS, "<", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				*token.NewToken(token.PLUS, "+", nil, 1),
				*token.NewToken(token.NUMBER, "4", 4.0, 1),
				semicolonToken,
			},
			"((<1> PLUS <2>) LESS (<3> PLUS <4>))",
		},
		{
			"asociatividad izquierda",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.LESS, "<", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.LESS, "<", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"((<1> LESS <2>) LESS <3>)",
		},
	}

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmts := setUpParser(t, test.tokens)
			expr := assertExpressionStatement(t, stmts)
			assertString(t, expr, test.expected)
		})
	}
}

func TestPrecedenceTerm(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		expected string
	}{
		{
			"factor es mas prioritario que term",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.PLUS, "+", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.STAR, "*", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"(<1> PLUS (<2> STAR <3>))",
		},
		{
			"asociatividad izquierda",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.MINUS, "-", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.MINUS, "-", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"((<1> MINUS <2>) MINUS <3>)",
		},
	}

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmts := setUpParser(t, test.tokens)
			expr := assertExpressionStatement(t, stmts)
			assertString(t, expr, test.expected)
		})
	}
}

func TestPrecedenceFactor(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		expected string
	}{
		{
			"unary es as prioritario que factor",
			[]token.Token{
				*token.NewToken(token.MINUS, "-", nil, 1),
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.STAR, "*", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				semicolonToken,
			},
			"((MINUS <1>) STAR <2>)",
		},
		{
			"asociatividad izquierda",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.STAR, "*", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.SLASH, "/", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"((<1> STAR <2>) SLASH <3>)",
		},
	}

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmts := setUpParser(t, test.tokens)
			expr := assertExpressionStatement(t, stmts)
			assertString(t, expr, test.expected)
		})
	}
}

func TestPrecedenceGrouping(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		expected string
	}{
		{
			"grouping rompe la precedencia default",
			[]token.Token{
				*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.PLUS, "+", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
				*token.NewToken(token.STAR, "*", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"((group (<1> PLUS <2>)) STAR <3>)",
		},
	}

	// Como expression statements
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmts := setUpParser(t, test.tokens)
			expr := assertExpressionStatement(t, stmts)
			assertString(t, expr, test.expected)
		})
	}
}

// TESTS -- Casos de error

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
		{
			"parentesis sin cerrar (expresion incompleta)",
			[]token.Token{
				*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected ')'",
		},
		{
			"token inesperado",
			[]token.Token{
				*token.NewToken(token.PLUS, "+", nil, 1),
				semicolonToken,
			},
			"expected expression, got `+`",
		},
		{
			"falta operando derecho (expresion incompleta)",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.PLUS, "+", nil, 1),
				semicolonToken,
			},
			"expected expression, got `;`",
		},
		{
			"print sin ';'",
			[]token.Token{
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
			},
			"expected ';' after statement",
		},
		{
			"expression statement sin ';'",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
			},
			"expected ';' after statement",
		},
		{
			"print sin expresion",
			[]token.Token{
				printToken,
				semicolonToken,
			},
			"expected expression, got `;`",
		},
		{
			"print de print",
			[]token.Token{
				printToken,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected expression, got `print`",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setUpParserExpectError(t, test.tokens, test.expectedErrorContent)
		})
	}
}
