package types

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	contracttable "github.com/gofed/symbols-extractor/pkg/parser/contracts/table"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/stack"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

// TypeParser implementation is responsible for Go data type parsing/processing
type TypeParser interface {
	Parse(d ast.Expr) (gotypes.DataType, error)
}

type ExprAttribute struct {
	DataTypeList []gotypes.DataType
	TypeVarList  []typevars.Interface
	// PropagationSequence []string
}

func (e *ExprAttribute) AddTypeVar(typevar typevars.Interface) *ExprAttribute {
	e.TypeVarList = append(e.TypeVarList, typevar)
	glog.Infof("Adding TypeVar: %v\n", typevars.TypeVar2String(typevar))
	return e
}

func ExprAttributeFromDataType(list ...gotypes.DataType) *ExprAttribute {
	return &ExprAttribute{
		DataTypeList: list,
	}
}

// ExpressionParser implemenation is responsible for Go expression parsing/processing
type ExpressionParser interface {
	Parse(expr ast.Expr) (*ExprAttribute, error)
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
	ParseValueSpec(spec *ast.ValueSpec) ([]*symbols.SymbolDef, error)
}

// Config for a parser
type Config struct {
	// package name
	PackageName string
	// file
	FileName string
	// per file symbol table
	SymbolTable *stack.Stack
	// per subset of packages symbol table
	GlobalSymbolTable *global.Table
	// per file allocatable ST
	AllocatedSymbolsTable *alloctable.Table
	// types parser
	TypeParser TypeParser
	// expresesion parser
	ExprParser ExpressionParser
	// statement parser
	StmtParser StatementParser
	// per-file contract table
	ContractTable *contracttable.Table
	// symbols accessor
	SymbolsAccessor *accessors.Accessor
}

func (c *Config) SymbolPos(pos token.Pos) string {
	return fmt.Sprintf("%v:%v", c.FileName, pos)
}
