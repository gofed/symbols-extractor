package symboltable

import (
	"encoding/json"
	"fmt"
	"strings"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

// Participants:
// - symbol table (global per package, one-level table)
// - multi-level symbol table (one symbol table for each block)
// - allocated symbol table (global per package)
//   - RFE: store file:line:column as well (possible generate a link to symbol's location)

type Table struct {
	symbols map[string]map[string]*SymbolDef
	// for methods of data types
	methods map[string]map[string]*SymbolDef
	Symbols map[string][]*SymbolDef `json:"symbols"`
}

func (t *Table) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string][]*SymbolDef{
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

	t.Symbols = map[string][]*SymbolDef{
		VariableSymbol: make([]*SymbolDef, 0),
		FunctionSymbol: make([]*SymbolDef, 0),
		DataTypeSymbol: make([]*SymbolDef, 0),
	}

	t.symbols = map[string]map[string]*SymbolDef{
		VariableSymbol: make(map[string]*SymbolDef, 0),
		FunctionSymbol: make(map[string]*SymbolDef, 0),
		DataTypeSymbol: make(map[string]*SymbolDef, 0),
	}

	t.methods = make(map[string]map[string]*SymbolDef)

	for _, symbolType := range SymbolTypes {
		if objMap[symbolType] != nil {
			var m []*SymbolDef
			if err := json.Unmarshal(*objMap[symbolType], &m); err != nil {
				return err
			}

			t.Symbols[symbolType] = m

			for _, item := range m {
				if symbolType == FunctionSymbol {
					if err := t.AddFunction(item); err != nil {
						return err
					}
				} else {
					t.symbols[symbolType][item.Name] = item
				}
			}
		}
	}

	return nil
}

func NewTable() *Table {
	return &Table{
		symbols: map[string]map[string]*SymbolDef{
			VariableSymbol: make(map[string]*SymbolDef, 0),
			FunctionSymbol: make(map[string]*SymbolDef, 0),
			DataTypeSymbol: make(map[string]*SymbolDef, 0),
		},
		methods: make(map[string]map[string]*SymbolDef),
		Symbols: map[string][]*SymbolDef{
			VariableSymbol: make([]*SymbolDef, 0),
			FunctionSymbol: make([]*SymbolDef, 0),
			DataTypeSymbol: make([]*SymbolDef, 0),
		},
	}
}

func (t *Table) addSymbol(symbolType string, name string, sym *SymbolDef) error {
	if def, ko := t.symbols[symbolType][name]; ko {
		// If the symbol definition is empty, we just allocated a name for the symbol.
		// Later on, it gets populated with its definition.
		if def.Def != nil {
			return fmt.Errorf("Symbol '%s' already exists", name)
		}
		def.Def = sym.Def
		return nil
	}

	t.symbols[symbolType][name] = sym
	t.Symbols[symbolType] = append(t.Symbols[symbolType], sym)

	if symbolType == FunctionSymbol {
		if method, ok := sym.Def.(*gotypes.Method); ok {
			switch receiverExpr := method.Receiver.(type) {
			case *gotypes.Pointer:
				ident, ok := receiverExpr.Def.(*gotypes.Identifier)
				if !ok {
					return fmt.Errorf("Expected receiver as a pointer to an identifier, got %#v instead", receiverExpr.Def)
				}
				if _, ok := t.methods[ident.Def]; !ok {
					t.methods[ident.Def] = make(map[string]*SymbolDef, 0)
				}
				t.methods[ident.Def][sym.Name] = sym
				glog.Infof("Adding method %#v of data type %v", sym, ident.Def)
			case *gotypes.Identifier:
				if _, ok := t.methods[receiverExpr.Def]; !ok {
					t.methods[receiverExpr.Def] = make(map[string]*SymbolDef, 0)
				}
				t.methods[receiverExpr.Def][sym.Name] = sym
				glog.Infof("Adding method %#v of data type %v", sym, receiverExpr.Def)
			default:
				return fmt.Errorf("Receiver data type %#v of %#v not recognized", method.Receiver, method)
			}
		}
	}

	return nil
}

