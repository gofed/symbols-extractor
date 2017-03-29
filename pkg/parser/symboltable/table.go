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

const (
	VariableSymbol = "variables"
	FunctionSymbol = "functions"
	DataTypeSymbol = "datatypes"
)

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
		var m []*gotypes.SymbolDef
		if err := json.Unmarshal(*objMap[symbolType], &m); err != nil {
			return err
		}

		t.Symbols[symbolType] = m

		for _, item := range m {
			t.symbols[symbolType][item.Name] = item
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
	if _, ko := t.symbols[symbolType][name]; ko {
		return fmt.Errorf("Symbol '%s' already exists", sym.Name)
	}

	t.symbols[symbolType][name] = sym
	t.Symbols[symbolType] = append(t.Symbols[symbolType], sym)

	return nil
}

func (t *Table) AddVariable(sym *gotypes.SymbolDef) error {
	return t.addSymbol(VariableSymbol, sym.Name, sym)
}

func (t *Table) AddDataType(sym *gotypes.SymbolDef) error {
	// TODO(jchaloup): extend the SymbolDef to generate symbol names for various symbol types
	return t.addSymbol(DataTypeSymbol, sym.Name, sym)
}

func (t *Table) AddFunction(sym *gotypes.SymbolDef) error {
	return t.addSymbol(FunctionSymbol, sym.Name, sym)
}

func (t *Table) Lookup(key string) (*gotypes.SymbolDef, error) {

	for _, symbolType := range SymbolTypes {
		if sym, ok := t.symbols[symbolType][key]; ok {
			fmt.Printf("Symbol found: %#v\n", sym)
			return sym, nil
		}
	}

	return nil, fmt.Errorf("Symbol `%v` not found", key)
}

// Stack is a multi-level symbol table for parsing blocks of code
type Stack struct {
	tables []*Table
	size   int
}

// NewStack creates an empty stack with no symbol table
func NewStack() *Stack {
	return &Stack{
		tables: make([]*Table, 0),
		size:   0,
	}
}

// Push pushes a new symbol table at the top of the stack
func (s *Stack) Push() {
	s.tables = append(s.tables, NewTable())
	s.size++
}

// Pop pops the top most symbol table from the stack
func (s *Stack) Pop() {
	if s.size > 0 {
		s.tables = s.tables[:s.size-1]
		s.size--
	} else {
		panic("Popping over an empty stack of symbol tables")
		// If you reached this line you are a magician
	}
}

func (s *Stack) AddVariable(sym *gotypes.SymbolDef) error {
	if s.size > 0 {
		return s.tables[s.size-1].AddVariable(sym)
	}
	return fmt.Errorf("Symbol table stack is empty")
}

func (s *Stack) AddDataType(sym *gotypes.SymbolDef) error {
	if s.size > 0 {
		return s.tables[s.size-1].AddDataType(sym)
	}
	return fmt.Errorf("Symbol table stack is empty")
}

func (s *Stack) AddFunction(sym *gotypes.SymbolDef) error {
	if s.size > 0 {
		return s.tables[s.size-1].AddFunction(sym)
	}
	return fmt.Errorf("Symbol table stack is empty")
}

// Lookup looks for the first occurrence of a symbol with the given name
func (s *Stack) Lookup(name string) (*gotypes.SymbolDef, error) {
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.size - 1; i >= 0; i-- {
		def, err := s.tables[i].Lookup(name)
		if err == nil {
			fmt.Printf("Table %v: symbol: %#v\n", i, def)
			return def, nil
		}
	}
	return nil, fmt.Errorf("Symbol %v not found", name)
}

func (s *Stack) Print() {
	for i := s.size - 1; i >= 0; i-- {
		fmt.Printf("Table %v: symbol: %#v\n", i, s.tables[i])
	}
}
