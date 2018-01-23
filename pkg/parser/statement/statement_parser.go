package statement

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/propagation"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/symbols"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

// Parser parses go statements, e.g. block, declaration and definition of a function/method
type Parser struct {
	*types.Config
	lastConstType gotypes.DataType
}

// New creates an instance of a statement parser
func New(config *types.Config) types.StatementParser {
	return &Parser{
		Config: config,
	}
}

func (ep *Parser) parseReceiver(receiver ast.Expr, skip_allocated bool) (gotypes.DataType, error) {
	glog.V(2).Infof("Processing Receiver: %#v\n", receiver)
	// Receiver's type must be of the form T or *T (possibly using parentheses) where T is a type name.
	switch typedExpr := receiver.(type) {
	case *ast.Ident:
		// search the identifier in the symbol table
		def, _, err := ep.SymbolTable.Lookup(typedExpr.Name)
		if err != nil {
			// Return an error so the function body processing can be postponed
			// TODO(jchaloup): return more information about the missing symbol so the
			// body can be re-processed right after the symbol is stored into the symbol table.
			return nil, err
		}

		if !skip_allocated {
			ep.AllocatedSymbolsTable.AddDataType(def.Package, typedExpr.Name, ep.Config.SymbolPos(typedExpr.Pos()))
		}

		return &gotypes.Identifier{
			Def:     typedExpr.Name,
			Package: ep.PackageName,
		}, nil
	case *ast.StarExpr:
		switch idExpr := typedExpr.X.(type) {
		case *ast.Ident:
			// search the identifier in the symbol table
			def, _, err := ep.SymbolTable.Lookup(idExpr.Name)
			if err != nil {
				// Return an error so the function body processing can be postponed
				// TODO(jchaloup): return more information about the missing symbol so the
				// body can be re-processed right after the symbol is stored into the symbol table.
				return nil, err
			}

			if !skip_allocated {
				ep.AllocatedSymbolsTable.AddDataType(def.Package, idExpr.Name, ep.Config.SymbolPos(idExpr.Pos()))
			}

			return &gotypes.Pointer{
				Def: &gotypes.Identifier{
					Def:     idExpr.Name,
					Package: ep.PackageName,
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
	if d.Name != nil {
		glog.V(2).Infof("Processing function %q declaration: %#v\n", d.Name.Name, d)
	} else {
		glog.V(2).Infof("Processing function declaration: %#v\n", d)
	}
	// parseFunction does not store name of params, resp. results
	// as the names are not important. Just params, resp. results ordering is.
	// Thus, this method is used to parse function's signature only.
	funcDef, err := sp.TypeParser.Parse(d.Type)
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
	if len(d.Recv.List) != 1 || len(d.Recv.List[0].Names) > 2 {
		if len(d.Recv.List) != 1 {
			return nil, fmt.Errorf("Method %q has no receiver", d.Name.Name)
		}
		return nil, fmt.Errorf("Receiver is not a single parameter: %#v, %#v", d.Recv, d.Recv.List[0].Names)
	}

	recDef, err := sp.parseReceiver(d.Recv.List[0].Type, false)
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
	if funcDecl.Name == nil {
		glog.V(2).Infof("Processing function body of %#v\n", funcDecl)
	} else {
		glog.V(2).Infof("Processing function body of %q: %#v\n", funcDecl.Name.Name, funcDecl)
	}
	// Function/method signature is already stored in a symbol table.
	// From function/method's AST get its receiver, parameters and results,
	// construct a first level of a multi-level symbol table stack..
	// For each new block (including the body) push another level into the stack.
	sp.SymbolTable.Push()
	if err := sp.parseFuncHeadVariables(funcDecl); err != nil {
		sp.SymbolTable.Pop()
		return fmt.Errorf("sp.ParseFuncBody: %v", err)
	}
	sp.SymbolTable.Push()
	defer func() {
		sp.SymbolTable.Pop()
		sp.SymbolTable.Pop()
	}()
	// if funcDecl.Body is nil, then the function/method is declared only and its definition
	// is most likely in .s file(s)
	if funcDecl.Body == nil {
		return nil
	}

	// The stack will always have at least one symbol table (with receivers, resp. parameters, resp. results)
	for _, statement := range funcDecl.Body.List {
		if err := sp.Parse(statement); err != nil {
			return err
		}
	}

	return nil
}

func (sp *Parser) ParseConstValueSpec(constSpec types.ConstSpec) ([]*symbols.SymbolDef, error) {
	var symbolsDef = make([]*symbols.SymbolDef, 0)

	sp.Config.Iota = constSpec.IotaIdx
	var names []string
	spec := constSpec.Spec
	for _, name := range spec.Names {
		names = append(names, name.Name)
	}
	glog.V(2).Infof("\n\nProcessing const value spec: %#v\n\tNames: %#v\n", spec, strings.Join(names, ","))

	var typeDef gotypes.DataType
	if spec.Type != nil {
		def, err := sp.TypeParser.Parse(spec.Type)
		if err != nil {
			return nil, err
		}
		typeDef = def
	}

	nLen := len(spec.Names)
	vLen := len(spec.Values)

	if nLen != vLen {
		return nil, fmt.Errorf("ValueSpec %#v has different number of identifiers on LHS (%v) than a number of results of invocation on RHS (%v)", spec, nLen, vLen)
	}

	// https://splice.com/blog/iota-elegant-constants-golang/
	for i := 0; i < vLen; i++ {
		glog.V(2).Infof("----Processing ast.ValueSpec[%v]: %#v of type %#v\n", i, spec.Values[i], spec.Type)
		valueExprAttr, err := sp.ExprParser.Parse(spec.Values[i])
		if err != nil {
			return nil, err
		}
		if len(valueExprAttr.DataTypeList) != 1 {
			return nil, fmt.Errorf("Expecting a single expression. Got a list instead: %#v", valueExprAttr.DataTypeList)
		}
		if valueExprAttr.DataTypeList[0].GetType() != gotypes.ConstantType {
			return nil, fmt.Errorf("Expecting a constant expression. Got %#v instead", valueExprAttr.DataTypeList[0])
		}
		if spec.Names[i].Name == "_" {
			continue
		}
		sDef := &symbols.SymbolDef{
			Name:    spec.Names[i].Name,
			Package: sp.PackageName,
			Def:     nil,
			Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, spec.Names[i].Pos()),
		}
		if sp.SymbolTable.CurrentLevel() > 0 {
			sDef.Package = ""
		}
		if typeDef == nil {
			// untyped constant (unless the type is propagates through the value)
			sDef.Def = valueExprAttr.DataTypeList[0]
			sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
				X:            valueExprAttr.TypeVarList[0],
				Y:            typevars.VariableFromSymbolDef(sDef),
				ExpectedType: sDef.Def,
				Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, spec.Names[i].Pos()),
			})
		} else {
			// typeDef must be identifier
			if typeDef.GetType() != gotypes.IdentifierType && typeDef.GetType() != gotypes.SelectorType {
				return nil, fmt.Errorf("Expected a type %#v of a constant to be an identifier. Got %v instead", typeDef, typeDef.GetType())
			}

			// typed constant, we must check the types are compatible
			bIdent, err := sp.Config.SymbolsAccessor.TypeToSimpleBuiltin(typeDef)
			if err != nil {
				return nil, err
			}

			if bIdent.Package == "builtin" && sp.Config.SymbolsAccessor.IsIntegral(bIdent.Def) {
				literal := valueExprAttr.DataTypeList[0].(*gotypes.Constant).Literal
				if literal != "*" {
					ma := propagation.NewMultiArith()
					ma.AddXFromString(
						valueExprAttr.DataTypeList[0].(*gotypes.Constant).Literal,
					)
					lit, err := ma.XToLiteral(bIdent.Def)

					if err != nil {
						return nil, err
					}
					valueExprAttr.DataTypeList[0].(*gotypes.Constant).Literal = lit
				}
			}

			// The statement
			//   const a int8 = 2
			// is equivalent to
			//  const a = int8(2)

			var typeDefIdent *gotypes.Identifier
			if typeDef.GetType() == gotypes.IdentifierType {
				typeDefIdent = typeDef.(*gotypes.Identifier)
			} else {
				sel := typeDef.(*gotypes.Selector)
				typeDefIdent = &gotypes.Identifier{
					Package: sel.Prefix.(*gotypes.Packagequalifier).Path,
					Def:     sel.Item,
				}
			}

			sDef.Def = &gotypes.Constant{
				Def:     typeDefIdent.Def,
				Package: typeDefIdent.Package,
				Literal: valueExprAttr.DataTypeList[0].(*gotypes.Constant).Literal,
			}

			// TODO(jchaloup): most likely one should generate contracts.IsTypecastableTo as well
			sp.Config.ContractTable.AddContract(&contracts.TypecastsTo{
				X:            valueExprAttr.TypeVarList[0],
				Type:         typevars.MakeConstant(typeDefIdent.Package, typeDefIdent),
				Y:            typevars.VariableFromSymbolDef(sDef),
				ExpectedType: sDef.Def,
			})
		}
		symbolsDef = append(symbolsDef, sDef)
	}

	return symbolsDef, nil
}

