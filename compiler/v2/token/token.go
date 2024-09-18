package token

import "fmt"

type TokenType uint8

const (
	TokenTypeIllegal TokenType = iota
	TokenTypeComma
	TokenTypeIntegerLiteral
	TokenTypeStringLiteral
	TokenTypeLeftParenthesis
	TokenTypeRightParenthesis
	TokenTypeLeftBrace
	TokenTypeRightBrace
	TokenTypeSemicolon
	TokenTypeAssign
	TokenTypeTilde
	TokenTypeLeftBracket
	TokenTypeRightBracket
	TokeTypeGreater
	TokenTypeLess
	TokenTypePlus
	TokenTypeMinus
	TokenTypeAsterisk
	TokenTypeSlash
	TokenTypeAmpersand
	TokenTypeVerticalBar
	TokenTypeDot
	TokenTypeClass
	TokenTypeConstructor
	TokenTypeFunction
	TokenTypeMethod
	TokenTypeField
	TokenTypeStatic
	TokenTypeVar
	TokenTypeInt
	TokenTypeChar
	TokenTypeBoolean
	TokenTypeVoid
	TokenTypeTrue
	TokenTypeFalse
	TokenTypeNull
	TokenTypeThis
	TokenTypeLet
	TokenTypeDo
	TokenTypeIf
	TokenTypeWhile
	TokenTypeElse
	TokenTypeReturn
	TokenTypeEOF

	TokenTypeIdentifier
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeIllegal:
		return "illegal"
	case TokenTypeComma:
		return "comma"
	case TokenTypeIntegerLiteral:
		return "integerLiteral"
	case TokenTypeStringLiteral:
		return "stringLiteral"
	case TokenTypeLeftParenthesis:
		return "leftParenthesis"
	case TokenTypeRightParenthesis:
		return "rightParenthesis"
	case TokenTypeLeftBrace:
		return "leftBrace"
	case TokenTypeRightBrace:
		return "rightBrace"
	case TokenTypeSemicolon:
		return "semicolon"
	case TokenTypeAssign:
		return "assign"
	case TokenTypeTilde:
		return "tilde"
	case TokenTypeLeftBracket:
		return "leftBracket"
	case TokenTypeRightBracket:
		return "bracket"
	case TokeTypeGreater:
		return "greater"
	case TokenTypeLess:
		return "less"
	case TokenTypePlus:
		return "plus"
	case TokenTypeMinus:
		return "minus"
	case TokenTypeAsterisk:
		return "asterisk"
	case TokenTypeSlash:
		return "slash"
	case TokenTypeAmpersand:
		return "ampersand"
	case TokenTypeVerticalBar:
		return "verticalBar"
	case TokenTypeDot:
		return "dot"
	case TokenTypeClass:
		return "class"
	case TokenTypeConstructor:
		return "constructor"
	case TokenTypeFunction:
		return "function"
	case TokenTypeMethod:
		return "method"
	case TokenTypeField:
		return "field"
	case TokenTypeStatic:
		return "static"
	case TokenTypeVar:
		return "var"
	case TokenTypeInt:
		return "int"
	case TokenTypeChar:
		return "char"
	case TokenTypeBoolean:
		return "boolean"
	case TokenTypeVoid:
		return "void"
	case TokenTypeTrue:
		return "true"
	case TokenTypeFalse:
		return "false"
	case TokenTypeNull:
		return "null"
	case TokenTypeThis:
		return "this"
	case TokenTypeLet:
		return "let"
	case TokenTypeDo:
		return "do"
	case TokenTypeIf:
		return "if"
	case TokenTypeWhile:
		return "while"
	case TokenTypeElse:
		return "else"
	case TokenTypeReturn:
		return "return"
	case TokenTypeEOF:
		return "EOF"
	case TokenTypeIdentifier:
		return "identifier"
	default:
		panic(fmt.Sprintf("unknown token type: %d", t))
	}
}

type Token struct {
	TokenType TokenType
	Literal   string
}

func (t Token) String() string {
	return t.Literal
}

var keywords = map[string]TokenType{
	"class":       TokenTypeClass,
	"constructor": TokenTypeConstructor,
	"function":    TokenTypeFunction,
	"method":      TokenTypeMethod,
	"field":       TokenTypeField,
	"static":      TokenTypeStatic,
	"var":         TokenTypeVar,
	"int":         TokenTypeInt,
	"char":        TokenTypeChar,
	"boolean":     TokenTypeBoolean,
	"void":        TokenTypeVoid,
	"true":        TokenTypeTrue,
	"false":       TokenTypeFalse,
	"null":        TokenTypeNull,
	"this":        TokenTypeThis,
	"let":         TokenTypeLet,
	"do":          TokenTypeDo,
	"if":          TokenTypeIf,
	"while":       TokenTypeWhile,
	"else":        TokenTypeElse,
	"return":      TokenTypeReturn,
}

func LookupIdentifier(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenTypeIdentifier
}

func Foo() {

}
