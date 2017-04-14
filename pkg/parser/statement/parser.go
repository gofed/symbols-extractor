package statement

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
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
	typeParser types.TypeParser
	// expresesion parser
	exprParser types.ExpressionParser
}

// New creates an instance of a statement parser
func New(packageName string, symbolTable *symboltable.Stack, allocatedSymbolsTable *alloctable.Table, typeParser types.TypeParser, exprParser types.ExpressionParser) types.StatementParser {
	return &Parser{
		packageName:           packageName,
		symbolTable:           symbolTable,
		allocatedSymbolsTable: allocatedSymbolsTable,
		typeParser:            typeParser,
		exprParser:            exprParser,
	}
}

func (ep *Parser) parseReceiver(receiver ast.Expr, skip_allocated bool) (gotypes.DataType, error) {
	// Receiver's type must be of the form T or *T (possibly using parentheses) where T is a type name.
	switch typedExpr := receiver.(type) {
	case *ast.Ident:
		// search the identifier in the symbol table
		def, err := ep.symbolTable.Lookup(typedExpr.Name)
		if err != nil {
			fmt.Printf("Lookup error: %v\n", err)
			// Return an error so the function body processing can be postponed
			// TODO(jchaloup): return more information about the missing symbol so the
			// body can be re-processed right after the symbol is stored into the symbol table.
			return nil, err
		}

		if !skip_allocated {
			ep.allocatedSymbolsTable.AddSymbol(def.Package, typedExpr.Name)
		}

		return &gotypes.Identifier{
			Def: typedExpr.Name,
		}, nil
	case *ast.StarExpr:
		fmt.Printf("Start: %#v\n", typedExpr)
		switch idExpr := typedExpr.X.(type) {
		case *ast.Ident:
			// search the identifier in the symbol table
			def, err := ep.symbolTable.Lookup(idExpr.Name)
			if err != nil {
				fmt.Printf("Lookup error: %v\n", err)
				// Return an error so the function body processing can be postponed
				// TODO(jchaloup): return more information about the missing symbol so the
				// body can be re-processed right after the symbol is stored into the symbol table.
				return nil, err
			}

			if !skip_allocated {
				ep.allocatedSymbolsTable.AddSymbol(def.Package, idExpr.Name)
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

func (sp *Parser) ParseFuncDecl(d *ast.FuncDecl) (gotypes.DataType, error) {
	// parseFunction does not store name of params, resp. results
	// as the names are not important. Just params, resp. results ordering is.
	// Thus, this method is used to parse function's signature only.
	funcDef, err := sp.typeParser.Parse(d.Type)
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

	if d.Recv == nil {
		// Empty receiver => Function
		return funcDef, nil
	}

	// Receiver has a single parametr
	// https://golang.org/ref/spec#Receiver
	if len((*d.Recv).List) != 1 || len((*d.Recv).List[0].Names) != 1 {
		return nil, fmt.Errorf("Receiver is not a single parameter")
	}

	//fmt.Printf("Rec Name: %#v\n", (*d.Recv).List[0].Names[0].Name)

	recDef, err := sp.parseReceiver((*d.Recv).List[0].Type, false)
	if err != nil {
		return nil, err
	}

	methodDef := &gotypes.Method{
		Def:      funcDef,
		Receiver: recDef,
	}

	return methodDef, nil
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
		if err := sp.Parse(statement); err != nil {
			panic(err)
			return err
		}
	}

	//stack.Print()
	sp.symbolTable.Pop()
	sp.symbolTable.Pop()

	return nil
}

func (sp *Parser) ParseValueSpec(spec *ast.ValueSpec) ([]*gotypes.SymbolDef, error) {
	fmt.Printf("ValueSpec: %#v\n", spec)
	fmt.Printf("ValueSpec.Names: %#v\n", spec.Names)
	fmt.Printf("ValueSpec.Type: %#v\n", spec.Type)
	fmt.Printf("ValueSpec.Values: %#v\n", spec.Values)
	nLen := len(spec.Names)
	vLen := len(spec.Values)

	fmt.Printf("(%v, %v)\n", nLen, vLen)

	if nLen < vLen {
		return nil, fmt.Errorf("ValueSpec %#v has less number of identifieries on LHS (%v) than a number of expressions on RHS (%v)", spec, nLen, vLen)
	}

	var typeDef gotypes.DataType
	if spec.Type != nil {
		def, err := sp.typeParser.Parse(spec.Type)
		if err != nil {
			return nil, err
		}
		typeDef = def
	}

	fmt.Printf("typeDef: %#v\n", typeDef)

	for i := 0; i < vLen; i++ {
		if typeDef == nil && spec.Values[i] == nil {
			return nil, fmt.Errorf("No type nor value in ValueSpec declaration")
		}
		// TODO(jchaloup): if the variable type is an interface and the variable value type is a concrete type
		//                 note somewhere the concrete type must implemented the interface
		valueExpr, err := sp.exprParser.Parse(spec.Values[i])
		if err != nil {
			return nil, err
		}
		if len(valueExpr) != 1 {
			return nil, fmt.Errorf("Expecting a single expression. Got a list instead: %#v", valueExpr)
		}
		fmt.Printf("Name: %#v, Value: %#v, Type: %#v", spec.Names[i].Name, valueExpr[0], typeDef)
		// Put the variables/consts into the symbol table
		if spec.Names[i].Name == "_" {
			continue
		}
		if typeDef != nil {
			sp.symbolTable.AddVariable(&gotypes.SymbolDef{
				Name:    spec.Names[i].Name,
				Package: sp.packageName,
				Def:     typeDef,
			})
		} else {
			sp.symbolTable.AddVariable(&gotypes.SymbolDef{
				Name:    spec.Names[i].Name,
				Package: sp.packageName,
				Def:     valueExpr[0],
			})
		}
	}

	// TODO(jchaloup): return a list of SymbolDefs
	var symbolsDef = make([]*gotypes.SymbolDef, 0)

	for i := vLen; i < nLen; i++ {
		if typeDef == nil {
			return nil, fmt.Errorf("No type in ValueSpec declaration for identifier at pos %v (starting index from 1)", i+1)
		}
		if spec.Names[i].Name == "_" {
			continue
		}
		symbolsDef = append(symbolsDef, &gotypes.SymbolDef{
			Name:    spec.Names[i].Name,
			Package: sp.packageName,
			Def:     typeDef,
		})
	}

	return symbolsDef, nil
}

func (sp *Parser) parseDeclStmt(statement *ast.DeclStmt) error {
	// expr.
	fmt.Printf("decl: %#v\n", statement.Decl)
	switch decl := statement.Decl.(type) {
	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			fmt.Printf("gendecl.spec: %#v\n", spec)
			switch genDeclSpec := spec.(type) {
			case *ast.ValueSpec:
				defs, err := sp.ParseValueSpec(genDeclSpec)
				if err != nil {
					return err
				}
				for _, def := range defs {
					// TPDP(jchaloup): we should store all variables or non.
					// Given the error is set only if the variable already exists, it should not matter so much.
					if err := sp.symbolTable.AddVariable(def); err != nil {
						return nil
					}
				}
			}
		}
	}
	//panic("Decl panic")
	return nil
}

func (sp *Parser) parseLabeledStmt(statement *ast.LabeledStmt) error {
	// the label is typeless
	return sp.Parse(statement.Stmt)
}

func (sp *Parser) parseExprStmt(statement *ast.ExprStmt) error {
	_, err := sp.exprParser.Parse(statement.X)
	return err
}

func (sp *Parser) parseSendStmt(statement *ast.SendStmt) error {
	if _, err := sp.exprParser.Parse(statement.Chan); err != nil {
		return err
	}
	// TODO(jchaloup): should check the statement.Chan type is really a channel.
	if _, err := sp.exprParser.Parse(statement.Value); err != nil {
		return err
	}
	return nil
}

func (sp *Parser) parseIncDecStmt(statement *ast.IncDecStmt) error {
	// both --,++ has no type information
	// TODO(jchaloup): check the --/++ can be carried over the statement.X
	_, err := sp.exprParser.Parse(statement.X)
	return err
}

func (sp *Parser) parseAssignStmt(statement *ast.AssignStmt) error {
	// expr.Lhs = expr.Rhs
	// left-hand sice expression must be an identifier or a selector
	exprsSize := len(statement.Lhs)
	// Some assignments are of a different number of expression on both sides.
	// E.g. value, ok := somemap[key]
	// TODO(jchaloup): cover all the cases as well
	if exprsSize != len(statement.Rhs) {
		return fmt.Errorf("Number of expression of the left-hand side differs from ones on the right-hand side for: %#v vs. %#v", statement.Lhs, statement.Rhs)
	}

	// If the assignment token is token.DEFINE a variable gets stored into the symbol table.
	// If it is already there and has the same type, do not do anything. Error if the type is different.
	// If it is not there yet, add it into the table
	// If the token is token.ASSIGN the variable must be in the symbol table.
	// If it is and has the same type, do not do anything, Error, if the type is different.
	// If it is not there yet, error.
	fmt.Printf("Ass token: %v %v %v\n", statement.Tok, token.ASSIGN, token.DEFINE)

	for i := 0; i < exprsSize; i++ {
		// If the left-hand side id a selector (e.g. struct.field), we alredy know data type of the id.
		// So, just store the field's data type into the allocated symbol table
		//switch lhsExpr := expr.
		fmt.Printf("Lhs: %#v\n", statement.Lhs[i])
		fmt.Printf("Rhs: %#v\n", statement.Rhs[i])

		def, err := sp.exprParser.Parse(statement.Rhs[i])
		if err != nil {
			return fmt.Errorf("Error when parsing Rhs[%v] expression of %#v: %v", i, statement, err)
		}

		if len(def) != 1 {
			return fmt.Errorf("Assignment element at pos %v does not return a single result: %#v", i, def)
		}

		fmt.Printf("Ass type: %#v\n", def)

		switch lhsExpr := statement.Lhs[i].(type) {
		case *ast.Ident:
			// skip the anonymous variables
			if lhsExpr.Name == "_" {
				continue
			}

			// TODO(jchaloup): If the statement.Tok is not token.DEFINE, don't add the variable to the symbol table.
			//                 Instead, check the varible is of the same type (or compatible) as the already stored one.
			sp.symbolTable.AddVariable(&gotypes.SymbolDef{
				Name:    lhsExpr.Name,
				Package: sp.packageName,
				Def:     def[0],
			})
		default:
			return fmt.Errorf("Lhs of an assignment type %#v is not recognized", statement.Lhs[i])
		}
	}
	return nil
}

func (sp *Parser) parseGoStmt(statement *ast.GoStmt) error {
	_, err := sp.exprParser.Parse(statement.Call)
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
		if err := sp.parseAssignStmt(statement.Init.(*ast.AssignStmt)); err != nil {
			return err
		}
	}

	fmt.Printf("\nCond: %#v\n", statement.Cond)
	_, err := sp.exprParser.Parse(statement.Cond)
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
		if err := sp.Parse(blockItem); err != nil {
			return err
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

func (sp *Parser) Parse(statement ast.Stmt) error {
	switch stmtExpr := statement.(type) {
	case *ast.DeclStmt:
		fmt.Printf("DeclStmt: %#v\n", stmtExpr)
		return sp.parseDeclStmt(stmtExpr)
	case *ast.LabeledStmt:
		fmt.Printf("LabeledStmt: %#v\n", stmtExpr)
		return sp.parseLabeledStmt(stmtExpr)
	case *ast.ExprStmt:
		fmt.Printf("ExprStmt: %#v\n", stmtExpr)
		return sp.parseExprStmt(stmtExpr)
	case *ast.SendStmt:
		fmt.Printf("SendStmt: %#v\n", stmtExpr)
		return sp.parseSendStmt(stmtExpr)
	case *ast.IncDecStmt:
		fmt.Printf("IncDecStmt: %#v\n", stmtExpr)
		return sp.parseIncDecStmt(stmtExpr)
	case *ast.AssignStmt:
		fmt.Printf("AssignStmt: %#v\n", stmtExpr)
		return sp.parseAssignStmt(stmtExpr)
	case *ast.GoStmt:
		fmt.Printf("GoStmt: %#v\n", stmtExpr)
		return sp.parseGoStmt(stmtExpr)
	case *ast.ReturnStmt:
		fmt.Printf("Return: %#v\n", stmtExpr)
		for _, result := range stmtExpr.Results {
			exprType, err := sp.exprParser.Parse(result)
			if err != nil {
				return err
			}
			fmt.Printf("====ExprType: %#v\n", exprType)
		}
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

		def, err := sp.parseReceiver((*funcDecl.Recv).List[0].Type, true)
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
			def, err := sp.typeParser.Parse(field.Type)
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
			def, err := sp.typeParser.Parse(field.Type)
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