func (sp *Parser) ParseValueSpec(spec *ast.ValueSpec) ([]*symbols.SymbolDef, error) {
	var names []string
	for _, name := range spec.Names {
		names = append(names, name.Name)
	}
	glog.V(2).Infof("\n\nProcessing value spec: %#v\n\tNames: %#v\n", spec, strings.Join(names, ","))

	var typeDef gotypes.DataType
	if spec.Type != nil {
		def, err := sp.TypeParser.Parse(spec.Type)
		if err != nil {
			return nil, err
		}
		typeDef = def
	}

	var symbolsDef = make([]*symbols.SymbolDef, 0)

	// var a, b, _, ... TYPE
	if len(spec.Values) == 0 {
		if spec.Type == nil {
			return nil, fmt.Errorf("ValueSpec %#v has no type, thus it needs at least one value on the RHS", spec)
		}
		for i, name := range spec.Names {
			glog.V(2).Infof("Name[%v]: %v", i, name.String())
			if name.Name == "_" {
				continue
			}
			sDef := &symbols.SymbolDef{
				Name:    name.String(),
				Package: sp.PackageName,
				Def:     typeDef,
				Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, name.Pos()),
			}
			symbolsDef = append(symbolsDef, sDef)
			if sp.SymbolTable.CurrentLevel() > 0 {
				sDef.Package = ""
			}
			sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
				X:            typevars.MakeConstant(sp.PackageName, typeDef),
				Y:            typevars.VariableFromSymbolDef(sDef),
				ExpectedType: typeDef,
				Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, name.Pos()),
			})
		}
		return symbolsDef, nil
	}

	var lhs []ast.Expr
	for _, name := range spec.Names {
		lhs = append(lhs, name)
	}

	return sp.parseAssignments(&ast.AssignStmt{
		Lhs:    lhs,
		Rhs:    spec.Values,
		TokPos: spec.Pos(),
		Tok:    token.VAR,
	}, typeDef)
}

func (sp *Parser) parseDeclStmt(statement *ast.DeclStmt) error {
	switch decl := statement.Decl.(type) {
	case *ast.GenDecl:
		switch decl.Tok {
		case token.TYPE:
			for _, spec := range decl.Specs {
				genDeclSpec := spec.(*ast.TypeSpec)
				glog.V(2).Infof("Processing type spec declaration %#v\n", genDeclSpec)
				if err := sp.SymbolTable.AddDataType(&symbols.SymbolDef{
					Name: genDeclSpec.Name.Name,
					// local variable has package origin as well (though the symbol gets dropped later on)
					Package: sp.PackageName,
					Def:     nil,
					Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, spec.Pos()),
				}); err != nil {
					return err
				}

				// TODO(jchaloup): capture the current state of the allocated symbol table
				// JIC the parsing ends with end error. Which can result into re-parsing later on.
				// Which can result in re-allocation. It should be enough two-level allocated symbol table.
				typeDef, err := sp.TypeParser.Parse(genDeclSpec.Type)
				if err != nil {
					return err
				}

				if err := sp.SymbolTable.AddDataType(&symbols.SymbolDef{
					Name:    genDeclSpec.Name.Name,
					Package: sp.PackageName,
					Def:     typeDef,
					Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, spec.Pos()),
				}); err != nil {
					return err
				}

				// In case the data type is not defined on the global level,
				// it needs to be duplicated and stored there,
				// so the contract runner can still access it
				if sp.SymbolTable.CurrentLevel() > 0 {
					if err := sp.SymbolTable.AddVirtualDataType(&symbols.SymbolDef{
						Name:    genDeclSpec.Name.Name,
						Package: sp.PackageName,
						Def:     typeDef,
						Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, spec.Pos()),
					}); err != nil {
						return err
					}
				}
			}
		case token.VAR:
			for _, spec := range decl.Specs {
				genDeclSpec := spec.(*ast.ValueSpec)
				glog.V(2).Infof("Processing value spec declaration %#v\n", genDeclSpec)
				defs, err := sp.ParseValueSpec(genDeclSpec)
				if err != nil {
					return err
				}
				for _, def := range defs {
					// TPDP(jchaloup): we should store all variables or non.
					// Given the error is set only if the variable already exists, it should not matter so much.
					if err := sp.SymbolTable.AddVariable(def); err != nil {
						return nil
					}
				}
			}
		case token.CONST:
			var lastConstSpecValue []ast.Expr
			var lastConstSpecType ast.Expr
			for i, spec := range decl.Specs {
				constSpec := spec.(*ast.ValueSpec)
				// Is the value empty? Copy the previous expression
				if constSpec.Values == nil {
					if lastConstSpecValue == nil {
						return fmt.Errorf("Unable to re-costruct a value for const %#v", constSpec)
					}
					constSpec.Values = lastConstSpecValue
				} else {
					lastConstSpecType = nil
				}

				if constSpec.Type == nil && lastConstSpecType != nil {
					constSpec.Type = lastConstSpecType
				}

				for j, constName := range constSpec.Names {
					cSpec := &ast.ValueSpec{
						Doc:     constSpec.Doc,
						Names:   []*ast.Ident{constName},
						Type:    constSpec.Type,
						Values:  nil,
						Comment: constSpec.Comment,
					}

					if constSpec.Values != nil && j < len(constSpec.Values) {
						cSpec.Values = []ast.Expr{constSpec.Values[j]}
					}

					defs, err := sp.ParseConstValueSpec(types.ConstSpec{
						IotaIdx: uint(i),
						Spec:    cSpec,
					})
					if err != nil {
						return err
					}
					for _, def := range defs {
						// TPDP(jchaloup): we should store all variables or non.
						// Given the error is set only if the variable already exists, it should not matter so much.
						if err := sp.SymbolTable.AddVariable(def); err != nil {
							return err
						}
					}
				}

				lastConstSpecValue = constSpec.Values
				// once there is a type it applies to all following untyped constants until there is a new one
				if constSpec.Type != nil {
					lastConstSpecType = constSpec.Type
				}
			}
		case token.IMPORT:
			// applies only on the global scope
		default:
			panic(fmt.Sprintf("Unexpected ast.GenDecl %#v", decl))
		}
	default:
		panic(fmt.Errorf("Unrecognized declaration: %#v", statement.Decl))
	}
	return nil
}

func (sp *Parser) parseLabeledStmt(statement *ast.LabeledStmt) error {
	glog.V(2).Infof("Processing labeled %q statement %#v\n", statement.Label.Name, statement)
	// the label is typeless
	return sp.Parse(statement.Stmt)
}

func (sp *Parser) parseExprStmt(statement *ast.ExprStmt) error {
	glog.V(2).Infof("Processing expression statement  %#v\n", statement)
	_, err := sp.ExprParser.Parse(statement.X)
	return err
}

func (sp *Parser) parseSendStmt(statement *ast.SendStmt) error {
	glog.V(2).Infof("Processing statement statement  %#v\n", statement)
	chanAttr, err := sp.ExprParser.Parse(statement.Chan)
	if err != nil {
		return err
	}
	// TODO(jchaloup): should check the statement.Chan type is really a channel.
	valueAttr, err := sp.ExprParser.Parse(statement.Value)
	if err != nil {
		return err
	}

	sp.Config.ContractTable.AddContract(&contracts.IsSendableTo{
		X: valueAttr.TypeVarList[0],
		Y: chanAttr.TypeVarList[0],
	})

	return nil
}

func (sp *Parser) parseIncDecStmt(statement *ast.IncDecStmt) error {
	glog.V(2).Infof("Processing inc dec statement  %#v\n", statement)
	// both --,++ has no type information
	// TODO(jchaloup): check the --/++ can be carried over the statement.X
	attr, err := sp.ExprParser.Parse(statement.X)
	if err != nil {
		return err
	}

	sp.Config.ContractTable.AddContract(&contracts.IsIncDecable{
		X: attr.TypeVarList[0],
	})

	return nil
}

