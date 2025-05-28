use crate::ast::*;
use std::fmt::Write;

#[derive(Debug, Clone, PartialEq)]
pub struct GenerationError(String);

impl std::fmt::Display for GenerationError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "Generation Error: {}", self.0)
    }
}

impl std::error::Error for GenerationError {}

// Main generation function
pub fn generate(program: &Program) -> Result<String, GenerationError> {
    let mut rust_code = String::new();
    writeln!(rust_code, "fn main() {{").unwrap();

    for statement in &program.statements {
        generate_statement(statement, &mut rust_code, 1)?;
    }

    writeln!(rust_code, "}}").unwrap();
    Ok(rust_code)
}

// Helper function for indentation
fn indent(level: usize) -> String {
    "    ".repeat(level)
}

// Map SIMPLELANG type strings to Rust type strings
fn map_type(simple_type: &str) -> String {
    match simple_type {
        "int" => "i64".to_string(),
        "float" => "f64".to_string(),
        "bool" => "bool".to_string(),
        "string" => "String".to_string(),
        // If not a known simple type, assume it's already a valid Rust type or needs specific handling.
        _ => simple_type.to_string(), 
    }
}

// Statement generation
fn generate_statement(statement: &Statement, writer: &mut String, indent_level: usize) -> Result<(), GenerationError> {
    write!(writer, "{}", indent(indent_level)).unwrap();
    match statement {
        Statement::LetDecl { name, type_ann, mutable, value_expr } => {
            write!(writer, "let {}{}", if *mutable { "mut " } else { "" }, name).unwrap();
            if let Some(ann) = type_ann {
                write!(writer, ": {}", map_type(ann)).unwrap();
            }
            write!(writer, " = ").unwrap();
            generate_expression(value_expr, writer)?;
            writeln!(writer, ";").unwrap();
        }
        Statement::Assignment { name, value_expr } => {
            write!(writer, "{} = ", name).unwrap();
            generate_expression(value_expr, writer)?;
            writeln!(writer, ";").unwrap();
        }
        Statement::ExprStatement { expr } => {
            generate_expression(expr, writer)?;
            writeln!(writer, ";").unwrap();
        }
        Statement::If { condition, then_block, else_if_blocks, else_block } => {
            write!(writer, "if ").unwrap();
            generate_expression(condition, writer)?;
            write!(writer, " ").unwrap(); 
            generate_block(then_block, writer, indent_level)?;

            for (else_if_condition, else_if_block) in else_if_blocks {
                write!(writer, " else if ").unwrap();
                generate_expression(else_if_condition, writer)?;
                write!(writer, " ").unwrap();
                generate_block(else_if_block, writer, indent_level)?;
            }

            if let Some(eb) = else_block {
                write!(writer, " else ").unwrap();
                generate_block(eb, writer, indent_level)?;
            }
            writeln!(writer).unwrap(); 
        }
        Statement::Loop { body_block } => {
            write!(writer, "loop ").unwrap();
            generate_block(body_block, writer, indent_level)?;
            writeln!(writer).unwrap();
        }
        Statement::While { condition, body_block } => {
            write!(writer, "while ").unwrap();
            generate_expression(condition, writer)?;
            write!(writer, " ").unwrap();
            generate_block(body_block, writer, indent_level)?;
            writeln!(writer).unwrap();
        }
        Statement::For { initializer, condition, increment, body_block } => {
            // Outer scope for the initializer if it's a LetDecl
            let needs_outer_scope = matches!(initializer, Some(box Statement::LetDecl{..}));
            if needs_outer_scope {
                // This creates a slight oddity if the initializer isn't a let decl,
                // but is required if `let` is used in the initializer part of a C-style for.
                // A more robust solution might involve desugaring `for` into a block with the initializer
                // and then a loop. For now, this handles simple `let` initializers.
                // write!(writer, "{{\n", indent(indent_level)).unwrap();
                // let effective_indent_level = indent_level + if needs_outer_scope { 1 } else { 0 };
            }

            if let Some(init_stmt) = initializer {
                 // Generate initializer without its own line's indent, but respect its content's indent if it's a block (not typical for for-init)
                let mut temp_writer = String::new();
                generate_statement(init_stmt, &mut temp_writer, 0)?; // Generate with 0 base indent
                write!(writer, "{}", temp_writer.trim_start()).unwrap(); // Add to main writer, trim leading spaces from its own generation
            }
            
            write!(writer, "while ").unwrap();
            if let Some(cond_expr) = condition {
                generate_expression(cond_expr, writer)?;
            } else {
                write!(writer, "true").unwrap(); 
            }
            write!(writer, " ").unwrap(); 
            
            // Open block for while body
            writeln!(writer, "{{").unwrap();
            for stmt in &body_block.statements {
                generate_statement(stmt, writer, indent_level + 1)?;
            }
            if let Some(inc_expr) = increment {
                 write!(writer, "{}", indent(indent_level + 1)).unwrap();
                 generate_expression(inc_expr, writer)?;
                 writeln!(writer, ";").unwrap();
            }
            writeln!(writer, "{}}}", indent(indent_level)).unwrap();

            if needs_outer_scope {
                // writeln!(writer, "{}}}", indent(indent_level -1 )).unwrap(); // Close outer scope
            }
        }
        Statement::Print { expr, newline } => {
            let macro_name = if *newline { "println!" } else { "print!" };
            // Basic version: assumes expr directly maps to a displayable type.
            // More robust: check expr type, use "{:?}" for complex types if no direct Display.
            write!(writer, "{}(\"{{}}\", ", macro_name).unwrap();
            generate_expression(expr, writer)?;
            writeln!(writer, ");").unwrap();
        }
        Statement::Break => {
            writeln!(writer, "break;").unwrap();
        }
        Statement::Continue => {
            writeln!(writer, "continue;").unwrap();
        }
    }
    Ok(())
}

