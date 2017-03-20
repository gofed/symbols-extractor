package symboltable

import (
	"encoding/json"
	"fmt"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type Symbol interface {
	GetPosition() DeclPos
	GetName() string
	GetType() string
}

type SymbolTable interface {
	LookupSymbol(string) Symbol
	// LookupSymbolTable(string) SymbolTable
	AddSymbol(string, Symbol) error
	//TODO: iterate over table or return list of symbol names? or both?
	//          Iterate() (string, Symbol) --- or just one of them
	//      NOTE: or symbol table could provide methods below,
	//            but get full list of symbols in table should be provided
	// methods below are requested for creation of correct JSON format
	// according to used JSON schema
	//FIXME: discuss it with jchaloup;
	//        (a) change schema,
	//        (b) change structures
	//        (c) require methods below
	//          GetVariables()  []string
	//          GetFunctions()  []string
	//          GetTypes()      []string
	//        (d) current style of solution - lists are created directly
	//            inside MarshalJSON method of Package
	GetList() []string
}

type DeclPos struct {
	File string
	Line uint
}

//TODO: now all types have same structure, but changes are expected
//      over time - mainly for DeclVar, DefConst
type DeclType struct {
	Pos  DeclPos          `json:"-"`
	Name string           `json:"name"`
	Def  gotypes.DataType `json:"def"`
}

type DeclFunc struct {
	Pos  DeclPos          `json:"-"`
	Name string           `json:"name"`
	Def  gotypes.DataType `json:"def"`
}

type DeclVar struct {
	Pos  DeclPos          `json:"-"`
	Name string           `json:"name"`
	Def  gotypes.DataType `json:"def,omitempty"`
}

type DefConst struct {
	Pos  DeclPos          `json:"-"`
	Name string           `json:"name"`
	Def  gotypes.DataType `json:"def"`
}

// TODO: add JSON

func (sym *DeclType) GetPosition() DeclPos { return sym.Pos }
func (sym *DeclFunc) GetPosition() DeclPos { return sym.Pos }
func (sym *DeclVar) GetPosition() DeclPos  { return sym.Pos }
func (sym *DefConst) GetPosition() DeclPos { return sym.Pos }
func (sym *DeclType) GetName() string      { return sym.Name }
func (sym *DeclFunc) GetName() string      { return sym.Name }
func (sym *DeclVar) GetName() string       { return sym.Name }
func (sym *DefConst) GetName() string      { return sym.Name }
func (sym *DeclType) GetType() string      { return sym.Def.GetType() } //FIXME
func (sym *DeclFunc) GetType() string      { return sym.Def.GetType() } //FIXME
func (sym *DeclVar) GetType() string       { return sym.Def.GetType() } //FIXME
func (sym *DefConst) GetType() string      { return "" }                //FIXME

// JSON (de)serialisation for Decl* and DefConst types
///////////////////////////////////////////////////////////////////

//func (o *DeclType) MarshalJSON() (b []byte, error) {
//	return json.Marshal(&struct {
//}

///////////////////////////////////////////////////////////////////

//TODO: maybe something like this would be part of types.go? Used just
//      to find data type of processed symbol
//FIXME: signature... would be fine use same as json.Unmarshal
func unmarshalDataType(b []byte) (gotypes.DataType, error) {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return nil, err
	}

	var symbolType string
	if err := json.Unmarshal(*objMap["type"], &symbolType); err != nil {
		return nil, err
	}

	var symbol gotypes.DataType
	switch symbolType {
	case gotypes.FunctionType:
		var t gotypes.Function
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		//FIXME: methods and hash-names
		symbol = &t
	case gotypes.MapType:
		var t gotypes.Map
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.SliceType:
		var t gotypes.Slice
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.StructFieldsItemType:
		var t gotypes.StructFieldsItem
		if err := json.Unmarshal(b, t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.StructType:
		var t gotypes.Struct
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.SelectorType:
		var t gotypes.Selector
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.InterfaceMethodsItemType:
		var t gotypes.InterfaceMethodsItem
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.InterfaceType:
		var t gotypes.Interface
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.EllipsisType:
		var t gotypes.Ellipsis
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.ArrayType:
		var t gotypes.Array
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.IdentifierType:
		var t gotypes.Identifier
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.PointerType:
		var t gotypes.Pointer
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.MethodType:
		var t gotypes.Method
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	case gotypes.ChannelType:
		var t gotypes.Channel
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		symbol = &t
	default:
		return nil, fmt.Errorf("JSON unmarshal: Unknown symbol type '%v'", symbolType)
	}

	return symbol, nil
}

func (o *DeclType) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	var symbol gotypes.DataType
	var err error

	if err = json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err = json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	if symbol, err = unmarshalDataType(*objMap["def"]); err != nil {
		return err
	}

	o.Def = symbol
	return nil
}

