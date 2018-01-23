package accessors

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

type Accessor struct {
	// package name
	packageName string
	// per package symbol table
	symbolTable symbols.SymbolLookable
	// per subset of packages symbol table
	globalSymbolTable *global.Table
}

func NewAccessor(globalSymbolTable *global.Table) *Accessor {
	return &Accessor{
		globalSymbolTable: globalSymbolTable,
	}
}

func (a *Accessor) SetCurrentTable(packageName string, symbolTable symbols.SymbolLookable) *Accessor {
	a.packageName = packageName
	a.symbolTable = symbolTable

	return a
}

func (a *Accessor) CurrentTable() (string, symbols.SymbolLookable) {
	return a.packageName, a.symbolTable
}

func (a *Accessor) GetBuiltin(name string) (*symbols.SymbolDef, symbols.SymbolType, error) {
	table, err := a.globalSymbolTable.Lookup("builtin")
	if err != nil {
		return nil, "", err
	}
	return table.Lookup(name)
}

func (a *Accessor) GetBuiltinDataType(name string) (*symbols.SymbolDef, symbols.SymbolLookable, error) {
	table, err := a.globalSymbolTable.Lookup("builtin")
	if err != nil {
		return nil, nil, err
	}
	def, err := table.LookupDataType(name)
	return def, table, err
}

func (a *Accessor) IsBuiltin(name string) bool {
	table, err := a.globalSymbolTable.Lookup("builtin")
	if err != nil {
		glog.Warning(err)
		return false
	}
	if _, _, err := table.Lookup(name); err == nil {
		return true
	}
	return false
}

func (a *Accessor) IsBuiltinConstantType(name string) bool {
	if a.IsIntegral(name) {
		return true
	}
	if a.IsFloating(name) {
		return true
	}
	switch name {
	case "complex64", "complex128":
		return true
	case "string":
		return true
	case "bool":
		return true
	}
	return false
}

func (a *Accessor) IsIntegral(name string) bool {
	switch name {
	case "int", "int8", "int16", "int32", "int64":
		return true
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "rune", "byte":
		return true
	case "uintptr":
		return true
	default:
		return false
	}
}

func (a *Accessor) IsCIntegral(name string) bool {
	// TODO(jchaloup): read the list from a external resource
	switch name {
	case "int", "long", "size_t", "uintptr_t":
		return true
	default:
		return false
	}
}

func (a *Accessor) IsUintegral(name string) bool {
	switch name {
	case "int", "int8", "int16", "int32", "int64":
		return false
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "rune", "byte":
		return true
	case "uintptr":
		return true
	default:
		return false
	}
}

func (a *Accessor) IsFloating(name string) bool {
	return name == "float32" || name == "float64"
}

func (a *Accessor) IsComplex(name string) bool {
	return name == "complex64" || name == "complex128"
}

func (a *Accessor) IsPointerType(def gotypes.DataType) bool {
	switch def.GetType() {
	case gotypes.PointerType,
		gotypes.InterfaceType,
		gotypes.NilType,
		gotypes.FunctionType,
		gotypes.MethodType,
		gotypes.ArrayType,
		gotypes.SliceType,
		gotypes.MapType,
		gotypes.ChannelType:
		return true
	}
	return false
}

// Lookup retrieves a definition of identifier ident
func (a *Accessor) Lookup(ident *gotypes.Identifier) (*symbols.SymbolDef, symbols.SymbolType, error) {
	if ident.Package == "" {
		return nil, "", fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		return a.symbolTable.Lookup(ident.Def)
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, "", err
	}
	return table.Lookup(ident.Def)
}

func (a *Accessor) LookupMethod(ident *gotypes.Identifier, method string) (*symbols.SymbolDef, error) {
	if ident.Package == "" {
		return nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		return a.symbolTable.LookupMethod(ident.Def, method)
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, err
	}
	return table.LookupMethod(ident.Def, method)
}

func (a *Accessor) LookupAllMethods(ident *gotypes.Identifier) (map[string]*symbols.SymbolDef, error) {
	if ident.Package == "" {
		return nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		return a.symbolTable.LookupAllMethods(ident.Def)
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, err
	}
	return table.LookupAllMethods(ident.Def)
}

