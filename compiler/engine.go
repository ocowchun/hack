package compiler

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Engine struct {
	tokenizer *Tokenizer
}

func NewEngine(reader io.Reader) *Engine {
	return &Engine{
		tokenizer: NewTokenizer(reader),
	}
}

type CompileError struct {
	line    string
	token   Token
	message string
	err     error
}

func (e CompileError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("%s line: %s, token: %s", e.message, e.line, e.token)
	}
	return fmt.Sprintf("%s line: %s, token: %s", e.err, e.line, e.token)
}

func NewCompileError(err error, line string, token Token, message string) CompileError {
	return CompileError{
		line:    line,
		token:   token,
		message: message,
		err:     err,
	}
}

func (engine *Engine) CompileClass() (Class, error) {
	// TODO
	//var err error
	class := Class{
		varDec:        make([]ClassVarDec, 0),
		subroutineDec: make([]SubroutineDec, 0),
	}
	token, err := engine.tokenizer.Next()
	if err == io.EOF {
		return class, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	} else if err != nil {
		return class, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	} else if token.Type() != KeywordTokenType || token.Content() != "class" {
		msg := fmt.Sprintf("class must start with `class` but got %s", token.Content())
		return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)

	}

	token, err = engine.nextToken()
	if err == io.EOF {
		msg := fmt.Sprintf("class must start with has a name")
		return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	} else if err != nil {
		return class, err
	} else if token.Type() != IdentifierTokenType {
		return class, fmt.Errorf("class must start with has a name but got token type %s with content %s", token.Type(), token.Content())
	}
	className, err := BuildClassName(token)
	if err != nil {
		return class, err
	}
	class.name = className

	token, err = engine.nextToken()
	if err == io.EOF {
		return class, fmt.Errorf("expect a `{` but got nothing")
	} else if err != nil {
		return class, err
	} else if token.Type() != SymbolTokenType || token.Content() != "{" {
		return class, fmt.Errorf("expecte { after class name but got token type %s with content %s", token.Type(), token.Content())
	}

	for {
		token, err = engine.tokenizer.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return class, err
		}

		switch token.Type() {
		case UnknownTokenType:
			return class, fmt.Errorf("token type is unknow, content: %s", token.content)
		case KeywordTokenType:
			if token.Content() == "static" || token.Content() == "field" {
				// classVarDec -> start with static, or field
				// should not have classVarDec after subroutineDec is not empty
				if len(class.subroutineDec) > 0 {
					return class, fmt.Errorf("classVarDec must declare before subroutineDec")
				}

				classVarDec, err := engine.CompileClassVarDec()
				if err != nil {
					return class, err
				}
				class.varDec = append(class.varDec, classVarDec)
			} else if token.Content() == "constructor" || token.content == "function" || token.content == "method" {
				// subroutineDec -> start with constructor,function,or method
				subroutineDec, err := engine.CompileSubroutineDec()
				if err != nil {
					return class, err
				}
				class.subroutineDec = append(class.subroutineDec, subroutineDec)
			} else {
				msg := fmt.Sprintf("expected classVarDec or subroutineDec, but got symbol with content: %s", token.content)
				return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
			}

		case SymbolTokenType:
			if token.Content() == "}" {
				break
			}

			msg := fmt.Sprintf("expected '}', but got symbol with content: %s", token.content)
			return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		case IntegerConstantTokenType:
			msg := fmt.Sprintf("expected classVarDec or subroutineDec, but got integerConstant with content: %s", token.content)
			return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		case StringConstantTokenType:
			msg := fmt.Sprintf("expected classVarDec or subroutineDec, but got stringConstant with content: %s", token.content)
			return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		case IdentifierTokenType:
			msg := fmt.Sprintf("expected classVarDec or subroutineDec, but got identifier with content: %s", token.content)
			return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		default:
			msg := fmt.Sprintf("expected classVarDec or subroutineDec, but got content: %s", token.content)
			return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		}
	}

	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		return class, nil
	} else if err != nil {
		return class, err
	}

	msg := fmt.Sprintf("unexpected token %s after class end", token.Content())
	return class, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
}

func (engine *Engine) CompileClassVarDec() (ClassVarDec, error) {
	classVarDec := ClassVarDec{}
	token, err := engine.tokenizer.Current()
	if err != nil {
		return classVarDec, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}
	if token.Type() != KeywordTokenType {
		msg := fmt.Sprintf("expect keyword but got %s when parsing ClassVarDec", token.Type())
		return classVarDec, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	}
	// (`static`|`field`) type varName (`,` varName)* `;`
	switch token.Content() {
	case "static":
		classVarDec.scope = StaticClassVarScope
	case "field":
		classVarDec.scope = FieldClassVarScope
	default:
		msg := fmt.Sprintf("expect a `static` or a `field` but got %s when parsing ClassVarDec", token.Content())
		return classVarDec, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	}

	token, err = engine.tokenizer.Next()
	if err != nil {
		return classVarDec, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}
	typee, err := compileType(token)
	if err != nil {
		return classVarDec, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}
	classVarDec.typee = typee

	token, err = engine.tokenizer.Next()
	if err != nil {
		return classVarDec, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}
	varName, err := BuildVarName(token)
	if err != nil {
		return classVarDec, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}
	classVarDec.varNames = []VarName{varName}

	for {
		token, err = engine.tokenizer.Next()
		if err != nil {
			return classVarDec, err
		}
		if token.Type() == SymbolTokenType && token.Content() == ";" {
			break
		}
		if token.Type() == SymbolTokenType && token.Content() == "," {
			token, err = engine.tokenizer.Next()
			if err != nil {
				return classVarDec, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
			}
		} else {
			msg := fmt.Sprintf("expect a `;` or a `,` but got %s when parsing ClassVarDec", token.Content())
			return classVarDec, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		}

		varName, err = BuildVarName(token)
		if err != nil {
			return classVarDec, err
		}
		classVarDec.varNames = append(classVarDec.varNames, varName)
	}

	return classVarDec, nil
}

