package ast

import (
	"bytes"
	"fmt"
	"hack/compiler/v2/token"
	"strings"
)

type AstNode interface {
	TokenLiteral() string
	String() string
}

type Expression interface {
	AstNode
	expressionNode()
}

type Statement interface {
	AstNode
	statementNode()
}
type Structure interface {
	AstNode
	structureNode()
}

type FieldScope uint8

const (
	FieldScopeStatic FieldScope = iota
	FieldScopeInstance
)

func (f FieldScope) String() string {
	switch f {
	case FieldScopeStatic:
		return "static"
	case FieldScopeInstance:
		return "field"
	default:
		panic("unknown field scope")
	}
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

type Field struct {
	Token       token.Token
	Scope       FieldScope
	Type        string
	Identifiers []*Identifier
}

func (f *Field) structureNode() {}
func (f *Field) TokenLiteral() string {
	return f.Token.Literal
}

func (f *Field) String() string {
	var out bytes.Buffer
	idents := make([]string, len(f.Identifiers))
	for i, identifier := range f.Identifiers {
		idents[i] = identifier.String()
	}

	out.WriteString(f.Scope.String())
	out.WriteString(" ")
	out.WriteString(f.Type)
	out.WriteString(" ")
	out.WriteString(strings.Join(idents, ", "))
	out.WriteString(";")
	return out.String()
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (b *BlockStatement) statementNode() {}
func (b *BlockStatement) TokenLiteral() string {
	return b.Token.Literal
}
func (b *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range b.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type SubroutineType uint8

const (
	SubroutineTypeConstructor SubroutineType = iota
	SubroutineTypeFunction
	SubroutineTypeMethod
)

func (s SubroutineType) String() string {
	switch s {
	case SubroutineTypeConstructor:
		return "constructor"
	case SubroutineTypeFunction:
		return "function"
	case SubroutineTypeMethod:
		return "method"
	default:
		panic(fmt.Sprintf("unknown subroutine type %d", s))
	}
}

type Parameter struct {
	Token token.Token
	Type  string
	Name  *Identifier
}

func (p *Parameter) structureNode() {}
func (p *Parameter) TokenLiteral() string {
	return p.Token.Literal
}
func (p *Parameter) String() string {
	var out bytes.Buffer

	out.WriteString(p.Type)
	out.WriteString(" ")
	out.WriteString(p.Name.String())
	return out.String()
}

type Subroutine struct {
	Token      token.Token
	Type       SubroutineType
	Name       *Identifier
	ReturnType string
	Parameters []*Parameter
	Body       *BlockStatement
}

func (s *Subroutine) structureNode() {}
func (s *Subroutine) TokenLiteral() string {
	return s.Token.Literal
}
func (s *Subroutine) String() string {
	var output bytes.Buffer
	params := make([]string, len(s.Parameters))
	for i, para := range s.Parameters {
		params[i] = para.String()
	}

	output.WriteString(s.Type.String())
	output.WriteString(" ")
	output.WriteString(s.ReturnType)
	output.WriteString(" ")
	output.WriteString(s.Name.String())
	output.WriteString("(")
	output.WriteString(strings.Join(params, ", "))
	output.WriteString("){")
	output.WriteString(s.Body.String())
	output.WriteString("}")

	return output.String()
}

type Class struct {
	Token       token.Token
	Identifier  *Identifier
	Fields      []*Field
	Subroutines []*Subroutine
}

func (c *Class) structureNode() {}
func (c *Class) TokenLiteral() string {
	return c.Token.Literal
}
func (c *Class) String() string {
	var output bytes.Buffer

	fields := make([]string, len(c.Fields))
	subs := make([]string, len(c.Subroutines))
	for i, field := range c.Fields {
		fields[i] = field.String()
	}
	for i, sub := range c.Subroutines {
		subs[i] = sub.String()
	}
	output.WriteString("class ")
	output.WriteString(c.Identifier.String())
	output.WriteString(" {")
	output.WriteString(strings.Join(fields, ", "))
	output.WriteString(strings.Join(subs, ", "))
	output.WriteString("}")

	return output.String()
}

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Index Expression
	Value Expression
}

func (l *LetStatement) statementNode()       {}
func (l *LetStatement) TokenLiteral() string { return l.Token.Literal }
func (l *LetStatement) String() string {
	var output bytes.Buffer

	output.WriteString("let ")
	output.WriteString(l.Name.String())
	if l.Index != nil {
		output.WriteString("[")
		output.WriteString(l.Index.String())
		output.WriteString("]")
	}
	output.WriteString(" = ")
	output.WriteString(l.Value.String())
	output.WriteString(";")

	return output.String()
}

type SubroutineCall struct {
	Token          token.Token
	CalleeName     *Identifier
	SubroutineName *Identifier
	Arguments      []Expression
}

func (s *SubroutineCall) expressionNode() {}
func (s *SubroutineCall) TokenLiteral() string {
	return s.Token.Literal
}
func (s *SubroutineCall) String() string {
	var output bytes.Buffer

	args := make([]string, len(s.Arguments))
	for i, a := range s.Arguments {
		args[i] = a.String()
	}
	if s.CalleeName != nil {
		output.WriteString(s.CalleeName.String())
		output.WriteString(".")
	}
	output.WriteString(s.SubroutineName.String())
	output.WriteString("(")
	output.WriteString(strings.Join(args, ", "))
	output.WriteString(")")

	return output.String()

}

type DoStatement struct {
	Token          token.Token
	SubroutineCall *SubroutineCall
}

func (d *DoStatement) statementNode() {}
func (d *DoStatement) TokenLiteral() string {
	return d.Token.Literal
}
func (d *DoStatement) String() string {
	var output bytes.Buffer
	output.WriteString("do ")
	output.WriteString(d.SubroutineCall.String())
	output.WriteString(";")
	return output.String()
}

type IntegerLiteral struct {
	Token token.Token
	Value int32
}

func (i *IntegerLiteral) expressionNode() {}
func (i *IntegerLiteral) TokenLiteral() string {
	return i.Token.Literal
}
func (i *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", i.Value)
}

type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (r *ReturnStatement) statementNode() {}
func (r *ReturnStatement) TokenLiteral() string {
	return r.Token.Literal
}
func (r *ReturnStatement) String() string {
	var output bytes.Buffer
	output.WriteString("return")
	if r.Value != nil {
		output.WriteString(" ")
		output.WriteString(r.Value.String())
	}
	output.WriteString(";")
	return output.String()
}

type KeywordConstantLiteral struct {
	Token token.Token
	Value string
}

func (t *KeywordConstantLiteral) expressionNode() {}
func (t *KeywordConstantLiteral) TokenLiteral() string {
	return t.Token.Literal
}
func (t *KeywordConstantLiteral) String() string {
	return t.Value
}

type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (i *IfStatement) statementNode() {}
func (i *IfStatement) TokenLiteral() string {
	return i.Token.Literal
}
func (i *IfStatement) String() string {
	var output bytes.Buffer
	output.WriteString("if ")
	output.WriteString("(")
	output.WriteString(i.Condition.String())
	output.WriteString("){")
	output.WriteString(i.Consequence.String())
	output.WriteString("}")
	if i.Alternative != nil {
		output.WriteString("else {")
		output.WriteString(i.Alternative.String())
		output.WriteString("}")
	}
	return output.String()
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (i *InfixExpression) expressionNode() {}
func (i *InfixExpression) TokenLiteral() string {
	return i.Token.Literal
}
func (i *InfixExpression) String() string {
	var output bytes.Buffer
	output.WriteString("(")
	output.WriteString(i.Left.String())
	output.WriteString(" " + i.Operator + " ")
	output.WriteString(i.Right.String())
	output.WriteString(")")
	return output.String()
}

type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStatement
}

func (w *WhileStatement) statementNode() {}
func (w *WhileStatement) TokenLiteral() string {
	return w.Token.Literal
}
func (w *WhileStatement) String() string {
	var output bytes.Buffer
	output.WriteString("while ")
	output.WriteString("(")
	output.WriteString(w.Condition.String())
	output.WriteString("){")
	output.WriteString(w.Body.String())
	output.WriteString("}")
	return output.String()
}

type PrefixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
}

func (p *PrefixExpression) expressionNode() {}
func (p *PrefixExpression) TokenLiteral() string {
	return p.Token.Literal
}
func (p *PrefixExpression) String() string {
	var output bytes.Buffer
	output.WriteString("(")
	output.WriteString(p.Operator)
	output.WriteString(p.Left.String())
	output.WriteString(")")
	return output.String()
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (i *IndexExpression) expressionNode() {}
func (i *IndexExpression) TokenLiteral() string {
	return i.Token.Literal
}
func (i *IndexExpression) String() string {
	var output bytes.Buffer
	output.WriteString(i.Left.String())
	output.WriteString("[")
	output.WriteString(i.Index.String())
	output.WriteString("]")
	return output.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (s *StringLiteral) expressionNode() {}
func (s *StringLiteral) TokenLiteral() string {
	return s.Token.Literal
}
func (s *StringLiteral) String() string {
	return s.Value
}
