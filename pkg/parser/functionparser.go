package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

var binaryOperators = []token.Token{
	token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
	token.AND, token.OR, token.XOR, token.SHL, token.SHR, token.AND_NOT,
	token.LAND, token.LOR, token.EQL, token.LSS, token.GTR, token.NEQ, token.LEQ, token.GEQ,
}

func isBinaryOperator(operator token.Token) bool {
	for _, tok := range binaryOperators {
		if tok == operator {
			return true
		}
	}
	return false
}

var builtinTypes = map[string]struct{}{
	//TODO(pstodulk):
	//  - generate the map from: https://golang.org/src/builtin/builtin.go
	//  - NOTE: there are just documentation types too (like Type1) which
	//          should not be part of this map
	//          - append all types except of Doc. only?
	"uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
	"int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {},
	"float32": {}, "float64": {},
	"complex64": {}, "complex128": {},
	"string": {}, "byte": {}, "rune": {},
	"chan": {}, "bool": {},
	"uintptr": {}, "error": {},
}

func isBuiltin(ident string) bool {
	_, ok := builtinTypes[ident]

	return ok
}

// functionParser parses declaration and definition of a function/method
type functionParser struct {
	// package name
	packageName string
	// per file symbol table
	symbolTable *symboltable.Stack
	// per file allocatable ST
	allocatedSymbolsTable *AllocatedSymbolsTable
	// types parser
	typesParser *typesParser
}

// NewFunctionParser create an instance of a function parser
func NewFunctionParser(packageName string, symbolTable *symboltable.Stack, allocatedSymbolsTable *AllocatedSymbolsTable, typesParser *typesParser) *functionParser {
	return &functionParser{
		packageName:           packageName,
		symbolTable:           symbolTable,
		allocatedSymbolsTable: allocatedSymbolsTable,
		typesParser:           typesParser,
	}
}

func (fp *functionParser) parseReceiver(receiver ast.Expr, skip_allocated bool) (gotypes.DataType, error) {
	// Receiver's type must be of the form T or *T (possibly using parentheses) where T is a type name.
	switch typedExpr := receiver.(type) {
	case *ast.Ident:
		// search the identifier in the symbol table
		def, err := fp.symbolTable.Lookup(typedExpr.Name)
		if err != nil {
			fmt.Printf("Lookup error: %v\n", err)
			// Return an error so the function body processing can be postponed
			// TODO(jchaloup): return more information about the missing symbol so the
			// body can be re-processed right after the symbol is stored into the symbol table.
			return nil, err
		}

		if !skip_allocated {
			fp.allocatedSymbolsTable.AddSymbol(def.Package, typedExpr.Name)
		}

		return &gotypes.Identifier{
			Def: typedExpr.Name,
		}, nil
	case *ast.StarExpr:
		fmt.Printf("Start: %#v\n", typedExpr)
		switch idExpr := typedExpr.X.(type) {
		case *ast.Ident:
			// search the identifier in the symbol table
			def, err := fp.symbolTable.Lookup(idExpr.Name)
			if err != nil {
				fmt.Printf("Lookup error: %v\n", err)
				// Return an error so the function body processing can be postponed
				// TODO(jchaloup): return more information about the missing symbol so the
				// body can be re-processed right after the symbol is stored into the symbol table.
				return nil, err
			}

			if !skip_allocated {
				fp.allocatedSymbolsTable.AddSymbol(def.Package, idExpr.Name)
			}

			return &gotypes.Pointer{
				Def: &gotypes.Identifier{
					Def: idExpr.Name,
				},
			}, nil
		default:
			return nil, fmt.Errorf("Method receiver %#v is not a pointer to an identifier", idExpr)
		}
	default:
		return nil, fmt.Errorf("Method receiver %#v is not a pointer to an identifier not an identifier", typedExpr)
	}
}

