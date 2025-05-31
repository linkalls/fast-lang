package lexer

import (
	"testing"

	"github.com/linkalls/zeno-lang/token"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5
let ten = 10.5
let add = fn(x, y) {
  x + y
}

let result = add(five, ten)
!-/+5
5 < 10 > 5

if (5 < 10) {
    return true
} else {
    return false
}

10 == 10
10 != 9
"foobar"
"foo bar"
// This is a comment
/* This is a 
   multi-line comment */
let x = 5
while (x > 0) {
    print("Hello")
    x = x - 1
}
for (let i = 0 i < 10 i = i + 1) {
    println("World")
}
loop {
    break
}
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ""},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.FLOAT, "10.5"},
		{token.SEMICOLON, ""},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FN, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ""},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ""},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ""},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.DIVIDE, "/"},
		{token.PLUS, "+"},
		{token.INT, "5"},
		{token.SEMICOLON, ""},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ""},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ""},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ""},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ""},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.WHILE, "while"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.GT, ">"},
		{token.INT, "0"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.PRINT, "print"},
		{token.LPAREN, "("},
		{token.STRING, "Hello"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ""},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.IDENT, "x"},
		{token.MINUS, "-"},
		{token.INT, "1"},
		{token.SEMICOLON, ""},
		{token.RBRACE, "}"},
		{token.FOR, "for"},
		{token.LPAREN, "("},
		{token.LET, "let"},
		{token.IDENT, "i"},
		{token.ASSIGN, "="},
		{token.INT, "0"},
		{token.SEMICOLON, ""},
		{token.IDENT, "i"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.SEMICOLON, ""},
		{token.IDENT, "i"},
		{token.ASSIGN, "="},
		{token.IDENT, "i"},
		{token.PLUS, "+"},
		{token.INT, "1"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.PRINTLN, "println"},
		{token.LPAREN, "("},
		{token.STRING, "World"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ""},
		{token.RBRACE, "}"},
		{token.LOOP, "loop"},
		{token.LBRACE, "{"},
		{token.BREAK, "break"},
		{token.SEMICOLON, ""},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	input := `"hello world"
"test\nstring"
"tab\ttest"
"quote\"test"
"backslash\\test"`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.STRING, "hello world"},
		{token.SEMICOLON, ""},
		{token.STRING, "test\nstring"},
		{token.SEMICOLON, ""},
		{token.STRING, "tab\ttest"},
		{token.SEMICOLON, ""},
		{token.STRING, "quote\"test"},
		{token.SEMICOLON, ""},
		{token.STRING, "backslash\\test"},
		{token.SEMICOLON, ""},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Type == token.STRING {
			processed := ProcessStringLiteral(tok.Literal)
			if processed != tt.expectedLiteral {
				t.Fatalf("tests[%d] - processed literal wrong. expected=%q, got=%q",
					i, tt.expectedLiteral, processed)
			}
		} else if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `// Single line comment
let x = 5
/* Multi-line
   comment */
let y = 10
/* Another comment */ let z = 15`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ""},
		{token.LET, "let"},
		{token.IDENT, "y"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ""},
		{token.LET, "let"},
		{token.IDENT, "z"},
		{token.ASSIGN, "="},
		{token.INT, "15"},
		{token.SEMICOLON, ""},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestFloatNumbers(t *testing.T) {
	input := `3.14159
0.5
123.456
1.0`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.FLOAT, "3.14159"},
		{token.SEMICOLON, ""},
		{token.FLOAT, "0.5"},
		{token.SEMICOLON, ""},
		{token.FLOAT, "123.456"},
		{token.SEMICOLON, ""},
		{token.FLOAT, "1.0"},
		{token.SEMICOLON, ""},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestComparisonOperators(t *testing.T) {
	input := `5 <= 10
10 >= 5
5 && 10
5 || 10`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "5"},
		{token.LTE, "<="},
		{token.INT, "10"},
		{token.SEMICOLON, ""},
		{token.INT, "10"},
		{token.GTE, ">="},
		{token.INT, "5"},
		{token.SEMICOLON, ""},
		{token.INT, "5"},
		{token.AND, "&&"},
		{token.INT, "10"},
		{token.SEMICOLON, ""},
		{token.INT, "5"},
		{token.OR, "||"},
		{token.INT, "10"},
		{token.SEMICOLON, ""},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestProcessStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello\\nworld", "hello\nworld"},
		{"tab\\there", "tab\there"},
		{"quote\\\"test", "quote\"test"},
		{"backslash\\\\test", "backslash\\test"},
		{"mixed\\n\\t\\\"content\\\\", "mixed\n\t\"content\\"},
	}

	for i, tt := range tests {
		result := ProcessStringLiteral(tt.input)
		if result != tt.expected {
			t.Errorf("test[%d] - ProcessStringLiteral wrong. input=%q, expected=%q, got=%q",
				i, tt.input, tt.expected, result)
		}
	}
}
