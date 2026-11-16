package compiler

import "tp-lox/internal/token"

type Precedence int

const (
	PREC_NONE       Precedence = iota
	PREC_ASSIGNMENT            // =
	PREC_OR                    // or
	PREC_AND                   // and
	PREC_EQUALITY              // == !=
	PREC_COMPARISON            // < > <= >=
	PREC_TERM                  // + -
	PREC_FACTOR                // * / %
	PREC_UNARY                 // ! -
	PREC_CALL                  // . ()
	PREC_PRIMARY
)

type parseFn func(c *Compiler, canAssign bool) error

type parseRule struct {
	prefix     parseFn    // si el token EMPIEZA una expresión
	infix      parseFn    // si el token aparece DESPUÉS de un operando
	precedence Precedence // qué tan fuerte es como operador infijo
}

var rules map[token.TokenType]parseRule

func init() {
	rules = map[token.TokenType]parseRule{
		token.LEFT_PAREN:    {prefix: (*Compiler).grouping, infix: (*Compiler).call, precedence: PREC_CALL},
		token.DOT:           {prefix: nil, infix: (*Compiler).dot, precedence: PREC_CALL},
		token.NUMBER:        {prefix: (*Compiler).value, infix: nil, precedence: PREC_NONE},
		token.STRING:        {prefix: (*Compiler).value, infix: nil, precedence: PREC_NONE},
		token.NIL:           {prefix: (*Compiler).literal, infix: nil, precedence: PREC_NONE},
		token.FALSE:         {prefix: (*Compiler).literal, infix: nil, precedence: PREC_NONE},
		token.TRUE:          {prefix: (*Compiler).literal, infix: nil, precedence: PREC_NONE},
		token.MINUS:         {prefix: (*Compiler).unary, infix: (*Compiler).binary, precedence: PREC_TERM},
		token.PLUS:          {prefix: nil, infix: (*Compiler).binary, precedence: PREC_TERM},
		token.STAR:          {prefix: nil, infix: (*Compiler).binary, precedence: PREC_FACTOR},
		token.SLASH:         {prefix: nil, infix: (*Compiler).binary, precedence: PREC_FACTOR},
		token.PERCENT:       {prefix: nil, infix: (*Compiler).binary, precedence: PREC_FACTOR},
		token.BANG:          {prefix: (*Compiler).unary, infix: nil, precedence: PREC_NONE},
		token.BANG_EQUAL:    {prefix: nil, infix: (*Compiler).binary, precedence: PREC_EQUALITY},
		token.EQUAL_EQUAL:   {prefix: nil, infix: (*Compiler).binary, precedence: PREC_EQUALITY},
		token.GREATER:       {prefix: nil, infix: (*Compiler).binary, precedence: PREC_COMPARISON},
		token.GREATER_EQUAL: {prefix: nil, infix: (*Compiler).binary, precedence: PREC_COMPARISON},
		token.LESS:          {prefix: nil, infix: (*Compiler).binary, precedence: PREC_COMPARISON},
		token.LESS_EQUAL:    {prefix: nil, infix: (*Compiler).binary, precedence: PREC_COMPARISON},
		token.AND:           {prefix: nil, infix: (*Compiler).logicAnd, precedence: PREC_AND},
		token.OR:            {prefix: nil, infix: (*Compiler).logicOr, precedence: PREC_OR},
		token.IDENTIFIER:    {prefix: (*Compiler).variable, infix: nil, precedence: PREC_NONE},
		token.THIS:          {prefix: (*Compiler).this, infix: nil, precedence: PREC_NONE},
		token.SUPER:         {prefix: (*Compiler).super, infix: nil, precedence: PREC_NONE},
	}
}

func getRule(t token.TokenType) parseRule {
	return rules[t]
}