func (fp *functionParser) parseFuncDecl(d *ast.FuncDecl) (gotypes.DataType, error) {
	// parseFunction does not store name of params, resp. results
	// as the names are not important. Just params, resp. results ordering is.
	// Thus, this method is used to parse function's signature only.
	funcDef, err := fp.typesParser.parseFunction(d.Type)
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

		recDef, err := fp.parseReceiver((*d.Recv).List[0].Type, false)
		if err != nil {
			return nil, err
		}

		methodDef := &gotypes.Method{
			Def:      funcDef,
			Receiver: recDef,
		}

		fp.symbolTable.AddFunction(&gotypes.SymbolDef{
			Name:    d.Name.Name,
			Package: fp.packageName,
			Def:     methodDef,
		})

		//printDataType(methodDef)

		return methodDef, nil
	}

	// Empty receiver => Function
	fp.symbolTable.AddFunction(&gotypes.SymbolDef{
		Name:    d.Name.Name,
		Package: fp.packageName,
		Def:     funcDef,
	})

	//printDataType(funcDef)

	return funcDef, nil
}

func (fp *functionParser) parseBasicLit(lit *ast.BasicLit) (gotypes.DataType, error) {
	switch lit.Kind {
	case token.INT:
		return &gotypes.Builtin{}, nil
	case token.FLOAT:
		return &gotypes.Builtin{}, nil
	case token.IMAG:
		return &gotypes.Builtin{}, nil
	case token.STRING:
		return &gotypes.Builtin{}, nil
	case token.CHAR:
		return &gotypes.Builtin{}, nil
	default:
		return nil, fmt.Errorf("Unrecognize BasicLit: %#v\n", lit.Kind)
	}
}

