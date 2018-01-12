package expression

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/propagation"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

func init() {
	if testing.Verbose() {
		flag.Set("alsologtostderr", "true")
		flag.Set("v", "5")
	}
}

var binaryOperators = map[token.Token]struct{}{
	token.ADD: {}, token.SUB: {}, token.MUL: {}, token.QUO: {}, token.REM: {},
	token.AND: {}, token.OR: {}, token.XOR: {}, token.SHL: {}, token.SHR: {},
	token.AND_NOT: {}, token.LAND: {}, token.LOR: {}, token.EQL: {},
	token.LSS: {}, token.GTR: {}, token.NEQ: {}, token.LEQ: {}, token.GEQ: {},
}

func isBinaryOperator(operator token.Token) bool {
	_, ok := binaryOperators[operator]

	return ok
}

// Parser parses go expressions
type Parser struct {
	*types.Config
}

// parseBasicLit consumes *ast.BasicLit and produces Builtin
func (ep *Parser) parseBasicLit(lit *ast.BasicLit) (*types.ExprAttribute, error) {
	glog.Infof("Processing BasicLit: %#v\n", lit)
	// TODO(jchaloup): store the literal as well so one can argue about []int{1, 2.0} and []int{1, 2.2}
	var builtin *gotypes.Constant
	switch lit.Kind {
	case token.INT:
		builtin = &gotypes.Constant{Def: "int"}
	case token.FLOAT:
		builtin = &gotypes.Constant{Def: "float64"}
	case token.IMAG:
		builtin = &gotypes.Constant{Def: "complex128"}
	case token.STRING:
		builtin = &gotypes.Constant{Def: "string"}
	case token.CHAR:
		builtin = &gotypes.Constant{Def: "rune"}
	default:
		return nil, fmt.Errorf("Unrecognize BasicLit: %#v\n", lit.Kind)
	}

	builtin.Untyped = true
	if lit.Kind == token.CHAR {
		// convert rune to uint8 number
		builtin.Literal = fmt.Sprintf("%v", lit.Value[1])
	} else {
		builtin.Literal = lit.Value
	}
	builtin.Package = "builtin"

	return types.ExprAttributeFromDataType(builtin).AddTypeVar(typevars.MakeConstant("builtin", builtin)), nil
}

// to parse *gotypes.Array, *gotypes.Slice, *gotypes.Map
func (ep *Parser) parseKeyValueLikeExpr(litDataType gotypes.DataType, lit *ast.CompositeLit, litType gotypes.DataType) (typevars.Interface, error) {
	glog.Infof("Processing parseKeyValueLikeExpr")
	var valueTypeVar typevars.Interface
	var keyTypeVar typevars.Interface
	var valueType gotypes.DataType

	outputTypeVar := ep.Config.ContractTable.NewVirtualVar()

	switch elmExpr := litType.(type) {
	case *gotypes.Slice:
		valueType = elmExpr.Elmtype
		valueTypeVar = typevars.MakeListValue(outputTypeVar)
		keyTypeVar = typevars.MakeListKey()
	case *gotypes.Array:
		valueType = elmExpr.Elmtype
		valueTypeVar = typevars.MakeListValue(outputTypeVar)
		keyTypeVar = typevars.MakeListKey()
	case *gotypes.Map:
		valueType = elmExpr.Valuetype
		valueTypeVar = typevars.MakeMapValue(outputTypeVar)
		keyTypeVar = typevars.MakeMapKey(outputTypeVar)
	default:
		return nil, fmt.Errorf("Unknown CL type for KV elements: %#v", litType)
	}

	ep.Config.ContractTable.AddContract(&contracts.IsIndexable{
		X:   outputTypeVar,
		Key: keyTypeVar,
		Pos: fmt.Sprintf("(3)%v:%v", ep.Config.FileName, lit.Pos()),
	})

	for _, litElement := range lit.Elts {
		var valueExpr ast.Expr
		if kvExpr, ok := litElement.(*ast.KeyValueExpr); ok {
			attr, err := ep.Parse(kvExpr.Key)
			if err != nil {
				return nil, err
			}
			ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X:            keyTypeVar,
				Y:            attr.TypeVarList[0],
				ExpectedType: attr.DataTypeList[0],
			})
			valueExpr = kvExpr.Value
		} else {
			valueExpr = litElement
		}

		if clExpr, ok := valueExpr.(*ast.CompositeLit); ok {
			var typeDef gotypes.DataType
			// If the CL type of the KV element is omitted, it needs to be reconstructed from the CL type itself
			if clExpr.Type == nil {
				if pointer, ok := valueType.(*gotypes.Pointer); ok {
					typeDef = pointer.Def
					// TODO(jchaloup): should check if the pointer.X is not a pointer and fail if it is
				} else {
					typeDef = valueType
				}
			}
			// type A struct {
			// 	f int
			// }
			// type B []A
			// type C map[int]B
			// c := C{
			// 	0: {
			// 		A{
			//      A{f: 2},
			//    },
			// 	},
			// }
			// will get parsed even if it is a semantic error.
			// The semantic analysis is not currently run and
			// it is assumed each Go code is semantically correct.
			// TODO(jchaloup): find a way to detect the semantic error and report it
			attr, err := ep.parseCompositeLit(clExpr, typeDef)
			if err != nil {
				return nil, err
			}

			// If the CL type is explicitly given, it has higher priority over
			// itemOf relation. Thus the type propagates to CL's elements and
			// if them itemOf relation changes to IsCompatibleWith contract.
			if clExpr.Type == nil {
				ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
					X:            valueTypeVar,
					Y:            attr.TypeVarList[0],
					ExpectedType: attr.DataTypeList[0],
				})
			} else {
				ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
					X:            valueTypeVar,
					Y:            attr.TypeVarList[0],
					ExpectedType: attr.DataTypeList[0],
				})
			}
			continue
		}

		// E.g. in case the value is a return value of a function or a basic literal
		attr, err := ep.Parse(valueExpr)
		if err != nil {
			return nil, err
		}
		if len(attr.DataTypeList) != 1 {
			return nil, fmt.Errorf("Expected single expression for KV value, got %#v", attr.DataTypeList)
		}

		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X:            valueTypeVar,
			Y:            attr.TypeVarList[0],
			ExpectedType: attr.DataTypeList[0],
		})
	}

	// qid.id is a type variable, not a constant
	if litDataType != nil {
		var outputOriginalTypeVar typevars.Interface
		outputOriginalTypeVar = typevars.MakeConstant(ep.Config.PackageName, litDataType)
		// qid.id is a type variable, not a constant
		if sel, sok := litDataType.(*gotypes.Selector); sok {
			if qid, qok := sel.Prefix.(*gotypes.Packagequalifier); qok {
				outputOriginalTypeVar = typevars.MakeVar(
					qid.Path,
					sel.Item,
					"", // no position for a global symbol
				)
			}
		}
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: outputOriginalTypeVar,
			Y: outputTypeVar,
		})
	}

	return outputTypeVar, nil
}

