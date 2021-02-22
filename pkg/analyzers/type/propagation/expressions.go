package propagation

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/compatibility"
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

type opr struct {
	pkg, id   string
	oPkg, oId string

	untyped        bool
	variable       bool
	literal        string
	underlyingType *accessors.UnderlyingType
}

func (xType *opr) equals(yType *opr) bool {
	xId := xType.oId
	if xType.oPkg == "builtin" {
		if xType.oId == "byte" {
			xId = "uint8"
		} else if xType.oId == "rune" {
			xId = "int32"
		}
	}

	yId := yType.oId
	if yType.oPkg == "builtin" {
		if yType.oId == "byte" {
			yId = "uint8"
		} else if yType.oId == "rune" {
			yId = "int32"
		}
	}

	// x unchanged?
	if xType.pkg == xType.oPkg && xType.id == xType.oId {
		return yType.pkg == xType.pkg && yType.id == yType.id
	}

	// y unchanged?
	if yType.pkg == yType.oPkg && yType.id == yType.oId {
		return yType.pkg == xType.pkg && yType.id == yType.id
	}

	return xId == yId && xType.oPkg == yType.oPkg
}

func (xType *opr) getOriginalType() string {
	if xType.oPkg == "" {
		return xType.oId
	}
	return fmt.Sprintf("%v.%v", xType.oPkg, xType.oId)
}

func (xType *opr) getType() (string, string) {
	if xType.oPkg == "builtin" {
		if xType.oId == "byte" {
			return "builtin", "uint8"
		} else if xType.oId == "rune" {
			return "builtin", "int32"
		}
	}
	return xType.oPkg, xType.oId
}

//////////////////////////////////////////////////////
// type category methods
//////////////////////////////////////////////////////

func (c *Config) isIntegral(x *opr) bool {
	if x.pkg == "builtin" {
		return c.symbolsAccessor.IsIntegral(x.id)
	}
	if x.pkg == "C" {
		return c.symbolsAccessor.IsCIntegral(x.id)
	}
	return false
}

func (c *Config) isUintegral(x *opr) bool {
	if x.pkg == "builtin" {
		return c.symbolsAccessor.IsUintegral(x.id)
	}
	return false
}

func (c *Config) isFloating(x *opr) bool {
	if x.pkg == "builtin" {
		return c.symbolsAccessor.IsFloating(x.id)
	}
	return false
}

func (c *Config) isComplex(x *opr) bool {
	if x.pkg == "builtin" {
		return c.symbolsAccessor.IsComplex(x.id)
	}
	return false
}

func (c *Config) areIntegral(x, y *opr) bool {
	return c.isIntegral(x) && c.isIntegral(y)
}

func (c *Config) areFloating(x, y *opr) bool {
	if c.isFloating(x) {
		return c.isFloating(y) || c.isIntegral(y)
	}
	if c.isFloating(y) {
		return c.isFloating(x) || c.isIntegral(x)
	}
	return false
}

func (c *Config) areComplex(x, y *opr) bool {
	if c.isComplex(x) {
		return c.isIntegral(y) || c.isFloating(y) || c.isComplex(y)
	}
	if c.isComplex(y) {
		return c.isIntegral(x) || c.isFloating(x) || c.isComplex(x)
	}
	return false
}

func (c *Config) areBooleans(x, y *opr) bool {
	if x.pkg == "builtin" && y.pkg == "builtin" {
		if x.id == "bool" && y.id == "bool" {
			return true
		}
	}
	return false
}

func (c *Config) areStrings(x, y *opr) bool {
	if x.pkg == "builtin" && y.pkg == "builtin" {
		if x.id == "string" && y.id == "string" {
			return true
		}
	}
	return false
}

func (c *Config) arePointers(x, y *opr) bool {
	if x.pkg == "<pointer>" && y.pkg == "<pointer>" {
		return true
	}

	if x.pkg == "<pointer>" && y.pkg == "<nil>" {
		return true
	}

	if x.pkg == "<nil>" && y.pkg == "<pointer>" {
		return true
	}

	return false
}

func (c *Config) areStructs(x, y *opr) bool {
	if x.pkg == "<struct>" && y.pkg == "<struct>" {
		return true
	}
	return false
}

func (c *Config) areInterfaces(x, y *opr) bool {
	if x.pkg == "<interface>" || y.pkg == "<interface>" {
		return true
	}

	return false
}

func (c *Config) areValueless(x, y *opr) bool {
	if x.literal == "*" || y.literal == "*" {
		return true
	}
	return false
}

func (c *Config) getOperandType(x gotypes.DataType) (*opr, error) {
	// Get underlying type
	underlyingType, err := c.symbolsAccessor.ResolveToUnderlyingType(x)
	if err != nil {
		return nil, err
	}

	if underlyingType.SymbolType == gotypes.InterfaceType {
		return &opr{
			pkg:            "<interface>",
			underlyingType: underlyingType,
		}, nil
	}

	if underlyingType.SymbolType == gotypes.NilType {
		return &opr{
			pkg: "<nil>",
		}, nil
	}

	if underlyingType.SymbolType == gotypes.PointerType {
		return &opr{
			pkg:            "<pointer>",
			underlyingType: underlyingType,
		}, nil
	}

	if underlyingType.SymbolType == gotypes.StructType {
		return &opr{
			pkg:            "<struct>",
			underlyingType: underlyingType,
		}, nil
	}

	oType := &opr{}

	// Get original type
	switch d := x.(type) {
	case *gotypes.Constant:
		oType.oPkg, oType.oId = d.Package, d.Def
		oType.literal = d.Literal
		oType.untyped = d.Untyped
	case *gotypes.Builtin:
		oType.oPkg, oType.oId = "builtin", d.Def
		oType.variable = true
		oType.literal = d.Literal
		oType.untyped = d.Untyped
	case *gotypes.Identifier:
		oType.oPkg, oType.oId = d.Package, d.Def
		oType.variable = true
	case *gotypes.Selector:
		qid, ok := d.Prefix.(*gotypes.Packagequalifier)
		if !ok {
			return nil, fmt.Errorf("Operand of a binary expression is expected to be a qualified identifier, got %#v instead", d)
		}
		oType.oPkg, oType.oId = qid.Path, d.Item
		oType.variable = true
	case *gotypes.Pointer:
		// pointer to C symbol
		if underlyingType.SymbolType != gotypes.IdentifierType {
			return nil, fmt.Errorf("Operand of a binary expression is expected to be a pointer to C data type, got pointer to %#v instead", d.Def)
		}
		ident := underlyingType.Def.(*gotypes.Identifier)
		oType.oPkg, oType.oId = ident.Package, ident.Def
		oType.variable = true
	default:
		return nil, fmt.Errorf("Unrecognized operand type %#v", x)
	}

	switch underlyingType.SymbolType {
	case gotypes.BuiltinType:
		if x.GetType() == gotypes.BuiltinType {
			b := underlyingType.Def.(*gotypes.Builtin)
			oType.pkg, oType.id = "builtin", b.Def
		} else {
			b := underlyingType.Def.(*gotypes.Identifier)
			oType.pkg, oType.id = b.Package, b.Def
		}
	case gotypes.IdentifierType:
		// C-language data types
		b := underlyingType.Def.(*gotypes.Identifier)
		oType.pkg, oType.id = b.Package, b.Def
	default:
		panic(fmt.Errorf("Unrecognized underlying TTTType: %#v", underlyingType))
	}

	if oType.pkg == "builtin" {
		if oType.id == "byte" {
			oType.id = "uint8"

		} else if oType.id == "rune" {
			oType.id = "int32"
		}
	}
	return oType, nil
}

func (c *Config) isCompatibleWithInt(xType *opr, xDataType gotypes.DataType) bool {
	if c.isIntegral(xType) {
		return true
	}
	if xDataType.GetType() != gotypes.ConstantType {
		return false
	}
	// Must be covertible to integer, in case it is float untyped constant
	return NewMultiArith().AddXFromString(xDataType.(*gotypes.Constant).Literal).IsXInt()
}

func (c *Config) isCompatibleWithUint(xType *opr, xDataType gotypes.DataType) bool {
	if c.isUintegral(xType) {
		return true
	}
	if xDataType.GetType() != gotypes.ConstantType {
		return false
	}
	// Must be covertible to integer, in case it is float untyped constant
	return NewMultiArith().AddXFromString(xDataType.(*gotypes.Constant).Literal).IsXUint()
}

