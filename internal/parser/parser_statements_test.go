package parser

import (
	"testing"
	"tp-lox/internal/token"
)

// Tests de la gramática de statements: print, expression-statement, var, block, if, etc.
// Cubre la sintaxis propia de cada statement, no la gramática de las expresiones que
// envuelven.

func TestParsePrint(t *testing.T) {
	t.Run("print of a literal", func(t *testing.T) {
		tokens := []token.Token{
			printToken,
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		expr := assertPrintStatement(t, stmts)
		assertLiteral(t, expr, 1.0)
	})

	t.Run("print wraps an arbitrary expression", func(t *testing.T) {
		tokens := []token.Token{
			printToken,
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			*token.NewToken(token.PLUS, "+", nil, 1),
			*token.NewToken(token.NUMBER, "2", 2.0, 1),
			semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		expr := assertPrintStatement(t, stmts)
		assertBinary(t, expr, token.PLUS, 1.0, 2.0)
	})
}

func TestParsePrintErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
		{
			"print sin ';'",
			[]token.Token{
				printToken,
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

func TestParseExpressionStatementErrors(t *testing.T) {
	tokens := []token.Token{
		*token.NewToken(token.NUMBER, "1", 1.0, 1),
	}

	t.Run("expression statement sin ';'", func(t *testing.T) {
		setUpParserExpectError(t, tokens, "expected ';' after statement")
	})
}

func TestParseVariables(t *testing.T) {
	nonInitializedVar := []token.Token{
		varToken,
		*token.NewToken(token.IDENTIFIER, "x", nil, 1),
		semicolonToken,
	}

	t.Run("non_initialized_var", func(t *testing.T) {
		stmts := setUpParser(t, nonInitializedVar)
		expr := assertVariableStatement(t, stmts, "x")
		if expr != nil {
			t.Errorf("expected nil initializer, got %v", expr)
		}
	})

	initializedVar := []token.Token{
		varToken,
		*token.NewToken(token.IDENTIFIER, "x", nil, 1),
		*token.NewToken(token.EQUAL, "=", nil, 1),
		*token.NewToken(token.NUMBER, "42", 42.0, 1),
		semicolonToken,
	}

	t.Run("initialized_var", func(t *testing.T) {
		stmts := setUpParser(t, initializedVar)
		expr := assertVariableStatement(t, stmts, "x")
		assertLiteral(t, expr, 42.0)
	})
}

func TestParseVariablesErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
		{
			"var sin identificador",
			[]token.Token{
				varToken,
				semicolonToken,
			},
			"expected identifier after 'var', got `;`",
		},
		{
			"var con '=' sin expresion",
			[]token.Token{
				varToken,
				*token.NewToken(token.IDENTIFIER, "x", nil, 1),
				*token.NewToken(token.EQUAL, "=", nil, 1),
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

func TestParseBlock(t *testing.T) {
	t.Run("Empty block", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(block)")
	})

	t.Run("Block with statements", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			varToken,
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			*token.NewToken(token.EQUAL, "=", nil, 1),
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			semicolonToken,
			printToken,
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			semicolonToken,
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(block (var IDENTIFIER<a> <1>) (print <a>))")
	})

	t.Run("Block with expression statement", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			*token.NewToken(token.PLUS, "+", nil, 1),
			*token.NewToken(token.NUMBER, "2", 2.0, 1),
			semicolonToken,
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(block (expr (<1> PLUS <2>)))")
	})

	t.Run("Nested blocks", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(block (block))")
	})

	t.Run("Deeply nested blocks with declaration", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			varToken,
			*token.NewToken(token.IDENTIFIER, "x", nil, 1),
			semicolonToken,
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(block (block (block (var IDENTIFIER<x>))))")
	})
}

func TestParseBlockErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
		{
			"unclosed empty block",
			[]token.Token{
				*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			},
			"expected '}' after block",
		},
		{
			"unclosed block with declaration",
			[]token.Token{
				*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
				varToken,
				*token.NewToken(token.IDENTIFIER, "a", nil, 1),
				semicolonToken,
			},
			"expected '}' after block",
		},
		{
			"unclosed nested block",
			[]token.Token{
				*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
				*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
				*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
			},
			"expected '}' after block",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setUpParserExpectError(t, test.tokens, test.expectedErrorContent)
		})
	}
}

