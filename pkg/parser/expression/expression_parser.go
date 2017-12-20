package expression

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
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
	var builtin *gotypes.Builtin
	switch lit.Kind {
	case token.INT:
		builtin = &gotypes.Builtin{Def: "int", Untyped: true}
	case token.FLOAT:
		builtin = &gotypes.Builtin{Def: "float", Untyped: true}
	case token.IMAG:
		builtin = &gotypes.Builtin{Def: "imag", Untyped: true}
	case token.STRING:
		builtin = &gotypes.Builtin{Def: "string", Untyped: true}
	case token.CHAR:
		builtin = &gotypes.Builtin{Def: "char", Untyped: true}
	default:
		return nil, fmt.Errorf("Unrecognize BasicLit: %#v\n", lit.Kind)
	}

	return types.ExprAttributeFromDataType(builtin).AddTypeVar(&typevars.Constant{
		DataType: builtin,
	}), nil
}

// to parse *gotypes.Array, *gotypes.Slice, *gotypes.Map
func (ep *Parser) parseKeyValueLikeExpr(litDataType gotypes.DataType, lit *ast.CompositeLit, litType gotypes.DataType) (typevars.Interface, error) {
	glog.Infof("Processing parseKeyValueLikeExpr")
	var valueTypeVar typevars.Interface
	var keyTypeVar typevars.Interface
	var valueType gotypes.DataType

	var outputTypeVar typevars.Interface = ep.Config.ContractTable.NewVirtualVar()

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
			ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X:            valueTypeVar,
				Y:            attr.TypeVarList[0],
				ExpectedType: attr.DataTypeList[0],
			})
			continue
		}

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
		outputOriginalTypeVar = typevars.MakeConstant(litDataType)
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
func (ep *Parser) parseCompositeLitStructElements(litDataType gotypes.DataType, lit *ast.CompositeLit, structDef *gotypes.Struct, structSymbol *symboltable.SymbolDef) (typevars.Interface, error) {
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
			if structSymbol != nil {
				ep.AllocatedSymbolsTable.AddDataTypeField(structSymbol.Package, structSymbol.Name, keyDefIdentifier.Name)
			}
			ep.Config.ContractTable.AddContract(&contracts.HasField{
				X:     structOutputTypeVar,
				Field: keyDefIdentifier.Name,
			})
			fieldTypeVar = typevars.MakeField(structOutputTypeVar, keyDefIdentifier.Name, 0)
			valueExpr = kvExpr.Value
		} else {
			if fieldCounter >= fieldLen {
				return nil, fmt.Errorf("Number of fields of the CL is greater than the number of fields of underlying struct %#v", structDef)
			}
			if structSymbol != nil {
				// TODO(jchaloup): Should we count the anonymous field as well? Maybe make a AddDataTypeAnonymousField?
				ep.AllocatedSymbolsTable.AddDataTypeField(structSymbol.Package, structSymbol.Name, structDef.Fields[fieldCounter].Name)
			}
			ep.Config.ContractTable.AddContract(&contracts.HasField{
				X:     structOutputTypeVar,
				Index: fieldCounter,
			})
			fieldTypeVar = typevars.MakeField(structOutputTypeVar, "", fieldCounter)
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
		structOriginalTypeVar = typevars.MakeConstant(litDataType)
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

func (ep *Parser) findFirstNonidDataType(typeDef gotypes.DataType) (*types.ExprAttribute, error) {
	var symbolDef *symboltable.SymbolDef
	switch typeDefType := typeDef.(type) {
	case *gotypes.Selector:
		_, def, err := ep.retrieveQidDataType(typeDefType.Prefix, &ast.Ident{Name: typeDefType.Item})
		if err != nil {
			return nil, err
		}
		symbolDef = def
	case *gotypes.Identifier:
		def, _, err := ep.Config.Lookup(typeDefType)
		if err != nil {
			return nil, err
		}
		symbolDef = def
	default:
		return types.ExprAttributeFromDataType(typeDef), nil
	}
	if symbolDef.Def == nil {
		return nil, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
	}
	return ep.findFirstNonidDataType(symbolDef.Def)
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
	nonIdentLitTypeDef, err := ep.findFirstNonidDataType(litTypedef)
	if err != nil {
		return nil, err
	}
	glog.Infof("nonIdentLitTypeDef: %#v", nonIdentLitTypeDef)

	// If the CL type is anonymous struct, array, slice or map don't store fields into the allocated symbols table (AST)
	var typeVar typevars.Interface
	switch litTypeExpr := nonIdentLitTypeDef.DataTypeList[0].(type) {
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
		typeVar, err = ep.parseKeyValueLikeExpr(origLitTypedef, lit, nonIdentLitTypeDef.DataTypeList[0])
		if err != nil {
			return nil, err
		}
	default:
		panic(fmt.Errorf("Unsupported CL type: %#v", nonIdentLitTypeDef.DataTypeList[0]))
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
			// The data type of the variable is not accounted as it is not implicitely used
			// The variable itself carries the data type and as long as the variable does not
			// get used, the data type can change.
			// TODO(jchaloup): return symbol origin
			if st == symboltable.FunctionSymbol {
				return types.ExprAttributeFromDataType(def.Def).AddTypeVar(
					typevars.FunctionFromSymbolDef(def),
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
		case symboltable.VariableSymbol:
			switch ident.Name {
			case "true", "false":
				return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: "bool"}).AddTypeVar(
					typevars.MakeConstant(&gotypes.Builtin{Def: "bool"}),
				), nil
			case "nil":
				return types.ExprAttributeFromDataType(&gotypes.Nil{}).AddTypeVar(
					typevars.MakeConstant(&gotypes.Nil{}),
				), nil
			case "iota":
				return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: "iota"}).AddTypeVar(
					typevars.MakeConstant(&gotypes.Builtin{Def: "iota"}),
				), nil
			default:
				return nil, fmt.Errorf("Unsupported built-in type: %v", ident.Name)
			}
		case symboltable.FunctionSymbol:
			return types.ExprAttributeFromDataType(symbolDef.Def).AddTypeVar(
				typevars.FunctionFromSymbolDef(symbolDef),
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

	y := ep.Config.ContractTable.NewVirtualVar()

	var yAttr *types.ExprAttribute

	// TODO(jchaloup): check the token is really a unary operator
	// TODO(jchaloup): add the missing unary operator tokens
	switch expr.Op {
	// variable address
	case token.AND:
		ep.Config.ContractTable.AddContract(&contracts.UnaryOp{
			OpToken: expr.Op,
			X:       attr.TypeVarList[0],
			Y:       y,
		})
		yAttr = types.ExprAttributeFromDataType(&gotypes.Pointer{Def: attr.DataTypeList[0]})
	// channel
	case token.ARROW:
		nonIdentDefAttr, err := ep.findFirstNonidDataType(attr.DataTypeList[0])
		if err != nil {
			return nil, err
		}
		if nonIdentDefAttr.DataTypeList[0].GetType() != gotypes.ChannelType {
			return nil, fmt.Errorf("<-OP operator expectes OP to be a channel, got %v instead at %v", nonIdentDefAttr.DataTypeList[0].GetType(), expr.Pos())
		}
		ep.Config.ContractTable.AddContract(&contracts.IsReceiveableFrom{
			X: attr.TypeVarList[0],
			Y: y,
		})
		yAttr = types.ExprAttributeFromDataType(nonIdentDefAttr.DataTypeList[0].(*gotypes.Channel).Value)
		// other
	case token.XOR, token.OR, token.SUB, token.NOT, token.ADD:
		// is token.OR really a unary operator?
		ep.Config.ContractTable.AddContract(&contracts.UnaryOp{
			OpToken: expr.Op,
			X:       attr.TypeVarList[0],
			Y:       y,
		})
		yAttr = types.ExprAttributeFromDataType(attr.DataTypeList[0])
	default:
		return nil, fmt.Errorf("Unary operator %#v (%#v) not recognized", expr.Op, token.ADD)
	}
	return yAttr.AddTypeVar(y), nil
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

	switch expr.Op {
	case token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR:
		// We should check both operands are compatible with the operator.
		// However, it requires a hard-coded knowledge of all available data types.
		// I.e. we must know int, int8, int16, int32, etc. types exists and if one
		// of the operators is untyped integral constant, the operator is valid.
		// As the list of all data types is read from the builtin package,
		// it is not possible to keep a clean solution and provide
		// the check for the operands validity at the same time.
		z := ep.Config.ContractTable.NewVirtualVar()
		ep.Config.ContractTable.AddContract(&contracts.BinaryOp{
			OpToken: expr.Op,
			X:       xAttr.TypeVarList[0],
			Y:       yAttr.TypeVarList[0],
			Z:       z,
		})
		return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: "bool"}).AddTypeVar(z), nil
	}

	// If both types are built-in, just return built-in
	if xAttr.DataTypeList[0].GetType() == yAttr.DataTypeList[0].GetType() && xAttr.DataTypeList[0].GetType() == gotypes.BuiltinType {
		xt := xAttr.DataTypeList[0].(*gotypes.Builtin)
		yt := yAttr.DataTypeList[0].(*gotypes.Builtin)
		if xt.Def != yt.Def {
			switch expr.Op {
			case token.SHL, token.SHR:
				// The right operand in a shift expression must have unsigned integer type
				// or be an untyped constant that can be converted to unsigned integer type.
				// If the left operand of a non-constant shift expression is an untyped constant,
				// it is first converted to the type it would assume if the shift expression
				// were replaced by its left operand alone.

				// The right operand can be ignored (it is always converted to an integer).
				// If the left operand is untyped int => untyped int (or int for non-constant expressions)
				// if the left operand is typed int => return int
				z := ep.Config.ContractTable.NewVirtualVar()
				ep.Config.ContractTable.AddContract(&contracts.BinaryOp{
					OpToken: expr.Op,
					X:       xAttr.TypeVarList[0],
					Y:       yAttr.TypeVarList[0],
					Z:       z,
				})
				if xt.Untyped {
					// const a = 8.0 << 1
					// a is of type int, checked at https://play.golang.org/
					return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: "int", Untyped: true}).AddTypeVar(z), nil
				}
				return types.ExprAttributeFromDataType(xt).AddTypeVar(z), nil
			case token.AND, token.OR, token.MUL, token.SUB, token.QUO, token.ADD, token.AND_NOT, token.REM, token.XOR:
				// The same reasoning as with the relational operators
				// Experiments:
				// 1+1.0 is untyped float64
				if xt.Untyped && yt.Untyped {

				}
				z := ep.Config.ContractTable.NewVirtualVar()
				ep.Config.ContractTable.AddContract(&contracts.BinaryOp{
					OpToken: expr.Op,
					X:       xAttr.TypeVarList[0],
					Y:       yAttr.TypeVarList[0],
					Z:       z,
				})
				if xt.Untyped {
					return types.ExprAttributeFromDataType(yt).AddTypeVar(z), nil
				}
				if yt.Untyped {
					return types.ExprAttributeFromDataType(xt).AddTypeVar(z), nil
				}
				// byte(2)&uint8(3) is uint8
				if (xt.Def == "byte" && yt.Def == "uint8") || (yt.Def == "byte" && xt.Def == "uint8") {
					return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: "uint8"}).AddTypeVar(z), nil
				}
			}
			return nil, fmt.Errorf("Binary operation %q over two different built-in types: %q != %q", expr.Op, xt, yt)
		}
		z := ep.Config.ContractTable.NewVirtualVar()
		ep.Config.ContractTable.AddContract(&contracts.BinaryOp{
			OpToken: expr.Op,
			X:       xAttr.TypeVarList[0],
			Y:       yAttr.TypeVarList[0],
			Z:       z,
		})
		if xt.Untyped {
			{
				byteSlice, _ := json.Marshal(&gotypes.Builtin{Def: xt.Def, Untyped: yt.Untyped})
				glog.Infof("xx+yy: %v\n", string(byteSlice))
			}
			return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: xt.Def, Untyped: yt.Untyped}).AddTypeVar(z), nil
		}
		if yt.Untyped {
			{
				byteSlice, _ := json.Marshal(&gotypes.Builtin{Def: xt.Def, Untyped: xt.Untyped})
				glog.Infof("xx+yy: %v\n", string(byteSlice))
			}
			return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: xt.Def, Untyped: xt.Untyped}).AddTypeVar(z), nil
		}
		// both operands are typed => Untyped = false
		{
			byteSlice, _ := json.Marshal(&gotypes.Builtin{Def: xt.Def})
			glog.Infof("xx+yy: %v\n", string(byteSlice))
		}
		return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: xt.Def}).AddTypeVar(z), nil
	}

	glog.Infof("Binaryexpr.x: %#v\nBinaryexpr.y: %#v\n", xAttr.DataTypeList[0], yAttr.DataTypeList[0])

	var xIdent *gotypes.Identifier
	var xIsIdentifier bool
	switch xExpr := xAttr.DataTypeList[0].(type) {
	case *gotypes.Identifier:
		xIdent = xExpr
		xIsIdentifier = true
	case *gotypes.Selector:
		qid, ok := xExpr.Prefix.(*gotypes.Packagequalifier)
		if !ok {
			return nil, fmt.Errorf("Left operand of a binary expression is expected to be a qualified identifier, got %#v instead", xExpr)
		}
		xIdent = &gotypes.Identifier{
			Def:     xExpr.Item,
			Package: qid.Name,
		}
		xIsIdentifier = true
	}

	var yIdent *gotypes.Identifier
	var yIsIdentifier bool
	switch yExpr := yAttr.DataTypeList[0].(type) {
	case *gotypes.Identifier:
		yIdent = yExpr
		yIsIdentifier = true
	case *gotypes.Selector:
		qid, ok := yExpr.Prefix.(*gotypes.Packagequalifier)
		if !ok {
			return nil, fmt.Errorf("Right operand of a binary expression is expected to be a qualified identifier, got %#v instead", yExpr)
		}
		yIdent = &gotypes.Identifier{
			Def:     yExpr.Item,
			Package: qid.Name,
		}
		yIsIdentifier = true
	}

	if !xIsIdentifier && !yIsIdentifier {
		return nil, fmt.Errorf("At least one operand of a binary expression has to be an identifier or a qualified identifier")
	}

	z := ep.Config.ContractTable.NewVirtualVar()
	ep.Config.ContractTable.AddContract(&contracts.BinaryOp{
		OpToken: expr.Op,
		X:       xAttr.TypeVarList[0],
		Y:       yAttr.TypeVarList[0],
		Z:       z,
	})

	// Even here we assume existence of untyped constants.
	// If there is only one identifier it means the other operand is a built-in type.
	// Assuming the code is written correctly (it compiles), resulting data type of the operation
	// is always the identifier.
	// TODO(jchaloup): if an operand is a data type identifier, we need to return the data type itself, not its definition
	if xIsIdentifier {
		return types.ExprAttributeFromDataType(xIdent).AddTypeVar(z), nil
	}
	return types.ExprAttributeFromDataType(yIdent).AddTypeVar(z), nil
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
		var def *symboltable.SymbolDef
		var defType symboltable.SymbolType
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
		_, sd, err := ep.retrieveQidDataType(typeDef.Prefix, &ast.Ident{Name: typeDef.Item})
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
	processArgs := func(functionTypeVar *typevars.Function, args []ast.Expr, params []gotypes.DataType) error {
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

		y := typevars.MakeConstant(def)

		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X:            attr.TypeVarList[0],
			Y:            y,
			ExpectedType: attr.DataTypeList[0],
		})

		return types.ExprAttributeFromDataType(def).AddTypeVar(y), nil
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
				switch arglen := len(expr.Args); arglen {
				case 1:
					glog.Infof("Processing make arguments for make(type) type: %#v", expr.Args)
				case 2:
					glog.Infof("Processing make arguments for make(type, size) type: %#v", expr.Args)
					if err := processArgs(typevars.MakeFunction("builtin", "make", ""), []ast.Expr{expr.Args[1]}, nil); err != nil {
						return nil, err
					}
				case 3:
					glog.Infof("Processing make arguments for make(type, size, size) type: %#v", expr.Args)
					if err := processArgs(typevars.MakeFunction("builtin", "make", ""), []ast.Expr{expr.Args[1], expr.Args[2]}, nil); err != nil {
						return nil, err
					}
				default:
					return nil, fmt.Errorf("Expecting 1, 2 or 3 arguments of built-in function make, got %q instead", arglen)
				}
				typeDef, err := ep.TypeParser.Parse(expr.Args[0])
				return types.ExprAttributeFromDataType(typeDef).AddTypeVar(typevars.MakeConstant(typeDef)), err
			case "new":
				// The new built-in function allocates memory. The first argument is a type,
				// not a value, and the value returned is a pointer to a newly
				// allocated zero value of that type.
				if len(expr.Args) != 1 {
					return nil, fmt.Errorf("Len of new args != 1, it is %#v", expr.Args)
				}
				typeDef, err := ep.TypeParser.Parse(expr.Args[0])
				return types.ExprAttributeFromDataType(
					&gotypes.Pointer{Def: typeDef},
				).AddTypeVar(
					typevars.MakeConstant(&gotypes.Pointer{Def: typeDef}),
				), err
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

	var f *typevars.Function
	switch d := attr.TypeVarList[0].(type) {
	case *typevars.Function:
		f = d
	case *typevars.Variable:
		f = typevars.MakeFunction(d.Package, d.Name, d.Pos)
	case *typevars.Constant, *typevars.Field, *typevars.ReturnType, *typevars.ListValue:
		newVar := ep.Config.ContractTable.NewVirtualVar()
		ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
			X: attr.TypeVarList[0],
			Y: newVar,
		})
		f = typevars.MakeVirtualFunction(newVar)
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
		X: exprDefAttr.TypeVarList[0],
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
		X: typevars.MakeConstant(def),
		Y: newVar,
	})

	return types.ExprAttributeFromDataType(def).AddTypeVar(
		typevars.MakeVirtualFunction(newVar),
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

func (ep *Parser) retrieveQidDataType(qidprefix gotypes.DataType, item *ast.Ident) (symboltable.SymbolLookable, *symboltable.SymbolDef, error) {
	// qid.structtype expected
	qid, ok := qidprefix.(*gotypes.Packagequalifier)
	if !ok {
		return nil, nil, fmt.Errorf("Expecting a qid.structtype when retrieving a struct from a selector expression")
	}
	glog.Infof("Trying to retrieve a symbol %#v from package %v\n", item.String(), qid.Path)
	qidst, err := ep.GlobalSymbolTable.Lookup(qid.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid.Path, err)
	}

	structDef, piErr := qidst.LookupDataType(item.String())
	if piErr != nil {
		return nil, nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", item.String(), qid.Path, piErr)
	}
	ep.AllocatedSymbolsTable.AddDataType(qid.Path, item.Name, ep.Config.SymbolPos(item.Pos()))
	return qidst, structDef, nil
}

