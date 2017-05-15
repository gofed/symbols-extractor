package expression

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"testing"

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

func (ep *Parser) parseBasicLit(lit *ast.BasicLit) (gotypes.DataType, error) {
	glog.Infof("Processing BasicLit: %#v\n", lit)
	switch lit.Kind {
	case token.INT:
		return &gotypes.Builtin{Def: "int", Untyped: true}, nil
	case token.FLOAT:
		return &gotypes.Builtin{Def: "float", Untyped: true}, nil
	case token.IMAG:
		return &gotypes.Builtin{Def: "imag", Untyped: true}, nil
	case token.STRING:
		return &gotypes.Builtin{Def: "string", Untyped: true}, nil
	case token.CHAR:
		return &gotypes.Builtin{Def: "char", Untyped: true}, nil
	default:
		return nil, fmt.Errorf("Unrecognize BasicLit: %#v\n", lit.Kind)
	}
}

func (ep *Parser) parseKeyValueLikeExpr(expr ast.Expr, litType ast.Expr) (gotypes.DataType, error) {
	var valueExpr ast.Expr
	if kvExpr, ok := expr.(*ast.KeyValueExpr); ok {
		_, err := ep.Parse(kvExpr.Key)
		if err != nil {
			return nil, err
		}
		valueExpr = kvExpr.Value
	} else {
		valueExpr = expr
	}

	if _, ok := valueExpr.(*ast.CompositeLit); ok {
		var valueType ast.Expr
		switch elmExpr := litType.(type) {
		case *ast.SliceExpr:
			valueType = elmExpr.X
		case *ast.ArrayType:
			valueType = elmExpr.Elt
		case *ast.MapType:
			valueType = elmExpr.Value
		case *ast.Ident:
			valueType = elmExpr
		default:
			return nil, fmt.Errorf("Unknown CL type for KV elements: %#v", litType)
		}

		// If the CL type of the KV element is omitted, it needs to be reconstructed from the CL type itself
		if clExpr, ok := valueExpr.(*ast.CompositeLit); ok {
			if clExpr.Type == nil {
				pointer, ok := valueType.(*ast.StarExpr)
				if ok {
					clExpr.Type = pointer.X
					// TODO(jchaloup): should check if the pointer.X is not a pointer and fail if it is
				} else {
					clExpr.Type = valueType
				}
			}
		}

	}

	def, err := ep.Parse(valueExpr)
	if err != nil {
		return nil, err
	}
	if len(def) != 1 {
		return nil, fmt.Errorf("Expected single expression for KV value, got %#v", def)
	}
	return def[0], nil
}

