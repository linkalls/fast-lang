pub mod ast;
pub mod lexer;
pub mod parser;
pub mod generator;

// Keep the original add function and its test for now, or remove if not needed.
pub fn add(left: usize, right: usize) -> usize {
    left + right
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        let result = add(2, 2);
        assert_eq!(result, 4);
    }
}
