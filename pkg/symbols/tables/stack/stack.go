package stack

import (
	"encoding/json"
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables"
	"github.com/golang/glog"
)

// Stack is a multi-level symbol table for parsing blocks of code
type Stack struct {
	Tables  []*tables.Table               `json:"tables"`
	Size    int                           `json:"size"`
	Imports map[string]*symbols.SymbolDef `json:"imports"`
}

// NewStack creates an empty stack with no symbol table
func New() *Stack {
	return &Stack{
		Tables: []*tables.Table{
			tables.NewTable(),
		},
		Size:    0,
		Imports: make(map[string]*symbols.SymbolDef, 0),
	}
}

// Push pushes a new symbol table at the top of the stack
func (s *Stack) Push() {
	s.Tables = append(s.Tables, tables.NewTable())
	s.Size++
	glog.Infof("Pushing to symbol table stack %v\n", s.Size)
}

// Pop pops the top most symbol table from the stack
func (s *Stack) Pop() {
	if s.Size > 0 {
		s.Tables = s.Tables[:s.Size-1]
		s.Size--
	} else {
		panic("Popping over an empty stack of symbol tables")
		// If you reached this line you are a magician
	}
	glog.Infof("Popping symbol table stack %v\n", s.Size)
}

func (s *Stack) AddImport(sym *symbols.SymbolDef) error {
	s.Imports[sym.Name] = sym
	return nil
}

func (s *Stack) AddVariable(sym *symbols.SymbolDef) error {
	if s.Size > 0 {
		glog.Infof("====Adding %v variable at level %v\n", sym.Name, s.Size-1)
		// In order to distinguish between global and local variable
		// all local variable are packageless
		if s.Size > 1 {
			sym.Package = ""
		}
		return s.Tables[s.Size-1].AddVariable(sym)
	}
	return fmt.Errorf("Symbol table stack is empty")
}

func (s *Stack) AddDataType(sym *symbols.SymbolDef) error {
	if s.Size > 0 {
		glog.Infof("====Adding %#v datatype at level %v\n", sym, s.Size-1)
		return s.Tables[s.Size-1].AddDataType(sym)
	}
	return fmt.Errorf("Symbol table stack is empty")
}

func (s *Stack) AddFunction(sym *symbols.SymbolDef) error {
	glog.Infof("====Adding function %q as: %#v", sym.Name, sym.Def)
	if s.Size > 0 {
		return s.Tables[s.Size-1].AddFunction(sym)
	}
	return fmt.Errorf("Symbol table stack is empty")
}

func (s *Stack) LookupVariable(name string) (*symbols.SymbolDef, error) {
	glog.Infof("====Looking up a variable %q", name)
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.Size - 1; i >= 0; i-- {
		def, err := s.Tables[i].LookupVariable(name)
		if err == nil {
			return def, nil
		}
	}
	// if the variable is not found, check the qids
	if def, ok := s.Imports[name]; ok {
		return def, nil
	}
	return nil, fmt.Errorf("Symbol %v not found", name)
}

func (s *Stack) LookupMethod(datatype, methodName string) (*symbols.SymbolDef, error) {
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.Size - 1; i >= 0; i-- {
		def, err := s.Tables[i].LookupMethod(datatype, methodName)
		if err == nil {
			return def, nil
		}
	}
	return nil, fmt.Errorf("Method %q of data type %q not found", methodName, datatype)
}

func (s *Stack) LookupAllMethods(datatype string) (map[string]*symbols.SymbolDef, error) {
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.Size - 1; i >= 0; i-- {
		defs, err := s.Tables[i].LookupAllMethods(datatype)
		if err == nil {
			return defs, nil
		}
	}
	return nil, fmt.Errorf("Methods of data type %q not found", datatype)
}

func (s *Stack) LookupFunction(name string) (*symbols.SymbolDef, error) {
	glog.Infof("====Looking up a function %q", name)
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.Size - 1; i >= 0; i-- {
		def, err := s.Tables[i].LookupFunction(name)
		if err == nil {
			return def, nil
		}
	}
	return nil, fmt.Errorf("Symbol %v not found", name)
}

func (s *Stack) LookupDataType(name string) (*symbols.SymbolDef, error) {
	glog.Infof("====Looking up a data type %q, s.Size = %v", name, s.Size)
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.Size - 1; i >= 0; i-- {
		def, err := s.Tables[i].LookupDataType(name)
		if err == nil {
			return def, nil
		}
	}
	return nil, fmt.Errorf("Symbol %v not found", name)
}

func (s *Stack) LookupVariableLikeSymbol(name string) (*symbols.SymbolDef, symbols.SymbolType, error) {
	glog.Infof("====Looking up a variablelike %q", name)
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.Size - 1; i >= 0; i-- {
		def, st, err := s.Tables[i].LookupVariableLikeSymbol(name)
		if err == nil {
			return def, st, nil
		}
	}
	// if the variable is not found, check the qids
	if def, ok := s.Imports[name]; ok {
		return def, symbols.VariableSymbol, nil
	}
	return nil, symbols.SymbolType(""), fmt.Errorf("VariableLike Symbol %v not found", name)
}

func (s *Stack) Exists(name string) bool {
	if _, _, err := s.Lookup(name); err == nil {
		return true
	}
	if _, ok := s.Imports[name]; ok {
		return true
	}
	return false
}

// Lookup looks for the first occurrence of a symbol with the given name
func (s *Stack) Lookup(name string) (*symbols.SymbolDef, symbols.SymbolType, error) {
	// The top most item on the stack is the right most item in the simpleSlice
	for i := s.Size - 1; i >= 0; i-- {
		def, st, err := s.Tables[i].Lookup(name)
		if err == nil {
			return def, st, nil
		}
	}
	return nil, symbols.SymbolType(""), fmt.Errorf("Symbol %v not found", name)
}

func (s *Stack) Reset() error {
	s.Tables = s.Tables[:1]
	s.Size = 1

	return nil
}

func (s *Stack) CurrentLevel() int {
	return s.Size - 1
}

// Table gets a symbol table at given level
// Level 0 corresponds to the file level symbol table (the top most block)
func (s *Stack) Table(level int) (*tables.Table, error) {
	if level < 0 || s.Size-1 < level {
		return nil, fmt.Errorf("No symbol table found for level %v", level)
	}
	return s.Tables[level], nil
}

func (s *Stack) Print() {
	for i := s.Size - 1; i >= 0; i-- {
		fmt.Printf("Table %v: symbol: %#v\n", i, s.Tables[i])
	}
}

func (s *Stack) Json() {
	x, _ := json.Marshal(s)
	fmt.Print(string(x))
}

func (s *Stack) PrintTop() {
	fmt.Printf("TableSymbols: %#v\n", s.Tables[s.Size-1])
}

var _ symbols.SymbolLookable = &Stack{}