func (engine *Engine) CompileSubroutineDec() (SubroutineDec, error) {
	subroutineDec := SubroutineDec{}
	token, err := engine.tokenizer.Current()
	if err != nil {
		return subroutineDec, err
	}

	subroutineType := MethodSubroutineType
	switch token.Content() {
	case "constructor":
		subroutineType = ConstructorSubroutineType
	case "function":
		subroutineType = FunctionSubroutineType
	case "method":
		subroutineType = MethodSubroutineType
	default:
		return subroutineDec, fmt.Errorf("expect a subroutineType but got %s", token.Content())
	}
	subroutineDec.subroutineType = subroutineType

	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		return subroutineDec, fmt.Errorf("expect a type but got nothing")
	} else if err != nil {
		return subroutineDec, err
	}
	returnType := ReturnType{isVoid: false}
	if token.Type() == KeywordTokenType {
		switch token.Content() {
		case "void":
			returnType.isVoid = true
		case "int":
			returnType.typee = Type{primitiveClassName: "int"}
		case "char":
			returnType.typee = Type{primitiveClassName: "char"}
		case "boolean":
			returnType.typee = Type{primitiveClassName: "boolean"}
		default:
			return subroutineDec, fmt.Errorf("expect a return type but got %s", token.Content())
		}
	} else if token.Type() == IdentifierTokenType {
		className, err := BuildClassName(token)
		if err != nil {
			return subroutineDec, err
		}
		returnType.typee = Type{className: className}
	} else {
		return subroutineDec, fmt.Errorf("expect a return type but got %s", token.Content())
	}
	subroutineDec.returnType = returnType

	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		return subroutineDec, fmt.Errorf("expect an identifier but got nothing")
	} else if err != nil {
		return subroutineDec, err
	} else if token.Type() != IdentifierTokenType {
		return subroutineDec, fmt.Errorf("expect an identifier but got %s with content: %s", token.Type(), token.Content())
	}
	subroutineName, err := BuildSubroutineName(token)
	if err != nil {
		return subroutineDec, err
	}
	subroutineDec.name = subroutineName

	//token, err = engine.tokenizer.Next()
	err = engine.nextAndCheck(Token{tokenType: SymbolTokenType, content: "("})
	if err != nil {
		return subroutineDec, err
	}
	//if err == io.EOF {
	//	return subroutineDec, fmt.Errorf("expect a `(` but got nothing")
	//} else if err != nil {
	//	return subroutineDec, err
	//} else if token.Type() != SymbolTokenType || token.Content() != "(" {
	//	return subroutineDec, fmt.Errorf("expect a `(` but got %s with content: %s", token.Type(), token.Content())
	//}

	_, err = engine.tokenizer.Next()
	if err != nil {
		return subroutineDec, err
	}

	parameterList, err := engine.CompileParameterList()
	if err != nil {
		return subroutineDec, err
	}
	subroutineDec.parameters = parameterList

	token, err = engine.tokenizer.Current()
	if err != nil {
		return subroutineDec, err
	} else if token.Type() != SymbolTokenType || token.Content() != ")" {
		return subroutineDec, fmt.Errorf("expect a `)` but got %s with content: %s", token.Type(), token.Content())
	}

	_, err = engine.tokenizer.Next()
	if err != nil {
		return subroutineDec, err
	}
	body, err := engine.CompileSubroutineBody()
	if err != nil {
		return subroutineDec, err
	}
	subroutineDec.body = body

	return subroutineDec, nil
}

func (engine *Engine) CompileSubroutineBody() (SubroutineBody, error) {
	subroutineBody := SubroutineBody{
		varDecs: make([]*VarDec, 0),
	}
	// {varDecs* statements }

	err := engine.check(Token{tokenType: SymbolTokenType, content: "{"})
	if err != nil {
		return subroutineBody, err
	}
	_, err = engine.nextToken()
	if err != nil {
		return subroutineBody, err
	}
	for {
		token, err := engine.currentToken()
		if err == io.EOF {
			return subroutineBody, fmt.Errorf("reach EOF when parsing SubroutineBody")
		} else if err != nil {
			return subroutineBody, err
		}

		// handle close case first!
		if token.Type() == SymbolTokenType && token.content == "}" {
			break
		}

		if token.Type() == KeywordTokenType && token.content == "var" {
			// varDecs case
			// TODO: should not have varDecs case after statements is not empty
			varDec, err := engine.CompileVarDec()
			if err != nil {
				return subroutineBody, err
			}

			subroutineBody.varDecs = append(subroutineBody.varDecs, &varDec)
		} else {
			// statements case
			statements, err := engine.CompileStatements()
			if err != nil {
				return subroutineBody, err
			}
			subroutineBody.statements = statements
			break
		}
	}

	return subroutineBody, nil
}

// CompileStatements compile statements}
// if not tracking `}`, how to decide a statements parsing is finished?
// When parsing element* like (statement* or varName*), I decide to let compileXXX return the result once it meet a token it doesn't know how to deal with
func (engine *Engine) CompileStatements() (Statements, error) {
	statements := Statements{}
	for {
		token, err := engine.tokenizer.Current()
		if err != nil {
			if err == io.EOF {
				return statements, nil
			}
			return statements, err
		}

		if token.Type() != KeywordTokenType {
			return statements, nil
		}
		shouldNext := true

		switch token.Content() {
		case "let":
			letStatement, err := engine.CompileLetStatement()
			if err != nil {
				return statements, err
			}
			statements.statements = append(statements.statements, letStatement)

		case "if":
			ifStatement, err := engine.CompileIfStatement()
			if err != nil {
				return statements, err
			}
			statements.statements = append(statements.statements, ifStatement)
			shouldNext = false
		case "while":
			whileStatement, err := engine.CompileWhileStatement()
			if err != nil {
				return statements, err
			}
			statements.statements = append(statements.statements, whileStatement)
		case "do":
			doStatement, err := engine.CompileDoStatement()

			if err != nil {
				return statements, err
			}
			statements.statements = append(statements.statements, doStatement)
		case "return":
			returnStatement, err := engine.CompileReturnStatement()

			if err != nil {
				return statements, err
			}
			statements.statements = append(statements.statements, returnStatement)
		default:
			return statements, nil
		}
		if shouldNext {
			_, err = engine.tokenizer.Next()
			if err != nil {
				if err == io.EOF {
					return statements, nil
				}
				return statements, err
			}
		}
	}
}
func (engine *Engine) CompileWhileStatement() (WhileStatement, error) {
	statement := WhileStatement{}
	// `while` `(` expression `)` `{` statements `}`
	err := engine.check(Token{tokenType: KeywordTokenType, content: "while"})
	if err != nil {
		return statement, err
	}
	err = engine.nextAndCheck(Token{tokenType: SymbolTokenType, content: "("})
	if err != nil {
		return statement, err
	}

	_, err = engine.nextToken()
	if err != nil {
		return statement, err
	}
	expression, err := engine.CompileExpression()
	if err != nil {
		return statement, err
	}
	statement.expression = &expression

	err = engine.check(Token{tokenType: SymbolTokenType, content: ")"})
	err = engine.nextAndCheck(Token{tokenType: SymbolTokenType, content: "{"})

	_, err = engine.nextToken()
	if err != nil {
		return statement, err
	}
	statements, err := engine.CompileStatements()
	if err != nil {
		return statement, err
	}
	statement.statements = statements

	err = engine.check(Token{tokenType: SymbolTokenType, content: "}"})

	return statement, err
}

