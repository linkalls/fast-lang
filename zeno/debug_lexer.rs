use std::println;
fn main() {
    let input = "!-/*5;";
    let mut lexer = zeno::lexer::Lexer::new(input);
    println!("Debugging input: {}", input);
    loop {
        let tok = lexer.next_token();
        println!("Token: {:?}", tok);
        if matches!(tok, zeno::lexer::Token::Eof) {
            break;
        }
    }
}