// Expression generation
fn generate_expression(expression: &Expr, writer: &mut String) -> Result<(), GenerationError> {
    match expression {
        Expr::Integer(val) => write!(writer, "{}_i64", val).unwrap(),
        Expr::Float(val) => {
            if val.fract() == 0.0 {
                write!(writer, "{}.0_f64", val).unwrap(); // Ensure it's treated as float e.g. 10.0
            } else {
                write!(writer, "{}_f64", val).unwrap();
            }
        }
        Expr::StringLiteral(s) => {
            write!(writer, "\"{}\"", s.escape_default().to_string()).unwrap();
        }
        Expr::Boolean(b) => write!(writer, "{}", b).unwrap(),
        Expr::Identifier(name) => write!(writer, "{}", name).unwrap(),
        Expr::BinaryOp { left, op, right } => {
            // Parenthesize all binary operations for safety and clarity.
            write!(writer, "(").unwrap();
            generate_expression(left, writer)?;
            match op {
                BinaryOperator::Plus => write!(writer, " + ").unwrap(),
                BinaryOperator::Minus => write!(writer, " - ").unwrap(),
                BinaryOperator::Multiply => write!(writer, " * ").unwrap(),
                BinaryOperator::Divide => write!(writer, " / ").unwrap(),
                BinaryOperator::Modulo => write!(writer, " % ").unwrap(),
                BinaryOperator::Eq => write!(writer, " == ").unwrap(),
                BinaryOperator::NotEq => write!(writer, " != ").unwrap(),
                BinaryOperator::Lt => write!(writer, " < ").unwrap(),
                BinaryOperator::Lte => write!(writer, " <= ").unwrap(),
                BinaryOperator::Gt => write!(writer, " > ").unwrap(),
                BinaryOperator::Gte => write!(writer, " >= ").unwrap(),
                BinaryOperator::And => write!(writer, " && ").unwrap(),
                BinaryOperator::Or => write!(writer, " || ").unwrap(),
            }
            generate_expression(right, writer)?;
            write!(writer, ")").unwrap();
        }
        Expr::UnaryOp { op, expr } => {
            // Parenthesize unary operations as well.
            write!(writer, "(").unwrap();
            match op {
                UnaryOperator::Not => write!(writer, "!").unwrap(),
                UnaryOperator::Negate => write!(writer, "-").unwrap(),
            }
            generate_expression(expr, writer)?;
            write!(writer, ")").unwrap();
        }
        Expr::Call { callee, args } => {
            write!(writer, "{}(", callee).unwrap();
            for (i, arg) in args.iter().enumerate() {
                if i > 0 {
                    write!(writer, ", ").unwrap();
                }
                generate_expression(arg, writer)?;
            }
            write!(writer, ")").unwrap();
        }
    }
    Ok(())
}