func (ep *Parser) parseCompositeLitArrayLikeElements(lit *ast.CompositeLit) error {
	for _, litElement := range lit.Elts {
		_, err := ep.parseKeyValueLikeExpr(litElement, lit.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ep *Parser) parseCompositeLitStructElements(lit *ast.CompositeLit, structDef *gotypes.Struct, structSymbol *gotypes.SymbolDef) error {
	fieldCounter := 0
	fieldLen := len(structDef.Fields)
	for _, litElement := range lit.Elts {
		var valueExpr ast.Expr
		// if the struct fields name are omitted, the order matters
		// TODO(jchaloup): should check if all elements are KeyValueExpr or not (otherwise go compilation fails)
		if kvExpr, ok := litElement.(*ast.KeyValueExpr); ok {
			// The field must be an identifier
			keyDefIdentifier, ok := kvExpr.Key.(*ast.Ident)
			if !ok {
				return fmt.Errorf("Struct's field %#v is not an identifier", litElement)
			}
			if structSymbol != nil {
				ep.AllocatedSymbolsTable.AddDataTypeField(structSymbol.Package, structSymbol.Name, keyDefIdentifier.Name)
			}
			valueExpr = kvExpr.Value
		} else {
			if fieldCounter >= fieldLen {
				return fmt.Errorf("Number of fields of the CL is greater than the number of fields of underlying struct %#v", structDef)
			}
			if structSymbol != nil {
				// TODO(jchaloup): Should we count the anonymous field as well? Maybe make a AddDataTypeAnonymousField?
				ep.AllocatedSymbolsTable.AddDataTypeField(structSymbol.Package, structSymbol.Name, structDef.Fields[fieldCounter].Name)
			}
			valueExpr = litElement
		}

		// process the field value
		_, err := ep.Parse(valueExpr)
		if err != nil {
			return err
		}
		fieldCounter++
	}

	return nil
}

func (ep *Parser) parseCompositeLitElements(lit *ast.CompositeLit, symbolDef *gotypes.SymbolDef) error {
	glog.Infof("Processing parseCompositeLitElements symbolDef: %#v", symbolDef)
	switch clTypeDataType := symbolDef.Def.(type) {
	case *gotypes.Struct:
		if err := ep.parseCompositeLitStructElements(lit, clTypeDataType, symbolDef); err != nil {
			return err
		}
	case *gotypes.Array:
		if err := ep.parseCompositeLitArrayLikeElements(lit); err != nil {
			return err
		}
	case *gotypes.Slice:
		if err := ep.parseCompositeLitArrayLikeElements(lit); err != nil {
			return err
		}
	case *gotypes.Map:
		if err := ep.parseCompositeLitArrayLikeElements(lit); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported 2 type %#v in ClTypeIdentifierDef of %#v\n", symbolDef.Def, symbolDef)
	}
	return nil
}

func (ep *Parser) parseCompositeLit(lit *ast.CompositeLit) (gotypes.DataType, error) {
	glog.Infof("Processing CompositeLit: %#v\n", lit)
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
	litTypedef, err := ep.TypeParser.Parse(lit.Type)
	if err != nil {
		return nil, err
	}

	// If the CL type is anonymous struct, array, slice or map don't store fields into the allocated symbols table (AST)
	switch litTypeExpr := litTypedef.(type) {
	case *gotypes.Struct:
		// anonymous structure -> we can ignore field's allocation
		if err := ep.parseCompositeLitStructElements(lit, litTypeExpr, nil); err != nil {
			return nil, err
		}
	case *gotypes.Array:
		if err := ep.parseCompositeLitArrayLikeElements(lit); err != nil {
			return nil, err
		}
	case *gotypes.Slice:
		if err := ep.parseCompositeLitArrayLikeElements(lit); err != nil {
			return nil, err
		}
	case *gotypes.Map:
		if err := ep.parseCompositeLitArrayLikeElements(lit); err != nil {
			return nil, err
		}
	case *gotypes.Selector:
		// If it is a selector, it must be qui.id (as the litTypeExpr is a type)
		qid, ok := litTypeExpr.Prefix.(*gotypes.Packagequalifier)
		if !ok {
			return nil, fmt.Errorf("Expecting package qualifier in CL selector: %#v", litTypeExpr)
		}

		st, err := ep.GlobalSymbolTable.Lookup(qid.Path)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid, err)
		}

		symbolDef, _, piErr := st.Lookup(litTypeExpr.Item)
		if piErr != nil {
			return nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", litTypeExpr.Item, qid.Path, piErr)
		}

		ep.AllocatedSymbolsTable.AddSymbol(qid.Path, litTypeExpr.Item)

		if err := ep.parseCompositeLitElements(lit, symbolDef); err != nil {
			return nil, err
		}
	case *gotypes.Identifier:
		// If the LC type is an identifier, determine a type which is defined by the identifier
		ClTypeIdentifierDef, _, err := ep.Config.Lookup(litTypeExpr)
		if err != nil {
			return nil, fmt.Errorf("Unable to find definition of CL type of identifier %#v\n", litTypeExpr)
		}

		if err := ep.parseCompositeLitElements(lit, ClTypeIdentifierDef); err != nil {
			return nil, err
		}
	default:
		panic(fmt.Errorf("Unsupported CL type: %#v", litTypedef))
	}

	return litTypedef, nil
}

// Identifier is always a variable.
// It is either a qid, local variable or global variable of the current package.
// It can be also an identifier of a function.
// In all cases data type of the variable/function/qid is returned
func (ep *Parser) parseIdentifier(ident *ast.Ident) (gotypes.DataType, error) {
	glog.Infof("Processing variable-like identifier: %#v\n", ident)

	// TODO(jchaloup): check for variables/functions first
	// E.g. user can define
	//    make := func() {}
	// which overwrites the builtin.make function

	// Does the local symbol exists at all?
	var postponedErr error
	if !ep.SymbolTable.Exists(ident.Name) {
		postponedErr = fmt.Errorf("parseIdentifier: Symbol %q not yet processed", ident.Name)
	} else {
		// If it is a variable, return its definition
		if def, err := ep.SymbolTable.LookupVariableLikeSymbol(ident.Name); err == nil {
			byteSlice, _ := json.Marshal(def)
			glog.Infof("Variable by identifier found: %v\n", string(byteSlice))
			// The data type of the variable is not accounted as it is not implicitely used
			// The variable itself carries the data type and as long as the variable does not
			// get used, the data type can change.
			return def.Def, nil
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
				return &gotypes.Builtin{Def: "bool"}, nil
			case "nil":
				return &gotypes.Nil{}, nil
			case "iota":
				return &gotypes.Builtin{Def: "iota"}, nil
			default:
				return nil, fmt.Errorf("Unsupported built-in type: %v", ident.Name)
			}
		case symboltable.FunctionSymbol:
			return symbolDef.Def, nil
		default:
			return nil, fmt.Errorf("Unsupported symbol type: %v", symbolType)
		}
	}

	return nil, postponedErr
}

