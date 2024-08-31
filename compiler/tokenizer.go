package compiler

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Tokenizer struct {
	reader       *bufio.Scanner
	hasNext      bool
	currentLine  string
	buffer       string
	currentToken Token
}

func NewTokenizer(reader io.Reader) *Tokenizer {
	//r := bufio.NewReader(reader)
	r := bufio.NewScanner(reader)

	tokenizer := &Tokenizer{reader: r, hasNext: true}

	return tokenizer
}

type TokenType uint8

const (
	//class,constructor,function,method,field,static,var,int,char,boolean,void,true,false,null,this,let,do,if,else
	// while,return
	UnknownTokenType TokenType = iota
	KeywordTokenType
	// {, },(,),[,],.,`,`,;,+,-,*,/,&,|,<,>,=,~
	SymbolTokenType
	// a decimal number in the range 0 ..32767
	IntegerConstantTokenType
	// '"' a sequence of Unicode characters, not including double quote or newline '"'
	StringConstantTokenType
	// a sequence of letters, digits and underscore(_), not starting with a digit
	IdentifierTokenType
)

type Token struct {
	content   string
	tokenType TokenType
}

func NewToken(content string, tokenType TokenType) Token {
	return Token{
		content:   content,
		tokenType: tokenType,
	}
}

func (t Token) Type() TokenType {
	return t.tokenType
}

func (t Token) Content() string {
	return t.content
}

func (t TokenType) String() string {
	switch t {
	case KeywordTokenType:
		return "keyword"
	case SymbolTokenType:
		return "symbol"
	case IntegerConstantTokenType:
		return "integerConstant"
	case StringConstantTokenType:
		return "stringConstant"
	case IdentifierTokenType:
		return "identifier"
	default:
		panic("unhandled default case")
	}
}

func (t Token) String() string {
	content := t.content
	switch content {
	case "<":
		content = "&lt;"
	case ">":
		content = "&gt;"
	case "\"":
		content = "&quot;"
	case "&":
		content = "&amp;"
	}
	return fmt.Sprintf("<%s> %s </%s>", t.tokenType, content, t.tokenType)
}

var keywordMap = map[string]bool{
	"class":       true,
	"constructor": true,
	"function":    true,
	"method":      true,
	"field":       true,
	"static":      true,
	"var":         true,
	"int":         true,
	"char":        true,
	"boolean":     true,
	"void":        true,
	"true":        true,
	"false":       true,
	"null":        true,
	"this":        true,
	"let":         true,
	"do":          true,
	"if":          true,
	"while":       true,
	"else":        true,
	"return":      true,
}
var symbolMap = map[string]bool{
	"{": true,
	"}": true,
	"(": true,
	")": true,
	"[": true,
	"]": true,
	".": true,
	",": true,
	";": true,
	"+": true,
	"-": true,
	"*": true,
	"/": true,
	"&": true,
	"|": true,
	"<": true,
	">": true,
	"=": true,
	"~": true,
}

func (t *Tokenizer) Current() (Token, error) {
	if !t.hasNext {
		return Token{}, io.EOF
	}
	// TODO:handle init!
	return t.currentToken, nil
}

func (t *Tokenizer) CurrentLine() string {
	return t.currentLine
}

