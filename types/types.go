package types

// Type represents a Zeno type
type Type interface {
	String() string
}

// BasicType represents basic types like int, bool, string, float
type BasicType struct {
	Name string
}

func (bt *BasicType) String() string {
	return bt.Name
}

// Common basic types
var (
	IntType    = &BasicType{Name: "int"}
	BoolType   = &BasicType{Name: "bool"}
	StringType = &BasicType{Name: "string"}
	FloatType  = &BasicType{Name: "float"}
	AnyType    = &BasicType{Name: "any"}   // Represents any type, similar to interface{}
)

// ArrayType represents an array type.
type ArrayType struct {
	ElementType Type // The type of the elements in the array
}

// String returns a string representation of the array type.
func (a *ArrayType) String() string {
	if a.ElementType == nil || a.ElementType == AnyType {
		// Represents an array of unknown/any element type, or an empty array
		// where type couldn't be inferred without context.
		return "[]any"
	}
	return "[]" + a.ElementType.String()
}

// ResultType represents a Result<T> type for error handling
type ResultType struct {
	ValueType Type // The type of the success value
}

// String returns a string representation of the result type.
func (r *ResultType) String() string {
	return "Result<" + r.ValueType.String() + ">"
}

// Symbol represents a variable or function in the symbol table
type Symbol struct {
	Name string
	Type Type
}

// SymbolTable manages variables and their types
type SymbolTable struct {
	symbols map[string]*Symbol
	parent  *SymbolTable // For nested scopes
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]*Symbol),
		parent:  parent,
	}
}

// Define defines a new symbol in this scope
func (st *SymbolTable) Define(name string, symbolType Type) *Symbol {
	symbol := &Symbol{Name: name, Type: symbolType}
	st.symbols[name] = symbol
	return symbol
}

// Resolve looks up a symbol in this scope and parent scopes
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	symbol, ok := st.symbols[name]
	if !ok && st.parent != nil {
		symbol, ok = st.parent.Resolve(name)
	}
	return symbol, ok
}