func (a *Accessor) LookupDataType(ident *gotypes.Identifier) (*symbols.SymbolDef, symbols.SymbolLookable, error) {
	if ident.Package == "" {
		return nil, nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		def, err := a.symbolTable.LookupDataType(ident.Def)
		return def, a.symbolTable, err
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, nil, err
	}
	def, err := table.LookupDataType(ident.Def)
	return def, table, err
}

func (a *Accessor) LookupVariableLike(ident *gotypes.Identifier) (*symbols.SymbolDef, symbols.SymbolLookable, error) {
	if ident.Package == "" {
		return nil, nil, fmt.Errorf("Identifier %#v does not set its Package field", ident)
	}
	if ident.Package == a.packageName {
		def, _, err := a.symbolTable.LookupVariableLikeSymbol(ident.Def)
		return def, a.symbolTable, err
	}
	table, err := a.globalSymbolTable.Lookup(ident.Package)
	if err != nil {
		return nil, nil, err
	}
	def, _, err := table.LookupVariableLikeSymbol(ident.Def)
	return def, table, err
}

func (a *Accessor) RetrieveQid(qidprefix gotypes.DataType, item *ast.Ident) (symbols.SymbolLookable, *symbols.SymbolDef, error) {
	// qid.structtype expected
	qid, ok := qidprefix.(*gotypes.Packagequalifier)
	if !ok {
		return nil, nil, fmt.Errorf("Expecting a qid.id when retrieving a symbol from a selector expression")
	}
	glog.V(2).Infof("Trying to retrieve a symbol %#v from package %v\n", item.String(), qid.Path)
	qidst, err := a.globalSymbolTable.Lookup(qid.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid.Path, err)
	}

	dataTypeDef, _, piErr := qidst.Lookup(item.String())
	if piErr != nil {
		return nil, nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", item.String(), qid.Path, piErr)
	}
	return qidst, dataTypeDef, nil
}

func (a *Accessor) RetrieveQidDataType(qidprefix gotypes.DataType, item *ast.Ident) (symbols.SymbolLookable, *symbols.SymbolDef, error) {
	// qid.structtype expected
	qid, ok := qidprefix.(*gotypes.Packagequalifier)
	if !ok {
		return nil, nil, fmt.Errorf("Expecting a qid.id when retrieving a symbol from a selector expression")
	}
	glog.V(2).Infof("Trying to retrieve a symbol %#v from package %v\n", item.String(), qid.Path)
	qidst, err := a.globalSymbolTable.Lookup(qid.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid.Path, err)
	}
	dataTypeDef, piErr := qidst.LookupDataType(item.String())
	if piErr != nil {
		return nil, nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", item.String(), qid.Path, piErr)
	}
	return qidst, dataTypeDef, nil
}

func (a *Accessor) IsDataTypeInterface(typeDef gotypes.DataType) (bool, error) {
	var symbolDef *symbols.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Interface:
		return true, nil
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return false, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		if typeDefType.Package == "builtin" {
			// error interface
			if typeDefType.Def == "error" {
				return true, nil
			}
			return false, nil
		}
		def, _, err := a.LookupDataType(typeDefType)
		if err != nil {
			return false, err
		}
		symbolDef = def
	default:
		return false, nil
	}
	if symbolDef.Def == nil {
		return false, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
	}
	if symbolDef.Package == "C" {
		return false, nil
	}
	return a.IsDataTypeInterface(symbolDef.Def)
}

func (a *Accessor) FindFirstNonidDataType(typeDef gotypes.DataType) (gotypes.DataType, error) {
	glog.V(2).Infof("FindFirstNonidDataType: %#v", typeDef)
	var symbolDef *symbols.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		if typeDefType.Package == "builtin" {
			return typeDef, nil
		}
		def, _, err := a.LookupDataType(typeDefType)
		if err != nil {
			return nil, err
		}
		symbolDef = def
	default:
		return typeDef, nil
	}
	if symbolDef.Def == nil {
		return nil, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
	}
	if symbolDef.Package == "C" {
		return typeDef, nil
	}
	return a.FindFirstNonidDataType(symbolDef.Def)
}

