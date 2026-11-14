package scanner

import (
	"testing"
	"tp-lox/internal/token"
	"tp-lox/internal/utils"
)

// HELPERS

func setUpScanner(t *testing.T, source string) []token.Token {
	t.Helper()
	s := NewScanner(source)
	tokens, err := s.Scan()
	utils.AssertNoError(t, err)
	return tokens
}

func assertToken(t *testing.T, token token.Token, expectedType token.TokenType, expectedLexeme string, expectedLiteral any) {
	t.Helper()
	if token.TokenType != expectedType {
		t.Errorf("expected token type %v, got %v", expectedType.String(), token.TokenType.String())
	}
	if token.Lexeme != expectedLexeme {
		t.Errorf("expected lexeme %q, got %q", expectedLexeme, token.Lexeme)
	}
	if token.Literal != expectedLiteral {
		t.Errorf("expected literal %v, got %v", expectedLiteral, token.Literal)
	}
}

func assertAllTokens(t *testing.T, tokens []token.Token, expectedTypes []token.TokenType, expectedLexems []string, expectedLiterals []any) {
	t.Helper()
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}
	for i, expectedType := range expectedTypes {
		if tokens[i].TokenType != expectedType {
			t.Errorf("expected token type %v at index %d, got %v", expectedType.String(), i, tokens[i].TokenType.String())
		}
	}
	for i, expectedLexeme := range expectedLexems {
		if tokens[i].Lexeme != expectedLexeme {
			t.Errorf("expected lexeme %q at index %d, got %q", expectedLexeme, i, tokens[i].Lexeme)
		}
	}
	for i, expectedLiteral := range expectedLiterals {
		if tokens[i].Literal != expectedLiteral {
			t.Errorf("expected literal %v at index %d, got %v", expectedLiteral, i, tokens[i].Literal)
		}
	}
}

func assertEndsWithEOF(t *testing.T, tokens []token.Token) {
	t.Helper()
	lastToken := tokens[len(tokens)-1]
	if lastToken.TokenType != token.EOF {
		t.Errorf("expected last token to be EOF, got %v", lastToken.TokenType.String())
	}
}

func assertInvalid(t *testing.T, source string, expectedErrorContent string) {
	t.Helper()
	s := NewScanner(source)
	_, err := s.Scan()
	utils.AssertError(t, err)
	if expectedErrorContent != "" {
		utils.AssertErrorMessage(t, err, expectedErrorContent)
	}
}

// TESTS -- Casos base válidos

func TestSimpleTokens(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected token.TokenType
	}{
		{"parentesis izquierdo", "(", token.LEFT_PAREN},
		{"parentesis derecho", ")", token.RIGHT_PAREN},
		{"llave izquierda", "{", token.LEFT_BRACE},
		{"llave derecha", "}", token.RIGHT_BRACE},
		{"coma", ",", token.COMMA},
		{"punto", ".", token.DOT},
		{"menos", "-", token.MINUS},
		{"mas", "+", token.PLUS},
		{"punto y coma", ";", token.SEMICOLON},
		{"asterisco", "*", token.STAR},
		{"porcentaje", "%", token.PERCENT},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertToken(t, tokens[0], test.expected, test.source, nil)
			assertEndsWithEOF(t, tokens)
		})
	}
}

func TestSingleCharacterTokens(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected token.TokenType
	}{
		{"slash", "/", token.SLASH},
		{"igual", "=", token.EQUAL},
		{"exclamacion", "!", token.BANG},
		{"mayor", ">", token.GREATER},
		{"menor", "<", token.LESS},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertToken(t, tokens[0], test.expected, test.source, nil)
			assertEndsWithEOF(t, tokens)
		})
	}
}

func TestDoubleCharacterTokens(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected token.TokenType
	}{
		{"igual igual", "==", token.EQUAL_EQUAL},
		{"diferente", "!=", token.BANG_EQUAL},
		{"mayor igual", ">=", token.GREATER_EQUAL},
		{"menor igual", "<=", token.LESS_EQUAL},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertToken(t, tokens[0], test.expected, test.source, nil)
			assertEndsWithEOF(t, tokens)
		})
	}
}

func TestValidStrings(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{"string vacío", `""`, ""},
		{"string de una palabra", `"Hello"`, "Hello"},
		{"string de varias palabras", `"Hello World"`, "Hello World"},
		{"string multi-linea", "\"Hello \nWorld\"", "Hello \nWorld"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertToken(t, tokens[0], token.STRING, test.source, test.expected)
			assertEndsWithEOF(t, tokens)
		})
	}
}

