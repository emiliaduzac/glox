package parser

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

type Parser struct {
	tokens  []token.Token
	current int
}

// NewParser crea un nuevo parser, a partir de una lista de tokens generados por el scanner.
func NewParser(tokens []token.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

func (p *Parser) Parse() ([]statements.Stmt, error) {
	stmts := []statements.Stmt{}
	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}

func (p *Parser) declaration() (statements.Stmt, error) {
	if p.isInTokenTypes(token.VAR) {
		stmt, err := p.varDeclaration()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	stmt, err := p.statement()
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) statement() (statements.Stmt, error) {
	if p.isInTokenTypes(token.PRINT) {
		stmt, err := p.printStatement()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if p.isInTokenTypes(token.LEFT_BRACE) {
		stmts, err := p.block()
		if err != nil {
			return nil, err
		}
		return statements.NewBlockStmt(stmts), nil
	}
	if p.isInTokenTypes(token.IF) {
		stmt, err := p.ifStatement()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}

	stmt, err := p.expressionStatement()
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) printStatement() (statements.Stmt, error) {
	expr, err := p.getStatement()
	if err != nil {
		return nil, err
	}
	return statements.NewPrintStmt(expr), nil
}

func (p *Parser) expressionStatement() (statements.Stmt, error) {
	expr, err := p.getStatement()
	if err != nil {
		return nil, err
	}
	return statements.NewExpressionStmt(expr), nil
}

func (p *Parser) block() ([]statements.Stmt, error) {
	var statements []statements.Stmt

	for !p.isTokenType(token.RIGHT_BRACE) && !p.isAtEnd() {
		decl, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, decl)
	}

	if !p.isInTokenTypes(token.RIGHT_BRACE) {
		return nil, p.expectedError("'}' after block")
	}

	return statements, nil
}

func (p *Parser) varDeclaration() (statements.Stmt, error) {
	if !p.isInTokenTypes(token.IDENTIFIER) {
		return nil, p.expectedError("identifier after 'var'")
	}

	varName := p.getPrevious()

	var initializer expression.Expr
	if p.isInTokenTypes(token.EQUAL) {
		var err error
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	} else {
		initializer = nil
	}

	if !p.isInTokenTypes(token.SEMICOLON) {
		return nil, p.expectedError("';' after variable declaration")
	}

	return statements.NewVarStmt(varName, initializer), nil
}

func (p *Parser) ifStatement() (statements.Stmt, error) {
	if !p.isInTokenTypes(token.LEFT_PAREN) {
		return nil, p.expectedError("'(' after 'if'")
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.isInTokenTypes(token.RIGHT_PAREN) {
		return nil, p.expectedError("')' after if condition")
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch statements.Stmt
	if p.isInTokenTypes(token.ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	} else {
		elseBranch = nil
	}

	return statements.NewIfStmt(condition, thenBranch, elseBranch), nil
}

func (p *Parser) getStatement() (expression.Expr, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.isInTokenTypes(token.SEMICOLON) {
		return nil, p.expectedError("';' after statement")
	}
	return expr, nil
}

// expression es el punto de entrada para parsear una expresión. Comienza el proceso
// de llamadas recursivas para parsear la expresión completa, siguiendo la jerarquía de
// precedencia de operadores.
func (p *Parser) expression() (expression.Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (expression.Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	if p.isInTokenTypes(token.EQUAL) {
		equals := p.getPrevious()

		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		if varExpr, ok := expr.(*expression.Variable); ok {
			return expression.NewAssign(varExpr.Name, value), nil
		}

		return nil, errors.NewLoxError(equals.Line, "invalid assignment target")
	}

	return expr, nil
}

func (p *Parser) equality() (expression.Expr, error) {
	var expr, err = p.comparison()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.getPrevious()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) comparison() (expression.Expr, error) {
	var expr, err = p.term()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.getPrevious()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) term() (expression.Expr, error) {
	var expr, err = p.factor()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.MINUS, token.PLUS) {
		operator := p.getPrevious()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) factor() (expression.Expr, error) {
	var expr, err = p.unary()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.SLASH, token.STAR, token.PERCENT) {
		operator := p.getPrevious()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) unary() (expression.Expr, error) {
	if p.isInTokenTypes(token.BANG, token.MINUS) {
		operator := p.getPrevious()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return expression.NewUnary(operator, right), nil
	}

	return p.primary()
}

func (p *Parser) primary() (expression.Expr, error) {
	if p.isInTokenTypes(token.FALSE) {
		return expression.NewLiteral(false), nil
	}
	if p.isInTokenTypes(token.TRUE) {
		return expression.NewLiteral(true), nil
	}
	if p.isInTokenTypes(token.NIL) {
		return expression.NewLiteral(nil), nil
	}

	if p.isInTokenTypes(token.NUMBER, token.STRING) {
		return expression.NewLiteral(p.getPrevious().Literal), nil
	}

	if p.isInTokenTypes(token.LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if !p.isInTokenTypes(token.RIGHT_PAREN) {
			return nil, errors.NewLoxError(p.getLine(), "expected ')' after expression")
		}
		return expression.NewGrouping(expr), nil
	}

	if p.isInTokenTypes(token.IDENTIFIER) {
		return expression.NewVariable(p.getPrevious()), nil
	}

	return nil, p.expectedError("expression")
}

// isAtEnd indica si el token actual es el último.
func (p *Parser) isAtEnd() bool {
	return p.tokens[p.current].TokenType == token.EOF
}

// isInTokenTypes verifica si el token actual es de alguno de los tipos de token pasados como argumento
func (p *Parser) isInTokenTypes(types ...token.TokenType) bool {
	for _, t := range types {
		if p.isTokenType(t) {
			p.advance()
			return true
		}
	}
	return false
}

// isTokenType verifica si el token actual es del tipo de token pasado como argumento
func (p *Parser) isTokenType(t token.TokenType) bool {
	return p.getCurrent().TokenType == t
}

func (p *Parser) getCurrent() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) getPrevious() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) getLine() int {
	return p.getCurrent().Line
}

func (p *Parser) getLexeme() string {
	return p.getCurrent().Lexeme
}

func (p *Parser) expectedError(expected string) error {
	got := fmt.Sprintf("`%s`", p.getLexeme())
	if p.isTokenType(token.EOF) {
		got = "EOF"
	}
	return errors.NewLoxError(p.getLine(), fmt.Sprintf("expected %s, got %s instead", expected, got))
}

// advance avanza al siguiente token y devuelve el token consumido
func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.getPrevious()
}
