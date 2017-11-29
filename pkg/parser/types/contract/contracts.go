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
	AssignmentType = "assigment"
)

// Contract between a data type and its application
type Contract interface {
	// Get the contract type (kind)
	GetType() string
}

// Common parts for all contracts
type CommonData struct {
	// The name of package where contract was made
	Package            string
	// The expected type for this contract. ExpectedType is computed once
	// during parsing (during the time when a contract is first made) and
	// it is persistently saved for later use
	// TODO(jkucera): We use `gotypes.DataType` for now, but this is not
	// good for later marshalling as there is a danger of large data amount
	// as the one (possibly composite) data type can be marshalled many
	// times. Suggestions:
	// a) custom marshalling/unmarshalling
	//    - use a section in json file that represents a map with keys as
	//      data type hash and value as marshalled data type definition
	//    - use these hashes in place of data types
	ExpectedType       gotypes.DataType
	// Is true if the expected data type for this contract was derived
	// according to the Golang rules.
	DataTypeWasDerived bool
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

// Contract for a declaration/assignment
// TODO(jkucera): Assignment to field, assignment to container member
type Assignment struct {
	CommonData
	// Contract of right-hand side expression
	Parent  Contract
	// Name of a symbol
	Name    string
	// True if it is a declarative assignment
	IsDecl  bool
	// True if it is a constant declaration
	IsConst bool
}

func (c *Assignment) GetType() string {
	return AssignmentType
}
