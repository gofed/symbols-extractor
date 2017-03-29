package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

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

	tp.symbolTable.AddDataType(&gotypes.SymbolDef{
		Name: d.Name.Name,
		Def:  typeDef,
	})

	//printDataType(typeDef)
	return typeDef, nil
}

func (tp *typesParser) parseFuncDecl(d *ast.FuncDecl) (gotypes.DataType, error) {
	// parseFunction does not store name of params, resp. results
	// as the names are not important. Just params, resp. results ordering is.
	// Thus, this method is used to parse function's signature only.
	funcDef, err := tp.parseFunction(d.Type)
	if err != nil {
		return nil, err
	}

	// Names of function params, resp. results are needed only when function's body is processed.
	// Thus, we can collect params, resp. results definition from symbol table and get names from
	// function's AST. (the ast is processed twice but in the second case, only params, resp. results names are read).

	// The function/method signature belongs to a package/file level symbol table.
	// The functonn/methods's params, resp. results, resp. receiver identifiers belong to
	// multi-level symbol table that is function/method's scoped. Once the body is left,
	// the multi-level symbol table is dropped.

	if d.Recv != nil {
		// Receiver has a single parametr
		// https://golang.org/ref/spec#Receiver
		if len((*d.Recv).List) != 1 || len((*d.Recv).List[0].Names) != 1 {
			return nil, fmt.Errorf("Receiver is not a single parameter")
		}

		//fmt.Printf("Rec Name: %#v\n", (*d.Recv).List[0].Names[0].Name)

		recDef, err := tp.parseTypeExpr((*d.Recv).List[0].Type)
		if err != nil {
			return nil, err
		}

		methodDef := &gotypes.Method{
			Def:      funcDef,
			Receiver: recDef,
		}

		tp.symbolTable.AddFunction(&gotypes.SymbolDef{
			Name: d.Name.Name,
			Def:  methodDef,
		})

		//printDataType(methodDef)

		return methodDef, nil
	}

	//printDataType(funcDef)

	return funcDef, nil
}

func (tp *typesParser) parseDefExpr(expr ast.Expr) {
	fmt.Printf("Expr: %#v\n", expr)
	switch exprType := expr.(type) {
	// Basic literal carries
	case *ast.BasicLit:
		fmt.Printf("BasicLit: %#v\n", exprType)
	}
}

func (tp *typesParser) parseStatement(statement ast.Stmt) error {
	switch stmtExpr := statement.(type) {
	case *ast.ReturnStmt:
		fmt.Printf("Return: %#v\n", stmtExpr)
		for _, result := range stmtExpr.Results {
			tp.parseDefExpr(result)
		}
	}

	return nil
}

func (tp *typesParser) parseFuncBody(funcDecl *ast.FuncDecl) error {
	// Function/method signature is already stored in a symbol table.
	// From function/method's AST get its receiver, parameters and results,
	// construct a first level of a multi-level symbol table stack..
	// For each new block (including the body) push another level into the stack.

	stack := symboltable.NewStack()

	stack.Push()

	if funcDecl.Recv != nil {
		// Receiver has a single parametr
		// https://golang.org/ref/spec#Receiver
		if len((*funcDecl.Recv).List) != 1 || len((*funcDecl.Recv).List[0].Names) != 1 {
			return fmt.Errorf("Receiver is not a single parameter")
		}

		def, err := tp.parseTypeExpr((*funcDecl.Recv).List[0].Type)
		if err != nil {
			return err
		}
		stack.AddDataType(&gotypes.SymbolDef{
			Name: (*funcDecl.Recv).List[0].Names[0].Name,
			Def:  def,
		})
	}

	if funcDecl.Type.Params != nil {
		for _, field := range funcDecl.Type.Params.List {
			def, err := tp.parseTypeExpr(field.Type)
			if err != nil {
				return err
			}

			// field.Names is always non-empty if param's datatype is defined
			for _, name := range field.Names {
				fmt.Printf("Name: %v\n", name.Name)
				stack.AddDataType(&gotypes.SymbolDef{
					Name: name.Name,
					Def:  def,
				})
			}
		}
	}

	if funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			def, err := tp.parseTypeExpr(field.Type)
			if err != nil {
				return err
			}

			for _, name := range field.Names {
				fmt.Printf("Name: %v\n", name.Name)
				stack.AddDataType(&gotypes.SymbolDef{
					Name: name.Name,
					Def:  def,
				})
			}
		}
	}

	stack.Push()

	// The stack will always have at least one symbol table (with receivers, resp. parameters, resp. results)
	for _, statement := range funcDecl.Body.List {
		fmt.Printf("statement: %#v\n", statement)
		if err := tp.parseStatement(statement); err != nil {
			return err
		}
	}

	//stack.Print()

	// Here!!! The symbol type analysis is carried here!!! Yes, HERE!!!

	return nil
}