// *gotypes.Struct
func (ep *Parser) parseCompositeLitStructElements(litDataType gotypes.DataType, lit *ast.CompositeLit, structDef *gotypes.Struct, structSymbol *symbols.SymbolDef) (typevars.Interface, error) {
	glog.Infof("Processing parseCompositeLitStructElements")
	structOutputTypeVar := ep.Config.ContractTable.NewVirtualVar()

	fieldCounter := 0
	fieldLen := len(structDef.Fields)
	for _, litElement := range lit.Elts {
		var fieldTypeVar *typevars.Field
		var valueExpr ast.Expr
		// if the struct fields name are omitted, the order matters
		// TODO(jchaloup): should check if all elements are KeyValueExpr or not (otherwise go compilation fails)
		if kvExpr, ok := litElement.(*ast.KeyValueExpr); ok {
			// The field must be an identifier
			keyDefIdentifier, ok := kvExpr.Key.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("Struct's field %#v is not an identifier", litElement)
			}
			ep.Config.ContractTable.AddContract(&contracts.HasField{
				X:     structOutputTypeVar,
				Field: keyDefIdentifier.Name,
				Pos:   fmt.Sprintf("%v:%v", ep.Config.FileName, keyDefIdentifier.Pos()),
			})
			fieldTypeVar = typevars.MakeField(structOutputTypeVar, keyDefIdentifier.Name, 0, fmt.Sprintf("%v:%v", ep.Config.FileName, keyDefIdentifier.Pos()))
			valueExpr = kvExpr.Value
		} else {
			if fieldCounter >= fieldLen {
				return nil, fmt.Errorf("Number of fields of the CL is greater than the number of fields of underlying struct %#v", structDef)
			}
			ep.Config.ContractTable.AddContract(&contracts.HasField{
				X:     structOutputTypeVar,
				Index: fieldCounter,
				Pos:   fmt.Sprintf("%v:%v", ep.Config.FileName, litElement.Pos()),
			})
			fieldTypeVar = typevars.MakeField(structOutputTypeVar, "", fieldCounter, fmt.Sprintf("%v:%v", ep.Config.FileName, litElement.Pos()))
			valueExpr = litElement
		}

		// process the field value
		attr, err := ep.Parse(valueExpr)
		if err != nil {
			return nil, err
		}
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X:            fieldTypeVar,
			Y:            attr.TypeVarList[0],
			ExpectedType: attr.DataTypeList[0],
		})
		fieldCounter++
	}

	if litDataType != nil {
		var structOriginalTypeVar typevars.Interface
		structOriginalTypeVar = typevars.MakeConstant(ep.Config.PackageName, litDataType)
		// qid.id is a type variable, not a constant
		if sel, sok := litDataType.(*gotypes.Selector); sok {
			if qid, qok := sel.Prefix.(*gotypes.Packagequalifier); qok {
				structOriginalTypeVar = typevars.MakeVar(
					qid.Path,
					sel.Item,
					"", // no position for a global symbol
				)
			}
		}
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: structOriginalTypeVar,
			Y: structOutputTypeVar,
		})
	}

	return structOutputTypeVar, nil
}

// parseCompositeLit consumes ast.CompositeLit and produces data type of the root composite literal
func (ep *Parser) parseCompositeLit(lit *ast.CompositeLit, typeDef gotypes.DataType) (*types.ExprAttribute, error) {
	glog.Infof("Processing CompositeLit: %#v\t\ttypeDef: %#v\n", lit, typeDef)
	// https://golang.org/ref/spec#Composite_literals
	// The LiteralType's underlying type must be a struct, array, slice, or map
	//  type (the grammar enforces this constraint except when the type is given as a TypeName)
	//
	// lit.Type{
	//   Elts[0],
	//   Elts[1],
	//   Elts[2],
	//}

	// Other links:
	// - https://medium.com/golangspec/composite-literals-in-go-10dc62eec06a
	// - https://groups.google.com/forum/#!msg/golang-nuts/s4CRRj6f6mw/SNKRAGbDZf0J
	// - http://grokbase.com/t/gg/golang-nuts/139exrszge/go-nuts-accessing-an-anonymous-struct-or-missing-type-in-composite-literal
	// - http://www.hnwatcher.com/r/2447374/Stupid-Go-lang-Tricks
	// - http://stackoverflow.com/a/30716481/4983496
	//
	// Experiments:
	// - Given "type Identifier struct {Def string}" type definition,
	//   double referrence of "&(&Identifier{"iii"})" ends with "invalid pointer type **Identifier for composite literal"
	// - The same holds for "&(&struct {field1 string; field2 int}{field1: "", field2: 3})"

	// Based on the Composite Literal grammar, if the CL type is omitted,
	// it must be reconstructuble from the context. Thus, if the is ommited,
	// i.e. the CL type expr is set to nil, the type is constructed from the parent CL type
	// ast.
	//
	// If the CL type is a pointer to pointer (and possibly on), the compiler fails
	// when the CL type is ommited when used in CL parent elements. Thus,
	// we just need to check the CL type is at most a pointer to a non-pointer type.
	// If it is not and the CL type is still set to nil, it is a semantic error.

	// The CL type can be processed independently of the CL elements
	var litTypedef gotypes.DataType
	var origLitTypedef gotypes.DataType
	if typeDef != nil {
		litTypedef = typeDef
	} else {
		def, err := ep.TypeParser.Parse(lit.Type)
		if err != nil {
			return nil, err
		}
		litTypedef = def
		origLitTypedef = def
	}

	glog.Infof("litTypedef: %#v", litTypedef)
	nonIdentLitTypeDef, err := ep.SymbolsAccessor.FindFirstNonidDataType(litTypedef)
	if err != nil {
		return nil, err
	}
	glog.Infof("nonIdentLitTypeDef: %#v", nonIdentLitTypeDef)

	// If the CL type is anonymous struct, array, slice or map don't store fields into the allocated symbols table (AST)
	var typeVar typevars.Interface
	switch litTypeExpr := nonIdentLitTypeDef.(type) {
	case *gotypes.Struct:
		// anonymous structure -> we can ignore field's allocation
		var err error
		typeVar, err = ep.parseCompositeLitStructElements(origLitTypedef, lit, litTypeExpr, nil)
		if err != nil {
			return nil, err
		}
	case *gotypes.Array, *gotypes.Slice, *gotypes.Map:
		glog.Infof("parseCompositeLitArrayLikeElements: %#v\n", lit)
		var err error
		// if origLitTypedef != nil {
		// 	switch origLitTypedef.(type) {
		// 	case *gotypes.Array, *gotypes.Slice, *gotypes.Map:
		// 	default:
		// 		origLitTypedef = nil
		// 	}
		// }
		typeVar, err = ep.parseKeyValueLikeExpr(origLitTypedef, lit, nonIdentLitTypeDef)
		if err != nil {
			return nil, err
		}
	default:
		panic(fmt.Errorf("Unsupported CL type: %#v", nonIdentLitTypeDef))
	}

	// If the CL type is explicitly given, it has higher priority over
	// itemOf relation. Thus the type propagates to CL's elements and
	// if them itemOf relation changes to IsCompatibleWith contract.
	if lit.Type != nil {
		def, err := ep.TypeParser.Parse(lit.Type)
		if err != nil {
			return nil, err
		}
		ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
			X:            typevars.MakeConstant(ep.Config.PackageName, def),
			Y:            typeVar,
			ExpectedType: def,
			Pos:          fmt.Sprintf("%v:%v", ep.Config.FileName, lit.Pos()),
		})
	}

	return types.ExprAttributeFromDataType(litTypedef).AddTypeVar(typeVar), nil
}

