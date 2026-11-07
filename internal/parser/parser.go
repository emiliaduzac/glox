package parser

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

type Parser struct {
	cursor *token.TokenCursor
}

// NewParser crea un nuevo parser, a partir de una lista de tokens generados por el scanner.
func NewParser(tokens []token.Token) *Parser {
	return &Parser{
		cursor: token.NewTokenCursor(tokens),
	}
}

func (p *Parser) Parse() ([]statements.Stmt, error) {
	stmts := []statements.Stmt{}
	for !p.cursor.IsAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}

func (p *Parser) declaration() (statements.Stmt, error) {
	if p.cursor.Match(token.FUN) {
		return p.function("function")
	}
	if p.cursor.Match(token.VAR) {
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
	if p.cursor.Match(token.PRINT) {
		return p.printStatement()
	}
	if p.cursor.Match(token.LEFT_BRACE) {
		stmts, err := p.block()
		if err != nil {
			return nil, err
		}
		return statements.NewBlockStmt(stmts), nil
	}
	if p.cursor.Match(token.IF) {
		return p.ifStatement()
	}
	if p.cursor.Match(token.WHILE) {
		return p.whileStatement()
	}
	if p.cursor.Match(token.FOR) {
		return p.forStatement()
	}
	if p.cursor.Match(token.RETURN) {
		return p.returnStatement()
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

	for !p.cursor.Check(token.RIGHT_BRACE) && !p.cursor.IsAtEnd() {
		decl, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, decl)
	}

	if err := p.cursor.Consume(token.RIGHT_BRACE, "'}' after block"); err != nil {
		return nil, err
	}

	return statements, nil
}

func (p *Parser) varDeclaration() (statements.Stmt, error) {
	if err := p.cursor.Consume(token.IDENTIFIER, "identifier after 'var'"); err != nil {
		return nil, err
	}

	varName := p.cursor.Previous()

	var initializer expression.Expr
	if p.cursor.Match(token.EQUAL) {
		var err error
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	} else {
		initializer = nil
	}

	if err := p.cursor.Consume(token.SEMICOLON, "';' after variable declaration"); err != nil {
		return nil, err
	}

	return statements.NewVarStmt(varName, initializer), nil
}

func (p *Parser) ifStatement() (statements.Stmt, error) {
	if err := p.cursor.Consume(token.LEFT_PAREN, "'(' after 'if'"); err != nil {
		return nil, err
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if err := p.cursor.Consume(token.RIGHT_PAREN, "')' after if condition"); err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch statements.Stmt
	if p.cursor.Match(token.ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return statements.NewIfStmt(condition, thenBranch, elseBranch), nil
}

func (p *Parser) whileStatement() (statements.Stmt, error) {
	if err := p.cursor.Consume(token.LEFT_PAREN, "'(' after 'while'"); err != nil {
		return nil, err
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if err := p.cursor.Consume(token.RIGHT_PAREN, "')' after while condition"); err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return statements.NewWhileStmt(condition, body), nil
}

func (p *Parser) forStatement() (statements.Stmt, error) {
	if err := p.cursor.Consume(token.LEFT_PAREN, "'(' after 'for'"); err != nil {
		return nil, err
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
	if err := p.cursor.Consume(token.SEMICOLON, "';' after statement"); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) returnStatement() (statements.Stmt, error) {
	keyword := p.cursor.Previous()
	var value expression.Expr
	if !p.cursor.Check(token.SEMICOLON) {
		var err error
		value, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if err := p.cursor.Consume(token.SEMICOLON, "';' after return value"); err != nil {
		return nil, err
	}

	return statements.NewReturnStmt(keyword, value), nil
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

	if p.cursor.Match(token.EQUAL) {
		equals := p.cursor.Previous()

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

	for p.cursor.Match(token.OR) {
		operator := p.cursor.Previous()
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

	for p.cursor.Match(token.AND) {
		operator := p.cursor.Previous()
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

	for p.cursor.Match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.cursor.Previous()
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

	for p.cursor.Match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.cursor.Previous()
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

	for p.cursor.Match(token.MINUS, token.PLUS) {
		operator := p.cursor.Previous()
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

	for p.cursor.Match(token.SLASH, token.STAR, token.PERCENT) {
		operator := p.cursor.Previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) unary() (expression.Expr, error) {
	if p.cursor.Match(token.BANG, token.MINUS) {
		operator := p.cursor.Previous()
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
		if p.cursor.Match(token.LEFT_PAREN) {
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

	if !p.cursor.Check(token.RIGHT_PAREN) {
		for {
			if len(arguments) >= 255 {
				return nil, p.cursor.ExpectedError("cannot have more than 255 arguments")
			}
			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, arg)
			if !p.cursor.Match(token.COMMA) {
				break
			}
		}
	}

	if err := p.cursor.Consume(token.RIGHT_PAREN, "')' after arguments"); err != nil {
		return nil, err
	}

	paren := p.cursor.Previous()
	return expression.NewCall(callee, paren, arguments), nil
}

func (p *Parser) primary() (expression.Expr, error) {
	if p.cursor.Match(token.FALSE) {
		return expression.NewLiteral(false), nil
	}
	if p.cursor.Match(token.TRUE) {
		return expression.NewLiteral(true), nil
	}
	if p.cursor.Match(token.NIL) {
		return expression.NewLiteral(nil), nil
	}

	if p.cursor.Match(token.NUMBER, token.STRING) {
		return expression.NewLiteral(p.cursor.Previous().Literal), nil
	}

	if p.cursor.Match(token.LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if !p.cursor.Match(token.RIGHT_PAREN) {
			return nil, errors.NewLoxError(p.cursor.GetLine(), "expected ')' after expression")
		}
		return expression.NewGrouping(expr), nil
	}

	if p.cursor.Match(token.IDENTIFIER) {
		return expression.NewVariable(p.cursor.Previous()), nil
	}

	return nil, p.cursor.ExpectedError("expression")
}

func (p *Parser) function(kind string) (statements.Stmt, error) {
	if err := p.cursor.Consume(token.IDENTIFIER, fmt.Sprintf("expect %s name", kind)); err != nil {
		return nil, err
	}

	name := p.cursor.Previous()

	if err := p.cursor.Consume(token.LEFT_PAREN, fmt.Sprintf("'(' after %s name", kind)); err != nil {
		return nil, err
	}

	var parameters []token.Token
	if !p.cursor.Check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				return nil, p.cursor.ExpectedError("cannot have more than 255 parameters")
			}
			if err := p.cursor.Consume(token.IDENTIFIER, "expect parameter name"); err != nil {
				return nil, err
			}
			parameters = append(parameters, p.cursor.Previous())
			if !p.cursor.Match(token.COMMA) {
				break
			}
		}
	}

	if err := p.cursor.Consume(token.RIGHT_PAREN, "')' after parameters"); err != nil {
		return nil, err
	}

	if err := p.cursor.Consume(token.LEFT_BRACE, fmt.Sprintf("'{' before %s body", kind)); err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return statements.NewFunctionStmt(name, parameters, body), nil
}
