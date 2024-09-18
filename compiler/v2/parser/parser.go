package parser

import (
	"fmt"
	"hack/compiler/v2/ast"
	"hack/compiler/v2/lexer"
	"hack/compiler/v2/token"
	"strconv"
)

type Parser struct {
	l              *lexer.Lexer
	currentToken   token.Token
	peekToken      token.Token
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type prefixParseFn func() (ast.Expression, error)
type infixParseFn func(ast.Expression) (ast.Expression, error)

func New(lexer *lexer.Lexer) *Parser {
	p := &Parser{l: lexer}
	p.nextToken()
	p.nextToken()

	//TODO: register prefixParseFn
	prefixParseFns := make(map[token.TokenType]prefixParseFn)
	prefixParseFns[token.TokenTypeIntegerLiteral] = p.parseIntegerLiteral
	prefixParseFns[token.TokenTypeStringLiteral] = p.parseStringLiteral
	prefixParseFns[token.TokenTypeIdentifier] = p.parseIdentifier
	prefixParseFns[token.TokenTypeTrue] = p.parseKeywordConstantLiteral
	prefixParseFns[token.TokenTypeFalse] = p.parseKeywordConstantLiteral
	prefixParseFns[token.TokenTypeNull] = p.parseKeywordConstantLiteral
	prefixParseFns[token.TokenTypeThis] = p.parseKeywordConstantLiteral
	prefixParseFns[token.TokenTypeLeftParenthesis] = p.parseParenthesisExpression
	prefixParseFns[token.TokenTypeMinus] = p.parsePrefixExpression
	prefixParseFns[token.TokenTypeTilde] = p.parsePrefixExpression
	p.prefixParseFns = prefixParseFns

	//TODO: register infixParseFn
	infixParseFns := make(map[token.TokenType]infixParseFn)
	infixParseFns[token.TokeTypeGreater] = p.parseInfixExpression
	infixParseFns[token.TokenTypeLess] = p.parseInfixExpression
	infixParseFns[token.TokenTypePlus] = p.parseInfixExpression
	infixParseFns[token.TokenTypeMinus] = p.parseInfixExpression
	infixParseFns[token.TokenTypeAsterisk] = p.parseInfixExpression
	infixParseFns[token.TokenTypeSlash] = p.parseInfixExpression
	infixParseFns[token.TokenTypeAmpersand] = p.parseInfixExpression
	infixParseFns[token.TokenTypeVerticalBar] = p.parseInfixExpression
	infixParseFns[token.TokenTypeAssign] = p.parseInfixExpression
	infixParseFns[token.TokenTypeDot] = p.parseObjectCall
	infixParseFns[token.TokenTypeLeftParenthesis] = p.parseCallExpression
	infixParseFns[token.TokenTypeLeftBracket] = p.parseIndexExpression
	p.infixParseFns = infixParseFns

	return p
}

func (p *Parser) ParseClass() (*ast.Class, error) {
	klass := &ast.Class{
		Fields:      make([]*ast.Field, 0),
		Subroutines: make([]*ast.Subroutine, 0),
	}
	if p.currentTokenIs(token.TokenTypeClass) {
		klass.Token = p.currentToken
		p.nextToken()

		identifier, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		klass.Identifier = identifier.(*ast.Identifier)
		if !p.expectPeek(token.TokenTypeLeftBrace) {
			return nil, fmt.Errorf("expected peek token to be left brace but found %s", p.currentToken.Literal)
		}
		p.nextToken()

		for p.currentTokenIs(token.TokenTypeStatic) || p.currentTokenIs(token.TokenTypeField) {
			f, err := p.parseField()
			if err != nil {
				return nil, err
			}
			klass.Fields = append(klass.Fields, f)
			p.nextToken()
		}

		for {
			if p.currentTokenIs(token.TokenTypeConstructor) || p.currentTokenIs(token.TokenTypeFunction) || p.currentTokenIs(token.TokenTypeMethod) {
				subroutine, err := p.parseSubroutine()
				if err != nil {
					return nil, err
				}
				klass.Subroutines = append(klass.Subroutines, subroutine)
			} else if p.currentTokenIs(token.TokenTypeRightBrace) {
				break
			} else {
				return nil, fmt.Errorf("expected current token to be right brace or subroutine but found %s", p.currentToken.TokenType)
			}
			p.nextToken()
		}

	} else {
		return nil, fmt.Errorf("expected start with class token")
	}

	return klass, nil
}
func (p *Parser) parseIdentifier() (ast.Expression, error) {
	if p.currentTokenIs(token.TokenTypeIdentifier) {
		return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}, nil
	}
	return nil, fmt.Errorf("expected current token to be identifier, got %s", p.currentToken.TokenType)
}

