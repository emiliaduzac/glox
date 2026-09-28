package parser

import (
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

// Tests de la gramática de expresiones: literales, unary, binary, grouping, assignment y sus
// reglas de precedencia/asociatividad.

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
			stmts := setUpParser(t, []token.Token{*test.token, semicolonToken})
			expr := assertExpressionStatement(t, stmts)
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
			tokens := append(test.tokens, semicolonToken)
			stmts := setUpParser(t, tokens)
			expr := assertExpressionStatement(t, stmts)
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
			tokens := append(test.tokens, semicolonToken)
			stmts := setUpParser(t, tokens)
			expr := assertExpressionStatement(t, stmts)
			assertBinary(t, expr, test.operator, test.left, test.right)
		})
	}
}

func TestParseBinaryErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setUpParserExpectError(t, test.tokens, test.expectedErrorContent)
		})
	}
}

func TestParseGrouping(t *testing.T) {
	tokens := []token.Token{
		leftParen,
		*token.NewToken(token.NUMBER, "123", 123.0, 1),
		rightParen,
		semicolonToken,
	}

	stmts := setUpParser(t, tokens)
	expr := assertExpressionStatement(t, stmts)
	grouping, ok := expr.(*expression.Grouping)
	if !ok {
		t.Fatalf("expected a Grouping expression, got %T", expr)
	}
	assertLiteral(t, grouping.Expr, 123.0)
}

func TestParseGroupingErrors(t *testing.T) {
	tokens := []token.Token{
		*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
		*token.NewToken(token.NUMBER, "1", 1.0, 1),
		semicolonToken,
	}

	t.Run("parentesis sin cerrar (expresion incompleta)", func(t *testing.T) {
		setUpParserExpectError(t, tokens, "expected ')'")
	})
}

func TestParseAssignment(t *testing.T) {

	t.Run("Simple assignment", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			*token.NewToken(token.EQUAL, "=", nil, 1),
			*token.NewToken(token.NUMBER, "5", 5.0, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(= a <5>)")
	})

	t.Run("Assignment with binary expression on right", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			*token.NewToken(token.EQUAL, "=", nil, 1),
			*token.NewToken(token.NUMBER, "5", 5.0, 1),
			*token.NewToken(token.PLUS, "+", nil, 1),
			*token.NewToken(token.NUMBER, "3", 3.0, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(= a (<5> PLUS <3>))")
	})
}

func TestParseAssignmentErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
		{
			"invalid assignment target on arithmetic expression",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.PLUS, "+", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				*token.NewToken(token.EQUAL, "=", nil, 1),
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"invalid assignment target",
		},
		{
			"invalid assignment target on grouped variable",
			[]token.Token{
				*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
				*token.NewToken(token.IDENTIFIER, "a", nil, 1),
				*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
				*token.NewToken(token.EQUAL, "=", nil, 1),
				*token.NewToken(token.NUMBER, "5", 5.0, 1),
				semicolonToken,
			},
			"invalid assignment target",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setUpParserExpectError(t, test.tokens, test.expectedErrorContent)
		})
	}
}

func TestParseLogical(t *testing.T) {
	t.Run("or", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			orToken,
			*token.NewToken(token.NUMBER, "2", 2.0, 1),
			semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(<1> OR <2>)")
	})

	t.Run("and", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			andToken,
			*token.NewToken(token.NUMBER, "2", 2.0, 1),
			semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(<1> AND <2>)")
	})
}

func TestParseLogicalErrors(t *testing.T) {
	tokens := []token.Token{
		*token.NewToken(token.NUMBER, "1", 1.0, 1),
		orToken,
		semicolonToken,
	}

	t.Run("falta operando derecho (expresion incompleta)", func(t *testing.T) {
		setUpParserExpectError(t, tokens, "expected expression, got `;`")
	})
}

