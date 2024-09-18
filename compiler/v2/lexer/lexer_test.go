package lexer

import (
	"hack/compiler/v2/token"
	"strings"
	"testing"
)

func TestLexer_NextToken(t *testing.T) {
	content := `
this is 1 "test";
int a1;
int a2; // test a2
 // abc

	/**
	* this is a multi line comment
	* ahhhh
	*/
int a3;
/** multi lien format in single line */
boolean b1;

`

	reader := strings.NewReader(content)
	lexer := New(reader)

	for _, expected := range []token.Token{
		{TokenType: token.TokenTypeThis, Literal: "this"},
		{TokenType: token.TokenTypeIdentifier, Literal: "is"},
		{TokenType: token.TokenTypeIntegerLiteral, Literal: "1"},
		{TokenType: token.TokenTypeStringLiteral, Literal: "test"},
		{TokenType: token.TokenTypeSemicolon, Literal: ";"},
		{TokenType: token.TokenTypeInt, Literal: "int"},
		{TokenType: token.TokenTypeIdentifier, Literal: "a1"},
		{TokenType: token.TokenTypeSemicolon, Literal: ";"},
		{TokenType: token.TokenTypeInt, Literal: "int"},
		{TokenType: token.TokenTypeIdentifier, Literal: "a2"},
		{TokenType: token.TokenTypeSemicolon, Literal: ";"},
		{TokenType: token.TokenTypeInt, Literal: "int"},
		{TokenType: token.TokenTypeIdentifier, Literal: "a3"},
		{TokenType: token.TokenTypeSemicolon, Literal: ";"},
		{TokenType: token.TokenTypeBoolean, Literal: "boolean"},
		{TokenType: token.TokenTypeIdentifier, Literal: "b1"},
		{TokenType: token.TokenTypeSemicolon, Literal: ";"},
		{TokenType: token.TokenTypeEOF, Literal: ""},
	} {
		actual := lexer.NextToken()
		if expected != actual {
			t.Fatalf("expected %s, but got : %s", expected, actual)
		}
	}
}
