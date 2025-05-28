use crate::ast::*;
use crate::lexer::{Lexer, Token}; // Token is already imported

use std::collections::HashMap;
use std::sync::LazyLock; // Import LazyLock

#[derive(PartialEq, PartialOrd, Debug, Clone, Copy)]
enum Precedence {
    LOWEST,
    // ASSIGN, // Removed: Assignment is a statement, not an expression precedence.
    OR,          // ||
    AND,         // &&
    EQUALS,      // ==, !=
    LESSGREATER, // <, >, <=, >=
    SUM,         // +, -
    PRODUCT,     // *, /, %
    PREFIX,      // -X or !X
    CALL,        // myFunction(X)
}

// Assuming Precedence enum is defined above (it is)
// Assuming Token enum is imported correctly (it is via crate::lexer::Token)

static PRECEDENCES: LazyLock<HashMap<Token, Precedence>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    // The line for Token::Assign mapping to Precedence::ASSIGN is confirmed to be already commented out or absent.
    // m.insert(Token::Assign, Precedence::ASSIGN); // This line should remain commented or absent.
    m.insert(Token::Or, Precedence::OR);
    m.insert(Token::And, Precedence::AND);
    m.insert(Token::Eq, Precedence::EQUALS);
    m.insert(Token::NotEq, Precedence::EQUALS);
    m.insert(Token::Lt, Precedence::LESSGREATER);
    m.insert(Token::Lte, Precedence::LESSGREATER);
    m.insert(Token::Gt, Precedence::LESSGREATER);
    m.insert(Token::Gte, Precedence::LESSGREATER);
    m.insert(Token::Plus, Precedence::SUM);
    m.insert(Token::Minus, Precedence::SUM); // Also prefix, handled separately
    m.insert(Token::Multiply, Precedence::PRODUCT);
    m.insert(Token::Divide, Precedence::PRODUCT);
    m.insert(Token::Modulo, Precedence::PRODUCT);
    m.insert(Token::LParen, Precedence::CALL); // For call expressions like func()
    // Prefix operators like Token::Bang (!) and Token::Minus (-) for prefix are handled by their parsing functions,
    // not by infix precedence lookup here.
    m
});

fn token_precedence(token: &Token) -> Precedence {
    PRECEDENCES.get(token).cloned().unwrap_or(Precedence::LOWEST)
}

pub struct Parser<'a> {
    lexer: Lexer<'a>,
    current_token: Token,
    peek_token: Token,
    errors: Vec<String>,
}

impl<'a> Parser<'a> {
    pub fn new(lexer: Lexer<'a>) -> Self {
        let mut p = Parser {
            lexer,
            current_token: Token::Eof, // Placeholder
            peek_token: Token::Eof,    // Placeholder
            errors: Vec::new(),
        };
        p.next_token();
        p.next_token();
        p
    }

    fn next_token(&mut self) {
        self.current_token = self.peek_token.clone();
        self.peek_token = self.lexer.next_token();
    }

    fn current_token_is(&self, t: &Token) -> bool {
        &self.current_token == t
    }

    fn peek_token_is(&self, t: &Token) -> bool {
        &self.peek_token == t
    }

    fn expect_peek(&mut self, t: Token) -> bool {
        if self.peek_token_is(&t) {
            self.next_token();
            true
        } else {
            self.peek_error(&t);
            false
        }
    }

    fn peek_error(&mut self, t: &Token) {
        let msg = format!(
            "expected next token to be {:?}, got {:?} instead. (current: {:?})",
            t, self.peek_token, self.current_token
        );
        self.errors.push(msg);
    }
    
    fn current_precedence(&self) -> Precedence {
        token_precedence(&self.current_token)
    }

    fn peek_precedence(&self) -> Precedence {
        token_precedence(&self.peek_token)
    }

    pub fn parse_program(&mut self) -> Result<Program, Vec<String>> {
        let mut program = Program { statements: Vec::new() };

        while !self.current_token_is(&Token::Eof) {
            match self.parse_statement() {
                Some(statement) => program.statements.push(statement),
                None => { 
                    // If parse_statement returns None, it means a severe error occurred,
                    // or it was an empty statement (e.g. just ';').
                    // Errors should have been logged. We can try to recover by advancing.
                    // However, parse_statement itself should advance tokens.
                    // This path might indicate we're not at EOF but can't parse a statement.
                }
            }
            // parse_statement is responsible for consuming all tokens related to the statement,
            // including the trailing semicolon if applicable.
            // So, we should *not* call self.next_token() here IF parse_statement does its job.
            // Let's adjust parse_statement to always advance to the next token that begins a new statement.
            // For now, the original loop structure where parse_program calls next_token is common:
             self.next_token(); // Consume the last token of the statement (e.g., ';', '}')
                                // or the token that caused parse_statement to return None.
        }

        if self.errors.is_empty() {
            Ok(program)
        } else {
            Err(self.errors.clone())
        }
    }

    pub fn errors(&self) -> &Vec<String> {
        &self.errors
    }

