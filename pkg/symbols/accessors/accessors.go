package accessors

import (
	"fmt"
	"go/ast"

	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

type Accessor struct {
	// package name
	packageName string
	// per package symbol table
	symbolTable symbols.SymbolLookable
	// per subset of packages symbol table
	globalSymbolTable *global.Table
}

func NewAccessor(globalSymbolTable *global.Table) *Accessor {
	return &Accessor{
		globalSymbolTable: globalSymbolTable,
	}
}

func (a *Accessor) SetCurrentTable(packageName string, symbolTable symbols.SymbolLookable) *Accessor {
	a.packageName = packageName
	a.symbolTable = symbolTable

	return a
}

func (a *Accessor) GetBuiltin(name string) (*symbols.SymbolDef, symbols.SymbolType, error) {
	table, err := a.globalSymbolTable.Lookup("builtin")
	if err != nil {
		return nil, "", err
	}
	return table.Lookup(name)
}

func (a *Accessor) IsBuiltin(name string) bool {
	table, err := a.globalSymbolTable.Lookup("builtin")
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
func (a *Accessor) Lookup(ident *gotypes.Identifier) (*symbols.SymbolDef, symbols.SymbolType, error) {
	if ident.Package == "" {
		return nil, "", fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		return a.symbolTable.Lookup(ident.Def)
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, "", err
	}
	return table.Lookup(ident.Def)
}

func (a *Accessor) LookupMethod(ident *gotypes.Identifier, method string) (*symbols.SymbolDef, error) {
	if ident.Package == "" {
		return nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		return a.symbolTable.LookupMethod(ident.Def, method)
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, err
	}
	return table.LookupMethod(ident.Def, method)
}

func (a *Accessor) LookupDataType(ident *gotypes.Identifier) (*symbols.SymbolDef, symbols.SymbolLookable, error) {
	if ident.Package == "" {
		return nil, nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		def, err := a.symbolTable.LookupDataType(ident.Def)
		return def, a.symbolTable, err
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, nil, err
	}
	def, err := table.LookupDataType(ident.Def)
	return def, table, err
}

func (a *Accessor) RetrieveQidDataType(qidprefix gotypes.DataType, item *ast.Ident) (symbols.SymbolLookable, *symbols.SymbolDef, error) {
	// qid.structtype expected
	qid, ok := qidprefix.(*gotypes.Packagequalifier)
	if !ok {
		return nil, nil, fmt.Errorf("Expecting a qid.id when retrieving a symbol from a selector expression")
	}
	glog.Infof("Trying to retrieve a symbol %#v from package %v\n", item.String(), qid.Path)
	qidst, err := a.globalSymbolTable.Lookup(qid.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid.Path, err)
	}

	dataTypeDef, piErr := qidst.LookupDataType(item.String())
	if piErr != nil {
		return nil, nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", item.String(), qid.Path, piErr)
	}
	return qidst, dataTypeDef, nil
}

func (a *Accessor) FindFirstNonidDataType(typeDef gotypes.DataType) (gotypes.DataType, error) {
	var symbolDef *symbols.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		def, _, err := a.LookupDataType(typeDefType)
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
	return a.FindFirstNonidDataType(symbolDef.Def)
}
