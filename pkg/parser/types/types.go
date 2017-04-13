package types

import (
	"go/ast"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type TypeParser interface {
	Parse(d ast.Expr) (gotypes.DataType, error)
}

type ExpressionParser interface {
	Parse(expr ast.Expr) ([]gotypes.DataType, error)
}

type StatementParser interface {
	Parse(statement ast.Stmt) error
}