func (sp *Parser) parseAssignments(statement *ast.AssignStmt, typeDef gotypes.DataType) ([]*symbols.SymbolDef, error) {
	glog.V(2).Infof("Processing assignment statement  %#v\n", statement)

	// define general Rhs index function
	rhsIndexer := func(i int) (gotypes.DataType, typevars.Interface, error) {
		glog.V(2).Infof("Calling general Rhs index function at pos %v", statement.Pos())
		defAttr, err := sp.ExprParser.Parse(statement.Rhs[i])
		if err != nil {
			return nil, nil, fmt.Errorf("Error when parsing Rhs[%v] expression of %#v: %v at %v", i, statement, err, statement.Pos())
		}
		if len(defAttr.DataTypeList) != 1 {
			return nil, nil, fmt.Errorf("Assignment element at pos %v does not return a single result: %#v", i, defAttr.DataTypeList)
		}
		return defAttr.DataTypeList[0], defAttr.TypeVarList[0], err
	}

	exprsSize := len(statement.Lhs)
	rExprSize := len(statement.Rhs)
	// Some assignments are of a different number of expression on both sides.
	// E.g. value, ok := somemap[key]
	//      ret1, ..., retn = func(...)
	//		  expr, ok = expr.(type)
	if exprsSize != rExprSize {
		if rExprSize != 1 {
			return nil, fmt.Errorf("Number of expressions on the left-hand side differs from ones on the right-hand side for: %#v vs. %#v", statement.Lhs, statement.Rhs)
		}
		// function call or key reading?
		switch typeExpr := statement.Rhs[0].(type) {
		case *ast.CallExpr:
			var isCgo bool
			if sel, isSel := typeExpr.Fun.(*ast.SelectorExpr); isSel {
				if ident, isIdent := sel.X.(*ast.Ident); isIdent {
					if ident.String() == "C" {
						isCgo = true
					}
				}
			}

			callExprDefAttr, err := sp.ExprParser.Parse(typeExpr)
			if err != nil {
				return nil, err
			}
			callExprDefLen := len(callExprDefAttr.DataTypeList)

			if exprsSize != callExprDefLen {
				if isCgo {
					if exprsSize != callExprDefLen+1 {
						return nil, fmt.Errorf("CGO: Number of expressions on the left-hand side differs from number of results of invocation on the right-hand side for: %#v vs. %#v at %v", statement.Lhs, callExprDefAttr.DataTypeList, statement.Pos())
					}
				} else {
					return nil, fmt.Errorf("Number of expressions on the left-hand side differs from number of results of invocation on the right-hand side for: %#v vs. %#v", statement.Lhs, callExprDefAttr.DataTypeList)
				}
			}
			rhsIndexer = func(i int) (gotypes.DataType, typevars.Interface, error) {
				glog.V(2).Infof("Calling CallExpr Rhs index function")
				if i < callExprDefLen {
					return callExprDefAttr.DataTypeList[i], callExprDefAttr.TypeVarList[i], nil
				}
				if i == callExprDefLen && isCgo {
					// TODO(jchaloup) return Identifier instead of the Builtin
					return &gotypes.Identifier{Package: "builtin", Def: "error"}, &typevars.CGO{Package: "builtin", DataType: &gotypes.Identifier{Package: "builtin", Def: "error"}}, nil
				}
				panic("rhsIndexer: out of range")
			}
			rExprSize = len(callExprDefAttr.DataTypeList)
		case *ast.IndexExpr:
			if exprsSize != 2 {
				return nil, fmt.Errorf("Expecting two expression on the RHS when accessing an index expression for val, ok = indexexpr[key], got %#v instead", statement.Lhs)
			}

			rhsIndexer = func(i int) (gotypes.DataType, typevars.Interface, error) {
				glog.V(2).Infof("Calling IndexExpr Rhs index function, i = %v", i)
				if i == 0 {
					defAttr, err := sp.ExprParser.Parse(statement.Rhs[i])
					if err != nil {
						return nil, nil, fmt.Errorf("Error when parsing Rhs[%v] expression of %#v: %v at %v", i, statement, err, statement.Pos())
					}
					{
						byteSlice, _ := json.Marshal(defAttr.DataTypeList[0])
						glog.V(2).Infof("\n\tHHHH: %v\n", string(byteSlice))
					}
					if len(defAttr.DataTypeList) != 1 {
						return nil, nil, fmt.Errorf("Assignment element at pos %v does not return a single result: %#v", i, defAttr.DataTypeList)
					}

					return defAttr.DataTypeList[0], defAttr.TypeVarList[0], err
				}
				if i == 1 {
					return &gotypes.Identifier{Package: "builtin", Def: "bool"}, typevars.MakeConstant(sp.Config.PackageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}), nil
				}
				return nil, nil, fmt.Errorf("Rhs index %v out of range", i)
			}
			rExprSize = 2
		case *ast.TypeAssertExpr:
			if exprsSize != 2 {
				return nil, fmt.Errorf("Expecting two expression on the RHS when type-asserting an expression for val, ok = expr.(datatype), got %#v instead", statement.Lhs)
			}
			xDefAttr, err := sp.ExprParser.Parse(typeExpr.X)
			if err != nil {
				return nil, err
			}
			if len(xDefAttr.DataTypeList) != 1 {
				return nil, fmt.Errorf("Index expression is not a single expression: %#v", xDefAttr.DataTypeList)
			}
			typeDef, err := sp.TypeParser.Parse(typeExpr.Type)
			if err != nil {
				return nil, err
			}
			rhsIndexer = func(i int) (gotypes.DataType, typevars.Interface, error) {
				glog.V(2).Infof("Calling TypeAssertExpr Rhs index function")
				if i == 0 {
					y := typevars.MakeConstant(sp.Config.PackageName, typeDef)
					sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
						X:            xDefAttr.TypeVarList[0],
						Y:            y,
						ExpectedType: typeDef,
					})
					return typeDef, y, nil
				}
				if i == 1 {
					return &gotypes.Identifier{Package: "builtin", Def: "bool"}, typevars.MakeConstant(sp.Config.PackageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}), nil
				}
				return nil, nil, fmt.Errorf("Rhs index %v out of range", i)
			}
			rExprSize = 2
		default:
			panic(fmt.Errorf("Expecting *ast.CallExpr, *ast.IndexExpr or *ast.TypeAssertExpr, got %#v instead", statement.Rhs[0]))
		}
	}

	var symbolsDef = make([]*symbols.SymbolDef, 0)
	// If the assignment token is token.DEFINE a variable gets stored into the symbol table.
	// If it is already there and has the same type, do not do anything. Error if the type is different.
	// If it is not there yet, add it into the table
	// If the token is token.ASSIGN the variable must be in the symbol table.
	// If it is and has the same type, do not do anything, Error, if the type is different.
	// If it is not there yet, error.
	for i := 0; i < exprsSize; i++ {
		// If the left-hand side id a selector (e.g. struct.field), we alredy know data type of the id.
		// So, just store the field's data type into the allocated symbol table
		//switch lhsExpr := expr.
		rhsExpr, rhsTypeVar, err := rhsIndexer(i)
		if err != nil {
			return nil, err
		}
		glog.V(2).Infof("Assignment LHs[%v]: %#v\n", i, statement.Lhs[i])
		glog.V(2).Infof("Assignment RHs[%v]: %#v\n", i, rhsExpr)

		// Skip parenthesis
		parExpr, ok := statement.Lhs[i].(*ast.ParenExpr)
		for ok {
			statement.Lhs[i] = parExpr.X
			parExpr, ok = statement.Lhs[i].(*ast.ParenExpr)
		}

		switch lhsExpr := statement.Lhs[i].(type) {
		case *ast.Ident:
			// skip the anonymous variables
			if lhsExpr.Name == "_" {
				continue
			}

			// TODO(jchaloup): If the statement.Tok is not token.DEFINE, don't add the variable to the symbol table.
			//                 Instead, check the varible is of the same type (or compatible) as the already stored one.

			glog.V(2).Infof("Adding contract for token %v", statement.Tok)
			// :=
			switch statement.Tok {
			case token.DEFINE:
				sDef := &symbols.SymbolDef{
					Name:    lhsExpr.Name,
					Package: sp.PackageName,
					Def:     rhsExpr,
					Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
				}
				if sp.SymbolTable.CurrentLevel() > 0 {
					sDef.Package = ""
				}
				symbolsDef = append(symbolsDef, sDef)
				sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
					X:            rhsTypeVar,
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: sDef.Def,
					ToVariable:   true,
					Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
				})
			case token.VAR:

				var sDef *symbols.SymbolDef
				if typeDef == nil {
					// constant changes to identifier
					if rhsExpr.GetType() == gotypes.ConstantType {
						c := rhsExpr.(*gotypes.Constant)
						rhsExpr = &gotypes.Identifier{
							Package: c.Package,
							Def:     c.Def,
						}
					} else if rhsExpr.GetType() == gotypes.BuiltinType {
						b := rhsExpr.(*gotypes.Builtin)
						rhsExpr = &gotypes.Identifier{
							Package: "builtin",
							Def:     b.Def,
						}
					}
					sDef = &symbols.SymbolDef{
						Name:    lhsExpr.Name,
						Package: sp.PackageName,
						Def:     rhsExpr,
						Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
					}
					if sp.SymbolTable.CurrentLevel() > 0 {
						sDef.Package = ""
					}
					sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
						X:            rhsTypeVar,
						Y:            typevars.VariableFromSymbolDef(sDef),
						ExpectedType: sDef.Def,
						ToVariable:   true,
						Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
					})
				} else {
					sDef = &symbols.SymbolDef{
						Name:    lhsExpr.Name,
						Package: sp.PackageName,
						Def:     typeDef,
						Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
					}
					if sp.SymbolTable.CurrentLevel() > 0 {
						sDef.Package = ""
					}
					sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
						X:            rhsTypeVar,
						Y:            typevars.VariableFromSymbolDef(sDef),
						ExpectedType: sDef.Def,
					})
					sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
						// TODO(jchaloup): determine the correct package of the typeDef
						X:            typevars.MakeConstant(sp.Config.PackageName, typeDef),
						Y:            typevars.VariableFromSymbolDef(sDef),
						ExpectedType: sDef.Def,
						ToVariable:   true,
						Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
					})
				}
				symbolsDef = append(symbolsDef, sDef)
			default:
				sDef, err := sp.SymbolTable.LookupVariable(lhsExpr.Name)
				if err != nil {
					return nil, err
				}
				sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
					X:            rhsTypeVar,
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: sDef.Def,
				})
			}
		case *ast.SelectorExpr, *ast.StarExpr:
			attr, err := sp.ExprParser.Parse(statement.Lhs[i])
			if err != nil {
				return nil, err
			}
			sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X:            rhsTypeVar,
				Y:            attr.TypeVarList[0],
				ExpectedType: attr.DataTypeList[0],
			})
		case *ast.IndexExpr:
			// This goes straight to the indexExpr parsing method
			attr, err := sp.ExprParser.Parse(lhsExpr)
			if err != nil {
				return nil, err
			}
			sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X:            rhsTypeVar,
				Y:            attr.TypeVarList[0],
				ExpectedType: attr.DataTypeList[0],
			})
		default:
			return nil, fmt.Errorf("Lhs[%v] of an assignment type %#v is not recognized", i, statement.Lhs[i])
		}
	}
	return symbolsDef, nil
}

