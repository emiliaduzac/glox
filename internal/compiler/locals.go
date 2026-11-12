package compiler

import "tp-lox/internal/token"

type Local struct {
	name  token.Token
	depth int
}