func (a *Accessor) FindFirstNonidVariable(typeDef gotypes.DataType) (gotypes.DataType, error) {
	glog.V(2).Infof("FindFirstNonidVariable: %#v", typeDef)
	var symbolDef *symbols.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		if typeDefType.Package == "builtin" {
			return typeDef, nil
		}
		def, _, err := a.LookupVariableLike(typeDefType)
		if err != nil {
			return nil, err
		}
		symbolDef = def
	default:
		return typeDef, nil
	}
	if symbolDef.Def == nil {
		return nil, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
	}
	if symbolDef.Package == "C" {
		return typeDef, nil
	}
	return a.FindFirstNonidVariable(symbolDef.Def)
}

func (a *Accessor) FindFirstNonIdSymbol(typeDef gotypes.DataType) (gotypes.DataType, error) {
	glog.V(2).Infof("FindFirstNonidSymbol: %#v", typeDef)
	var symbolDef *symbols.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		if typeDefType.Package == "builtin" {
			return typeDef, nil
		}
		def, _, err := a.Lookup(typeDefType)
		if err != nil {
			return nil, err
		}
		symbolDef = def
	default:
		return typeDef, nil
	}
	if symbolDef.Def == nil {
		return nil, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
	}
	if symbolDef.Package == "C" {
		return typeDef, nil
	}
	return a.FindFirstNonIdSymbol(symbolDef.Def)
}

func (a *Accessor) FindLastIdDataType(typeDef gotypes.DataType) (gotypes.DataType, error) {
	var symbolDef *symbols.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		if typeDefType.Package == "builtin" {
			return typeDef, nil
		}
		def, _, err := a.LookupDataType(typeDefType)
		if err != nil {
			return nil, err
		}
		symbolDef = def
	default:
		return nil, fmt.Errorf("Expected identifier, got %#v instead", typeDef)
	}
	if symbolDef.Def == nil {
		return nil, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
	}
	return a.FindFirstNonidDataType(symbolDef.Def)
}

// I need to end up with:
// - anonymous struct definition 		=> "" + struct definition
// - typed struct (its identifier) 	=> qid.id + struct definition
// - pointer (and its type)					=> just it is a pointer for now
// - builtin												=> just the type name

type UnderlyingType struct {
	Def        gotypes.DataType
	SymbolType string
	Id         string
	Package    string
}

