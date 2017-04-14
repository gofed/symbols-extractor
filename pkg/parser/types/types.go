package types

import (
	"go/ast"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

// TypeParser implementation is responsible for Go data type parsing/processing
type TypeParser interface {
	Parse(d ast.Expr) (gotypes.DataType, error)
}

// ExpressionParser implemenation is responsible for Go expression parsing/processing
type ExpressionParser interface {
	Parse(expr ast.Expr) ([]gotypes.DataType, error)
}

// StatementParser implemenation is responsible for:
// - statement parsing/processing (block is considered a statement)
// -
type StatementParser interface {
	// Parse a general statement (there are no particular variables attached to the statement)
	Parse(statement ast.Stmt) error
	// ParseFuncDecl parses function declaration only
	ParseFuncDecl(d *ast.FuncDecl) (gotypes.DataType, error)
	// ParseFuncBody parses function body with pushing function parameter(s) and/or its receiver into the symbol table
	ParseFuncBody(funcDecl *ast.FuncDecl) error
	// ParseValueSpec parses variable/constant definition/declaration
	ParseValueSpec(spec *ast.ValueSpec) ([]*gotypes.SymbolDef, error)
}