// parseIdentifier consumes ast.Ident and produces:
// - if the identifier is qid, qid definition is returned
// - if the identifier is a function, function definition is returned
// - if the identifier is a local/global variable, data type of the variable is returned
//   (e.g. the data type can be a method/function definition)
// Unless origin of a data type is part of the data type definition itself, it is not propagated.
// Assumptions:
// - the identifier is either a qid, a local/global variable or a function of the current package.
//   identifiers of other packages are handled differently (TODO(jchaloup): describe where and how)
// - the identifier can not be a method (though, it can be a value of a local/global variable)
func (ep *Parser) parseIdentifier(ident *ast.Ident) (*types.ExprAttribute, error) {
	glog.Infof("Processing variable-like identifier: %#v\n", ident)

	// Does the local symbol exists at all?
	var postponedErr error
	if !ep.SymbolTable.Exists(ident.Name) {
		// E.g. user can define
		//    make := func() {}
		// which overwrites the builtin.make function
		postponedErr = fmt.Errorf("parseIdentifier: Symbol %q not yet processed", ident.Name)
	} else {
		// If it is a variable, return its definition
		if def, st, err := ep.SymbolTable.LookupVariableLikeSymbol(ident.Name); err == nil {
			byteSlice, _ := json.Marshal(def)
			glog.Infof("Variable by identifier found: %v\n", string(byteSlice))

			if def.Block == 0 && st == symbols.VariableSymbol && def.Def.GetType() != gotypes.PackagequalifierType {
				ep.AllocatedSymbolsTable.AddVariable(def.Package, def.Name, fmt.Sprintf("%v.%v", ep.Config.FileName, ident.Pos()))
			}
			// The data type of the variable is not accounted as it is not implicitely used
			// The variable itself carries the data type and as long as the variable does not
			// get used, the data type can change.
			// TODO(jchaloup): return symbol origin
			if st == symbols.FunctionSymbol {
				return types.ExprAttributeFromDataType(def.Def).AddTypeVar(
					typevars.VariableFromSymbolDef(def),
				), nil
			}
			return types.ExprAttributeFromDataType(def.Def).AddTypeVar(
				typevars.VariableFromSymbolDef(def),
			), nil
		}
	}

	// Maybe it is a builtin variable
	table, err := ep.GlobalSymbolTable.Lookup("builtin")
	if err != nil {
		return nil, err
	}
	symbolDef, symbolType, symbolErr := table.Lookup(ident.Name)

	if symbolErr == nil {
		switch symbolType {
		case symbols.VariableSymbol:
			switch ident.Name {
			case "true", "false":
				c := &gotypes.Constant{Package: "builtin", Def: "bool", Literal: ident.Name, Untyped: true}
				return types.ExprAttributeFromDataType(c).AddTypeVar(
					typevars.MakeConstant("builtin", c),
				), nil
			case "nil":
				return types.ExprAttributeFromDataType(&gotypes.Nil{}).AddTypeVar(
					typevars.MakeConstant("builtin", &gotypes.Nil{}),
				), nil
			case "iota":
				c := &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: fmt.Sprintf("%v", ep.Config.Iota)}
				return types.ExprAttributeFromDataType(c).AddTypeVar(
					typevars.MakeConstant("builtin", c),
				), nil
			default:
				return nil, fmt.Errorf("Unsupported built-in type: %v", ident.Name)
			}
		case symbols.FunctionSymbol:
			return types.ExprAttributeFromDataType(symbolDef.Def).AddTypeVar(
				typevars.VariableFromSymbolDef(symbolDef),
			), nil
		default:
			return nil, fmt.Errorf("Unsupported symbol type: %v", symbolType)
		}
	}

	// if postponedErr is nil, something is wrong
	if postponedErr == nil {
		panic("postponedErr can not be nil here")
	}

	return nil, postponedErr
}

// parseUnaryExpr consumes ast.UnaryExpr and produces:
// - token.AND is understood as a dereference operator
// - token.ARROW is understood as a channel operator
// - token.XOR, token.OR, token.SUB, token.NOT, token.ADD are understood as
//   their unary equivalents (^OP, |OP, -OP, ~OP, +OP)
func (ep *Parser) parseUnaryExpr(expr *ast.UnaryExpr) (*types.ExprAttribute, error) {
	glog.Infof("Processing UnaryExpr: %#v\n", expr)
	attr, err := ep.Parse(expr.X)
	if err != nil {
		return nil, err
	}

	if len(attr.DataTypeList) != 1 {
		return nil, fmt.Errorf("Operand of an unary operator is not a single value")
	}

	yDataType, err := propagation.New(ep.Config.SymbolsAccessor).UnaryExpr(expr.Op, attr.DataTypeList[0])
	if err != nil {
		return nil, err
	}

	y := ep.Config.ContractTable.NewVirtualVar()
	if expr.Op == token.ARROW {
		ep.Config.ContractTable.AddContract(&contracts.IsReceiveableFrom{
			X:            attr.TypeVarList[0],
			Y:            y,
			ExpectedType: yDataType,
		})
		// TODO(jchaloup): add ReceivedFrom contract
	} else {
		if expr.Op == token.AND {
			ep.Config.ContractTable.AddContract(&contracts.IsReferenceable{
				X: attr.TypeVarList[0],
			})
			ep.Config.ContractTable.AddContract(&contracts.ReferenceOf{
				X: attr.TypeVarList[0],
				Y: y,
			})
		} else {
			ep.Config.ContractTable.AddContract(&contracts.UnaryOp{
				OpToken:      expr.Op,
				X:            attr.TypeVarList[0],
				Y:            y,
				ExpectedType: yDataType,
			})
		}
	}

	return types.ExprAttributeFromDataType(yDataType).AddTypeVar(y), nil
}

// parseBinaryExpr consumes ast.BinaryExpr and produces:
// - if the operator results in boolean data type, the boolean is returned (without checking data type of operands)
// - if both operands are untyped, their result data type is untyped
// - if any of the operands is typed, their result data type is typed
// - if one of the operands is of byte and the other one of uint8(3), their result data type is uint8
// - if both operands are non-builtin typed data types, data type of the first operand is returned
// - if exactly one of the operands is non-builtin typed data type, the non-builtin typed data type is returned
// Facts::
// - result data type is always another identifier, not data type definition itself
//   i.e. MyInt + int = MyInt, not MyInt definition
func (ep *Parser) parseBinaryExpr(expr *ast.BinaryExpr) (*types.ExprAttribute, error) {
	glog.Infof("Processing Binaryexpr: %#v\n", expr)
	if !isBinaryOperator(expr.Op) {
		return nil, fmt.Errorf("Binary operator %#v not recognized", expr.Op)
	}

	// Given all binary operators are valid only for build-in types
	// and any operator result is built-in type again,
	// the Buildin is return.
	// However, as the operands itself can be built of
	// user defined type returning built-in type, both operands must be processed.

	// X Op Y
	xAttr, yErr := ep.Parse(expr.X)
	if yErr != nil {
		return nil, yErr
	}

	yAttr, xErr := ep.Parse(expr.Y)
	if xErr != nil {
		return nil, xErr
	}

	{
		byteSlice, _ := json.Marshal(xAttr.DataTypeList[0])
		glog.Infof("xx: %v\n", string(byteSlice))
	}

	{
		byteSlice, _ := json.Marshal(yAttr.DataTypeList[0])
		glog.Infof("yy: %v\n", string(byteSlice))
	}

	if len(xAttr.DataTypeList) != 1 {
		return nil, fmt.Errorf("First operand of a binary operator is not a single value")
	}

	if len(yAttr.DataTypeList) != 1 {
		return nil, fmt.Errorf("Second operand of a binary operator is not a single value")
	}

	zDataType, err := propagation.New(ep.Config.SymbolsAccessor).BinaryExpr(expr.Op, xAttr.DataTypeList[0], yAttr.DataTypeList[0])
	if err != nil {
		return nil, fmt.Errorf("parseBinaryExpr: %v, at pos %v", err, expr.Pos())
	}

	z := ep.Config.ContractTable.NewVirtualVar()
	ep.Config.ContractTable.AddContract(&contracts.BinaryOp{
		OpToken:      expr.Op,
		X:            xAttr.TypeVarList[0],
		Y:            yAttr.TypeVarList[0],
		Z:            z,
		ExpectedType: zDataType,
		Pos:          fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()),
	})
	return types.ExprAttributeFromDataType(zDataType).AddTypeVar(z), nil
}