func (fp *functionParser) parseCompositeLiteralElements(clType *gotypes.SymbolDef, elements []ast.Expr, indent int) (gotypes.DataType, error) {
	// The CL type can be one of the following:
	// - identifier:
	//		 struct identifier, can it be anything else?
	//	   With the struct I define anonymous data type whose fields ara access later so I need to store its definition as well.
	// - selectpr:
	//     qui.identifier (can it be anything else)
	// - array:
	//     each item can be processed separatelly, just delegate the valueType further
	// - map:
	//		 https://golang.org/ref/spec#Map_types:
	//		   "the key type must not be a function, map, or slice"
	//     can I use anything else than a built-in type for the keyType?
	//     each item of the map can be processed separately
	//     Given a map does not have fields, there is no need to inspect the keyType for them.
	//     It can potentially contain user defined identifiers
	// var elementType gotypes.DataType
	// switch clTypeExpr := clType.Def.(type) {
	// case *gotypes.Builtin:
	// 	// all built-in types are ignored
	// 	return clTypeExpr, nil
	// case *gotypes.Slice:
	// 	// each item of a slice is of the same type => ignore it
	// 	elementType = clTypeExpr.Elmtype
	// case *gotypes.Struct:
	// 	elementType = clTypeExpr
	// }

	fmt.Printf("======================================Inside parseCompositeLiteralElements: identation: %v\n", indent)
	fmt.Printf("%vclType: %#v\n", strings.Repeat("\t", indent), clType)
	fmt.Printf("%vElements: %#v\n", strings.Repeat("\t", indent), elements)

	for _, elem := range elements {
		fmt.Printf("%vElement: %#v\n", strings.Repeat("\t", indent), elem)
		switch elemExpr := elem.(type) {
		case *ast.BasicLit:
			return &gotypes.Builtin{}, nil
		case *ast.Ident:
			// Search for the symbol definition in the symbol table.
			// If the identifier points to a struct data type, we need to store struct's fields as well.
			// TODO(jchaloup): find a way how to propagate struct's data type name into the recursive call stack
			fmt.Printf("%v==ast.Ident %#v\ttype: %#v\n", strings.Repeat("\t", indent), elemExpr, clType)
			// Here, it is always a variable or a built-in type
			// TODO(jchaloup): check if the variable is global and count it into the allocatable symbol table
			return &gotypes.Identifier{Def: elemExpr.Name}, nil
		case *ast.UnaryExpr:
			fmt.Printf("%v==ast.UnaryExpr %#v\ttype: %#v\n", strings.Repeat("\t", indent), elemExpr, clType)
			d, e := fp.parseCompositeLiteralElements(nil, []ast.Expr{elemExpr.X}, indent+1)
			if e != nil {
				return d, e
			}
		case *ast.KeyValueExpr:
			// allowed for a struct and map only
			switch keyValueDef := clType.Def.(type) {
			case *gotypes.Struct:
				elemKey, ok := elemExpr.Key.(*ast.Ident)
				if !ok {
					return nil, fmt.Errorf("ast.KeyValueExpr key %#v is not an identifier", elemExpr.Key)
				}
				var keyDef gotypes.DataType
				for _, item := range keyValueDef.Fields {
					//fmt.Printf("\t%vField: %v: %#v\n", strings.Repeat("\t", indent), item.Name, item.Def)
					if item.Name == elemKey.Name {
						//fmt.Printf("\t%vField %v found: %#v\n", strings.Repeat("\t", indent), elemKey.Name, item.Def)
						//fp.allocatedSymbolsTable.AddDataTypeField(structDefsymbol.Package, structDefsymbol.Name, elemKey.Name)
						keyDef = item.Def
						break
					}
				}
				if keyDef == nil {
					return nil, fmt.Errorf("Unable to find a key %v in struct %#v of a CompositeLiteral", elemKey.Name, keyValueDef)
				}
				fmt.Printf("%v==ast.KeyValueExpr %v Element:\n%v\ttype: %#v\n%v\tvalue: %#v\n", strings.Repeat("\t", indent), elemKey.Name, strings.Repeat("\t", indent), keyDef, strings.Repeat("\t", indent), elemExpr.Value)
				// as a value can be of a different type then its key (e.g. interface vs. type implementing the interface), each value type of a struct is re-parsed fron the value type
				d, e := fp.parseCompositeLiteralElements(nil, []ast.Expr{elemExpr.Value}, indent+1)
				if e != nil {
					return d, e
				}
			case *gotypes.Map:
				// the key can be basically anything that is a literal, id or can be evaluated into the literal, resp. id
				// Given the key is already parsed, we can skip it
				fmt.Printf("%vMapKeyType: %#v\n", strings.Repeat("\t", indent), keyValueDef.Keytype)
				// The same holds for the value, we already processed it so all data types are already accounted
				fmt.Printf("%vMapValueType: %#v\n", strings.Repeat("\t", indent), keyValueDef.Valuetype)
				fmt.Printf("%velemExpr: %#v\n", strings.Repeat("\t", indent), elemExpr.Value)
				d, e := fp.parseCompositeLiteralElements(nil, []ast.Expr{elemExpr.Value}, indent+1)
				if e != nil {
					return d, e
				}
			default:
				return nil, fmt.Errorf("ast.KeyValueExpr is allowed for a struct and map type only. Instead, %#v type is available", clType)
			}

		case *ast.CompositeLit:
			fmt.Printf("%vCL: %#v\n", strings.Repeat("\t", indent), elemExpr)
			// nil CL types = array/slice items
			fmt.Printf("%v==ast.CompositeLit Element:\n%v\ttype: %#v\n%v\tvalue: %#v\n", strings.Repeat("\t", indent), strings.Repeat("\t", indent), nil, strings.Repeat("\t", indent), elemExpr)
			if elemExpr.Type != nil {
				clElemType, err := fp.typesParser.parseTypeExpr(elemExpr.Type)
				if err != nil {
					return nil, err
				}
				fmt.Printf("%velemExpr.Type: %#v\t%#v\n", strings.Repeat("\t", indent), elemExpr.Type, clElemType)
				if clElemTypeIdent, ok := clElemType.(*gotypes.Identifier); ok {
					def, err := fp.symbolTable.Lookup(clElemTypeIdent.Def)
					if err != nil {
						return nil, err
					}
					// TODO(jchaloup): add the symbol into the allocated symbol table
					d, e := fp.parseCompositeLiteralElements(def, elemExpr.Elts, indent+1)
					if e != nil {
						return d, e
					}
				} else {
					d, e := fp.parseCompositeLiteralElements(&gotypes.SymbolDef{Def: clElemType}, elemExpr.Elts, indent+1)
					if e != nil {
						return d, e
					}
				}
			} else {
				// slice/array
				if clType == nil {
					panic(fmt.Errorf("Expecting slice type for %#v. Got nil.", elemExpr))
				}
				slExpr, ok := clType.Def.(*gotypes.Slice)
				if !ok {
					panic(fmt.Errorf("Expecting slice, got %#v\n", clType.Def))
				}
				d, err := fp.parseCompositeLiteralElements(&gotypes.SymbolDef{Def: slExpr.Elmtype}, elemExpr.Elts, indent+1)
				if err != nil {
					return d, err
				}
			}
		default:
			panic(fmt.Errorf("Unknown element type: %#v", elem))
		}
	}
	return nil, nil
}