func (p *Parser) parseField() (*ast.Field, error) {
	f := &ast.Field{}
	if p.currentTokenIs(token.TokenTypeStatic) {
		f.Scope = ast.FieldScopeStatic
	} else if p.currentTokenIs(token.TokenTypeField) {
		f.Scope = ast.FieldScopeInstance
	} else {
		return nil, fmt.Errorf("expected field, got %s", p.currentToken.TokenType)
	}
	p.nextToken()

	t, err := p.parseType()
	if err != nil {
		return nil, err
	}
	f.Type = t
	p.nextToken()

	identifiers, err := p.parseIdentifiers()
	if err != nil {
		return nil, err
	}
	f.Identifiers = identifiers

	return f, nil
}

func (p *Parser) parseType() (string, error) {
	switch p.currentToken.TokenType {
	case token.TokenTypeInt:
		return "int", nil
	case token.TokenTypeBoolean:
		return "boolean", nil
	case token.TokenTypeChar:
		return "char", nil
	case token.TokenTypeIdentifier:
		return p.currentToken.Literal, nil
	default:
		return "", fmt.Errorf("expected int, boolean, char or identifier, got %s", p.currentToken.TokenType)
	}
}

func (p *Parser) parseIdentifiers() ([]*ast.Identifier, error) {
	idents := make([]*ast.Identifier, 0)
	if p.currentTokenIs(token.TokenTypeIdentifier) {
		idents = append(idents, &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal})
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected identifier, got %s", p.currentToken.TokenType)
	}

	for p.currentTokenIs(token.TokenTypeComma) {
		if p.expectPeek(token.TokenTypeIdentifier) {
			idents = append(idents, &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal})
			p.nextToken()
		} else {
			return nil, fmt.Errorf("expected peek token to be identifier, got %s", p.peekToken.TokenType)
		}
	}
	return idents, nil
}

func (p *Parser) parseSubroutine() (*ast.Subroutine, error) {
	subroutine := &ast.Subroutine{Token: p.currentToken}
	if p.currentTokenIs(token.TokenTypeConstructor) {
		subroutine.Type = ast.SubroutineTypeConstructor
	} else if p.currentTokenIs(token.TokenTypeFunction) {
		subroutine.Type = ast.SubroutineTypeFunction
	} else if p.currentTokenIs(token.TokenTypeMethod) {
		subroutine.Type = ast.SubroutineTypeMethod
	} else {
		return nil, fmt.Errorf("expected current token to be constructor, function, or method, got %s", p.currentToken.TokenType)
	}
	p.nextToken()

	if p.currentTokenIs(token.TokenTypeVoid) {
		subroutine.ReturnType = "void"
	} else {
		t, err := p.parseType()
		if err != nil {
			return nil, err
		}
		subroutine.ReturnType = t
	}
	p.nextToken()

	name, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}
	subroutine.Name = name.(*ast.Identifier)
	if !p.expectPeek(token.TokenTypeLeftParenthesis) {
		return nil, fmt.Errorf("expected peek token to be left parenthesis, got %s", p.peekToken.Literal)
	}
	p.nextToken()

	parameters, err := p.parseParameterList()
	if err != nil {
		return nil, err
	}
	subroutine.Parameters = parameters
	if !p.currentTokenIs(token.TokenTypeRightParenthesis) {
		return nil, fmt.Errorf("expected current token to be right parenthesis, got %s", p.currentToken.TokenType)
	}
	if !p.expectPeek(token.TokenTypeLeftBrace) {
		return nil, fmt.Errorf("expected peek token to be left brace but found %s", p.peekToken.TokenType)
	}

	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	subroutine.Body = body

	return subroutine, nil
}