// parseStarExpr consumes ast.StarExpr and produces:
// - if the expression is a pointer, pointed data type is returned
// Errors:
// - if the expression is non-pointer data type
func (ep *Parser) parseStarExpr(expr *ast.StarExpr) (*types.ExprAttribute, error) {
	glog.Infof("Processing StarExpr: %#v\n", expr)
	attr, err := ep.Parse(expr.X)
	if err != nil {
		return nil, err
	}

	if len(attr.DataTypeList) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}

	val, ok := attr.DataTypeList[0].(*gotypes.Pointer)
	if !ok {
		return nil, fmt.Errorf("Accessing a value of non-pointer type: %#v", attr.DataTypeList[0])
	}
	ep.Config.ContractTable.AddContract(&contracts.IsDereferenceable{
		X: attr.TypeVarList[0],
	})

	y := ep.Config.ContractTable.NewVirtualVar()
	ep.Config.ContractTable.AddContract(&contracts.DereferenceOf{
		X: attr.TypeVarList[0],
		Y: y,
	})

	return types.ExprAttributeFromDataType(val.Def).AddTypeVar(y), nil
}

// parseParenExpr consumes ast.Expr and drops direct sequence of parenthesis
func (ep *Parser) parseParenExpr(expr ast.Expr) (e ast.Expr) {
	e = expr
	for {
		if pe, ok := e.(*ast.ParenExpr); ok {
			e = pe.X
			continue
		}
		return
	}
}

func (ep *Parser) isDataType(expr ast.Expr) (bool, error) {
	glog.Infof("Detecting isDataType for %#v at %v", expr, expr.Pos())
	switch exprType := expr.(type) {
	case *ast.Ident:
		// user defined type
		// try to find any local symbol, if it is not found, the local symbol is unknown
		var postponedErr error
		if !ep.SymbolTable.Exists(exprType.Name) {
			postponedErr = fmt.Errorf("isDataType: symbol %q is unknown", exprType.Name)
		} else {
			if _, err := ep.SymbolTable.LookupDataType(exprType.Name); err == nil {
				return true, nil
			}
		}

		// not a user defined type
		if table, err := ep.GlobalSymbolTable.Lookup("builtin"); err == nil {
			glog.Infof("isDataType Builtin seaching for %q", exprType.Name)
			if _, err := table.LookupDataType(exprType.Name); err == nil {
				glog.Info("isDataType Builtin found")
				return true, nil
			}
			return false, nil
		}

		// neither user defined type, nor builtin type
		return false, postponedErr
	case *ast.SelectorExpr:
		// Selector based data type is in a form qid.typeid
		// If the selector.X is not an identifier => variable
		ident, ok := exprType.X.(*ast.Ident)
		if !ok {
			return false, nil
		}
		// is ident qid?
		qiddef, err := ep.SymbolTable.LookupVariable(ident.Name)
		if err != nil {
			return false, nil
		}
		qid, ok := qiddef.Def.(*gotypes.Packagequalifier)
		if !ok {
			return false, nil
		}
		if table, err := ep.GlobalSymbolTable.Lookup(qid.Path); err == nil {
			if _, err := table.LookupDataType(exprType.Sel.Name); err == nil {
				return true, nil
			}
		}
		// the symbol table for qid not found => get catched later
		return false, nil
	case *ast.ArrayType:
		return true, nil
	case *ast.StarExpr:
		return ep.isDataType(exprType.X)
	case *ast.ParenExpr:
		return ep.isDataType(exprType.X)
	case *ast.FuncLit:
		// Empty body => anonymous function type assertion
		// E.g. (func(int, int))funcid
		if exprType.Body == nil {
			return true, nil
		}
		return false, nil
		// TODO(jchaloup): what about (&functionidvar)(args)?
	case *ast.FuncType:
		return true, nil
	case *ast.InterfaceType:
		return true, nil
	case *ast.TypeAssertExpr:
		return false, nil
	case *ast.ChanType:
		return true, nil
	case *ast.MapType:
		return true, nil
	case *ast.CallExpr:
		return false, nil
	case *ast.StructType:
		return true, nil
	case *ast.IndexExpr:
		return false, nil
	default:
		// TODO(jchaloup): yes? As now it is anonymous data type. Or should we check for each such type?
		panic(fmt.Errorf("Unrecognized isDataType expr: %#v at %v", expr, expr.Pos()))
	}
	return false, nil
}

func (ep *Parser) getFunctionDef(def gotypes.DataType) (*types.ExprAttribute, []string, error) {
	glog.Infof("getFunctionDef of %#v", def)
	switch typeDef := def.(type) {
	case *gotypes.Identifier:
		// local definition
		var def *symbols.SymbolDef
		var defType symbols.SymbolType
		var err error

		if typeDef.Package == "" || typeDef.Package == ep.PackageName {
			def, defType, err = ep.SymbolTable.Lookup(typeDef.Def)
		} else {
			table, tErr := ep.GlobalSymbolTable.Lookup(typeDef.Package)
			if tErr != nil {
				return nil, nil, tErr
			}
			def, defType, err = table.Lookup(typeDef.Def)
		}
		if err != nil {
			return nil, nil, err
		}
		if def.Def == nil {
			return nil, nil, fmt.Errorf("Symbol %q not yet fully processed", def.Name)
		}
		if defType.IsFunctionType() {
			return types.ExprAttributeFromDataType(def.Def), nil, nil
		}
		if defType.IsDataType() {
			return ep.getFunctionDef(def.Def)
		}
		if !defType.IsVariable() {
			return nil, nil, fmt.Errorf("Function call expression %#v is not expected to be a type", def)
		}
		return ep.getFunctionDef(def.Def)
	case *gotypes.Selector:
		_, sd, err := ep.SymbolsAccessor.RetrieveQidDataType(typeDef.Prefix, &ast.Ident{Name: typeDef.Item})
		if err != nil {
			return nil, nil, err
		}
		return ep.getFunctionDef(sd.Def)
	case *gotypes.Function:
		return types.ExprAttributeFromDataType(def), nil, nil
	case *gotypes.Method:
		return types.ExprAttributeFromDataType(def), nil, nil
	case *gotypes.Pointer:
		return nil, nil, fmt.Errorf("Can not invoke non-function expression: %#v", def)
	default:
		panic(fmt.Errorf("Unrecognized getFunctionDef def: %#v", def))
	}
	return nil, nil, nil
}