    fn parse_statement(&mut self) -> Option<Statement> {
        let stmt = match self.current_token {
            Token::Let | Token::Mut => self.parse_let_or_mut_statement(), // Updated dispatch
            Token::If => self.parse_if_statement(),
            Token::Loop => self.parse_loop_statement(),
            Token::While => self.parse_while_statement(),
            Token::For => self.parse_for_statement(),
            Token::Print | Token::Println => self.parse_print_statement(),
            Token::Break => self.parse_break_statement(),
            Token::Continue => self.parse_continue_statement(),
            Token::Identifier(ident_name) => { // Capture ident_name directly
                if self.peek_token_is(&Token::Assign) {
                    // This is an Assignment statement
                    let name = ident_name.clone(); // Clone the captured name
                    self.next_token(); // Consume Identifier, current_token is now Assign
                    self.next_token(); // Consume Assign, current_token is now the start of the RHS expression
                    
                    match self.parse_expression(Precedence::LOWEST) {
                        Some(value_expr) => Some(Statement::Assignment { name, value_expr }),
                        None => {
                            self.errors.push(format!("Expected expression after '=' for assignment to '{}'", name));
                            None
                        }
                    }
                } else {
                    // Not an assignment, so it's a regular expression statement (e.g. function call or just an expression)
                    // parse_expression_statement expects current_token to be the start of the expression.
                    // Since current_token is already the Identifier, this is correct.
                    self.parse_expression_statement()
                }
            }
            Token::Semicolon => { // Empty statement
                // self.next_token(); // The main loop will advance.
                return None; 
            }
            // If it's none of the above, try to parse it as an expression statement.
            // This includes cases like literals, prefix expressions, etc., starting a statement.
            _ => self.parse_expression_statement(),
        };

        // After parsing the core statement, check for an optional semicolon.
        // This applies to statements that are not block-based (if, while, for, loop end with '}')
        // and are not control flow keywords like break/continue that might not need them.
        // Specifically, LetDecl, Assignment, ExprStatement, Print often have semicolons.
        match stmt {
            Some(Statement::LetDecl{..}) | 
            Some(Statement::Assignment{..}) | 
            Some(Statement::ExprStatement{..}) | 
            Some(Statement::Print{..}) |
            Some(Statement::Break) | // Break and Continue can also be optionally terminated
            Some(Statement::Continue) => {
                if self.peek_token_is(&Token::Semicolon) {
                    self.next_token(); // Consume the optional semicolon. current_token is now the semicolon.
                }
            }
            // For block statements (If, While, For, Loop), they end with '}', no semicolon needed after the '}'.
            // None is for empty semicolon statements, already handled.
            _ => {} 
        }
        stmt
    }

    // Renamed from parse_let_statement to handle both 'let' and 'mut'
    fn parse_let_or_mut_statement(&mut self) -> Option<Statement> {
        let is_mutable = match self.current_token {
            Token::Mut => true,
            Token::Let => false,
            _ => { // Should not be called if current_token is not Let or Mut
                self.errors.push(format!("parse_let_or_mut_statement called with unexpected token {:?}", self.current_token));
                return None;
            }
        };
        // The dispatcher (parse_statement) has already identified 'let' or 'mut'.
        // We now expect an identifier for the variable name.
        // So, we look at peek_token.

        if !matches!(self.peek_token, Token::Identifier(_)) {
            self.peek_error(&Token::Identifier("IDENTIFIER_NAME".to_string()));
            return None;
        }
        self.next_token(); // Consume the identifier token
        let name = match &self.current_token {
            Token::Identifier(n) => n.clone(),
            _ => return None, // Unlikely due to the check above
        };

        // Optional Type Annotation
        let mut type_annotation: Option<String> = None;
        if self.peek_token_is(&Token::Colon) {
            self.next_token(); // Consume ':'
            
            if !matches!(self.peek_token, Token::Identifier(_)) {
                 self.peek_error(&Token::Identifier("TYPE_NAME".to_string()));
                 // Decide if this is a fatal error for the statement or if we can proceed without type_ann
                 // For now, let's make it fatal for the type annotation part.
                 return None; 
            }
            self.next_token(); // Consume the type identifier
            match &self.current_token {
                 Token::Identifier(type_name_str) => {
                    type_annotation = Some(type_name_str.clone());
                }
                _ => return None, // Unlikely
            }
        }

        // Assignment Operator
        if !self.expect_peek(Token::Assign) {
            // peek_error already called by expect_peek
            return None;
        }
        // current_token is now '='
        self.next_token(); // Consume '=', move to the start of the expression

        // Value Expression
        let value_expr = match self.parse_expression(Precedence::LOWEST) {
            Some(expr) => expr,
            None => {
                self.errors.push(format!("Expected expression after '=' for variable '{}'", name));
                return None;
            }
        };
        
        // Optional semicolon is handled by the main parse_statement function's suffix logic.

        Some(Statement::LetDecl {
            name,
            mutable: is_mutable,
            type_ann: type_annotation,
            value_expr,
        })
    }

    // Removed parse_assignment_statement as its logic is now in parse_statement's Identifier arm.

    fn parse_expression_statement(&mut self) -> Option<Statement> {
        // current_token is the beginning of an expression
        let expr = self.parse_expression(Precedence::LOWEST)?;
        // Semicolon is now optional, will be handled by parse_statement's suffix check.
        Some(Statement::ExprStatement { expr })
    }

    fn parse_block_statement(&mut self) -> Option<Block> {
        // Expects self.current_token to be LBrace (ensured by caller using expect_peek).
        // No, the plan was for THIS function to consume it, or rather, expect_peek in caller consumes it.
        // Let's stick to: caller uses expect_peek, so current_token *is* LBrace when this is called.
        // Then this function consumes it to move to the content of the block.

        // self.next_token(); // Consume '{', current_token is now the first token of the block or '}'

        let mut statements = Vec::new();
        // Current token is LBRACE. Consume it and move to the first statement or RBRACE
        self.next_token(); 

        while !self.current_token_is(&Token::RBrace) && !self.current_token_is(&Token::Eof) {
            // parse_statement parses one statement.
            // The optional semicolon logic is handled within parse_statement itself.
            // The main loop in parse_program calls next_token() after parse_statement().
            // We replicate that pattern here for statements within a block.
            match self.parse_statement() {
                Some(statement) => statements.push(statement),
                None => {
                    // If parse_statement returns None (e.g., for an empty ';' or a parsing error for that statement),
                    // we still need to advance to avoid getting stuck, unless it's already EOF or RBrace.
                }
            }
            // Crucially, advance the token to prepare for the next statement or the end of the block.
            // This is similar to how parse_program's loop works.
            // If current_token is already RBrace or Eof, the loop condition will handle it.
            self.next_token();
        }

        if !self.current_token_is(&Token::RBrace) {
            self.errors.push(format!("Unterminated block: expected '}}', got {:?}", self.current_token));
            return None;
        }
        // Do NOT consume the RBrace here.
        // The current_token is now RBrace. The caller (e.g. parse_if_statement)
        // will finish, and the main parse_program loop's next_token() will consume the RBrace.
        Some(Block { statements })
    }