func (p *Parser) parseParameterList() ([]*ast.Parameter, error) {
	parameters := make([]*ast.Parameter, 0)
	for !p.currentTokenIs(token.TokenTypeRightParenthesis) {
		para := &ast.Parameter{
			Token: p.currentToken,
		}
		t, err := p.parseType()
		if err != nil {
			return nil, err
		}
		para.Type = t
		p.nextToken()

		name, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		para.Name = name.(*ast.Identifier)
		parameters = append(parameters, para)
		p.nextToken()

		if p.currentTokenIs(token.TokenTypeComma) {
			p.nextToken()
		}
	}
	return parameters, nil
}
func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.currentToken.TokenType {
	case token.TokenTypeLet:
		return p.parseLetStatement()
	case token.TokenTypeDo:
		return p.parseDoStatement()
	case token.TokenTypeReturn:
		return p.parseReturnStatement()
	case token.TokenTypeIf:
		return p.parseIfStatement()
	case token.TokenTypeWhile:
		return p.parseWhileStatement()
	default:
		return nil, fmt.Errorf("failed to parse statement with %s", p.currentToken.TokenType)
	}
}

func (p *Parser) parseWhileStatement() (ast.Statement, error) {
	statement := &ast.WhileStatement{
		Token: p.currentToken,
	}
	if !p.expectPeek(token.TokenTypeLeftParenthesis) {
		return nil, fmt.Errorf("expected peek token to be left parenthesis, got %s", p.peekToken.TokenType)
	}
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	statement.Condition = exp

	if !p.expectPeek(token.TokenTypeRightParenthesis) {
		return nil, fmt.Errorf("expected peek token to be right parenthesis, got %s", p.peekToken.TokenType)
	}
	if !p.expectPeek(token.TokenTypeLeftBrace) {
		return nil, fmt.Errorf("expected peek token to be left brace but found %s", p.peekToken.TokenType)
	}
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	statement.Body = body
	if !p.currentTokenIs(token.TokenTypeRightBrace) {
		return nil, fmt.Errorf("expected current token to be right brace but found %s", p.peekToken.TokenType)
	}
	return statement, nil
}

func (p *Parser) parseIfStatement() (ast.Statement, error) {
	statement := &ast.IfStatement{
		Token: p.currentToken,
	}
	if !p.expectPeek(token.TokenTypeLeftParenthesis) {
		return nil, fmt.Errorf("expected peek token to be left parenthesis, got %s", p.peekToken.TokenType)
	}
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	statement.Condition = exp

	if !p.expectPeek(token.TokenTypeRightParenthesis) {
		return nil, fmt.Errorf("expected peek token to be right parenthesis, got %s", p.peekToken.TokenType)
	}
	if !p.expectPeek(token.TokenTypeLeftBrace) {
		return nil, fmt.Errorf("expected peek token to be left brace but found %s", p.peekToken.TokenType)
	}
	consequence, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	statement.Consequence = consequence
	if !p.currentTokenIs(token.TokenTypeRightBrace) {
		return nil, fmt.Errorf("expected current token to be right brace but found %s", p.peekToken.TokenType)
	}

	// TODO: handle else

	return statement, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	statement := &ast.ReturnStatement{Token: p.currentToken}
	p.nextToken()

	if !p.currentTokenIs(token.TokenTypeSemicolon) {
		exp, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		statement.Value = exp
		if !p.expectPeek(token.TokenTypeSemicolon) {
			return nil, fmt.Errorf("expected peek token to be semicolon got %s", p.peekToken.TokenType)
		}
	}
	return statement, nil
}

func (p *Parser) parseDoStatement() (*ast.DoStatement, error) {
	statement := &ast.DoStatement{Token: p.currentToken}
	p.nextToken()
	call, err := p.parseSubroutineCall()
	if err != nil {
		return nil, err
	}
	statement.SubroutineCall = call.(*ast.SubroutineCall)

	if !p.expectPeek(token.TokenTypeSemicolon) {
		return nil, fmt.Errorf("expected peek token to be semicolon got %s", p.peekToken.TokenType)
	}

	return statement, nil
}