func (a *Accessor) ResolveToUnderlyingType(typeDef gotypes.DataType) (*UnderlyingType, error) {
	glog.V(2).Infof("ResolveToUnderlyingType: %#v\n", typeDef)
	switch d := typeDef.(type) {
	// anonymous struct
	case *gotypes.Struct:
		return &UnderlyingType{
			Def:        typeDef,
			SymbolType: gotypes.StructType,
		}, nil
	// interfaces
	case *gotypes.Interface:
		return &UnderlyingType{
			SymbolType: gotypes.InterfaceType,
		}, nil
	// nil is still special
	case *gotypes.Nil:
		return &UnderlyingType{
			SymbolType: gotypes.NilType,
		}, nil
	case *gotypes.Pointer:
		return &UnderlyingType{
			SymbolType: gotypes.PointerType,
			Def:        typeDef,
		}, nil
	// any pointer type
	case *gotypes.Array,
		*gotypes.Channel,
		// Function, Slice and Map only comparable with nil
		*gotypes.Function,
		*gotypes.Method,
		*gotypes.Slice,
		*gotypes.Map:
		return &UnderlyingType{
			SymbolType: gotypes.PointerType,
			Def:        typeDef,
		}, nil
	// identifier
	case *gotypes.Builtin:
		return &UnderlyingType{
			Def:        typeDef,
			SymbolType: gotypes.BuiltinType,
		}, nil
	case *gotypes.Identifier, *gotypes.Constant:
		var ident *gotypes.Identifier
		if constant, ok := d.(*gotypes.Constant); ok {
			ident = &gotypes.Identifier{
				Package: constant.Package,
				Def:     constant.Def,
			}
		} else {
			ident = d.(*gotypes.Identifier)
		}

		if ident.Package == "builtin" {
			if ident.Def == "error" {
				def, _, err := a.Lookup(ident)
				if err != nil {
					return nil, err
				}
				return &UnderlyingType{
					SymbolType: gotypes.InterfaceType,
					Package:    "builtin",
					Id:         "error",
					Def:        def.Def,
				}, nil
			}
			return &UnderlyingType{
				Def:        ident,
				SymbolType: gotypes.BuiltinType,
			}, nil
		}
		if ident.Package == "C" {
			return &UnderlyingType{
				Def:        ident,
				SymbolType: gotypes.IdentifierType,
			}, nil
		}
		def, _, err := a.LookupDataType(ident)
		if err != nil {
			return nil, err
		}
		if def.Def == nil {
			return nil, fmt.Errorf("Symbol %q not yet fully processed", def.Name)
		}
		switch def.Def.(type) {
		case *gotypes.Struct:
			return &UnderlyingType{
				Def:        def.Def,
				SymbolType: gotypes.StructType,
				Id:         ident.Def,
				Package:    ident.Package,
			}, nil
		case *gotypes.Interface:
			return &UnderlyingType{
				Def:        def.Def,
				SymbolType: gotypes.InterfaceType,
				Id:         ident.Def,
				Package:    ident.Package,
			}, nil
		}

		return a.ResolveToUnderlyingType(def.Def)
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(d.Prefix, &ast.Ident{Name: d.Item})
		if err != nil {
			return nil, err
		}
		if def.Def == nil {
			return nil, fmt.Errorf("Symbol %q not yet fully processed", def.Name)
		}

		switch def.Def.(type) {
		case *gotypes.Struct:
			return &UnderlyingType{
				Def:        def.Def,
				SymbolType: gotypes.StructType,
				Id:         d.Item,
				Package:    d.Prefix.(*gotypes.Packagequalifier).Path,
			}, nil
		case *gotypes.Interface:
			return &UnderlyingType{
				Def:        def.Def,
				SymbolType: gotypes.InterfaceType,
				Id:         d.Item,
				Package:    d.Prefix.(*gotypes.Packagequalifier).Path,
			}, nil
		}
		return a.ResolveToUnderlyingType(def.Def)
	default:
		panic(fmt.Errorf("Unrecognized underlying type %#v", typeDef))
	}
}

func (a *Accessor) TypeToSimpleBuiltin(typeDef gotypes.DataType) (*gotypes.Identifier, error) {
	builtin, err := a.FindFirstNonidDataType(typeDef)
	if err != nil {
		return nil, fmt.Errorf("Expected a type %#v to have simple buitin underlying type: %v", typeDef, err)
	}
	if ident, ok := builtin.(*gotypes.Identifier); ok {
		if ident.Package == "C" {
			// just assume it is correctly used
			return ident, nil
		}
	}
	if sel, ok := builtin.(*gotypes.Selector); ok {
		qid := sel.Prefix.(*gotypes.Packagequalifier)
		if qid.Path == "C" {
			return &gotypes.Identifier{
				Package: "C",
				Def:     sel.Item,
			}, nil
		}
	}
	// pointer to C?
	if a.IsPointerType(builtin) {
		return &gotypes.Identifier{
			Package: "<builtin>",
			// better idea?
			Def: "pointer",
		}, nil
	}
	if builtin.GetType() != gotypes.IdentifierType {
		return nil, fmt.Errorf("Expected a type %#v to have simple buitin underlying type. Got %v instead", typeDef, builtin.GetType())
	}
	ident := builtin.(*gotypes.Identifier)
	if ident.Package != "builtin" || !a.IsBuiltinConstantType(ident.Def) {
		return nil, fmt.Errorf("Expected identifier %#v to have simple buitin underlying type. Got %#v instead", typeDef, ident)
	}
	return ident, nil
}

