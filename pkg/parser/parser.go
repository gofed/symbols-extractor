package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	gost "github.com/gofed/symbols-extractor/pkg/types/symboltable"
)

//TODO: add func ParsePackage

//NOTE: it's ugly but big modification is really expected here...
// This will be main function used from outside for parsing of whole files
func ParseFiles(fname []string, symtab gost.SymbolTable) (gost.SymbolTable, error) {
	//TODO: support parsing through all files, for now, just one file is processed
	//      but in future is expected set of files as input
	//TODO: at first, create ASTs of all files, then start parsing
	var sym *gost.DeclType
	fset := token.NewFileSet()
	gofile := fname[0]
	f, err := parser.ParseFile(fset, gofile, nil, 0)
	if err != nil {
		//TODO: log or err.... log.Println(err)
		return nil, err
	}
	//TODO: use generator in future to get declared types instead of range ...*,
	//      so it will be possible to parse types from various places in code
	//      of this file in future -- global variables are expected to be used
	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			// get position of type declaration, which is part of 'decl' var
			tokPosition := fset.Position(decl.TokPos)
			position := gost.DeclPos{
				File: tokPosition.Filename,
				Line: uint(tokPosition.Line), //TODO: may use int instead of uint
			}
			for _, spec := range decl.Specs {
				switch d := spec.(type) {
				case *ast.TypeSpec:
					//TODO: for now parse just non-conflict types
					parsedType, err := parseTypeSpec(d)
					if err != nil {
						return nil, err
					}
					sym = &gost.DeclType{
						Pos:  position,
						Name: d.Name.Name,
						Def:  parsedType,
					}
					symtab.AddSymbol(d.Name.Name, sym)
				}
			}
		}
	}
	return symtab, nil
}

func printDataType(dataType gotypes.DataType) {
	byteSlice, _ := json.Marshal(dataType)
	fmt.Printf("\n%v\n", string(byteSlice))
}

func (tp *typesParser) parseTypeSpec(d *ast.TypeSpec) (gotypes.DataType, error) {
	// Here I get new type's definition.
	// The new type's id is not stored in the definition.
	// It is stored separatelly.
	typeDef, err := tp.parseTypeExpr(d.Type)
	if err != nil {
		return nil, err
	}

	// TODO(jchaloup): store type's ID and definition into a symbol table.
	// Or make it a part of the definition itself
	// --- pstodulk: symbol table, I am going to do that
	printDataType(typeDef)
	return typeDef, nil
}

func (tp *typesParser) parseStruct(typedExpr *ast.StructType) (*gotypes.Struct, error) {
	structType := &gotypes.Struct{}
	structType.Fields = make([]gotypes.StructFieldsItem, 0)

	if typedExpr.Fields.List == nil {
		return structType, nil
	}

	for _, field := range typedExpr.Fields.List {
		// anonymous field?
		if field.Names == nil {
			def, err := tp.parseTypeExpr(field.Type)
			if err != nil {
				return nil, err
			}

			item := gotypes.StructFieldsItem{
				Name: "",
				Def:  def,
			}

			structType.Fields = append(structType.Fields, item)
			// named fields
		} else {
			for _, name := range field.Names {
				def, err := tp.parseTypeExpr(field.Type)
				if err != nil {
					return nil, err
				}

				item := gotypes.StructFieldsItem{
					Name: name.Name,
					Def:  def,
				}
				structType.Fields = append(structType.Fields, item)
			}
		}
	}
	return structType, nil
}

func (tp *typesParser) parseIdentifier(typedExpr *ast.Ident) (*gotypes.Identifier, error) {
	// TODO(jchaloup): store symbol's origin as well
	return &gotypes.Identifier{
		Def: typedExpr.Name,
	}, nil
}

func (tp *typesParser) parseStar(typedExpr *ast.StarExpr) (*gotypes.Pointer, error) {
	// X.Sel
	def, err := tp.parseTypeExpr(typedExpr.X)
	if err != nil {
		return nil, err
	}
	return &gotypes.Pointer{
		Def: def,
	}, nil
}

func (tp *typesParser) parseChan(typedExpr *ast.ChanType) (*gotypes.Channel, error) {
	def, err := tp.parseTypeExpr(typedExpr.Value)
	if err != nil {
		return nil, err
	}

	channel := &gotypes.Channel{
		Value: def,
	}

	switch typedExpr.Dir {
	case ast.SEND:
		channel.Dir = "1"
	case ast.RECV:
		channel.Dir = "2"
	default:
		channel.Dir = "3"
	}

	return channel, nil
}

func (tp *typesParser) parseEllipsis(typedExpr *ast.Ellipsis) (*gotypes.Ellipsis, error) {
	// X.Sel
	def, err := tp.parseTypeExpr(typedExpr.Elt)
	if err != nil {
		return nil, err
	}
	return &gotypes.Ellipsis{
		Def: def,
	}, nil
}

