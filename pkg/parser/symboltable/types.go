package symboltable

import (
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type SymbolLookable interface {
	LookupVariable(key string) (*gotypes.SymbolDef, error)
	LookupVariableLikeSymbol(key string) (*gotypes.SymbolDef, error)
	LookupFunction(key string) (*gotypes.SymbolDef, error)
	LookupDataType(key string) (*gotypes.SymbolDef, error)
	LookupMethod(datatype, methodName string) (*gotypes.SymbolDef, error)
	Lookup(key string) (*gotypes.SymbolDef, SymbolType, error)
	Exists(name string) bool
}

type SymbolType string

const (
	VariableSymbol = "variables"
	FunctionSymbol = "functions"
	DataTypeSymbol = "datatypes"
)

func (s SymbolType) IsVariable() bool {
	return s == VariableSymbol
}

func (s SymbolType) IsDataType() bool {
	return s == DataTypeSymbol
}

func (s SymbolType) IsFunctionType() bool {
	return s == FunctionSymbol
}

var (
	SymbolTypes = []string{VariableSymbol, FunctionSymbol, DataTypeSymbol}
)

type SymbolTable interface {
	SymbolLookable
	AddVariable(sym *gotypes.SymbolDef) error
	AddDataType(sym *gotypes.SymbolDef) error
	AddFunction(sym *gotypes.SymbolDef) error
}