func TestValidNumbers(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected float64
	}{
		{"unumero de n digito", "5", 5.0},
		{"numero de varios digitos", "12345", 12345.0},
		{"numero decimal", "3.14", 3.14},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertToken(t, tokens[0], token.NUMBER, test.source, test.expected)
			assertEndsWithEOF(t, tokens)
		})
	}
}

func TestValidKeyWords(t *testing.T) {
	tests := []struct {
		source   string
		expected token.TokenType
	}{
		{"and", token.AND},
		{"class", token.CLASS},
		{"else", token.ELSE},
		{"false", token.FALSE},
		{"fun", token.FUN},
		{"for", token.FOR},
		{"if", token.IF},
		{"nil", token.NIL},
		{"or", token.OR},
		{"print", token.PRINT},
		{"return", token.RETURN},
		{"super", token.SUPER},
		{"this", token.THIS},
		{"true", token.TRUE},
		{"var", token.VAR},
		{"while", token.WHILE},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertToken(t, tokens[0], test.expected, test.source, nil)
			assertEndsWithEOF(t, tokens)
		})
	}
}

func TestValidIdentifiers(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"identificador compuesto de letras", "variable"},
		{"identificador compuesto de letras, inicia con _", "_variable"},
		{"identificador compuesto de letras y numeros", "variable1"},
		{"identificador compuesto de letras, numeros y _", "var_1"},
		{"identificador compuesto de letras finaliza con _", "variable_"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertToken(t, tokens[0], token.IDENTIFIER, test.source, nil)
			assertEndsWithEOF(t, tokens)
		})
	}
}

func TestIgnoredCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"comentario de una linea", "// esto es un comentario"},
		{"espacios en blanco", "     "},
		{"tabulaciones", "\t\t\t"},
		{"saltos de linea", "\n\n\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := setUpScanner(t, test.source)

			assertEndsWithEOF(t, tokens)
			if len(tokens) != 1 {
				t.Errorf("expected only EOF token, got %d tokens", len(tokens))
			}
		})
	}
}

func TestEmptySource(t *testing.T) {
	tokens := setUpScanner(t, "")

	assertEndsWithEOF(t, tokens)
}

// TESTS -- Casos combinación de tokens

func TestOperation(t *testing.T) {
	tokens := setUpScanner(t, "1 + 2")
	expectedTypes := []token.TokenType{token.NUMBER, token.PLUS, token.NUMBER, token.EOF}
	expectedLexemes := []string{"1", "+", "2", ""}
	expectedLiterals := []any{1.0, nil, 2.0, nil}

	assertAllTokens(t, tokens, expectedTypes, expectedLexemes, expectedLiterals)
}

func TestVariableAssignment(t *testing.T) {
	tokens := setUpScanner(t, "var x= 10;")
	expectedTypes := []token.TokenType{token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.SEMICOLON, token.EOF}
	expectedLexemes := []string{"var", "x", "=", "10", ";", ""}
	expectedLiterals := []any{nil, nil, nil, 10.0, nil, nil}

	assertAllTokens(t, tokens, expectedTypes, expectedLexemes, expectedLiterals)
}

// TESTS -- Casos inválidos

func TestInvalidStrings(t *testing.T) {
	tests := []struct {
		name                 string
		source               string
		expectedErrorContent string
	}{
		{"string comilla simple", `'Hello'`, "unexpected character"},
		{"string vacio sin comilla final", `"`, "unterminated string"},
		{"sin comilla final", `"Hello`, "unterminated string"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertInvalid(t, test.source, test.expectedErrorContent)
		})
	}
}

func TestInvalidNumbers(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"finaliza con punto", "123."},
		{"muchos puntos", "1.2.3"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertInvalid(t, test.source, "invalid number format")
		})
	}
}

func TestInvalidCharacters(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"arroba", "@"},
		{"pesos", "$"},
		{"comilla simple", "'"},
		{"backslash", "\\"},
		{"pregunta apertura", "¿"},
		{"pregunta cierre", "?"},
		{"hashtag", "#"},
		{"flecha", "^"},
		{"pipe", "|"},
		{"exclamacion apertura", "¡"},
		{"corchete apertura", "["},
		{"corchete cierre", "]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertInvalid(t, test.source, "unexpected character")
		})
	}
}

func TestContinueAndBreak(t *testing.T) {
	t.Run("continue es una palabra reservada", func(t *testing.T) {
		tokens := setUpScanner(t, "continue")

		assertToken(t, tokens[0], token.CONTINUE, "continue", nil)
		assertEndsWithEOF(t, tokens)
	})

	t.Run("break es una palabra reservada", func(t *testing.T) {
		tokens := setUpScanner(t, "break")

		assertToken(t, tokens[0], token.BREAK, "break", nil)
		assertEndsWithEOF(t, tokens)
	})
}
