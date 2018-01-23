package tables

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/symbols"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

// Participants:
// - symbol table (global per package, one-level table)
// - multi-level symbol table (one symbol table for each block)
// - allocated symbol table (global per package)
//   - RFE: store file:line:column as well (possible generate a link to symbol's location)

type Table struct {
	symbols map[string]map[string]*symbols.SymbolDef
	// for methods of data types
	methods    map[string]map[string]*symbols.SymbolDef
	Symbols    map[string][]*symbols.SymbolDef `json:"symbols"`
	PackageQID string                          `json:"qid"`
	Imports    []string                        `json:"imports"`
}

func (t *Table) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	t.Symbols = map[string][]*symbols.SymbolDef{
		symbols.VariableSymbol: make([]*symbols.SymbolDef, 0),
		symbols.FunctionSymbol: make([]*symbols.SymbolDef, 0),
		symbols.DataTypeSymbol: make([]*symbols.SymbolDef, 0),
	}

	t.symbols = map[string]map[string]*symbols.SymbolDef{
		symbols.VariableSymbol: make(map[string]*symbols.SymbolDef, 0),
		symbols.FunctionSymbol: make(map[string]*symbols.SymbolDef, 0),
		symbols.DataTypeSymbol: make(map[string]*symbols.SymbolDef, 0),
	}

	t.methods = make(map[string]map[string]*symbols.SymbolDef)

	if err := json.Unmarshal(*objMap["qid"], &t.PackageQID); err != nil {
		return err
	}

	if _, ok := objMap["imports"]; ok && objMap["imports"] != nil {
		if err := json.Unmarshal(*objMap["imports"], &t.Imports); err != nil {
			return err
		}
	}

	var symbolsObjMap map[string]*json.RawMessage

	if err := json.Unmarshal(*objMap["symbols"], &symbolsObjMap); err != nil {
		return err
	}

	for _, symbolType := range symbols.SymbolTypes {
		if symbolsObjMap[symbolType] != nil {
			var m []*symbols.SymbolDef
			if err := json.Unmarshal(*symbolsObjMap[symbolType], &m); err != nil {
				return err
			}

			t.Symbols[symbolType] = m

			for _, item := range m {
				if symbolType == symbols.FunctionSymbol {
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
		symbols: map[string]map[string]*symbols.SymbolDef{
			symbols.VariableSymbol: make(map[string]*symbols.SymbolDef, 0),
			symbols.FunctionSymbol: make(map[string]*symbols.SymbolDef, 0),
			symbols.DataTypeSymbol: make(map[string]*symbols.SymbolDef, 0),
		},
		methods: make(map[string]map[string]*symbols.SymbolDef),
		Symbols: map[string][]*symbols.SymbolDef{
			symbols.VariableSymbol: make([]*symbols.SymbolDef, 0),
			symbols.FunctionSymbol: make([]*symbols.SymbolDef, 0),
			symbols.DataTypeSymbol: make([]*symbols.SymbolDef, 0),
		},
	}
}

func (t *Table) addSymbol(symbolType string, name string, sym *symbols.SymbolDef) error {
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

	if symbolType == symbols.FunctionSymbol {
		if method, ok := sym.Def.(*gotypes.Method); ok {
			switch receiverExpr := method.Receiver.(type) {
			case *gotypes.Pointer:
				ident, ok := receiverExpr.Def.(*gotypes.Identifier)
				if !ok {
					return fmt.Errorf("Expected receiver as a pointer to an identifier, got %#v instead", receiverExpr.Def)
				}
				if _, ok := t.methods[ident.Def]; !ok {
					t.methods[ident.Def] = make(map[string]*symbols.SymbolDef, 0)
				}
				t.methods[ident.Def][sym.Name] = sym
				glog.V(2).Infof("Adding method %#v of data type %v", sym, ident.Def)
			case *gotypes.Identifier:
				if _, ok := t.methods[receiverExpr.Def]; !ok {
					t.methods[receiverExpr.Def] = make(map[string]*symbols.SymbolDef, 0)
				}
				t.methods[receiverExpr.Def][sym.Name] = sym
				glog.V(2).Infof("Adding method %#v of data type %v", sym, receiverExpr.Def)
			default:
				return fmt.Errorf("Receiver data type %#v of %#v not recognized", method.Receiver, method)
			}
		}
	}

	return nil
}

func (t *Table) AddVariable(sym *symbols.SymbolDef) error {
	// TODO(jchaloup): Given one can re-assign a variable (if the new type is assignable to the old one)
	//                 we need to allow update of variable's data type.
	return t.addSymbol(symbols.VariableSymbol, sym.Name, sym)
}

func (t *Table) AddDataType(sym *symbols.SymbolDef) error {
	// TODO(jchaloup): extend the SymbolDef to generate symbol names for various symbol types
	return t.addSymbol(symbols.DataTypeSymbol, sym.Name, sym)
}

func (t *Table) AddFunction(sym *symbols.SymbolDef) error {
	// Function/Method is always stored with its definition
	if def, ok := sym.Def.(*gotypes.Method); ok {
		switch rExpr := def.Receiver.(type) {
		case *gotypes.Identifier:
			glog.V(2).Infof("Storing method %q into a symbol table", strings.Join([]string{rExpr.Def, sym.Name}, "."))
			return t.addSymbol(symbols.FunctionSymbol, strings.Join([]string{rExpr.Def, sym.Name}, "."), sym)
		case *gotypes.Pointer:
			if ident, ok := rExpr.Def.(*gotypes.Identifier); ok {
				glog.V(2).Infof("Storing method %q into a symbol table", strings.Join([]string{ident.Def, sym.Name}, "."))
				return t.addSymbol(symbols.FunctionSymbol, strings.Join([]string{ident.Def, sym.Name}, "."), sym)
			}
			return fmt.Errorf("Expecting a pointer to an identifier as a receiver when adding %q method, got %#v instead", sym.Name, rExpr)
		default:
			return fmt.Errorf("Expecting an identifier or a pointer to an identifier as a receiver when adding %q method, got %#v instead", sym.Name, def)
		}
	}
	return t.addSymbol(symbols.FunctionSymbol, sym.Name, sym)
}

func (t *Table) LookupVariable(key string) (*symbols.SymbolDef, error) {
	if sym, ok := t.symbols[symbols.VariableSymbol][key]; ok {
		return sym, nil
	}
	return nil, fmt.Errorf("Variable `%v` not found", key)
}

func (t *Table) LookupVariableLikeSymbol(key string) (*symbols.SymbolDef, symbols.SymbolType, error) {
	for _, symbolType := range symbols.SymbolTypes {
		if symbolType == symbols.DataTypeSymbol {
			continue
		}
		if sym, ok := t.symbols[symbolType][key]; ok {
			return sym, symbols.SymbolType(symbolType), nil
		}
	}

	return nil, symbols.SymbolType(""), fmt.Errorf("VariableLike symbol `%v` not found", key)
}

func (t *Table) LookupFunction(key string) (*symbols.SymbolDef, error) {
	if sym, ok := t.symbols[symbols.FunctionSymbol][key]; ok {
		return sym, nil
	}
	return nil, fmt.Errorf("Function `%v` not found", key)
}

func (t *Table) LookupDataType(key string) (*symbols.SymbolDef, error) {
	if sym, ok := t.symbols[symbols.DataTypeSymbol][key]; ok {
		return sym, nil
	}
	return nil, fmt.Errorf("DataType `%v` not found", key)
}

func (t *Table) LookupMethod(datatype, methodName string) (*symbols.SymbolDef, error) {
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

func (t *Table) LookupAllMethods(datatype string) (map[string]*symbols.SymbolDef, error) {
	methods, ok := t.methods[datatype]
	if !ok {
		return nil, fmt.Errorf("Data type %q not found", datatype)
	}
	return methods, nil
}

func (t *Table) Exists(name string) bool {
	if _, _, err := t.Lookup(name); err == nil {
		return true
	}
	return false
}

func (t *Table) Lookup(key string) (*symbols.SymbolDef, symbols.SymbolType, error) {

	for _, symbolType := range symbols.SymbolTypes {
		if sym, ok := t.symbols[symbolType][key]; ok {
			return sym, symbols.SymbolType(symbolType), nil
		}
	}

	return nil, symbols.SymbolType(""), fmt.Errorf("Symbol `%v` not found", key)
}

var _ symbols.SymbolLookable = &Table{}
