package symbols

type SymbolLookable interface {
	LookupVariable(key string) (*SymbolDef, error)
	LookupVariableLikeSymbol(key string) (*SymbolDef, SymbolType, error)
	LookupFunction(key string) (*SymbolDef, error)
	LookupDataType(key string) (*SymbolDef, error)
	LookupMethod(datatype, methodName string) (*SymbolDef, error)
	LookupAllMethods(datatype string) (map[string]*SymbolDef, error)
	Lookup(key string) (*SymbolDef, SymbolType, error)
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
	AddVariable(sym *SymbolDef) error
	AddDataType(sym *SymbolDef) error
	AddFunction(sym *SymbolDef) error
}