func (fp *functionParser) parseCompositeLit(lit *ast.CompositeLit) (gotypes.DataType, error) {
	// https://golang.org/ref/spec#Composite_literals
	// The LiteralType's underlying type must be a struct, array, slice, or map
	//  type (the grammar enforces this constraint except when the type is given as a TypeName)
	//
	// lit.Type{
	//   Elts[0],
	//   Elts[1],
	//   Elts[2],
	//}

	// The CL type can be processed independently of the CL elements
	fmt.Printf("CL Type: %#v\n", lit.Type)
	def, err := fp.typesParser.parseTypeExpr(lit.Type)
	if err != nil {
		return nil, err
	}

	// TODO(jchaloup): the typeIdent can be a selector as well (qui.id), check that as well
	if clElemTypeIdent, ok := def.(*gotypes.Identifier); ok {
		clElemTypeIdentDef, err := fp.symbolTable.Lookup(clElemTypeIdent.Def)
		if err != nil {
			panic(err)
			return nil, err
		}
		// TODO(jchaloup): add the symbol into the allocated symbol table
		_, err = fp.parseCompositeLiteralElements(clElemTypeIdentDef, lit.Elts, 0)
		if err != nil {
			panic(err)
		}
	} else {
		_, err = fp.parseCompositeLiteralElements(&gotypes.SymbolDef{Def: def}, lit.Elts, 0)
		if err != nil {
			panic(err)
		}
	}

	return nil, nil
}

func (fp *functionParser) parseIdentifier(ident *ast.Ident) (gotypes.DataType, error) {
	// TODO(jchaloup): put the nil into the allocated symbol table
	if ident.Name == "nil" {
		return &gotypes.Nil{}, nil
	}

	// Check if the symbol is in the symbol table.
	// It is either a local variable or a global variable (as it is used inside an expression).
	// If the symbol is not found, it means it is not defined and never will be.

	// If it is a variable, return its definition
	if def, err := fp.symbolTable.LookupVariable(ident.Name); err == nil {
		fmt.Printf("Variable used: %v.%v %#v\n", def.Package, def.Name, def.Def)
		return def.Def, nil
	}

	// Otherwise it is a data type of a function declation -> return just the data type identifier
	def, err := fp.symbolTable.Lookup(ident.Name)
	if err != nil {
		fmt.Printf("Lookup error: %v\n", err)
		// Return an error so the function body processing can be postponed
		// TODO(jchaloup): return more information about the missing symbol so the
		// body can be re-processed right after the symbol is stored into the symbol table.
		return nil, err
	}

	// TODO(jchaloup): put the identifier's type into the allocated symbol table
	fmt.Printf("Symbol used: %v.%v %#v\n", def.Package, def.Name, def.Def)
	return &gotypes.Identifier{Def: def.Name}, nil
}

