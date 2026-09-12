package scanner

import (
	"fmt"
	"tp-lox/internal/token"
)

type Scanner struct {
	tokens  []token.Token
	source  string
	current int
}

// NewScanner crea un nuevo scanner, a partir de un string fuente a escanear.
func NewScanner(source string) *Scanner {
	return &Scanner{
		tokens: []token.Token{},
		source: source,
	}
}

// Scan escanea el string fuente y devuelve un slice de tokens y un error si encuentra un caracter inesperado.
func (s *Scanner) Scan() ([]token.Token, error) {
	for s.current < len(s.source) {
		if err := s.scanToken(); err != nil {
			return s.tokens, err
		}
	}
	s.addToken(token.EOF, "")
	return s.tokens, nil
}

// scanToken escanea un token del string fuente y lo agrega a la lista de tokens.
// Devuelve error si encuentra un caracter inesperado.
func (s *Scanner) scanToken() error {
	currentChar := s.advance()

	switch currentChar {
	// Tokens de un solo caracter
	case '(':
		s.addToken(token.LEFT_PAREN, string(currentChar))
	case ')':
		s.addToken(token.RIGHT_PAREN, string(currentChar))
	case '{':
		s.addToken(token.LEFT_BRACE, string(currentChar))
	case '}':
		s.addToken(token.RIGHT_BRACE, string(currentChar))
	case ',':
		s.addToken(token.COMMA, string(currentChar))
	case '.':
		s.addToken(token.DOT, string(currentChar))
	case '-':
		s.addToken(token.MINUS, string(currentChar))
	case '+':
		s.addToken(token.PLUS, string(currentChar))
	case ';':
		s.addToken(token.SEMICOLON, string(currentChar))
	case '*':
		s.addToken(token.STAR, string(currentChar))
	case '%':
		s.addToken(token.PERCENT, string(currentChar))

	// Tokens de uno o dos caracteres
	case '/':
		if s.isNext('/') {
			s.advanceAllLine() // Comentario, ignorar hasta el final de la línea
		} else {
			s.addToken(token.SLASH, string(currentChar))
		}
	case '!':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.BANG_EQUAL, "!="), *token.NewToken(token.BANG, "!"))
	case '=':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.EQUAL_EQUAL, "=="), *token.NewToken(token.EQUAL, "="))
	case '>':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.GREATER_EQUAL, ">="), *token.NewToken(token.GREATER, ">"))
	case '<':
		s.addSingleOrDoubleCharToken('=', *token.NewToken(token.LESS_EQUAL, "<="), *token.NewToken(token.LESS, "<"))
	default:
		// faltan manejar los literales y palabras reservadas todavia
		return fmt.Errorf("error: unexpected character: %c", currentChar)
	}
	return nil
}

func (s *Scanner) addToken(tokenType token.TokenType, lexeme string) {
	s.tokens = append(s.tokens, token.Token{TokenType: tokenType, Lexeme: lexeme})
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
	if s.current+1 >= len(s.source) {
		return false
	}
	return s.source[s.current+1] == expected
}

// addSingleOrDoubleCharToken agrega un token de uno o dos caracteres dependiendo de si el siguiente caracter es el esperado.
func (s *Scanner) addSingleOrDoubleCharToken(nextExpected byte, doubleToken token.Token, singleToken token.Token) {
	if s.isNext(nextExpected) {
		s.advance()
		s.addToken(doubleToken.GetTokenType(), doubleToken.GetLexeme())
	} else {
		s.addToken(singleToken.GetTokenType(), singleToken.GetLexeme())
	}
}