func (ep *Parser) parseUnaryExpr(expr *ast.UnaryExpr) (gotypes.DataType, error) {
	glog.Infof("Processing UnaryExpr: %#v\n", expr)
	def, err := ep.Parse(expr.X)
	if err != nil {
		return nil, err
	}

	if len(def) != 1 {
		return nil, fmt.Errorf("Operand of an unary operator is not a single value")
	}

	// TODO(jchaloup): check the token is really a unary operator
	// TODO(jchaloup): add the missing unary operator tokens
	switch expr.Op {
	// variable address
	case token.AND:
		return &gotypes.Pointer{
			Def: def[0],
		}, nil
	// channel
	case token.ARROW:
		if def[0].GetType() != gotypes.ChannelType {
			return nil, fmt.Errorf("<-OP operator expectes OP to be a channel, got %v instead", def[0].GetType())
		}
		return def[0].(*gotypes.Channel).Value, nil
		// other
	case token.XOR, token.OR, token.SUB, token.NOT, token.ADD:
		return def[0], nil
	default:
		return nil, fmt.Errorf("Unary operator %#v (%#v) not recognized", expr.Op, token.ADD)
	}
}

func (ep *Parser) parseBinaryExpr(expr *ast.BinaryExpr) (gotypes.DataType, error) {
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
	x, yErr := ep.Parse(expr.X)
	if yErr != nil {
		return nil, yErr
	}

	y, xErr := ep.Parse(expr.Y)
	if xErr != nil {
		return nil, xErr
	}

	{
		byteSlice, _ := json.Marshal(x)
		glog.Infof("xx: %v\n", string(byteSlice))
	}

	{
		byteSlice, _ := json.Marshal(y)
		glog.Infof("yy: %v\n", string(byteSlice))
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
		return &gotypes.Builtin{Def: "bool"}, nil
	}

	// If both types are built-in, just return built-in
	if x[0].GetType() == y[0].GetType() && x[0].GetType() == gotypes.BuiltinType {
		xt := x[0].(*gotypes.Builtin)
		yt := y[0].(*gotypes.Builtin)
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
				if xt.Untyped {
					// const a = 8.0 << 1
					// a is of type int, checked at https://play.golang.org/
					return &gotypes.Builtin{Def: "int", Untyped: true}, nil
				}
				return xt, nil
			case token.AND, token.OR, token.MUL, token.SUB, token.QUO, token.ADD, token.AND_NOT, token.REM, token.XOR:
				// The same reasoning as with the relational operators
				// Experiments:
				// 1+1.0 is untyped float64
				if xt.Untyped && yt.Untyped {

				}
				if xt.Untyped {
					return yt, nil
				}
				if yt.Untyped {
					return xt, nil
				}
				// byte(2)&uint8(3) is uint8
				if (xt.Def == "byte" && yt.Def == "uint8") || (yt.Def == "byte" && xt.Def == "uint8") {
					return &gotypes.Builtin{Def: "uint8"}, nil
				}
			}
			return nil, fmt.Errorf("Binary operation %q over two different built-in types: %q != %q", expr.Op, xt, yt)
		}
		if xt.Untyped {
			{
				byteSlice, _ := json.Marshal(&gotypes.Builtin{Def: xt.Def, Untyped: yt.Untyped})
				glog.Infof("xx+yy: %v\n", string(byteSlice))
			}
			return &gotypes.Builtin{Def: xt.Def, Untyped: yt.Untyped}, nil
		}
		if yt.Untyped {
			{
				byteSlice, _ := json.Marshal(&gotypes.Builtin{Def: xt.Def, Untyped: xt.Untyped})
				glog.Infof("xx+yy: %v\n", string(byteSlice))
			}
			return &gotypes.Builtin{Def: xt.Def, Untyped: xt.Untyped}, nil
		}
		// both operands are typed => Untyped = false
		{
			byteSlice, _ := json.Marshal(&gotypes.Builtin{Def: xt.Def})
			glog.Infof("xx+yy: %v\n", string(byteSlice))
		}
		return &gotypes.Builtin{Def: xt.Def}, nil
	}

	glog.Infof("Binaryexpr.x: %#v\nBinaryexpr.y: %#v\n", x[0], y[0])

	// At least one of the type is an identifier
	xIdent, xOk := x[0].(*gotypes.Identifier)
	yIdent, yOk := y[0].(*gotypes.Identifier)

	if !xOk && !yOk {
		return nil, fmt.Errorf("At least one operand of a binary operator %v must be an identifier, at %v", expr.Op, expr.Pos())
	}

	// Even here we assume existence of untyped constants.
	// If there is only one identifier it means the other operand is a built-in type.
	// Assuming the code is written correctly (it compiles), resulting data type of the operation
	// is always the identifier.
	if xOk {
		return xIdent, nil
	}
	return yIdent, nil
}