func (sp *Parser) parseAssignStmt(statement *ast.AssignStmt) error {
	glog.V(2).Infof("Processing assignment statement  %#v\n", statement)

	// define general Rhs index function
	rhsIndexer := func(i int) (gotypes.DataType, typevars.Interface, error) {
		glog.V(2).Infof("Calling general Rhs index function at pos %v", statement.Pos())
		defAttr, err := sp.ExprParser.Parse(statement.Rhs[i])
		if err != nil {
			return nil, nil, fmt.Errorf("Error when parsing Rhs[%v] expression of %#v: %v at %v", i, statement, err, statement.Pos())
		}
		if len(defAttr.DataTypeList) != 1 {
			return nil, nil, fmt.Errorf("Assignment element at pos %v does not return a single result: %#v", i, defAttr.DataTypeList)
		}
		return defAttr.DataTypeList[0], defAttr.TypeVarList[0], err
	}

	exprsSize := len(statement.Lhs)
	rExprSize := len(statement.Rhs)
	// Some assignments are of a different number of expression on both sides.
	// E.g. value, ok := somemap[key]
	//      ret1, ..., retn = func(...)
	//		  expr, ok = expr.(type)
	if exprsSize != rExprSize {
		if rExprSize != 1 {
			return fmt.Errorf("Number of expressions on the left-hand side differs from ones on the right-hand side for: %#v vs. %#v", statement.Lhs, statement.Rhs)
		}
		// function call or key reading?
		switch typeExpr := skipPars(statement.Rhs[0]).(type) {
		case *ast.CallExpr:
			var isCgo bool
			if sel, isSel := typeExpr.Fun.(*ast.SelectorExpr); isSel {
				if ident, isIdent := sel.X.(*ast.Ident); isIdent {
					if ident.String() == "C" {
						isCgo = true
					}
				}
			}

			callExprDefAttr, err := sp.ExprParser.Parse(typeExpr)
			if err != nil {
				return err
			}
			callExprDefLen := len(callExprDefAttr.DataTypeList)

			if exprsSize != callExprDefLen {
				if isCgo {
					if exprsSize != callExprDefLen+1 {
						return fmt.Errorf("CGO: Number of expressions on the left-hand side differs from number of results of invocation on the right-hand side for: %#v vs. %#v at %v", statement.Lhs, callExprDefAttr.DataTypeList, statement.Pos())
					}
				} else {
					return fmt.Errorf("Number of expressions on the left-hand side differs from number of results of invocation on the right-hand side for: %#v vs. %#v", statement.Lhs, callExprDefAttr.DataTypeList)
				}
			}
			rhsIndexer = func(i int) (gotypes.DataType, typevars.Interface, error) {
				glog.V(2).Infof("Calling CallExpr Rhs index function")
				if i < callExprDefLen {
					return callExprDefAttr.DataTypeList[i], callExprDefAttr.TypeVarList[i], nil
				}
				if i == callExprDefLen && isCgo {
					// TODO(jchaloup) return Identifier instead of the Builtin
					return &gotypes.Identifier{Package: "builtin", Def: "error"}, &typevars.CGO{Package: "builtin", DataType: &gotypes.Identifier{Package: "builtin", Def: "error"}}, nil
				}
				panic("rhsIndexer: out of range")
			}
			rExprSize = len(callExprDefAttr.DataTypeList)
		case *ast.IndexExpr:
			if exprsSize != 2 {
				return fmt.Errorf("Expecting two expression on the RHS when accessing an index expression for val, ok = indexexpr[key], got %#v instead", statement.Lhs)
			}

			rhsIndexer = func(i int) (gotypes.DataType, typevars.Interface, error) {
				glog.V(2).Infof("Calling IndexExpr Rhs index function, i = %v", i)
				if i == 0 {
					defAttr, err := sp.ExprParser.Parse(statement.Rhs[i])
					if err != nil {
						return nil, nil, fmt.Errorf("Error when parsing Rhs[%v] expression of %#v: %v at %v", i, statement, err, statement.Pos())
					}
					{
						byteSlice, _ := json.Marshal(defAttr.DataTypeList[0])
						glog.V(2).Infof("\n\tHHHH: %v\n", string(byteSlice))
					}
					if len(defAttr.DataTypeList) != 1 {
						return nil, nil, fmt.Errorf("Assignment element at pos %v does not return a single result: %#v", i, defAttr.DataTypeList)
					}
					// TODO(jchaloup): generate ListValue or MapValue contract
					return defAttr.DataTypeList[0], defAttr.TypeVarList[0], err
				}
				if i == 1 {
					return &gotypes.Identifier{Package: "builtin", Def: "bool"}, typevars.MakeConstant(sp.Config.PackageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}), nil
				}
				return nil, nil, fmt.Errorf("Rhs index %v out of range", i)
			}
			rExprSize = 2
		case *ast.TypeAssertExpr:
			if exprsSize != 2 {
				return fmt.Errorf("Expecting two expression on the RHS when type-asserting an expression for val, ok = expr.(datatype), got %#v instead", statement.Lhs)
			}
			xDefAttr, err := sp.ExprParser.Parse(typeExpr.X)
			if err != nil {
				return err
			}
			if len(xDefAttr.DataTypeList) != 1 {
				return fmt.Errorf("Index expression is not a single expression: %#v", xDefAttr.DataTypeList)
			}
			typeDef, err := sp.TypeParser.Parse(typeExpr.Type)
			if err != nil {
				return err
			}
			rhsIndexer = func(i int) (gotypes.DataType, typevars.Interface, error) {
				glog.V(2).Infof("Calling TypeAssertExpr Rhs index function")
				if i == 0 {
					y := typevars.MakeConstant(sp.Config.PackageName, typeDef)
					sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
						X:            xDefAttr.TypeVarList[0],
						Y:            y,
						ExpectedType: typeDef,
					})
					return typeDef, y, nil
				}
				if i == 1 {
					return &gotypes.Identifier{Package: "builtin", Def: "bool"}, typevars.MakeConstant(sp.Config.PackageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}), nil
				}
				return nil, nil, fmt.Errorf("Rhs index %v out of range", i)
			}
			rExprSize = 2
		case *ast.UnaryExpr:
			// val, ok <- channel
			if typeExpr.Op != token.ARROW {
				return fmt.Errorf("Expected <- to a channel, got %#v instead", typeExpr)
			}
			if exprsSize != 2 {
				return fmt.Errorf("Expecting two expression on the RHS when receiving val, ok = <-var, got %#v instead", statement.Lhs)
			}
			xDefAttr, err := sp.ExprParser.Parse(typeExpr.X)
			if err != nil {
				return err
			}
			if len(xDefAttr.DataTypeList) != 1 {
				return fmt.Errorf("Index expression is not a single expression: %#v", xDefAttr.DataTypeList)
			}
			// in case it is an identifier
			nonIdent, err := sp.SymbolsAccessor.FindFirstNonidDataType(xDefAttr.DataTypeList[0])
			if err != nil {
				return err
			}

			channel, ok := nonIdent.(*gotypes.Channel)
			if !ok {
				return fmt.Errorf("Expected <- to a channel, got %#v instead", nonIdent)
			}
			if channel.Dir == "1" {
				return fmt.Errorf("Expected a channel %#v to be at least receiveable", channel)
			}
			rhsIndexer = func(i int) (gotypes.DataType, typevars.Interface, error) {
				glog.V(2).Infof("Calling <-channel Rhs index function")
				if i == 0 {
					y := sp.Config.ContractTable.NewVirtualVar()
					sp.Config.ContractTable.AddContract(&contracts.IsReceiveableFrom{
						X: xDefAttr.TypeVarList[0],
						Y: y,
					})

					return channel.Value, y, nil
				}
				if i == 1 {
					return &gotypes.Identifier{Package: "builtin", Def: "bool"}, typevars.MakeConstant(sp.Config.PackageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}), nil
				}
				return nil, nil, fmt.Errorf("Rhs index %v out of range", i)
			}
			rExprSize = 2
		default:
			panic(fmt.Errorf("Expecting *ast.CallExpr, *ast.IndexExpr or *ast.TypeAssertExpr, got %#v instead", statement.Rhs[0]))
		}
	}

	// If the assignment token is token.DEFINE a variable gets stored into the symbol table.
	// If it is already there and has the same type, do not do anything. Error if the type is different.
	// If it is not there yet, add it into the table
	// If the token is token.ASSIGN the variable must be in the symbol table.
	// If it is and has the same type, do not do anything, Error, if the type is different.
	// If it is not there yet, error.
	var sDefs []*symbols.SymbolDef

	for i := 0; i < exprsSize; i++ {
		// If the left-hand side id a selector (e.g. struct.field), we alredy know data type of the id.
		// So, just store the field's data type into the allocated symbol table
		//switch lhsExpr := expr.
		rhsExpr, rhsTypeVar, err := rhsIndexer(i)
		if err != nil {
			return err
		}
		glog.V(2).Infof("Assignment LHs[%v]: %#v\n", i, statement.Lhs[i])
		glog.V(2).Infof("Assignment RHs[%v]: %#v\n", i, rhsExpr)

		// Skip parenthesis
		parExpr, ok := statement.Lhs[i].(*ast.ParenExpr)
		for ok {
			statement.Lhs[i] = parExpr.X
			parExpr, ok = statement.Lhs[i].(*ast.ParenExpr)
		}

		switch lhsExpr := statement.Lhs[i].(type) {
		case *ast.Ident:
			// skip the anonymous variables
			if lhsExpr.Name == "_" {
				continue
			}

			// TODO(jchaloup): If the statement.Tok is not token.DEFINE, don't add the variable to the symbol table.
			//                 Instead, check the varible is of the same type (or compatible) as the already stored one.

			glog.V(2).Infof("Adding contract for token %v", statement.Tok)
			// :=
			if statement.Tok == token.DEFINE {
				// constant changes to identifier
				if rhsExpr.GetType() == gotypes.ConstantType {
					c := rhsExpr.(*gotypes.Constant)
					rhsExpr = &gotypes.Identifier{
						Package: c.Package,
						Def:     c.Def,
					}
				} else if rhsExpr.GetType() == gotypes.BuiltinType {
					b := rhsExpr.(*gotypes.Builtin)
					rhsExpr = &gotypes.Identifier{
						Package: "builtin",
						Def:     b.Def,
					}
				}
				sDef := &symbols.SymbolDef{
					Name:    lhsExpr.Name,
					Package: sp.PackageName,
					Def:     rhsExpr,
					Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
				}
				if sp.SymbolTable.CurrentLevel() > 0 {
					sDef.Package = ""
				}
				sDefs = append(sDefs, sDef)
				sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
					X:            rhsTypeVar,
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: sDef.Def,
					ToVariable:   true,
					Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Lhs[i].Pos()),
				})
			} else {
				sDef, err := sp.SymbolTable.LookupVariable(lhsExpr.Name)
				if err != nil {
					return err
				}
				sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
					X:            rhsTypeVar,
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: sDef.Def,
				})
			}
		case *ast.SelectorExpr, *ast.StarExpr:
			attr, err := sp.ExprParser.Parse(statement.Lhs[i])
			if err != nil {
				return err
			}
			sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X:            rhsTypeVar,
				Y:            attr.TypeVarList[0],
				ExpectedType: attr.DataTypeList[0],
			})
		case *ast.IndexExpr:
			// This goes straight to the indexExpr parsing method
			attr, err := sp.ExprParser.Parse(lhsExpr)
			if err != nil {
				return err
			}
			sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X:            rhsTypeVar,
				Y:            attr.TypeVarList[0],
				ExpectedType: attr.DataTypeList[0],
			})
		default:
			return fmt.Errorf("Lhs[%v] of an assignment type %#v is not recognized", i, statement.Lhs[i])
		}
	}
	for _, sDef := range sDefs {
		sp.SymbolTable.AddVariable(sDef)
	}
	return nil
}

