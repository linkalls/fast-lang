package token

// TokenType represents the type of a token
type TokenType string

// Token represents a token in the Zeno language
type Token struct {
	Type    TokenType
	Literal string
}

// Token types for the Zeno language
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers + literals
	IDENT  TokenType = "IDENT"  // add, foobar, x, y, ...
	INT    TokenType = "INT"    // 1343456
	FLOAT  TokenType = "FLOAT"  // 3.14159
	STRING TokenType = "STRING" // "foobar"

	// Keywords
	LET      TokenType = "LET"
	PUB      TokenType = "PUB"
	IMPORT   TokenType = "IMPORT"
	FROM     TokenType = "FROM"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	LOOP     TokenType = "LOOP"
	WHILE    TokenType = "WHILE"
	FOR      TokenType = "FOR"
	FN       TokenType = "FN"
	RETURN   TokenType = "RETURN"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	// PRINT    TokenType = "PRINT"    // Removed as keyword
	// PRINTLN  TokenType = "PRINTLN"  // Removed as keyword
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	MULTIPLY TokenType = "*"
	DIVIDE   TokenType = "/"
	MODULO   TokenType = "%"
	BANG     TokenType = "!"
	EQ       TokenType = "=="
	NOT_EQ   TokenType = "!="
	LT       TokenType = "<"
	LTE      TokenType = "<="
	GT       TokenType = ">"
	GTE      TokenType = ">="
	AND      TokenType = "&&"
	OR       TokenType = "||"

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ""
	COLON     TokenType = ":"
	DOTDOTDOT TokenType = "..."
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"
)

// keywords maps string literals to their token types
var keywords = map[string]TokenType{
	"let":      LET,
	"pub":      PUB,
	"import":   IMPORT,
	"from":     FROM,
	"if":       IF,
	"else":     ELSE,
	"loop":     LOOP,
	"while":    WHILE,
	"for":      FOR,
	"fn":       FN,
	"return":   RETURN,
	"true":     TRUE,
	"false":    FALSE,
	// "print":    PRINT,    // Removed as keyword
	// "println":  PRINTLN,  // Removed as keyword
	"break":    BREAK,
	"continue": CONTINUE,
}

// LookupIdent checks if the identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
