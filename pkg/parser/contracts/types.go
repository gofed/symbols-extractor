package contracts

import (
	"fmt"
	"go/token"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type Type string

type Contract interface {
	GetType() Type
}

var BinaryOpType Type = "binaryop"
var UnaryOpType Type = "unaryop"
var PropagatesToType Type = "propagatesto"
var IsCompatibleWithType Type = "iscompatiblewith"
var IsInvocableType Type = "isinvocable"
var HasFieldType Type = "hasfield"

func Contract2String(c Contract) string {
	switch d := c.(type) {
	case *BinaryOp:
		return fmt.Sprintf("BinaryOpContract:\n\tX=%v,\n\tY=%v,\n\tZ=%v,\n\top=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), typevars.TypeVar2String(d.Z), d.OpToken)
	case *UnaryOp:
		return fmt.Sprintf("UnaryOpContract:\n\tX=%v,\n\tY=%v,\n\top=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.OpToken)
	case *PropagatesTo:
		return fmt.Sprintf("PropagatesTo:\n\tX=%v,\n\tY=%v,\n\tE=%#v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.ExpectedType)
	case *IsCompatibleWith:
		return fmt.Sprintf("IsCompatibleWith:\n\tX=%v\n\tY=%v\n\tE=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.ExpectedType)
	case *IsInvocable:
		return fmt.Sprintf("IsInvocable:\n\tF=%v,\n\targCount=%v", typevars.TypeVar2String(d.F), d.ArgsCount)
	case *HasField:
		return fmt.Sprintf("HasField:\n\tX=%v,\n\tField=%v,\n\tIndex=%v", typevars.TypeVar2String(&d.X), d.Field, d.Index)
	default:
		panic(fmt.Sprintf("Contract %#v not recognized", c))
	}
}

// BinaryOp represents contract between two typevars
type BinaryOp struct {
	// OpToken gives information about particular binary operation.
	// E.g. '+' can be used with integers and strings, '-' can not be used with strings.
	// As long as the operands are compatible with the operation, the contract holds.
	OpToken token.Token
	// Z = X op Y
	X, Y typevars.Interface
	Z    typevars.Interface
	// TODO(jchaloup): add expected type
}

type UnaryOp struct {
	OpToken token.Token
	// Y = op X
	X, Y typevars.Interface
	// TODO(jchaloup): add expected type
}

type PropagatesTo struct {
	X, Y         typevars.Interface
	ExpectedType gotypes.DataType
}

type IsCompatibleWith struct {
	X, Y         typevars.Interface
	ExpectedType gotypes.DataType
}

type IsInvocable struct {
	F         typevars.Interface
	ArgsCount int
}

type HasField struct {
	X     typevars.Variable
	Field string
	Index int
}

func (b *BinaryOp) GetType() Type {
	return BinaryOpType
}

func (b *UnaryOp) GetType() Type {
	return UnaryOpType
}

func (p *PropagatesTo) GetType() Type {
	return PropagatesToType
}

func (i *IsCompatibleWith) GetType() Type {
	return IsCompatibleWithType
}

func (i *IsInvocable) GetType() Type {
	return IsInvocableType
}

func (i *HasField) GetType() Type {
	return HasFieldType
}