func (o *DeclFunc) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	var symbol gotypes.DataType
	var err error

	if err = json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	//FIXME: methods and hash-names
	if err = json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	if symbol, err = unmarshalDataType(*objMap["def"]); err != nil {
		return err
	}

	o.Def = symbol
	return nil
}

func (o *DeclVar) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	var symbol gotypes.DataType
	var err error

	if err = json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err = json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	// variable could be represented just by name in json scheme
	// -- handle:  foo := barType
	if _, ok := objMap["def"]; ok {
		if symbol, err = unmarshalDataType(*objMap["def"]); err != nil {
			return err
		}

		o.Def = symbol
	}
	return nil
}

func (o *DefConst) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	var symbol gotypes.DataType
	var err error

	if err = json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err = json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	if symbol, err = unmarshalDataType(*objMap["def"]); err != nil {
		return err
	}

	o.Def = symbol
	return nil
}

///////////////////////////////////////////////////////////////////

type Package struct {
	Project     string
	Name        string
	Imports     []string
	SymbolTable SymbolTable
}

// JSON serialisation for Package datatype, which covers now SymbolTable itself
// according to currently used JSON scheme (golang-project-exported-api.json)
// TODO: ... it is really neccessary? it seems more like design of structures
//       or json scheme should be modified; this is written in ugly way
func (pkg *Package) MarshalJSON() (b []byte, e error) {
	var symbol Symbol
	var symID string
	types := make([]*DeclType, 0)
	funcs := make([]*DeclFunc, 0)
	vars := make([]*DeclVar, 0)
	consts := make([]*DefConst, 0)

	// assort all symbols from SymbolTable to lists above; these will be used
	// for correct serialisation
	for _, symID = range pkg.SymbolTable.GetList() {
		symbol = pkg.SymbolTable.LookupSymbol(symID)
		if symbol == nil {
			// this should not happened, but to be sure that used symbol table
			// is implemented correctly
			return nil, fmt.Errorf("Symbol with ID (%#v) has not been found.", symID)
		}

		switch symbolType := symbol.(type) {
		case *DeclType:
			types = append(types, symbolType)
		case *DeclFunc:
			funcs = append(funcs, symbolType)
		case *DeclVar:
			vars = append(vars, symbolType)
		case *DefConst:
			consts = append(consts, symbolType)
		default:
			return nil, fmt.Errorf("Unexpected symbol type: (%T).", symbolType)
		}
	}

	return json.Marshal(&struct {
		Variables []*DeclVar  `json:"variables"`
		DataTypes []*DeclType `json:"datatypes"`
		Functions []*DeclFunc `json:"functions"`
		Constants []*DefConst `json:"constants,omitempty"`
		Name      string      `json:"package"`
	}{
		Variables: vars,
		DataTypes: types,
		Functions: funcs,
		Constants: consts,
		Name:      pkg.Name,
	})
}