func (t *Table) AddVariable(sym *SymbolDef) error {
	// TODO(jchaloup): Given one can re-assign a variable (if the new type is assignable to the old one)
	//                 we need to allow update of variable's data type.
	return t.addSymbol(VariableSymbol, sym.Name, sym)
}

func (t *Table) AddDataType(sym *SymbolDef) error {
	// TODO(jchaloup): extend the SymbolDef to generate symbol names for various symbol types
	return t.addSymbol(DataTypeSymbol, sym.Name, sym)
}

func (t *Table) AddFunction(sym *SymbolDef) error {
	// Function/Method is always stored with its definition
	if def, ok := sym.Def.(*gotypes.Method); ok {
		switch rExpr := def.Receiver.(type) {
		case *gotypes.Identifier:
			glog.Infof("Storing method %q into a symbol table", strings.Join([]string{rExpr.Def, sym.Name}, "."))
			return t.addSymbol(FunctionSymbol, strings.Join([]string{rExpr.Def, sym.Name}, "."), sym)
		case *gotypes.Pointer:
			if ident, ok := rExpr.Def.(*gotypes.Identifier); ok {
				glog.Infof("Storing method %q into a symbol table", strings.Join([]string{ident.Def, sym.Name}, "."))
				return t.addSymbol(FunctionSymbol, strings.Join([]string{ident.Def, sym.Name}, "."), sym)
			}
			return fmt.Errorf("Expecting a pointer to an identifier as a receiver when adding %q method, got %#v instead", sym.Name, rExpr)
		default:
			return fmt.Errorf("Expecting an identifier or a pointer to an identifier as a receiver when adding %q method, got %#v instead", sym.Name, def)
		}
	}
	return t.addSymbol(FunctionSymbol, sym.Name, sym)
}

func (t *Table) LookupVariable(key string) (*SymbolDef, error) {
	if sym, ok := t.symbols[VariableSymbol][key]; ok {
		return sym, nil
	}
	return nil, fmt.Errorf("Variable `%v` not found", key)
}

func (t *Table) LookupVariableLikeSymbol(key string) (*SymbolDef, SymbolType, error) {
	for _, symbolType := range SymbolTypes {
		if symbolType == DataTypeSymbol {
			continue
		}
		if sym, ok := t.symbols[symbolType][key]; ok {
			return sym, SymbolType(symbolType), nil
		}
	}

	return nil, SymbolType(""), fmt.Errorf("VariableLike symbol `%v` not found", key)
}

func (t *Table) LookupFunction(key string) (*SymbolDef, error) {
	if sym, ok := t.symbols[FunctionSymbol][key]; ok {
		return sym, nil
	}
	return nil, fmt.Errorf("Function `%v` not found", key)
}

func (t *Table) LookupDataType(key string) (*SymbolDef, error) {
	if sym, ok := t.symbols[DataTypeSymbol][key]; ok {
		return sym, nil
	}
	return nil, fmt.Errorf("DataType `%v` not found", key)
}

func (t *Table) LookupMethod(datatype, methodName string) (*SymbolDef, error) {
	methods, ok := t.methods[datatype]
	if !ok {
		return nil, fmt.Errorf("Data type %q not found", datatype)
	}
	method, ok := methods[methodName]
	if !ok {
		return nil, fmt.Errorf("Method %q of data type %q not found", methodName, datatype)
	}
	return method, nil
}

func (t *Table) Exists(name string) bool {
	if _, _, err := t.Lookup(name); err == nil {
		return true
	}
	return false
}

func (t *Table) Lookup(key string) (*SymbolDef, SymbolType, error) {

	for _, symbolType := range SymbolTypes {
		if sym, ok := t.symbols[symbolType][key]; ok {
			return sym, SymbolType(symbolType), nil
		}
	}

	return nil, SymbolType(""), fmt.Errorf("Symbol `%v` not found", key)
}

var _ SymbolLookable = &Table{}