func (c *Config) isCompatibleWithFloat(xType *opr, xDataType gotypes.DataType) bool {
	if c.isIntegral(xType) || c.isFloating(xType) {
		return true
	}
	if xDataType.GetType() != gotypes.ConstantType {
		return false
	}
	// Must be covertible to integer, in case it is float untyped constant
	return NewMultiArith().AddXFromString(xDataType.(*gotypes.Constant).Literal).IsXFloat()
}

func (c *Config) isConstant(x gotypes.DataType) bool {
	return x.GetType() == gotypes.ConstantType
}

func (c *Config) isUntypedVar(x gotypes.DataType) bool {
	return x.GetType() == gotypes.BuiltinType
}

func (c *Config) isValueLess(x gotypes.DataType) bool {
	constant, ok := x.(*gotypes.Constant)
	if !ok {
		return false
	}
	return constant.Literal == "*"
}

//////////////////////////////////////////////////////
// data type propagation methods
//////////////////////////////////////////////////////

func (c *Config) binaryExprNumeric(exprOp token.Token, xType, yType *opr, xDataType, yDataType gotypes.DataType) (gotypes.DataType, error) {
	ma := NewMultiArith()

	// Types of operands:
	// - untyped constants
	// - typed constants
	// - untyped variables (until assigned to a variable), e.g. var > var, it is true, can be asigned to any type defined as bool
	// - typed varible a.k.a identifiers
	//
	// Examples:
	// - a := 2 // int variable
	// - 2.0 << a // untyped int variable since the a is int variable
	// The <<, resp. >> are the only operands that can produce untyped variables.
	// At the same time, the untyped variables are produced only if the LHS is untyped int constant.
	// Otherwise, the product of the <<, resp. >> is a typed variable

	//////////////////////////////////////
	// UIV +|-|*|/|%|&|||&^|^ UIV = UIV
	// UIV <<\>> UIV = UIV
	// UIV ==|!=|<=|<|>=|> UIV = bool

	// called over constants only, only constants has literals to operate with

	glog.V(2).Infof("Performing %v operation", exprOp)
	switch exprOp {
	case token.REM, token.AND, token.OR, token.AND_NOT, token.XOR:
		if exprOp == token.REM {
			// just in case the first operand is typed int and the second operant is int compatible
			if !c.isIntegral(xType) {
				return nil, fmt.Errorf("mismatched types %v %v %v", xType.getOriginalType(), exprOp, yType.getOriginalType())
			}
			if !c.isCompatibleWithInt(yType, yDataType) {
				return nil, fmt.Errorf("mismatched types %v %v %v", xType.getOriginalType(), exprOp, yType.getOriginalType())
			}
		} else {
			// purely integral ops
			if !c.areIntegral(xType, yType) {
				return nil, fmt.Errorf("%v (%v) %v %v (%v) not supported", xType.literal, xType.getOriginalType(), exprOp, yType.literal, yType.getOriginalType())
			}
		}
		// Constants
		if c.isConstant(xDataType) && c.isConstant(yDataType) {
			var targetType *opr
			untyped := false
			if xType.untyped && yType.untyped {
				if c.isValueLess(xDataType) || c.isValueLess(yDataType) {
					return &gotypes.Constant{
						Package: xType.oPkg,
						Def:     xType.oId,
						Literal: "*",
						Untyped: true,
					}, nil
				}
				untyped = true
				targetType = xType
			} else if xType.untyped {
				if c.isValueLess(yDataType) {
					// x is untyped int constant
					return yDataType, nil
				}
				targetType = yType
			} else if yType.untyped {
				if c.isValueLess(xDataType) {
					// x is untyped int constant
					return xDataType, nil
				}
				targetType = xType
			} else {
				if !xType.equals(yType) {
					return nil, fmt.Errorf("%v can not be converted to %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				if c.isValueLess(xDataType) || c.isValueLess(yDataType) {
					return &gotypes.Constant{
						Package: xType.oPkg,
						Def:     xType.oId,
						Literal: "*",
					}, nil
				}
				targetType = xType
			}
			ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
			ma.AddYFromString(yDataType.(*gotypes.Constant).Literal)
			if err := ma.Perform(exprOp).Error(); err != nil {
				return nil, err
			}
			lit, err := ma.ZToLiteral(targetType.id, false)
			if err != nil {
				return nil, err
			}
			return &gotypes.Constant{Package: targetType.oPkg, Def: targetType.oId, Literal: lit, Untyped: untyped}, nil
		}
		// Y is constant
		if c.isConstant(yDataType) {
			if c.isUntypedVar(xDataType) {
				b := xDataType.(*gotypes.Builtin)
				// Product of << or >> operation.
				// So the Y constant must be integral since the <<, resp. >> does not accept non-integral LHS.
				// See https://golang.org/ref/spec#Operators and the "shift expression" paragraph with examples.
				if !c.isIntegral(yType) {
					return nil, fmt.Errorf("Op %v expects the LHS to be int, got %#v instead", b.Literal, yType.getOriginalType())
				}
				// Y is untyped int constant
				if yType.untyped {
					b.Literal = ""
					return b, nil
				}
				// Y is typed int constant
				return &gotypes.Identifier{
					Package: "builtin",
					Def:     yDataType.(*gotypes.Constant).Def,
				}, nil
			}
			// Y is typed variable
			// is the Y constant compatible with the typed variable X?
			if yType.untyped {
				// Y is untyped constant
				// TODO(jchaloup): add error check for all Add?FromString calls to avoid panics
				ma.AddXFromString(yDataType.(*gotypes.Constant).Literal)
				if _, err := ma.XToLiteral(xType.id); err != nil {
					return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				return xDataType, nil
			}
			// both X,Y are typed
			if !xType.equals(yType) {
				return nil, fmt.Errorf("mismatched types %v and %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
			}
			return xDataType, nil
		}
		// X is constant
		if c.isConstant(xDataType) {
			// Y is untyped variable
			if c.isUntypedVar(yDataType) {
				b := yDataType.(*gotypes.Builtin)
				if !c.isIntegral(xType) {
					return nil, fmt.Errorf("Op %v expects the LHS to be int, got %#v instead", b.Literal, xType.getOriginalType())
				}
				// Y is untyped int constant
				if xType.untyped {
					b.Literal = ""
					return b, nil
				}
				// Y is typed int constant
				return &gotypes.Identifier{
					Package: "builtin",
					Def:     xDataType.(*gotypes.Constant).Def,
				}, nil
			}
			// X is typed variable
			// is the Y constant compatible with the typed variable X?
			if xType.untyped {
				// Y is untyped constant
				ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
				if _, err := ma.XToLiteral(yType.id); err != nil {
					return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				return yDataType, nil
			}
			// both X,Y are typed
			if !xType.equals(yType) {
				return nil, fmt.Errorf("mismatched types %v and %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
			}
			return yDataType, nil
		}
		// Both X,Y are variable
		if c.isUntypedVar(xDataType) && c.isUntypedVar(yDataType) {
			return &gotypes.Builtin{Def: "int", Untyped: true}, nil
		}
		// X is untyped var
		if c.isUntypedVar(xDataType) {
			// is Y compatible with X?
			if !c.isIntegral(yType) {
				return nil, fmt.Errorf("The second operand %v of %v is not int", yDataType, exprOp)
			}
			return yDataType, nil
		}
		// Y is untyped var
		if c.isUntypedVar(yDataType) {
			// is X compatible with Y?
			if !c.isIntegral(xType) {
				return nil, fmt.Errorf("The first operand %v of %v is not int", xDataType, exprOp)
			}
			return xDataType, nil
		}
		// both X,Y are typed
		if !xType.equals(yType) {
			return nil, fmt.Errorf("mismatched types %v and %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
		}
		return xDataType, nil
	case token.SHL, token.SHR:
		// operating with floats that has no floating part
		if !c.isCompatibleWithInt(xType, xDataType) {
			return nil, fmt.Errorf("The first operand %v of %v is not int", xDataType, exprOp)
		}
		if !c.isCompatibleWithUint(yType, yDataType) && !c.isCompatibleWithInt(yType, yDataType) {
			return nil, fmt.Errorf("The second operand %v of %v is not uint nor with int (runtime panic if negative)", yDataType, exprOp)
		}
		if xDataType.GetType() == gotypes.ConstantType && yDataType.GetType() == gotypes.ConstantType {
			var targetType *opr
			untyped := false
			if xType.untyped && yType.untyped {
				untyped = true
				targetType = &opr{
					pkg:     "builtin",
					id:      "int",
					oPkg:    "builtin",
					oId:     "int",
					untyped: true,
				}
			} else if xType.untyped {
				if c.isValueLess(yDataType) {
					constant := xDataType.(*gotypes.Constant)
					constant.Literal = "*"
					return constant, nil
				}
				targetType = xType
				untyped = true
			} else if yType.untyped {
				if c.isValueLess(xDataType) {
					// Y must be untyped int constant
					return xDataType, nil
				}
				targetType = xType
			} else {
				if !xType.equals(yType) {
					return nil, fmt.Errorf("%v can not be converted to %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				if c.isValueLess(xDataType) || c.isValueLess(yDataType) {
					return &gotypes.Constant{
						Package: xType.oPkg,
						Def:     xType.oId,
						Literal: "*",
					}, nil
				}
				targetType = xType
			}
			ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
			ma.AddYFromString(yDataType.(*gotypes.Constant).Literal)
			if err := ma.Perform(exprOp).Error(); err != nil {
				return nil, err
			}
			lit, err := ma.ZToLiteral(targetType.id, false)
			if err != nil {
				return nil, err
			}
			return &gotypes.Constant{Package: targetType.oPkg, Def: targetType.oId, Literal: lit, Untyped: untyped}, nil
		}
		if c.isConstant(xDataType) {
			// is Y is a variable, X must be integral
			if !c.isIntegral(xType) {
				return nil, fmt.Errorf("The first operand %v of %v is not int", xDataType, exprOp)
			}
			if xType.untyped {
				return &gotypes.Builtin{Def: xType.oId, Untyped: true, Literal: fmt.Sprintf("%v", exprOp)}, nil
			}
			return &gotypes.Identifier{Package: xType.oPkg, Def: xType.oId}, nil
		}
		// Y is uint (either constant or variable). So the product is a variable.
		// Either untyped variable or typed varible, which is determined by the xDataType only
		if c.isUntypedVar(xDataType) {
			b := xDataType.(*gotypes.Builtin)
			b.Literal = fmt.Sprintf("%v", exprOp)
			return b, nil
		}
		return xDataType, nil
	case token.MUL, token.QUO, token.ADD, token.SUB:
		// ints, float and complex numbers allowed
		// Both X,Y are constants
		if xDataType.GetType() == gotypes.ConstantType && yDataType.GetType() == gotypes.ConstantType {
			var targetType *opr
			untyped := false
			if xType.untyped && yType.untyped {
				untyped = true
				if c.areComplex(xType, yType) {
					targetType = &opr{
						pkg:  "builtin",
						id:   "complex128",
						oPkg: "builtin",
						oId:  "complex128",
					}
				} else if c.areFloating(xType, yType) {
					targetType = &opr{
						pkg:  "builtin",
						id:   "float64",
						oPkg: "builtin",
						oId:  "float64",
					}
				} else {
					targetType = xType
				}
				if c.isValueLess(xDataType) || c.isValueLess(yDataType) {
					return &gotypes.Constant{
						Package: targetType.oPkg,
						Def:     targetType.oId,
						Literal: "*",
						Untyped: true,
					}, nil
				}
			} else if xType.untyped {
				// it's either untyped int, untyped float or untyped complex
				// Y is int => X must be convertible to int
				if c.isIntegral(yType) {
					if !c.isCompatibleWithInt(xType, xDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
					if c.isValueLess(yDataType) {
						// x must be untyped int constant and it is (or at least convertible to it)
						return yDataType, nil
					}
				} else if c.isFloating(yType) {
					if !c.isCompatibleWithFloat(xType, xDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
					if c.isValueLess(yDataType) {
						// Y must be untyped int constant and it is (or at least convertible to it)
						return yDataType, nil
					}
				}
				targetType = yType
			} else if yType.untyped {
				if c.isIntegral(xType) {
					if !c.isCompatibleWithInt(yType, yDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
					if c.isValueLess(xDataType) {
						// Y must be untyped int constant and it is (or at least convertible to it)
						return xDataType, nil
					}
				} else if c.isFloating(xType) {
					if !c.isCompatibleWithFloat(yType, yDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
					if c.isValueLess(xDataType) {
						// Y must be untyped int constant and it is (or at least convertible to it)
						return xDataType, nil
					}
				}
				targetType = xType
			} else {
				if !xType.equals(yType) {
					return nil, fmt.Errorf("%v can not be converted to %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				if c.isValueLess(xDataType) || c.isValueLess(yDataType) {
					return &gotypes.Constant{
						Package: xType.oPkg,
						Def:     xType.oId,
						Literal: "*",
					}, nil
				}
				targetType = xType
			}
			ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
			ma.AddYFromString(yDataType.(*gotypes.Constant).Literal)
			if err := ma.Perform(exprOp).Error(); err != nil {
				return nil, err
			}
			if exprOp == token.QUO && c.areIntegral(xType, yType) {
				ma.ZFloor()
			}
			lit, err := ma.ZToLiteral(targetType.id, false)
			if err != nil {
				return nil, err
			}
			return &gotypes.Constant{Package: targetType.oPkg, Def: targetType.oId, Literal: lit, Untyped: untyped}, nil
		}
		if c.isConstant(yDataType) {
			if c.isUntypedVar(xDataType) {
				b := xDataType.(*gotypes.Builtin)
				// Product of << or >> operation.
				// So the Y constant must be integral since the <<, resp. >> does not accept non-integral LHS.
				// See https://golang.org/ref/spec#Operators and the "shift expression" paragraph with examples.
				if (b.Literal == "<<" || b.Literal == ">>") && !c.isIntegral(yType) {
					return nil, fmt.Errorf("Op %v expects the LHS to be int, got %#v instead", b.Literal, yType.getOriginalType())
				}
				// Y is untyped int constant
				if yType.untyped {
					return xDataType, nil
				}
				// Y is typed int constant
				return &gotypes.Identifier{
					Package: "builtin",
					Def:     yDataType.(*gotypes.Constant).Def,
				}, nil
			}
			// Y is typed variable
			// is the Y constant compatible with the typed variable X?
			if yType.untyped {
				// Y is untyped constant
				ma.AddXFromString(yDataType.(*gotypes.Constant).Literal)
				var targetType string
				if xType.pkg == "C" {
					goIdent, err := c.symbolsAccessor.CToGoUnderlyingType(&gotypes.Identifier{Package: "C", Def: xType.id})
					if err != nil {
						return nil, err
					}
					targetType = goIdent.Def
				} else {
					targetType = xType.id
				}
				if _, err := ma.XToLiteral(targetType); err != nil {
					return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				return xDataType, nil
			}
			// both X,Y are typed
			if !xType.equals(yType) {
				return nil, fmt.Errorf("mismatched types %v and %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
			}
			return xDataType, nil
		}
		if c.isConstant(xDataType) {
			if c.isUntypedVar(yDataType) {
				b := yDataType.(*gotypes.Builtin)
				// Product of << or >> operation.
				// So the Y constant must be integral since the <<, resp. >> does not accept non-integral LHS.
				// See https://golang.org/ref/spec#Operators and the "shift expression" paragraph with examples.
				if !c.isIntegral(xType) {
					return nil, fmt.Errorf("Op %v expects the RHS to be int, got %#v instead", b.Literal, xType.getOriginalType())
				}
				// Y is untyped int constant
				if xType.untyped {
					return yDataType, nil
				}
				// Y is typed int constant
				return &gotypes.Identifier{
					Package: "builtin",
					Def:     xDataType.(*gotypes.Constant).Def,
				}, nil
			}
			// X is typed variable
			// is the Y constant compatible with the typed variable X?
			if xType.untyped {
				// Y is untyped constant
				ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
				var targetType string
				if yType.pkg == "C" {
					goIdent, err := c.symbolsAccessor.CToGoUnderlyingType(&gotypes.Identifier{Package: "C", Def: yType.id})
					if err != nil {
						return nil, err
					}
					targetType = goIdent.Def
				} else {
					targetType = yType.id
				}
				if _, err := ma.XToLiteral(targetType); err != nil {
					return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				return yDataType, nil
			}
			// both X,Y are typed
			if !xType.equals(yType) {
				return nil, fmt.Errorf("mismatched types %v and %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
			}
			return yDataType, nil
		}
		// Both X,Y are variable
		if c.isUntypedVar(xDataType) && c.isUntypedVar(yDataType) {
			return &gotypes.Builtin{Def: "int", Untyped: true}, nil
		}
		// X is untyped var
		if c.isUntypedVar(xDataType) {
			// is Y compatible with X?
			if !c.isIntegral(yType) {
				return nil, fmt.Errorf("The second operand %v of %v is not int", yDataType, exprOp)
			}
			return yDataType, nil
		}
		// Y is untyped var
		if c.isUntypedVar(yDataType) {
			// is X compatible with Y?
			if !c.isIntegral(xType) {
				return nil, fmt.Errorf("The first operand %v of %v is not int", xDataType, exprOp)
			}
			return xDataType, nil
		}
		// both X,Y are typed
		if !xType.equals(yType) {
			return nil, fmt.Errorf("mismatched types %v and %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
		}
		return xDataType, nil
	case token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR:
		// Both X,Y are constants
		if xDataType.GetType() == gotypes.ConstantType && yDataType.GetType() == gotypes.ConstantType {
			if xType.untyped && yType.untyped {
				if c.isValueLess(xDataType) || c.isValueLess(yDataType) {
					return &gotypes.Constant{Package: "builtin", Def: "bool", Literal: "*", Untyped: true}, nil
				}

				ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
				ma.AddYFromString(yDataType.(*gotypes.Constant).Literal)
				if err := ma.Perform(exprOp).Error(); err != nil {
					return nil, err
				}
				var lit string
				if ma.ZBool() {
					lit = "true"
				} else {
					lit = "false"
				}
				return &gotypes.Constant{Package: "builtin", Def: "bool", Literal: lit, Untyped: true}, nil
			} else if xType.untyped {
				// it's either untyped int, untyped float or untyped complex
				// Y is int => X must be convertible to int
				if c.isIntegral(yType) {
					if !c.isCompatibleWithInt(xType, xDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
					if c.isValueLess(yDataType) {
						// x must be untyped int constant and it is (or at least convertible to it)
						return &gotypes.Constant{Package: "builtin", Def: "bool", Literal: "*", Untyped: true}, nil
					}
				} else if c.isFloating(yType) {
					if !c.isCompatibleWithFloat(xType, xDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
				}
			} else if yType.untyped {
				if c.isIntegral(xType) {
					if !c.isCompatibleWithInt(yType, yDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
					if c.isValueLess(xDataType) {
						// Y must be untyped int constant and it is (or at least convertible to it)
						return &gotypes.Constant{Package: "builtin", Def: "bool", Literal: "*", Untyped: true}, nil
					}
				} else if c.isFloating(xType) {
					if !c.isCompatibleWithFloat(yType, yDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
				}
			} else {
				if !xType.equals(yType) {
					return nil, fmt.Errorf("%v can not be converted to %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
				}
				if c.isValueLess(xDataType) || c.isValueLess(yDataType) {
					return &gotypes.Constant{Package: "builtin", Def: "bool", Literal: "*", Untyped: true}, nil
				}
			}
			ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
			ma.AddYFromString(yDataType.(*gotypes.Constant).Literal)
			if err := ma.Perform(exprOp).Error(); err != nil {
				return nil, err
			}
			var lit string
			if ma.ZBool() {
				lit = "true"
			} else {
				lit = "false"
			}
			return &gotypes.Constant{Package: "builtin", Def: "bool", Literal: lit, Untyped: true}, nil
		}
		if c.isConstant(yDataType) {
			if yType.untyped {
				if c.isIntegral(xType) {
					if !c.isCompatibleWithInt(yType, yDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
				} else if c.isFloating(xType) {
					if !c.isCompatibleWithFloat(yType, yDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
				}
			}
			return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
		}
		if c.isConstant(xDataType) {
			if xType.untyped {
				if c.isIntegral(yType) {
					if !c.isCompatibleWithInt(xType, xDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
				} else if c.isFloating(yType) {
					if !c.isCompatibleWithFloat(xType, xDataType) {
						return nil, fmt.Errorf("mismatched types %v and %v for %v op", xType.getOriginalType(), yType.getOriginalType(), exprOp)
					}
				}
			}
			return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
		}
		// both constants
		if c.isUntypedVar(xDataType) && c.isUntypedVar(yDataType) {
			return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
		}
		// X is untyped var
		if c.isUntypedVar(xDataType) {
			// is Y compatible with X?
			if !c.isIntegral(yType) {
				return nil, fmt.Errorf("The second operand %v of %v is not int", yDataType, exprOp)
			}
			return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
		}
		if c.isUntypedVar(yDataType) {
			// is X compatible with Y?
			if !c.isIntegral(xType) {
				return nil, fmt.Errorf("The first operand %v of %v is not int", xDataType, exprOp)
			}
			return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
		}
		if !xType.equals(yType) {
			return nil, fmt.Errorf("%v can not be converted to %v for %v operation", xType.getOriginalType(), yType.getOriginalType(), exprOp)
		}
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	default:
		return nil, fmt.Errorf("binaryExprNumeric: Unsupported operation %v over %v.%v and %v.%v", exprOp, xType.oPkg, xType.oId, yType.oPkg, yType.oId)
	}
}

func (c *Config) binaryExprBooleans(exprOp token.Token, xType, yType *opr, xDataType, yDataType gotypes.DataType) (gotypes.DataType, error) {
	if exprOp != token.LAND && exprOp != token.LOR && exprOp != token.EQL && exprOp != token.NEQ {
		return nil, fmt.Errorf("binaryExprBooleans: Unsupported operation %v over booleans", exprOp)
	}

	resolveBool := func(exprOp token.Token, x, y string) string {
		switch exprOp {
		case token.LAND:
			if x == "true" && y == "true" {
				return "true"
			}
			return "false"
		case token.LOR:
			if x == "true" || y == "true" {
				return "true"
			}
			return "false"
		case token.EQL:
			if x == y {
				return "true"
			}
			return "false"
		case token.NEQ:
			if x == y {
				return "false"
			}
			return "true"
		default:
			panic(fmt.Errorf("unrecognized op %v", exprOp))
		}
	}

	if xType.untyped && yType.untyped {
		if xDataType.GetType() == gotypes.ConstantType && yDataType.GetType() == gotypes.ConstantType {
			return &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: resolveBool(exprOp, xType.literal, yType.literal)}, nil
		}
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	}
	// y is typed, either typed constant or typed identifier
	if xType.untyped {
		if yDataType.GetType() == gotypes.ConstantType {
			if xDataType.GetType() == gotypes.ConstantType {
				if exprOp == token.EQL || exprOp == token.NEQ {
					return &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: resolveBool(exprOp, xType.literal, yType.literal)}, nil
				}
				return &gotypes.Constant{Package: yType.oPkg, Def: yType.oId, Literal: resolveBool(exprOp, xType.literal, yType.literal)}, nil
			}
			if exprOp == token.EQL || exprOp == token.NEQ {
				return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
			}
			return &gotypes.Identifier{Package: yType.oPkg, Def: yType.oId}, nil
		}
		// xDataType has a type => no builtin
		return yDataType, nil
	}

	if yType.untyped {
		// x is typed, either typed constant or typed identifier
		if xDataType.GetType() == gotypes.ConstantType {
			if yDataType.GetType() == gotypes.ConstantType {
				if exprOp == token.EQL || exprOp == token.NEQ {
					return &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: resolveBool(exprOp, xType.literal, yType.literal)}, nil
				}
				return &gotypes.Constant{Package: xType.oPkg, Def: xType.oId, Literal: resolveBool(exprOp, xType.literal, yType.literal)}, nil
			}
			if exprOp == token.EQL || exprOp == token.NEQ {
				return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
			}
			return &gotypes.Identifier{Package: xType.oPkg, Def: xType.oId}, nil
		}
		// xDataType has a type => no builtin
		return xDataType, nil
	}

	// both operands are typed
	if xType.oId == yType.oId && xType.oPkg == yType.oPkg {
		// if both are constants, perform the operation, otherwise return the identifier
		if xDataType.GetType() == gotypes.ConstantType && yDataType.GetType() == gotypes.ConstantType {
			if exprOp == token.EQL || exprOp == token.NEQ {
				return &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: resolveBool(exprOp, xType.literal, yType.literal)}, nil
			}
			return &gotypes.Constant{Package: xType.oPkg, Def: xType.oId, Literal: resolveBool(exprOp, xType.literal, yType.literal)}, nil
		}
		if exprOp == token.EQL || exprOp == token.NEQ {
			return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
		}
		return &gotypes.Identifier{Package: xType.oPkg, Def: xType.oId}, nil
	}
	return nil, fmt.Errorf("binaryExprBooleans: Unsupported operation %v over %v.%v and %v.%v", exprOp, xType.oPkg, xType.oId, yType.oPkg, yType.oId)
}

func (c *Config) binaryExprStrings(exprOp token.Token, xType, yType *opr, xDataType, yDataType gotypes.DataType) (gotypes.DataType, error) {
	if xType.untyped && yType.untyped {
		if exprOp == token.ADD {
			return &gotypes.Constant{Package: "builtin", Def: "string", Untyped: true}, nil
		}
		return &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true}, nil
	}
	if xType.untyped {
		if yDataType.GetType() == gotypes.ConstantType {
			if exprOp == token.ADD {
				return &gotypes.Constant{Package: xType.oPkg, Def: xType.oId}, nil
			}
			return &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true}, nil
		}
		if exprOp == token.ADD {
			return yDataType, nil
		}
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	}
	if yType.untyped {
		if xDataType.GetType() == gotypes.ConstantType {
			if exprOp == token.ADD {
				return &gotypes.Constant{Package: yType.oPkg, Def: yType.oId}, nil
			}
			return &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true}, nil
		}
		if exprOp == token.ADD {
			return xDataType, nil
		}
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	}
	// both operands are typed
	if xType.equals(yType) {
		if exprOp == token.ADD {
			return &gotypes.Constant{Package: yType.oPkg, Def: yType.oId}, nil
		}
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	}
	return nil, fmt.Errorf("binaryExprStrings: Unsupported operation %v over %v.%v and %v.%v", exprOp, xType.oPkg, xType.oId, yType.oPkg, yType.oId)
}

func (c *Config) binaryExprStructs(exprOp token.Token, xType, yType *opr, xDataType, yDataType gotypes.DataType) (gotypes.DataType, error) {

	// structs  must be of the same type
	if xType.underlyingType.SymbolType != gotypes.StructType {
		return nil, fmt.Errorf("Expected %#v to be a struct", xDataType)
	}
	if yType.underlyingType.SymbolType != gotypes.StructType {
		return nil, fmt.Errorf("Expected %#v to be a struct", yDataType)
	}

	// both typed structs
	if xType.underlyingType.Id != "" && yType.underlyingType.Id != "" {
		if xType.underlyingType.Id != yType.underlyingType.Id || xType.underlyingType.Package != yType.underlyingType.Package {
			return nil, fmt.Errorf("Can not compare structs of two different types: %#v, %#v", xDataType, yDataType)
		}
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	}
	// at least one is anonymous => structs must have the same fields of the same type

	panic("binaryExprStructs: NYI")
}

func (c *Config) binaryExprInterfaces(exprOp token.Token, xType, yType *opr, xDataType, yDataType gotypes.DataType) (gotypes.DataType, error) {
	// if one of the operands is nil => bool right away
	if xType.pkg == "<nil>" && yType.pkg == "<nil>" {
		return nil, fmt.Errorf("mismatched types nil %v nil", exprOp)
	}

	if xType.pkg == "<nil>" || yType.pkg == "<nil>" {
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	}

	// two interfaces must be the same
	if xType.pkg == "<interface>" && yType.pkg == "<interface>" {
		if xType.underlyingType.Id != yType.underlyingType.Id {
			// Have both interface the same list of methods?
			// get definition of both interfaces
			var i1, i2 *gotypes.Interface
			if xDataType.GetType() == gotypes.InterfaceType {
				i1 = xDataType.(*gotypes.Interface)
			} else {
				sDef1, _, err1 := c.symbolsAccessor.LookupDataType(&gotypes.Identifier{Package: xType.underlyingType.Package, Def: xType.underlyingType.Id})
				if err1 != nil {
					return nil, err1
				}
				i1 = sDef1.Def.(*gotypes.Interface)
			}

			if yDataType.GetType() == gotypes.InterfaceType {
				i2 = yDataType.(*gotypes.Interface)
			} else {
				sDef2, _, err2 := c.symbolsAccessor.LookupDataType(&gotypes.Identifier{Package: yType.underlyingType.Package, Def: yType.underlyingType.Id})
				if err2 != nil {
					return nil, err2
				}
				i2 = sDef2.Def.(*gotypes.Interface)
			}

			if len(i1.Methods) != 0 && len(i2.Methods) != 0 {
				if !compatibility.InterfacesEqual(i1, i2) {
					return nil, fmt.Errorf("mismatched interface types %v.%v %v %v.%v", xType.underlyingType.Package, xType.underlyingType.Id, exprOp, yType.underlyingType.Package, yType.underlyingType.Id)
				}
			}
			return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
		}
		return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
	}

	// one of the operands is a data type
	// if the data type is not a pointer, it has to have pointer-less receivers
	// TODO(jchaloup): check if the data type implementes the interface

	//if xType.pkg
	//c.symbolsAccessor.LookupAllMethods(  )
	return &gotypes.Builtin{Def: "bool", Untyped: true}, nil
}

func (c *Config) BinaryExpr(exprOp token.Token, xDataType, yDataType gotypes.DataType) (gotypes.DataType, error) {

	xType, xE := c.getOperandType(xDataType)
	if xE != nil {
		return nil, xE
	}

	yType, yE := c.getOperandType(yDataType)
	if yE != nil {
		return nil, yE
	}

	glog.V(2).Infof("xType: %#v\n", xType)
	glog.V(2).Infof("yType: %#v\n", yType)

	// https://golang.org/ref/spec#Arithmetic_operators
	// Integer ops
	if c.areIntegral(xType, yType) {
		switch exprOp {
		case token.MUL, token.QUO, token.ADD, token.SUB, token.REM, token.SHL, token.SHR, token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR, token.AND, token.OR, token.AND_NOT, token.XOR:
			dt, err := c.binaryExprNumeric(exprOp, xType, yType, xDataType, yDataType)
			if err != nil {
				return nil, err
			}
			return dt, err
		default:
			return nil, fmt.Errorf("BinaryExpr: Unsupported operation %v over %v and %v", exprOp, xType.id, yType.id)
		}
	}

	// Floating ops
	if c.areFloating(xType, yType) {
		switch exprOp {
		case token.MUL, token.QUO, token.SUB, token.ADD, token.SHL, token.SHR, token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR, token.REM:
			dt, err := c.binaryExprNumeric(exprOp, xType, yType, xDataType, yDataType)
			if err != nil {
				return nil, err
			}
			return dt, err
		default:
			return nil, fmt.Errorf("areFloating: Unsupported operation %v over %v and %v", exprOp, xType.id, yType.id)
		}
	}

	// Complex ops
	if c.areComplex(xType, yType) {
		switch exprOp {
		case token.MUL, token.QUO, token.SUB, token.ADD, token.EQL, token.NEQ:
			dt, err := c.binaryExprNumeric(exprOp, xType, yType, xDataType, yDataType)
			if err != nil {
				return nil, err
			}
			return dt, err
		default:
			return nil, fmt.Errorf("areComplex: Unsupported operation %v over %v and %v", exprOp, xType.id, yType.id)
		}
	}

	// Boolean ops
	if c.areBooleans(xType, yType) {
		switch exprOp {
		case token.LAND, token.LOR, token.EQL, token.NEQ:
			dt, err := c.binaryExprBooleans(exprOp, xType, yType, xDataType, yDataType)
			if err != nil {
				return nil, err
			}
			return dt, err
		}
	}

	// String ops
	if c.areStrings(xType, yType) {
		switch exprOp {
		case token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR, token.ADD:
			dt, err := c.binaryExprStrings(exprOp, xType, yType, xDataType, yDataType)
			if err != nil {
				return nil, err
			}
			return dt, err
		}
	}

	// Pointers
	if c.arePointers(xType, yType) {
		switch exprOp {
		case token.EQL, token.NEQ:
			return &gotypes.Identifier{Package: "builtin", Def: "bool"}, nil
		}
	}

	// Structs
	if c.areStructs(xType, yType) {
		switch exprOp {
		case token.EQL, token.NEQ:
			return c.binaryExprStructs(exprOp, xType, yType, xDataType, yDataType)
		}
	}

	// interfaces
	if c.areInterfaces(xType, yType) {
		switch exprOp {
		case token.EQL, token.NEQ:
			return c.binaryExprInterfaces(exprOp, xType, yType, xDataType, yDataType)
		}
	}

	return nil, fmt.Errorf("General: Unsupported operation %v over %v.%v and %v.%v", exprOp, xType.oPkg, xType.oId, yType.oPkg, yType.oId)
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
	case token.SUB:
		// aplicable only for numbers
		xType, xE := c.getOperandType(xDataType)
		if xE != nil {
			return nil, xE
		}

		if c.isIntegral(xType) || c.isFloating(xType) || c.isComplex(xType) {
			ma := NewMultiArith()
			switch exprOp {
			case token.SUB:
				// ints, float and complex numbers allowed
				// Both X,Y are constants
				if xDataType.GetType() == gotypes.ConstantType {
					if c.isValueLess(xDataType) {
						return &gotypes.Constant{
							Package: xType.oPkg,
							Def:     xType.oId,
							Literal: "*",
							Untyped: xType.untyped,
						}, nil
					}

					ma.AddXFromString(xDataType.(*gotypes.Constant).Literal)
					if err := ma.PerformUnary(exprOp).Error(); err != nil {
						return nil, err
					}

					lit, err := ma.ZToLiteral(xType.id, false)
					if err != nil {
						return nil, err
					}
					return &gotypes.Constant{Package: xType.oPkg, Def: xType.oId, Literal: lit, Untyped: xType.untyped}, nil
				}
			}
		}
		// TODO(jchaloup): implement the remaining variants
		return xDataType, nil
	case token.XOR:
		var yDataType gotypes.DataType
		// assuming the data type is alway typed (^ needs to know the type to flip the right number of bits), e.g. ^uintptr(0)
		if xDataType.GetType() == gotypes.ConstantType {
			underlyingDataType, err := c.symbolsAccessor.ResolveToUnderlyingType(xDataType)
			if err != nil {
				return nil, fmt.Errorf("unable to resolve underlying type for %#v: %v", xDataType, err)
			}

			underlyingTargetType := "int"
			switch underlyingDataType.Def.GetType() {
			case gotypes.BuiltinType:
				underlyingTargetType = underlyingDataType.Def.(*gotypes.Builtin).Def
			case gotypes.IdentifierType:
				underlyingTargetType = underlyingDataType.Def.(*gotypes.Identifier).Def
			}

			constant := xDataType.(*gotypes.Constant)

			switch underlyingTargetType {
			case "int":
				// TODO(jchaloup): the architecture must be input of the processing
				p, _ := strconv.ParseInt(constant.Literal, 10, 64)
				o := ^int(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "int8":
				p, _ := strconv.ParseInt(constant.Literal, 10, 8)
				o := ^int8(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "int16":
				p, _ := strconv.ParseInt(constant.Literal, 10, 16)
				o := ^int16(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "int32":
				p, _ := strconv.ParseInt(constant.Literal, 10, 32)
				o := ^int32(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "int64":
				p, _ := strconv.ParseInt(constant.Literal, 10, 64)
				o := ^int64(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "uint":
				p, _ := strconv.ParseInt(constant.Literal, 10, 64)
				o := ^uint64(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "uint8":
				p, _ := strconv.ParseInt(constant.Literal, 10, 8)
				o := ^uint8(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "uint16":
				p, _ := strconv.ParseInt(constant.Literal, 10, 16)
				o := ^uint16(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "uint32":
				p, _ := strconv.ParseInt(constant.Literal, 10, 32)
				o := ^uint32(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "uint64":
				p, _ := strconv.ParseInt(constant.Literal, 10, 64)
				o := ^uint64(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "rune":
				p, _ := strconv.ParseInt(constant.Literal, 10, 32)
				o := ^int32(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "byte":
				p, _ := strconv.ParseInt(constant.Literal, 10, 8)
				o := ^uint8(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			case "uintptr":
				p, _ := strconv.ParseInt(constant.Literal, 10, 32)
				o := ^uintptr(p)
				yDataType = &gotypes.Constant{Package: constant.Package, Def: constant.Def, Literal: fmt.Sprintf("%v", o), Untyped: constant.Untyped}
			default:
				panic(fmt.Errorf("%v underlying type not recognized", constant.Def))
				return nil, fmt.Errorf("%v underlying type not recognized", constant.Def)
			}

			return yDataType, nil
		}
		return xDataType, nil
	case token.OR, token.NOT, token.ADD:
		// TODO(jchaloup): perform the operations
		return xDataType, nil
	default:
		return nil, fmt.Errorf("Unary operator %#v (%#v) not recognized", exprOp, token.ADD)
	}
}

func (c *Config) SelectorExpr(xDataType gotypes.DataType, item string) (*accessors.FieldAttribute, error) {
	// X.Sel a.k.a Prefix.Item

	glog.V(2).Infof("SelectorExpr: %#v, %v", xDataType, item)

	jsonMarshal := func(msg string, i interface{}) {
		byteSlice, _ := json.Marshal(i)
		glog.V(2).Infof("%v: %v\n", msg, string(byteSlice))
	}

	// The struct and an interface are the only data type from which a field/method is retriveable
	switch xType := xDataType.(type) {
	// If the X expression is a qualified id, the selector is a symbol from a package pointed by the id
	case *gotypes.Packagequalifier:
		glog.V(2).Infof("Trying to retrieve a symbol %#v from package %v\n", item, xType.Path)
		_, packageIdent, sType, err := c.symbolsAccessor.RetrieveQid(xType, &ast.Ident{Name: item})
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve a symbol table for %q package: %v", xType, err)
		}
		jsonMarshal("packageIdent", packageIdent)
		if sType == symbols.VariableSymbol {
			return &accessors.FieldAttribute{
				DataType: packageIdent.Def,
			}, nil
		} else {
			return &accessors.FieldAttribute{
				DataType: &gotypes.Identifier{Package: xType.Path, Def: item},
			}, nil
		}
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
	case *gotypes.Identifier, *gotypes.Constant:
		var ident *gotypes.Identifier
		if i, ok := xDataType.(*gotypes.Identifier); ok {
			ident = i
		} else {
			constant := xDataType.(*gotypes.Constant)
			ident = &gotypes.Identifier{
				Package: constant.Package,
				Def:     constant.Def,
			}
		}

		// Get struct/interface definition given by its identifier
		defSymbol, symbolTable, err := c.symbolsAccessor.LookupDataType(ident)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %v from the symbol table", ident.Def)
		}
		if defSymbol.Def == nil {
			return nil, fmt.Errorf("Trying to retrieve a field/method from a data type %#v that is not yet fully processed", ident)
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
				accessors.NewFieldAccessor(symbolTable, defSymbol, &ast.Ident{Name: item}),
			)
		case *gotypes.Selector:
			return c.symbolsAccessor.RetrieveDataTypeField(
				accessors.NewFieldAccessor(symbolTable, defSymbol, &ast.Ident{Name: item}),
			)
		default:
			// check data types with receivers
			glog.V(2).Infof("Retrieving method %q of a non-struct non-interface data type %#v", item, ident)
			def, err := c.symbolsAccessor.LookupMethod(ident, item)
			if err != nil {
				return nil, fmt.Errorf("Trying to retrieve a field/method %q from non-struct/non-interface data type: %#v: %v", item, defSymbol, err)
			}
			return &accessors.FieldAttribute{
				DataType: def.Def,
				IsMethod: true,
			}, nil
		}

	// anonymous struct
	case *gotypes.Struct:
		_, currentSt := c.symbolsAccessor.CurrentTable()
		return c.symbolsAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(currentSt, &symbols.SymbolDef{Def: xType}, &ast.Ident{Name: item}),
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
		return nil, fmt.Errorf("Trying to retrieve a %q field from a non-struct data type when parsing selector expression %#v", item, xDataType)
	}
}

func (c *Config) MakKeyExpr(xDataType gotypes.DataType) (gotypes.DataType, error) {
	// One can not do &(&(a))
	var indexExpr gotypes.DataType
	pointer, ok := xDataType.(*gotypes.Pointer)
	if ok {
		indexExpr = pointer.Def
	} else {
		indexExpr = xDataType
	}
	glog.V(2).Infof("IndexExprDef: %#v", indexExpr)

	nonIdentDef, err := c.symbolsAccessor.FindFirstNonidDataType(indexExpr)
	if err != nil {
		return nil, err
	}

	indexExpr = nonIdentDef

	mapDef, ok := indexExpr.(*gotypes.Map)
	if !ok {
		return nil, fmt.Errorf("MakKeyExpr expected map, got %#v instead", indexExpr)
	}
	return mapDef.Keytype, nil

}

func (c *Config) IndexExpr(xDataType, idxDataType gotypes.DataType) (gotypes.DataType, string, error) {
	// One can not do &(&(a))
	var indexExpr gotypes.DataType
	pointer, ok := xDataType.(*gotypes.Pointer)
	if ok {
		indexExpr = pointer.Def
	} else {
		indexExpr = xDataType
	}
	glog.V(2).Infof("IndexExprDef: %#v", indexExpr)
	// In case we have
	// type A []int
	// type B A
	// type C B
	// c := (C)([]int{1,2,3})
	// c[1]
	if indexExpr.GetType() == gotypes.IdentifierType || indexExpr.GetType() == gotypes.SelectorType {
		for {
			var symbolDef *symbols.SymbolDef
			if indexExpr.GetType() == gotypes.IdentifierType {
				xType := indexExpr.(*gotypes.Identifier)

				if xType.Package == "builtin" {
					break
				}
				def, _, err := c.symbolsAccessor.LookupDataType(xType)
				if err != nil {
					return nil, "", err
				}
				symbolDef = def
			} else {
				_, sd, err := c.symbolsAccessor.RetrieveQidDataType(indexExpr.(*gotypes.Selector).Prefix, &ast.Ident{Name: indexExpr.(*gotypes.Selector).Item})
				if err != nil {
					return nil, "", err
				}
				symbolDef = sd
			}

			if symbolDef.Def == nil {
				return nil, "", fmt.Errorf("Symbol %q not yet fully processed", symbolDef.Name)
			}

			indexExpr = symbolDef.Def
			if symbolDef.Def.GetType() == gotypes.IdentifierType || symbolDef.Def.GetType() == gotypes.SelectorType {
				continue
			}
			break
		}
	}
	// TODO(jchaloup): check idxDataType as well, .e.g. it is an integer type when accessing slice, array or string

	// Get definition of the X from the symbol Table (it must be a variable of a data type)
	// and get data type of its array/map members
	switch xType := indexExpr.(type) {
	case *gotypes.Map:
		return xType.Valuetype, indexExpr.GetType(), nil
	case *gotypes.Array:
		return xType.Elmtype, indexExpr.GetType(), nil
	case *gotypes.Slice:
		return xType.Elmtype, indexExpr.GetType(), nil
	case *gotypes.Builtin:
		if xType.Def == "string" {
			// Checked at https://play.golang.org/
			return &gotypes.Identifier{Package: "builtin", Def: "uint8"}, gotypes.BuiltinType, nil
		}
		return nil, "", fmt.Errorf("Accessing item of built-in non-string type: %#v", xType)
	case *gotypes.Identifier:
		if xType.Def == "string" && xType.Package == "builtin" {
			// Checked at https://play.golang.org/
			return &gotypes.Identifier{Package: "builtin", Def: "uint8"}, gotypes.BuiltinType, nil
		}
		return nil, "", fmt.Errorf("Accessing item of built-in non-string type: %#v", xType)
	case *gotypes.Constant:
		// only for strings, other constant types can not be indexed
		if xType.Def == "string" && xType.Package == "builtin" {
			// Checked at https://play.golang.org/
			return &gotypes.Identifier{Package: "builtin", Def: "uint8"}, gotypes.BuiltinType, nil
		}
		return nil, "", fmt.Errorf("Accessing item of built-in non-string type: %#v", xType)
	case *gotypes.Ellipsis:
		return xType.Def, indexExpr.GetType(), nil
	default:
		panic(fmt.Errorf("Unrecognized indexExpr type: %#v", xDataType))
	}
}

func (c *Config) RangeExpr(xDataType gotypes.DataType) (gotypes.DataType, gotypes.DataType, error) {
	var rangeExpr gotypes.DataType
	// over-approximation but given we run the go build before the procesing
	// this is a valid processing
	pointer, ok := xDataType.(*gotypes.Pointer)
	if ok {
		rangeExpr = pointer.Def
	} else {
		rangeExpr = xDataType
	}

	// Identifier or a qid.Identifier
	var err error
	rangeExpr, err = c.symbolsAccessor.FindFirstNonidDataType(rangeExpr)
	if err != nil {
		return nil, nil, err
	}

	// From https://golang.org/ref/spec#For_range
	//
	// Range expression                          1st value          2nd value
	//
	// array or slice  a  [n]E, *[n]E, or []E    index    i  int    a[i]       E
	// string          s  string type            index    i  int    see below  rune
	// map             m  map[K]V                key      k  K      m[k]       V
	// channel         c  chan E, <-chan E       element  e  E
	switch xExprType := rangeExpr.(type) {
	case *gotypes.Array:
		return &gotypes.Builtin{Def: "int"}, xExprType.Elmtype, nil
	case *gotypes.Slice:
		return &gotypes.Builtin{Def: "int"}, xExprType.Elmtype, nil
	case *gotypes.Builtin:
		if xExprType.Def != "string" {
			fmt.Errorf("Expecting string in range Builtin expression. Got %#v instead.", xDataType)
		}
		return &gotypes.Builtin{Def: "int"}, &gotypes.Builtin{Def: "rune"}, nil
	case *gotypes.Identifier:
		if xExprType.Package != "builtin" || xExprType.Def != "string" {
			fmt.Errorf("Expecting string in range Identifier expression. Got %#v instead.", xDataType)
		}
		return &gotypes.Builtin{Def: "int"}, &gotypes.Builtin{Def: "rune"}, nil
	case *gotypes.Constant:
		if xExprType.Package != "builtin" || xExprType.Def != "string" {
			fmt.Errorf("Expecting string in range Constant expression. Got %#v instead.", xDataType)
		}
		return &gotypes.Builtin{Def: "int"}, &gotypes.Builtin{Def: "rune"}, nil
	case *gotypes.Map:
		return xExprType.Keytype, xExprType.Valuetype, nil
	case *gotypes.Channel:
		return xExprType.Value, nil, nil
	case *gotypes.Ellipsis:
		return &gotypes.Builtin{Def: "int"}, xExprType.Def, nil
	default:
		return nil, nil, fmt.Errorf("Unknown type of range expression: %#v", rangeExpr)
	}
}

func (c *Config) TypecastExpr(xDataType, tDataType gotypes.DataType) (gotypes.DataType, error) {
	glog.V(2).Infof("TypecastExpr, xDataType: %#v, tDataType: %#v", xDataType, tDataType)

	if constant, ok := xDataType.(*gotypes.Constant); ok {
		// if the tDataType is interface => no constant
		glog.V(2).Infof("Checking if IsDataTypeInterface(%#v)", tDataType)
		isI, err := c.symbolsAccessor.IsDataTypeInterface(tDataType)
		glog.V(2).Infof("Checking if IsDataTypeInterface(%#v) is %v, err: %v", tDataType, isI, err)
		if err != nil {
			return nil, err
		}

		glog.V(2).Infof("Checking if IsDataTypeInterface(%#v) is %v", tDataType, isI)

		if isI {
			// no more a constant
			return tDataType, nil
		}

		ident, err := c.symbolsAccessor.TypeToSimpleBuiltin(tDataType)
		if err != nil {
			return nil, err
		}

		if ident.Package == "<builtin>" && ident.Def == "pointer" {
			cIdent, err := c.symbolsAccessor.TypeToSimpleBuiltin(&gotypes.Identifier{Package: constant.Package, Def: constant.Def})
			if err != nil {
				return nil, err
			}
			if cIdent.Package == "builtin" && cIdent.Def == "string" {
				// []byte(str)
				nonIdent, err := c.symbolsAccessor.FindFirstNonidDataType(tDataType)
				if err != nil {
					return nil, err
				}
				slice, ok := nonIdent.(*gotypes.Slice)
				if !ok {
					return nil, fmt.Errorf("Expected []byte, got %#v instead", nonIdent)
				}
				ident, ok := slice.Elmtype.(*gotypes.Identifier)
				if !ok {
					return nil, fmt.Errorf("Expected []byte, got slice of %#v instead", slice.Elmtype)
				}
				if ident.Package == "builtin" && (ident.Def == "byte" || ident.Def == "rune") {
					return tDataType, nil
				}

				return nil, fmt.Errorf("Expected []byte, got slice of %v.%v instead", ident.Package, ident.Def)
			}
			// only if the constant type is uintptr
			if constant.Package != "builtin" || constant.Def != "uintptr" {
				return nil, fmt.Errorf("Casting a non uintptr constant %#v to a pointer type %#v", xDataType, tDataType)
			}
			return tDataType, nil
		}

		// tDataType must be identifier
		if tDataType.GetType() != gotypes.IdentifierType && tDataType.GetType() != gotypes.SelectorType {
			return nil, fmt.Errorf("Expected a casted type %#v of a constant to be an identifier. Got %v instead", tDataType, tDataType.GetType())
		}

		// typed constant, we must check the types are compatible (during the contract evaluation)
		bIdent, err := c.symbolsAccessor.TypeToSimpleBuiltin(tDataType)
		if err != nil {
			return nil, err
		}

		if bIdent.Package == "builtin" && c.symbolsAccessor.IsIntegral(bIdent.Def) {
			// is valueless literal?
			if constant.Literal != "*" {
				ma := NewMultiArith()
				ma.AddXFromString(constant.Literal)
				lit, err := ma.XToLiteral(bIdent.Def)

				if err != nil {
					return nil, err
				}
				constant.Literal = lit
			}
		}
		if tDataType.GetType() == gotypes.IdentifierType {
			typeDefIdent := tDataType.(*gotypes.Identifier)
			return &gotypes.Constant{
				Def:     typeDefIdent.Def,
				Package: typeDefIdent.Package,
				Literal: constant.Literal,
			}, nil
		}
		typeDefIdent := tDataType.(*gotypes.Selector)
		return &gotypes.Constant{
			Def:     typeDefIdent.Item,
			Package: typeDefIdent.Prefix.(*gotypes.Packagequalifier).Path,
			Literal: constant.Literal,
		}, nil
	}

	return tDataType, nil
}

func (c *Config) BuiltinFunctionInvocation(name string, arguments []gotypes.DataType) ([]gotypes.DataType, error) {
	glog.V(2).Infof("Processing builtin %v function", name)
	switch name {
	case "imag", "real":
		if len(arguments) != 1 {
			return nil, fmt.Errorf("imag argument is not a single expression")
		}
		if constant, ok := arguments[0].(*gotypes.Constant); ok {
			fmt.Printf("Constant: %#v\n", constant)
			panic("CCC")
		}
		// variable => accepting only complex64 or complex128
		xType, xE := c.getOperandType(arguments[0])
		if xE != nil {
			return nil, xE
		}

		if xType.id == "complex64" {
			return []gotypes.DataType{
				&gotypes.Identifier{Package: "builtin", Def: "float32"},
			}, nil
		}
		if xType.id == "complex128" {
			return []gotypes.DataType{
				&gotypes.Identifier{Package: "builtin", Def: "float64"},
			}, nil
		}
		return nil, fmt.Errorf("invalid argument %v for %v", xType.getOriginalType(), name)
	case "complex":
		if len(arguments) != 2 {
			return nil, fmt.Errorf("Expected two arguments of the complex function, got %v instead", len(arguments))
		}
		// accepting up to float types
		realType, rE := c.getOperandType(arguments[0])
		if rE != nil {
			return nil, rE
		}
		imagType, iE := c.getOperandType(arguments[1])
		if iE != nil {
			return nil, iE
		}

		// if !c.isFloating(realType) || !c.isFloating(imagType) {
		// 	return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
		// }
		// both constants => return constant
		if c.isConstant(arguments[0]) && c.isConstant(arguments[1]) {
			var targetType string
			if realType.untyped && imagType.untyped {
				if !c.isCompatibleWithFloat(realType, arguments[0]) || !c.isCompatibleWithFloat(realType, arguments[1]) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				targetType = "complex128"
			} else if realType.untyped {
				if !c.isCompatibleWithFloat(realType, arguments[0]) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				if !c.isFloating(imagType) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				if _, err := NewMultiArith().AddXFromString(arguments[0].(*gotypes.Constant).Literal).XToLiteral(imagType.id); err != nil {
					return nil, err
				}
				targetType = imagType.id
			} else if imagType.untyped {
				if !c.isCompatibleWithFloat(imagType, arguments[1]) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				if !c.isFloating(realType) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				if _, err := NewMultiArith().AddXFromString(arguments[1].(*gotypes.Constant).Literal).XToLiteral(realType.id); err != nil {
					return nil, err
				}
				targetType = realType.id
			} else {
				if !realType.equals(imagType) || !c.isFloating(realType) || !c.isFloating(imagType) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				targetType = realType.id
			}

			if targetType == "float32" {
				targetType = "complex64"
			} else {
				targetType = "complex128"
			}

			// untyped complex
			// create a complex number
			cMa := NewMultiArith()
			cMa.AddXFromString("1i")
			cMa.AddYFromString(arguments[1].(*gotypes.Constant).Literal)
			if err := cMa.Perform(token.MUL).Error(); err != nil {
				return nil, err
			}
			imagLit, err := cMa.ZToLiteral(targetType, false)
			if err != nil {
				return nil, err
			}

			ma := NewMultiArith()
			ma.AddXFromString(arguments[0].(*gotypes.Constant).Literal)
			ma.AddYFromString(imagLit)

			if err := ma.Perform(token.ADD).Error(); err != nil {
				return nil, err
			}

			lit, err := ma.ZToLiteral(targetType, false)
			if err != nil {
				return nil, err
			}
			return []gotypes.DataType{
				&gotypes.Constant{Package: "builtin", Def: targetType, Literal: lit, Untyped: true},
			}, nil
		}
		// at least one operand is a variable
		if c.isConstant(arguments[0]) {
			if !c.isFloating(imagType) {
				return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
			}
			if realType.untyped {
				if !c.isCompatibleWithFloat(realType, arguments[0]) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				if _, err := NewMultiArith().AddXFromString(arguments[0].(*gotypes.Constant).Literal).XToLiteral(imagType.id); err != nil {
					return nil, err
				}
			} else {
				if !realType.equals(imagType) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
			}
			var targetType string
			if imagType.id == "float32" {
				targetType = "complex64"
			} else {
				targetType = "complex128"
			}
			return []gotypes.DataType{
				&gotypes.Identifier{Package: "builtin", Def: targetType},
			}, nil
		}
		if c.isConstant(arguments[1]) {
			if !c.isFloating(realType) {
				return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
			}
			if imagType.untyped {
				if !c.isCompatibleWithFloat(imagType, arguments[1]) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
				if _, err := NewMultiArith().AddXFromString(arguments[1].(*gotypes.Constant).Literal).XToLiteral(realType.id); err != nil {
					return nil, err
				}
			} else {
				if !realType.equals(imagType) {
					return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
				}
			}
			var targetType string
			if realType.id == "float32" {
				targetType = "complex64"
			} else {
				targetType = "complex128"
			}
			return []gotypes.DataType{
				&gotypes.Identifier{Package: "builtin", Def: targetType},
			}, nil
		}
		if !realType.equals(imagType) || !c.isFloating(realType) {
			return nil, fmt.Errorf("mismatched types complex(%v, %v)", realType.getOriginalType(), imagType.getOriginalType())
		}
		var targetType string
		if realType.id == "float32" {
			targetType = "complex64"
		} else {
			targetType = "complex128"
		}
		return []gotypes.DataType{
			&gotypes.Identifier{Package: "builtin", Def: targetType},
		}, nil
	}
	panic("PPPP")
}