func (engine *Engine) currentToken() (Token, error) {
	return engine.tokenizer.Current()
}

func (engine *Engine) nextToken() (Token, error) {
	return engine.tokenizer.Next()
}

// check check token equals to expectedToken or not, return nil if matched
func (engine *Engine) check(expectedToken Token) error {
	token, err := engine.tokenizer.Current()
	if err != nil {
		return err
	}
	if token.Type() != expectedToken.Type() {
		return fmt.Errorf("expected token.type to be %s but got %s", expectedToken.Type(), token.Type())
	}
	if token.Content() != expectedToken.Content() {
		return fmt.Errorf("expected token.content to be %s but got %s", expectedToken.Content(), token.Content())
	}
	return nil
}

// nextAndCheck run engine.tokenizer.Next and check token
func (engine *Engine) nextAndCheck(expectedToken Token) error {
	_, err := engine.tokenizer.Next()
	if err != nil {
		return err
	}
	return engine.check(expectedToken)
}

func (engine *Engine) CompileLetStatement() (LetStatement, error) {
	statement := LetStatement{}
	token, err := engine.tokenizer.Current()
	if err != nil {
		return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}
	// let varName ([expression])? = expression;

	if token.Type() != KeywordTokenType || token.Content() != "let" {
		msg := fmt.Sprintf("expect `let` but got %s when parsing LetStatement", token.Content())
		return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	}

	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		msg := fmt.Sprintf("reach EOF when parsing LetStatement")
		return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	} else if err != nil {
		return statement, err
	}

	varName, err := BuildVarName(token)
	if err != nil {
		return statement, err
	}
	statement.varName = varName

	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		msg := fmt.Sprintf("reach EOF when parsing LetStatement")
		return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	} else if err != nil {
		return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}

	// compile [expression]
	if token.Type() == SymbolTokenType && token.Content() == "[" {
		token, err = engine.tokenizer.Next()
		if err == io.EOF {
			msg := fmt.Sprintf("reach EOF when parsing LetStatement")
			return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		} else if err != nil {
			return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}

		expression, err := engine.CompileExpression()
		if err != nil {
			return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
		statement.varNameExpression = &expression

		token, err = engine.tokenizer.Current()
		if err == io.EOF {
			msg := fmt.Sprintf("reach EOF when parsing LetStatement")
			return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		} else if err != nil {
			return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
		if token.Type() != SymbolTokenType || token.Content() != "]" {
			msg := fmt.Sprintf("expect `]` but got %s when parsing LetStatement", token.Content())
			return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		}

		_, err = engine.tokenizer.Next()
		if err == io.EOF {
			msg := fmt.Sprintf("reach EOF when parsing LetStatement")
			return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
		} else if err != nil {
			return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
	}

	err = engine.check(Token{tokenType: SymbolTokenType, content: "="})
	if err != nil {
		return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}

	// compile expression
	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		msg := fmt.Sprintf("reach EOF when parsing LetStatement")
		return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	} else if err != nil {
		return statement, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
	}

	expression, err := engine.CompileExpression()
	if err != nil {
		return statement, err
	}
	statement.expression = &expression
	if err != nil {
		return statement, err
	}

	// check `;`
	token, err = engine.tokenizer.Current()
	if err == io.EOF {
		msg := fmt.Sprintf("reach EOF when parsing LetStatement")
		return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	} else if err != nil {
		return statement, err
	}
	if token.Type() != SymbolTokenType || token.Content() != ";" {
		msg := fmt.Sprintf("expect `;` but got %s when parsing LetStatement", token.Content())
		return statement, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)
	}

	return statement, nil
}
func (engine *Engine) CompileReturnStatement() (ReturnStatement, error) {
	returnStatement := ReturnStatement{}
	// return expression? `;`

	token, err := engine.tokenizer.Current()
	if err != nil {
		return returnStatement, err
	}

	if token.Type() != KeywordTokenType || token.Content() != "return" {
		return returnStatement, fmt.Errorf("expect `return` got %s when parsing ReturnStatement", token.Content())
	}

	token, err = engine.tokenizer.Next()
	if err != nil {
		return returnStatement, err
	}

	if token.Type() != SymbolTokenType || token.Content() != ";" {
		expression, err := engine.CompileExpression()
		if err != nil {
			return returnStatement, err
		}
		returnStatement.expression = &expression
	}

	err = engine.check(Token{tokenType: SymbolTokenType, content: ";"})
	return returnStatement, err
}

func (engine *Engine) CompileDoStatement() (DoStatement, error) {
	doStatement := DoStatement{}
	// `do` subroutineCall;
	token, err := engine.tokenizer.Current()
	if err != nil {
		return doStatement, err
	}

	if token.Type() != KeywordTokenType || token.Content() != "do" {
		return doStatement, fmt.Errorf("expect `do` got %s when parsing DoStatement", token.Content())
	}

	token, err = engine.tokenizer.Next()
	if err != nil {
		return doStatement, err
	}
	_, err = engine.tokenizer.Next()
	if err != nil {
		return doStatement, err
	}
	subroutineCall, err := engine.CompileSubroutineCall(token)
	if err != nil {
		return doStatement, err
	}
	doStatement.subroutineCall = subroutineCall

	token, err = engine.tokenizer.Next()
	if err != nil {
		return doStatement, err
	}

	if token.Type() != SymbolTokenType || token.Content() != ";" {
		return doStatement, fmt.Errorf("expect `;` got %s when parsing DoStatement", token.Content())
	}
	return doStatement, nil
}