func (a *Accessor) CToGoUnderlyingType(cIdent *gotypes.Identifier) (*gotypes.Identifier, error) {
	if cIdent.Package != "C" {
		return nil, fmt.Errorf("Expected C identifier, got %v identifier instead", cIdent.Package)
	}

	switch cIdent.Def {
	case "size_t":
		return &gotypes.Identifier{Package: "builtin", Def: "int"}, nil
	}

	return nil, fmt.Errorf("Unrecognized C identifier %q", cIdent.Def)
}

// TODO(jchaloup): return symbol table of the returned data type
func (a *Accessor) RetrieveInterfaceMethod(pkgsymboltable symbols.SymbolLookable, interfaceDefsymbol *symbols.SymbolDef, method string) (*FieldAttribute, error) {
	glog.V(2).Infof("Retrieving Interface method %q from %#v\n", method, interfaceDefsymbol)
	if interfaceDefsymbol.Def.GetType() != gotypes.InterfaceType {
		return nil, fmt.Errorf("Trying to retrieve a %v method from a non-interface data type: %#v", method, interfaceDefsymbol.Def)
	}

	type embeddedInterfacesItem struct {
		symbolTable symbols.SymbolLookable
		symbolDef   *symbols.SymbolDef
	}
	var embeddedInterfaces []embeddedInterfacesItem

	var methodItem *gotypes.InterfaceMethodsItem

	for _, item := range interfaceDefsymbol.Def.(*gotypes.Interface).Methods {
		methodName := item.Name
		// anonymous field (can be embedded struct as well)
		if methodName == "" {
			// Given a variable can be of interface data type,
			// embedded interface needs to be checked as well.
			if item.Def == nil {
				return nil, fmt.Errorf("Symbol of embedded interface not fully processed")
			}
			itemExpr := item.Def
			if pointerExpr, isPointer := item.Def.(*gotypes.Pointer); isPointer {
				itemExpr = pointerExpr.Def
			}

			itemSymbolTable := pkgsymboltable
			if itemExpr.GetType() == gotypes.BuiltinType {
				itemExpr = &gotypes.Identifier{
					Def:     itemExpr.(*gotypes.Builtin).Def,
					Package: "builtin",
				}
				table, err := a.globalSymbolTable.Lookup("builtin")
				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve a symbol table for 'builtin' package: %v", err)
				}
				itemSymbolTable = table
			}

			switch fieldType := itemExpr.(type) {
			case *gotypes.Identifier:
				if fieldType.Def == method {
					methodItem = &item
					break
				}
				var def *symbols.SymbolDef
				var err error
				// check if the field is an embedded struct
				if fieldType.Package == "builtin" {
					table, e := a.globalSymbolTable.Lookup("builtin")
					if e != nil {
						return nil, fmt.Errorf("Unable to retrieve a symbol table for 'builtin' package: %v", e)
					}
					def, err = table.LookupDataType(fieldType.Def)
				} else {
					def, err = itemSymbolTable.LookupDataType(fieldType.Def)
				}

				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve1 %q type definition when retrieving a field", fieldType.Def)
				}
				if def.Def == nil {
					return nil, fmt.Errorf("Symbol %q not yet fully processed", fieldType.Def)
				}
				if _, ok := def.Def.(*gotypes.Interface); ok {
					embeddedInterfaces = append(embeddedInterfaces, embeddedInterfacesItem{symbolTable: itemSymbolTable, symbolDef: def})
				}
				continue
			case *gotypes.Selector:
				{
					byteSlice, _ := json.Marshal(fieldType)
					glog.V(2).Infof("++++%v\n", string(byteSlice))
				}
				// qid expected
				st, sd, err := a.RetrieveQidDataType(fieldType.Prefix, &ast.Ident{Name: fieldType.Item})
				if err != nil {
					return nil, err
				}

				embeddedInterfaces = append(embeddedInterfaces, embeddedInterfacesItem{symbolTable: st, symbolDef: sd})
				continue
			default:
				panic(fmt.Errorf("Unknown interface anonymous field type %#v", item))
			}
		}
		if methodName == method {
			methodItem = &item
			break
		}
	}

	glog.V(2).Info("Retrieving methods from embedded interfaces")
	if len(embeddedInterfaces) != 0 {
		for _, item := range embeddedInterfaces {
			if fieldDef, err := a.RetrieveInterfaceMethod(item.symbolTable, item.symbolDef, method); err == nil {
				return fieldDef, nil
			}
		}
	}

	if methodItem != nil {
		// if interfaceDefsymbol.Name != "" {
		// 	ep.AllocatedSymbolsTable.AddDataTypeField(interfaceDefsymbol.Package, interfaceDefsymbol.Name, method)
		// }
		return &FieldAttribute{
			DataType: methodItem.Def,
			IsMethod: true,
		}, nil
	}

	return nil, fmt.Errorf("Unable to find a method %v in interface %#v", method, interfaceDefsymbol)
}