// parseCallExpr consumes ast.CallExpr and produces:
// - if the call expression is a data type, the data type itself is returned
// - if the call expression is a function/method, result data type of the function/method is returned
// Assumptions:
// - type assertion is always valid (checked during compilation)
// - all arguments are always assignable to function/method parameters
// Errors:
// - number of arguments is different from a number of parameters (including variable length of parameters)
func (ep *Parser) parseCallExpr(expr *ast.CallExpr) (*types.ExprAttribute, error) {
	glog.Infof("Processing CallExpr: %#v\n", expr)
	defer glog.Infof("Leaving CallExpr: %#v\n", expr)

	expr.Fun = ep.parseParenExpr(expr.Fun)

	// params = list of function/method parameters definition (from its signature)
	processArgs := func(functionTypeVar *typevars.Variable, args []ast.Expr, params []gotypes.DataType) error {
		// TODO(jchaloup): check the arguments can be assigned to the parameters
		// TODO(jchaloup): generate type contract for each argument
		if params != nil {
			if len(args) == 1 {
				attr, err := ep.Parse(args[0])
				if err != nil {
					return err
				}
				// e.g. f(...) (a, b, c); g(a,b,c) ...; g(f(...))
				if len(params) == len(attr.DataTypeList) {
					for i := range attr.DataTypeList {
						if functionTypeVar != nil {
							ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
								X:            attr.TypeVarList[i],
								Y:            typevars.MakeArgument(functionTypeVar, i),
								ExpectedType: attr.DataTypeList[i],
							})
						}
					}
					return nil
				}
				if len(attr.DataTypeList) != 1 {
					return fmt.Errorf("Argument %#v of a call expression does not have one return value", args[0])
				}

				// E.g. func f(a string, b ...int) called as f("aaa"), the 'b' is nil then
				if functionTypeVar != nil {
					ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
						X:            attr.TypeVarList[0],
						Y:            typevars.MakeArgument(functionTypeVar, 0),
						ExpectedType: attr.DataTypeList[0],
					})
				}
				return nil
			}
		}
		for i, arg := range args {
			// an argument is passed to the function so its data type does not affect the result data type of the call
			attr, err := ep.Parse(arg)
			if err != nil {
				return err
			}

			if attr == nil || len(attr.DataTypeList) != 1 {
				return fmt.Errorf("Argument %#v of a call expression does not have one return value", arg)
			}

			if functionTypeVar != nil {
				ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
					X:            attr.TypeVarList[0],
					Y:            typevars.MakeArgument(functionTypeVar, i),
					ExpectedType: attr.DataTypeList[0],
				})
			}
		}
		return nil
	}

	// data type => explicit type casting
	// TODO(jchaloup): if len(expr.Args) != 1 => not a data type
	isType, err := ep.isDataType(expr.Fun)
	if err != nil {
		return nil, err
	}
	if isType {
		glog.Infof("isDataType of %#v is true", expr.Fun)
		def, err := ep.TypeParser.Parse(expr.Fun)
		if err != nil {
			return nil, err
		}

		attr, err := ep.Parse(expr.Args[0])
		if err != nil {
			return nil, err
		}

		if attr == nil || len(attr.DataTypeList) != 1 {
			return nil, fmt.Errorf("Argument %#v of a call expression does not have one return value", expr.Args[0])
		}

		castedDef, err := propagation.New(ep.Config.SymbolsAccessor).TypecastExpr(attr.DataTypeList[0], def)
		if err != nil {
			return nil, fmt.Errorf("Unable to type-cast: %v", err)
		}

		glog.Infof("Casted to %#v\n", castedDef)

		newVar := ep.Config.ContractTable.NewVirtualVar()

		// get type's origin
		dt := attr.DataTypeList[0]
		var dtOrigin string
		pointer, ok := dt.(*gotypes.Pointer)
		for {
			pointer, ok = dt.(*gotypes.Pointer)
			if !ok {
				break
			}
			dt = pointer.Def
		}

		switch d := dt.(type) {
		case *gotypes.Identifier:
			dtOrigin = d.Package
		case *gotypes.Selector:
			qid, ok := d.Prefix.(*gotypes.Packagequalifier)
			if !ok {
				fmt.Printf("slector not qid: %#v\n", d.Prefix)
				panic("SELECTOR not QID!!!")
			}
			dtOrigin = qid.Path
		default:
			dtOrigin = ep.Config.PackageName
		}

		ep.Config.ContractTable.AddContract(&contracts.TypecastsTo{
			X:            attr.TypeVarList[0],
			Type:         typevars.MakeConstant(dtOrigin, def),
			Y:            newVar,
			ExpectedType: def,
		})

		return types.ExprAttributeFromDataType(castedDef).AddTypeVar(newVar), nil
	}
	glog.Infof("isDataType of %#v is false", expr.Fun)

	// function
	attr, err := ep.ExprParser.Parse(expr.Fun)
	if err != nil {
		return nil, err
	}

	funcDefAttr, _, err := ep.getFunctionDef(attr.DataTypeList[0])
	if err != nil {
		return nil, err
	}

	var params []gotypes.DataType
	var results []gotypes.DataType

	// a) f1() (int, int), f2(int, int) => f2(f1())
	// b) f(arg, arg, arg) => f(a,a,a)
	switch funcType := funcDefAttr.DataTypeList[0].(type) {
	case *gotypes.Function:
		// built-in make or new?
		if ident, isIdent := expr.Fun.(*ast.Ident); isIdent {
			switch ident.Name {
			case "make":
				// arglen == 1: make(type) type
				// arglen == 2: make(type, size) type
				// arglen == 3: make(type, size, cap) type
				f := typevars.MakeVar("builtin", ident.Name, fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()))
				arglen := len(expr.Args)
				switch arglen {
				case 1:
					glog.Infof("Processing make arguments for make(type) type: %#v", expr.Args)
				case 2:
					glog.Infof("Processing make arguments for make(type, size) type: %#v", expr.Args)
					if err := processArgs(f, []ast.Expr{expr.Args[1]}, nil); err != nil {
						return nil, err
					}
				case 3:
					glog.Infof("Processing make arguments for make(type, size, size) type: %#v", expr.Args)
					if err := processArgs(f, []ast.Expr{expr.Args[1], expr.Args[2]}, nil); err != nil {
						return nil, err
					}
				default:
					return nil, fmt.Errorf("Expecting 1, 2 or 3 arguments of built-in function make, got %q instead", arglen)
				}
				ep.Config.ContractTable.AddContract(&contracts.IsInvocable{
					F:         f,
					ArgsCount: arglen,
				})

				typeDef, err := ep.TypeParser.Parse(expr.Args[0])
				// TODO(jchaloup): set the right type's package
				return types.ExprAttributeFromDataType(typeDef).AddTypeVar(
					typevars.MakeConstant(ep.Config.PackageName, typeDef),
				), err
			case "new":
				// The new built-in function allocates memory. The first argument is a type,
				// not a value, and the value returned is a pointer to a newly
				// allocated zero value of that type.
				if len(expr.Args) != 1 {
					return nil, fmt.Errorf("Len of new args != 1, it is %#v", expr.Args)
				}
				f := typevars.MakeVar("builtin", ident.Name, fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()))
				ep.Config.ContractTable.AddContract(&contracts.IsInvocable{
					F:         f,
					ArgsCount: 1,
				})

				typeDef, err := ep.TypeParser.Parse(expr.Args[0])
				return types.ExprAttributeFromDataType(
					&gotypes.Pointer{Def: typeDef},
				).AddTypeVar(
					// TODO(jchaloup): set the right type's package
					typevars.MakeConstant(ep.Config.PackageName, &gotypes.Pointer{Def: typeDef}),
				), err
			case "len":
				// if the argument is array or a pointer to an array
				// check if the len can be determined (in case then len is a constant).
				// See https://golang.org/ref/spec#Length_and_capacity
				// TODO(jchaloup): return the correct value/type of the len built-in function
				f := typevars.MakeVar("builtin", ident.Name, fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()))
				if err := processArgs(f, []ast.Expr{expr.Args[0]}, nil); err != nil {
					return nil, err
				}
				ep.Config.ContractTable.AddContract(&contracts.IsInvocable{
					F:         f,
					ArgsCount: 1,
				})

				// Produce an int typed constant for the moment so we don't report any false alarms
				c := &gotypes.Constant{
					Package: "builtin",
					Def:     "int",
					Literal: "*", // Greedy approximation
				}
				return types.ExprAttributeFromDataType(c).AddTypeVar(
					typevars.MakeConstant("builtin", c),
				), nil
			case "imag", "real":
				attr, err := ep.Parse(expr.Args[0])
				if err != nil {
					return nil, err
				}

				if len(attr.DataTypeList) != 1 {
					return nil, fmt.Errorf("%v's argument is not a single expression", ident.Name)
				}

				results, err = propagation.New(ep.Config.SymbolsAccessor).BuiltinFunctionInvocation(ident.Name, attr.DataTypeList)
				if err != nil {
					return nil, err
				}
				f := typevars.MakeVar("builtin", ident.Name, fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()))
				ep.Config.ContractTable.AddContract(&contracts.IsInvocable{
					F:         f,
					ArgsCount: 1,
				})
				return types.ExprAttributeFromDataType(results[0]).AddTypeVar(
					typevars.MakeConstant("builtin", results[0]),
				), nil
			case "complex":
				realAttr, err := ep.Parse(expr.Args[0])
				if err != nil {
					return nil, err
				}
				if len(realAttr.DataTypeList) != 1 {
					return nil, fmt.Errorf("%v's real argument is not a single expression", ident.Name)
				}
				imagAttr, err := ep.Parse(expr.Args[1])
				if err != nil {
					return nil, err
				}
				if len(imagAttr.DataTypeList) != 1 {
					return nil, fmt.Errorf("%v's imag argument is not a single expression", ident.Name)
				}
				results, err = propagation.New(ep.Config.SymbolsAccessor).BuiltinFunctionInvocation(ident.Name, []gotypes.DataType{realAttr.DataTypeList[0], imagAttr.DataTypeList[0]})
				if err != nil {
					return nil, err
				}
				f := typevars.MakeVar("builtin", ident.Name, fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()))
				ep.Config.ContractTable.AddContract(&contracts.IsInvocable{
					F:         f,
					ArgsCount: 2,
				})
				return types.ExprAttributeFromDataType(results[0]).AddTypeVar(
					typevars.MakeConstant("builtin", results[0]),
				), nil
			}
		}
		// built-in unsafe.Sizeof
		// https://golang.org/pkg/unsafe/
		if sel, isSel := expr.Fun.(*ast.SelectorExpr); isSel {
			if ident, ok := sel.X.(*ast.Ident); ok {
				if ident.Name == "unsafe" {
					switch sel.Sel.Name {
					case "Sizeof", "Alignof", "Offsetof":
						if err := processArgs(typevars.MakeVar("unsafe", sel.Sel.Name, ""), []ast.Expr{expr.Args[0]}, nil); err != nil {
							return nil, err
						}
						// always produces a constant, taking a data type and providing its size (during compilation)
						c := &gotypes.Constant{
							Package: "builtin",
							Def:     "uintptr",
							Literal: "*", // Greedy approximation
						}
						return types.ExprAttributeFromDataType(c).AddTypeVar(
							typevars.MakeConstant("builtin", c),
						), nil
					}
				}
			}
		}

		params = funcType.Params
		results = funcType.Results
	case *gotypes.Method:
		params = funcType.Def.(*gotypes.Function).Params
		results = funcType.Def.(*gotypes.Function).Results
	default:
		return nil, fmt.Errorf("Symbol %#v of %#v to be called is not a function", funcType, expr)
	}

	var f *typevars.Variable
	switch d := attr.TypeVarList[0].(type) {
	case *typevars.Variable:
		if !strings.HasPrefix(d.Name, "virtual.var") {
			f = &typevars.Variable{Package: d.Package, Name: d.Name, Pos: fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos())}
		} else {
			f = &typevars.Variable{Package: d.Package, Name: d.Name}
		}
		if d.Package == "" {
			// function literals defined through file level variables need to be
			// propagated to positions where they are invoked
			ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
				X:   d,
				Y:   f,
				Pos: fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()),
			})
		}
	case *typevars.Constant, *typevars.Field, *typevars.ReturnType, *typevars.ListValue:
		newVar := ep.Config.ContractTable.NewVirtualVar()
		ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
			X:   attr.TypeVarList[0],
			Y:   newVar,
			Pos: fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()),
		})
		f = newVar
	default:
		panic(fmt.Sprintf("expr %#v expected to be a function, got %#v instead", expr, attr.TypeVarList[0]))
	}

	ep.Config.ContractTable.AddContract(&contracts.IsInvocable{
		F:         f,
		ArgsCount: len(expr.Args),
	})

	if err := processArgs(f, expr.Args, params); err != nil {
		return nil, err
	}

	outputAttr := types.ExprAttributeFromDataType(results...)
	for i := range results {
		z := typevars.MakeReturn(f, i)
		outputAttr.AddTypeVar(z)
	}
	return outputAttr, nil
}

