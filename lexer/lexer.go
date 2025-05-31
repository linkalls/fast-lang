package lexer

import (
	"strconv" // Added: for strconv.ParseInt
	"strings" // Added: for strings.Builder and strings.Contains (though Contains might not be used anymore)
	"unicode"

	"github.com/linkalls/zeno-lang/token"
)

// Lexer represents the lexical analyzer
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

// New creates a new instance of Lexer
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar gives us the next character and advances our position in the input string
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL character signifies "EOF"
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// peekChar returns the next character without advancing our position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment skips single-line and multi-line comments
func (l *Lexer) skipComment() bool {
	if l.ch == '/' && l.peekChar() == '/' {
		// Single-line comment
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		return true
	} else if l.ch == '/' && l.peekChar() == '*' {
		// Multi-line comment
		l.readChar() // consume '/'
		l.readChar() // consume '*'

		for {
			if l.ch == 0 {
				// Unterminated comment
				break
			}
			if l.ch == '*' && l.peekChar() == '/' {
				l.readChar() // consume '*'
				l.readChar() // consume '/'
				break
			}
			l.readChar()
		}
		return true
	}
	return false
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float)
func (l *Lexer) readNumber() (token.TokenType, string) {
	position := l.position
	tokenType := token.INT

	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = token.FLOAT
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return tokenType, l.input[position:l.position]
}

// readString reads a string literal
func (l *Lexer) readString() (string, bool) {
	position := l.position + 1 // skip opening quote
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar() // consume backslash
			if l.ch == 0 {
				break
			}
		}
	}

	if l.ch == 0 {
		// Unterminated string
		return "", false
	}

	str := l.input[position:l.position]
	return str, true
}

// NextToken returns the next token in the input
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	// Skip comments
	for l.skipComment() {
		l.skipWhitespace()
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '*':
		tok = newToken(token.MULTIPLY, l.ch)
	case '/':
		if l.skipComment() {
			return l.NextToken()
		} else {
			tok = newToken(token.DIVIDE, l.ch)
		}
	case '%':
		tok = newToken(token.MODULO, l.ch)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '.':
		if l.peekChar() == '.' {
			// Look ahead one more character to check for ...
			nextPos := l.readPosition + 1
			if nextPos < len(l.input) && l.input[nextPos] == '.' {
				// It's a variadic operator ...
				l.readChar() // consume second .
				l.readChar() // consume third .
				tok = token.Token{Type: token.DOTDOTDOT, Literal: "..."}
			} else {
				tok = newToken(token.ILLEGAL, l.ch)
			}
		} else {
			tok = newToken(token.ILLEGAL, l.ch) // Single dot not supported yet
		}
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '"':
		str, ok := l.readString()
		if !ok {
			tok = newToken(token.ILLEGAL, l.ch)
		} else {
			tok.Type = token.STRING
			tok.Literal = str
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tokenType, literal := l.readNumber()
			tok.Type = tokenType
			tok.Literal = literal
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func processEscapeSequence(str string) string {
	var result strings.Builder
	result.Grow(len(str))
	for i := 0; i < len(str); i++ {
		if str[i] == '\\' && i+1 < len(str) {
			i++
			switch str[i] {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			case 'u':
				if i+4 < len(str) {
					val, err := strconv.ParseInt(str[i+1:i+5], 16, 32)
					if err == nil {
						result.WriteRune(rune(val))
						i += 4
						continue
					}
				}
				result.WriteByte('u')
			case 'x':
				if i+2 < len(str) {
					val, err := strconv.ParseInt(str[i+1:i+3], 16, 32)
					if err == nil {
						result.WriteByte(byte(val))
						i += 2
						continue
					}
				}
				result.WriteByte('x')
			default:
				result.WriteByte('\\')
				result.WriteByte(str[i])
			}
		} else {
			result.WriteByte(str[i])
		}
	}
	return result.String()
}

func ProcessStringLiteral(literal string) string {
	return processEscapeSequence(literal)
}