func (fp *functionParser) parseUnaryExpr(expr *ast.UnaryExpr) (gotypes.DataType, error) {
	def, err := fp.parseExpr(expr.X)
	if err != nil {
		return nil, err
	}

	if len(def) != 1 {
		return nil, fmt.Errorf("Operand of an unary operator is not a single value")
	}

	// TODO(jchaloup): check the token is really a unary operator
	switch expr.Op {
	// variable address
	case token.AND:
		return &gotypes.Pointer{
			Def: def[0],
		}, nil
	default:
		return nil, fmt.Errorf("Unary operator %#v not recognized", expr.Op)
	}
}

func (fp *functionParser) parseBinaryExpr(expr *ast.BinaryExpr) (gotypes.DataType, error) {
	if !isBinaryOperator(expr.Op) {
		return nil, fmt.Errorf("Binary operator %#v not recognized", expr.Op)
	}

	// Given all binary operators are valid only for build-in types
	// and any operator result is built-in type again,
	// the Buildin is return.
	// However, as the operands itself can be built of
	// user defined type returning built-in type, both operands must be processed.

	// X Op Y
	_, yErr := fp.parseExpr(expr.X)
	if yErr != nil {
		return nil, yErr
	}

	_, xErr := fp.parseExpr(expr.Y)
	if xErr != nil {
		return nil, xErr
	}

	return &gotypes.Builtin{}, nil
}

func (fp *functionParser) parseStarExpr(expr *ast.StarExpr) (gotypes.DataType, error) {
	def, err := fp.parseExpr(expr.X)
	if err != nil {
		return nil, err
	}

	if len(def) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}

	return &gotypes.Pointer{
		Def: def[0],
	}, nil
}

func (fp *functionParser) parseCallExpr(expr *ast.CallExpr) ([]gotypes.DataType, error) {
	// TODO(jchaloup): check if the type() casting is of ast.CallExpr as well
	for _, arg := range expr.Args {
		// an argument is passed to the function so its data type does not affect the result data type of the call
		def, err := fp.parseExpr(arg)
		if err != nil {
			return nil, err
		}
		if len(def) != 1 {
			return nil, fmt.Errorf("Argument %#v of a call expression does not have one return value", arg)
		}
		// TODO(jchaloup): data type of the argument itself can be propagated to the function/method
		// to provide more information about the type if the function parameter is an interface.
	}

	var function gotypes.DataType

	fmt.Printf("CallExpr::: %#v\n", expr.Fun)
	switch exprType := expr.Fun.(type) {
	case *ast.Ident:
		// simple function ID
		fmt.Printf("Function name: %v\n", exprType.Name)
		def, err := fp.symbolTable.Lookup(exprType.Name)
		if err != nil {
			fmt.Printf("Lookup error: %v\n", err)
			// Return an error so the function body processing can be postponed
			// TODO(jchaloup): return more information about the missing symbol so the
			// body can be re-processed right after the symbol is stored into the symbol table.
			return nil, err
		}

		// Get function's definition from the symbol table (if it exists)
		// and use its result type as the return type.
		// If a function/method call is a part of an expression, the function/method has only ony result type.
		fmt.Printf("Def: %#v\n", def.Def)

		// Store the symbol to the allocated ST
		fp.allocatedSymbolsTable.AddSymbol(def.Package, exprType.Name)
		function = def.Def
	case *ast.SelectorExpr:
		fmt.Printf("S Function name: %#v\n", exprType)
		def, err := fp.parseSelectorExpr(exprType)
		if err != nil {
			return []gotypes.DataType{def}, err
		}
		function = def
		fmt.Printf("Def: %#v\nerr: %v\n", def, err)
	default:
		return nil, fmt.Errorf("Call expr not recognized: %#v", expr)
	}

	switch funcType := function.(type) {
	case *gotypes.Function:
		return funcType.Results, nil
	case *gotypes.Method:
		return funcType.Def.(*gotypes.Function).Results, nil
	default:
		return nil, fmt.Errorf("Symbol to be called is not a function")
	}
}

func (fp *functionParser) parseStructType(expr *ast.StructType) (gotypes.DataType, error) {
	for _, field := range expr.Fields.List {
		for _, name := range field.Names {
			fmt.Printf("FieldName: %#v\n", name.Name)
		}
		fmt.Printf("Field: %#v\n", field.Type)
	}

	panic("Panic")
}