// parseSliceExpr consumes ast.SliceExpr and produces a data type of its slice value
// Assumptions:
// - All Low, High, Max expression are valid slice indexes
func (ep *Parser) parseSliceExpr(expr *ast.SliceExpr) (*types.ExprAttribute, error) {
	// From https://golang.org/ref/spec#Slice_types:
	// 		"The elements can be addressed by integer indices 0 through len(s)-1."
	if expr.Low != nil {
		attr, err := ep.Parse(expr.Low)
		if err != nil {
			return nil, err
		}
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: attr.TypeVarList[0],
			Y: typevars.MakeListKey(),
		})
	}
	if expr.High != nil {
		attr, err := ep.Parse(expr.High)
		if err != nil {
			return nil, err
		}
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: attr.TypeVarList[0],
			Y: typevars.MakeListKey(),
		})
	}
	if expr.Max != nil {
		attr, err := ep.Parse(expr.Max)
		if err != nil {
			return nil, err
		}
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: attr.TypeVarList[0],
			Y: typevars.MakeListKey(),
		})
	}

	exprDefAttr, err := ep.Parse(expr.X)
	if err != nil {
		return nil, err
	}

	if len(exprDefAttr.DataTypeList) != 1 {
		return nil, fmt.Errorf("SliceExpr is not a single argument")
	}

	ep.Config.ContractTable.AddContract(&contracts.IsIndexable{
		X:       exprDefAttr.TypeVarList[0],
		IsSlice: true,
		Pos:     fmt.Sprintf("(1)%v:%v", ep.Config.FileName, expr.Pos()),
	})

	return exprDefAttr, nil
}

// parseChanType consumes ast.ChanType and produces a data type corresponding to channel definition
// Example:
// - (<-chan int)(varible) // type casting to a channel type
func (ep *Parser) parseChanType(expr *ast.ChanType) (*types.ExprAttribute, error) {
	valueDefAttr, err := ep.Parse(expr.Value)
	if err != nil {
		return nil, err
	}

	if len(valueDefAttr.DataTypeList) != 1 {
		return nil, fmt.Errorf("ChanType is not a single argument")
	}

	channel := &gotypes.Channel{Value: valueDefAttr.DataTypeList[0]}

	switch expr.Dir {
	case ast.SEND:
		channel.Dir = "1"
	case ast.RECV:
		channel.Dir = "2"
	default:
		channel.Dir = "3"
	}

	return types.ExprAttributeFromDataType(channel), nil
}

