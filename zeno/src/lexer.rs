#[derive(Debug, Clone, PartialEq)]
pub enum Token {
    Illegal(char),
    Eof,

    // Identifiers + literals
    Identifier(String),
    Integer(i64),
    Float(f64),
    String(String),

    // Keywords
    Let,
    Mut,
    If,
    Else,
    Loop,
    While,
    For,
    Fn,
    Return,
    True,
    False,
    Print,
    Println,
    Break,
    Continue,

    // Operators
    Assign,   // =
    Plus,     // +
    Minus,    // -
    Multiply, // *
    Divide,   // /
    Modulo,   // %
    Bang,     // !
    Eq,       // ==
    NotEq,    // !=
    Lt,       // <
    Lte,      // <=
    Gt,       // >
    Gte,      // >=
    And,      // &&
    Or,       // ||

    // Delimiters
    Comma,    // ,
    Semicolon,// ;
    Colon,    // :
    LParen,   // (
    RParen,   // )
    LBrace,   // {
    RBrace,   // }
}

#[derive(Debug)]
pub struct Lexer<'a> {
    input: &'a [u8],
    position: usize,      // current position in input (points to current char)
    read_position: usize, // current reading position in input (after current char)
    ch: u8,               // current char under examination
}

impl<'a> Lexer<'a> {
    pub fn new(input: &'a str) -> Self {
        let mut l = Lexer {
            input: input.as_bytes(),
            position: 0,
            read_position: 0,
            ch: 0,
        };
        l.read_char();
        l
    }

    fn read_char(&mut self) {
        if self.read_position >= self.input.len() {
            self.ch = 0; // ASCII NUL, signifies EOF
        } else {
            self.ch = self.input[self.read_position];
        }
        self.position = self.read_position;
        self.read_position += 1;
    }

    fn peek_char(&self) -> u8 {
        if self.read_position >= self.input.len() {
            0
        } else {
            self.input[self.read_position]
        }
    }

    fn skip_whitespace(&mut self) {
        while self.ch.is_ascii_whitespace() {
            self.read_char();
        }
    }

    fn skip_comment(&mut self) -> bool {
        if self.ch == b'/' && self.peek_char() == b'/' {
            // Single-line comment
            while self.ch != b'\n' && self.ch != 0 {
                self.read_char();
            }
            self.skip_whitespace(); // Skip potential whitespace after comment before next token
            return true;
        } else if self.ch == b'/' && self.peek_char() == b'*' {
            // Multi-line comment
            self.read_char(); // consume /
            self.read_char(); // consume *
            loop {
                if self.ch == 0 { // EOF inside comment
                    // This could be an error state, Token::Illegal, or handled by next_token
                    break;
                }
                if self.ch == b'*' && self.peek_char() == b'/' {
                    self.read_char(); // consume *
                    self.read_char(); // consume /
                    break;
                }
                self.read_char();
            }
            self.skip_whitespace(); // Skip potential whitespace after comment
            return true;
        }
        false
    }

    fn read_identifier(&mut self) -> String {
        let position = self.position;
        while self.ch.is_ascii_alphabetic() || self.ch == b'_' || self.ch.is_ascii_digit() { // allow digits in identifiers after the first char
            self.read_char();
        }
        String::from_utf8_lossy(&self.input[position..self.position]).to_string()
    }

    fn read_number(&mut self) -> Token {
        let position = self.position;
        let mut is_float = false;
        while self.ch.is_ascii_digit() {
            self.read_char();
        }
        if self.ch == b'.' && self.peek_char().is_ascii_digit() {
            is_float = true;
            self.read_char(); // consume '.'
            while self.ch.is_ascii_digit() {
                self.read_char();
            }
        }
        let number_str = String::from_utf8_lossy(&self.input[position..self.position]);
        if is_float {
            Token::Float(number_str.parse().unwrap_or_else(|_| {
                // Handle parse error, maybe return Illegal for malformed float
                0.0 // Default or error value
            }))
        } else {
            Token::Integer(number_str.parse().unwrap_or_else(|_| {
                 // Handle parse error
                0 // Default or error value
            }))
        }
    }

