package accessors

import (
	"encoding/json"
	"fmt"
	"go/ast"

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

func (a *Accessor) RetrieveQid(qidprefix gotypes.DataType, item *ast.Ident) (symbols.SymbolLookable, *symbols.SymbolDef, error) {
	// qid.structtype expected
	qid, ok := qidprefix.(*gotypes.Packagequalifier)
	if !ok {
		return nil, nil, fmt.Errorf("Expecting a qid.id when retrieving a symbol from a selector expression")
	}
	glog.Infof("Trying to retrieve a symbol %#v from package %v\n", item.String(), qid.Path)
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
	glog.Infof("Trying to retrieve a symbol %#v from package %v\n", item.String(), qid.Path)
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

func (a *Accessor) FindFirstNonidDataType(typeDef gotypes.DataType) (gotypes.DataType, error) {
	var symbolDef *symbols.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := a.RetrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
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
	return a.FindFirstNonidDataType(symbolDef.Def)
}

// TODO(jchaloup): return symbol table of the returned data type
func (a *Accessor) RetrieveInterfaceMethod(pkgsymboltable symbols.SymbolLookable, interfaceDefsymbol *symbols.SymbolDef, method string) (gotypes.DataType, error) {
	glog.Infof("Retrieving Interface method %q from %#v\n", method, interfaceDefsymbol)
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
				// check if the field is an embedded struct
				def, err := itemSymbolTable.LookupDataType(fieldType.Def)
				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve %q type definition when retrieving a field", fieldType.Def)
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
					glog.Infof("++++%v\n", string(byteSlice))
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

	glog.Info("Retrieving methods from embedded interfaces")
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
		return methodItem.Def, nil
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
	glog.Infof("Retrieving struct field at %q-th position from %#v\n", idx, structDef)

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

// Get a struct's field.
// Given a struct can embedded another struct from a different package, the method must be able to access
// symbol tables of other packages. Thus recursively process struct's definition up to all its embedded fields.
// TODO(jchaloup): return symbol table of the returned data type
func (a *Accessor) RetrieveDataTypeField(accessor *FieldAccessor) (gotypes.DataType, error) {
	glog.Infof("Retrieving data type field %q from %#v\n", accessor.field.String(), accessor.dataTypeDef)
	// Only data type declaration is known
	if accessor.dataTypeDef.Def == nil {
		return nil, fmt.Errorf("Data type definition of %q is not known", accessor.dataTypeDef.Name)
	}

	// Any data type can have its own methods.
	// Or, a struct can embedded any data type with its own methods
	if accessor.dataTypeDef.Def.GetType() != gotypes.StructType {
		glog.Infof("Processing %#v", accessor.dataTypeDef.Def)
		if accessor.dataTypeDef.Def.GetType() == gotypes.IdentifierType {
			ident := accessor.dataTypeDef.Def.(*gotypes.Identifier)
			// check methods of the type itself if there is any match
			glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return method.Def, nil
			}

			if ident.Package == "builtin" {
				// built-in => only methods
				// Check data type methods
				glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
				if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
					return method.Def, nil
				}
				return nil, fmt.Errorf("Unable to find a field %v in data type %#v", accessor.field.String(), accessor.dataTypeDef)
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
			return a.RetrieveDataTypeField(&FieldAccessor{
				symbolTable:    pkgST,
				dataTypeDef:    def,
				field:          accessor.field,
				fieldsOnly:     true,
				dropFieldsOnly: true,
			})
		}

		if accessor.dataTypeDef.Def.GetType() == gotypes.SelectorType {
			// check methods of the type itself if there is any match
			glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return method.Def, nil
			}

			selector := accessor.dataTypeDef.Def.(*gotypes.Selector)
			// qid?
			st, sd, err := a.RetrieveQidDataType(selector.Prefix, &ast.Ident{Name: selector.Item})
			if err != nil {
				return nil, err
			}

			return a.RetrieveDataTypeField(&FieldAccessor{
				symbolTable:    st,
				dataTypeDef:    sd,
				field:          accessor.field,
				fieldsOnly:     true,
				dropFieldsOnly: true,
			})
		}

		if !accessor.fieldsOnly {
			if accessor.dataTypeDef.Def.GetType() == gotypes.InterfaceType {
				return a.RetrieveInterfaceMethod(accessor.symbolTable, accessor.dataTypeDef, accessor.field.String())
			}
			// Check data type methods
			glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return method.Def, nil
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
			case *gotypes.Builtin:
				if !accessor.methodsOnly && fieldType.Def == accessor.field.String() {
					fieldItem = &item
					break ITEMS_LOOP
				}
				table, err := a.globalSymbolTable.Lookup("builtin")
				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve a symbol table for 'builtin' package: %v", err)
				}

				// check if the field is an embedded struct
				def, err := table.LookupDataType(fieldType.Def)
				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve %q type definition when retrieving a field", fieldType.Def)
				}
				if def.Def == nil {
					return nil, fmt.Errorf("Symbol %q not yet fully processed", fieldType.Def)
				}
				embeddedDataTypes = append(embeddedDataTypes, embeddedDataTypesItem{symbolTable: table, symbolDef: def})
				continue
			case *gotypes.Identifier:
				if !accessor.methodsOnly && fieldType.Def == accessor.field.String() {
					fieldItem = &item
					break ITEMS_LOOP
				}
				// check if the field is an embedded struct
				def, err := accessor.symbolTable.LookupDataType(fieldType.Def)
				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve %q type definition when retrieving a field", fieldType.Def)
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
					glog.Infof("++++%v\n", string(byteSlice))
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
		return fieldItem.Def, nil
	}

	// First, check methods, then embedded structs

	// Check data type methods
	if !accessor.fieldsOnly {
		glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
		if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
			return method.Def, nil
		}
	}

	glog.Info("Retrieving fields from embedded structs")
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
				return fieldDefAttr, nil
			}
		}
	}

	{
		byteSlice, _ := json.Marshal(accessor.dataTypeDef)
		glog.Infof("++++%v\n", string(byteSlice))
	}
	return nil, fmt.Errorf("Unable to find a field %v in struct %#v", accessor.field.String(), accessor.dataTypeDef)
}
