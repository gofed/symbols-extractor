package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

// Participants:
// - symbol table (global per package, one-level table)
// - multi-level symbol table (one symbol table for each block)
// - allocated symbol table (global per package)
//   - RFE: store file:line:column as well (possible generate a link to symbol's location)

const (
	VariableSymbol = "variables"
	FunctionSymbol = "functions"
	DataTypeSymbol = "datatypes"
)

var (
	SymbolTypes = []string{VariableSymbol, FunctionSymbol, DataTypeSymbol}
)

type SymbolTable struct {
	symbols map[string]map[string]*gotypes.SymbolDef
	Symbols map[string][]*gotypes.SymbolDef `json:"symbols"`
}

func (st *SymbolTable) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string][]*gotypes.SymbolDef{
		VariableSymbol: st.Symbols[VariableSymbol],
		DataTypeSymbol: st.Symbols[DataTypeSymbol],
		FunctionSymbol: st.Symbols[FunctionSymbol],
	})
}

func (st *SymbolTable) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	st.Symbols = map[string][]*gotypes.SymbolDef{
		VariableSymbol: make([]*gotypes.SymbolDef, 0),
		FunctionSymbol: make([]*gotypes.SymbolDef, 0),
		DataTypeSymbol: make([]*gotypes.SymbolDef, 0),
	}

	st.symbols = map[string]map[string]*gotypes.SymbolDef{
		VariableSymbol: make(map[string]*gotypes.SymbolDef, 0),
		FunctionSymbol: make(map[string]*gotypes.SymbolDef, 0),
		DataTypeSymbol: make(map[string]*gotypes.SymbolDef, 0),
	}

	for _, symbolType := range SymbolTypes {
		var m []*gotypes.SymbolDef
		if err := json.Unmarshal(*objMap[symbolType], &m); err != nil {
			return err
		}

		st.Symbols[symbolType] = m

		for _, item := range m {
			st.symbols[symbolType][item.Name] = item
		}
	}

	return nil
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
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

func (o *SymbolTable) addSymbol(symbolType string, name string, sym *gotypes.SymbolDef) error {
	if _, ko := o.symbols[symbolType][name]; ko {
		return fmt.Errorf("Symbol '%s' already exists", sym.Name)
	}

	o.symbols[symbolType][name] = sym
	o.Symbols[symbolType] = append(o.Symbols[symbolType], sym)
	return nil
}

func (o *SymbolTable) AddVariable(sym *gotypes.SymbolDef) error {
	return o.addSymbol(VariableSymbol, sym.Name, sym)
}

func (o *SymbolTable) AddDataType(sym *gotypes.SymbolDef) error {
	// TODO(jchaloup): extend the SymbolDef to generate symbol names for various symbol types
	return o.addSymbol(DataTypeSymbol, sym.Name, sym)
}

func (o *SymbolTable) AddFunction(sym *gotypes.SymbolDef) error {
	return o.addSymbol(FunctionSymbol, sym.Name, sym)
}

func (st *SymbolTable) Lookup(key string) (*gotypes.SymbolDef, error) {

	for _, symbolType := range SymbolTypes {
		if sym, ok := st.symbols[symbolType][key]; ok {
			return sym, nil
		}
	}

	return nil, fmt.Errorf("Symbol `%v` not found", key)
}

// - count a number of each symbol used (to watch how intensively is a given symbol used)
// - each AS table is per file, AS package is a union of file AS tables
//

type AllocatedSymbolsTable struct {
	File    string
	Package string
	// symbol's name is in a PACKAGE.ID form
	// if the PACKAGE is empty, the ID is considired as the embedded symbol
	symbols map[string]int
}

func NewAllocatableSymbolsTable() *AllocatedSymbolsTable {
	return &AllocatedSymbolsTable{
		symbols: make(map[string]int),
	}
}

func (ast *AllocatedSymbolsTable) AddSymbol(origin, id string) {
	var key string
	if origin != "" {
		key = strings.Join([]string{origin, id}, ".")
	} else {
		key = id
	}

	count, ok := ast.symbols[key]
	if !ok {
		ast.symbols[key] = 1
	} else {
		ast.symbols[key] = count + 1
	}
}

func (ast *AllocatedSymbolsTable) Print() {
	for key := range ast.symbols {
		fmt.Printf("%v:\t%v\n", key, ast.symbols[key])
	}
}
