package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"

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

var builtinTypes = []string{
	// TODO(jchaloup): extend the list with all built-in types
	"string", "int", "error",
	"uint64",
}

func isBuiltin(ident string) bool {
	for _, id := range builtinTypes {
		if id == ident {
			return true
		}
	}
	return false
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

func (fp *functionParser) parseIdentifier(ident *ast.Ident) (gotypes.DataType, error) {
	// TODO(jchaloup): put the nil into the allocated symbol table
	if ident.Name == "nil" {
		return &gotypes.Nil{}, nil
	}

	// Check if the symbol is in the symbol table.
	// It is either a local variable or a global variable (as it is used inside an expression).
	// If the symbol is not found, it means it is not defined and never will be.
	def, err := fp.symbolTable.Lookup(ident.Name)
	if err != nil {
		fmt.Printf("Lookup error: %v\n", err)
		// Return an error so the function body processing can be postponed
		// TODO(jchaloup): return more information about the missing symbol so the
		// body can be re-processed right after the symbol is stored into the symbol table.
		return nil, err
	}

	// TODO(jchaloup): put the identifier's type into the allocated symbol table
	fmt.Printf("Symbol used: %v.%v\n", def.Package, def.Name)
	return def.Def, nil
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

func (fp *functionParser) getDataTypeField(def gotypes.DataType, field string) (gotypes.DataType, error) {
	// The struct is the only data type from which a field is retriveable

	var structDef *gotypes.SymbolDef

	switch expr := def.(type) {
	case *gotypes.Builtin:
		return nil, fmt.Errorf("Cannot retrieve field %v from a built-in type", field)
	case *gotypes.Identifier:
		// If the def is an identifier, retrieve struct's definition from the symbol table
		def, err := fp.symbolTable.Lookup(expr.Def)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", expr.Def)
		}
		fmt.Printf("UsingI: %v.%v\n", def.Package, def.Name)
		fp.allocatedSymbolsTable.AddSymbol(def.Package, def.Name)
		structDef = def
	case *gotypes.Pointer:
		iDef, ok := expr.Def.(*gotypes.Identifier)
		if !ok {
			return nil, fmt.Errorf("Cannot retrieve field %v from a pointer pointing to non-Identifier", field)
		}
		def, err := fp.symbolTable.Lookup(iDef.Def)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", expr.Def)
		}
		structDef = def
	default:
		return nil, fmt.Errorf("Unable to recognize access field expression: %#v", def)
	}

	fmt.Printf("Struct: %#v\n", structDef)

	structExpr, ok := structDef.Def.(*gotypes.Struct)
	if !ok {
		return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type: %#v", field, structDef.Def)
	}

	for _, item := range structExpr.Fields {
		fmt.Printf("\tField: %v: %#v\n", item.Name, item.Def)
		if item.Name == field {
			fmt.Printf("Field %v found: %#v\n", field, item.Def)
			fp.allocatedSymbolsTable.AddDataTypeField(structDef.Package, structDef.Name, field)
			return item.Def, nil
		}
	}
	return nil, fmt.Errorf("Unable to find a field %v in struct %#v", field, structExpr)
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

	fmt.Printf("X: %#v\n", xDef)
	// It is either a data type or it is a pointer to a data type
	// xExpr, ok := xDef[0].(*gotypes.Pointer)
	// if ok {
	// 	return fp.getDataTypeField(xExpr.Def, expr.Sel.Name)
	// } else {
	// 	return fp.getDataTypeField(xDef[0], expr.Sel.Name)
	// }
	return fp.getDataTypeField(xDef[0], expr.Sel.Name)
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
	case *ast.Ident:
		fmt.Printf("Ident: %#v\n", exprType)
		def, err := fp.parseIdentifier(exprType)
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