    fn parse_if_statement(&mut self) -> Option<Statement> {
        // current_token is If. No opening parenthesis expected.
        self.next_token(); // Consume 'if', current is start of condition

        let condition = self.parse_expression(Precedence::LOWEST)?;
        // After parse_expression, current_token is the last token of the condition.
        // We now expect LBrace for the 'then' block.

        if !self.expect_peek(Token::LBrace) { // Expects peek to be LBrace, then consumes it.
            self.errors.push(format!("Expected '{{' after if condition, got {:?}", self.current_token));
            return None;
        }
        // current_token is now LBrace
        let then_block = self.parse_block_statement()?;
        // after parse_block_statement, current_token is '}'

        let mut else_if_blocks = Vec::new();
        let mut else_block = None;

        // current_token is '}' from then_block. Peek for 'else'.
        while self.peek_token_is(&Token::Else) {
            self.next_token(); // Consume '}' (from then_block or previous else-if block), current is 'else'
            
            if self.peek_token_is(&Token::If) { // This is 'else if'
                self.next_token(); // consume 'else', current is 'if'
                self.next_token(); // consume 'if', current is start of else-if-condition
                
                let else_if_condition = self.parse_expression(Precedence::LOWEST)?;
                
                if !self.expect_peek(Token::LBrace) {
                    self.errors.push(format!("Expected '{{' after else if condition, got {:?}", self.current_token));
                    return None;
                }
                let else_if_then_block = self.parse_block_statement()?;
                else_if_blocks.push((else_if_condition, else_if_then_block));
            } else { // This is an 'else' block
                if !self.expect_peek(Token::LBrace) { 
                    self.errors.push(format!("Expected '{{' for else block, got {:?}", self.current_token));
                    return None; 
                }
                else_block = Some(self.parse_block_statement()?); 
                break; 
            }
        }
        Some(Statement::If { condition, then_block, else_if_blocks, else_block })
    }
    
    fn parse_loop_statement(&mut self) -> Option<Statement> {
        if !self.expect_peek(Token::LBrace) { 
            self.errors.push(format!("Expected '{{' after 'loop', got {:?}, peek: {:?}", self.current_token, self.peek_token));
            return None;
        }
        let body_block = self.parse_block_statement()?; 
        Some(Statement::Loop { body_block })
    }

    fn parse_while_statement(&mut self) -> Option<Statement> {
        // current_token is While. No opening parenthesis expected.
        self.next_token(); // Consume 'while', current is start of condition

        let condition = self.parse_expression(Precedence::LOWEST)?;
        // After parse_expression, current_token is the last token of the condition.
        // We now expect LBrace for the loop body.

        if !self.expect_peek(Token::LBrace) { // Expects peek to be LBrace, then consumes it.
            self.errors.push(format!("Expected '{{' after while condition, got {:?}", self.current_token));
            return None;
        }
        // current_token is now LBrace
        let body_block = self.parse_block_statement()?;
        Some(Statement::While { condition, body_block })
    }

    fn parse_for_statement(&mut self) -> Option<Statement> {
        // current_token is For. No opening parenthesis expected.
        self.next_token(); // Consume 'for', current is start of initializer or first ';'
        
        // 1. Initializer
        let initializer = if self.current_token_is(&Token::Semicolon) {
            None 
        } else {
            // parse_statement itself will handle optional semicolons for the initializer statement.
            // The semicolon *for the for-loop structure* is mandatory here.
            self.parse_statement() 
        };

        if !self.current_token_is(&Token::Semicolon) {
            // This error means the initializer (if present) didn't end where we expected,
            // or if it was None, the first token wasn't a semicolon.
            // If parse_statement consumed an optional semicolon, current_token would be that semicolon.
            // If it did not (optional semicolon was absent), current_token is last token of init.
            // So, we must advance if current is not already the semicolon.
            if self.peek_token_is(&Token::Semicolon) { // If init was `let x = 1` (no semi), current is 1, peek is ;
                self.next_token(); // current is now ;
            } else if !self.current_token_is(&Token::Semicolon) { // If init was `let x = 1;` current is ;, this is false. If `let x = 1` (no semi) and next is not semi.
                 self.errors.push(format!("Expected ';' after for loop initializer, got {:?} (peek: {:?})", self.current_token, self.peek_token));
                 return None;
            }
        }
        self.next_token(); // Consume ';' after initializer, current is start of condition or second ';'

        let condition = if self.current_token_is(&Token::Semicolon) {
            None 
        } else {
            self.parse_expression(Precedence::LOWEST)
        };
        
        if !self.current_token_is(&Token::Semicolon) {
            if self.peek_token_is(&Token::Semicolon) { // If cond was `x < 1` (no semi), current is 1, peek is ;
                self.next_token(); // current is now ;
            } else if !self.current_token_is(&Token::Semicolon) {
                 self.errors.push(format!("Expected ';' after for loop condition, got {:?} (peek: {:?})", self.current_token, self.peek_token));
                 return None;
            }
        }
        self.next_token(); // Consume ';' after condition, current is start of increment or ')'

        let increment = if self.current_token_is(&Token::RParen) {
            None 
        } else {
            self.parse_expression(Precedence::LOWEST)
        };
        
        // After parsing increment, current_token is the last token of the increment expression.
        // After parsing increment, current_token is the last token of the increment expression,
        // OR it's the token that made us decide there's no increment (e.g., LBrace).
        // No closing parenthesis expected. We directly expect LBrace for the body.
        
        if !self.current_token_is(&Token::LBrace) {
            // If current_token is not LBrace after parsing increment (or deciding there's no increment),
            // it's an error. parse_expression for increment should leave us on its last token.
            // So we need to expect_peek for LBrace.
            if !self.expect_peek(Token::LBrace) {
                 self.errors.push(format!("Expected '{{' for for-loop body, got {:?} (peek: {:?})", self.current_token, self.peek_token));
                 return None;
            }
        }
        // current_token is now LBrace
        let body_block = self.parse_block_statement()?;
        Some(Statement::For { initializer: initializer.map(Box::new), condition, increment, body_block })
    }