type FieldAccessor struct {
	symbolTable symbols.SymbolLookable
	dataTypeDef *symbols.SymbolDef
	field       *ast.Ident
	methodsOnly bool
	fieldsOnly  bool
	// if set to true, lookup through fields only (no methods)
	// when processing embedded data types, set the flag back to false
	// This is usefull when a data type is defined as data type of another struct.
	// In that case, methods of the other struct are not inherited.
	dropFieldsOnly bool
}

func NewFieldAccessor(symbolTable symbols.SymbolLookable, dataTypeDef *symbols.SymbolDef, field *ast.Ident) *FieldAccessor {
	return &FieldAccessor{
		symbolTable: symbolTable,
		dataTypeDef: dataTypeDef,
		field:       field,
	}
}

func (f *FieldAccessor) SetMethodsOnly() *FieldAccessor {
	f.methodsOnly = true
	return f
}

func (f *FieldAccessor) SetFieldsOnly() *FieldAccessor {
	f.fieldsOnly = true
	return f
}

func (f *FieldAccessor) SetDropFieldsOnly() *FieldAccessor {
	f.dropFieldsOnly = true
	return f
}

func (a *Accessor) RetrieveStructFieldAtIndex(structDef *gotypes.Struct, idx int) (gotypes.DataType, error) {
	glog.V(2).Infof("Retrieving struct field at %q-th position from %#v\n", idx, structDef)

	if idx < 0 || idx >= len(structDef.Fields) {
		return nil, fmt.Errorf("Invalid index %v out of struct's <0; %v> interval", idx, len(structDef.Fields))
	}

	for i, field := range structDef.Fields {
		if i == idx {
			return field.Def, nil
		}
	}

	panic("RetrieveStructFieldAtIndex: unexpected behaviour, end of the world")
}

type FieldAttribute struct {
	// field definition
	gotypes.DataType
	// field's origin
	SymbolTable symbols.SymbolLookable
	// origin of the field
	Origin []*symbols.SymbolDef
	// is the field a method?
	IsMethod bool
}

