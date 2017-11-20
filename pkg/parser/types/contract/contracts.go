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
	ResourceType = "resource"
	BinaryOpType = "binaryOp"
)

// Contract between a data type and its application
type Contract interface {
	GetType() string
}

type Resource struct {
	gotypes.DataType
	Package string
	Name    string
	// TODO(jchaloup): add additional flags, e.g. DataTypeForced, FunctionValue
}

func (r *Resource) GetType() string {
	return ResourceType
}

type BinaryOp struct {
	X, Y Contract
}

func (c *BinaryOp) GetType() string {
	return BinaryOpType
}