    fn parse_print_statement(&mut self) -> Option<Statement> {
        let newline = self.current_token_is(&Token::Println);

        if !self.expect_peek(Token::LParen) { return None; } 
        self.next_token(); 

        let expr = self.parse_expression(Precedence::LOWEST)?; 

        if !self.expect_peek(Token::RParen) { return None; } 
        // Semicolon is now optional, will be handled by parse_statement's suffix check.
        Some(Statement::Print { expr, newline })
    }
    
    fn parse_break_statement(&mut self) -> Option<Statement> {
        // Semicolon is now optional, will be handled by parse_statement's suffix check.
        Some(Statement::Break)
    }

    fn parse_continue_statement(&mut self) -> Option<Statement> {
        // Semicolon is now optional, will be handled by parse_statement's suffix check.
        Some(Statement::Continue)
    }
    
    // `parse_return_statement` would be here if `return` keyword was part of the language.

    // === Expression Parsing (Pratt Parser) ===

    fn parse_expression(&mut self, precedence: Precedence) -> Option<Expr> {
        // Prefix part
        let mut left_expr_opt = match self.current_token {
            Token::Identifier(_) => self.parse_identifier(),
            Token::Integer(_) => self.parse_integer_literal(),
            Token::Float(_) => self.parse_float_literal(),
            Token::String(_) => self.parse_string_literal(),
            Token::True | Token::False => self.parse_boolean_literal(),
            Token::Bang | Token::Minus => self.parse_prefix_expression(), // Note: Minus is also infix
            Token::LParen => self.parse_grouped_expression(),
            ref tok if is_prefix_operator(tok) => self.parse_prefix_expression(), // General prefix
            _ => {
                self.errors.push(format!("No prefix parse function for {:?} found. Peek: {:?}", self.current_token, self.peek_token));
                return None;
            }
        };

        // Infix part
        // After prefix parsing, current_token is the *last* token of the prefix expression.
        // We need to look at peek_token for the infix operator.
        while !self.peek_token_is(&Token::Semicolon) && precedence < self.peek_precedence() {
            let peeked_token = self.peek_token.clone();
            if !is_infix_operator(&peeked_token) && peeked_token != Token::LParen /* for call */ {
                return left_expr_opt;
            }

            self.next_token(); // Consume the prefix expression's last token, current_token is now the infix operator or '(' for call
            
            left_expr_opt = match self.current_token {
                // Binary operators
                Token::Plus | Token::Minus | Token::Multiply | Token::Divide | Token::Modulo |
                Token::Eq | Token::NotEq | Token::Lt | Token::Lte | Token::Gt | Token::Gte |
                Token::And | Token::Or => {
                    self.parse_infix_expression(left_expr_opt?)
                }
                Token::LParen => { // Call expression like identifier(args)
                    self.parse_call_expression(left_expr_opt?)
                }
                _ => {
                    // This should not be reached if is_infix_operator and precedence checks are correct
                    return left_expr_opt; 
                }
            };
        }
        left_expr_opt
    }
    
    fn parse_identifier(&mut self) -> Option<Expr> {
        // current_token is Identifier
        match &self.current_token {
            Token::Identifier(name) => Some(Expr::Identifier(name.clone())),
            _ => None, 
        }
    }

    fn parse_integer_literal(&mut self) -> Option<Expr> {
        // current_token is Integer
        match self.current_token {
            Token::Integer(val) => Some(Expr::Integer(val)),
            _ => None,
        }
    }

    fn parse_float_literal(&mut self) -> Option<Expr> {
        // current_token is Token::Float(String)
        match &self.current_token {
            Token::Float(s_val) => {
                match s_val.parse::<f64>() {
                    Ok(f_val) => Some(Expr::Float(f_val)),
                    Err(_) => {
                        self.errors.push(format!("Could not parse float string '{}'", s_val));
                        None
                    }
                }
            }
            _ => { // Should not happen if called on Token::Float
                self.errors.push(format!("Expected float literal, got {:?}", self.current_token));
                None
            }
        }
    }
    
    fn parse_string_literal(&mut self) -> Option<Expr> {
        // current_token is String
        match &self.current_token {
            Token::String(val) => Some(Expr::StringLiteral(val.clone())),
            _ => None,
        }
    }

    fn parse_boolean_literal(&mut self) -> Option<Expr> {
        // current_token is True or False
        match self.current_token {
            Token::True => Some(Expr::Boolean(true)),
            Token::False => Some(Expr::Boolean(false)),
            _ => None,
        }
    }

    fn parse_prefix_expression(&mut self) -> Option<Expr> {
        // current_token is the prefix operator (e.g., !, -)
        let operator_token = self.current_token.clone();
        let op = match operator_token {
            Token::Bang => UnaryOperator::Not,
            Token::Minus => UnaryOperator::Negate,
            _ => {
                self.errors.push(format!("Unknown prefix operator: {:?}", operator_token));
                return None;
            }
        };
        self.next_token(); // Consume prefix operator, current_token is now start of operand
        let expr = self.parse_expression(Precedence::PREFIX)?;
        // After parse_expression, current_token is the last token of the operand.
        Some(Expr::UnaryOp { op, expr: Box::new(expr) })
    }

