package compiler

import "tp-lox/internal/token"

type FunctionType int

const (
	TYPE_FUNCTION FunctionType = iota
	TYPE_SCRIPT
	TYPE_METHOD
	TYPE_INITIALIZER
)

type parserState struct {
	cursor       *token.TokenCursor
	errs         []error
	panicMode    bool
	currentClass *classCompiler
}

type Local struct {
	Name       token.Token
	Depth      int
	IsCaptured bool
}

type Loop struct {
	loopStart  int
	scopeDepth int
	breakJumps []int
}

type Upvalue struct {
	index   byte
	isLocal bool
}

type classCompiler struct {
	enclosing     *classCompiler
	hasSuperclass bool
}

var (
	thisToken  = token.Token{TokenType: token.THIS, Lexeme: "this"}
	superToken = token.Token{TokenType: token.SUPER, Lexeme: "super"}
)