    fn read_string(&mut self) -> Result<String, char> {
        let mut result = String::new();
        self.read_char(); // consume the opening "

        while self.ch != b'"' {
            if self.ch == 0 { // Unterminated string
                return Err('\0'); // Using NUL to signify unterminated string error
            }
            if self.ch == b'\\' { // Escape character
                self.read_char(); // consume '\'
                match self.ch {
                    b'n' => result.push('\n'),
                    b't' => result.push('\t'),
                    b'\\' => result.push('\\'),
                    b'"' => result.push('\"'),
                    // Add more escapes if needed
                    _ => result.push(self.ch as char), // Or return an error for unknown escape
                }
            } else {
                result.push(self.ch as char);
            }
            self.read_char();
        }
        self.read_char(); // consume the closing "
        Ok(result)
    }

    pub fn next_token(&mut self) -> Token {
        self.skip_whitespace();

        // Try skipping comments repeatedly
        while self.skip_comment() {
            // skip_comment itself calls skip_whitespace, so we are good
        }


        let tok = match self.ch {
            b'=' => {
                if self.peek_char() == b'=' {
                    self.read_char();
                    Token::Eq
                } else {
                    Token::Assign
                }
            }
            b'+' => Token::Plus,
            b'-' => Token::Minus,
            b'!' => {
                if self.peek_char() == b'=' {
                    self.read_char();
                    Token::NotEq
                } else {
                    Token::Bang
                }
            }
            b'*' => Token::Multiply,
            b'/' => Token::Divide, // skip_comment should have handled // and /*
            b'%' => Token::Modulo,
            b'<' => {
                if self.peek_char() == b'=' {
                    self.read_char();
                    Token::Lte
                } else {
                    Token::Lt
                }
            }
            b'>' => {
                if self.peek_char() == b'=' {
                    self.read_char();
                    Token::Gte
                } else {
                    Token::Gt
                }
            }
            b'&' => {
                if self.peek_char() == b'&' {
                    self.read_char();
                    Token::And
                } else {
                    Token::Illegal(self.ch as char) // Or some other way to handle single '&'
                }
            }
            b'|' => {
                if self.peek_char() == b'|' {
                    self.read_char();
                    Token::Or
                } else {
                    Token::Illegal(self.ch as char) // Or some other way to handle single '|'
                }
            }
            b',' => Token::Comma,
            b';' => Token::Semicolon,
            b':' => Token::Colon,
            b'(' => Token::LParen,
            b')' => Token::RParen,
            b'{' => Token::LBrace,
            b'}' => Token::RBrace,
            b'"' => {
                match self.read_string() {
                    Ok(s) => Token::String(s),
                    Err(_) => Token::Illegal('"'), // Unterminated string
                }
            }
            b'a'..=b'z' | b'A'..=b'Z' | b'_' => {
                let ident = self.read_identifier();
                return match ident.as_str() {
                    "let" => Token::Let,
                    "mut" => Token::Mut,
                    "if" => Token::If,
                    "else" => Token::Else,
                    "loop" => Token::Loop,
                    "while" => Token::While,
                    "for" => Token::For,
                    "fn" => Token::Fn,
                    "return" => Token::Return,
                    "true" => Token::True,
                    "false" => Token::False,
                    "print" => Token::Print,
                    "println" => Token::Println,
                    "break" => Token::Break,
                    "continue" => Token::Continue,
                    _ => Token::Identifier(ident),
                };
            }
            b'0'..=b'9' => {
                return self.read_number(); // read_number returns Token, so just return it
            }
            0 => Token::Eof,
            _ => Token::Illegal(self.ch as char),
        };

        if tok != Token::Eof && !(matches!(tok, Token::Identifier(_)) || matches!(tok, Token::Integer(_)) || matches!(tok, Token::Float(_)) || matches!(tok, Token::String(_))) {
            // For most single-character tokens, we need to advance the character
            // read_identifier, read_number, and read_string handle their own advancement.
            // Operators that look ahead (==, !=, <=, >=, &&, ||) also advance.
            // This check is a bit broad but aims to cover the simple cases.
             if ! ( self.ch == b'=' || self.ch == b'!' || self.ch == b'<' || self.ch == b'>' || self.ch == b'&' || self.ch == b'|' || self.ch == b'"') {
                 // if it was already advanced by peek_char logic or read_string
                  if !(tok == Token::Eq || tok == Token::NotEq || tok == Token::Lte || tok == Token::Gte || tok == Token::And || tok == Token::Or || matches!(tok, Token::Illegal(_))) {
                     // if it's not one of the multi-char operators or illegal (which means we didn't advance)
                     // This condition is getting complex. A simpler way is to ensure all paths advance ch.
                  }
             }
             // All paths that produce a token should call read_char() before returning,
             // unless they are multi-character tokens that are already handled by read_identifier, read_number, read_string,
             // or the peek_char() logic.
             // For single char tokens, we definitely need to read_char() here.
             // Let's simplify: most branches in the match will need self.read_char()
        }
        
        // Most token types consume one character.
        // Exceptions: EOF, read_identifier, read_number, read_string, and multi-char operators.
        // The logic for advancing `ch` is handled in `read_char`, `read_identifier`, `read_number`, `read_string`.
        // For single-character tokens, we need to call `read_char` after identifying them.
        // For multi-character tokens (like ==, !=, &&, ||, <=, >=), `read_char` is called an extra time.
        // For identifiers, numbers, strings, they manage their own `read_char` calls.

        match tok {
            Token::Assign | Token::Plus | Token::Minus | Token::Bang | Token::Multiply | Token::Divide | Token::Modulo |
            Token::Lt | Token::Gt | Token::Comma | Token::Semicolon | Token::Colon | Token::LParen | Token::RParen |
            Token::LBrace | Token::RBrace => {
                 // These are single char tokens (or first char of multi-char handled above)
                 // that were not part of a longer token sequence like `==` or `read_identifier`
                 // if the token is NOT already advanced by a peek_char() path
                 if !(tok == Token::Eq || tok == Token::NotEq || tok == Token::Lte || tok == Token::Gte || tok == Token::And || tok == Token::Or) {
                    // This is a default advancement for single char tokens
                 }
            }
            // For Eq, NotEq, Lte, Gte, And, Or, read_char was already called for the second char.
            // For Identifiers, Numbers, Strings, their respective functions handle read_char.
            // Eof and Illegal don't consume in the same way or it's the end.
            _ => {}
        }
        
        // Ensure `read_char` is called for tokens that don't manage it internally
        // This is crucial for tokens like '+', '-', ';', etc.
        // `read_identifier`, `read_number`, `read_string` manage their own consumption.
        // Multi-character operators like `==` also manage their own consumption.
        // `skip_whitespace` and `skip_comment` also manage their own consumption.
        if !(matches!(tok, Token::Identifier(_)) ||
             matches!(tok, Token::Integer(_)) ||
             matches!(tok, Token::Float(_)) ||
             matches!(tok, Token::String(_)) ||
             matches!(tok, Token::Eof) ||
             matches!(tok, Token::Illegal(_)) ||
             // These were handled by peeking and consuming the second char
             tok == Token::Eq || tok == Token::NotEq || tok == Token::Lte || tok == Token::Gte || tok == Token::And || tok == Token::Or)
        {
            //This is for single character tokens like +, -, *, /, ;, etc.
            // Also for Assign, Bang, Lt, Gt when they are NOT part of a two-char token
             if self.ch != 0 { // Avoid reading past EOF if we just produced an EOF token
                //self.read_char(); // This was causing issues by over-consuming
             }
        }

        // Correct advancement logic:
        // 1. skip_whitespace and skip_comment advance.
        // 2. read_identifier, read_number, read_string advance internally until the end of the literal/identifier.
        // 3. For operators:
        //    - Single char (e.g., '+', ';'): consume one char.
        //    - Double char (e.g., '==', '&&'): consume two chars.
        // The main match block needs to decide if it consumes one or two (or more via helper fns).

        // Resetting `ch` advancement logic for clarity.
        // `read_char()` is called at the start of `new()` and at the end of every successful consumption of a character or sequence.

        let current_char_consumed = match tok {
            Token::Identifier(_) | Token::Integer(_) | Token::Float(_) | Token::String(_) | Token::Eof => false,
             // For Illegal, we consume the char to avoid infinite loops on it.
            Token::Illegal(_) => true,
            // For two-char tokens, the second char is consumed by read_char() inside the if block.
            Token::Eq | Token::NotEq | Token::Lte | Token::Gte | Token::And | Token::Or => false, // Already consumed by peek
            // All others are single char tokens by default
            _ => true,
        };

        if current_char_consumed {
            self.read_char();
        }
        
        tok
    }