    fn parse_infix_expression(&mut self, left: Expr) -> Option<Expr> {
        // current_token is the infix operator (e.g. +, ==)
        let operator_token = self.current_token.clone();
        let op = match operator_token {
            Token::Plus => BinaryOperator::Plus,
            Token::Minus => BinaryOperator::Minus,
            Token::Multiply => BinaryOperator::Multiply,
            Token::Divide => BinaryOperator::Divide,
            Token::Modulo => BinaryOperator::Modulo,
            Token::Eq => BinaryOperator::Eq,
            Token::NotEq => BinaryOperator::NotEq,
            Token::Lt => BinaryOperator::Lt,
            Token::Lte => BinaryOperator::Lte,
            Token::Gt => BinaryOperator::Gt,
            Token::Gte => BinaryOperator::Gte,
            Token::And => BinaryOperator::And,
            Token::Or => BinaryOperator::Or,
            _ => {
                self.errors.push(format!("Unknown infix operator: {:?}", operator_token));
                return None;
            }
        };
        let precedence = self.current_precedence();
        self.next_token(); // Consume infix operator, current_token is now start of right operand
        let right = self.parse_expression(precedence)?;
        // After parse_expression, current_token is the last token of the right operand.
        Some(Expr::BinaryOp { left: Box::new(left), op, right: Box::new(right) })
    }
    
    fn parse_grouped_expression(&mut self) -> Option<Expr> {
        // current_token is LParen
        self.next_token(); // Consume '(', current_token is start of inner expression
        let expr = self.parse_expression(Precedence::LOWEST);
        // After parse_expression, current_token is last token of inner expression.
        // Expect peek_token to be RParen.
        if !self.expect_peek(Token::RParen) { // Consumes RParen, current_token is now RParen
            return None; 
        }
        // current_token is now RParen.
        expr
    }

    fn parse_call_expression(&mut self, function_identifier_expr: Expr) -> Option<Expr> {
        // `function_identifier_expr` is the expression for the function name (e.g. Identifier("add"))
        // `current_token` is LParen, consumed from parse_expression's infix loop.
        let callee = match function_identifier_expr {
            Expr::Identifier(name) => name,
            _ => {
                self.errors.push(format!("Expected function name (identifier) for call, got {:?}", function_identifier_expr));
                return None;
            }
        };

        let mut args = Vec::new();
        if self.peek_token_is(&Token::RParen) { // No arguments: add()
            self.next_token(); // Consume ')', current_token is now ')'
        } else {
            // Has arguments
            self.next_token(); // Consume '(', current_token is start of first argument
            args.push(self.parse_expression(Precedence::LOWEST)?);
            // After parse_expression, current_token is last token of first argument.

            while self.peek_token_is(&Token::Comma) {
                self.next_token(); // Consume argument's last token, current is now ','
                self.next_token(); // Consume ',', current_token is start of next argument
                args.push(self.parse_expression(Precedence::LOWEST)?);
                // current_token is last token of this argument.
            }
            // Expect ')'
            if !self.expect_peek(Token::RParen) { // Consumes ')', current_token is now ')'
                return None; 
            }
        }
        // current_token is now RParen.
        Some(Expr::Call { callee, args })
    }
}

// Helper function to identify tokens that can start a prefix expression
fn is_prefix_operator(token: &Token) -> bool {
    matches!(token, Token::Bang | Token::Minus)
}

// Helper function to identify tokens that can be infix operators
fn is_infix_operator(token: &Token) -> bool {
    matches!(token, Token::Plus | Token::Minus | Token::Multiply | Token::Divide | Token::Modulo |
                    Token::Eq | Token::NotEq | Token::Lt | Token::Lte | Token::Gt | Token::Gte |
                    Token::And | Token::Or)
}


#[cfg(test)]
mod tests {
    use super::*;
    use crate::lexer::Lexer;

    // Helper to run parser and check for errors and statement counts
    fn run_parser_test(input: &str, expected_stmts: usize, expected_errors: usize) -> Program {
        let l = Lexer::new(input);
        let mut p = Parser::new(l);
        let program_result = p.parse_program();

        if expected_errors > 0 {
            assert!(program_result.is_err(), "Expected parsing errors, but got Ok for input: '{}'", input);
            let errors = program_result.unwrap_err();
            assert_eq!(errors.len(), expected_errors, "Wrong number of parsing errors for input: '{}'. Got: {:?}, Expected: {}", input, errors, expected_errors);
            return Program { statements: vec![] }; // Dummy program for error cases
        } else {
            assert!(program_result.is_ok(), "Expected successful parse for input: '{}', but got errors: {:?}", input, program_result.unwrap_err());
            let program = program_result.unwrap();
            assert_eq!(program.statements.len(), expected_stmts, "Wrong number of statements for input: '{}'. Got: {}, Expected: {}", input, program.statements.len(), expected_stmts);
            assert!(p.errors.is_empty(), "Parser reported errors unexpectedly for input: '{}': {:?}", input, p.errors); // Should be caught by program_result.is_ok()
            program
        }
    }

    #[test]
    fn test_let_statements() {
        let tests = vec![
            ("let x = 5;", "x", None, false, Expr::Integer(5)),
            ("let y = 10.5", "y", None, false, Expr::Float(10.5)),
            ("let z = true", "z", None, false, Expr::Boolean(true)),
            ("let s = \"hello\";", "s", None, false, Expr::StringLiteral("hello".to_string())),
        ];

        for (input, name, type_ann_opt, mutable_expected, value_expr_expected) in tests {
            let program = run_parser_test(input, 1, 0);
            match &program.statements[0] {
                Statement::LetDecl { name: n, type_ann, mutable, value_expr } => {
                    assert_eq!(n, name);
                    assert_eq!(*type_ann, type_ann_opt.map(String::from));
                    assert_eq!(*mutable, mutable_expected);
                    assert_eq!(*value_expr, value_expr_expected);
                }
                _ => panic!("Expected LetDecl, got {:?}", program.statements[0]),
            }
        }
    }

