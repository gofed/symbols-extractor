package types

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	contracttable "github.com/gofed/symbols-extractor/pkg/parser/contracts/table"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
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
	ParseValueSpec(spec *ast.ValueSpec) ([]*symboltable.SymbolDef, error)
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
}

func (c *Config) SymbolPos(pos token.Pos) string {
	return fmt.Sprintf("%v:%v", c.FileName, pos)
}

func (c *Config) GetBuiltin(name string) (*symboltable.SymbolDef, symboltable.SymbolType, error) {
	table, err := c.GlobalSymbolTable.Lookup("builtin")
	if err != nil {
		return nil, "", err
	}
	return table.Lookup(name)
}

func (c *Config) IsBuiltin(name string) bool {
	table, err := c.GlobalSymbolTable.Lookup("builtin")
	if err != nil {
		glog.Warning(err)
		return false
	}
	if _, _, err := table.Lookup(name); err == nil {
		return true
	}
	return false
}

// Lookup retrieves a definition of identifier ident
func (c *Config) Lookup(ident *gotypes.Identifier) (*symboltable.SymbolDef, symboltable.SymbolType, error) {
	if ident.Package == "" {
		return nil, "", fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == c.PackageName {
		return c.SymbolTable.Lookup(ident.Def)
	}
	table, err := c.GlobalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, "", err
	}
	return table.Lookup(ident.Def)
}

func (c *Config) LookupMethod(ident *gotypes.Identifier, method string) (*symboltable.SymbolDef, error) {
	if ident.Package == "" {
		return nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == c.PackageName {
		return c.SymbolTable.LookupMethod(ident.Def, method)
	}
	table, err := c.GlobalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, err
	}
	return table.LookupMethod(ident.Def, method)
}

func (c *Config) LookupDataType(ident *gotypes.Identifier) (*symboltable.SymbolDef, symboltable.SymbolLookable, error) {
	if ident.Package == "" {
		return nil, nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == c.PackageName {
		def, err := c.SymbolTable.LookupDataType(ident.Def)
		return def, c.SymbolTable, err
	}
	table, err := c.GlobalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, nil, err
	}
	def, err := table.LookupDataType(ident.Def)
	return def, table, err
}

func (c *Config) RetrieveQidDataType(qidselector *gotypes.Selector) (symboltable.SymbolLookable, *symboltable.SymbolDef, error) {
	// qid.structtype expected
	qid, ok := qidselector.Prefix.(*gotypes.Packagequalifier)
	if !ok {
		return nil, nil, fmt.Errorf("Expecting a qid.structtype when retrieving a struct from a selector expression")
	}
	glog.Infof("Trying to retrieve a symbol %#v from package %v\n", qidselector.Item, qid.Path)
	qidst, err := c.GlobalSymbolTable.Lookup(qid.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid.Path, err)
	}

	dataTypeDef, piErr := qidst.LookupDataType(qidselector.Item)
	if piErr != nil {
		return nil, nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", qidselector.Item, qid.Path, piErr)
	}
	return qidst, dataTypeDef, nil
}

func (c *Config) FindFirstNonidDataType(typeDef gotypes.DataType) (gotypes.DataType, error) {
	var symbolDef *symboltable.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := c.RetrieveQidDataType(typeDefType)
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		def, _, err := c.LookupDataType(typeDefType)
		if err != nil {
			return nil, err
		}
		symbolDef = def
	default:
		return typeDef, nil
	}
	if symbolDef.Def == nil {
		return nil, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
	}
	return c.FindFirstNonidDataType(symbolDef.Def)
}