// Next returns Token{}, and error, if error = io.EOF it reaches the end
func (t *Tokenizer) Next() (Token, error) {
	if len(t.buffer) > 0 {
		newBuffer := ""
		for _, c := range t.buffer {
			if len(newBuffer) == 0 && isSpace(c) {
				continue
			}
			newBuffer += string(c)
		}
		// handle comment line
		if strings.HasPrefix(newBuffer, "//") {
			t.buffer = ""
		} else {
			t.buffer = newBuffer
		}
	}

	isMultipleLineComment := false
	for len(t.buffer) == 0 {
		if !t.reader.Scan() {
			t.hasNext = false
			return Token{}, io.EOF
		}
		line := ""
		for _, c := range t.reader.Text() {
			// handle space line
			if len(line) == 0 && isSpace(c) {
				continue
			}
			line += string(c)
		}

		// handle single comment line
		if strings.HasPrefix(line, "//") || (strings.HasPrefix(line, "/**") && strings.HasSuffix(line, "*/")) {
			continue
		}

		// handle multiple line comment
		if strings.HasPrefix(line, "/**") {
			isMultipleLineComment = true
			continue
		}
		if isMultipleLineComment {
			if strings.HasSuffix(line, "*/") {
				isMultipleLineComment = false
				continue
			} else if strings.HasPrefix(line, "*") {
				continue
			} else {
				return Token{}, fmt.Errorf("invalid line: %s when handling multiple line comment", line)
			}
		}

		t.currentLine = line
		t.buffer = t.currentLine
	}

	buffer := t.buffer
	content := ""
	isCompleted := false
	tokenType := UnknownTokenType

	i := 0
	for i < len(buffer) {
		r := rune(buffer[i])
		if content == "" {
			// add anything except space or tab
			if !isSpace(r) {
				if isDigit(r) {
					tokenType = IntegerConstantTokenType
					content += string(r)
				} else if r == '"' {
					tokenType = StringConstantTokenType
				} else {
					content += string(r)
					if _, ok := symbolMap[content]; ok {
						tokenType = SymbolTokenType
						buffer = buffer[i+1:]
						isCompleted = true
						break
					}
				}
			}
		} else {

			if tokenType == IntegerConstantTokenType {
				if isDigit(r) {
					if isZero(rune(content[0])) {
						return Token{}, fmt.Errorf("leading zero is not supported")
					}
					content += string(r)
					num, err := strconv.ParseInt(content, 10, 32)
					if err != nil {
						return Token{}, fmt.Errorf("failed to parse integer, error: %s", err)
					}
					if num > 32767 {
						// at most 32767
						return Token{}, fmt.Errorf("integer can't greater than 32767")
					}
				} else {
					if isSpace(r) {
						isCompleted = true
						buffer = buffer[i+1:]
						break
					} else {
						if _, ok := symbolMap[string(r)]; ok {
							isCompleted = true
							buffer = buffer[i:]
							break
						}
						// TODO: handle symbol!
						content += string(r)
						return Token{}, fmt.Errorf("failed to parse %s as integer", content)
					}
				}
			} else if tokenType == StringConstantTokenType {
				if r == '"' {
					isCompleted = true
					buffer = buffer[i+1:]
					break
				} else {
					content += string(r)
				}
			} else {
				// either keyword or identifier
				if isSpace(r) {
					isCompleted = true
					buffer = buffer[i+1:]
				} else if _, ok := symbolMap[string(r)]; ok {
					isCompleted = true
					buffer = buffer[i:]
				} else {
					content += string(r)
				}

				if isCompleted {
					if _, ok := keywordMap[content]; ok {
						tokenType = KeywordTokenType
					} else {
						if validateIdentifierFormat(content) {
							tokenType = IdentifierTokenType
						} else {
							return Token{}, fmt.Errorf("failed to parse %s as identifier", content)
						}
					}
					break
				}
			}
		}
		i += 1
	}

	if !isCompleted {
		// handle valid case first
		if tokenType == IntegerConstantTokenType {
			buffer = ""
		} else if tokenType == UnknownTokenType {
			if _, ok := keywordMap[content]; ok {
				tokenType = KeywordTokenType
			} else {
				if validateIdentifierFormat(content) {
					tokenType = IdentifierTokenType
				} else {
					return Token{}, fmt.Errorf("failed to parse %s as identifier", content)
				}
			}
			buffer = ""
		} else {
			return Token{}, fmt.Errorf("doesn't support multiple-line statement %s", t.currentLine)
		}

	}
	t.buffer = buffer

	t.currentToken = Token{tokenType: tokenType, content: content}
	return t.currentToken, nil
}

func validateIdentifierFormat(word string) bool {
	if len(word) == 0 {
		return false
	}

	pattern := `^[a-zA-Z0-9_]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(word) {
		return false
	}

	return !isDigit(rune(word[0]))
}

func isSpace(r rune) bool {
	if r <= '\u00FF' {
		// Obvious ASCII ones: \t through \r plus space. Plus two Latin-1 oddballs.
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	// High-valued ones.
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}

func isZero(r rune) bool {
	return r == '0'
}

func isDigit(r rune) bool {
	if isZero(r) {
		return true
	} else if r == '1' {
		return true
	} else if r == '2' {
		return true
	} else if r == '3' {
		return true
	} else if r == '4' {
		return true
	} else if r == '5' {
		return true
	} else if r == '6' {
		return true
	} else if r == '7' {
		return true
	} else if r == '8' {
		return true
	} else if r == '9' {
		return true
	}
	return false
}
