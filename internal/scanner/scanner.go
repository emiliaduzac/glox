package scanner

import (
	"fmt"
	"tp-lox/internal/token"
)

type Scanner struct {
	tokens  []token.Token
	source  string
	current int
	start   int
}

// NewScanner crea un nuevo scanner, a partir de un string fuente a escanear.
func NewScanner(source string) *Scanner {
	return &Scanner{
		tokens:  []token.Token{},
		source:  source,
		current: 0,
		start:   0,
	}
}

// Scan escanea el string fuente y devuelve un slice de tokens y un error si encuentra un caracter inesperado.
func (s *Scanner) Scan() ([]token.Token, error) {
	for !s.isAtEnd() {
		s.start = s.current
		if err := s.scanToken(); err != nil {
			return s.tokens, err
		}
	}
	s.addToken(token.EOF, "", nil)
	return s.tokens, nil
}

// scanToken escanea un token del string fuente y lo agrega a la lista de tokens.
// Devuelve error si encuentra un caracter inesperado.
func (s *Scanner) scanToken() error {
	currentChar := s.advance()

	switch currentChar {
	case ' ', '\r', '\t', '\n':
		// Ignorar espacios en blanco y saltos de línea (no necesitamos avanzar el índice ya que el advance() ya lo hizo)

	// Tokens de un solo caracter
	case '(':
		s.addToken(token.LEFT_PAREN, string(currentChar), nil)
	case ')':
		s.addToken(token.RIGHT_PAREN, string(currentChar), nil)
	case '{':
		s.addToken(token.LEFT_BRACE, string(currentChar), nil)
	case '}':
		s.addToken(token.RIGHT_BRACE, string(currentChar), nil)
	case ',':
		s.addToken(token.COMMA, string(currentChar), nil)
	case '.':
		s.addToken(token.DOT, string(currentChar), nil)
	case '-':
		s.addToken(token.MINUS, string(currentChar), nil)
	case '+':
		s.addToken(token.PLUS, string(currentChar), nil)
	case ';':
		s.addToken(token.SEMICOLON, string(currentChar), nil)
	case '*':
		s.addToken(token.STAR, string(currentChar), nil)
	case '%':
		s.addToken(token.PERCENT, string(currentChar), nil)

	// Tokens de uno o dos caracteres
	case '/':
		if s.isNext('/') {
			s.advanceAllLine() // Comentario, ignorar hasta el final de la línea
		} else {
			s.addToken(token.SLASH, string(currentChar), nil)
		}
	case '!':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.BANG_EQUAL, "!=", nil), *token.NewToken(token.BANG, "!", nil))
	case '=':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.EQUAL_EQUAL, "==", nil), *token.NewToken(token.EQUAL, "=", nil))
	case '>':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.GREATER_EQUAL, ">=", nil), *token.NewToken(token.GREATER, ">", nil))
	case '<':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.LESS_EQUAL, "<=", nil), *token.NewToken(token.LESS, "<", nil))

	// Literales
	case '"':
		return s.findStringLiteral()

	default:
		if s.isDigit(currentChar) {
			return s.findNumberLiteral()
		} else {
			// faltan manejar los literales identifiers todavia
			return fmt.Errorf("unexpected character: %c", currentChar)
		}
	}
	return nil
}

// isAtEnd verifica si el scanner llegó al final del string fuente.
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) addToken(tokenType token.TokenType, lexeme string, literal any) {
	s.tokens = append(s.tokens, token.Token{TokenType: tokenType, Lexeme: lexeme, Literal: literal})
}

func (s *Scanner) advance() byte {
	c := s.source[s.current]
	s.current++
	return c
}

func (s *Scanner) advanceAllLine() {
	for s.current < len(s.source) && s.source[s.current] != '\n' {
		s.current++
	}
}

// isNext verifica si el siguiente caracter es el esperado.
// Valida contra el indice actual ya que el advance() ya incrementó el índice.
func (s *Scanner) isNext(expected byte) bool {
	if s.current >= len(s.source) {
		return false
	}
	return s.source[s.current] == expected
}

// addSingleOrDoubleCharToken agrega un token de uno o dos caracteres dependiendo de si el siguiente caracter es el esperado.
func (s *Scanner) addSingleOrDoubleCharToken(nextExpected byte, doubleToken token.Token, singleToken token.Token) {
	if s.isNext(nextExpected) {
		s.advance()
		s.addToken(doubleToken.GetTokenType(), doubleToken.GetLexeme(), nil)
	} else {
		s.addToken(singleToken.GetTokenType(), singleToken.GetLexeme(), nil)
	}
}

// findStringLiteral busca un literal de cadena en el string fuente.
// Devuelve error si encuentra un literal de cadena sin terminar.
func (s *Scanner) findStringLiteral() error {
	for !s.isAtEnd() && !s.isNext('"') && !s.isNext('\n') {
		s.advance()
	}

	if s.isNext('"') {
		s.advance() // comilla de cierre
		lexeme := s.source[s.start:s.current]
		literal := lexeme[1 : len(lexeme)-1]
		s.addToken(token.STRING, lexeme, literal)
	} else {
		return fmt.Errorf("unterminated string")
	}
	return nil
}

func (s *Scanner) findNumberLiteral() error {
	foundDot := false
	// Avanzamos mientras el caracter actual sea un dígito o un punto decimal
	for !s.isAtEnd() && (s.isDigit(s.source[s.current]) || s.isNext('.')) {
		c := s.advance()
		// Si encontramos un punto decimal, verificamos si ya habíamos encontrado uno antes
		if c == '.' {
			if foundDot {
				return fmt.Errorf("invalid number format: multiple decimal points")
			}
			foundDot = true
		}
	}

	// Verificamos que el ultimo caracter leído no sea un punto decimal
	if s.source[s.current-1] == '.' {
		return fmt.Errorf("invalid number format: an not end with a decimal point")
	}

	lexeme := s.source[s.start:s.current]
	s.addToken(token.NUMBER, lexeme, lexeme)
	return nil
}
