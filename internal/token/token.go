package token

type TokenType int

const (
	LEFT_PAREN TokenType = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	STAR
	PERCENT
	SLASH

	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	IDENTIFIER
	STRING
	NUMBER

	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	EOF
)

var TokenKeyWords = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"fun":    FUN,
	"for":    FOR,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   any
}

// NewToken crea un nuevo token con el tipo y lexema indicados
func NewToken(tokenType TokenType, lexeme string, literal any) *Token {
	return &Token{
		TokenType: tokenType,
		Lexeme:    lexeme,
		Literal:   literal,
	}
}

// GetLexeme devuelve el lexema del token
func (t Token) GetLexeme() string {
	return t.Lexeme
}

// GetTokenType devuelve el tipo del token
func (t Token) GetTokenType() TokenType {
	return t.TokenType
}

// GetLiteral devuelve el literal del token
func (t Token) GetLiteral() any {
	return t.Literal
}