// Block generation
fn generate_block(block: &Block, writer: &mut String, indent_level: usize) -> Result<(), GenerationError> {
    writeln!(writer, "{{").unwrap();
    for statement in &block.statements {
        generate_statement(statement, writer, indent_level + 1)?;
    }
    write!(writer, "{}}}", indent(indent_level)).unwrap(); 
    Ok(())
}


#[cfg(test)]
mod tests {
    use super::*;
    use crate::lexer::Lexer;
    use crate::parser::Parser;

    fn run_generator_test(input_simplelang: &str, expected_rust_substrings: Vec<&str>) -> String {
        let l = Lexer::new(input_simplelang);
        let mut p = Parser::new(l);
        let program_result = p.parse_program();
        
        if let Err(parser_errors) = &program_result {
             eprintln!("Parser errors for input:\n{}\nErrors: {:?}", input_simplelang, parser_errors);
        }
        assert!(program_result.is_ok(), "Parser failed");
        
        let program = program_result.unwrap();
        let rust_code_result = generate(&program);

        if let Err(gen_error) = &rust_code_result {
            eprintln!("Generator error for input:\n{}\nError: {}", input_simplelang, gen_error);
        }
        assert!(rust_code_result.is_ok(), "Generator failed");
        
        let rust_code = rust_code_result.unwrap();
        println!("\n--- SimpleLang Input:\n{}\n--- Generated Rust Output: ---\n{}\n---------------------------\n", input_simplelang, rust_code);

        for sub in &expected_rust_substrings {
            assert!(rust_code.contains(sub), "Generated code does not contain expected substring: '{}'.\nFull code:\n{}", sub, rust_code);
        }
        
        // Basic check for balanced braces
        let mut brace_count = 0;
        for char_code in rust_code.chars(){
            if char_code == '{' { brace_count += 1; }
            else if char_code == '}' { brace_count -=1; }
        }
        assert_eq!(brace_count, 0, "Unbalanced braces in generated code for input:\n{}", input_simplelang);
        rust_code
    }

    #[test]
    fn test_generate_let_and_assign() {
        run_generator_test("let x = 10; let mut y: float = 20.0; y = x + 15.5;", vec![
            "let x = 10_i64;",
            "let mut y: f64 = 20.0_f64;",
            "y = (x + 15.5_f64);",
        ]);
        run_generator_test("let z = 1 + 2", vec!["let z = (1_i64 + 2_i64);"]);
        run_generator_test("mut count = 0", vec!["let mut count = 0_i64;"]);
    }

    #[test]
    fn test_for_loop_only_initializer() {
        let code = "for (let i = 10;;) { print(i); break; }";
        run_generator_test(code, vec![
            "let mut i = 10_i64;", // Assuming let implies mut for typical for-loop rebinds, or parser handles it. Generator should reflect AST.
                                 // If AST makes `i` immutable, then `let i = 10_i64;` is correct.
                                 // Current AST for `for`'s initializer is `Option<Box<Statement>>`.
                                 // If it's `LetDecl{mutable: false}`, then this test should reflect that.
                                 // The generator for `LetDecl` respects `mutable`.
                                 // Let's assume for-loop initializers are often mutable in spirit,
                                 // but the AST/parser rule for `let i = 0` in for might make it immutable.
                                 // For test robustness, let's ensure the Zeno code reflects this.
                                 // The example `for (let i = 0; ...)` implies `i` can be mutable.
                                 // The parser for `for (let i =0; ...)` will create a LetDecl.
                                 // If `mut` is not used, it's immutable.
                                 // The generator's for-loop desugaring should correctly place this let.

            // Correcting the Zeno code to make `i` mutable if it's intended to be changed by an increment (even if missing here)
            // Or, if `i` is not changed in the loop and only used, immutable is fine.
            // The current generator for `for` puts initializer, then `while condition { body; increment }`.
            // So, `let i = 10; while true { print(i); break; }` is the expected Rust.
            "let i = 10_i64;",
            "while true {",
            "print!(\"{}\", i);",
            "break;",
            "}",
        ]);
        
        let code_mut = "for (let mut i = 0;;) { print(i); break; }";
         run_generator_test(code_mut, vec![
            "let mut i = 0_i64;",
            "while true {",
            "print!(\"{}\", i);",
            "break;",
            "}",
        ]);
    }

    #[test]
    fn test_generate_print_statements() {
        run_generator_test("print(123); println(\"hello\"); print(true);", vec![
            "print!(\"{}\", 123_i64);",
            "println!(\"{}\", \"hello\");",
            "print!(\"{}\", true);",
        ]);
    }