func (sp *Parser) parseGoStmt(statement *ast.GoStmt) error {
	glog.V(2).Infof("Processing Go statement  %#v\n", statement)
	_, err := sp.ExprParser.Parse(statement.Call)
	return err
}

func (sp *Parser) parseBranchStmt(statement *ast.BranchStmt) error {
	// TODO(jchaloup): just panic here!!!
	glog.V(2).Infof("Processing branch statement  %#v\n", statement)
	return nil
}

func (sp *Parser) parseSwitchStmt(statement *ast.SwitchStmt) error {
	glog.V(2).Infof("Processing switch statement  %#v\n", statement)
	// ExprSwitchStmt = "switch" [ SimpleStmt ";" ] [ Expression ] "{" { ExprCaseClause } "}" .
	// ExprCaseClause = ExprSwitchCase ":" StatementList .
	// ExprSwitchCase = "case" ExpressionList | "default" .
	if statement.Init != nil {
		sp.SymbolTable.Push()
		defer sp.SymbolTable.Pop()

		if err := sp.StmtParser.Parse(statement.Init); err != nil {
			return err
		}
	}

	var stmtTagTypeVar typevars.Interface
	if statement.Tag == nil {
		stmtTagTypeVar = typevars.MakeConstant("builtin", &gotypes.Identifier{Package: "builtin", Def: "bool"})
	} else {
		attr, err := sp.ExprParser.Parse(statement.Tag)
		if err != nil {
			return err
		}
		stmtTagTypeVar = attr.TypeVarList[0]
	}

	for _, stmt := range statement.Body.List {
		caseStmt, ok := stmt.(*ast.CaseClause)
		if !ok {
			return fmt.Errorf("Expected *ast.CaseClause in switch's body. Got %#v\n", stmt)
		}
		for _, caseExpr := range caseStmt.List {
			caseAttr, err := sp.ExprParser.Parse(caseExpr)
			if err != nil {
				return nil
			}
			sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X: caseAttr.TypeVarList[0],
				Y: stmtTagTypeVar,
			})
		}

		if caseStmt.Body != nil {
			if err := sp.Parse(&ast.BlockStmt{
				List: caseStmt.Body,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (sp *Parser) parseTypeSwitchStmt(statement *ast.TypeSwitchStmt) error {
	glog.V(2).Infof("Processing type switch statement  %#v\n", statement)
	// TypeSwitchStmt  = "switch" [ SimpleStmt ";" ] TypeSwitchGuard "{" { TypeCaseClause } "}" .
	// TypeSwitchGuard = [ identifier ":=" ] PrimaryExpr "." "(" "type" ")" .
	// TypeCaseClause  = TypeSwitchCase ":" StatementList .
	// TypeSwitchCase  = "case" TypeList | "default" .
	// TypeList        = Type { "," Type } .
	if statement.Init != nil {
		sp.SymbolTable.Push()
		defer sp.SymbolTable.Pop()

		if err := sp.StmtParser.Parse(statement.Init); err != nil {
			return err
		}
	}

	sp.SymbolTable.Push()
	defer sp.SymbolTable.Pop()

	var rhsIdentifier *gotypes.Identifier
	var typeTypeVar typevars.Interface
	var typeDef gotypes.DataType

	switch assignStmt := statement.Assign.(type) {
	case *ast.AssignStmt:
		sp.SymbolTable.Push()
		defer sp.SymbolTable.Pop()

		if len(assignStmt.Lhs) != 1 {
			return fmt.Errorf("LHS of the type assert assignment must be a 'single' assignable expression")
		}

		if len(assignStmt.Rhs) != 1 {
			return fmt.Errorf("RHS of the type assert assignment must be a 'single' expression")
		}

		typeAssert, ok := assignStmt.Rhs[0].(*ast.TypeAssertExpr)
		if !ok {
			return fmt.Errorf("Expecting type assert in type switch assignemnt statement")
		}
		if typeAssert.Type != nil {
			glog.Warningf("type assert type is not set to literal 'type'. It is %#v instead\n", typeAssert.Type)
		}
		typeAttr, err := sp.ExprParser.Parse(typeAssert.X)
		if err != nil {
			return err
		}
		typeTypeVar = typeAttr.TypeVarList[0]
		typeDef = typeAttr.DataTypeList[0]

		ident, ok := assignStmt.Lhs[0].(*ast.Ident)
		if !ok {
			return fmt.Errorf("Expecting identifier on the RHS of the type switch assigment. Got %#v instead", assignStmt.Lhs[0])
		}

		if ident.Name != "_" {
			rhsIdentifier = &gotypes.Identifier{
				Def:     ident.Name,
				Package: sp.PackageName,
			}
		}
	case *ast.ExprStmt:
		typeAssert, ok := assignStmt.X.(*ast.TypeAssertExpr)
		if !ok {
			return fmt.Errorf("Expecting type assert in type switch")
		}
		if typeAssert.Type != nil {
			glog.Warningf("type assert type is not set to literal 'type'. It is %#v instead\n", typeAssert.Type)
		}
		typeAttr, err := sp.ExprParser.Parse(typeAssert.X)
		if err != nil {
			return nil
		}
		typeTypeVar = typeAttr.TypeVarList[0]
		typeDef = typeAttr.DataTypeList[0]
	default:
		return fmt.Errorf("Unsupported statement in type switch")
	}

	for _, stmt := range statement.Body.List {
		caseStmt, ok := stmt.(*ast.CaseClause)
		if !ok {
			return fmt.Errorf("Expected *ast.CaseClause in switch's body. Got %#v\n", stmt)
		}
		if err := func() error {
			sp.SymbolTable.Push()
			defer sp.SymbolTable.Pop()
			// in case there are at least two data types (or none = default case), the type is interface{}
			if len(caseStmt.List) != 1 {
				for _, caseExpr := range caseStmt.List {
					// case nil:
					if ident, ok := skipPars(caseExpr).(*ast.Ident); ok && ident.Name == "nil" {
						continue
					}

					rhsType, err := sp.TypeParser.Parse(caseExpr)
					if err != nil {
						return err
					}
					sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
						X:    typevars.MakeConstant(sp.Config.PackageName, rhsType),
						Y:    typeTypeVar,
						Weak: true,
					})
				}
				if rhsIdentifier != nil {
					sDef := &symbols.SymbolDef{
						Name:    rhsIdentifier.Def,
						Def:     typeDef,
						Package: sp.PackageName,
						// The Pos in this context is not applicable to the typevars in each
						// case block the rhsIdentifier.Def has different scope
						// TODO(jchaloup): find a different way to state the position of the rhsIdentifier
						Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, caseStmt.Pos()),
					}
					sp.SymbolTable.AddVariable(sDef)
					sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
						X:   typevars.MakeConstant(sp.Config.PackageName, typeDef),
						Y:   typevars.VariableFromSymbolDef(sDef),
						Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, caseStmt.Pos()),
					})
				}
			} else {
				// case nil:
				skipCase := false
				if ident, ok := skipPars(caseStmt.List[0]).(*ast.Ident); ok && ident.Name == "nil" {
					skipCase = true
				}

				if !skipCase {
					rhsType, err := sp.TypeParser.Parse(caseStmt.List[0])
					if err != nil {
						return err
					}
					sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
						X:    typevars.MakeConstant(sp.Config.PackageName, rhsType),
						Y:    typeTypeVar,
						Weak: true,
					})
					if rhsIdentifier != nil {
						sDef := &symbols.SymbolDef{
							Name:    rhsIdentifier.Def,
							Package: sp.PackageName,
							Def:     rhsType,
							// The Pos in this context is not applicable to the typevars in each
							// case block the rhsIdentifier.Def has different scope
							// TODO(jchaloup): find a different way to state the position of the rhsIdentifier
							Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, caseStmt.Pos()),
						}
						sp.SymbolTable.AddVariable(sDef)
						sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
							X:   typevars.MakeConstant(sp.Config.PackageName, rhsType),
							Y:   typevars.VariableFromSymbolDef(sDef),
							Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, caseStmt.Pos()),
						})
					}
				}
			}

			if caseStmt.Body != nil {
				if err := sp.Parse(&ast.BlockStmt{
					List: caseStmt.Body,
				}); err != nil {
					return err
				}
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

func skipPars(e ast.Expr) ast.Expr {
	p, ok := e.(*ast.ParenExpr)
	if ok {
		return skipPars(p.X)
	}
	return e
}

// SelectStmt = "select" "{" { CommClause } "}" .
// CommClause = CommCase ":" StatementList .
// CommCase   = "case" ( SendStmt | RecvStmt ) | "default" .
// RecvStmt   = [ ExpressionList "=" | IdentifierList ":=" ] RecvExpr .
// RecvExpr   = Expression .
func (sp *Parser) parseSelectStmt(statement *ast.SelectStmt) error {
	glog.V(2).Infof("Processing select statement  %#v\n", statement)
	// TODO(jchaloup): deal with select{}
	for _, stmt := range statement.Body.List {
		if err := func() error {
			sp.SymbolTable.Push()
			defer sp.SymbolTable.Pop()

			glog.V(2).Infof("Processing select statement clause %#v\n", stmt)

			commClause, ok := stmt.(*ast.CommClause)
			if !ok {
				return fmt.Errorf("Select must be a list of commClause statements")
			}

			// default has empty Comm
			if commClause.Comm != nil {
				glog.V(2).Infof("commClause.Comm: %#v", commClause.Comm)
				switch clause := commClause.Comm.(type) {
				case *ast.ExprStmt:
					if _, err := sp.ExprParser.Parse(clause.X); err != nil {
						return err
					}
				case *ast.SendStmt:
					sendChanAtrr, err := sp.ExprParser.Parse(clause.Chan)
					if err != nil {
						return err
					}
					sendValueAtrr, err := sp.ExprParser.Parse(clause.Value)
					if err != nil {
						return err
					}
					sp.Config.ContractTable.AddContract(&contracts.IsSendableTo{
						X: sendValueAtrr.TypeVarList[0],
						Y: sendChanAtrr.TypeVarList[0],
					})
				case *ast.AssignStmt:
					if len(clause.Rhs) != 1 {
						return fmt.Errorf("Expecting a single expression on the RHS of a clause assigment")
					}

					chExpr, ok := clause.Rhs[0].(*ast.UnaryExpr)
					if !ok {
						return fmt.Errorf("Expecting unary expression in a form of <-chan. Got clause.Rhs[0]=%#v instead", clause.Rhs[0])
					}

					if chExpr.Op != token.ARROW {
						return fmt.Errorf("Expecting unary expression in a form of <-chan. Got chExpr=%#v instead", chExpr)
					}

					rhsExprAttr, err := sp.ExprParser.Parse(chExpr.X)
					if err != nil {
						return err
					}

					if len(rhsExprAttr.DataTypeList) != 1 {
						return fmt.Errorf("Expecting unary expression in a form of <-chan. Got rhsExprAttr.DataTypeList=%#v instead", rhsExprAttr.DataTypeList)
					}

					channel, err := sp.Config.SymbolsAccessor.FindFirstNonIdSymbol(rhsExprAttr.DataTypeList[0])
					if err != nil {
						return err
					}

					rhsChannel, ok := channel.(*gotypes.Channel)
					if !ok {
						return fmt.Errorf("Expecting unary expression in a form of <-chan. Got rhsExprAttr.DataTypeList[0]=%#v instead", channel)
					}

					switch lhsLen := len(clause.Lhs); lhsLen {
					case 0:
						return fmt.Errorf("Expecting at least one expression on the LHS of a clause assigment")
					case 1:
						y := sp.Config.ContractTable.NewVirtualVar()
						sp.Config.ContractTable.AddContract(&contracts.IsReceiveableFrom{
							X: rhsExprAttr.TypeVarList[0],
							Y: y,
						})

						if clause.Tok == token.DEFINE {
							// All LHS expressions must be identifiers
							// TODO(jchaloup): the expression can be surrounded by parenthesis!!! Make sure they are checked as well
							ident, ok := clause.Lhs[0].(*ast.Ident)
							if !ok {
								return fmt.Errorf("Expecting an identifier in select clause due to := assignment")
							}
							sDef := &symbols.SymbolDef{
								Name:    ident.Name,
								Package: sp.PackageName,
								Def:     rhsChannel.Value,
								Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, ident.Pos()),
							}
							sp.SymbolTable.AddVariable(sDef)

							sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
								X:   y,
								Y:   typevars.VariableFromSymbolDef(sDef),
								Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, ident.Pos()),
							})
						} else {
							if err := func() error {
								nonP := skipPars(clause.Lhs[0])
								if ident, ok := nonP.(*ast.Ident); ok && ident.Name == "_" {
									return nil
								}
								assignAttr, err := sp.ExprParser.Parse(clause.Lhs[0])
								if err != nil {
									return err
								}
								sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
									X: y,
									Y: assignAttr.TypeVarList[0],
								})
								return nil
							}(); err != nil {
								return err
							}
						}
					case 2:
						// Given a case is an implicit block, there is no need to check if any of to-be-declared variables is already declared
						// See http://www.tapirgames.com/blog/golang-block-and-scope
						// If an ordinary assignment (with = symbol) is used, all variables/selectors(and other assignable expression) must be already defined,
						// so they are not included as new variables.
						y := sp.Config.ContractTable.NewVirtualVar()
						sp.Config.ContractTable.AddContract(&contracts.IsReceiveableFrom{
							X: rhsExprAttr.TypeVarList[0],
							Y: y,
						})
						if clause.Tok == token.DEFINE {
							// All LHS expressions must be identifiers
							// TODO(jchaloup): the expression can be surrounded by parenthesis!!! Make sure they are checked as well
							ident1, ok := clause.Lhs[0].(*ast.Ident)
							if !ok {
								return fmt.Errorf("Expecting an identifier in select clause due to := assignment")
							}
							ident2, ok := clause.Lhs[1].(*ast.Ident)
							if !ok {
								return fmt.Errorf("Expecting an identifier in select clause due to := assignment")
							}

							sDef1 := &symbols.SymbolDef{
								Name:    ident1.Name,
								Package: sp.PackageName,
								Def:     rhsChannel.Value,
								Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, ident1.Pos()),
							}
							sp.SymbolTable.AddVariable(sDef1)
							sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
								X:   y,
								Y:   typevars.VariableFromSymbolDef(sDef1),
								Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, ident1.Pos()),
							})

							sDef2 := &symbols.SymbolDef{
								Name:    ident2.Name,
								Package: sp.PackageName,
								Def:     &gotypes.Identifier{Package: "builtin", Def: "bool"},
								Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, ident2.Pos()),
							}
							sp.SymbolTable.AddVariable(sDef2)
							sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
								X:   typevars.MakeConstant("builtin", &gotypes.Identifier{Package: "builtin", Def: "bool"}),
								Y:   typevars.VariableFromSymbolDef(sDef2),
								Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, ident2.Pos()),
							})
						} else {
							// Both vars can be just _
							// a := make(chan int, 1)
							// a <- 2
							// select {
							// 	case _, ((_)) = <- a:
							// }
							if err := func() error {
								nonP := skipPars(clause.Lhs[0])
								if ident, ok := nonP.(*ast.Ident); ok && ident.Name == "_" {
									return nil
								}
								assignAttr, err := sp.ExprParser.Parse(clause.Lhs[0])
								if err != nil {
									return err
								}
								sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
									X: y,
									Y: assignAttr.TypeVarList[0],
								})
								return nil
							}(); err != nil {
								return err
							}

							if err := func() error {
								nonP := skipPars(clause.Lhs[1])
								if ident, ok := nonP.(*ast.Ident); ok && ident.Name == "_" {
									return nil
								}
								assignOkAttr, err := sp.ExprParser.Parse(clause.Lhs[1])
								if err != nil {
									return err
								}
								sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
									X: typevars.MakeConstant("builtin", &gotypes.Identifier{Package: "builtin", Def: "bool"}),
									Y: assignOkAttr.TypeVarList[0],
								})
								return nil
							}(); err != nil {
								return err
							}

						}
					default:
						return fmt.Errorf("Expecting at most two expression on the LHS of a clause assigment")
					}
				default:
					return fmt.Errorf("Unable to recognize selector CommClause CommCase. Got %#v\n", commClause.Comm)
				}
			}

			if commClause.Body != nil {
				if err := sp.Parse(&ast.BlockStmt{
					List: commClause.Body,
				}); err != nil {
					return err
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}

// ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
// Condition = Expression .
func (sp *Parser) parseForStmt(statement *ast.ForStmt) error {
	glog.V(2).Infof("Processing for statement  %#v\n", statement)
	sp.SymbolTable.Push()
	defer sp.SymbolTable.Pop()

	if statement.Init != nil {
		sp.StmtParser.Parse(statement.Init)
	}

	if statement.Cond != nil {
		if _, err := sp.ExprParser.Parse(statement.Cond); err != nil {
			return err
		}
	}

	if statement.Post != nil {
		sp.StmtParser.Parse(statement.Post)
	}

	return sp.Parse(statement.Body)
}

// RangeClause = [ ExpressionList "=" | IdentifierList ":=" ] "range" Expression .
func (sp *Parser) parseRangeStmt(statement *ast.RangeStmt) error {
	glog.V(2).Infof("Processing range statement  %#v with token %v\n", statement, token.DEFINE)
	// If the assignment is = all expression on the LHS are already defined.
	// If the assignment is := all expression on the LHS are identifiers
	xExprAttr, err := sp.ExprParser.Parse(statement.X)
	if err != nil {
		return err
	}

	sp.SymbolTable.Push()
	defer sp.SymbolTable.Pop()

	var xVarType *typevars.Variable
	if d, ok := xExprAttr.TypeVarList[0].(*typevars.Variable); ok {
		xVarType = d
	} else {
		xVarType = sp.Config.ContractTable.NewVirtualVar()
		sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
			X:   xExprAttr.TypeVarList[0],
			Y:   xVarType,
			Pos: fmt.Sprintf("%v:%v", sp.Config.FileName, statement.Pos()),
		})
	}

	sp.Config.ContractTable.AddContract(&contracts.IsRangeable{
		X:   xVarType,
		Pos: fmt.Sprintf("%v/%v:%v", sp.Config.PackageName, sp.Config.FileName, statement.Pos()),
	})

	key, value, rErr := propagation.New(sp.Config.SymbolsAccessor).RangeExpr(xExprAttr.DataTypeList[0])
	if rErr != nil {
		return rErr
	}

	if statement.Key != nil {
		if statement.Tok == token.DEFINE {
			keyIdent, ok := statement.Key.(*ast.Ident)
			if !ok {
				return fmt.Errorf("Expected key as an identifier in the for range statement. Got %#v instead.", statement.Key)
			}
			glog.V(2).Infof("Processing range.Key variable %v", keyIdent.Name)

			if keyIdent.Name != "_" {
				sDef := &symbols.SymbolDef{
					Name:    keyIdent.Name,
					Package: sp.PackageName,
					Def:     key,
					Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, keyIdent.Pos()),
				}
				if err := sp.SymbolTable.AddVariable(sDef); err != nil {
					return err
				}
				sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
					X:            typevars.MakeRangeKey(xVarType),
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: key,
					Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, keyIdent.Pos()),
				})
			}
		} else {
			keyExpr := statement.Key
			for {
				if pe, ok := keyExpr.(*ast.ParenExpr); ok {
					keyExpr = pe.X
					continue
				}
				break
			}

			keyIdent, ok := keyExpr.(*ast.Ident)
			if !ok {
				return fmt.Errorf("Expected key as an identifier in the for range statement. Got %#v instead.", statement.Key)
			}
			glog.V(2).Infof("Processing range.Key variable %v", keyIdent.Name)

			if keyIdent.Name != "_" {
				keyAttr, err := sp.ExprParser.Parse(keyIdent)
				if err != nil {
					return nil
				}
				sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
					X:            typevars.MakeRangeKey(xVarType),
					Y:            keyAttr.TypeVarList[0],
					ExpectedType: keyAttr.DataTypeList[0],
				})
			}
		}
	}

	if statement.Value != nil {
		if statement.Tok == token.DEFINE {
			valueIdent, ok := statement.Value.(*ast.Ident)
			if !ok {
				return fmt.Errorf("Expected value as an identifier in the for range statement. Got %#v instead.", statement.Value)
			}
			glog.V(2).Infof("Processing range.Value variable %v", valueIdent.Name)
			if valueIdent.Name != "_" && value != nil {
				sDef := &symbols.SymbolDef{
					Name:    valueIdent.Name,
					Package: sp.PackageName,
					Def:     value,
					Pos:     fmt.Sprintf("%v:%v", sp.Config.FileName, valueIdent.Pos()),
				}
				if err := sp.SymbolTable.AddVariable(sDef); err != nil {
					return err
				}
				sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
					X:            typevars.MakeRangeValue(xVarType),
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: value,
					Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, valueIdent.Pos()),
				})
			}
		} else {
			valueExpr := statement.Value
			for {
				if pe, ok := valueExpr.(*ast.ParenExpr); ok {
					valueExpr = pe.X
					continue
				}
				break
			}

			valueIdent, ok := statement.Value.(*ast.Ident)
			if !ok {
				return fmt.Errorf("Expected value as an identifier in the for range statement. Got %#v instead.", statement.Value)
			}
			glog.V(2).Infof("Processing range.Value variable %v", valueIdent.Name)

			if valueIdent.Name != "_" && value != nil {
				valueAttr, err := sp.ExprParser.Parse(statement.Value)
				if err != nil {
					return nil
				}
				sp.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
					X:            typevars.MakeRangeValue(xVarType),
					Y:            valueAttr.TypeVarList[0],
					ExpectedType: valueAttr.DataTypeList[0],
				})
			}
		}
	}

	return sp.Parse(statement.Body)
}

