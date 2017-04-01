package parser

import (
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

// functionParser parses declaration and definition of a function/method
type functionParser struct {
	// per file symbol table
	symbolTable *symboltable.Table
	// per file allocatable ST
	allocatedSymbolsTable *AllocatedSymbolsTable
	// types parser
	typesParser *typesParser
}

// NewFunctionParser create an instance of a function parser
func NewFunctionParser(symbolTable *symboltable.Table, allocatedSymbolsTable *AllocatedSymbolsTable, typesParser *typesParser) *functionParser {
	return &functionParser{
		symbolTable:           symbolTable,
		allocatedSymbolsTable: allocatedSymbolsTable,
		typesParser:           typesParser,
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

		recDef, err := fp.typesParser.parseTypeExpr((*d.Recv).List[0].Type)
		if err != nil {
			return nil, err
		}

		methodDef := &gotypes.Method{
			Def:      funcDef,
			Receiver: recDef,
		}

		fp.symbolTable.AddFunction(&gotypes.SymbolDef{
			Name: d.Name.Name,
			Def:  methodDef,
		})

		//printDataType(methodDef)

		return methodDef, nil
	}

	// Empty receiver => Function
	fp.symbolTable.AddFunction(&gotypes.SymbolDef{
		Name: d.Name.Name,
		Def:  funcDef,
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

	// TODO(jchaloup): put the identifier into the allocated symbol table
	return &gotypes.Identifier{
		Def: ident.Name,
	}, nil
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

func (fp *functionParser) parseCallExpr(expr *ast.CallExpr) ([]gotypes.DataType, error) {
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

	fmt.Printf("%#v\n", expr.Fun)
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
		// TODO(jchaloup): get the symbol's origin from the higher parser structures
		fp.allocatedSymbolsTable.AddSymbol("", exprType.Name)

		switch funcType := def.Def.(type) {
		case *gotypes.Function:
			return funcType.Results, nil
		case *gotypes.Method:
			return funcType.Def.(*gotypes.Function).Results, nil
		default:
			return nil, fmt.Errorf("Symbol to be called is not a function")
		}
	}

	return nil, fmt.Errorf("Call expr not recognized: %#v", expr)
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

	stack := symboltable.NewStack()

	stack.Push()

	if funcDecl.Recv != nil {
		// Receiver has a single parametr
		// https://golang.org/ref/spec#Receiver
		if len((*funcDecl.Recv).List) != 1 || len((*funcDecl.Recv).List[0].Names) != 1 {
			return fmt.Errorf("Receiver is not a single parameter")
		}

		def, err := fp.typesParser.parseTypeExpr((*funcDecl.Recv).List[0].Type)
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
			def, err := fp.typesParser.parseTypeExpr(field.Type)
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
			def, err := fp.typesParser.parseTypeExpr(field.Type)
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
		fmt.Printf("\n\nstatement: %#v\n", statement)
		if err := fp.parseStatement(statement); err != nil {
			return err
		}
	}

	//stack.Print()

	// Here!!! The symbol type analysis is carried here!!! Yes, HERE!!!

	return nil
}