func (engine *Engine) CompileIfStatement() (IfStatement, error) {
	ifStatement := IfStatement{}
	// `if` `(` expression `)` `{` statements `}` (`else` `{` statements `}`)?

	token, err := engine.tokenizer.Current()
	if err != nil {
		return ifStatement, err
	}

	if token.Type() != KeywordTokenType || token.Content() != "if" {
		return ifStatement, fmt.Errorf("expect `if` got %s when parsing IfStatement", token.Content())
	}

	token, err = engine.tokenizer.Next()
	if err != nil {
		return ifStatement, err
	}
	if token.Type() != SymbolTokenType || token.Content() != "(" {
		return ifStatement, fmt.Errorf("expect `(` got %s when parsing IfStatement", token.Content())
	}

	_, err = engine.tokenizer.Next()
	if err != nil {
		return ifStatement, err
	}
	expression, err := engine.CompileExpression()
	if err != nil {
		return ifStatement, err
	}
	ifStatement.expression = &expression

	token, err = engine.tokenizer.Current()
	if err != nil {
		return ifStatement, err
	}
	if token.Type() != SymbolTokenType || token.Content() != ")" {
		return ifStatement, fmt.Errorf("expect `)` got %s when parsing IfStatement", token.Content())
	}

	token, err = engine.tokenizer.Next()
	if err != nil {
		return ifStatement, err
	}
	if token.Type() != SymbolTokenType || token.Content() != "{" {
		return ifStatement, fmt.Errorf("expect `{` got %s when parsing IfStatement", token.Content())
	}

	_, err = engine.tokenizer.Next()
	if err != nil {
		return ifStatement, err
	}
	trueStatements, err := engine.CompileStatements()
	if err != nil {
		return ifStatement, err
	}
	ifStatement.trueStatements = trueStatements

	token, err = engine.tokenizer.Current()
	if err != nil {
		return ifStatement, err
	}
	if token.Type() != SymbolTokenType || token.Content() != "}" {
		return ifStatement, fmt.Errorf("expect `}` got %s when parsing IfStatement", token.Content())
	}

	token, err = engine.tokenizer.Next()
	if err != nil {
		if err == io.EOF {
			return ifStatement, nil
		}
		return ifStatement, err
	}

	if token.Type() == KeywordTokenType && token.Content() == "else" {
		ifStatement.hasElse = true
		token, err = engine.tokenizer.Next()
		if err != nil {
			return ifStatement, err
		}
		if token.Type() != SymbolTokenType || token.Content() != "{" {
			return ifStatement, fmt.Errorf("expect `{` got %s when parsing IfStatement", token.Content())
		}

		_, err = engine.nextToken()
		if err != nil {
			return ifStatement, err
		}
		falseStatement, err := engine.CompileStatements()
		if err != nil {
			return ifStatement, err
		}
		ifStatement.falseStatements = falseStatement

		err = engine.check(Token{tokenType: SymbolTokenType, content: "}"})
		if err != nil {
			return ifStatement, err
		}
		_, err = engine.nextToken()
		if err != nil && err != io.EOF {
			return ifStatement, err
		}
	}

	return ifStatement, nil
}

func (engine *Engine) CompileExpressionList() (ExpressionList, error) {
	// (expression,(`,` expression)*)?
	expressionList := ExpressionList{expressions: make([]Expression, 0)}
	for {
		token, err := engine.tokenizer.Current()

		if err != nil {
			if err == io.EOF {
				return expressionList, nil
			}
			return expressionList, err
		}

		if len(expressionList.expressions) > 0 {
			if token.Type() == SymbolTokenType && token.Content() == "," {
				token, err = engine.tokenizer.Next()
				if err != nil {
					return expressionList, err
				}
			} else {
				// reach expressionList end
				return expressionList, nil
			}
		}

		// check it's valid to parse as expressionList, otherwise treat it as expressionList is completed
		if token.Type() == IntegerConstantTokenType {

		} else if token.Type() == StringConstantTokenType {

		} else if token.Type() == IdentifierTokenType {

		} else if token.Type() == SymbolTokenType {
			switch token.Content() {
			case "-":
			case "~":
			case "(":
			case ",":
				if len(expressionList.expressions) > 0 {
					_, err = engine.tokenizer.Next()
					if err != nil {
						return expressionList, nil
					}
				} else {
					return expressionList, nil
				}
			default:
				return expressionList, nil
			}
		} else if token.Type() == KeywordTokenType {
			_, err = BuildKeywordConstant(token)
			if err != nil {
				return expressionList, nil
			}
		} else {
			return expressionList, nil
		}

		expression, err := engine.CompileExpression()
		if err != nil {
			return expressionList, err
		}
		expressionList.expressions = append(expressionList.expressions, expression)
	}
}

func (engine *Engine) CompileExpression() (Expression, error) {
	expression := Expression{}
	// term (op term)
	leftTerm, err := engine.CompileTerm()
	if err != nil {
		return expression, err
	}
	expression.leftTerm = &leftTerm

	token, err := engine.tokenizer.Current()
	if err == io.EOF {
		return expression, nil
	} else if err != nil {
		return expression, err
	}

	if isOp(token) {
		op, err := BuildOp(token)
		if err != nil {
			return expression, err
		}
		expression.op = op

		_, err = engine.tokenizer.Next()
		if err == io.EOF {
			return expression, fmt.Errorf("reach EOF when parsing Expression")
		} else if err != nil {
			return expression, err
		}

		rightTerm, err := engine.CompileTerm()
		if err != nil {
			return expression, err
		}
		expression.rightTerm = &rightTerm
	}

	return expression, nil
}

func isOp(token Token) bool {
	if token.Type() != SymbolTokenType {
		return false
	}
	switch token.Content() {
	case "+":
		return true

	case "-":
		return true

	case "*":
		return true

	case "/":
		return true

	case "&":
		return true

	case "|":
		return true

	case "<":
		return true

	case ">":
		return true

	case "=":
		return true
	}
	return false
}