func (sp *Parser) parseDeferStmt(statement *ast.DeferStmt) error {
	glog.V(2).Infof("Processing defer statement  %#v\n", statement)
	_, err := sp.ExprParser.Parse(statement.Call)
	return err
}

func (sp *Parser) parseIfStmt(statement *ast.IfStmt) error {
	glog.V(2).Infof("Processing if statement  %#v\n", statement)
	// If Init; Cond { Body } Else

	// The Init part is basically another block
	if statement.Init != nil {
		sp.SymbolTable.Push()
		defer sp.SymbolTable.Pop()
		if err := sp.Parse(statement.Init); err != nil {
			return err
		}
	}

	_, err := sp.ExprParser.Parse(statement.Cond)
	if err != nil {
		return err
	}

	// Process the If-body
	if err := sp.parseBlockStmt(statement.Body); err != nil {
		return err
	}

	if statement.Else != nil {
		return sp.Parse(statement.Else)
	}
	return nil
}

func (sp *Parser) parseBlockStmt(statement *ast.BlockStmt) error {
	glog.V(2).Infof("Processing block statement  %#v\n", statement)
	sp.SymbolTable.Push()
	defer sp.SymbolTable.Pop()

	for _, blockItem := range statement.List {
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
		return sp.parseDeclStmt(stmtExpr)
	case *ast.LabeledStmt:
		return sp.parseLabeledStmt(stmtExpr)
	case *ast.ExprStmt:
		return sp.parseExprStmt(stmtExpr)
	case *ast.SendStmt:
		return sp.parseSendStmt(stmtExpr)
	case *ast.IncDecStmt:
		return sp.parseIncDecStmt(stmtExpr)
	case *ast.AssignStmt:
		return sp.parseAssignStmt(stmtExpr)
	case *ast.GoStmt:
		return sp.parseGoStmt(stmtExpr)
	case *ast.ReturnStmt:
		glog.V(2).Infof("Processing ast.ReturnStmt: %#v", stmtExpr.Results)
		for _, result := range stmtExpr.Results {
			_, err := sp.ExprParser.Parse(result)
			if err != nil {
				return err
			}
		}
	case *ast.BranchStmt:
		return sp.parseBranchStmt(stmtExpr)
	case *ast.BlockStmt:
		return sp.parseBlockStmt(stmtExpr)
	case *ast.IfStmt:
		return sp.parseIfStmt(stmtExpr)
	case *ast.SwitchStmt:
		return sp.parseSwitchStmt(stmtExpr)
	case *ast.TypeSwitchStmt:
		return sp.parseTypeSwitchStmt(stmtExpr)
	case *ast.SelectStmt:
		return sp.parseSelectStmt(stmtExpr)
	case *ast.ForStmt:
		return sp.parseForStmt(stmtExpr)
	case *ast.RangeStmt:
		return sp.parseRangeStmt(stmtExpr)
	case *ast.DeferStmt:
		return sp.parseDeferStmt(stmtExpr)
	case *ast.EmptyStmt:
		return nil
	default:
		panic(fmt.Errorf("Unknown statement %#v", statement))
	}

	return nil
}