func (p *Parser) parseSubroutineCall() (ast.Expression, error) {
	call := &ast.SubroutineCall{Token: p.currentToken}
	name, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}
	if p.peekTokenIs(token.TokenTypeDot) {
		call.CalleeName = name.(*ast.Identifier)
		p.nextToken()
		p.nextToken()
		name, err = p.parseIdentifier()
		if err != nil {
			return nil, err
		}
	}
	call.SubroutineName = name.(*ast.Identifier)

	if !p.expectPeek(token.TokenTypeLeftParenthesis) {
		return nil, fmt.Errorf("expected peek token to be left parenthesis, got %s", p.peekToken.TokenType)
	}

	expressions, err := p.parseExpressions()
	if err != nil {
		return nil, err
	}
	call.Arguments = expressions
	if !p.currentTokenIs(token.TokenTypeRightParenthesis) {
		return nil, fmt.Errorf("expected current token to be right parenthesis, got %s", p.peekToken.TokenType)
	}

	return call, nil
}

func (p *Parser) parseExpressions() ([]ast.Expression, error) {
	expressions := make([]ast.Expression, 0)
	p.nextToken()
	// TODO
	for !p.currentTokenIs(token.TokenTypeRightParenthesis) {
		exp, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, exp)
		if p.peekTokenIs(token.TokenTypeComma) {
			p.nextToken()
		}
		p.nextToken()
	}
	return expressions, nil
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, error) {
	let := &ast.LetStatement{Token: p.currentToken}
	p.nextToken()

	name, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}
	let.Name = name.(*ast.Identifier)
	p.nextToken()

	if p.currentTokenIs(token.TokenTypeLeftBracket) {
		p.nextToken()
		index, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		let.Index = index

		if !p.expectPeek(token.TokenTypeRightBracket) {
			return nil, fmt.Errorf("expected peek token to be right bracket, got %s", p.currentToken.TokenType)
		}
		p.nextToken()
	}
	if !p.currentTokenIs(token.TokenTypeAssign) {
		return nil, fmt.Errorf("expected current token to be assign got %s", p.peekToken.TokenType)
	}
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	let.Value = exp

	if !p.expectPeek(token.TokenTypeSemicolon) {
		return nil, fmt.Errorf("expected peek token to be semicolon got %s", p.peekToken.TokenType)
	}

	return let, nil
}

const (
	LOWEST uint8 = iota
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	DOT
	CALL
	INDEX
)

var precedenceTable = map[token.TokenType]uint8{
	token.TokenTypeAssign:          EQUALS,
	token.TokenTypeLess:            LESSGREATER,
	token.TokeTypeGreater:          LESSGREATER,
	token.TokenTypePlus:            SUM,
	token.TokenTypeMinus:           SUM,
	token.TokenTypeAsterisk:        PRODUCT,
	token.TokenTypeSlash:           PRODUCT,
	token.TokenTypeDot:             DOT,
	token.TokenTypeLeftParenthesis: CALL,
	token.TokenTypeLeftBracket:     INDEX,
}