func (ep *Parser) parseStarExpr(expr *ast.StarExpr) (gotypes.DataType, error) {
	glog.Infof("Processing StarExpr: %#v\n", expr)
	def, err := ep.Parse(expr.X)
	if err != nil {
		return nil, err
	}

	if len(def) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}

	val, ok := def[0].(*gotypes.Pointer)
	if !ok {
		return nil, fmt.Errorf("Accessing a value of non-pointer type: %#v", def[0])
	}
	return val.Def, nil
}

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
	glog.Infof("Detecting isDataType for %#v", expr)
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
	default:
		// TODO(jchaloup): yes? As now it is anonymous data type. Or should we check for each such type?
		panic(fmt.Errorf("Unrecognized isDataType expr: %#v", expr))
	}
	return false, nil
}

func (ep *Parser) getFunctionDef(def gotypes.DataType) (gotypes.DataType, error) {
	switch typeDef := def.(type) {
	case *gotypes.Identifier:
		// local definition
		if typeDef.Package == "" {
			def, defType, err := ep.SymbolTable.Lookup(typeDef.Def)
			if err != nil {
				return nil, err
			}
			if defType.IsFunctionType() {
				return def.Def, nil
			}
			if !defType.IsVariable() {
				return nil, fmt.Errorf("Function call expression %#v is not expected to be a type", def)
			}
			return ep.getFunctionDef(def.Def)
		}
		panic("typeDef.Package != \"\"")
	case *gotypes.Function:
		return def, nil
	case *gotypes.Method:
		return def, nil
	case *gotypes.Pointer:
		return nil, fmt.Errorf("Can not invoke non-function expression: %#v", def)
	default:
		panic(fmt.Errorf("Unrecognized getFunctionDef def: %#v", def))
	}
	return nil, nil
}

func (ep *Parser) parseCallExpr(expr *ast.CallExpr) ([]gotypes.DataType, error) {
	glog.Infof("Processing CallExpr: %#v\n", expr)
	defer glog.Infof("Leaving CallExpr: %#v\n", expr)

	expr.Fun = ep.parseParenExpr(expr.Fun)

	processArgs := func(args []ast.Expr, params []gotypes.DataType) error {
		if params != nil {
			if len(args) == 1 {
				def, err := ep.Parse(args[0])
				if err != nil {
					return err
				}
				if len(params) == len(def) {
					return nil
				}
				if len(def) != 1 {
					return fmt.Errorf("Argument %#v of a call expression does not have one return value", args[0])
				}
				return nil
			}
		}
		for _, arg := range args {
			// an argument is passed to the function so its data type does not affect the result data type of the call
			def, err := ep.Parse(arg)
			if err != nil {
				return err
			}

			if len(def) != 1 {
				return fmt.Errorf("Argument %#v of a call expression does not have one return value", arg)
			}
			// TODO(jchaloup): data type of the argument itself can be propagated to the function/method
			// to provide more information about the type if the function parameter is an interface.
		}
		return nil
	}

	// data type => explicit type casting
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
		if err := processArgs(expr.Args, nil); err != nil {
			return nil, err
		}
		return []gotypes.DataType{def}, nil
	}
	glog.Infof("isDataType of %#v is false", expr.Fun)

	// function
	def, err := ep.ExprParser.Parse(expr.Fun)
	if err != nil {
		return nil, err
	}

	funcDef, err := ep.getFunctionDef(def[0])
	if err != nil {
		return nil, err
	}

	// a) f1() (int, int), f2(int, int) => f2(f1())
	// b) f(arg, arg, arg) => f(a,a,a)

	switch funcType := funcDef.(type) {
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
					typeDef, err := ep.TypeParser.Parse(expr.Args[0])
					return []gotypes.DataType{typeDef}, err
				case 2:
					glog.Infof("Processing make arguments for make(type, size) type: %#v", expr.Args)
					if err := processArgs([]ast.Expr{expr.Args[1]}, nil); err != nil {
						return nil, err
					}
					typeDef, err := ep.TypeParser.Parse(expr.Args[0])
					return []gotypes.DataType{typeDef}, err
				case 3:
					glog.Infof("Processing make arguments for make(type, size, size) type: %#v", expr.Args)
					if err := processArgs([]ast.Expr{expr.Args[1], expr.Args[2]}, nil); err != nil {
						return nil, err
					}
					typeDef, err := ep.TypeParser.Parse(expr.Args[0])
					return []gotypes.DataType{typeDef}, err
				default:
					return nil, fmt.Errorf("Expecting 1, 2 or 3 arguments of built-in function make, got %q instead", arglen)
				}
			case "new":
				// The new built-in function allocates memory. The first argument is a type,
				// not a value, and the value returned is a pointer to a newly
				// allocated zero value of that type.
				if len(expr.Args) != 1 {
					return nil, fmt.Errorf("Len of new args != 1, it is %#v", expr.Args)
				}
				typeDef, err := ep.TypeParser.Parse(expr.Args[0])
				return []gotypes.DataType{&gotypes.Pointer{Def: typeDef}}, err
			}
		}
		if err := processArgs(expr.Args, funcType.Params); err != nil {
			return nil, err
		}
		return funcType.Results, nil
	case *gotypes.Method:
		if err := processArgs(expr.Args, funcType.Def.(*gotypes.Function).Params); err != nil {
			return nil, err
		}
		return funcType.Def.(*gotypes.Function).Results, nil
	default:
		return nil, fmt.Errorf("Symbol %#v of %#v to be called is not a function", funcType, expr)
	}
}