func (sp *Parser) parseFuncHeadVariables(funcDecl *ast.FuncDecl) error {
	sp.AllocatedSymbolsTable.Lock()
	defer sp.AllocatedSymbolsTable.Unlock()
	// TOOD(jchaloup): check the id of the receiver is not the same
	// as any of the function params/return ids

	// Store symbol definitions of all params once all of them are parsed.
	// Otherwise, func PostForm(url string, data url.Values) will get processed
	// properly. The url in url.Values will be interpreted as string,
	// instead of qid.
	// The same holds for func(net, addr string) (net.Conn, error).
	// The net in the return type will get picked from the net argument of type string
	// So both params and results must be stored at the end.
	var sDefs []*symbols.SymbolDef

	if funcDecl.Type.Params != nil {
		for i, field := range funcDecl.Type.Params.List {
			glog.V(2).Infof("Parsing funcDecl.Type.Params[%v].Type: %#v\n", i, field.Type)
			def, err := sp.TypeParser.Parse(field.Type)
			if err != nil {
				return fmt.Errorf("sp.TypeParser.Parse Params: %v", err)
			}

			// field.Names is always non-empty if param's datatype is defined
			for _, name := range field.Names {
				sDef := &symbols.SymbolDef{
					Name: name.Name,
					Def:  def,
					Pos:  fmt.Sprintf("%v:%v", sp.Config.FileName, name.Pos()),
				}
				sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
					X:            typevars.MakeConstant(sp.Config.PackageName, def),
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: sDef.Def,
					Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, name.Pos()),
				})
				sDefs = append(sDefs, sDef)
			}
		}
	}

	if funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			def, err := sp.TypeParser.Parse(field.Type)
			if err != nil {
				return fmt.Errorf("sp.TypeParser.Parse Results: %v", err)
			}

			for _, name := range field.Names {
				sDef := &symbols.SymbolDef{
					Name: name.Name,
					Def:  def,
					Pos:  fmt.Sprintf("%v:%v", sp.Config.FileName, name.Pos()),
				}
				sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
					X:            typevars.MakeConstant(sp.Config.PackageName, def),
					Y:            typevars.VariableFromSymbolDef(sDef),
					ExpectedType: sDef.Def,
					Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, name.Pos()),
				})
				sDefs = append(sDefs, sDef)
			}
		}
	}

	if funcDecl.Recv != nil {
		// Receiver has a single parametr
		// https://golang.org/ref/spec#Receiver
		if len(funcDecl.Recv.List) != 1 || len(funcDecl.Recv.List[0].Names) > 2 {
			if len(funcDecl.Recv.List) != 1 {
				return fmt.Errorf("Method %q has no receiver", funcDecl.Name.Name)
			}
			return fmt.Errorf("Receiver is not a single parameter: %#v, %#v", funcDecl.Recv, funcDecl.Recv.List[0].Names)
		}

		def, err := sp.parseReceiver(funcDecl.Recv.List[0].Type, true)
		if err != nil {
			return fmt.Errorf("sp.parseReceiver: %v", err)
		}
		// the receiver can be typed only
		if funcDecl.Recv.List[0].Names != nil {
			sDef := &symbols.SymbolDef{
				Name: (*funcDecl.Recv).List[0].Names[0].Name,
				Def:  def,
				Pos:  fmt.Sprintf("%v:%v", sp.Config.FileName, (*funcDecl.Recv).List[0].Names[0].Pos()),
			}
			sp.Config.ContractTable.AddContract(&contracts.PropagatesTo{
				X:            typevars.MakeConstant(sp.Config.PackageName, def),
				Y:            typevars.VariableFromSymbolDef(sDef),
				ExpectedType: sDef.Def,
				Pos:          fmt.Sprintf("%v:%v", sp.Config.FileName, (*funcDecl.Recv).List[0].Names[0].Pos()),
			})
			sp.SymbolTable.AddVariable(sDef)
		}
	}

	for _, sDef := range sDefs {
		sp.SymbolTable.AddVariable(sDef)
	}
	return nil
}