func (fp *functionParser) parseSelectorExpr(expr *ast.SelectorExpr) (gotypes.DataType, error) {
	// X.Sel a.k.a Prefix.Item
	xDef, xErr := fp.parseExpr(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	if len(xDef) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}

	fmt.Printf("X: %#v\t%#v\n", xDef, expr.X)

	// The struct is the only data type from which a field is retriveable
	fmt.Printf("structExpr: %#v\n", xDef[0])

	var structIdent *gotypes.Identifier
	if typePointer, ok := xDef[0].(*gotypes.Pointer); ok {
		structIdent, ok = typePointer.Def.(*gotypes.Identifier)
		if !ok {
			return nil, fmt.Errorf("Trying to retrieve a %v field from a pointer to non-struct data type: %#v", expr.Sel.Name, typePointer.Def)
		}
	} else {
		structIdent, ok = xDef[0].(*gotypes.Identifier)
		if !ok {
			return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type: %#v", expr.Sel.Name, xDef[0])
		}
	}

	fmt.Printf("structIdent: %#v\n", structIdent)

	// Get struct's definition given by its identifier
	structDefsymbol, err := fp.symbolTable.Lookup(structIdent.Def)
	if err != nil {
		return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", structIdent.Def)
	}

	fmt.Printf("structDef: %#v\n", structDefsymbol)
	structDef, ok := structDefsymbol.Def.(*gotypes.Struct)
	if !ok {
		return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type: %#v", expr.Sel.Name, structDefsymbol.Def)
	}

	for _, item := range structDef.Fields {
		fmt.Printf("\tField: %v: %#v\n", item.Name, item.Def)
		if item.Name == expr.Sel.Name {
			fmt.Printf("Field %v found: %#v\n", expr.Sel.Name, item.Def)
			fp.allocatedSymbolsTable.AddDataTypeField(structDefsymbol.Package, structDefsymbol.Name, expr.Sel.Name)
			return item.Def, nil
		}
	}
	return nil, fmt.Errorf("Unable to find a field %v in struct %#v", expr.Sel.Name, structDefsymbol)
}

func (fp *functionParser) parseIndexExpr(expr *ast.IndexExpr) (gotypes.DataType, error) {
	// X[Index]
	// The Index can be a simple literal or another compound expression
	_, indexErr := fp.parseExpr(expr.Index)
	if indexErr != nil {
		return nil, indexErr
	}

	xDef, xErr := fp.parseExpr(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	if len(xDef) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}

	// Get definition of the X from the symbol Table (it must be a variable of a data type)
	// and get data type of its array/map members
	switch xType := xDef[0].(type) {
	case *gotypes.Identifier:
		return xType, nil
	case *gotypes.Map:
		return xType.Valuetype, nil
	case *gotypes.Array:
		return xType.Elmtype, nil
	case *gotypes.Slice:
		return xType.Elmtype, nil
	default:
		panic(fmt.Errorf("Unrecognized indexExpr type: %#v", xDef[0]))
	}
}

func (fp *functionParser) parseTypeAssertExpr(expr *ast.TypeAssertExpr) (gotypes.DataType, error) {
	// X.(Type)
	_, xErr := fp.parseExpr(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	// We should check if the data type really implements all methods of the interface.
	// Or we can assume it does and just return the Type itself
	// TODO(jchaloup): check the data type Type really implements interface of X (if it is an interface)

	fmt.Printf("TypeAssert type: %#v\n", expr.Type)

	// Assertion type can be an identifier or a pointer to an identifier.
	// Here, the symbol definition of the data type is not returned as it is lookup later by the caller
	switch typeType := expr.Type.(type) {
	case *ast.Ident:
		// TODO(jchaloup): check the type is not built-in type. If it is, return the &Builtin{}.
		return &gotypes.Identifier{
			Def: typeType.Name,
		}, nil
	case *ast.StarExpr:
		iDef, ok := typeType.X.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("TypeAssert type %#v is not a pointer to an identifier", expr.Type)
		}
		return &gotypes.Pointer{
			Def: &gotypes.Identifier{
				Def: iDef.Name,
			},
		}, nil
	default:
		return nil, fmt.Errorf("Unsupported TypeAssert type: %#v\n", typeType)
	}
}