func (ep *Parser) parseSliceExpr(expr *ast.SliceExpr) (gotypes.DataType, error) {
	if expr.Low != nil {
		if _, err := ep.Parse(expr.Low); err != nil {
			return nil, err
		}
	}
	if expr.High != nil {
		if _, err := ep.Parse(expr.High); err != nil {
			return nil, err
		}
	}
	if expr.Max != nil {
		if _, err := ep.Parse(expr.Max); err != nil {
			return nil, err
		}
	}

	exprDef, err := ep.Parse(expr.X)
	if err != nil {
		return nil, err
	}

	if len(exprDef) != 1 {
		return nil, fmt.Errorf("SliceExpr is not a single argument")
	}

	return exprDef[0], nil
}

func (ep *Parser) parseChanType(expr *ast.ChanType) (gotypes.DataType, error) {
	valueDef, err := ep.Parse(expr.Value)
	if err != nil {
		return nil, err
	}

	if len(valueDef) != 1 {
		return nil, fmt.Errorf("ChanType is not a single argument")
	}

	channel := &gotypes.Channel{Value: valueDef[0]}

	switch expr.Dir {
	case ast.SEND:
		channel.Dir = "1"
	case ast.RECV:
		channel.Dir = "2"
	default:
		channel.Dir = "3"
	}

	return channel, nil
}

func (ep *Parser) parseFuncLit(expr *ast.FuncLit) (gotypes.DataType, error) {
	glog.Infof("Processing FuncLit: %#v\n", expr)
	if err := ep.StmtParser.ParseFuncBody(&ast.FuncDecl{
		Type: expr.Type,
		Body: expr.Body,
	}); err != nil {
		return nil, err
	}

	return ep.TypeParser.Parse(expr.Type)
}

func (ep *Parser) parseStructType(expr *ast.StructType) (gotypes.DataType, error) {
	glog.Infof("Processing StructType: %#v\n", expr)
	return ep.TypeParser.Parse(expr)
}

func (ep *Parser) retrieveQidStruct(qidselector *gotypes.Selector) (symboltable.SymbolLookable, *gotypes.SymbolDef, error) {
	// qid.structtype expected
	qid, ok := qidselector.Prefix.(*gotypes.Packagequalifier)
	if !ok {
		return nil, nil, fmt.Errorf("Expecting a qid.structtype when retrieving a struct from a selector expression")
	}
	glog.Infof("Trying to retrieve a symbol %#v from package %v\n", qidselector.Item, qid.Path)
	qidst, err := ep.GlobalSymbolTable.Lookup(qid.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid.Path, err)
	}

	structDef, piErr := qidst.LookupDataType(qidselector.Item)
	if piErr != nil {
		return nil, nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", qidselector.Item, qid.Path, piErr)
	}
	return qidst, structDef, nil
}