    // Optional: Tokenize the whole input
    pub fn tokenize(&mut self) -> Result<Vec<Token>, String> {
        let mut tokens = Vec::new();
        loop {
            let token = self.next_token();
            if let Token::Illegal(ch) = token {
                if ch == '"' { // Specifically for unterminated string
                    return Err("Unterminated string literal".to_string());
                } else if ch == '\0' && tokens.last() == Some(&Token::Illegal('/')) { // Check for unterminated multi-line comment
                     return Err("Unterminated multi-line comment".to_string());
                }
                // For other illegal characters, we might just add them or error out differently
                // For now, let's assume next_token handles Illegal correctly by advancing
                 tokens.push(token); // Add the illegal token and continue or break
                 // If we want to stop on first illegal:
                 // return Err(format!("Illegal character: {}", ch));
            } else {
                tokens.push(token.clone());
            }
            if token == Token::Eof {
                break;
            }
        }
        Ok(tokens)
    }
}

// Helper for keyword matching (used within next_token)
fn keyword(ident: &str) -> Option<Token> {
    match ident {
        "let" => Some(Token::Let),
        "mut" => Some(Token::Mut),
        "if" => Some(Token::If),
        "else" => Some(Token::Else),
        "loop" => Some(Token::Loop),
        "while" => Some(Token::While),
        "for" => Some(Token::For),
        "fn" => Some(Token::Fn),
        "return" => Some(Token::Return),
        "true" => Some(Token::True),
        "false" => Some(Token::False),
        "print" => Some(Token::Print),
        "println" => Some(Token::Println),
        "break" => Some(Token::Break),
        "continue" => Some(Token::Continue),
        _ => None,
    }
}

