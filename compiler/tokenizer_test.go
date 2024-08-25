package compiler

import (
	"io"
	"strings"
	"testing"
)

func TestTokenizer_Next(t *testing.T) {
	reader := strings.NewReader(" this is 1 \"test\";")
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "this", tokenType: KeywordTokenType},
		{content: "is", tokenType: IdentifierTokenType},
		{content: "1", tokenType: IntegerConstantTokenType},
		{content: "test", tokenType: StringConstantTokenType},
		{content: ";", tokenType: SymbolTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_with_space_only(t *testing.T) {
	reader := strings.NewReader("     ")
	tokenizer := NewTokenizer(reader)

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_with_comment(t *testing.T) {
	reader := strings.NewReader(" // this is a boring comment")
	tokenizer := NewTokenizer(reader)

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_case2(t *testing.T) {
	reader := strings.NewReader(" int a1")
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "int", tokenType: KeywordTokenType},
		{content: "a1", tokenType: IdentifierTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_case3(t *testing.T) {
	reader := strings.NewReader(" boolean a")
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "boolean", tokenType: KeywordTokenType},
		{content: "a", tokenType: IdentifierTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_case4(t *testing.T) {
	reader := strings.NewReader(" 42")
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "42", tokenType: IntegerConstantTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_case5(t *testing.T) {
	reader := strings.NewReader(" \"space \" ")
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "space ", tokenType: StringConstantTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_case6(t *testing.T) {
	reader := strings.NewReader("-13")
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "-", tokenType: SymbolTokenType},
		{content: "13", tokenType: IntegerConstantTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_case7(t *testing.T) {
	reader := strings.NewReader("-13 // this is a comment")
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "-", tokenType: SymbolTokenType},
		{content: "13", tokenType: IntegerConstantTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}

func TestTokenizer_Next_case8(t *testing.T) {
	code := `
/**
* this is test
*/
-13
	`
	reader := strings.NewReader(code)
	tokenizer := NewTokenizer(reader)

	for _, expected := range []Token{
		{content: "-", tokenType: SymbolTokenType},
		{content: "13", tokenType: IntegerConstantTokenType},
	} {
		actual, err := tokenizer.Next()
		if err != nil {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}

		if expected != actual {
			t.Fatalf("expected %s, but got error: %s", expected, err)
		}
	}

	_, err := tokenizer.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, but got: %s", err)
	}
}
