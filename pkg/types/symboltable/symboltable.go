package symboltable

import (
	"encoding/json"
	"fmt"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type SymbolTable interface {
	LookupSymbol(string) SymbolDef
	//AddSymbol(string, SymbolDef) error
	AddFunction(string, SymbolDef) error
	AddVariable(string, SymbolDef) error
	AddConstant(string, SymbolDef) error
	AddDataType(string, SymbolDef) error
}

type DeclPos struct {
	File string
	Line uint
}

//NOTE: Def can be empty for variable, but not for the others;
type SymbolDef struct {
	Pos  DeclPos          `json:"-"`
	Name string           `json:"name"`
	Def  gotypes.DataType `json:"def,omitempty"`
}


func (sym *SymbolDef) GetPosition() DeclPos { return sym.Pos }
func (sym *SymbolDef) GetName() string      { return sym.Name }
func (sym *SymbolDef) GetType() string      { return sym.Def.GetType() }

//TODO: maybe something like this would be part of types.go? Used just
//      to find data type of processed symbol
//TODO: signature... would be fine use same as json.Unmarshal
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
		//FIXME: methods and hash-names - see prefix
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

func (o *SymbolDef) UnmarshalJSON(b []byte) error {
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
	//TODO: really?
	variables map[string]*SymbolDef
	functions map[string]*SymbolDef
	datatypes map[string]*SymbolDef
	constants map[string]*SymbolDef

	Variables []*SymbolDef `json:"variables"`
	DataTypes []*SymbolDef `json:"datatypes"`
	Functions []*SymbolDef `json:"functions"`
	Constants []*SymbolDef `json:"constants,omitempty"`
	Name      string       `json:"package"`

	//Project     string
	//Imports     []string
}

func (o *Package) LookupSymbol(id string) *SymbolDef {
	if sym, ok := o.variables[id]; ok {
		return sym
	} else if sym, ok := o.functions[id]; ok {
		return sym
	} else if sym, ok := o.datatypes[id]; ok {
		return sym
	} else if sym, ok := o.constants[id]; ok {
		return sym
	}

	return nil
}

func (o *Package) Init() {
	o.variables = make(map[string]*SymbolDef)
	o.functions = make(map[string]*SymbolDef)
	o.constants = make(map[string]*SymbolDef)
	o.datatypes = make(map[string]*SymbolDef)

	o.Variables = make([]*SymbolDef, 0)
	o.Functions = make([]*SymbolDef, 0)
	o.Constants = make([]*SymbolDef, 0)
	o.DataTypes = make([]*SymbolDef, 0)
}

const errSymExists string = "Symbol '%s' already exists in symbol %T"

//TODO: IT could be replaced by AddSymbol, but I don't know reliable way
//      how to recognize each item for current structure and methods
//      of SymbolDef; think about that

func (o *Package) AddFunction(id string, sym *SymbolDef) error {
	if _, ko := o.functions[id]; ko {
		return fmt.Errorf(errSymExists, sym.Name, sym)
	}

	o.functions[id] = sym
	o.Functions = append(o.Functions, sym)
	return nil
}

func (o *Package) AddVariable(id string, sym *SymbolDef) error {
	if _, ko := o.functions[id]; ko {
		return fmt.Errorf(errSymExists, sym.Name, sym)
	}

	o.variables[id] = sym
	o.Variables = append(o.Variables, sym)
	return nil
}

func (o *Package) AddDataType(id string, sym *SymbolDef) error {
	if _, ko := o.datatypes[id]; ko {
		return fmt.Errorf(errSymExists, sym.Name, sym)
	}

	o.datatypes[id] = sym
	o.DataTypes = append(o.DataTypes, sym)
	return nil
}

func (o *Package) AddConstant(id string, sym *SymbolDef) error {
	if _, ko := o.functions[id]; ko {
		return fmt.Errorf(errSymExists, sym.Name, sym)
	}

	o.constants[id] = sym
	o.Constants = append(o.Constants, sym)
	return nil
}

// Load Package and symbol table according to given JSON data
//NOTE: Unmarshal is required as the default func loads only public data
//NOTE:  -- another way is create func that fill internal maps when symbols
//          are already loaded - but it is not clear solution
func (pkg *Package) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["package"], &pkg.Name); err != nil {
		return err
	}

	// init pkg because of internal map structures
	pkg.Init()

	// load data types + functions
	//FIXME: separate functions?
	{
		var symbols []*json.RawMessage
		if err := json.Unmarshal(*objMap["datatypes"], &symbols); err != nil {
			return err
		}
		for _, rawSym := range symbols {
			sym := &SymbolDef{}
			if err := json.Unmarshal(*rawSym, &sym); err != nil {
				return err
			}

			if err := pkg.AddDataType(sym.Name, sym); err != nil {
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
			sym := &SymbolDef{}
			if err := json.Unmarshal(*rawSym, &sym); err != nil {
				return err
			}

			if err := pkg.AddVariable(sym.Name, sym); err != nil {
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
			sym := &SymbolDef{}
			if err := json.Unmarshal(*rawSym, &sym); err != nil {
				return err
			}

			if err := pkg.AddConstant(sym.Name, sym); err != nil {
				return err
			}
		}
	}

	// load functions
	{
		var symbols []*json.RawMessage
		if err := json.Unmarshal(*objMap["functions"], &symbols); err != nil {
			return err
		}
		for _, rawSym := range symbols {
			sym := &SymbolDef{}
			if err := json.Unmarshal(*rawSym, &sym); err != nil {
				return err
			}

			//FIXME method name - see prefix
			if err := pkg.AddFunction(sym.Name, sym); err != nil {
				return err
			}
		}
	}

	return nil
}

// EXPECTED EOF

//TODO: This is temporary type, as the hashmap will be part
//       of a project type in future, which would be part of different module
//TODO: Move it to separate package - project maybe
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

//TODO: move this to the parser - will be used to check, whether we should add
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
	"chan": {},
	//"interface": {},
	//"struct":    {},
	"string": {},
}

func IsBuiltin(name string) bool {
	_, ok := builtinList[name]
	return ok
}
