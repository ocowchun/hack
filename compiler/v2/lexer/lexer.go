package lexer

import (
	"bufio"
	"bytes"
	"hack/compiler/v2/token"
	"io"
	"unicode"
)

type Lexer struct {
	scanner         *bufio.Scanner
	currentLine     []rune
	currentRune     rune
	currentPosition int
	peekPosition    int
	isEOF           bool
}

func New(reader io.Reader) *Lexer {
	l := &Lexer{
		scanner:         bufio.NewScanner(reader),
		currentPosition: 0,
		peekPosition:    0,
		currentLine:     make([]rune, 0),
		isEOF:           false,
	}
	l.nextRune()

	return l
}

func (l *Lexer) nextLine() {
	if l.isEOF {
		return
	}
	l.isEOF = !l.scanner.Scan()
	if l.isEOF {
		return
	}
	l.currentLine = []rune(l.scanner.Text())
	l.peekPosition = 0
	l.currentPosition = 0

}

func (l *Lexer) nextRune() {
	for l.peekPosition == len(l.currentLine) {
		l.nextLine()
		if l.isEOF {
			return
		}
	}
	l.currentPosition = l.peekPosition
	l.currentRune = l.currentLine[l.currentPosition]
	l.peekPosition++
}
func (l *Lexer) peekIsEndOfLine() bool {
	return l.peekPosition == len(l.currentLine) || l.isEOF
}

func (l *Lexer) PeekRune() (rune, bool) {
	if l.peekPosition >= len(l.currentLine) {
		return 0, false
	}
	return l.currentLine[l.peekPosition], true
}

func (l *Lexer) NextToken() token.Token {
	//l.skipWhitespace()
	l.skipCommentAndWhitespace()
	if l.isEOF {
		return token.Token{TokenType: token.TokenTypeEOF, Literal: ""}
	}
	// handle comment
	var tok token.Token

	switch l.currentRune {
	case ',':
		tok = token.Token{
			TokenType: token.TokenTypeComma,
			Literal:   ",",
		}
	case '(':
		tok = token.Token{
			TokenType: token.TokenTypeLeftParenthesis,
			Literal:   "(",
		}
	case ')':
		tok = token.Token{
			TokenType: token.TokenTypeRightParenthesis,
			Literal:   ")",
		}
	case '{':
		tok = token.Token{
			TokenType: token.TokenTypeLeftBrace,
			Literal:   "{",
		}
	case '}':
		tok = token.Token{
			TokenType: token.TokenTypeRightBrace,
			Literal:   "}",
		}
	case ';':
		tok = token.Token{
			TokenType: token.TokenTypeSemicolon,
			Literal:   ";",
		}
	case '=':
		tok = token.Token{
			TokenType: token.TokenTypeAssign,
			Literal:   "=",
		}
	case '~':
		tok = token.Token{
			TokenType: token.TokenTypeTilde,
			Literal:   "~",
		}
	case '[':
		tok = token.Token{
			TokenType: token.TokenTypeLeftBracket,
			Literal:   "[",
		}
	case ']':
		tok = token.Token{
			TokenType: token.TokenTypeRightBracket,
			Literal:   "]",
		}
	case '>':
		tok = token.Token{
			TokenType: token.TokeTypeGreater,
			Literal:   ">",
		}
	case '<':
		tok = token.Token{
			TokenType: token.TokenTypeLess,
			Literal:   "<",
		}
	case '+':
		tok = token.Token{
			TokenType: token.TokenTypePlus,
			Literal:   "+",
		}
	case '-':
		tok = token.Token{
			TokenType: token.TokenTypeMinus,
			Literal:   "-",
		}
	case '*':
		tok = token.Token{
			TokenType: token.TokenTypeAsterisk,
			Literal:   "*",
		}
	case '/':
		tok = token.Token{
			TokenType: token.TokenTypeSlash,
			Literal:   "/",
		}
	case '&':
		tok = token.Token{
			TokenType: token.TokenTypeAmpersand,
			Literal:   "&",
		}
	case '|':
		tok = token.Token{
			TokenType: token.TokenTypeVerticalBar,
			Literal:   "|",
		}
	case '.':
		tok = token.Token{
			TokenType: token.TokenTypeDot,
			Literal:   ".",
		}
	case '"':
		lit, err := l.readStringLiteral()
		if err != nil {
			return token.Token{TokenType: token.TokenTypeIllegal, Literal: "illegal"}
		}
		tok = token.Token{
			TokenType: token.TokenTypeStringLiteral,
			Literal:   lit,
		}

	default:
		if isLetter(l.currentRune) {
			lit := l.readIdentifier()
			tok = token.Token{
				TokenType: token.LookupIdentifier(lit),
				Literal:   lit,
			}

		} else if isDigit(l.currentRune) {
			tok = token.Token{
				TokenType: token.TokenTypeIntegerLiteral,
				Literal:   l.readIdentifier(),
			}
		} else {
			tok = token.Token{
				TokenType: token.TokenTypeIllegal,
				Literal:   "illegal",
			}
		}
		return tok
	}

	l.nextRune()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for !l.isEOF && unicode.IsSpace(l.currentRune) {
		l.nextRune()
	}
}

func (l *Lexer) skipCommentAndWhitespace() {
	l.skipWhitespace()
	if l.isEOF {
		return
	}
	for l.currentRune == '/' {
		next, ok := l.PeekRune()
		if !ok {
			return
		}
		if next == '/' {
			l.nextLine()
			l.nextRune()
		} else if next == '*' {
			//multi line comment
			prev := l.currentRune
			l.nextRune()
			if l.isEOF {
				return
			}

			for {
				if prev == '*' && l.currentRune == '/' {
					l.nextRune()
					break
				}
				prev = l.currentRune
				l.nextRune()
			}

		} else {

		}

		l.skipWhitespace()
		if l.isEOF {
			return
		}

	}

}

func (l *Lexer) readStringLiteral() (string, error) {
	var output bytes.Buffer
	l.nextRune()
	for {
		if l.isEOF {
			return "", io.EOF
		}
		if l.currentRune == '"' {
			break
		} else if l.currentRune == '\\' {
			prev := l.currentRune
			l.nextRune()
			if l.isEOF {
				return "", io.EOF
			}
			if l.currentRune != '"' {
				output.WriteRune(prev)
			}
		}

		output.WriteRune(l.currentRune)
		l.nextRune()
	}

	return output.String(), nil
}
func isLetter(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
func (l *Lexer) readNumber() string {
	var output bytes.Buffer
	for isDigit(l.currentRune) {
		output.WriteRune(l.currentRune)
		l.nextRune()
		if l.isEOF {
			break
		}
	}

	return output.String()
}
func (l *Lexer) readIdentifier() string {
	var output bytes.Buffer
	for isLetter(l.currentRune) || isDigit(l.currentRune) {
		output.WriteRune(l.currentRune)
		l.nextRune()
		if l.isEOF {
			break
		}
	}

	return output.String()
}