func (fp *functionParser) parseExpr(expr ast.Expr) ([]gotypes.DataType, error) {
	// Given an expression we must always return its final data type
	// User defined symbols has its corresponding structs under parser/pkg/types.
	// In order to cover all possible symbol data types, we need to cover
	// golang language embedded data types as well
	fmt.Printf("Expr: %#v\n", expr)

	switch exprType := expr.(type) {
	// Basic literal carries
	case *ast.BasicLit:
		fmt.Printf("BasicLit: %#v\n", exprType)
		def, err := fp.parseBasicLit(exprType)
		return []gotypes.DataType{def}, err
	case *ast.CompositeLit:
		fmt.Printf("CompositeLit: %#v\n", exprType)
		def, err := fp.parseCompositeLit(exprType)
		return []gotypes.DataType{def}, err
	case *ast.Ident:
		fmt.Printf("Ident: %#v\n", exprType)
		def, err := fp.parseIdentifier(exprType)
		return []gotypes.DataType{def}, err
	case *ast.UnaryExpr:
		fmt.Printf("UnaryExpr: %#v\n", exprType)
		def, err := fp.parseUnaryExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.BinaryExpr:
		fmt.Printf("BinaryExpr: %#v\n", exprType)
		def, err := fp.parseBinaryExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.CallExpr:
		fmt.Printf("CallExpr: %#v\n", exprType)
		// If the call expression is the most most expression,
		// it may have a different number of results
		return fp.parseCallExpr(exprType)
	case *ast.StructType:
		fmt.Printf("StructType: %#v\n", exprType)
		def, err := fp.parseStructType(exprType)
		return []gotypes.DataType{def}, err
	case *ast.IndexExpr:
		fmt.Printf("IndexExpr: %#v\n", exprType)
		def, err := fp.parseIndexExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.SelectorExpr:
		fmt.Printf("SelectorExpr: %#v\n", exprType)
		def, err := fp.parseSelectorExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.TypeAssertExpr:
		fmt.Printf("TypeAssertExpr: %#v\n", exprType)
		def, err := fp.parseTypeAssertExpr(exprType)
		return []gotypes.DataType{def}, err
	default:
		return nil, fmt.Errorf("Unrecognized expression: %#v\n", expr)
	}
}

func (fp *functionParser) parseAssignment(expr *ast.AssignStmt) (gotypes.DataType, error) {
	// expr.Lhs = expr.Rhs
	// left-hand sice expression must be an identifier or a selector
	exprsSize := len(expr.Lhs)
	// Some assignments are of a different number of expression on both sides.
	// E.g. value, ok := somemap[key]
	// TODO(jchaloup): cover all the cases as well
	if exprsSize != len(expr.Rhs) {
		return nil, fmt.Errorf("Number of expression of the left-hand side differs from ones on the right-hand side for: %#v vs. %#v", expr.Lhs, expr.Rhs)
	}

	// If the assignment token is token.DEFINE a variable gets stored into the symbol table.
	// If it is already there and has the same type, do not do anything. Error if the type is different.
	// If it is not there yet, add it into the table
	// If the token is token.ASSIGN the variable must be in the symbol table.
	// If it is and has the same type, do not do anything, Error, if the type is different.
	// If it is not there yet, error.
	fmt.Printf("Ass token: %v %v %v\n", expr.Tok, token.ASSIGN, token.DEFINE)

	for i := 0; i < exprsSize; i++ {
		// If the left-hand side id a selector (e.g. struct.field), we alredy know data type of the id.
		// So, just store the field's data type into the allocated symbol table
		//switch lhsExpr := expr.
		fmt.Printf("Lhs: %#v\n", expr.Lhs[i])
		fmt.Printf("Rhs: %#v\n", expr.Rhs[i])

		def, err := fp.parseExpr(expr.Rhs[i])
		if err != nil {
			return nil, fmt.Errorf("Error when parsing Rhs[%v] expression of %#v: %v", i, expr, err)
		}

		if len(def) != 1 {
			return nil, fmt.Errorf("Assignment element at pos %v does not return a single result: %#v", i, def)
		}

		fmt.Printf("Ass type: %#v\n", def)

		switch lhsExpr := expr.Lhs[i].(type) {
		case *ast.Ident:
			fp.symbolTable.AddVariable(&gotypes.SymbolDef{
				Name:    lhsExpr.Name,
				Package: fp.packageName,
				Def:     def[0],
			})
		default:
			return nil, fmt.Errorf("Lhs of an assignment type %#v is not recognized", expr.Lhs[i])
		}
	}
	return nil, nil
}

