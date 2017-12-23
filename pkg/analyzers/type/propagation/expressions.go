package propagation

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

type Config struct {
	// per subset of packages symbol table
	symbolsAccessor *accessors.Accessor
}

func New(symbolsAccessor *accessors.Accessor) *Config {
	return &Config{
		symbolsAccessor: symbolsAccessor,
	}
}

func (c *Config) BinaryExpr(exprOp token.Token, xDataType, yDataType gotypes.DataType) (gotypes.DataType, error) {
	switch exprOp {
	case token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR:
		// We should check both operands are compatible with the operator.
		// However, it requires a hard-coded knowledge of all available data types.
		// I.e. we must know int, int8, int16, int32, etc. types exists and if one
		// of the operators is untyped integral constant, the operator is valid.
		// As the list of all data types is read from the builtin package,
		// it is not possible to keep a clean solution and provide
		// the check for the operands validity at the same time.

		// TODO(jchaloup): contradicting the above, check if both operands are type compatible
		return &gotypes.Builtin{Def: "bool"}, nil
	}

	// If both types are built-in, just return built-in
	if xDataType.GetType() == yDataType.GetType() && xDataType.GetType() == gotypes.BuiltinType {
		xt := xDataType.(*gotypes.Builtin)
		yt := yDataType.(*gotypes.Builtin)
		if xt.Def != yt.Def {
			switch exprOp {
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
			return nil, fmt.Errorf("Binary operation %q over two different built-in types: %q != %q", exprOp, xt, yt)
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

	glog.Infof("Binaryexpr.x: %#v\nBinaryexpr.y: %#v\n", xDataType, yDataType)

	var xIdent *gotypes.Identifier
	var xIsIdentifier bool
	switch xExpr := xDataType.(type) {
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
	switch yExpr := yDataType.(type) {
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

	// Even here we assume existence of untyped constants.
	// If there is only one identifier it means the other operand is a built-in type.
	// Assuming the code is written correctly (it compiles), resulting data type of the operation
	// is always the identifier.
	// TODO(jchaloup): if an operand is a data type identifier, we need to return the data type itself, not its definition
	if xIsIdentifier {
		return xIdent, nil
	}
	return yIdent, nil
}

func (c *Config) UnaryExpr(exprOp token.Token, xDataType gotypes.DataType) (gotypes.DataType, error) {
	// TODO(jchaloup): check the token is really a unary operator
	// TODO(jchaloup): add the missing unary operator tokens
	switch exprOp {
	// variable address
	case token.AND:
		return &gotypes.Pointer{Def: xDataType}, nil
	// channel
	case token.ARROW:
		nonIdentDefAttr, err := c.symbolsAccessor.FindFirstNonidDataType(xDataType)
		if err != nil {
			return nil, err
		}
		if nonIdentDefAttr.GetType() != gotypes.ChannelType {
			return nil, fmt.Errorf("<-OP operator expectes OP to be a channel, got %v instead", nonIdentDefAttr.GetType())
		}
		return nonIdentDefAttr.(*gotypes.Channel).Value, nil
		// other
	case token.XOR, token.OR, token.SUB, token.NOT, token.ADD:
		// is token.OR really a unary operator?
		return xDataType, nil
	default:
		return nil, fmt.Errorf("Unary operator %#v (%#v) not recognized", exprOp, token.ADD)
	}
}

func (c *Config) SelectorExpr(xDataType gotypes.DataType, item string) (gotypes.DataType, error) {
	// X.Sel a.k.a Prefix.Item

	jsonMarshal := func(msg string, i interface{}) {
		byteSlice, _ := json.Marshal(i)
		glog.Infof("%v: %v\n", msg, string(byteSlice))
	}

	// The struct and an interface are the only data type from which a field/method is retriveable
	switch xType := xDataType.(type) {
	// If the X expression is a qualified id, the selector is a symbol from a package pointed by the id
	case *gotypes.Packagequalifier:
		glog.Infof("Trying to retrieve a symbol %#v from package %v\n", item, xType.Path)
		_, packageIdent, err := c.symbolsAccessor.RetrieveQid(xType, &ast.Ident{Name: item})
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", xType, err)
		}
		jsonMarshal("packageIdent", packageIdent)
		// TODO(jchaloup): check if the packageIdent is a global variable, method, function or a data type
		//                 based on that, use the corresponding Add method
		//ep.AllocatedSymbolsTable.AddSymbol(xType.Path, item, "")
		return packageIdent.Def, nil
	case *gotypes.Pointer:
		switch def := xType.Def.(type) {
		case *gotypes.Identifier:
			// Get struct's definition given by its identifier
			jsonMarshal("Looking for", def)
			structDefsymbol, symbolTable, err := c.symbolsAccessor.LookupDataType(def)
			if err != nil {
				return nil, fmt.Errorf("Cannot retrieve identifier %q from the symbol table: %v", def.Def, err)
			}
			jsonMarshal("Struct retrieved", structDefsymbol)

			return c.symbolsAccessor.RetrieveDataTypeField(
				accessors.NewFieldAccessor(symbolTable, structDefsymbol, &ast.Ident{Name: item}),
			)
		case *gotypes.Selector:
			// qid to different package
			// TODO(jchaloup): create a contract once the multi-package contracts collecting is implemented
			pq, ok := def.Prefix.(*gotypes.Packagequalifier)
			if !ok {
				return nil, fmt.Errorf("Trying to retrieve a %v field from a pointer to non-qualified struct data type: %#v", item, def)
			}
			structDefsymbol, symbolTable, err := c.symbolsAccessor.LookupDataType(&gotypes.Identifier{Def: def.Item, Package: pq.Path})
			if err != nil {
				return nil, err
			}
			return c.symbolsAccessor.RetrieveDataTypeField(
				accessors.NewFieldAccessor(symbolTable, structDefsymbol, &ast.Ident{Name: item}),
			)
		case *gotypes.Struct:
			// anonymous struct
			_, currentSt := c.symbolsAccessor.CurrentTable()
			return c.symbolsAccessor.RetrieveDataTypeField(
				accessors.NewFieldAccessor(currentSt, &symbols.SymbolDef{Def: def}, &ast.Ident{Name: item}),
			)
		default:
			return nil, fmt.Errorf("Trying to retrieve a %q field from a pointer to non-struct data type: %#v", item, xType.Def)
		}
	case *gotypes.Identifier:
		// Get struct/interface definition given by its identifier
		defSymbol, symbolTable, err := c.symbolsAccessor.LookupDataType(xType)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", xType.Def)
		}
		if defSymbol.Def == nil {
			return nil, fmt.Errorf("Trying to retrieve a field/method from a data type %#v that is not yet fully processed", xType)
		}
		switch defSymbol.Def.(type) {
		case *gotypes.Struct:
			return c.symbolsAccessor.RetrieveDataTypeField(
				accessors.NewFieldAccessor(symbolTable, defSymbol, &ast.Ident{Name: item}),
			)
		case *gotypes.Interface:
			return c.symbolsAccessor.RetrieveInterfaceMethod(symbolTable, defSymbol, item)
		case *gotypes.Identifier:
			return c.symbolsAccessor.RetrieveDataTypeField(
				accessors.NewFieldAccessor(symbolTable, defSymbol, &ast.Ident{Name: item}).SetFieldsOnly().SetDropFieldsOnly(),
			)
		default:
			// check data types with receivers
			glog.Infof("Retrieving method %q of a non-struct non-interface data type %#v", item, xType)
			def, err := c.symbolsAccessor.LookupMethod(xType, item)
			if err != nil {
				return nil, fmt.Errorf("Trying to retrieve a field/method from non-struct/non-interface data type: %#v: %v", defSymbol, err)
			}
			return def.Def, nil
		}
	// anonymous struct
	case *gotypes.Struct:
		_, currentSt := c.symbolsAccessor.CurrentTable()
		return c.symbolsAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(currentSt, &symbols.SymbolDef{
				Name:    "",
				Package: "",
				Def:     xType,
			}, &ast.Ident{Name: item}),
		)
	case *gotypes.Interface:
		// TODO(jchaloup): test the case when the interface is anonymous
		_, currentSt := c.symbolsAccessor.CurrentTable()
		return c.symbolsAccessor.RetrieveInterfaceMethod(currentSt, &symbols.SymbolDef{Def: xType}, item)
	case *gotypes.Selector:
		// qid.structtype expected
		st, sd, err := c.symbolsAccessor.RetrieveQidDataType(xType.Prefix, &ast.Ident{Name: xType.Item})
		if err != nil {
			return nil, err
		}
		return c.symbolsAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(st, sd, &ast.Ident{Name: item}),
		)
	case *gotypes.Builtin:
		// Check if the built-in type has some methods
		def, table, err := c.symbolsAccessor.GetBuiltinDataType(xType.Def)
		if err != nil {
			return nil, err
		}
		return c.symbolsAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(table, def, &ast.Ident{Name: item}).SetMethodsOnly(),
		)
	default:
		return nil, fmt.Errorf("Trying to retrieve a %v field from a non-struct data type when parsing selector expression %#v", item, xDataType)
	}
}