func (tp *typesParser) parseIdentifier(typedExpr *ast.Ident) (*gotypes.Identifier, error) {
	// TODO(jchaloup): store symbol's origin as well (in a case a symbol is imported without qid)
	// Check if the identifier is in the any of the global symbol tables (in a case a symbol is imported without qid).
	// If it is, the origin is known. If it is not, the origin is the current package.
	// Check if the identifier is embedded type first. Then the origin is empty.

	// Every data type definition consists of a set of identifiers.
	// Whenever an identifier is used in the definition,
	// it is allocated.
	tp.allocatedSymbolsTable.AddSymbol("", typedExpr.Name)

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
	// X.Sel a.k.a Prefix.Item

	id, ok := typedExpr.X.(*ast.Ident)
	// if the prefix is identifier only, the selector is in a form qui.identifier,
	// i.e. fully qualified identifier
	// TODO(jchaloup): check id.function(), is that selector as well or not?
	if ok {
		// TODO(jchaloup): replace the id.Name with full package path
		// E.g. instead of (qid, identifier) use (github.com/coreos/etcd/pkg/wait, Wait)
		tp.allocatedSymbolsTable.AddSymbol(id.Name, typedExpr.Sel.Name)
		return &gotypes.Selector{
			Item: typedExpr.Sel.Name,
			Prefix: &gotypes.Identifier{
				Def: id.Name,
			},
		}, nil
	}

	def, err := tp.parseTypeExpr(typedExpr.X)
	if err != nil {
		return nil, err
	}
	return &gotypes.Selector{
		Item:   typedExpr.Sel.Name,
		Prefix: def,
	}, nil
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

			// field.Names list must be singletion at least
			for i := 0; i < len(field.Names); i++ {
				params = append(params, def)
			}
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

			// results can be identifier free
			if len(field.Names) == 0 {
				results = append(results, def)
			} else {
				for i := 0; i < len(field.Names); i++ {
					results = append(results, def)
				}
			}

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

type typesParser struct {
	// per file symbol table
	symbolTable *symboltable.Table
	// per file allocatable ST
	allocatedSymbolsTable *AllocatedSymbolsTable

	// TODO(jchaloup):
	// - create a project scoped symbol table in a higher level struct (e.g. ProjectParser)
	// - merge all per file symbol tables continuously in the higher level struct (each time a new symbol definition is process)
	//
}

func NewParser() *typesParser {
	return &typesParser{
		symbolTable:           symboltable.NewTable(),
		allocatedSymbolsTable: NewAllocatableSymbolsTable(),
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
			_, err := tp.parseFuncDecl(decl)
			if err != nil {
				return err
			}

			// if an error is returned, put the function's AST into a context list
			// and continue with other definition
			tp.parseFuncBody(decl)
		}
	}

	tp.allocatedSymbolsTable.Print()

	byteSlice, _ := json.Marshal(tp.symbolTable)
	fmt.Printf("\nSymbol table: %v\n", string(byteSlice))

	newObject := &symboltable.Table{}
	if err := json.Unmarshal(byteSlice, newObject); err != nil {
		fmt.Printf("Error: %v", err)
	}

	fmt.Printf("\nAfter: %#v", newObject)

	return nil
}