    #[test]
    fn test_mut_statements() {
        let tests = vec![
            ("mut x = 5;", "x", None, true, Expr::Integer(5)),
            ("mut y = 10.5", "y", None, true, Expr::Float(10.5)),
            ("mut z: bool = true", "z", Some("bool"), true, Expr::Boolean(true)),
            ("mut s: string = \"hello\";", "s", Some("string"), true, Expr::StringLiteral("hello".to_string())),
        ];

        for (input, name, type_ann_opt, mutable_expected, value_expr_expected) in tests {
            let program = run_parser_test(input, 1, 0);
            match &program.statements[0] {
                Statement::LetDecl { name: n, type_ann, mutable, value_expr } => {
                    assert_eq!(n, name);
                    assert_eq!(*type_ann, type_ann_opt.map(String::from));
                    assert_eq!(*mutable, mutable_expected);
                    assert_eq!(*value_expr, value_expr_expected);
                }
                _ => panic!("Expected LetDecl (for mut), got {:?}", program.statements[0]),
            }
        }
    }
    
    #[test]
    fn test_let_statement_with_type() {
        let program = run_parser_test("let x: int = 5", 1, 0);
         match &program.statements[0] {
            Statement::LetDecl { name, type_ann, mutable, value_expr } => {
                assert_eq!(name, "x");
                assert_eq!(*type_ann, Some("int".to_string()));
                assert!(!*mutable);
                assert_eq!(*value_expr, Expr::Integer(5));
            }
            _ => panic!("Expected LetDecl"),
        }

        let program_float = run_parser_test("let y: float = 3.14;", 1, 0);
        match &program_float.statements[0] {
            Statement::LetDecl { name, type_ann, mutable, value_expr } => {
                 assert_eq!(name, "y");
                 assert_eq!(*type_ann, Some("float".to_string()));
                 assert!(!*mutable);
                 assert_eq!(*value_expr, Expr::Float(3.14_f64));
            }
            _ => panic!("Expected LetDecl"),
        }
    }
    
    #[test]
    fn test_let_statement_errors() {
        run_parser_test("let = 5;", 0, 1); // Missing identifier
        run_parser_test("let x 5;", 0, 1); // Missing assign
        run_parser_test("let x = ;", 0, 1); // Missing value expression
        run_parser_test("mut y : int = ;", 0, 1); // Missing value expr for mut with type
        run_parser_test("let z : = 10;", 0, 1); // Missing type name after colon
        run_parser_test("mut w : float 20.0;", 0, 1); // Missing assign for mut with type
    }


    #[test]
    fn test_assignment_statement() {
        let tests = vec![
            ("val = 100 + 20", "val", Expr::BinaryOp { 
                left: Box::new(Expr::Integer(100)), 
                op: BinaryOperator::Plus, 
                right: Box::new(Expr::Integer(20)) 
            }),
            ("another = \"text\";", "another", Expr::StringLiteral("text".to_string())),
            ("flag = true", "flag", Expr::Boolean(true)),
        ];
        for (input, expected_name, expected_expr) in tests {
            let program = run_parser_test(input, 1, 0);
            match &program.statements[0] {
                Statement::Assignment { name, value_expr } => {
                    assert_eq!(name, expected_name);
                    assert_eq!(*value_expr, expected_expr);
                }
                _ => panic!("Expected Assignment statement for input: {}", input),
            }
        }
        
        let full_input = "val = 1\n a = val";
        let program_seq = run_parser_test(full_input, 2, 0);
        assert!(matches!(&program_seq.statements[0], Statement::Assignment{..}));
        assert!(matches!(&program_seq.statements[1], Statement::Assignment{..}));
    }

    #[test]
    fn test_assignment_errors() {
        run_parser_test("x = ;", 0, 1); // Missing expression after =
        run_parser_test("x = ", 0, 1); // Missing expression, ends with EOF
        // Note: "let x;" is not an assignment error, it's a LetDecl error if '=' is missing.
        // Assignment errors primarily revolve around the RHS.
    }
    
    #[test]
    fn test_print_statements() {
        run_parser_test("print(x)", 1, 0);
        run_parser_test("println(y + 2);", 1, 0);
        let full_input = "print(1)\nprintln(2)";
        run_parser_test(full_input, 2, 0);
    }

    #[test]
    fn test_expression_statement_literals() {
        run_parser_test("5", 1, 0);
        run_parser_test("true;", 1, 0);
        run_parser_test("\"test_string\"", 1, 0);
        let program = run_parser_test("3.14", 1, 0); // Test float literal expression
        assert_eq!(program.statements[0], Statement::ExprStatement { expr: Expr::Float(3.14_f64) });
        let full_input = "3.14\nfalse";
        run_parser_test(full_input, 2, 0);
    }
    
    #[test]
    fn test_prefix_expressions() {
        run_parser_test("!true", 1, 0);
        let program = run_parser_test("-15.5", 1, 0); // Test prefix float
         match &program.statements[0] {
            Statement::ExprStatement{expr: Expr::UnaryOp{op: UnaryOperator::Negate, expr: inner_expr}} => {
                assert_eq!(**inner_expr, Expr::Float(15.5_f64));
            }
            _ => panic!("Not a prefix negation of float. Got {:?}", program.statements[0]),
        }
    }

    #[test]
    fn test_infix_expressions_simple_arithmetic() {
        run_parser_test("5 + 5", 1, 0);
        let program = run_parser_test("10 - 2.0", 1, 0); // Test infix with float
        match &program.statements[0] {
            Statement::ExprStatement{expr: Expr::BinaryOp{left, op, right}} => {
                 assert_eq!(**left, Expr::Integer(10));
                 assert_eq!(*op, BinaryOperator::Minus);
                 assert_eq!(**right, Expr::Float(2.0_f64));
            }
            _ => panic!("Not a binary op with float. Got {:?}", program.statements[0]),
        }
        run_parser_test("3 * 8", 1, 0);
    }
    
    #[test]
    fn test_operator_precedence_parsing() {
        // Semicolons are optional, so we can test the raw expressions.
        // Each of these should parse as a single expression statement.
        let tests = vec![
            "-a * b", 
            "!-a",
            "a + b + c",
            "a + b / c",
            "3 > 5 == false",
            "(1 + 2) * 3",
            "a + add(b * c) + d",
        ];
        for input in tests {
            run_parser_test(input, 1, 0); // Expect 1 expression statement, 0 errors
        }
        run_parser_test("let val = 1 + 2 * 3", 1, 0); // Let statement with precedence
    }
    