// parseFuncLit consumes ast.FuncLit and produces a data type corresponding to a function signature
// Example:
// - a := func(...) (...) {...}
func (ep *Parser) parseFuncLit(expr *ast.FuncLit) (*types.ExprAttribute, error) {
	glog.Infof("Processing FuncLit: %#v\n", expr)
	if err := ep.StmtParser.ParseFuncBody(&ast.FuncDecl{
		Type: expr.Type,
		Body: expr.Body,
	}); err != nil {
		return nil, err
	}

	def, err := ep.TypeParser.Parse(expr.Type)
	newVar := ep.Config.ContractTable.NewVirtualVar()
	ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
		X:   typevars.MakeConstant(ep.Config.PackageName, def),
		Y:   newVar,
		Pos: fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()),
	})

	return types.ExprAttributeFromDataType(def).AddTypeVar(
		newVar,
	), err
}

// parseStructType consumes ast.StructType and produces data type corresponding to struct definition
// Example:
// -  (struct{int a})(variable)	// type casting to a struct type
func (ep *Parser) parseStructType(expr *ast.StructType) (*types.ExprAttribute, error) {
	glog.Infof("Processing StructType: %#v\n", expr)
	def, err := ep.TypeParser.Parse(expr)
	return types.ExprAttributeFromDataType(def), err
}

func (ep *Parser) checkAngGetDataTypeMethod(expr *ast.SelectorExpr) (bool, *types.ExprAttribute, error) {
	parExpr, ok := expr.X.(*ast.ParenExpr)
	if !ok {
		return false, nil, nil
	}
	dataTypeIdent := parExpr.X
	pointer, ok := parExpr.X.(*ast.StarExpr)
	if ok {
		dataTypeIdent = pointer.X
	}

	var typeSymbolTable symbols.SymbolLookable
	var typeSymbolDef *symbols.SymbolDef

	switch dtExpr := dataTypeIdent.(type) {
	case *ast.Ident:
		if !ep.SymbolTable.Exists(dtExpr.Name) {
			return false, nil, fmt.Errorf("Symbol %v not found", dtExpr.Name)
		}
		def, err := ep.SymbolTable.LookupDataType(dtExpr.Name)
		if err != nil {
			return false, nil, nil
		}

		typeSymbolTable = ep.SymbolTable
		typeSymbolDef = def
	case *ast.SelectorExpr:
		// qid assumed
		ident, ok := dtExpr.X.(*ast.Ident)
		if !ok {
			return false, nil, fmt.Errorf("Expecting qid.id, got %#v for the Qid instead", dtExpr.X)
		}

		qidAttr, err := ep.parseIdentifier(ident)
		if err != nil {
			return false, nil, err
		}
		if _, isQid := qidAttr.DataTypeList[0].(*gotypes.Packagequalifier); !isQid {
			return false, nil, nil
		}

		st, sd, err := ep.SymbolsAccessor.RetrieveQidDataType(qidAttr.DataTypeList[0], &ast.Ident{Name: dtExpr.Sel.String()})
		if err != nil {
			return false, nil, err
		}

		typeSymbolTable = st
		typeSymbolDef = sd
	default:
		return false, nil, nil
	}

	methodDefAttr, err := ep.SymbolsAccessor.RetrieveDataTypeField(
		accessors.NewFieldAccessor(typeSymbolTable, typeSymbolDef, &ast.Ident{Name: expr.Sel.Name}).SetMethodsOnly(),
	)
	if err != nil {
		return false, nil, fmt.Errorf("Unable to find method %q of data type %v", expr.Sel.Name, typeSymbolDef.Name)
	}
	method, ok := methodDefAttr.DataType.(*gotypes.Method)
	if !ok {
		return false, nil, fmt.Errorf("Expected a method %q of data type %v, got %#v instead", expr.Sel.Name, typeSymbolDef.Name, methodDefAttr)
	}

	y := ep.Config.ContractTable.NewVirtualVar()

	receiverDef, err := ep.TypeParser.Parse(dataTypeIdent)
	if err != nil {
		return true, nil, err
	}

	ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
		X:   typevars.MakeConstant(ep.Config.PackageName, receiverDef),
		Y:   y,
		Pos: fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Pos()),
	})

	ep.Config.ContractTable.AddContract(&contracts.HasField{
		X:     y,
		Field: expr.Sel.Name,
		Pos:   fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Sel.Pos()),
	})

	return true, types.ExprAttributeFromDataType(method.Def.(*gotypes.Function)).AddTypeVar(
		typevars.MakeField(y, expr.Sel.Name, 0, fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Sel.Pos())),
	), nil
}

// parseSelectorExpr consumes ast.SelectorExpr and produces:
// - if the prefix is a data type receiver, data type's method (by selector item) definition is returned
// - if the prefix is a pointer to a data type receiver, data type's method (by selector item) definition is returned
// - if the prefix is a package quialifier, data type of a symbol pointed by selector item of the package is returned
// - if the prefix is a pointer to a package quialifier, data type of a symbol pointed by selector item of the package is returned
// - if the prefix is a struct, data type of a field/method pointed by the selector item is returned
// - if the prefix is a pointer to a struct, data type of a field/method pointed by the selector item is returned
// - if the prefix is an identifier of a data type, data type of a method pointed by the selector item is returned
// - if the prefix is pointer to identifier of a data type, data type of a method pointed by the selector item is returned
// - if the prefix is an interface data type, data type of a method pointed by the selector item is returned
func (ep *Parser) parseSelectorExpr(expr *ast.SelectorExpr) (*types.ExprAttribute, error) {
	glog.Infof("Processing SelectorExpr: %#v at %v\n", expr, expr.Pos())
	// Check for data type method cases
	// (*Receiver).method: use method of a data type as a value to store to a variable
	// (Receiver).method: the same, just the receiver is a data type itself
	hasMethod, function, err := ep.checkAngGetDataTypeMethod(expr)
	if err != nil {
		return nil, err
	}
	if hasMethod {
		return function, nil
	}
	// X.Sel a.k.a Prefix.Item
	xDefAttr, xErr := ep.Parse(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	if len(xDefAttr.DataTypeList) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}
	byteSlice, _ := json.Marshal(xDefAttr.DataTypeList[0])
	glog.Infof("\n\nSelectorExpr.X:\n\t%#v\n\tTypeVar: %#v\n\tfield:%#v\n\t%v at %v\n", xDefAttr.DataTypeList[0], xDefAttr.TypeVarList[0], expr.Sel, string(byteSlice), expr.Pos())

	fieldAttribute, err := propagation.New(ep.Config.SymbolsAccessor).SelectorExpr(xDefAttr.DataTypeList[0], expr.Sel.Name)
	if err != nil {
		return nil, err
	}

	// qid.id?
	if qid, ok := xDefAttr.DataTypeList[0].(*gotypes.Packagequalifier); ok {
		ep.AllocatedSymbolsTable.AddVariable(qid.Path, expr.Sel.Name, fmt.Sprintf("%v.%v", ep.Config.FileName, expr.Pos()))
		return types.ExprAttributeFromDataType(fieldAttribute.DataType).AddTypeVar(
			typevars.MakeVar(qid.Path, expr.Sel.Name, fmt.Sprintf("%v.%v", ep.Config.FileName, expr.Pos())),
		), nil
	}

	// Make sure the Field.X is always a virtual variable.
	// It is easier to evaluate during the data type propagation analysis.
	yVar := ep.typevar2variable(xDefAttr.TypeVarList[0])

	ep.Config.ContractTable.AddContract(&contracts.HasField{
		X:     yVar,
		Field: expr.Sel.Name,
		Pos:   fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Sel.Pos()),
	})

	return types.ExprAttributeFromDataType(fieldAttribute.DataType).AddTypeVar(
		typevars.MakeField(yVar, expr.Sel.Name, 0, fmt.Sprintf("%v:%v", ep.Config.FileName, expr.Sel.Pos())),
	), nil
}

