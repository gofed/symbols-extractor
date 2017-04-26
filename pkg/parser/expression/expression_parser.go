package expression

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"testing"

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

func (ep *Parser) retrieveQualifiedIdentifier(qid, symbolName string) (*gotypes.SymbolDef, error) {
	qidDef, defErr := ep.SymbolTable.LookupVariable(qid)
	if defErr != nil {
		return nil, fmt.Errorf("Unable to retrieve qid %q from the local symbol table: %v", qid, defErr)
	}

	qidPQ, ok := qidDef.Def.(*gotypes.PackageQualifier)
	if !ok {
		return nil, fmt.Errorf("Local variable %q is not a QID, it is %#v instead", qid, qidDef.Def)
	}
	st, err := ep.GlobalSymbolTable.Lookup(qidPQ.Path)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", qid, err)
	}

	packageIdent, _, piErr := st.Lookup(symbolName)
	if piErr != nil {
		return nil, fmt.Errorf("Unable to locate symbol %q in %q's symbol table: %v", symbolName, qid, piErr)
	}

	ep.AllocatedSymbolsTable.AddSymbol(qidPQ.Path, symbolName)

	return packageIdent, nil
}

func (ep *Parser) parseCompositeLitElements(lit *ast.CompositeLit, symbolDef *gotypes.SymbolDef) error {
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
		return fmt.Errorf("Unsupported 2 ClTypeIdentifierDef: %#v\n", symbolDef)
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
		qid, ok := litTypeExpr.Prefix.(*gotypes.Identifier)
		if !ok {
			return nil, fmt.Errorf("Expecting identifier in CL selector: %#v", litTypeExpr)
		}

		symbolDef, err := ep.retrieveQualifiedIdentifier(qid.Def, litTypeExpr.Item)
		if err != nil {
			return nil, err
		}

		if err := ep.parseCompositeLitElements(lit, symbolDef); err != nil {
			return nil, err
		}
	case *gotypes.Identifier:
		// If the LC type is an identifier, determine a type which is defined by the identifier
		ClTypeIdentifierDef, _, err := ep.SymbolTable.Lookup(litTypeExpr.Def)
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

func (ep *Parser) parseIdentifier(ident *ast.Ident) (gotypes.DataType, error) {
	glog.Infof("Processing identifier: %#v\n", ident)
	// TODO(jchaloup): put the nil into the allocated symbol table
	if ident.Name == "nil" {
		return &gotypes.Nil{}, nil
	}

	// true/false
	if ident.Name == "true" || ident.Name == "false" {
		return &gotypes.BuiltinLiteral{}, nil
	}

	// Check if the symbol is in the symbol table.
	// It is either a local variable or a global variable (as it is used inside an expression).
	// If the symbol is not found, it means it is not defined and never will be.

	// If it is a variable, return its definition
	if def, err := ep.SymbolTable.LookupVariable(ident.Name); err == nil {
		return def.Def, nil
	}

	// Otherwise it is a data type of a function declation -> return just the data type identifier
	def, _, err := ep.SymbolTable.Lookup(ident.Name)
	if err != nil {
		// Return an error so the function body processing can be postponed
		// TODO(jchaloup): return more information about the missing symbol so the
		// body can be re-processed right after the symbol is stored into the symbol table.
		return nil, err
	}

	// TODO(jchaloup): put the identifier's type into the allocated symbol table
	return &gotypes.Identifier{Def: def.Name}, nil
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
	case token.ARROW:
		if def[0].GetType() != gotypes.ChannelType {
			return nil, fmt.Errorf("<-OP operator expectes OP to be a channel, got %v instead", def[0].GetType())
		}
		return def[0].(*gotypes.Channel).Value, nil
	default:
		return nil, fmt.Errorf("Unary operator %#v not recognized", expr.Op)
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

	// If both types are built-in, just return built-in
	if x[0].GetType() == y[0].GetType() && x[0].GetType() == gotypes.BuiltinType {
		return &gotypes.Builtin{}, nil
	}

	// At least one of the type is an identifier
	xIdent, xOk := x[0].(*gotypes.Identifier)
	yIdent, yOk := y[0].(*gotypes.Identifier)

	if !xOk && !yOk {
		return nil, fmt.Errorf("Both operands of a binary operator must be identifiers")
	}

	// TODO(jchaloup): logical and/or over any two operands is always a bool

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

	return &gotypes.Pointer{
		Def: def[0],
	}, nil
}

func (ep *Parser) parseCallExpr(expr *ast.CallExpr) ([]gotypes.DataType, error) {
	glog.Infof("Processing CallExpr: %#v\n", expr)

	for _, arg := range expr.Args {
		// an argument is passed to the function so its data type does not affect the result data type of the call
		def, err := ep.Parse(arg)
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
	switch exprType := expr.Fun.(type) {
	case *ast.Ident:
		// simple function ID
		// check if the function name is a variable or a type

		// not a user defined type
		if def, err := ep.TypeParser.Parse(exprType); err == nil {
			if def.GetType() == gotypes.BuiltinType {
				return []gotypes.DataType{def}, nil
			}
		}

		def, symbolType, err := ep.SymbolTable.Lookup(exprType.Name)
		if err != nil {
			// Return an error so the function body processing can be postponed
			// TODO(jchaloup): return more information about the missing symbol so the
			// body can be re-processed right after the symbol is stored into the symbol table.
			return nil, err
		}
		if symbolType.IsDataType() {
			return []gotypes.DataType{
				&gotypes.Identifier{Def: exprType.Name},
			}, nil
		}

		// Get function's definition from the symbol table (if it exists)
		// and use its result type as the return type.
		// If a function/method call is a part of an expression, the function/method has only ony result type.

		// Store the symbol to the allocated ST
		ep.AllocatedSymbolsTable.AddSymbol(def.Package, exprType.Name)
		function = def.Def
	case *ast.SelectorExpr:
		def, err := ep.parseSelectorExpr(exprType)
		if err != nil {
			return []gotypes.DataType{def}, err
		}
		function = def
	case *ast.FuncLit:
		if err := ep.StmtParser.ParseFuncBody(&ast.FuncDecl{
			Type: exprType.Type,
			Body: exprType.Body,
		}); err != nil {
			return nil, err
		}

		def, err := ep.TypeParser.Parse(exprType.Type)
		if err != nil {
			return nil, err
		}

		function = def
	default:
		return nil, fmt.Errorf("Call expr not recognized: %#v", expr)
	}

	switch funcType := function.(type) {
	case *gotypes.Function:
		return funcType.Results, nil
	case *gotypes.Method:
		return funcType.Def.(*gotypes.Function).Results, nil
	default:
		return nil, fmt.Errorf("Symbol %#v to be called is not a function", function)
	}
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
	for _, field := range expr.Fields.List {
		for _, name := range field.Names {
			fmt.Printf("FieldName: %#v\n", name.Name)
		}
		fmt.Printf("Field: %#v\n", field.Type)
	}

	panic("Panic")
}

func (ep *Parser) retrieveStructField(structDefsymbol *gotypes.SymbolDef, field string) (gotypes.DataType, error) {
	glog.Infof("Retrieving StructType field %q from %#v\n", field, structDefsymbol)
	if structDefsymbol.Def.GetType() != gotypes.StructType {
		return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type: %#v", field, structDefsymbol.Def)
	}

	for _, item := range structDefsymbol.Def.(*gotypes.Struct).Fields {
		if item.Name == field {
			if structDefsymbol.Name != "" {
				ep.AllocatedSymbolsTable.AddDataTypeField(structDefsymbol.Package, structDefsymbol.Name, field)
			}
			return item.Def, nil
		}
	}
	return nil, fmt.Errorf("Unable to find a field %v in struct %#v", field, structDefsymbol)
}

func (ep *Parser) retrieveInterfaceMethod(interfaceDefsymbol *gotypes.SymbolDef, method string) (gotypes.DataType, error) {
	glog.Infof("Retrieving Interface method %q from %#v\n", method, interfaceDefsymbol)
	if interfaceDefsymbol.Def.GetType() != gotypes.InterfaceType {
		return nil, fmt.Errorf("Trying to retrieve a %v method from a non-interface data type: %#v", method, interfaceDefsymbol.Def)
	}

	for _, item := range interfaceDefsymbol.Def.(*gotypes.Interface).Methods {
		if item.Name == method {
			if interfaceDefsymbol.Name != "" {
				ep.AllocatedSymbolsTable.AddDataTypeField(interfaceDefsymbol.Package, interfaceDefsymbol.Name, method)
			}
			return item.Def, nil
		}
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

	// The struct and an interface are the only data type from which a field/method is retriveable
	switch xType := xDef[0].(type) {
	// If the X expression is a qualified id, the selector is a symbol from a package pointed by the id
	case *gotypes.PackageQualifier:
		fmt.Printf("Trying to retrieve a symbol %#v from package %v\n", expr.Sel.Name, xType.Path)
		// TODO(jchaloup): implement retrieval a symbols from other symbol tables
		panic("Symbol retrieval from other packages not yet implemented")
	case *gotypes.Pointer:
		def, ok := xType.Def.(*gotypes.Identifier)
		if !ok {
			return nil, fmt.Errorf("Trying to retrieve a %v field from a pointer to non-struct data type: %#v", expr.Sel.Name, xType.Def)
		}
		// Get struct's definition given by its identifier
		structDefsymbol, _, err := ep.SymbolTable.Lookup(def.Def)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", def.Def)
		}
		return ep.retrieveStructField(structDefsymbol, expr.Sel.Name)
	case *gotypes.Identifier:
		// Get struct/interface definition given by its identifier
		defSymbol, _, err := ep.SymbolTable.Lookup(xType.Def)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", xType.Def)
		}
		switch defSymbol.Def.(type) {
		case *gotypes.Struct:
			return ep.retrieveStructField(defSymbol, expr.Sel.Name)
		case *gotypes.Interface:
			return ep.retrieveInterfaceMethod(defSymbol, expr.Sel.Name)
		default:
			return nil, fmt.Errorf("Trying to retrieve a field/method from non-struct/non-interface data type: %#v", defSymbol)
		}
	// anonymous struct
	case *gotypes.Struct:
		return ep.retrieveStructField(&gotypes.SymbolDef{
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
	default:
		return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type: %#v", expr.Sel.Name, xDef[0])
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
	default:
		return nil, fmt.Errorf("Unrecognized expression: %#v\n", expr)
	}
}

func New(c *types.Config) types.ExpressionParser {
	return &Parser{
		Config: c,
	}
}
