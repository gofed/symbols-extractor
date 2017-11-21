package contract

import (
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

// The data type propagation can built up to an acyclic graph in general.
// Each node of the propagation graph makes a contract with a Go statement
// or with a Go expression. Recognized contracts:
// - assigment
// - unary or binary expression
// -

const (
	FunctionType = "function"
	LiteralType = "literal"
	BinaryOpType = "binaryOp"
)

// Contract between a data type and its application
type Contract interface {
	GetType() string
}

// Common parts for all contracts
type CommonData struct {
	// The name of package where contract was made.
	Package            string
	// The expected type for this contract.
	ExpectedType       gotypes.DataType
	// Is true if the expected data type for this contract was derived
	// according to the Golang rules.
	DataTypeWasDerived bool
}

// Contract for function definitions/declarations.
//
// From a Golang point of view
//
//   f := func() {}	and	func f() {}
//
// are different constructs. Both introduces a new symbol f, but the ways are
// different and we want to keep these ways.
type Function struct {
	*CommonData
	Name string // function name
}

func (f *Function) GetType() string {
	return FunctionType
}

// Contract for literals
//
// TODO(jkucera): Maybe split to BasicLit, CompositeLit and FuncLit
type Literal struct {
	*CommonData
	// TODO(jchaloup): add additional flags, e.g. DataTypeForced, FunctionValue
}

func (l *Literal) GetType() string {
	return LiteralType
}

type BinaryOp struct {
	X, Y Contract
}

func (c *BinaryOp) GetType() string {
	return BinaryOpType
}
