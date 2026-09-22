package scanner

import (
	"fmt"
	"strconv"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

type Scanner struct {
	tokens  []token.Token
	source  string
	current int
	start   int
	line    int
}

// NewScanner crea un nuevo scanner, a partir de un string fuente a escanear.
func NewScanner(source string) *Scanner {
	return &Scanner{
		tokens:  []token.Token{},
		source:  source,
		current: 0,
		start:   0,
		line:    1,
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
	case '\n':
		s.line++
	case ' ', '\r', '\t':
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
		s.addSingleOrDoubleCharToken('=', token.BANG_EQUAL, token.BANG)
	case '=':
		s.addSingleOrDoubleCharToken('=', token.EQUAL_EQUAL, token.EQUAL)
	case '>':
		s.addSingleOrDoubleCharToken('=', token.GREATER_EQUAL, token.GREATER)
	case '<':
		s.addSingleOrDoubleCharToken('=', token.LESS_EQUAL, token.LESS)

	// Literales cadena
	case '"':
		return s.findStringLiteral()

	default:
		if s.isDigit(currentChar) { // Literales numericos
			return s.findNumberLiteral()
		} else if s.isAlphaOrUnderscore(currentChar) { // Identificadores y palabras clave
			s.findIdentifierOrKeyword()
		} else {
			return errors.NewLoxError(s.line, fmt.Sprintf("unexpected character: `%c`", currentChar))
		}
	}
	return nil
}

// isAtEnd verifica si el scanner llegó al final del string fuente (current ya no apunta a ningun caracter)
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) isAlphaOrUnderscore(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
}

func (s *Scanner) addToken(tokenType token.TokenType, lexeme string, literal any) {
	s.tokens = append(s.tokens, token.Token{TokenType: tokenType, Lexeme: lexeme, Literal: literal, Line: s.line})
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

func (s *Scanner) getCurrentLexeme() string {
	return s.source[s.start:s.current]
}

// addSingleOrDoubleCharToken agrega un token de uno o dos caracteres dependiendo de si el siguiente caracter es el esperado.
func (s *Scanner) addSingleOrDoubleCharToken(nextExpected byte, doubleTokenType token.TokenType, singleTokenType token.TokenType) {
	if s.isNext(nextExpected) {
		s.advance()
		s.addToken(doubleTokenType, s.getCurrentLexeme(), nil)
	} else {
		s.addToken(singleTokenType, s.getCurrentLexeme(), nil)
	}
}

// findStringLiteral busca un literal de cadena en el string fuente.
// Devuelve error si encuentra un literal de cadena sin terminar.
func (s *Scanner) findStringLiteral() error {
	for !s.isAtEnd() && !s.isNext('"') {
		if s.source[s.current] == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isNext('"') {
		s.advance() // comilla de cierre
		lexeme := s.getCurrentLexeme()
		literal := lexeme[1 : len(lexeme)-1]
		s.addToken(token.STRING, lexeme, literal)
	} else {
		return errors.NewLoxError(s.line, "unterminated string")
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
				return errors.NewLoxError(s.line, "invalid number format: multiple decimal points")
			}
			foundDot = true
		}
	}

	// Verificamos que el ultimo caracter leído no sea un punto decimal
	if s.source[s.current-1] == '.' {
		return errors.NewLoxError(s.line, "invalid number format: cannot end with a decimal point")
	}

	lexeme := s.getCurrentLexeme()
	literal, err := strconv.ParseFloat(lexeme, 64)
	if err != nil {
		return errors.NewLoxError(s.line, fmt.Sprintf("invalid number format: `%v`", err))
	}
	s.addToken(token.NUMBER, lexeme, literal)
	return nil
}

// findIdentifierOrKeyword busca un identificador o palabra clave en el string fuente.
func (s *Scanner) findIdentifierOrKeyword() {
	// Avanzamos mientras el caracter actual sea alfanumérico o un guion bajo
	for !s.isAtEnd() && (s.isAlphaOrUnderscore(s.source[s.current]) || s.isDigit(s.source[s.current])) {
		s.advance()
	}

	lexeme := s.getCurrentLexeme()
	if tokenType, ok := token.TokenKeyWords[lexeme]; ok {
		s.addToken(tokenType, lexeme, nil)
	} else {
		s.addToken(token.IDENTIFIER, lexeme, nil)
	}
}
