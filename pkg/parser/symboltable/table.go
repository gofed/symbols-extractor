package symboltable

import (
	"encoding/json"
	"fmt"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

// Participants:
// - symbol table (global per package, one-level table)
// - multi-level symbol table (one symbol table for each block)
// - allocated symbol table (global per package)
//   - RFE: store file:line:column as well (possible generate a link to symbol's location)

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

var (
	SymbolTypes = []string{VariableSymbol, FunctionSymbol, DataTypeSymbol}
)

type Table struct {
	symbols map[string]map[string]*gotypes.SymbolDef
	Symbols map[string][]*gotypes.SymbolDef `json:"symbols"`
}

func (t *Table) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string][]*gotypes.SymbolDef{
		VariableSymbol: t.Symbols[VariableSymbol],
		DataTypeSymbol: t.Symbols[DataTypeSymbol],
		FunctionSymbol: t.Symbols[FunctionSymbol],
	})
}

func (t *Table) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	t.Symbols = map[string][]*gotypes.SymbolDef{
		VariableSymbol: make([]*gotypes.SymbolDef, 0),
		FunctionSymbol: make([]*gotypes.SymbolDef, 0),
		DataTypeSymbol: make([]*gotypes.SymbolDef, 0),
	}

	t.symbols = map[string]map[string]*gotypes.SymbolDef{
		VariableSymbol: make(map[string]*gotypes.SymbolDef, 0),
		FunctionSymbol: make(map[string]*gotypes.SymbolDef, 0),
		DataTypeSymbol: make(map[string]*gotypes.SymbolDef, 0),
	}

	for _, symbolType := range SymbolTypes {
		if objMap[symbolType] != nil {
			var m []*gotypes.SymbolDef
			if err := json.Unmarshal(*objMap[symbolType], &m); err != nil {
				return err
			}

			t.Symbols[symbolType] = m

			for _, item := range m {
				t.symbols[symbolType][item.Name] = item
			}
		}
	}

	return nil
}

func NewTable() *Table {
	return &Table{
		symbols: map[string]map[string]*gotypes.SymbolDef{
			VariableSymbol: make(map[string]*gotypes.SymbolDef, 0),
			FunctionSymbol: make(map[string]*gotypes.SymbolDef, 0),
			DataTypeSymbol: make(map[string]*gotypes.SymbolDef, 0),
		},
		Symbols: map[string][]*gotypes.SymbolDef{
			VariableSymbol: make([]*gotypes.SymbolDef, 0),
			FunctionSymbol: make([]*gotypes.SymbolDef, 0),
			DataTypeSymbol: make([]*gotypes.SymbolDef, 0),
		},
	}
}

func (t *Table) addSymbol(symbolType string, name string, sym *gotypes.SymbolDef) error {
	if def, ko := t.symbols[symbolType][name]; ko {
		// If the symbol definition is empty, we just allocated a name for the symbol.
		// Later on, it gets populated with its definition.
		if def.Def != nil {
			return fmt.Errorf("Symbol '%s' already exists", sym.Name)
		}
		def.Def = sym.Def
		return nil
	}

	t.symbols[symbolType][name] = sym
	t.Symbols[symbolType] = append(t.Symbols[symbolType], sym)

	return nil
}

func (t *Table) AddVariable(sym *gotypes.SymbolDef) error {
	// TODO(jchaloup): Given one can re-assign a variable (if the new type is assignable to the old one)
	//                 we need to allow update of variable's data type.
	return t.addSymbol(VariableSymbol, sym.Name, sym)
}

func (t *Table) AddDataType(sym *gotypes.SymbolDef) error {
	// TODO(jchaloup): extend the SymbolDef to generate symbol names for various symbol types
	return t.addSymbol(DataTypeSymbol, sym.Name, sym)
}

func (t *Table) AddFunction(sym *gotypes.SymbolDef) error {
	return t.addSymbol(FunctionSymbol, sym.Name, sym)
}

func (t *Table) LookupVariable(key string) (*gotypes.SymbolDef, error) {
	if sym, ok := t.symbols[VariableSymbol][key]; ok {
		fmt.Printf("Variable found: %#v\n", sym)
		return sym, nil
	}
	return nil, fmt.Errorf("Variable `%v` not found", key)
}

func (t *Table) Lookup(key string) (*gotypes.SymbolDef, SymbolType, error) {

	for _, symbolType := range SymbolTypes {
		if sym, ok := t.symbols[symbolType][key]; ok {
			fmt.Printf("Symbol found: %#v\n", sym)
			return sym, SymbolType(symbolType), nil
		}
	}

	return nil, SymbolType(""), fmt.Errorf("Symbol `%v` not found", key)
}
