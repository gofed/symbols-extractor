package statement

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

// Parser parses go statements, e.g. block, declaration and definition of a function/method
type Parser struct {
	// package name
	packageName string
	// per file symbol table
	symbolTable *symboltable.Stack
	// per file allocatable ST
	allocatedSymbolsTable *alloctable.Table
	// types parser
	typeParser *typeparser.Parser
	// expresesion parser
	exprParser *exprparser.Parser
}

// New creates an instance of a statement parser
func New(packageName string, symbolTable *symboltable.Stack, allocatedSymbolsTable *alloctable.Table, typeParser *typeparser.Parser, exprParser *exprparser.Parser) *Parser {
	return &Parser{
		packageName:           packageName,
		symbolTable:           symbolTable,
		allocatedSymbolsTable: allocatedSymbolsTable,
		typeParser:            typeParser,
		exprParser:            exprParser,
	}
}

func (sp *Parser) ParseFuncDecl(d *ast.FuncDecl) (gotypes.DataType, error) {
	// parseFunction does not store name of params, resp. results
	// as the names are not important. Just params, resp. results ordering is.
	// Thus, this method is used to parse function's signature only.
	funcDef, err := sp.typeParser.ParseFunction(d.Type)
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

		recDef, err := sp.exprParser.ParseReceiver((*d.Recv).List[0].Type, false)
		if err != nil {
			return nil, err
		}

		methodDef := &gotypes.Method{
			Def:      funcDef,
			Receiver: recDef,
		}

		sp.symbolTable.AddFunction(&gotypes.SymbolDef{
			Name:    d.Name.Name,
			Package: sp.packageName,
			Def:     methodDef,
		})

		//printDataType(methodDef)

		return methodDef, nil
	}

	// Empty receiver => Function
	sp.symbolTable.AddFunction(&gotypes.SymbolDef{
		Name:    d.Name.Name,
		Package: sp.packageName,
		Def:     funcDef,
	})

	//printDataType(funcDef)

	return funcDef, nil
}

func (sp *Parser) parseDeclStmt(statement *ast.DeclStmt) error {
	// expr.
	fmt.Printf("decl: %#v\n", statement.Decl)
	panic("Decl panic")
	return nil
}

func (sp *Parser) parseAssignment(expr *ast.AssignStmt) error {
	// expr.Lhs = expr.Rhs
	// left-hand sice expression must be an identifier or a selector
	exprsSize := len(expr.Lhs)
	// Some assignments are of a different number of expression on both sides.
	// E.g. value, ok := somemap[key]
	// TODO(jchaloup): cover all the cases as well
	if exprsSize != len(expr.Rhs) {
		return fmt.Errorf("Number of expression of the left-hand side differs from ones on the right-hand side for: %#v vs. %#v", expr.Lhs, expr.Rhs)
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

		def, err := sp.exprParser.ParseExpr(expr.Rhs[i])
		if err != nil {
			return fmt.Errorf("Error when parsing Rhs[%v] expression of %#v: %v", i, expr, err)
		}

		if len(def) != 1 {
			return fmt.Errorf("Assignment element at pos %v does not return a single result: %#v", i, def)
		}

		fmt.Printf("Ass type: %#v\n", def)

		switch lhsExpr := expr.Lhs[i].(type) {
		case *ast.Ident:
			// skip the anonymous variables
			if lhsExpr.Name == "_" {
				continue
			}

			sp.symbolTable.AddVariable(&gotypes.SymbolDef{
				Name:    lhsExpr.Name,
				Package: sp.packageName,
				Def:     def[0],
			})
		default:
			return fmt.Errorf("Lhs of an assignment type %#v is not recognized", expr.Lhs[i])
		}
	}
	return nil
}

func (sp *Parser) parseExprStmt(statement *ast.ExprStmt) error {
	_, err := sp.exprParser.ParseExpr(statement.X)
	return err
}

func (sp *Parser) parseIfStmt(statement *ast.IfStmt) error {
	// If Init; Cond { Body } Else

	// The Init part is basically another block
	fmt.Printf("\nInit: %#v\n", statement.Init)
	if statement.Init != nil {
		// The Init part must be an assignment statement
		if _, ok := statement.Init.(*ast.AssignStmt); !ok {
			return fmt.Errorf("If Init part must by an assignment statement if set")
		}
		sp.symbolTable.Push()
		defer sp.symbolTable.Pop()
		if err := sp.parseAssignment(statement.Init.(*ast.AssignStmt)); err != nil {
			return err
		}
	}

	fmt.Printf("\nCond: %#v\n", statement.Cond)
	_, err := sp.exprParser.ParseExpr(statement.Cond)
	if err != nil {
		return err
	}

	// Process the If-body
	fmt.Printf("\nBody: %#v\n", statement.Body)
	sp.parseBlockStmt(statement.Body)

	return nil
}

