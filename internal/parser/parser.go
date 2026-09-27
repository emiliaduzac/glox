package parser

import (
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
		return p.printStatement()
	}
	if p.isInTokenTypes(token.LEFT_BRACE) {
		stmts, err := p.block()
		if err != nil {
			return nil, err
		}
		return statements.NewBlockStmt(stmts), nil
	}
	if p.isInTokenTypes(token.IF) {
		return p.ifStatement()
	}
	if p.isInTokenTypes(token.WHILE) {
		return p.whileStatement()
	}
	if p.isInTokenTypes(token.FOR) {
		return p.forStatement()
	}

	return p.expressionStatement()
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
	}

	return statements.NewIfStmt(condition, thenBranch, elseBranch), nil
}

func (p *Parser) whileStatement() (statements.Stmt, error) {
	if !p.isInTokenTypes(token.LEFT_PAREN) {
		return nil, p.expectedError("'(' after 'while'")
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.isInTokenTypes(token.RIGHT_PAREN) {
		return nil, p.expectedError("')' after while condition")
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return statements.NewWhileStmt(condition, body), nil
}

func (p *Parser) forStatement() (statements.Stmt, error) {
	if !p.isInTokenTypes(token.LEFT_PAREN) {
		return nil, p.expectedError("'(' after 'for'")
	}

	// obtenemos las partes del for: inicializador, condición e incremento
	initializer, err := p.parseForInitializer()
	if err != nil {
		return nil, err
	}
	condition, err := p.parseForCondition()
	if err != nil {
		return nil, err
	}
	increment, err := p.parseForIncrement()
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	// de-sugaring: convertimos el for en un while con las partes obtenidas
	if increment != nil {
		body = statements.NewBlockStmt([]statements.Stmt{
			body,
			statements.NewExpressionStmt(increment),
		})
	}
	if condition == nil {
		condition = expression.NewLiteral(true)
	}
	body = statements.NewWhileStmt(condition, body)
	if initializer != nil {
		body = statements.NewBlockStmt([]statements.Stmt{
			initializer,
			body,
		})
	}

	return body, nil
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
	expr, err := p.or()
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

func (p *Parser) or() (expression.Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.OR) {
		operator := p.getPrevious()
		right, err := p.and()
		if err != nil {
			return nil, err
		}
		expr = expression.NewLogical(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) and() (expression.Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.AND) {
		operator := p.getPrevious()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = expression.NewLogical(expr, operator, right)
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

	return p.call()
}

func (p *Parser) call() (expression.Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.isInTokenTypes(token.LEFT_PAREN) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee expression.Expr) (expression.Expr, error) {
	var arguments []expression.Expr

	if !p.isTokenType(token.RIGHT_PAREN) {
		for {
			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, arg)
			if !p.isInTokenTypes(token.COMMA) {
				break
			}
		}
	}

	if !p.isInTokenTypes(token.RIGHT_PAREN) {
		return nil, p.expectedError("')' after arguments")
	}

	paren := p.getPrevious()
	return expression.NewCall(callee, paren, arguments), nil
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