// Load Package and symbol table according to given JSON data
func (pkg *Package) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["package"], &pkg.Name); err != nil {
		return err
	}

	pkg.SymbolTable = make(HST)

	// load data types + functions
	//FIXME: separate functions?
	{
		var symbols []*json.RawMessage
		if err := json.Unmarshal(*objMap["datatypes"], &symbols); err != nil {
			return err
		}
		for _, rawSym := range symbols {
			sym := &DeclType{}
			if err := json.Unmarshal(*rawSym, &sym); err != nil {
				return err
			}

			if err := pkg.SymbolTable.AddSymbol(sym.Name, sym); err != nil {
				return err
			}
		}
	}

	// load variables
	{
		var symbols []*json.RawMessage
		if err := json.Unmarshal(*objMap["variables"], &symbols); err != nil {
			return err
		}
		for _, rawSym := range symbols {
			sym := &DeclVar{}
			if err := json.Unmarshal(*rawSym, &sym); err != nil {
				return err
			}

			if err := pkg.SymbolTable.AddSymbol(sym.Name, sym); err != nil {
				return err
			}
		}
	}

	// load constants
	if _, ok := objMap["constants"]; ok {
		var symbols []*json.RawMessage
		if err := json.Unmarshal(*objMap["constants"], &symbols); err != nil {
			return err
		}
		for _, rawSym := range symbols {
			sym := &DefConst{}
			if err := json.Unmarshal(*rawSym, &sym); err != nil {
				return err
			}

			if err := pkg.SymbolTable.AddSymbol(sym.Name, sym); err != nil {
				return err
			}
		}
	}

	//// load functions
	//{
	//    var symbols []*json.RawMessage
	//    if err := json.Unmarshal(*objMap["functions"], &symbols); err != nil {
	//        return err
	//    }
	//    for _, rawSym := range symbols {
	//        sym := &DeclFunc{}
	//        if err := json.Unmarshal(*rawSym, &sym); err != nil {
	//          return err
	//        }
	//
	//        if err := pkg.SymbolTable.AddSymbol(sym.Name, sym); err != nil {
	//          return err
	//        }
	//    }
	//}

	return nil
}

//TODO: This is probably temporary type, as the hashmap will be part
//       of a project type in future, which would be part of different module
type PackageMap map[string]*Package

func (pkgmap *PackageMap) MarshalJSON() (b []byte, e error) {
	pkgs := make([]*Package, len(*pkgmap))
	i := 0
	for _, pkg := range *pkgmap {
		pkgs[i] = pkg
		i++
	}

	return json.Marshal(&struct {
		//NOTE: this will be later part of the project datatype
		Pkgs []*Package `json:"packages"`
	}{
		Pkgs: pkgs,
	})
}

func (pkgmap *PackageMap) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	var rawPackages []*json.RawMessage
	if err := json.Unmarshal(*objMap["packages"], &rawPackages); err != nil {
		return err
	}

	for _, rawPkg := range rawPackages {
		var pkg Package = Package{}
		if err := json.Unmarshal(*rawPkg, &pkg); err != nil {
			return err
		}

		(*pkgmap)[pkg.Name] = &pkg
	}

	return nil

}

//TODO: or struct with maps of symbols and symboltables
//      --- another symbol tables will be needed when we will parse bodies
//          of functions
type HST map[string]Symbol

//TODO
//type AllocatedSymbolTable struct {
//}

func (st HST) LookupSymbol(name string) Symbol {
	sym, ok := st[name]
	if ok {
		return sym
	}

	return nil
}

//func (st HST) LookupSymbolTable (string) SymbolTable {
//  //TODO: finish it or remove it
//  return nil
//}

func (st HST) AddSymbol(name string, sym Symbol) error {
	if st.LookupSymbol(name) != nil {
		return fmt.Errorf("Symbol '%s' already exists in symbol %T", name, sym)
	}

	if st == nil {
		st = make(HST)
	}

	st[name] = sym
	return nil
}

//TODO: doc
//FIXME: think about internal symbols (created by parser)
func (st HST) GetList() []string {
	symbols := make([]string, len(st))
	var i uint = 0
	for symbolID, _ := range st {
		symbols[i] = symbolID
		i++
	}

	return symbols
}

// function is useless now for HST
// func (o HST) MarshalJSON() (b []byte, e error) {}

//TODO: move this to gotypes - will be used to check, whether we should add
//      another item into symtab or not, because it is builtin
//NOTE: can be overwritten inside package, so use additional logic inside
//      parser package... like: look at symtab at first and then use IsBuiltin,
//      so we will be able to create (e.g.) variable of the same name...
//       --- check it for 'int' - it works for string type
var builtinList = map[string]struct{}{
	"uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
	"int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {},
	"float32": {}, "float64": {},
	"complex64": {}, "complex128": {},
	"chan":      {},
	"interface": {},
	"struct":    {},
}

func IsBuiltin(name string) bool {
	_, ok := builtinList[name]
	return ok
}