func (engine *Engine) CompileTerm() (Term, error) {
	term := Term{}
	// integerConstant| stringConstant | keywordConstant | varName | varName'[' expression ']' |subroutineCall
	// | '(' expression')' |unaryOp term
	token, err := engine.tokenizer.Current()
	if err != nil {
		return term, err
	}
	shouldNext := true
	if token.Type() == IntegerConstantTokenType {
		term.termType = IntegerConstantTermType

		num, err := strconv.ParseInt(token.Content(), 10, 32)
		if err != nil {
			return term, err
		}
		term.integerConstant = int32(num)
	} else if token.Type() == StringConstantTokenType {
		term.termType = StringConstantTermType
		term.stringConstant = token.Content()
	} else if token.Type() == KeywordTokenType {
		term.termType = KeywordConstantTermType

		keyword, err := BuildKeywordConstant(token)
		if err != nil {
			return term, err
		}
		term.keywordConstant = keyword
	} else if token.Type() == IdentifierTokenType {
		// handle varName
		// handle subroutineCall
		// handle varName[expression]
		firstToken := token

		token, err = engine.tokenizer.Next()
		hasNextToken := true
		if err == io.EOF {
			hasNextToken = false
		} else if err != nil {
			return term, err
		}

		if hasNextToken && token.Type() == SymbolTokenType && token.Content() == "[" {
			// handle varName[expression]
			varName, err := BuildVarName(firstToken)
			if err != nil {
				return term, err
			}

			term.termType = VarNameExpressionTermType
			term.varName = varName

			_, err = engine.tokenizer.Next()
			if err != nil {
				return term, err
			}

			expression, err := engine.CompileExpression()
			if err != nil {
				return term, err
			}
			term.expression = &expression
		} else if hasNextToken && token.Type() == SymbolTokenType && (token.Content() == "(" || token.Content() == ".") {
			// handle subroutineCall
			subroutineCall, err := engine.CompileSubroutineCall(firstToken)
			if err != nil {
				return term, err
			}
			term.termType = SubroutineCallTermType
			term.subroutineCall = subroutineCall
		} else {
			varName, err := BuildVarName(firstToken)
			if err != nil {
				return term, err
			}
			// handle varName
			term.termType = VarNameTermType
			term.varName = varName
			// no need to Next, because it doesn't use currentToken!
			shouldNext = false
		}
	} else if token.Type() == SymbolTokenType && token.Content() == "(" {
		term.termType = ExpressionTermType
		_, err = engine.tokenizer.Next()
		if err != nil {
			return term, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
		// handle expression
		expression, err := engine.CompileExpression()
		if err != nil {
			return term, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
		term.expression = &expression

		err = engine.check(Token{tokenType: SymbolTokenType, content: ")"})
		if err != nil {
			msg := fmt.Sprintf("expect `)` but got %s when parsing Term", token.Content())
			return term, NewCompileError(nil, engine.tokenizer.CurrentLine(), token, msg)

		}
	} else if token.Type() == SymbolTokenType && (token.Content() == "-" || token.Content() == "~") {
		term.termType = UnaryOpTermTermType
		// handle unaryOp term
		unaryOp, err := BuildUnaryOp(token)
		if err != nil {
			return term, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
		term.unaryOp = unaryOp

		_, err = engine.tokenizer.Next()
		if err != nil {
			return term, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}

		term2, err := engine.CompileTerm()
		if err != nil {
			return term, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
		term.term = &term2
		shouldNext = false
	}

	if shouldNext {
		_, err = engine.tokenizer.Next()
		// no need to handle EOF here, for example: if we parse a line `42`, then EOF is still valid case from CompileTerm's POV
		// the caller can decide it's an error or not according to it's context
		if err != nil && err != io.EOF {
			return term, NewCompileError(err, engine.tokenizer.CurrentLine(), token, "")
		}
	}

	return term, nil
}

func (engine *Engine) CompileSubroutineCall(firstToken Token) (SubroutineCall, error) {
	subroutineCall := SubroutineCall{}
	token, err := engine.tokenizer.Current()
	if err == io.EOF {
		return subroutineCall, fmt.Errorf("reach EOF when parsing SubroutineCall")
	} else if err != nil {
		return subroutineCall, err
	}

	if token.Content() == "(" {
		// subroutineName(expressionList)
		subroutineName, err := BuildSubroutineName(firstToken)
		if err != nil {
			return subroutineCall, err
		}
		subroutineCall.subroutineName = subroutineName

		_, err = engine.tokenizer.Next()
		if err != nil {
			return subroutineCall, err
		}
		expressionList, err := engine.CompileExpressionList()
		if err != nil {
			return subroutineCall, err
		}
		subroutineCall.expressionList = expressionList

		token, err = engine.tokenizer.Current()
		if err != nil {
			return subroutineCall, err
		}
		if token.Type() != SymbolTokenType || token.Content() != ")" {
			return subroutineCall, fmt.Errorf("expect a `)` but got %s when parsing SubroutineCall", token.Content())
		}
	} else if token.Content() == "." {
		// (className|varName).subroutineName(expressionList)
		if unicode.IsUpper(rune(firstToken.Content()[0])) {
			className, err := BuildClassName(firstToken)

			if err != nil {
				return subroutineCall, err
			}
			subroutineCall.className = className
		} else {
			varName, err := BuildVarName(firstToken)
			if err != nil {
				return subroutineCall, err
			}
			subroutineCall.varName = varName
		}

		// handle subroutineName
		token, err = engine.tokenizer.Next()
		if err != nil {
			return subroutineCall, err
		}
		subroutineName, err := BuildSubroutineName(token)
		if err != nil {
			return subroutineCall, err
		}
		subroutineCall.subroutineName = subroutineName

		token, err = engine.tokenizer.Next()
		if err != nil {
			return subroutineCall, err
		}
		if token.Type() != SymbolTokenType || token.Content() != "(" {
			return subroutineCall, fmt.Errorf("expect a `(` but got %s when parsing SubroutineCall", token.Content())
		}

		_, err = engine.nextToken()
		if err != nil {
			return subroutineCall, err
		}
		expressionList, err := engine.CompileExpressionList()
		if err != nil {
			return subroutineCall, err
		}
		subroutineCall.expressionList = expressionList
	} else {
		return subroutineCall, fmt.Errorf("expect a `)` or `.` but got %s when parsing SubroutineCall", token.Content())
	}

	return subroutineCall, nil
}
func (engine *Engine) CompileVarDec() (VarDec, error) {
	// var type varName,varName;
	varDec := VarDec{}
	token, err := engine.tokenizer.Current()
	if err == io.EOF {
		return varDec, fmt.Errorf("reach EOF when parsing VarDec")
	} else if err != nil {
		return varDec, err
	} else if token.Type() != KeywordTokenType || token.Content() != "var" {
		return varDec, fmt.Errorf("expect `var` but got %s parsing VarDec", token.Content())
	}

	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		return varDec, fmt.Errorf("reach EOF when parsing VarDec")
	} else if err != nil {
		return varDec, err
	}
	typee, err := compileType(token)
	if err != nil {
		return varDec, err
	}
	varDec.typee = typee

	for {
		token, err = engine.tokenizer.Next()
		if err == io.EOF {
			return varDec, fmt.Errorf("reach EOF when parsing VarDec")
		} else if err != nil {
			return varDec, err
		}

		// close when reach ;
		if token.Type() == SymbolTokenType && token.content == ";" {
			if len(varDec.names) > 0 {
				break
			} else {
				return varDec, fmt.Errorf("can't found variable name when parsing VarDec")
			}
		}

		// must start with `,` if len(varDecs.names) > 0
		if len(varDec.names) > 0 {
			if token.Type() == SymbolTokenType && token.content == "," {
				token, err = engine.tokenizer.Next()
				if err == io.EOF {
					return varDec, fmt.Errorf("reach EOF when parsing VarDec")
				} else if err != nil {
					return varDec, err
				}
			} else {
				return varDec, fmt.Errorf("expect a `,` but got token %s when parsing VarDec", token.Content())
			}
		}

		varName, err := BuildVarName(token)
		if err != nil {
			return varDec, err
		}
		varDec.names = append(varDec.names, varName)
	}

	_, err = engine.nextToken()
	return varDec, err
}

// CompileParameterList compile type1 var1, type2 var2
func (engine *Engine) CompileParameterList() (ParameterList, error) {
	parameterList := ParameterList{
		parameters: make([]Parameter, 0),
	}

	for {
		token, err := engine.tokenizer.Current()
		if err == io.EOF {
			// return when current state is valid and reach token doesn't know how to deal with
			return parameterList, nil
		} else if err != nil {
			return parameterList, err
		}

		// should start with `,` if len(parameters) > 0
		if len(parameterList.parameters) > 0 {
			if token.Type() == SymbolTokenType && token.content == "," {
				token, err = engine.tokenizer.Next()
				if err == io.EOF {
					return parameterList, nil
				} else if err != nil {
					return parameterList, err
				}
			} else {
				// return when current state is valid and reach token doesn't know how to deal with
				return parameterList, nil
			}
		} else {
			_, err = compileType(token)
			if err != nil {
				return parameterList, nil
			}
		}

		parameter, err := engine.compileParameter()
		if err != nil {
			return parameterList, err
		}
		parameterList.parameters = append(parameterList.parameters, parameter)

		// handle error by begin of for loop
		_, _ = engine.tokenizer.Next()
	}
}

func (engine *Engine) compileParameter() (Parameter, error) {
	parameter := Parameter{}
	token, err := engine.tokenizer.Current()
	if err != nil {
		return parameter, err
	}
	// i.e., int a
	typee, err := compileType(token)
	if err != nil {
		return parameter, err
	}
	parameter.typee = typee

	token, err = engine.tokenizer.Next()
	if err == io.EOF {
		return parameter, fmt.Errorf("reach EOF when parsing parameterList")
	} else if err != nil {
		return parameter, err
	}

	varName, err := BuildVarName(token)
	if err != nil {
		return parameter, err
	}
	parameter.name = varName

	return parameter, nil
}

func compileType(token Token) (Type, error) {
	res := Type{}
	if token.Type() == KeywordTokenType {
		switch token.Content() {
		case "int":
			res.primitiveClassName = "int"
		case "char":
			res.primitiveClassName = "char"
		case "boolean":
			res.primitiveClassName = "boolean"
		default:
			return res, fmt.Errorf("expect a type but got %s", token.Content())
		}
	} else if token.Type() == IdentifierTokenType {
		className, err := BuildClassName(token)
		if err != nil {
			return res, err
		}
		res.className = className
	} else {
		return res, fmt.Errorf("expect a type but got %s", token.Content())
	}
	return res, nil
}

// 'class' className '{' classVarDec* subroutineDec* '}'
type Class struct {
	name          ClassName
	varDec        []ClassVarDec
	subroutineDec []SubroutineDec
}

func (c Class) Name() ClassName {
	return c.name
}

func (c Class) VarDecs() []ClassVarDec {
	return c.varDec
}

func (c Class) SubroutineDecs() []SubroutineDec {
	return c.subroutineDec
}

// 'int'|'char'|'boolean'|className
type Type struct {
	primitiveClassName string
	className          ClassName
}

func (t Type) PrimitiveClassName() string {
	return t.primitiveClassName
}

func (t Type) String() string {
	if len(t.primitiveClassName) > 0 {
		return t.primitiveClassName
	}
	return t.className.identifier.content
}

func (t Type) Name() string {
	if len(t.primitiveClassName) > 0 {
		return t.primitiveClassName
	}
	return t.className.identifier.content
}

// ('static'|'field') type varName (',' varName)* ';'
type ClassVarScope uint8

const (
	StaticClassVarScope ClassVarScope = iota
	FieldClassVarScope
)

func (s ClassVarScope) String() string {
	switch s {
	case StaticClassVarScope:
		return "static"
	case FieldClassVarScope:
		return "field"
	default:
		return "unknown"
	}
}

type ClassVarDec struct {
	typee    Type
	scope    ClassVarScope
	varNames []VarName
}

func (d ClassVarDec) Type() Type {
	return d.typee
}

func (d ClassVarDec) Scope() ClassVarScope {
	return d.scope
}

func (d ClassVarDec) VarNames() []VarName {
	return d.varNames
}

type ClassName struct {
	identifier Identifier
}

func (n ClassName) Name() string {
	return n.identifier.Content()
}

func BuildClassName(token Token) (ClassName, error) {
	identifier, err := BuildIdentifier(token)
	if err != nil {
		return ClassName{}, err
	}
	// className must start with upper letter
	if !unicode.IsUpper(rune(identifier.Content()[0])) {
		return ClassName{}, fmt.Errorf("className must start with upper letter, but got %s", identifier.content)
	}

	return ClassName{identifier: identifier}, nil
}

type SubroutineType uint8

const (
	ConstructorSubroutineType SubroutineType = iota
	FunctionSubroutineType
	MethodSubroutineType
)

func (t SubroutineType) String() string {
	switch t {
	case ConstructorSubroutineType:
		return "constructor"
	case FunctionSubroutineType:
		return "function"
	case MethodSubroutineType:
		return "method"
	}
	return ""
}

type ReturnType struct {
	typee  Type
	isVoid bool
}

func (t ReturnType) String() string {
	if t.isVoid {
		return "void"
	}
	return t.typee.Name()
}

func (t ReturnType) IsVoid() bool {
	return t.isVoid
}

func (t ReturnType) Type() Type {
	return t.typee
}

// ('constructor'|'function'|'method') ('void'|type) subroutineName '('parameterList ')' subroutineBody
type SubroutineDec struct {
	subroutineType SubroutineType
	returnType     ReturnType
	name           SubroutineName
	parameters     ParameterList
	body           SubroutineBody
}

func (d SubroutineDec) SubroutineType() SubroutineType {
	return d.subroutineType
}

func (d SubroutineDec) ReturnType() ReturnType {
	return d.returnType
}

func (d SubroutineDec) Name() SubroutineName {
	return d.name
}
func (d SubroutineDec) Parameters() ParameterList {
	return d.parameters
}
func (d SubroutineDec) Body() SubroutineBody {
	return d.body
}

// ((type varName) (',' type varName)*)?
type ParameterList struct {
	parameters []Parameter
}

func (l ParameterList) Parameters() []Parameter {
	return l.parameters
}

type Parameter struct {
	typee Type
	name  VarName
}

func (p Parameter) Type() Type {
	return p.typee
}

func (p Parameter) Name() VarName {
	return p.name
}

// '{'varDecs* statements '}'
type SubroutineBody struct {
	varDecs    []*VarDec
	statements Statements
}

func (b SubroutineBody) VarDecs() []*VarDec {
	return b.varDecs
}

func (b SubroutineBody) Statements() Statements {
	return b.statements
}

type Statements struct {
	statements []Statement
}

func (s Statements) Statements() []Statement {
	return s.statements
}

type StatementType uint8

const (
	LetStatementType StatementType = iota
	IfStatementType
	WhileStatementType
	DoStatementType
	ReturnStatementType
)

func (s StatementType) String() string {
	switch s {
	case LetStatementType:
		return "Let"
	case IfStatementType:
		return "If"
	case WhileStatementType:
		return "While"
	case DoStatementType:
		return "Do"
	case ReturnStatementType:
		return "Return"
	}
	return ""
}

type Statement interface {
	StatementType() StatementType
}

// LetStatement 'let' varName ( '[' expression ']')? '=' expression ';'
type LetStatement struct {
	varName           VarName
	varNameExpression *Expression
	expression        *Expression
}

func (l LetStatement) VarName() VarName {
	return l.varName
}

func (l LetStatement) VarNameExpression() *Expression {
	return l.varNameExpression
}

func (l LetStatement) Expression() *Expression {
	return l.expression
}

func (l LetStatement) StatementType() StatementType {
	return LetStatementType
}

// IfStatement 'if' '(' expression ')' '{' statements '}' ( 'else' '{' statements '}' )?
type IfStatement struct {
	expression      *Expression
	trueStatements  Statements
	hasElse         bool
	falseStatements Statements
}

func (i IfStatement) Expression() *Expression {
	return i.expression
}
func (i IfStatement) TrueStatements() Statements {
	return i.trueStatements
}

func (i IfStatement) FalseStatements() Statements {
	return i.falseStatements
}
func (i IfStatement) HasElse() bool {
	return i.hasElse
}

func (i IfStatement) StatementType() StatementType {
	return IfStatementType
}

// WhileStatement 'while' '(' expression ')' '{'statements '}'
type WhileStatement struct {
	expression *Expression
	statements Statements
}

func (w WhileStatement) Expression() *Expression {
	return w.expression
}

func (w WhileStatement) Statements() Statements {
	return w.statements
}

func (w WhileStatement) StatementType() StatementType {
	return WhileStatementType
}

// DoStatement 'do' subroutineCall ';'
type DoStatement struct {
	subroutineCall SubroutineCall
}

func (d DoStatement) SubroutineCall() SubroutineCall {
	return d.subroutineCall
}

func (d DoStatement) StatementType() StatementType {
	return DoStatementType
}

// ReturnStatement 'return' expression? ';'
type ReturnStatement struct {
	expression *Expression
}

func (r ReturnStatement) Expression() *Expression {
	return r.expression
}

func (r ReturnStatement) HasExpression() bool {
	return r.expression != nil
}

func (r ReturnStatement) StatementType() StatementType {
	return ReturnStatementType
}

// Expression term (op term)*
type Expression struct {
	leftTerm  *Term
	op        Op
	rightTerm *Term
}

func (e Expression) LeftTerm() *Term {
	return e.leftTerm
}

func (e Expression) Op() Op {
	return e.op
}

func (e Expression) RightTerm() *Term {
	return e.rightTerm
}

func (e Expression) String() string {
	if e.rightTerm == nil {
		return e.leftTerm.String()
	}
	return fmt.Sprintf("%s %s %s", e.leftTerm, e.op, e.rightTerm)

}

func (e Expression) HasOpAndRightTerm() bool {
	return e.rightTerm != nil
}

type TermType uint8

const (
	IntegerConstantTermType TermType = iota
	StringConstantTermType
	KeywordConstantTermType
	VarNameTermType
	VarNameExpressionTermType
	SubroutineCallTermType
	ExpressionTermType
	UnaryOpTermTermType
)

func (t TermType) String() string {
	switch t {
	case IntegerConstantTermType:
		return "IntegerConstant"
	case StringConstantTermType:
		return "StringConstant"
	case KeywordConstantTermType:
		return "Keyword"
	case VarNameTermType:
		return "VarName"
	case VarNameExpressionTermType:
		return "VarNameExpression"
	case SubroutineCallTermType:
		return "SubroutineCall"
	case ExpressionTermType:
		return "Expression"
	case UnaryOpTermTermType:
		return "UnaryOp Term"
	}
	return ""
}

// integerConstant| stringConstant | keywordConstant | varName | varName'[' expression ']' |subroutineCall
// | '(' expression')' |unaryOp term
type Term struct {
	termType        TermType
	integerConstant int32
	stringConstant  string
	keywordConstant KeywordConstant
	varName         VarName
	expression      *Expression
	subroutineCall  SubroutineCall
	term            *Term
	unaryOp         UnaryOp
}

func (t Term) TermType() TermType {
	return t.termType
}

func (t Term) SubroutineCall() SubroutineCall {
	return t.subroutineCall
}
func (t Term) IntegerConstant() int32 {
	return t.integerConstant
}
func (t Term) StringConstant() string {
	return t.stringConstant
}
func (t Term) KeywordConstant() KeywordConstant {
	return t.keywordConstant
}
func (t Term) VarName() VarName {
	return t.varName
}
func (t Term) Expression() *Expression {
	return t.expression
}
func (t Term) Term() *Term {
	return t.term
}
func (t Term) UnaryOp() UnaryOp {
	return t.unaryOp
}

func (t Term) String() string {
	switch t.termType {
	case IntegerConstantTermType:
		return fmt.Sprintf("%d", t.integerConstant)
	case StringConstantTermType:
		return t.stringConstant
	case KeywordConstantTermType:
		return t.keywordConstant.String()
	case VarNameTermType:
		return t.varName.Name()
	case VarNameExpressionTermType:
		return t.expression.String()
	case SubroutineCallTermType:
		return t.subroutineCall.String()
	case ExpressionTermType:
		return t.expression.String()
	case UnaryOpTermTermType:
		return fmt.Sprintf("%s%s", t.unaryOp, t.term)
	}
	return ""
}

type SubroutineCallType uint8

const (
	SubroutineSubroutineCallType SubroutineCallType = iota
	ClassNameSubroutineCallType
	VarNameSubroutineCallType
)

// subroutineName '(' expressionList ')' | (className|varName) '.' subroutineName '(' expressionList ')'
type SubroutineCall struct {
	subroutineName SubroutineName
	expressionList ExpressionList
	className      ClassName
	varName        VarName
}

func (c SubroutineCall) SubroutineName() SubroutineName {
	return c.subroutineName
}
func (c SubroutineCall) ExpressionList() ExpressionList {
	return c.expressionList
}
func (c SubroutineCall) ClassName() ClassName {
	return c.className
}
func (c SubroutineCall) VarName() VarName {
	return c.varName
}

func (c SubroutineCall) String() string {
	if c.className.Name() != "" {
		return fmt.Sprintf("%s.%s(%s)", c.className.Name(), c.subroutineName.Name(), c.expressionList)
	} else if c.varName.Name() != "" {
		return fmt.Sprintf("%s.%s(%s)", c.varName.Name(), c.subroutineName.Name(), c.expressionList)
	} else {
		return fmt.Sprintf("%s(%s)", c.subroutineName.Name(), c.expressionList)
	}
}

// (expression ( ',' expression)* )?
type ExpressionList struct {
	expressions []Expression
}

func (l ExpressionList) Expressions() []Expression {
	return l.expressions
}

func (l ExpressionList) String() string {
	es := make([]string, 0)
	for _, expression := range l.expressions {
		es = append(es, expression.String())
	}
	return strings.Join(es, ",")
}

type Op uint8

const (
	PlusOp Op = iota
	MinusOp
	MultipleOp
	DivideOp
	AndOp
	OrOp
	GreaterOp
	LessOp
	EqualOp
)

func BuildOp(token Token) (Op, error) {
	res := PlusOp
	if token.Type() != SymbolTokenType {
		return res, fmt.Errorf("expect token.type to be Symbol but got %s", token.Type())
	}
	switch token.Content() {
	case "+":
		res = PlusOp
	case "-":
		res = MinusOp
	case "*":
		res = MultipleOp
	case "/":
		res = DivideOp
	case "&":
		res = AndOp
	case "|":
		res = OrOp
	case "<":
		res = LessOp
	case ">":
		res = GreaterOp
	case "=":
		res = EqualOp
	default:
		return res, fmt.Errorf("invalid symbol %s when parsing UnaryOp", token.Content())
	}
	return res, nil
}

func (o Op) String() string {
	switch o {
	case PlusOp:
		return "+"
	case MinusOp:
		return "-"
	case MultipleOp:
		return "*"
	case DivideOp:
		return "/"
	case AndOp:
		return "&"
	case OrOp:
		return "|"
	case GreaterOp:
		return ">"
	case LessOp:
		return "<"
	case EqualOp:
		return "="
	}

	return ""
}

type UnaryOp uint8

const (
	NegativeUnaryOp UnaryOp = iota
	TildeUnaryOp
)

func (o UnaryOp) String() string {
	switch o {
	case NegativeUnaryOp:
		return "-"
	case TildeUnaryOp:
		return "~"
	}
	return ""
}

func BuildUnaryOp(token Token) (UnaryOp, error) {
	res := NegativeUnaryOp
	if token.Type() != SymbolTokenType {
		return res, fmt.Errorf("expect token.type to be Symbol but got %s", token.Type())
	}
	switch token.Content() {
	case "-":
		res = NegativeUnaryOp
	case "~":
		res = TildeUnaryOp
	default:
		return res, fmt.Errorf("invalid symbol %s when parsing UnaryOp", token.Content())
	}
	return res, nil
}

type KeywordConstant uint8

const (
	TrueKeywordConstant KeywordConstant = iota
	FalseKeywordConstant
	NullKeywordConstant
	ThisKeywordConstant
)

func (k KeywordConstant) String() string {
	switch k {
	case TrueKeywordConstant:
		return "true"
	case FalseKeywordConstant:
		return "false"
	case NullKeywordConstant:
		return "null"
	case ThisKeywordConstant:
		return "this"

	}
	return ""
}

func BuildKeywordConstant(token Token) (KeywordConstant, error) {
	res := TrueKeywordConstant
	if token.Type() != KeywordTokenType {
		return res, fmt.Errorf("expect token.type to be Keyword, but got %s", token.Type())
	}
	switch token.Content() {
	case "true":
		res = TrueKeywordConstant

	case "false":
		res = FalseKeywordConstant

	case "null":
		res = NullKeywordConstant

	case "this":
		res = ThisKeywordConstant
	default:
		return res, fmt.Errorf("invalid keyword %s when parsing KeywordConstant", token.Content())
	}

	return res, nil
}

// 'var' type varName (',' type varName)* ';'
type VarDec struct {
	typee Type
	names []VarName
}

func (d VarDec) Type() Type {
	return d.typee
}

func (d VarDec) Names() []VarName {
	return d.names
}

type SubroutineName struct {
	identifier Identifier
}

func (s SubroutineName) Name() string {
	return s.identifier.Content()
}

func (s SubroutineName) String() string {
	return s.Name()
}

func BuildSubroutineName(token Token) (SubroutineName, error) {
	identifier, err := BuildIdentifier(token)
	if err != nil {
		return SubroutineName{}, err
	}
	// subroutineName must start with upper letter
	if !unicode.IsLower(rune(identifier.Content()[0])) {
		return SubroutineName{}, fmt.Errorf("subroutineName must start with lower letter, but got %s", identifier.content)
	}

	return SubroutineName{identifier: identifier}, nil
}

type VarName struct {
	identifier Identifier
}

func (v VarName) String() string {
	return v.Name()
}

func (v VarName) Name() string {
	return v.identifier.Content()
}

func BuildVarName(token Token) (VarName, error) {
	identifier, err := BuildIdentifier(token)
	if err != nil {
		return VarName{}, err
	}

	return VarName{identifier: identifier}, nil
}

type Identifier struct {
	content string
}

func (i Identifier) Content() string {
	return i.content
}

func BuildIdentifier(token Token) (Identifier, error) {
	if token.Type() != IdentifierTokenType {
		return Identifier{}, fmt.Errorf("expect token.type to be Identifier, but got %s", token.Type())
	}

	return Identifier{content: token.content}, nil
}