type dataTypeFieldAccessor struct {
	symbolTable symboltable.SymbolLookable
	dataTypeDef *symboltable.SymbolDef
	field       *ast.Ident
	methodsOnly bool
	fieldsOnly  bool
	// if set to true, lookup through fields only (no methods)
	// when processing embedded data types, set the flag back to false
	// This is usefull when a data type is defined as data type of another struct.
	// In that case, methods of the other struct are not inherited.
	dropFieldsOnly bool
}

// Get a struct's field.
// Given a struct can embedded another struct from a different package, the method must be able to Accessing
// symbol tables of other packages. Thus recursively process struct's definition up to all its embedded fields.
func (ep *Parser) retrieveDataTypeField(accessor *dataTypeFieldAccessor) (*types.ExprAttribute, error) {
	glog.Infof("Retrieving data type field %q from %#v\n", accessor.field.String(), accessor.dataTypeDef)
	// Only data type declaration is known
	if accessor.dataTypeDef.Def == nil {
		return nil, fmt.Errorf("Data type definition of %q is not known", accessor.dataTypeDef.Name)
	}

	// Any data type can have its own methods.
	// Or, a struct can embedded any data type with its own methods
	if accessor.dataTypeDef.Def.GetType() != gotypes.StructType {
		if accessor.dataTypeDef.Def.GetType() == gotypes.IdentifierType {
			ident := accessor.dataTypeDef.Def.(*gotypes.Identifier)
			// check methods of the type itself if there is any match
			glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return types.ExprAttributeFromDataType(method.Def), nil
			}

			if ident.Package == "builtin" {
				// built-in => only methods
				// Check data type methods
				glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
				if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
					return types.ExprAttributeFromDataType(method.Def), nil
				}
				return nil, fmt.Errorf("Unable to find a field %v in data type %#v", accessor.field.String(), accessor.dataTypeDef)
			}

			// if not built-in, it is an identifier of another type,
			// in which case all methods of the type are ignored
			// I.e. with 'type A B' data type of B is data type of A. However, methods of type B are not inherited.

			var pkgST symboltable.SymbolLookable
			if ident.Package == "" || ident.Package == ep.PackageName {
				pkgST = ep.SymbolTable
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
			return ep.retrieveDataTypeField(&dataTypeFieldAccessor{
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
				return types.ExprAttributeFromDataType(method.Def), nil
			}

			selector := accessor.dataTypeDef.Def.(*gotypes.Selector)
			// qid?
			st, sd, err := ep.Config.RetrieveQidDataType(selector)
			if err != nil {
				return nil, err
			}

			return ep.retrieveDataTypeField(&dataTypeFieldAccessor{
				symbolTable:    st,
				dataTypeDef:    sd,
				field:          accessor.field,
				fieldsOnly:     true,
				dropFieldsOnly: true,
			})
		}

		if !accessor.fieldsOnly {
			if accessor.dataTypeDef.Def.GetType() == gotypes.InterfaceType {
				return ep.retrieveInterfaceMethod(accessor.symbolTable, accessor.dataTypeDef, accessor.field.String())
			}
			// Check data type methods
			glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
			if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
				return types.ExprAttributeFromDataType(method.Def), nil
			}
		}
		return nil, fmt.Errorf("Unable to find a field %v in data type %#v", accessor.field.String(), accessor.dataTypeDef)
	}

	type embeddedDataTypesItem struct {
		symbolTable symboltable.SymbolLookable
		symbolDef   *symboltable.SymbolDef
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
				table, err := ep.GlobalSymbolTable.Lookup("builtin")
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
				st, sd, err := ep.retrieveQidDataType(fieldType.Prefix, &ast.Ident{Name: fieldType.Item})
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
		if accessor.dataTypeDef.Name != "" {
			ep.AllocatedSymbolsTable.AddDataTypeField(accessor.dataTypeDef.Package, accessor.dataTypeDef.Name, accessor.field.String())
		}
		return types.ExprAttributeFromDataType(fieldItem.Def), nil
	}

	// First, check methods, then embedded structs

	// Check data type methods
	if !accessor.fieldsOnly {
		glog.Infof("Retrieving method %q of data type %q", accessor.field.String(), accessor.dataTypeDef.Name)
		if method, err := accessor.symbolTable.LookupMethod(accessor.dataTypeDef.Name, accessor.field.String()); err == nil {
			return types.ExprAttributeFromDataType(method.Def), nil
		}
	}

	glog.Info("Retrieving fields from embedded structs")
	if len(embeddedDataTypes) != 0 {
		if accessor.dropFieldsOnly {
			accessor.fieldsOnly = false
		}
		for _, item := range embeddedDataTypes {
			if fieldDefAttr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
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

func (ep *Parser) retrieveInterfaceMethod(pkgsymboltable symboltable.SymbolLookable, interfaceDefsymbol *symboltable.SymbolDef, method string) (*types.ExprAttribute, error) {
	glog.Infof("Retrieving Interface method %q from %#v\n", method, interfaceDefsymbol)
	if interfaceDefsymbol.Def.GetType() != gotypes.InterfaceType {
		return nil, fmt.Errorf("Trying to retrieve a %v method from a non-interface data type: %#v", method, interfaceDefsymbol.Def)
	}

	type embeddedInterfacesItem struct {
		symbolTable symboltable.SymbolLookable
		symbolDef   *symboltable.SymbolDef
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
				table, err := ep.GlobalSymbolTable.Lookup("builtin")
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
				st, sd, err := ep.retrieveQidDataType(fieldType.Prefix, &ast.Ident{Name: fieldType.Item})
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
			if fieldDef, err := ep.retrieveInterfaceMethod(item.symbolTable, item.symbolDef, method); err == nil {
				return fieldDef, nil
			}
		}
	}

	if methodItem != nil {
		if interfaceDefsymbol.Name != "" {
			ep.AllocatedSymbolsTable.AddDataTypeField(interfaceDefsymbol.Package, interfaceDefsymbol.Name, method)
		}
		return types.ExprAttributeFromDataType(methodItem.Def), nil
	}

	return nil, fmt.Errorf("Unable to find a method %v in interface %#v", method, interfaceDefsymbol)
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

	var typeSymbolTable symboltable.SymbolLookable
	var typeSymbolDef *symboltable.SymbolDef

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
			return false, nil, fmt.Errorf("Expecting qid.id, got %#v for the qid instead", dtExpr.X)
		}

		qidAttr, err := ep.parseIdentifier(ident)
		if err != nil {
			return false, nil, err
		}
		if _, isQid := qidAttr.DataTypeList[0].(*gotypes.Packagequalifier); !isQid {
			return false, nil, fmt.Errorf("Expecting qid.id, got %#v for the qid instead", qidAttr.DataTypeList[0])
		}

		st, sd, err := ep.retrieveQidDataType(qidAttr.DataTypeList[0], &ast.Ident{Name: dtExpr.Sel.String()})
		if err != nil {
			return false, nil, err
		}

		typeSymbolTable = st
		typeSymbolDef = sd
	default:
		return false, nil, nil
	}

	methodDefAttr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
		symbolTable: typeSymbolTable,
		dataTypeDef: typeSymbolDef,
		field:       &ast.Ident{Name: expr.Sel.Name},
		methodsOnly: true,
	})
	if err != nil {
		return false, nil, fmt.Errorf("Unable to find method %q of data type %v", expr.Sel.Name, typeSymbolDef.Name)
	}
	method, ok := methodDefAttr.DataTypeList[0].(*gotypes.Method)
	if !ok {
		return false, nil, fmt.Errorf("Expected a method %q of data type %v, got %#v instead", expr.Sel.Name, typeSymbolDef.Name, methodDefAttr.DataTypeList[0])
	}

	y := ep.Config.ContractTable.NewVirtualVar()

	ep.Config.ContractTable.AddContract(&contracts.PropagatesTo{
		X: &typevars.Constant{
			DataType: typeSymbolDef.Def,
			Package:  typeSymbolDef.Package,
		},
		Y: y,
	})

	ep.Config.ContractTable.AddContract(&contracts.HasField{
		X:     y,
		Field: expr.Sel.Name,
	})

	return true, types.ExprAttributeFromDataType(method.Def.(*gotypes.Function)).AddTypeVar(
		typevars.MakeField(y, expr.Sel.Name, 0),
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

	var outputAttr *types.ExprAttribute
	var outputField string

	// The struct and an interface are the only data type from which a field/method is retriveable
	switch xType := xDefAttr.DataTypeList[0].(type) {
	// If the X expression is a qualified id, the selector is a symbol from a package pointed by the id
	case *gotypes.Packagequalifier:
		glog.Infof("Trying to retrieve a symbol %#v from package %v\n", expr.Sel.Name, xType.Path)
		st, err := ep.GlobalSymbolTable.Lookup(xType.Path)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", xType, err)
		}

		packageIdent, _, piErr := st.Lookup(expr.Sel.Name)
		if piErr != nil {
			return nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", expr.Sel.Name, xType, piErr)
		}
		byteSlice, _ := json.Marshal(packageIdent)
		glog.Infof("packageIdent: %v\n", string(byteSlice))
		// TODO(jchaloup): check if the packageIdent is a global variable, method, function or a data type
		//                 based on that, use the corresponding Add method
		ep.AllocatedSymbolsTable.AddSymbol(xType.Path, expr.Sel.Name, "")
		return types.ExprAttributeFromDataType(packageIdent.Def).AddTypeVar(
			typevars.MakeVar(
				xType.Path,
				expr.Sel.Name,
				"", // no position for a global symbol
			),
		), nil
	case *gotypes.Pointer:
		switch def := xType.Def.(type) {
		case *gotypes.Identifier:
			// Get struct's definition given by its identifier
			{
				byteSlice, _ := json.Marshal(def)
				glog.Infof("Looking for: %v\n", string(byteSlice))
			}
			structDefsymbol, symbolTable, err := ep.Config.LookupDataType(def)
			if err != nil {
				return nil, fmt.Errorf("Cannot retrieve identifier %q from the symbol table: %v", def.Def, err)
			}
			{
				byteSlice, _ := json.Marshal(structDefsymbol)
				glog.Infof("Struct retrieved: %v\n", string(byteSlice))
			}

			attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
				symbolTable: symbolTable,
				dataTypeDef: structDefsymbol,
				field:       &ast.Ident{Name: expr.Sel.Name},
			})

			if err != nil {
				return nil, err
			}
			outputAttr = attr
			outputField = expr.Sel.Name
		case *gotypes.Selector:
			// qid to different package
			// TODO(jchaloup): create a contract once the multi-package contracts collecting is implemented
			pq, ok := def.Prefix.(*gotypes.Packagequalifier)
			if !ok {
				return nil, fmt.Errorf("Trying to retrieve a %v field from a pointer to non-qualified struct data type: %#v", expr.Sel.Name, def)
			}
			structDefsymbol, symbolTable, err := ep.Config.LookupDataType(&gotypes.Identifier{Def: def.Item, Package: pq.Path})
			if err != nil {
				return nil, err
			}
			attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
				symbolTable: symbolTable,
				dataTypeDef: structDefsymbol,
				field:       &ast.Ident{Name: expr.Sel.Name},
			})
			if err != nil {
				return nil, err
			}
			outputAttr = attr
			outputField = expr.Sel.Name
		case *gotypes.Struct:
			// anonymous struct
			attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
				symbolTable: ep.SymbolTable,
				dataTypeDef: &symboltable.SymbolDef{Def: def},
				field:       &ast.Ident{Name: expr.Sel.Name},
			})
			if err != nil {
				return nil, err
			}
			outputAttr = attr
			outputField = expr.Sel.Name
		default:
			return nil, fmt.Errorf("Trying to retrieve a %q field from a pointer to non-struct data type: %#v", expr.Sel.Name, xType.Def)
		}
	case *gotypes.Identifier:
		// Get struct/interface definition given by its identifier
		defSymbol, symbolTable, err := ep.Config.LookupDataType(xType)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", xType.Def)
		}
		if defSymbol.Def == nil {
			return nil, fmt.Errorf("Trying to retrieve a field/method from a data type %#v that is not yet fully processed", xType)
		}
		switch defSymbol.Def.(type) {
		case *gotypes.Struct:
			attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
				symbolTable: symbolTable,
				dataTypeDef: defSymbol,
				field:       &ast.Ident{Name: expr.Sel.Name},
			})
			if err != nil {
				return nil, err
			}
			outputAttr = attr
			outputField = expr.Sel.Name
		case *gotypes.Interface:
			attr, err := ep.retrieveInterfaceMethod(symbolTable, defSymbol, expr.Sel.Name)
			if err != nil {
				return attr, err
			}
			outputAttr = attr
			outputField = expr.Sel.Name
		case *gotypes.Identifier:
			attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
				symbolTable:    symbolTable,
				dataTypeDef:    defSymbol,
				field:          &ast.Ident{Name: expr.Sel.Name},
				fieldsOnly:     true,
				dropFieldsOnly: true,
			})
			if err != nil {
				return attr, err
			}
			outputAttr = attr
			outputField = expr.Sel.Name
		default:
			// check data types with receivers
			glog.Infof("Retrieving method %q of a non-struct non-interface data type %#v", expr.Sel.Name, xType)
			def, err := ep.Config.LookupMethod(xType, expr.Sel.Name)
			if err != nil {
				return nil, fmt.Errorf("Trying to retrieve a field/method from non-struct/non-interface data type: %#v at %v: %v", defSymbol, expr.Pos(), err)
			}
			outputAttr = types.ExprAttributeFromDataType(def.Def)
			outputField = expr.Sel.Name
		}
	// anonymous struct
	case *gotypes.Struct:
		attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
			symbolTable: ep.SymbolTable,
			dataTypeDef: &symboltable.SymbolDef{
				Name:    "",
				Package: "",
				Def:     xType,
			},
			field: &ast.Ident{Name: expr.Sel.Name},
		})
		if err != nil {
			return nil, err
		}
		outputAttr = attr
		outputField = expr.Sel.Name
	case *gotypes.Interface:
		// TODO(jchaloup): test the case when the interface is anonymous
		attr, err := ep.retrieveInterfaceMethod(ep.SymbolTable, &symboltable.SymbolDef{
			Name:    "",
			Package: "",
			Def:     xType,
		}, expr.Sel.Name)
		if err != nil {
			return nil, err
		}
		outputAttr = attr
		outputField = expr.Sel.Name
	case *gotypes.Selector:
		// qid.structtype expected
		st, sd, err := ep.retrieveQidDataType(xType.Prefix, &ast.Ident{Name: xType.Item})
		if err != nil {
			return nil, err
		}
		attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
			symbolTable: st,
			dataTypeDef: sd,
			field:       &ast.Ident{Name: expr.Sel.Name},
		})
		if err != nil {
			return nil, err
		}
		outputAttr = attr
		outputField = expr.Sel.Name
	case *gotypes.Builtin:
		// Check if the built-in type has some methods
		table, err := ep.GlobalSymbolTable.Lookup("builtin")
		if err != nil {
			return nil, err
		}
		def, err := table.LookupDataType(xType.Def)
		if err != nil {
			return nil, err
		}
		attr, err := ep.retrieveDataTypeField(&dataTypeFieldAccessor{
			symbolTable: table,
			dataTypeDef: def,
			field:       &ast.Ident{Name: expr.Sel.Name},
			methodsOnly: true,
		})
		if err != nil {
			return nil, err
		}
		outputAttr = attr
		outputField = expr.Sel.Name
	default:
		return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type when parsing selector expression %#v at %v", expr.Sel.Name, xDefAttr.DataTypeList[0], expr.Pos())
	}

	ep.Config.ContractTable.AddContract(&contracts.HasField{
		X:     xDefAttr.TypeVarList[0],
		Field: outputField,
	})

	return outputAttr.AddTypeVar(
		typevars.MakeField(xDefAttr.TypeVarList[0], outputField, 0),
	), nil
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

	// One can not do &(&(a))
	var indexExpr gotypes.DataType
	pointer, ok := xDefAttr.DataTypeList[0].(*gotypes.Pointer)
	if ok {
		indexExpr = pointer.Def
	} else {
		indexExpr = xDefAttr.DataTypeList[0]
	}
	glog.Infof("IndexExprDef: %#v", indexExpr)
	// In case we have
	// type A []int
	// type B A
	// type C B
	// c := (C)([]int{1,2,3})
	// c[1]
	if indexExpr.GetType() == gotypes.IdentifierType || indexExpr.GetType() == gotypes.SelectorType {
		for {
			var symbolDef *symboltable.SymbolDef
			if indexExpr.GetType() == gotypes.IdentifierType {
				xType := indexExpr.(*gotypes.Identifier)

				if xType.Package == "builtin" {
					break
				}
				def, _, err := ep.Config.LookupDataType(xType)
				if err != nil {
					return nil, err
				}
				symbolDef = def
			} else {
				_, sd, err := ep.retrieveQidDataType(indexExpr.(*gotypes.Selector).Prefix, &ast.Ident{Name: indexExpr.(*gotypes.Selector).Item})
				if err != nil {
					return nil, err
				}
				symbolDef = sd
			}

			if symbolDef.Def == nil {
				return nil, fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
			}

			indexExpr = symbolDef.Def
			if symbolDef.Def.GetType() == gotypes.IdentifierType || symbolDef.Def.GetType() == gotypes.SelectorType {
				continue
			}
			break
		}
	}

	ep.Config.ContractTable.AddContract(&contracts.IsIndexable{
		X: xDefAttr.TypeVarList[0],
	})

	// Get definition of the X from the symbol Table (it must be a variable of a data type)
	// and get data type of its array/map members
	switch xType := indexExpr.(type) {
	case *gotypes.Map:
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: indexAttr.TypeVarList[0],
			Y: typevars.MakeMapKey(indexAttr.TypeVarList[0]),
		})
		y := typevars.MakeMapValue(xDefAttr.TypeVarList[0])
		return types.ExprAttributeFromDataType(xType.Valuetype).AddTypeVar(y), nil
	case *gotypes.Array:
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: indexAttr.TypeVarList[0],
			Y: typevars.MakeListKey(),
		})
		y := typevars.MakeListValue(xDefAttr.TypeVarList[0])
		return types.ExprAttributeFromDataType(xType.Elmtype).AddTypeVar(y), nil
	case *gotypes.Slice:
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: indexAttr.TypeVarList[0],
			Y: typevars.MakeListKey(),
		})
		y := typevars.MakeListValue(xDefAttr.TypeVarList[0])
		return types.ExprAttributeFromDataType(xType.Elmtype).AddTypeVar(y), nil
	case *gotypes.Builtin:
		if xType.Def == "string" {
			// Checked at https://play.golang.org/
			ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X: indexAttr.TypeVarList[0],
				Y: typevars.MakeListKey(),
			})
			y := typevars.MakeListValue(xDefAttr.TypeVarList[0])
			return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: "uint8"}).AddTypeVar(y), nil
		}
		return nil, fmt.Errorf("Accessing item of built-in non-string type: %#v", xType)
	case *gotypes.Identifier:
		if xType.Def == "string" && xType.Package == "builtin" {
			// Checked at https://play.golang.org/
			ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
				X: indexAttr.TypeVarList[0],
				Y: typevars.MakeListKey(),
			})
			y := typevars.MakeListValue(xDefAttr.TypeVarList[0])
			return types.ExprAttributeFromDataType(&gotypes.Builtin{Def: "uint8"}).AddTypeVar(y), nil
		}
		return nil, fmt.Errorf("Accessing item of built-in non-string type: %#v", xType)
	case *gotypes.Ellipsis:
		ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
			X: indexAttr.TypeVarList[0],
			Y: typevars.MakeListKey(),
		})
		y := typevars.MakeListValue(xDefAttr.TypeVarList[0])
		return types.ExprAttributeFromDataType(xType.Def).AddTypeVar(y), nil
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

	y := typevars.MakeConstant(def)

	ep.Config.ContractTable.AddContract(&contracts.IsCompatibleWith{
		X:            attr.TypeVarList[0],
		Y:            y,
		ExpectedType: attr.DataTypeList[0],
	})

	return types.ExprAttributeFromDataType(def).AddTypeVar(y), nil
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
	default:
		return nil, fmt.Errorf("Unrecognized expression: %#v\n", expr)
	}
}

func New(c *types.Config) types.ExpressionParser {
	return &Parser{
		Config: c,
	}
}
