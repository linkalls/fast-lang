use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

use clap::Parser as ClapParser; // Alias to avoid conflict if we have our own Parser

// Assuming your library crate is named 'zeno' (check Cargo.toml)
// and it exposes the lexer, parser, and generator.
use zeno::lexer::Lexer;
use zeno::parser::Parser; // Your actual parser struct
use zeno::generator; // Assuming a generate function like generator::generate()
use zeno::ast::Program; // Assuming Program is the root AST node

#[derive(ClapParser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Zeno source file to compile
    #[arg(required = true)]
    source_file: PathBuf,

    /// Output file for the generated Rust code
    #[arg(short, long)]
    output_rust_file: Option<PathBuf>,

    /// Output file for the compiled executable
    #[arg(short = 'O', long)] // Changed from -o to -O to avoid conflict with -o for rust file
    output_executable_file: Option<PathBuf>,
    
    /// Compile the generated Rust code using rustc
    #[arg(short, long)]
    compile: bool,

    /// Run the compiled executable if compilation is successful
    #[arg(short, long, requires = "compile")]
    run: bool,

    /// Keep the generated .rs file (default: false, delete if not specified)
    #[arg(short, long)]
    keep_rs: bool,
}

fn main() -> anyhow::Result<()> {
    let args = Args::parse();

    // 1. Read Zeno source file
    let source_code = fs::read_to_string(&args.source_file)
        .map_err(|e| anyhow::anyhow!("Failed to read source file '{}': {}", args.source_file.display(), e))?;

    // 2. Lexing
    // The Lexer in this project is an iterator.
    // For the parser, we re-initialize the lexer as it consumes the input.
    // We don't need to collect all tokens here unless we want to inspect them.
    // The parser will call next_token() on its own lexer instance.

    // 3. Parsing
    let mut parser_lexer = Lexer::new(&source_code); 
    let mut parser = Parser::new(parser_lexer); 
    let ast: Program = match parser.parse_program() {
        Ok(program) => program,
        Err(errors) => {
            eprintln!("Encountered parsing errors:");
            for error in errors {
                eprintln!("  - {}", error);
            }
            return Err(anyhow::anyhow!("Parsing failed with {} error(s).", parser.errors.len()));
        }
    };

    // 4. Code Generation
    let rust_code = generator::generate(&ast)
        .map_err(|e| anyhow::anyhow!("Code generation failed: {}", e))?; // GenerationError impls Display

    // 5. Determine output Rust file path
    let rust_output_path = args.output_rust_file.clone().unwrap_or_else(|| {
        args.source_file.with_extension("rs")
    });

    fs::write(&rust_output_path, &rust_code)
        .map_err(|e| anyhow::anyhow!("Failed to write generated Rust code to '{}': {}", rust_output_path.display(), e))?;
    
    println!("Generated Rust code written to: {}", rust_output_path.display());

    if args.compile {
        // 6. Compile generated Rust code
        let executable_path = args.output_executable_file.clone().unwrap_or_else(|| {
            // Use source file stem for executable name if not provided
            let mut exe_name = args.source_file.file_stem().unwrap_or_default().to_os_string();
            if cfg!(windows) {
                exe_name.push(".exe");
            }
            args.source_file.with_file_name(exe_name)
        });


        println!("Compiling generated Rust code with rustc...");
        let mut command = Command::new("rustc");
        command.arg(&rust_output_path);
        command.arg("-o");
        command.arg(&executable_path);
        
        // Add optimization flags for release-like build
        command.arg("-C");
        command.arg("opt-level=2");


        let output = command.output() // Use output() to capture stderr for better error reporting
            .map_err(|e| anyhow::anyhow!("Failed to execute rustc: {}", e))?;

        if !output.status.success() {
            eprintln!("rustc compilation failed.");
            eprintln!("--- rustc STDOUT ---");
            eprintln!("{}", String::from_utf8_lossy(&output.stdout));
            eprintln!("--- rustc STDERR ---");
            eprintln!("{}", String::from_utf8_lossy(&output.stderr));
            return Err(anyhow::anyhow!("rustc compilation failed. Status: {}", output.status));
        }
        println!("Compilation successful. Executable at: {}", executable_path.display());

        if args.run {
            // 7. Run compiled executable
            println!("Running executable '{}'...", executable_path.display());
            let mut run_command = Command::new(&executable_path);
            let run_status = run_command.status()
                .map_err(|e| anyhow::anyhow!("Failed to run executable '{}': {}", executable_path.display(), e))?;
            
            if !run_status.success() {
                eprintln!("Executable '{}' exited with error code: {:?}", executable_path.display(), run_status.code());
            } else {
                println!("Executable finished successfully.");
            }
        }
        
        // Delete the .rs file only if compilation was successful (or attempted) and --keep-rs is not set
        if !args.keep_rs {
             fs::remove_file(&rust_output_path)
                .map_err(|e| anyhow::anyhow!("Failed to delete temporary .rs file '{}': {}", rust_output_path.display(), e))?;
             println!("Removed temporary Rust file: {}", rust_output_path.display());
        }

    } else if !args.keep_rs && args.output_rust_file.is_none() {
        // If not compiling, and output_rust_file was not specified (meaning it defaulted to source_file.rs),
        // and --keep-rs is false, then it's implied the .rs file is temporary.
        // However, the prompt says "default: false, delete if not specified [if compiling]"
        // The current logic: if not compiling, the .rs file is kept unless user deletes it manually or -o was not used.
        // This is acceptable. The main auto-deletion happens after successful compilation without --keep-rs.
        println!("Generated Rust file kept at: {}. Use --compile to build or --keep-rs to always keep it.", rust_output_path.display());
    }

    Ok(())
}
