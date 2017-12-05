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
var PropagatesToType Type = "propagatesto"
var IsCompatibleWithType Type = "iscompatiblewith"

func Contract2String(c Contract) string {
	switch d := c.(type) {
	case *BinaryOp:
		return fmt.Sprintf("BinaryOpContract:\n\tX=%v,\n\tY=%v,\n\tZ=%v,\n\top=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), typevars.TypeVar2String(d.Z), d.OpToken)
	case *PropagatesTo:
		return fmt.Sprintf("PropagatesTo:\n\tX=%v,\n\tY=%v,\n\tE=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.ExpectedType)
	case *IsCompatibleWith:
		return fmt.Sprintf("IsCompatibleWith:\n\tX=%v\n\tY=%v\n\tE=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.ExpectedType)
	}
	return ""
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

func (b *BinaryOp) GetType() Type {
	return BinaryOpType
}

type PropagatesTo struct {
	X, Y         typevars.Interface
	ExpectedType gotypes.DataType
}

func (p *PropagatesTo) GetType() Type {
	return PropagatesToType
}

type IsCompatibleWith struct {
	X, Y         typevars.Interface
	ExpectedType gotypes.DataType
}

func (i *IsCompatibleWith) GetType() Type {
	return IsCompatibleWithType
}