func (sp *Parser) parseBlockStmt(statement *ast.BlockStmt) error {
	sp.symbolTable.Push()
	defer sp.symbolTable.Pop()

	fmt.Printf("Block statement: %#v\n", statement)
	for _, blockItem := range statement.List {
		fmt.Printf("BodyItem: %#v\n", blockItem)
		if err := sp.parseStatement(blockItem); err != nil {
			return nil
		}
	}
	return nil
}

// Statement grammer at https://golang.org/ref/spec#Statements
//
// Statement =
// 	Declaration | LabeledStmt | SimpleStmt |
// 	GoStmt | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt |
// 	FallthroughStmt | Block | IfStmt | SwitchStmt | SelectStmt | ForStmt |
// 	DeferStmt .
//
// SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .

func (sp *Parser) parseStatement(statement ast.Stmt) error {
	switch stmtExpr := statement.(type) {
	case *ast.DeclStmt:
		fmt.Printf("DeclStmt: %#v\n", stmtExpr)
		return sp.parseDeclStmt(stmtExpr)
	case *ast.ReturnStmt:
		fmt.Printf("Return: %#v\n", stmtExpr)
		for _, result := range stmtExpr.Results {
			exprType, err := sp.exprParser.ParseExpr(result)
			if err != nil {
				return err
			}
			fmt.Printf("====ExprType: %#v\n", exprType)
		}
	case *ast.ExprStmt:
		fmt.Printf("ExprStmt: %#v\n", stmtExpr)
		return sp.parseExprStmt(stmtExpr)
	case *ast.AssignStmt:
		fmt.Printf("AssignStmt: %#v\n", stmtExpr)
		return sp.parseAssignment(stmtExpr)
	case *ast.BlockStmt:
		fmt.Printf("BlockStmt: %#v\n", stmtExpr)
		return sp.parseBlockStmt(stmtExpr)
	case *ast.IfStmt:
		fmt.Printf("IfStmt: %#v\n", stmtExpr)
		return sp.parseIfStmt(stmtExpr)
	default:
		panic(fmt.Errorf("Unknown statement %#v", statement))
	}

	return nil
}

func (sp *Parser) parseFuncHeadVariables(funcDecl *ast.FuncDecl) error {
	if funcDecl.Recv != nil {
		// Receiver has a single parametr
		// https://golang.org/ref/spec#Receiver
		if len((*funcDecl.Recv).List) != 1 || len((*funcDecl.Recv).List[0].Names) != 1 {
			return fmt.Errorf("Receiver is not a single parameter")
		}

		def, err := sp.exprParser.ParseReceiver((*funcDecl.Recv).List[0].Type, true)
		if err != nil {
			return err
		}
		sp.symbolTable.AddVariable(&gotypes.SymbolDef{
			Name: (*funcDecl.Recv).List[0].Names[0].Name,
			Def:  def,
		})
	}

	if funcDecl.Type.Params != nil {
		for _, field := range funcDecl.Type.Params.List {
			def, err := sp.typeParser.ParseTypeExpr(field.Type)
			if err != nil {
				return err
			}

			// field.Names is always non-empty if param's datatype is defined
			for _, name := range field.Names {
				fmt.Printf("Name: %v\n", name.Name)
				sp.symbolTable.AddVariable(&gotypes.SymbolDef{
					Name: name.Name,
					Def:  def,
				})
			}
		}
	}

	if funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			def, err := sp.typeParser.ParseTypeExpr(field.Type)
			if err != nil {
				return err
			}

			for _, name := range field.Names {
				fmt.Printf("Name: %v\n", name.Name)
				sp.symbolTable.AddVariable(&gotypes.SymbolDef{
					Name: name.Name,
					Def:  def,
				})
			}
		}
	}
	return nil
}

func (sp *Parser) ParseFuncBody(funcDecl *ast.FuncDecl) error {
	// Function/method signature is already stored in a symbol table.
	// From function/method's AST get its receiver, parameters and results,
	// construct a first level of a multi-level symbol table stack..
	// For each new block (including the body) push another level into the stack.
	sp.symbolTable.Push()
	if err := sp.parseFuncHeadVariables(funcDecl); err != nil {
		return nil
	}
	sp.symbolTable.Push()
	byteSlice, err := json.Marshal(sp.symbolTable)
	fmt.Printf("\nTable: %v\nerr: %v", string(byteSlice), err)

	// The stack will always have at least one symbol table (with receivers, resp. parameters, resp. results)
	for _, statement := range funcDecl.Body.List {
		fmt.Printf("\n\nstatement: %#v\n", statement)
		if err := sp.parseStatement(statement); err != nil {
			panic(err)
			return err
		}
	}

	//stack.Print()
	sp.symbolTable.Pop()
	sp.symbolTable.Pop()

	return nil
}