func TestParseIf(t *testing.T) {
	t.Run("If without else", func(t *testing.T) {
		tokens := []token.Token{
			ifToken, leftParen, trueToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(if <TRUE> (print <1>))")
	})

	t.Run("If with else", func(t *testing.T) {
		tokens := []token.Token{
			ifToken, leftParen, trueToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
			elseToken,
			printToken, *token.NewToken(token.NUMBER, "2", 2.0, 1), semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(if <TRUE> (print <1>) (print <2>))")
	})

	t.Run("If with block as then branch", func(t *testing.T) {
		tokens := []token.Token{
			ifToken, leftParen, trueToken, rightParen,
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(if <TRUE> (block (print <1>)))")
	})

	t.Run("Else followed by another if", func(t *testing.T) {
		tokens := []token.Token{
			ifToken, leftParen, trueToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
			elseToken,
			ifToken, leftParen, falseToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "2", 2.0, 1), semicolonToken,
			elseToken,
			printToken, *token.NewToken(token.NUMBER, "3", 3.0, 1), semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(if <TRUE> (print <1>) (if <FALSE> (print <2>) (print <3>)))")
	})

	t.Run("Else belongs to the closest if", func(t *testing.T) {
		tokens := []token.Token{
			ifToken, leftParen, trueToken, rightParen,
			ifToken, leftParen, falseToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
			elseToken,
			printToken, *token.NewToken(token.NUMBER, "2", 2.0, 1), semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(if <TRUE> (if <FALSE> (print <1>) (print <2>)))")
	})

	t.Run("Condition with assignment", func(t *testing.T) {
		tokens := []token.Token{
			ifToken, leftParen,
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			*token.NewToken(token.EQUAL, "=", nil, 1),
			*token.NewToken(token.NUMBER, "1", 1.0, 1),
			rightParen,
			printToken, *token.NewToken(token.IDENTIFIER, "a", nil, 1), semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(if (= a <1>) (print <a>))")
	})

	t.Run("If inside a block", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			ifToken, leftParen, trueToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(block (if <TRUE> (print <1>)))")
	})
}

func TestParseIfErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
		{
			"if without '(' after 'if'",
			[]token.Token{
				ifToken,
				trueToken,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected '(' after 'if', got `true`",
		},
		{
			"if without ')' after condition",
			[]token.Token{
				ifToken,
				leftParen,
				trueToken,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected ')' after if condition, got `print`",
		},
		{
			"if with empty condition",
			[]token.Token{
				ifToken,
				leftParen,
				rightParen,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected expression, got `)`",
		},
		{
			"if with declaration as then branch",
			[]token.Token{
				ifToken,
				leftParen,
				trueToken,
				rightParen,
				varToken,
				*token.NewToken(token.IDENTIFIER, "a", nil, 1),
				semicolonToken,
			},
			"expected expression, got `var`",
		},
		{
			"if without then branch",
			[]token.Token{
				ifToken,
				leftParen,
				trueToken,
				rightParen,
				elseToken,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected expression, got `else`",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setUpParserExpectError(t, test.tokens, test.expectedErrorContent)
		})
	}
}

func TestParseWhile(t *testing.T) {
	t.Run("While with print body", func(t *testing.T) {
		tokens := []token.Token{
			whileToken, leftParen, trueToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(while <TRUE> (print <1>))")
	})

	t.Run("While with block body", func(t *testing.T) {
		tokens := []token.Token{
			whileToken, leftParen, trueToken, rightParen,
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(while <TRUE> (block (print <1>)))")
	})

	t.Run("Condition with comparison", func(t *testing.T) {
		tokens := []token.Token{
			whileToken, leftParen,
			*token.NewToken(token.IDENTIFIER, "a", nil, 1),
			*token.NewToken(token.LESS, "<", nil, 1),
			*token.NewToken(token.NUMBER, "10", 10.0, 1),
			rightParen,
			printToken, *token.NewToken(token.IDENTIFIER, "a", nil, 1), semicolonToken,
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(while (<a> LESS <10>) (print <a>))")
	})

	t.Run("While inside a block", func(t *testing.T) {
		tokens := []token.Token{
			*token.NewToken(token.LEFT_BRACE, "{", nil, 1),
			whileToken, leftParen, trueToken, rightParen,
			printToken, *token.NewToken(token.NUMBER, "1", 1.0, 1), semicolonToken,
			*token.NewToken(token.RIGHT_BRACE, "}", nil, 1),
		}
		stmts := setUpParser(t, tokens)
		assertStatementString(t, stmts, "(block (while <TRUE> (print <1>)))")
	})
}

func TestParseWhileErrors(t *testing.T) {
	tests := []struct {
		name                 string
		tokens               []token.Token
		expectedErrorContent string
	}{
		{
			"while without '(' after 'while'",
			[]token.Token{
				whileToken,
				trueToken,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected '(' after 'while', got `true`",
		},
		{
			"while without ')' after condition",
			[]token.Token{
				whileToken,
				leftParen,
				trueToken,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected ')' after while condition, got `print`",
		},
		{
			"while with empty condition",
			[]token.Token{
				whileToken,
				leftParen,
				rightParen,
				printToken,
				*token.NewToken(token.NUMBER, "1", 1.0, 1),
				semicolonToken,
			},
			"expected expression, got `)`",
		},
		{
			"while with declaration as body",
			[]token.Token{
				whileToken,
				leftParen,
				trueToken,
				rightParen,
				varToken,
				*token.NewToken(token.IDENTIFIER, "a", nil, 1),
				semicolonToken,
			},
			"expected expression, got `var`",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setUpParserExpectError(t, test.tokens, test.expectedErrorContent)
		})
	}
}