    #[test]
    fn test_call_expression_parsing() {
        let program = run_parser_test("myFunction(arg1, 2.5, arg3 + 4)", 1, 0);
         match &program.statements[0] {
            Statement::ExprStatement{ expr: Expr::Call{callee, args} } => {
                assert_eq!(callee, "myFunction");
                assert_eq!(args.len(), 3);
                assert!(matches!(args[0], Expr::Identifier(id) if id == "arg1"));
                assert_eq!(args[1], Expr::Float(2.5_f64)); // Check float arg
                assert!(matches!(args[2], Expr::BinaryOp{op: BinaryOperator::Plus, ..}));
            }
            _ => panic!("Not a call expression statement. Got {:?}", program.statements[0])
        }
        run_parser_test("anotherCall();", 1, 0);
    }
    
    #[test]
    fn test_parse_malformed_float_string() {
        // This test assumes the lexer might produce a Float(String) that is not a valid f64.
        // e.g. if lexer's read_number was less strict.
        // For now, our lexer's read_number for floats is already quite strict (digits.digits).
        // But if it were "1.2.3", this test would be relevant.
        // Let's simulate a bad token that the lexer *shouldn't* produce but tests parser robustness.
        
        // To test this properly, we'd need to manually create a Parser with a custom token stream,
        // or modify the lexer to produce such a token (which is counter-productive).
        // The current `run_parser_test` uses the actual lexer.
        // So, we'll rely on the fact that `s_val.parse::<f64>()` handles errors.
        // An input like "let x = 1.2.3;" would be caught by the lexer producing something like:
        // Let, Ident("x"), Assign, Float("1.2"), Illegal('.'), Integer(3), Semicolon
        // So the parser would likely error out before even calling parse_float_literal on "1.2.3".
        
        // If we force a Token::Float("not-a-float") into the parser:
        let mut lexer = Lexer::new(""); // Dummy lexer
        let mut parser = Parser::new(lexer);
        parser.current_token = Token::Float("not-a-float".to_string()); // Manually set current token
        
        let expr_result = parser.parse_float_literal();
        assert!(expr_result.is_none());
        assert_eq!(parser.errors.len(), 1);
        assert!(parser.errors[0].contains("Could not parse float string 'not-a-float'"));
    }

    #[test]
    fn test_if_statement_no_else() {
        let input = "if x < y { x = 1\n print(x) }"; 
        let program = run_parser_test(input, 1, 0);
        match &program.statements[0] {
            Statement::If { condition, then_block, else_if_blocks, else_block } => {
                assert!(matches!(condition, Expr::BinaryOp {op: BinaryOperator::Lt, ..}));
                assert_eq!(then_block.statements.len(), 2, "Then block should have 2 statements");
                assert!(else_if_blocks.is_empty());
                assert!(else_block.is_none());
            }
            _ => panic!("Not an if statement. Got {:?}", program.statements[0])
        }
    }

    #[test]
    fn test_if_empty_block() {
        let input = "if x {}";
        let program = run_parser_test(input, 1, 0);
        match &program.statements[0] {
            Statement::If { then_block, .. } => {
                assert!(then_block.statements.is_empty(), "Then block should be empty");
            }
            _ => panic!("Not an if statement"),
        }
    }
    
    #[test]
    fn test_if_else_statement_parsing() {
        let input = "if x > y { 1 } else { 0; }"; 
        let program = run_parser_test(input, 1, 0);
         match &program.statements[0] {
            Statement::If { else_block, .. } => {
                assert!(else_block.is_some());
                assert_eq!(else_block.as_ref().unwrap().statements.len(), 1);
            }
            _ => panic!("Not an if statement. Got {:?}", program.statements[0])
        }
    }

    #[test]
    fn test_if_else_if_else_complex() {
        let input = "if a == 1 { print(1) } else if a == 2 { print(2); } else { print(3) }"; 
        let program = run_parser_test(input, 1, 0);
        match &program.statements[0] {
            Statement::If { else_if_blocks, else_block, .. } => {
                assert_eq!(else_if_blocks.len(), 1);
                assert!(else_block.is_some());
                 assert_eq!(else_block.as_ref().unwrap().statements.len(), 1);
            }
            _ => panic!("Not an if-elseif-else statement. Got {:?}", program.statements[0])
        }
    }

    #[test]
    fn test_if_block_with_semicolons_only() {
        let input = "if true { ;;; }";
        let program = run_parser_test(input, 1, 0);
         match &program.statements[0] {
            Statement::If { then_block, .. } => {
                assert!(then_block.statements.is_empty(), "Block with only semicolons should be empty of actual statements.");
            }
            _ => panic!("Not an if statement"),
        }
    }
    
    #[test]
    fn test_loop_statement_parsing() {
        let input = "loop { print(1)\n break }";
        let program = run_parser_test(input, 1, 0);
         match &program.statements[0] {
            Statement::Loop { body_block } => {
                assert_eq!(body_block.statements.len(), 2);
            }
            _ => panic!("Not a loop statement. Got {:?}", program.statements[0])
        }

        let empty_loop = "loop {}";
        let program_empty = run_parser_test(empty_loop, 1, 0);
        match &program_empty.statements[0] {
            Statement::Loop { body_block } => {
                assert!(body_block.statements.is_empty());
            }
            _ => panic!("Not an empty loop statement"),
        }
    }
    
    #[test]
    fn test_while_statement_parsing() {
        let input = "while count < 10 { count = count + 1\n continue; }"; 
         let program = run_parser_test(input, 1, 0);
         match &program.statements[0] {
            Statement::While { condition, body_block } => {
                assert!(matches!(condition, Expr::BinaryOp{op: BinaryOperator::Lt, ..}));
                assert_eq!(body_block.statements.len(), 2);
                 assert!(matches!(&body_block.statements[1], Statement::Continue));
            }
            _ => panic!("Not a while statement. Got {:?}", program.statements[0])
        }
    }