// Get a struct's field.
// Given a struct can embedded another struct from a different package, the method must be able to access
// symbol tables of other packages. Thus recursively process struct's definition up to all its embedded fields.
// TODO(jchaloup): return symbol table of the returned data type
func (a *Accessor) RetrieveDataTypeField(accessor *FieldAccessor) (*FieldAttribute, error) {
	glog.V(2).Infof("Retrieving data type field %q from %#v\n", accessor.field.String(), accessor.dataTypeDef)
	// Only data type declaration is known
	if accessor.dataTypeDef.Def == nil {
		return nil, fmt.Errorf("Data type definition of %q is not known", accessor.dataTypeDef.Name)
	}

	// Any data type can have its own methods.
	// Or, a struct can embedded any data type with its own methods
	if accessor.dataTypeDef.Def.GetType() != gotypes.StructType {
		glog.V(2).Infof("Processing %#v", accessor.dataTypeDef.Def)
		if accessor.dataTypeDef.Def.GetType() == gotypes.IdentifierType {
			ident := accessor.dataTypeDef.Def.(*gotypes.Identifier)
			// check methods of the type itself if there is any match
			glog.V(2).Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return &FieldAttribute{
					DataType:    method.Def,
					SymbolTable: accessor.symbolTable,
					IsMethod:    true,
				}, nil
			}

			if ident.Package == "builtin" {
				// built-in => only methods
				// Check data type methods
				glog.V(2).Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
				if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
					return &FieldAttribute{
						DataType:    method.Def,
						SymbolTable: accessor.symbolTable,
						IsMethod:    true,
					}, nil
				}
				return nil, fmt.Errorf("Unable to find a field %v in builtin data type %#v", accessor.field.String(), accessor.dataTypeDef)
			}

			// if not built-in, it is an identifier of another type,
			// in which case all methods of the type are ignored
			// I.e. with 'type A B' data type of B is data type of A. However, methods of type B are not inherited.
			var pkgST symbols.SymbolLookable
			if ident.Package == "" || ident.Package == a.packageName {
				pkgST = a.symbolTable
			} else {
				pkgST = accessor.symbolTable
			}

			// No methods, just fields
			if !pkgST.Exists(ident.Def) {
				return nil, fmt.Errorf("Symbol %v not yet processed", ident.Def)
			}

			def, err := pkgST.LookupDataType(ident.Def)
			if err != nil {
				return nil, err
			}
			fieldAccessor, err := a.RetrieveDataTypeField(&FieldAccessor{
				symbolTable:    pkgST,
				dataTypeDef:    def,
				field:          accessor.field,
				fieldsOnly:     true,
				dropFieldsOnly: true,
			})
			if err != nil {
				return nil, err
			}
			fieldAccessor.Origin = append(fieldAccessor.Origin, accessor.dataTypeDef)
			return fieldAccessor, nil
		}

		if accessor.dataTypeDef.Def.GetType() == gotypes.SelectorType {
			// check methods of the type itself if there is any match
			glog.V(2).Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return &FieldAttribute{
					DataType:    method.Def,
					SymbolTable: accessor.symbolTable,
					IsMethod:    true,
				}, nil
			}

			selector := accessor.dataTypeDef.Def.(*gotypes.Selector)
			// qid?
			st, sd, err := a.RetrieveQidDataType(selector.Prefix, &ast.Ident{Name: selector.Item})
			if err != nil {
				return nil, err
			}

			// is it an interface
			if sd.Def.GetType() == gotypes.InterfaceType {
				return a.RetrieveInterfaceMethod(st, sd, accessor.field.Name)
			}

			fieldAccessor, err := a.RetrieveDataTypeField(&FieldAccessor{
				symbolTable:    st,
				dataTypeDef:    sd,
				field:          accessor.field,
				fieldsOnly:     true,
				dropFieldsOnly: true,
			})
			if err != nil {
				return nil, err
			}
			fieldAccessor.Origin = append(fieldAccessor.Origin, accessor.dataTypeDef)
			return fieldAccessor, nil
		}

		if accessor.dataTypeDef.Def.GetType() == gotypes.InterfaceType {
			return a.RetrieveInterfaceMethod(accessor.symbolTable, accessor.dataTypeDef, accessor.field.String())
		}

		if !accessor.fieldsOnly {
			// Check data type methods
			glog.V(2).Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return &FieldAttribute{
					DataType:    method.Def,
					SymbolTable: accessor.symbolTable,
				}, nil
			}
		}
		return nil, fmt.Errorf("Unable to find a field %v in data type %#v", accessor.field.String(), accessor.dataTypeDef)
	}

	type embeddedDataTypesItem struct {
		symbolTable symbols.SymbolLookable
		symbolDef   *symbols.SymbolDef
	}
	var embeddedDataTypes []embeddedDataTypesItem

	// Check struct field
	var fieldItem *gotypes.StructFieldsItem
