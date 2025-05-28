#[derive(Debug, Clone, PartialEq)]
pub struct Program {
    pub statements: Vec<Statement>,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Statement {
    LetDecl {
        name: String,
        type_ann: Option<String>,
        mutable: bool,
        value_expr: Expr,
    },
    Assignment {
        name: String,
        value_expr: Expr,
    },
    ExprStatement {
        expr: Expr,
    },
    If {
        condition: Expr,
        then_block: Block,
        else_if_blocks: Vec<(Expr, Block)>,
        else_block: Option<Block>,
    },
    While {
        condition: Expr,
        body_block: Block,
    },
    Loop {
        body_block: Block,
    },
    For {
        initializer: Option<Box<Statement>>,
        condition: Option<Expr>,
        increment: Option<Expr>,
        body_block: Block,
    },
    Print {
        expr: Expr,
        newline: bool,
    },
    Break,
    Continue,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Block {
    pub statements: Vec<Statement>,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Expr {
    Integer(i64),
    Float(f64),
    StringLiteral(String),
    Boolean(bool),
    Identifier(String),
    BinaryOp {
        left: Box<Expr>,
        op: BinaryOperator,
        right: Box<Expr>,
    },
    UnaryOp {
        op: UnaryOperator,
        expr: Box<Expr>,
    },
    Call {
        callee: String,
        args: Vec<Expr>,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub enum BinaryOperator {
    Plus,
    Minus,
    Multiply,
    Divide,
    Modulo,
    Eq,
    NotEq,
    Lt,
    Lte,
    Gt,
    Gte,
    And,
    Or,
}

#[derive(Debug, Clone, PartialEq)]
pub enum UnaryOperator {
    Not,
    Negate,
}