    #[test]
    fn test_generate_arithmetic_and_boolean_expressions() {
        run_generator_test("let v = (1 + 2) * 3 - 4 / 2 % 3;", vec!["let v = (((((1_i64 + 2_i64) * 3_i64) - (4_i64 / 2_i64)) % 3_i64));"]);
        run_generator_test("let b = !true && (false || (1 < 2));", vec!["let b = ((!true) && (false || (1_i64 < 2_i64)));"]);
    }
    
    #[test]
    fn test_generate_if_else_if_else() {
        let code = "if (x > 10) { print(1); } else if (x < 5) { print(2); } else { print(3); }";
        run_generator_test(code, vec![
            "if (x > 10_i64) {",
            "print!(\"{}\", 1_i64);",
            "} else if (x < 5_i64) {",
            "print!(\"{}\", 2_i64);",
            "} else {",
            "print!(\"{}\", 3_i64);",
            "}",
        ]);
    }

    #[test]
    fn test_generate_loop_with_break_continue() {
        let code = "let mut i = 0; loop { i = i + 1; if (i == 2) { continue; } if (i > 3) { break; } print(i); }";
        run_generator_test(code, vec![
            "loop {",
            "i = (i + 1_i64);",
            "if (i == 2_i64) {",
            "continue;",
            "if (i > 3_i64) {",
            "break;",
            "print!(\"{}\", i);",
        ]);
    }

    #[test]
    fn test_generate_while_loop() {
        let code = "let mut counter = 10; while (counter > 0) { print(counter); counter = counter - 1; }";
        run_generator_test(code, vec![
            "let mut counter = 10_i64;",
            "while (counter > 0_i64) {",
            "print!(\"{}\", counter);",
            "counter = (counter - 1_i64);",
            "}",
        ]);
    }

    #[test]
    fn test_generate_for_loop_as_while() {
        let code = "for (let i = 0; i < 3; i = i + 1) { println(i); }";
        run_generator_test(code, vec![
            "let mut i = 0_i64;",      // Initializer part of for
            "while (i < 3_i64) {",     // Condition part
            "println!(\"{}\", i);",  // Body
            "i = (i + 1_i64);",        // Increment part at end of block
            "}",                       // Closing while
        ]);
    }
    
    #[test]
    fn test_for_loop_empty_parts() {
        let code = "for (;;) { if check() { break; } }";
         run_generator_test(code, vec![
            "while true {", // No condition means true
            "if check() {",
            "break;",
            "}",
        ]);
    }

    #[test]
    fn test_generate_string_escaping() {
        run_generator_test("let a = \"Hello\\nWorld\\t\\\"Quoted\\\"\";", vec!["let a = \"Hello\\nWorld\\t\\\"Quoted\\\"\";"]);
    }

    #[test]
    fn test_generate_call_expression() {
        // This assumes `some_external_function` would be available in the Rust environment.
        run_generator_test("let res = some_external_function(arg1, 10 + 2, \"str_arg\");", 
            vec!["let res = some_external_function(arg1, (10_i64 + 2_i64), \"str_arg\");"]);
    }

    #[test]
    fn test_nested_blocks_and_indentation() {
        let code = r#"
        let x = 1;
        if (x == 1) {
            let y = 2;
            if (y == 2) {
                print(3);
                loop {
                    break;
                }
            } else {
                print(4);
            }
        } else {
            print(0);
        }
        "#;
        let generated_code = run_generator_test(code, vec![
            "fn main() {\n",
            "    let x = 1_i64;\n",
            "    if (x == 1_i64) {\n",
            "        let y = 2_i64;\n",
            "        if (y == 2_i64) {\n",
            "            print!(\"{}\", 3_i64);\n",
            "            loop {\n",
            "                break;\n",
            "            }\n", // loop
            "        } else {\n",
            "            print!(\"{}\", 4_i64);\n",
            "        }\n", // inner if-else
            "    } else {\n",
            "        print!(\"{}\", 0_i64);\n",
            "    }\n", // outer if-else
            "}\n", // main
        ]);
        // For more precise indentation check, can compare line by line with expected.
        // This is a basic structural check.
        assert!(generated_code.contains("        if (y == 2_i64) {\n"));
        assert!(generated_code.contains("            loop {\n"));
        assert!(generated_code.contains("                break;\n"));
    }
}