ITEMS_LOOP:
	for _, item := range accessor.dataTypeDef.Def.(*gotypes.Struct).Fields {
		fieldName := item.Name
		// anonymous field (can be embedded struct as well)
		if fieldName == "" {
			itemExpr := item.Def
			if itemExpr.GetType() == gotypes.PointerType {
				itemExpr = itemExpr.(*gotypes.Pointer).Def
			}
			switch fieldType := itemExpr.(type) {
			case *gotypes.Identifier:
				if !accessor.methodsOnly {
					var fieldName string
					// localy defined data type?
					if parts := strings.Split(fieldType.Def, "#"); len(parts) == 3 {
						fieldName = parts[2]
					} else {
						fieldName = fieldType.Def
					}

					if fieldName == accessor.field.String() {
						fieldItem = &item
						break ITEMS_LOOP
					}
				}
				var def *symbols.SymbolDef
				var err error
				// check if the field is an embedded struct
				if fieldType.Package == "builtin" {
					table, e := a.globalSymbolTable.Lookup("builtin")
					if e != nil {
						return nil, fmt.Errorf("Unable to retrieve a symbol table for 'builtin' package: %v", e)
					}
					def, err = table.LookupDataType(fieldType.Def)
				} else {
					def, err = accessor.symbolTable.LookupDataType(fieldType.Def)
				}
				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve3 %q type definition when retrieving a field", fieldType.Def)
				}
				if def.Def == nil {
					return nil, fmt.Errorf("Symbol %q not yet fully processed", fieldType.Def)
				}
				embeddedDataTypes = append(embeddedDataTypes, embeddedDataTypesItem{symbolTable: accessor.symbolTable, symbolDef: def})
				continue
			case *gotypes.Selector:
				if !accessor.methodsOnly && fieldType.Item == accessor.field.String() {
					fieldItem = &item
					break ITEMS_LOOP
				}
				{
					byteSlice, _ := json.Marshal(fieldType)
					glog.V(2).Infof("++++%v\n", string(byteSlice))
				}
				// qid expected
				st, sd, err := a.RetrieveQidDataType(fieldType.Prefix, &ast.Ident{Name: fieldType.Item})
				if err != nil {
					return nil, err
				}

				embeddedDataTypes = append(embeddedDataTypes, embeddedDataTypesItem{symbolTable: st, symbolDef: sd})
				continue
			default:
				panic(fmt.Errorf("Unknown data type anonymous field type %#v", itemExpr))
			}
		}
		if !accessor.methodsOnly && fieldName == accessor.field.String() {
			fieldItem = &item
			break ITEMS_LOOP
		}
	}

	if !accessor.methodsOnly && fieldItem != nil {
		// if accessor.dataTypeDef.Name != "" {
		// 	ep.AllocatedSymbolsTable.AddDataTypeField(accessor.dataTypeDef.Package, accessor.dataTypeDef.Name, accessor.field.String())
		// }
		return &FieldAttribute{
			DataType:    fieldItem.Def,
			SymbolTable: accessor.symbolTable,
			Origin:      []*symbols.SymbolDef{accessor.dataTypeDef},
		}, nil
	}

	// First, check methods, then embedded structs

	// Check data type methods
	if !accessor.fieldsOnly {
		glog.V(2).Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
		if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
			return &FieldAttribute{
				DataType:    method.Def,
				SymbolTable: accessor.symbolTable,
				IsMethod:    false,
			}, nil
		}
	}

	glog.V(2).Info("Retrieving fields from embedded structs")
	if len(embeddedDataTypes) != 0 {
		if accessor.dropFieldsOnly {
			accessor.fieldsOnly = false
		}
		for _, item := range embeddedDataTypes {
			if fieldDefAttr, err := a.RetrieveDataTypeField(&FieldAccessor{
				symbolTable: item.symbolTable,
				dataTypeDef: item.symbolDef,
				field:       accessor.field,
				methodsOnly: accessor.methodsOnly,
				fieldsOnly:  accessor.fieldsOnly,
			}); err == nil {
				fieldDefAttr.Origin = append(fieldDefAttr.Origin, item.symbolDef)
				return fieldDefAttr, nil
			}
		}
	}

	{
		byteSlice, _ := json.Marshal(accessor.dataTypeDef)
		glog.V(2).Infof("++++%v\n", string(byteSlice))
	}
	return nil, fmt.Errorf("Unable to find a field %v in struct %#v", accessor.field.String(), accessor.dataTypeDef)
}
