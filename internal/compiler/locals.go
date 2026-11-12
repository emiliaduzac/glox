package compiler

import "tp-lox/internal/token"

type Local struct {
	Name  token.Token
	Depth int
}