func (p *Parser) currentPrecedence() uint8 {
	precedence, ok := precedenceTable[p.currentToken.TokenType]
	if ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() uint8 {
	precedence, ok := precedenceTable[p.peekToken.TokenType]
	if ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) parseExpression(precedence uint8) (ast.Expression, error) {
	prefixFn, ok := p.prefixParseFns[p.currentToken.TokenType]
	if !ok {
		return nil, fmt.Errorf("can't found prefix parse function for %s", p.currentToken.TokenType)
	}
	left, err := prefixFn()
	if err != nil {
		return nil, err
	}

	for !p.peekTokenIs(token.TokenTypeSemicolon) && precedence < p.peekPrecedence() {
		infixFn, ok := p.infixParseFns[p.peekToken.TokenType]
		if !ok {
			break
		}
		p.nextToken()

		left, err = infixFn(left)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, error) {
	block := &ast.BlockStatement{Token: p.currentToken, Statements: []ast.Statement{}}
	p.nextToken()

	for !p.currentTokenIs(token.TokenTypeRightBrace) {
		statement, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		block.Statements = append(block.Statements, statement)
		p.nextToken()
	}

	return block, nil
}

func (p *Parser) currentTokenIs(tokenType token.TokenType) bool {
	return p.currentToken.TokenType == tokenType
}
func (p *Parser) peekTokenIs(tokenType token.TokenType) bool {
	return p.peekToken.TokenType == tokenType
}
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) expectPeek(tokenType token.TokenType) bool {
	if p.peekTokenIs(tokenType) {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) parseIntegerLiteral() (ast.Expression, error) {
	num, err := strconv.Atoi(p.currentToken.Literal)
	if err != nil {
		return nil, err
	}
	exp := &ast.IntegerLiteral{
		Token: p.currentToken,
		Value: int32(num),
	}
	return exp, nil
}
func (p *Parser) parseKeywordConstantLiteral() (ast.Expression, error) {
	switch p.currentToken.TokenType {
	case token.TokenTypeTrue:
		return &ast.KeywordConstantLiteral{Token: p.currentToken, Value: "true"}, nil
	case token.TokenTypeFalse:
		return &ast.KeywordConstantLiteral{Token: p.currentToken, Value: "false"}, nil
	case token.TokenTypeNull:
		return &ast.KeywordConstantLiteral{Token: p.currentToken, Value: "null"}, nil
	case token.TokenTypeThis:
		return &ast.KeywordConstantLiteral{Token: p.currentToken, Value: "this"}, nil
	default:
		return nil, fmt.Errorf("expected current token to be this got %s", p.currentToken.TokenType)
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	op := p.currentToken.Literal
	infix := &ast.InfixExpression{
		Token:    p.currentToken,
		Left:     left,
		Operator: op,
	}
	precedence := p.currentPrecedence()
	p.nextToken()
	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}
	infix.Right = right

	return infix, nil
}
func (p *Parser) parseParenthesisExpression() (ast.Expression, error) {
	p.nextToken()
	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if !p.expectPeek(token.TokenTypeRightParenthesis) {
		return nil, fmt.Errorf("expected peek token to be right parenthesis but got %s", p.peekToken.TokenType)
	}
	return exp, nil
}

func (p *Parser) parseObjectCall(exp ast.Expression) (ast.Expression, error) {
	callee, ok := exp.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier, got %s", exp.String())
	}
	subroutineCall := &ast.SubroutineCall{
		Token:      callee.Token,
		CalleeName: callee,
	}
	p.nextToken()

	subroutineName, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}
	subroutineCall.SubroutineName = subroutineName.(*ast.Identifier)
	if !p.expectPeek(token.TokenTypeLeftParenthesis) {
		return nil, fmt.Errorf("expected peek token to be left parenthesis, got %s", p.peekToken.TokenType)
	}

	args, err := p.parseExpressions()
	if err != nil {
		return nil, err
	}
	subroutineCall.Arguments = args

	return subroutineCall, nil
}

func (p *Parser) parseCallExpression(exp ast.Expression) (ast.Expression, error) {
	subroutineName, ok := exp.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier, got %s", exp.String())
	}
	subroutineCall := &ast.SubroutineCall{
		Token:          subroutineName.Token,
		SubroutineName: subroutineName,
	}
	args, err := p.parseExpressions()
	if err != nil {
		return nil, err
	}
	subroutineCall.Arguments = args

	return subroutineCall, nil
}

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	prefixExpression := &ast.PrefixExpression{Token: p.currentToken}
	prefixExpression.Operator = p.currentToken.Literal
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	prefixExpression.Left = exp
	return prefixExpression, nil
}

func (p *Parser) parseIndexExpression(left ast.Expression) (ast.Expression, error) {
	indexExpression := &ast.IndexExpression{Token: p.currentToken, Left: left}
	p.nextToken()
	index, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}
	indexExpression.Index = index

	if !p.expectPeek(token.TokenTypeRightBracket) {
		return nil, fmt.Errorf("expected peek token to be right bracket got %s", p.peekToken.TokenType)
	}

	return indexExpression, nil
}

func (p *Parser) parseStringLiteral() (ast.Expression, error) {
	return &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}, nil
}