func (tp *typesParser) parseSelector(typedExpr *ast.SelectorExpr) (*gotypes.Selector, error) {
	// X.Sel
	def, err := tp.parseTypeExpr(typedExpr.X)
	if err != nil {
		return nil, err
	}
	return &gotypes.Selector{
		Item:   typedExpr.Sel.Name,
		Prefix: def,
	}, nil
}

func (tp *typesParser) parseMap(typedExpr *ast.MapType) (*gotypes.Map, error) {
	keyDef, keyErr := tp.parseTypeExpr(typedExpr.Key)
	if keyErr != nil {
		return nil, keyErr
	}

	valueDef, valueErr := tp.parseTypeExpr(typedExpr.Value)
	if valueErr != nil {
		return nil, valueErr
	}

	return &gotypes.Map{
		Keytype:   keyDef,
		Valuetype: valueDef,
	}, nil
}

func (tp *typesParser) parseArray(typedExpr *ast.ArrayType) (gotypes.DataType, error) {
	def, err := tp.parseTypeExpr(typedExpr.Elt)
	if err != nil {
		return nil, err
	}
	if typedExpr.Len == nil {
		return &gotypes.Slice{
			Elmtype: def,
		}, nil
	}

	// TODO(jchaloup): store array's length as well
	return &gotypes.Array{
		Elmtype: def,
	}, nil
}

func (tp *typesParser) parseInterface(typedExpr *ast.InterfaceType) (*gotypes.Interface, error) {
	// TODO(jchaloup): extend the interface definition with embedded interfaces
	interfaceObj := &gotypes.Interface{}
	var methods []gotypes.InterfaceMethodsItem
	for _, m := range typedExpr.Methods.List {
		def, err := tp.parseTypeExpr(m.Type)
		if err != nil {
			return nil, err
		}

		for _, name := range m.Names {
			item := gotypes.InterfaceMethodsItem{
				Name: name.Name,
				Def:  def,
			}
			methods = append(methods, item)
		}
	}
	if len(methods) > 0 {
		interfaceObj.Methods = methods
	}

	return interfaceObj, nil
}

func (tp *typesParser) parseFunction(typedExpr *ast.FuncType) (*gotypes.Function, error) {
	functionType := &gotypes.Function{}

	var params []gotypes.DataType
	var results []gotypes.DataType

	if typedExpr.Params != nil {
		for _, field := range typedExpr.Params.List {
			def, err := tp.parseTypeExpr(field.Type)
			if err != nil {
				return nil, err
			}
			params = append(params, def)
		}

		if len(params) > 0 {
			functionType.Params = params
		}
	}

	if typedExpr.Results != nil {
		for _, field := range typedExpr.Results.List {
			def, err := tp.parseTypeExpr(field.Type)
			if err != nil {
				return nil, err
			}
			results = append(results, def)
		}

		if len(results) > 0 {
			functionType.Results = results
		}
	}

	return functionType, nil
}

func (tp *typesParser) parseTypeExpr(expr ast.Expr) (gotypes.DataType, error) {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		return tp.parseIdentifier(typedExpr)
	case *ast.StarExpr:
		return tp.parseStar(typedExpr)
	case *ast.ChanType:
		return tp.parseChan(typedExpr)
	case *ast.Ellipsis:
		return tp.parseEllipsis(typedExpr)
	case *ast.SelectorExpr:
		return tp.parseSelector(typedExpr)
	case *ast.MapType:
		return tp.parseMap(typedExpr)
	case *ast.ArrayType:
		return tp.parseArray(typedExpr)
	case *ast.StructType:
		return tp.parseStruct(typedExpr)
	case *ast.InterfaceType:
		return tp.parseInterface(typedExpr)
	case *ast.FuncType:
		return tp.parseFunction(typedExpr)
	}
	return nil, fmt.Errorf("ast.Expr (%#v) not recognized", expr)
}

type AllocatedSymbolsTable gost.HST

type typesParser struct {
	allocatedSymbolsTable *AllocatedSymbolsTable
}

func NewParser() *typesParser {
	return &typesParser{
		allocatedSymbolsTable: &AllocatedSymbolsTable{},
	}
}

func (tp *typesParser) Parse(gofile string) error {
	fset := token.NewFileSet()
	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, gofile, nil, 0)
	if err != nil {
		return err
	}

	// Print the imports from the file's AST.
	for _, d := range f.Decls {
		//fmt.Printf("%v\n", d)
		// accessing dynamic_value := interface_variable.(typename)
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch d := spec.(type) {
				case *ast.ImportSpec:
					// process imports first
					//fmt.Printf("%+v\n", d)
				case *ast.ValueSpec:
					// process value and constants as third
					//fmt.Printf("%+v\n", d)
				case *ast.TypeSpec:
					// process type definitions as second
					//fmt.Printf("%#v\n", d)
					_, err := tp.parseTypeSpec(d)
					if err != nil {
						return err
					}
				}
			}
		case *ast.FuncDecl:
			// process function definitions as the last
			//fmt.Printf("%+v\n", d)
		}
	}
	return nil
}