func (fp *functionParser) parseStatement(statement ast.Stmt) error {
	switch stmtExpr := statement.(type) {
	case *ast.ReturnStmt:
		fmt.Printf("Return: %#v\n", stmtExpr)
		for _, result := range stmtExpr.Results {
			exprType, err := fp.parseExpr(result)
			if err != nil {
				panic(err)
				return err
			}
			fmt.Printf("====ExprType: %#v\n", exprType)
		}
	case *ast.AssignStmt:
		fmt.Printf("AssignStmt: %#v\n", stmtExpr)
		_, err := fp.parseAssignment(stmtExpr)
		if err != nil {
			panic(err)
			return err
		}
	}

	return nil
}

func (fp *functionParser) parseFuncBody(funcDecl *ast.FuncDecl) error {
	// Function/method signature is already stored in a symbol table.
	// From function/method's AST get its receiver, parameters and results,
	// construct a first level of a multi-level symbol table stack..
	// For each new block (including the body) push another level into the stack.
	fp.symbolTable.Push()

	if funcDecl.Recv != nil {
		// Receiver has a single parametr
		// https://golang.org/ref/spec#Receiver
		if len((*funcDecl.Recv).List) != 1 || len((*funcDecl.Recv).List[0].Names) != 1 {
			return fmt.Errorf("Receiver is not a single parameter")
		}

		def, err := fp.parseReceiver((*funcDecl.Recv).List[0].Type, true)
		if err != nil {
			return err
		}
		fp.symbolTable.AddVariable(&gotypes.SymbolDef{
			Name: (*funcDecl.Recv).List[0].Names[0].Name,
			Def:  def,
		})
	}

	if funcDecl.Type.Params != nil {
		for _, field := range funcDecl.Type.Params.List {
			def, err := fp.typesParser.parseTypeExpr(field.Type)
			if err != nil {
				return err
			}

			// field.Names is always non-empty if param's datatype is defined
			for _, name := range field.Names {
				fmt.Printf("Name: %v\n", name.Name)
				fp.symbolTable.AddVariable(&gotypes.SymbolDef{
					Name: name.Name,
					Def:  def,
				})
			}
		}
	}

	if funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			def, err := fp.typesParser.parseTypeExpr(field.Type)
			if err != nil {
				return err
			}

			for _, name := range field.Names {
				fmt.Printf("Name: %v\n", name.Name)
				fp.symbolTable.AddVariable(&gotypes.SymbolDef{
					Name: name.Name,
					Def:  def,
				})
			}
		}
	}

	fp.symbolTable.Push()
	byteSlice, err := json.Marshal(fp.symbolTable)
	fmt.Printf("\nTable: %v\nerr: %v", string(byteSlice), err)

	// The stack will always have at least one symbol table (with receivers, resp. parameters, resp. results)
	for _, statement := range funcDecl.Body.List {
		fmt.Printf("\n\nstatement: %#v\n", statement)
		if err := fp.parseStatement(statement); err != nil {
			return err
		}
	}

	//stack.Print()
	fp.symbolTable.Pop()
	fp.symbolTable.Pop()

	return nil
}