    #[test]
    fn test_for_statement_complete() {
        let input = "for let i = 0; i < 10; i = i + 1 { print(i)\n print(i*2) }"; // No parens
        let program = run_parser_test(input, 1, 0);
        match &program.statements[0] {
            Statement::For { initializer, condition, increment, body_block } => {
                assert!(initializer.is_some());
                assert!(matches!(initializer.as_ref().unwrap().as_ref(), Statement::LetDecl{..}));
                assert!(condition.is_some());
                assert!(matches!(condition.as_ref().unwrap(), Expr::BinaryOp{op: BinaryOperator::Lt, ..}));
                assert!(increment.is_some());
                assert_eq!(body_block.statements.len(), 2);
            }
            _ => panic!("Not a for statement. Got {:?}", program.statements[0])
        }
    }

    #[test]
    fn test_for_statement_minimal() {
        let input = "for ;; { break }"; // No parens, empty clauses
        let program = run_parser_test(input, 1, 0);
         match &program.statements[0] {
            Statement::For { initializer, condition, increment, body_block } => {
                assert!(initializer.is_none());
                assert!(condition.is_none());
                assert!(increment.is_none());
                assert_eq!(body_block.statements.len(), 1);
            }
            _ => panic!("Not a for statement. Got {:?}", program.statements[0])
        }
    }

    #[test]
    fn test_for_statement_only_condition() {
        let input = "for ; i < 10 ; { i = i + 1 }"; // No parens
        let program = run_parser_test(input, 1, 0);
        match &program.statements[0] {
            Statement::For { initializer, condition, increment, body_block } => {
                assert!(initializer.is_none());
                assert!(condition.is_some());
                assert!(increment.is_none());
                assert_eq!(body_block.statements.len(), 1);
            }
            _ => panic!("Not a for statement. Got {:?}", program.statements[0])
        }
    }
    
    #[test]
    fn test_break_and_continue_optional_semicolon() {
        run_parser_test("loop { break }", 1, 0);
        run_parser_test("loop { continue; }", 1, 0);
        let input = "loop { if (x) {break} else {continue;} }";
        run_parser_test(input, 1, 0);
    }

    #[test]
    fn test_sequence_of_statements_mixed_semicolons() {
        let input = r#"
            let x = 10
            let y = 20;
            x = x + y
            print(x);
            let z = "done"
        "#;
        run_parser_test(input, 5, 0);
    }

    #[test]
    fn test_empty_statements_still_parse() {
        // An empty string or only whitespace should result in 0 statements, 0 errors.
        run_parser_test("", 0, 0);
        run_parser_test("   \n\n   ", 0, 0);
        // A single semicolon is an empty statement, typically results in None from parse_statement
        // and is not added to the program's statement list.
        // The run_parser_test helper might need adjustment if it strictly checks statement count vs None.
        // For now, let's test that multiple semicolons don't cause errors.
        run_parser_test(";;;", 0, 0); // Each semicolon is an empty statement, not added to program.
    }

    #[test]
    fn test_statements_in_block_mixed_semicolons() {
        let input = r#"
        if true {
            let a = 1
            print(a);
            let b = 2
            a = a + b
        }
        "#;
        let program = run_parser_test(input, 1, 0);
        match &program.statements[0] {
            Statement::If {then_block, ..} => {
                assert_eq!(then_block.statements.len(), 4, "Block should contain 4 statements");
            }
            _ => panic!("Not an if statement")
        }
    }

    #[test]
    fn test_unterminated_block_error() {
        let input = "if true { let x = 1"; // Missing closing brace
        run_parser_test(input, 0, 1); // Expect 1 statement parsed (the if), and 1 error for unclosed block.
                                      // Or 0 statements if parse_if_statement returns None due to block error.
                                      // Current run_parser_test expects program_result.is_err(), so 0 statements.
    }
    
    // Error tests should remain largely the same, as syntax errors unrelated to semicolons
    // should still be caught.
    #[test]
    fn test_error_let_missing_equals() {
        let input = "let x 5"; // No semicolon, but error is missing '='
        run_parser_test(input, 0, 1); 
    }

    #[test]
    fn test_error_unclosed_parenthesis_in_expression() {
        let input = "let x = (5 + 2"; // No semicolon, but error is unclosed '('
        run_parser_test(input, 0, 1); 
    }
    
    #[test]
    fn test_error_if_missing_condition_parentheses() {
        // This test is no longer relevant as parentheses are not expected.
        // Instead, an error might occur if a LBrace is not found after the condition.
        // run_parser_test("if x < 10 { print(x) }", 0, 1); // Original: error for missing (
        // New: if condition is just "x", then "<" would be unexpected.
        // If "x < 10" is parsed, then if no "{", that's an error.
        run_parser_test("if x print(x)", 0, 1); // Expects { after x
    }

    #[test]
    fn test_error_if_missing_body_braces() {
        let input = "if x < 10 print(x)"; // No parens, no braces
        run_parser_test(input, 0, 1); // Expects { after condition
    }
    
    #[test]
    fn test_error_for_loop_missing_internal_semicolons() {
        // Semicolons *inside* the for (init; cond; incr) structure are still mandatory.
        let input = "for let i = 0 i < 10 i = i + 1 {}"; // No parens, missing semicolons
        run_parser_test(input, 0, 2); 
    }
    
    #[test]
    fn test_for_loop_statement_no_brace() {
        let input = "for let i = 0; i < 1; i = i + 1 print(i)";
        run_parser_test(input, 0, 1); // Expects {
    }


    #[test]
    fn test_error_unexpected_token_in_statement() {
        // This test might change slightly if `5 + ;` becomes `5+` then an empty statement.
        // However, `+` expecting an operand is the primary error.
        let input = "let x = 5 +"; 
        run_parser_test(input, 1, 1); // Let statement, error in expression.
    }
}