// Get a struct's field.
// Given a struct can embedded another struct from a different package, the method must be able to Accessing
// symbol tables of other packages. Thus recursively process struct's definition up to all its embedded fields.
func (ep *Parser) retrieveDataTypeField(pkgsymboltable symboltable.SymbolLookable, structDefsymbol *gotypes.SymbolDef, field string) (gotypes.DataType, error) {
	glog.Infof("Retrieving data type field %q from %#v\n", field, structDefsymbol)
	// Only data type declaration is known
	if structDefsymbol.Def == nil {
		return nil, fmt.Errorf("Data type definition of %q is not known", structDefsymbol.Name)
	}

	// Any data type can have its own methods.
	if structDefsymbol.Def.GetType() != gotypes.StructType {
		// Check struct methods
		glog.Infof("Retrieving method %q of data type %q", field, structDefsymbol.Name)
		if method, err := pkgsymboltable.LookupMethod(structDefsymbol.Name, field); err == nil {
			return method.Def, nil
		}
		return nil, fmt.Errorf("Unable to find a field %v in data type %#v", field, structDefsymbol)
	}

	type embeddedStructsItem struct {
		symbolTable symboltable.SymbolLookable
		symbolDef   *gotypes.SymbolDef
	}
	var embeddedStructs []embeddedStructsItem

	// Check struct field
	var fieldItem *gotypes.StructFieldsItem
	for _, item := range structDefsymbol.Def.(*gotypes.Struct).Fields {
		fieldName := item.Name
		// anonymous field (can be embedded struct as well)
		if fieldName == "" {
			itemExpr := item.Def
			if pointerExpr, isPointer := item.Def.(*gotypes.Pointer); isPointer {
				itemExpr = pointerExpr
			}
			switch fieldType := itemExpr.(type) {
			case *gotypes.Identifier:
				if fieldType.Def == field {
					fieldItem = &item
					break
				}
				// check if the field is an embedded struct
				def, err := pkgsymboltable.LookupDataType(fieldType.Def)
				if err != nil {
					return nil, fmt.Errorf("Unable to retrieve %q type definition when retrieving a field", fieldType.Def)
				}
				if def.Def == nil {
					return nil, fmt.Errorf("Symbol %q not yet fully processed", fieldType.Def)
				}
				if _, ok := def.Def.(*gotypes.Struct); ok {
					embeddedStructs = append(embeddedStructs, embeddedStructsItem{symbolTable: pkgsymboltable, symbolDef: def})
				}
				continue
			case *gotypes.Selector:
				{
					byteSlice, _ := json.Marshal(fieldType)
					glog.Infof("++++%v\n", string(byteSlice))
				}
				// qid expected
				st, sd, err := ep.retrieveQidStruct(fieldType)
				if err != nil {
					return nil, err
				}

				embeddedStructs = append(embeddedStructs, embeddedStructsItem{symbolTable: st, symbolDef: sd})
				continue
			default:
				panic(fmt.Errorf("Unknown anonymous field type %#v", item.Def))
			}
		}
		if fieldName == field {
			fieldItem = &item
			break
		}
	}

	if fieldItem != nil {
		if structDefsymbol.Name != "" {
			ep.AllocatedSymbolsTable.AddDataTypeField(structDefsymbol.Package, structDefsymbol.Name, field)
		}
		return fieldItem.Def, nil
	}

	// First, check methods, then embedded structs

	// Check struct methods
	glog.Infof("Retrieving method %q of data type %q", field, structDefsymbol.Name)
	if method, err := pkgsymboltable.LookupMethod(structDefsymbol.Name, field); err == nil {
		return method.Def, nil
	}

	glog.Info("Retrieving fields from embedded structs")
	if len(embeddedStructs) != 0 {
		for _, item := range embeddedStructs {
			if fieldDef, err := ep.retrieveDataTypeField(item.symbolTable, item.symbolDef, field); err == nil {
				return fieldDef, nil
			}
		}
	}

	return nil, fmt.Errorf("Unable to find a field %v in struct %#v", field, structDefsymbol)
}

func (ep *Parser) retrieveInterfaceMethod(interfaceDefsymbol *gotypes.SymbolDef, method string) (gotypes.DataType, error) {
	glog.Infof("Retrieving Interface method %q from %#v\n", method, interfaceDefsymbol)
	if interfaceDefsymbol.Def.GetType() != gotypes.InterfaceType {
		return nil, fmt.Errorf("Trying to retrieve a %v method from a non-interface data type: %#v", method, interfaceDefsymbol.Def)
	}

	var methodItem *gotypes.InterfaceMethodsItem

	for _, item := range interfaceDefsymbol.Def.(*gotypes.Interface).Methods {
		methodName := item.Name
		// anonymous field (can be embedded struct as well)
		if methodName == "" {
			// Given a data type implements an interface, there is no need to check embedded interface.
			// Once we carry interface analysis, it will become relevant
			panic("Embedded interface, not yet implemented")
		}
		if methodName == method {
			methodItem = &item
			break
		}
	}

	if methodItem != nil {
		if interfaceDefsymbol.Name != "" {
			ep.AllocatedSymbolsTable.AddDataTypeField(interfaceDefsymbol.Package, interfaceDefsymbol.Name, method)
		}
		return methodItem.Def, nil
	}

	return nil, fmt.Errorf("Unable to find a method %v in interface %#v", method, interfaceDefsymbol)
}