func TestPrecedenceAssignment(t *testing.T) {
	t.Run("Assignment is right associative", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			*token.NewToken(token.EQUAL, "=", nil, 1),
			*token.NewToken(token.IDENTIFIER, "b", nil, 1),
			*token.NewToken(token.EQUAL, "=", nil, 1),
			*token.NewToken(token.NUMBER, "5", 5.0, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(= a (= b <5>))")
	})

	t.Run("Assignment has lower precedence than equality", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			*token.NewToken(token.EQUAL, "=", nil, 1),
			*token.NewToken(token.IDENTIFIER, "b", nil, 1),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			*token.NewToken(token.NUMBER, "5", 5.0, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(= a (<b> EQUAL_EQUAL <5>))")
	})
}

func TestPrecedenceLogical(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		expected string
	}{
		{
			"and es mas prioritario que or",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				orToken,
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				andToken,
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"(<1> OR (<2> AND <3>))",
		},
		{
			"equality es mas prioritario que and",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				andToken,
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
				*token.NewToken(token.NUMBER, "4", 4.0, 1),
				semicolonToken,
			},
			"((<1> EQUAL_EQUAL <2>) AND (<3> EQUAL_EQUAL <4>))",
		},
		{
			"asociatividad izquierda de or",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				orToken,
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				orToken,
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"((<1> OR <2>) OR <3>)",
		},
		{
			"asociatividad izquierda de and",
			[]token.Token{
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				andToken,
				*token.NewToken(token.NUMBER, "2", 2.0, 1),
				andToken,
				*token.NewToken(token.NUMBER, "3", 3.0, 1),
				semicolonToken,
			},
			"((<1> AND <2>) AND <3>)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmts := setUpParser(t, test.tokens)
			expr := assertExpressionStatement(t, stmts)
			assertString(t, expr, test.expected)
		})
	}
}

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmts := setUpParser(t, test.tokens)
			expr := assertExpressionStatement(t, stmts)
			assertString(t, expr, test.expected)
		})
	}
}

func TestPrecedenceGrouping(t *testing.T) {
	tokens := []token.Token{
		*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
		*token.NewToken(token.NUMBER, "1", 1.0, 1),
		*token.NewToken(token.PLUS, "+", nil, 1),
		*token.NewToken(token.NUMBER, "2", 2.0, 1),
		*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
		*token.NewToken(token.STAR, "*", nil, 1),
		*token.NewToken(token.NUMBER, "3", 3.0, 1),
		semicolonToken,
	}

	t.Run("grouping rompe la precedencia default", func(t *testing.T) {
		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "((group (<1> PLUS <2>)) STAR <3>)")
	})

}

func TestCallExpressions(t *testing.T) {
	t.Run("Call without arguments", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "foo", nil, 1),
			*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(call <foo>)")
	})

	t.Run("Call with arguments", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "add", nil, 1),
			*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			*token.NewToken(token.COMMA, ",", nil, 1),
			*token.NewToken(token.NUMBER, "2", 2.0, 1),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(call <add> <1> <2>)")
	})

	t.Run("Chained calls", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "getCallback", nil, 1),
			*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(call (call <getCallback>))")
	})

	t.Run("Call precedence higher than unary", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.BANG, "!", nil, 1),
			*token.NewToken(token.IDENTIFIER, "fn", nil, 1),
			*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			semicolonToken,
		}

		stmts := setUpParser(t, tokens)
		expr := assertExpressionStatement(t, stmts)
		assertString(t, expr, "(BANG (call <fn>))")
	})

	t.Run("Missing right paren error", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "foo", nil, 1),
			*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			semicolonToken,
		}

		setUpParserExpectError(t, tokens, "')' after arguments")
	})

	t.Run("Exceeds max arguments error", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.IDENTIFIER, "foo", nil, 1),
			*token.NewToken(token.LEFT_PAREN, "(", nil, 1),
		}

		for i := 0; i < 256; i++ {
			tokens = append(tokens, *token.NewToken(token.NUMBER, "1", 1.0, 1))
			if i < 255 {
				tokens = append(tokens, *token.NewToken(token.COMMA, ",", nil, 1))
			}
		}
		tokens = append(tokens, *token.NewToken(token.RIGHT_PAREN, ")", nil, 1), semicolonToken)

		setUpParserExpectError(t, tokens, "cannot have more than 255 arguments")
	})
}