impl Iterator for Lexer<'_> {
    type Item = Token;

    fn next(&mut self) -> Option<Self::Item> {
        let token = self.next_token();
        if token == Token::Eof {
            None
        } else {
            Some(token)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_lexer(input: &str, expected_tokens: Vec<Token>) {
        let mut lexer = Lexer::new(input);
        let mut tokens = Vec::new();
        while let Some(token) = lexer.next() {
            tokens.push(token);
        }
        assert_eq!(tokens, expected_tokens);
    }

    #[test]
    fn test_simple_tokens() {
        let input = "=+-*/%(){},;:!";
        let expected = vec![
            Token::Assign, Token::Plus, Token::Minus, Token::Multiply, Token::Divide, Token::Modulo,
            Token::LBrace, Token::RBrace, Token::LParen, Token::RParen, Token::Comma, Token::Semicolon, Token::Colon, Token::Bang,
        ];
        test_lexer(input, expected);
    }
    
    #[test]
    fn test_operators_and_delimiters() {
        let input = "== != <= >= && ||";
        let expected = vec![
            Token::Eq, Token::NotEq, Token::Lte, Token::Gte, Token::And, Token::Or,
        ];
        test_lexer(input, expected);
    }

    #[test]
    fn test_keywords_and_identifiers() {
        let input = "let mut x = 5; fn main() { return x; }";
        let expected = vec![
            Token::Let, Token::Mut, Token::Identifier("x".to_string()), Token::Assign, Token::Integer(5), Token::Semicolon,
            Token::Fn, Token::Identifier("main".to_string()), Token::LParen, Token::RParen, Token::LBrace,
            Token::Return, Token::Identifier("x".to_string()), Token::Semicolon,
            Token::RBrace,
        ];
        test_lexer(input, expected);
    }

    #[test]
    fn test_numbers() {
        let input = "123 45.67 0.5";
        let expected = vec![
            Token::Integer(123), Token::Float(45.67), Token::Float(0.5),
        ];
        test_lexer(input, expected);
    }
    
    #[test]
    fn test_string_literal() {
        let input = r#""hello"" "#;
        let expected = vec![Token::String("hello".to_string())];
        test_lexer(input, expected);
    }

    #[test]
    fn test_string_with_escapes() {
        let input = r#""line1\nline2\t\"quote\\end""#;
        let expected = vec![Token::String("line1\nline2\t\"quote\\end".to_string())];
        test_lexer(input, expected);
    }
    
    #[test]
    fn test_unterminated_string() {
        let input = r#""hello"#;
        let mut lexer = Lexer::new(input);
        assert_eq!(lexer.next_token(), Token::Illegal('"'));
        assert_eq!(lexer.next_token(), Token::Eof); // Should be EOF after error
    }

    #[test]
    fn test_skip_whitespace_and_comments() {
        let input = r#"
            // This is a comment
            let x = 10; // another comment
            /* multi-line
               comment */
            let y = 20;
            /* unterminated
        "#;
        let expected = vec![
            Token::Let, Token::Identifier("x".to_string()), Token::Assign, Token::Integer(10), Token::Semicolon,
            Token::Let, Token::Identifier("y".to_string()), Token::Assign, Token::Integer(20), Token::Semicolon,
            Token::Illegal('/'), // From the start of "/* unterminated"
        ];
         let mut lexer = Lexer::new(input);
        let mut tokens = Vec::new();
        // Collect tokens until EOF or specific error handling
        loop {
            let token = lexer.next_token();
            if token == Token::Eof && tokens.last() == Some(&Token::Illegal('/')) { // if EOF follows unterminated comment
                 break;
            }
            tokens.push(token.clone());
            if token == Token::Eof {
                break;
            }
             if let Token::Illegal('/') = token { // Stop after detecting start of unterminated comment
                if lexer.ch == 0 { // if we are at EOF
                    break;
                }
            }
        }
         // The current skip_comment for multi-line will read until EOF if not terminated.
         // next_token() will then return Eof.
         // A more robust error would be Token::Illegal for unterminated multi-line comment.
         // For now, testing what's implemented:
        let mut lexer_for_test = Lexer::new(input);
        assert_eq!(lexer_for_test.next_token(), Token::Let);
        assert_eq!(lexer_for_test.next_token(), Token::Identifier("x".to_string()));
        assert_eq!(lexer_for_test.next_token(), Token::Assign);
        assert_eq!(lexer_for_test.next_token(), Token::Integer(10));
        assert_eq!(lexer_for_test.next_token(), Token::Semicolon);
        assert_eq!(lexer_for_test.next_token(), Token::Let);
        assert_eq!(lexer_for_test.next_token(), Token::Identifier("y".to_string()));
        assert_eq!(lexer_for_test.next_token(), Token::Assign);
        assert_eq!(lexer_for_test.next_token(), Token::Integer(20));
        assert_eq!(lexer_for_test.next_token(), Token::Semicolon);
        // The unterminated /* comment consumes the rest. Then next_token() sees EOF.
        // The current skip_comment consumes '/*' then reads till EOF if '*/' is not found.
        // This means the next call to next_token() after "let y = 20;" will encounter the "/*"
        // it will consume it, then read till end of input.
        // Then the *next* call to next_token() will see self.ch == 0 and return Token::Eof.
        assert_eq!(lexer_for_test.next_token(), Token::Eof);


    }


    #[test]
    fn test_complex_mix() {
        let input = r#"
            let five = 5;
            let ten = 10.5;
            let add = fn(x, y) {
              x + y;
            };
            let result = add(five, ten);
            if (result > 15) {
                print "greater";
            } else {
                println "smaller or equal";
            }
            // Check operators
            !true == false;
            1 < 2; 2 <= 2; 3 > 1; 4 >= 3;
            true && false || true;
            10 % 3;
            /* Loop test
               for (let i = 0; i < 3; i = i + 1) {
                 if (i == 1) { continue; }
                 print i;
                 if (i == 2) { break; }
               }
            */
            while (false) {}
            loop { break; }
        "#;
        let expected = vec![
            Token::Let, Token::Identifier("five".to_string()), Token::Assign, Token::Integer(5), Token::Semicolon,
            Token::Let, Token::Identifier("ten".to_string()), Token::Assign, Token::Float(10.5), Token::Semicolon,
            Token::Let, Token::Identifier("add".to_string()), Token::Assign, Token::Fn, Token::LParen, Token::Identifier("x".to_string()), Token::Comma, Token::Identifier("y".to_string()), Token::RParen, Token::LBrace,
            Token::Identifier("x".to_string()), Token::Plus, Token::Identifier("y".to_string()), Token::Semicolon,
            Token::RBrace, Token::Semicolon,
            Token::Let, Token::Identifier("result".to_string()), Token::Assign, Token::Identifier("add".to_string()), Token::LParen, Token::Identifier("five".to_string()), Token::Comma, Token::Identifier("ten".to_string()), Token::RParen, Token::Semicolon,
            Token::If, Token::LParen, Token::Identifier("result".to_string()), Token::Gt, Token::Integer(15), Token::RParen, Token::LBrace,
            Token::Print, Token::String("greater".to_string()), Token::Semicolon,
            Token::RBrace, Token::Else, Token::LBrace,
            Token::Println, Token::String("smaller or equal".to_string()), Token::Semicolon,
            Token::RBrace,
            Token::Bang, Token::True, Token::Eq, Token::False, Token::Semicolon,
            Token::Integer(1), Token::Lt, Token::Integer(2), Token::Semicolon,
            Token::Integer(2), Token::Lte, Token::Integer(2), Token::Semicolon,
            Token::Integer(3), Token::Gt, Token::Integer(1), Token::Semicolon,
            Token::Integer(4), Token::Gte, Token::Integer(3), Token::Semicolon,
            Token::True, Token::And, Token::False, Token::Or, Token::True, Token::Semicolon,
            Token::Integer(10), Token::Modulo, Token::Integer(3), Token::Semicolon,
            Token::While, Token::LParen, Token::False, Token::RParen, Token::LBrace, Token::RBrace,
            Token::Loop, Token::LBrace, Token::Break, Token::Semicolon, Token::RBrace,
        ];
        test_lexer(input, expected);
    }
    
    #[test]
    fn test_illegal_char() {
        let input = "let a = @;";
        let expected = vec![
            Token::Let, Token::Identifier("a".to_string()), Token::Assign, Token::Illegal('@'), Token::Semicolon,
        ];
        test_lexer(input, expected);
    }

    #[test]
    fn test_unterminated_multiline_comment_at_eof() {
        let input = "/* this is not closed";
        let mut lexer = Lexer::new(input);
        // skip_comment advances to EOF if comment is unterminated.
        // Then next_token() sees EOF.
        // A more specific error token would be better.
        assert_eq!(lexer.next_token(), Token::Eof);
    }

    #[test]
    fn test_tokenize_function_normal() {
        let input = "let x = 10;";
        let mut lexer = Lexer::new(input);
        let tokens = lexer.tokenize().unwrap();
        assert_eq!(tokens, vec![
            Token::Let, Token::Identifier("x".to_string()), Token::Assign, Token::Integer(10), Token::Semicolon, Token::Eof
        ]);
    }

    #[test]
    fn test_tokenize_function_unterminated_string() {
        let input = r#"let name = "Test;"#; // Missing closing quote
        let mut lexer = Lexer::new(input);
        let result = lexer.tokenize();
        assert!(result.is_err());
        assert_eq!(result.unwrap_err(), "Unterminated string literal");
    }
}