func (ep *Parser) parseSelectorExpr(expr *ast.SelectorExpr) (gotypes.DataType, error) {
	glog.Infof("Processing SelectorExpr: %#v\n", expr)
	// X.Sel a.k.a Prefix.Item
	xDef, xErr := ep.Parse(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	if len(xDef) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}
	byteSlice, _ := json.Marshal(xDef[0])
	glog.Infof("SelectorExpr.X: %#v\tfield:%#v\t\t%v\n", xDef[0], expr.Sel, string(byteSlice))

	// The struct and an interface are the only data type from which a field/method is retriveable
	switch xType := xDef[0].(type) {
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
		ep.AllocatedSymbolsTable.AddSymbol(xType.Path, expr.Sel.Name)
		return packageIdent.Def, nil
	case *gotypes.Pointer:
		switch def := xType.Def.(type) {
		case *gotypes.Identifier:
			// Get struct's definition given by its identifier
			{
				byteSlice, _ := json.Marshal(def)
				glog.Infof("Looking for: %v\n", string(byteSlice))
			}
			structDefsymbol, err := ep.Config.LookupDataType(def)
			if err != nil {
				return nil, fmt.Errorf("Cannot retrieve identifier %q from the symbol table: %v", def.Def, err)
			}
			{
				byteSlice, _ := json.Marshal(structDefsymbol)
				glog.Infof("Struct retrieved: %v\n", string(byteSlice))
			}
			return ep.retrieveDataTypeField(ep.SymbolTable, structDefsymbol, expr.Sel.Name)
		case *gotypes.Selector:
			// qid to different package
			pq, ok := def.Prefix.(*gotypes.Packagequalifier)
			if !ok {
				return nil, fmt.Errorf("Trying to retrieve a %v field from a pointer to non-qualified struct data type: %#v", expr.Sel.Name, def)
			}
			structDefsymbol, err := ep.Config.LookupDataType(&gotypes.Identifier{Def: def.Item, Package: pq.Path})
			if err != nil {
				return nil, err
			}
			return ep.retrieveDataTypeField(ep.SymbolTable, structDefsymbol, expr.Sel.Name)
		case *gotypes.Struct:
			// anonymous struct
			return ep.retrieveDataTypeField(ep.SymbolTable, &gotypes.SymbolDef{Def: def}, expr.Sel.Name)
		default:
			return nil, fmt.Errorf("Trying to retrieve a %q field from a pointer to non-struct data type: %#v", expr.Sel.Name, xType.Def)
		}
	case *gotypes.Identifier:
		// Get struct/interface definition given by its identifier
		defSymbol, err := ep.Config.LookupDataType(xType)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", xType.Def)
		}
		if defSymbol.Def == nil {
			return nil, fmt.Errorf("Trying to retrieve a field/method from a data type %#v that is not yet fully processed", xType)
		}
		switch defSymbol.Def.(type) {
		case *gotypes.Struct:
			return ep.retrieveDataTypeField(ep.SymbolTable, defSymbol, expr.Sel.Name)
		case *gotypes.Interface:
			return ep.retrieveInterfaceMethod(defSymbol, expr.Sel.Name)
		default:
			// check data types with receivers
			glog.Infof("Retrieving method %q of a non-struct non-interface data type %#v", expr.Sel.Name, xType)
			def, err := ep.Config.LookupMethod(xType, expr.Sel.Name)
			if err != nil {
				return nil, fmt.Errorf("Trying to retrieve a field/method from non-struct/non-interface data type: %#v", defSymbol)
			}
			return def.Def, nil
		}
	// anonymous struct
	case *gotypes.Struct:
		return ep.retrieveDataTypeField(ep.SymbolTable, &gotypes.SymbolDef{
			Name:    "",
			Package: "",
			Def:     xType,
		}, expr.Sel.Name)
	case *gotypes.Interface:
		// TODO(jchaloup): test the case when the interface is anonymous
		return ep.retrieveInterfaceMethod(&gotypes.SymbolDef{
			Name:    "",
			Package: "",
			Def:     xType,
		}, expr.Sel.Name)
	case *gotypes.Selector:
		// qid.structtype expected
		st, sd, err := ep.retrieveQidStruct(xType)
		if err != nil {
			return nil, err
		}
		return ep.retrieveDataTypeField(st, sd, expr.Sel.Name)
	// case *gotypes.Builtin:
	// 	// Check if the built-in type has some methods
	// 	table, err := ep.GlobalSymbolTable.Lookup("builtin")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	def, err := ep.SymbolTable.LookupMethod(xType.Def, expr.Sel.Name)
	// 	fmt.Printf("def: %#v, err: %v\n", def, err)
	//
	// 	panic("JJJ")
	default:
		return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type when parsing selector expression %#v at %v", expr.Sel.Name, xDef[0], expr.Pos())
	}
}