func (ep *Parser) typevar2variable(xTypeVar typevars.Interface) *typevars.Variable {
	if d, ok := xTypeVar.(*typevars.Variable); ok {
		return d
	}
	yVarType := ep.Config.ContractTable.NewVirtualVar()
	ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
		X: xTypeVar,
		Y: yVarType,
	})
	return yVarType
}

// parseIndexExpr consumes ast.IndexExpr and produces:
// - if the expression is a map, data type of the map values is returned
// - if the expression is an array, data type of the array value is returned
// - if the expression is a slice, data type of the slice value is returned
// - if the expression as a string, uint8 data type is returned
// - if the expression as an ellipsis, data type of the ellipsis value is returned
func (ep *Parser) parseIndexExpr(expr *ast.IndexExpr) (*types.ExprAttribute, error) {
	glog.Infof("Processing IndexExpr: %#v\n", expr)
	// X[Index]
	// The Index can be a simple literal or another compound expression
	indexAttr, indexErr := ep.Parse(expr.Index)
	if indexErr != nil {
		return nil, indexErr
	}

	xDefAttr, xErr := ep.Parse(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	if len(xDefAttr.DataTypeList) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}

	indexExpr, xDefType, err := propagation.New(ep.Config.SymbolsAccessor).IndexExpr(xDefAttr.DataTypeList[0], indexAttr.DataTypeList[0])
	if err != nil {
		return nil, err
	}

	yVarType := ep.typevar2variable(xDefAttr.TypeVarList[0])
	ep.Config.ContractTable.AddContract(&contracts.IsIndexable{
		X:   yVarType,
		Key: indexAttr.TypeVarList[0],
		Pos: fmt.Sprintf("(2)%v:%v", ep.Config.FileName, expr.Pos()),
	})

	// Get definition of the X from the symbol Table (it must be a variable of a data type)
	// and get data type of its array/map members
	switch xDefType {
	case gotypes.MapType:
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: indexAttr.TypeVarList[0],
			Y: typevars.MakeMapKey(ep.typevar2variable(indexAttr.TypeVarList[0])),
		})
		return types.ExprAttributeFromDataType(indexExpr).AddTypeVar(
			typevars.MakeMapValue(yVarType),
		), nil
	case gotypes.ArrayType, gotypes.SliceType, gotypes.BuiltinType, gotypes.EllipsisType:
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: indexAttr.TypeVarList[0],
			Y: typevars.MakeListKey(),
		})
		return types.ExprAttributeFromDataType(indexExpr).AddTypeVar(
			typevars.MakeListValue(yVarType),
		), nil
	default:
		panic(fmt.Errorf("Unrecognized indexExpr type: %#v at %v", xDefAttr.DataTypeList[0], expr.Pos()))
	}
}

// parseTypeAssertExpr consumes ast.TypeAssertExpr and produces asserted data type
func (ep *Parser) parseTypeAssertExpr(expr *ast.TypeAssertExpr) (*types.ExprAttribute, error) {
	glog.Infof("Processing TypeAssertExpr: %#v\n", expr)
	// X.(Type)
	attr, xErr := ep.Parse(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	// We should check if the data type really implements all methods of the interface.
	// Or we can assume it does and just return the Type itself
	// TODO(jchaloup): check the data type Type really implements interface of X (if it is an interface)
	// TODO(jchaloup): generate a data type contract
	isDT, err := ep.isDataType(expr.Type)
	if err != nil {
		return nil, err
	}
	if !isDT {
		return nil, fmt.Errorf("TypeAssertExpression is expected to be a data type, got %#v instead", expr.Type)
	}

	def, err := ep.TypeParser.Parse(expr.Type)
	if err != nil {
		return nil, err
	}

	y := typevars.MakeConstant(ep.Config.PackageName, def)

	ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
		X:            attr.TypeVarList[0],
		Y:            y,
		ExpectedType: attr.DataTypeList[0],
	})

	return types.ExprAttributeFromDataType(def).AddTypeVar(y), nil
}

func (ep *Parser) parseEllipsis(expr *ast.Ellipsis) (*types.ExprAttribute, error) {
	glog.Infof("Processing Ellipsis: %#v\n", expr)
	if expr.Elt == nil {
		return types.ExprAttributeFromDataType(&gotypes.Ellipsis{}), nil
	}
	def, err := ep.Parse(expr.Elt)
	if err != nil {
		return nil, err
	}
	if len(def.DataTypeList) != 1 {
		return nil, fmt.Errorf("Expected a single expression of ellipses, got %#v instead", def.DataTypeList)
	}
	return types.ExprAttributeFromDataType(&gotypes.Ellipsis{
		Def: def.DataTypeList[0],
	}), nil
}

func (ep *Parser) Parse(expr ast.Expr) (*types.ExprAttribute, error) {
	// Given an expression we must always return its final data type
	// User defined symbols has its corresponding structs under parser/pkg/types.
	// In order to cover all possible symbol data types, we need to cover
	// golang language embedded data types as well
	switch exprType := expr.(type) {
	// Basic literal carries
	case *ast.BasicLit:
		return ep.parseBasicLit(exprType)
	case *ast.CompositeLit:
		return ep.parseCompositeLit(exprType, nil)
	case *ast.Ident:
		return ep.parseIdentifier(exprType)
	case *ast.StarExpr:
		return ep.parseStarExpr(exprType)
	case *ast.UnaryExpr:
		return ep.parseUnaryExpr(exprType)
	case *ast.BinaryExpr:
		return ep.parseBinaryExpr(exprType)
	case *ast.CallExpr:
		// If the call expression is the most most expression,
		// it may have a different number of results
		return ep.parseCallExpr(exprType)
	case *ast.StructType:
		return ep.parseStructType(exprType)
	case *ast.IndexExpr:
		return ep.parseIndexExpr(exprType)
	case *ast.SelectorExpr:
		return ep.parseSelectorExpr(exprType)
	case *ast.TypeAssertExpr:
		return ep.parseTypeAssertExpr(exprType)
	case *ast.FuncLit:
		return ep.parseFuncLit(exprType)
	case *ast.SliceExpr:
		return ep.parseSliceExpr(exprType)
	case *ast.ParenExpr:
		return ep.Parse(ep.parseParenExpr(exprType))
	case *ast.ChanType:
		return ep.parseChanType(exprType)
	case *ast.Ellipsis:
		return ep.parseEllipsis(exprType)
	default:
		return nil, fmt.Errorf("Unrecognized expression: %#v\n", expr)
	}
}

func New(c *types.Config) types.ExpressionParser {
	return &Parser{
		Config: c,
	}
}