func (ep *Parser) parseIndexExpr(expr *ast.IndexExpr) (gotypes.DataType, error) {
	glog.Infof("Processing IndexExpr: %#v\n", expr)
	// X[Index]
	// The Index can be a simple literal or another compound expression
	_, indexErr := ep.Parse(expr.Index)
	if indexErr != nil {
		return nil, indexErr
	}

	xDef, xErr := ep.Parse(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	if len(xDef) != 1 {
		return nil, fmt.Errorf("X of %#v does not return one value", expr)
	}

	// One can not do &(&(a))
	var indexEpxr gotypes.DataType
	pointer, ok := xDef[0].(*gotypes.Pointer)
	if ok {
		indexEpxr = pointer.Def
	} else {
		indexEpxr = xDef[0]
	}

	// Get definition of the X from the symbol Table (it must be a variable of a data type)
	// and get data type of its array/map members
	switch xType := indexEpxr.(type) {
	case *gotypes.Identifier:
		return xType, nil
	case *gotypes.Map:
		return xType.Valuetype, nil
	case *gotypes.Array:
		return xType.Elmtype, nil
	case *gotypes.Slice:
		return xType.Elmtype, nil
	case *gotypes.Builtin:
		if xType.Def == "string" {
			// Checked at https://play.golang.org/
			return &gotypes.Builtin{Def: "uint8"}, nil
		}
		return nil, fmt.Errorf("Accessing item of built-in non-string type: %#v", xType)
	case *gotypes.Ellipsis:
		return xType.Def, nil
	default:
		panic(fmt.Errorf("Unrecognized indexExpr type: %#v", xDef[0]))
	}
}

func (ep *Parser) parseTypeAssertExpr(expr *ast.TypeAssertExpr) (gotypes.DataType, error) {
	glog.Infof("Processing TypeAssertExpr: %#v\n", expr)
	// X.(Type)
	_, xErr := ep.Parse(expr.X)
	if xErr != nil {
		return nil, xErr
	}

	// We should check if the data type really implements all methods of the interface.
	// Or we can assume it does and just return the Type itself
	// TODO(jchaloup): check the data type Type really implements interface of X (if it is an interface)

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

	return def, nil
}

func (ep *Parser) Parse(expr ast.Expr) ([]gotypes.DataType, error) {
	// Given an expression we must always return its final data type
	// User defined symbols has its corresponding structs under parser/pkg/types.
	// In order to cover all possible symbol data types, we need to cover
	// golang language embedded data types as well
	switch exprType := expr.(type) {
	// Basic literal carries
	case *ast.BasicLit:
		def, err := ep.parseBasicLit(exprType)
		return []gotypes.DataType{def}, err
	case *ast.CompositeLit:
		def, err := ep.parseCompositeLit(exprType)
		return []gotypes.DataType{def}, err
	case *ast.Ident:
		def, err := ep.parseIdentifier(exprType)
		return []gotypes.DataType{def}, err
	case *ast.StarExpr:
		def, err := ep.parseStarExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.UnaryExpr:
		def, err := ep.parseUnaryExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.BinaryExpr:
		def, err := ep.parseBinaryExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.CallExpr:
		// If the call expression is the most most expression,
		// it may have a different number of results
		return ep.parseCallExpr(exprType)
	case *ast.StructType:
		def, err := ep.parseStructType(exprType)
		return []gotypes.DataType{def}, err
	case *ast.IndexExpr:
		def, err := ep.parseIndexExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.SelectorExpr:
		def, err := ep.parseSelectorExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.TypeAssertExpr:
		def, err := ep.parseTypeAssertExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.FuncLit:
		def, err := ep.parseFuncLit(exprType)
		return []gotypes.DataType{def}, err
	case *ast.SliceExpr:
		def, err := ep.parseSliceExpr(exprType)
		return []gotypes.DataType{def}, err
	case *ast.ParenExpr:
		return ep.Parse(ep.parseParenExpr(exprType))
	case *ast.ChanType:
		def, err := ep.parseChanType(exprType)
		return []gotypes.DataType{def}, err
	default:
		return nil, fmt.Errorf("Unrecognized expression: %#v\n", expr)
	}
}

func New(c *types.Config) types.ExpressionParser {
	return &Parser{
		Config: c,
	}
}
