package typevars

import (
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type Interface interface {
	GetType() Type
}

type Type string

var ConstantType Type = "Constant"
var VariableType Type = "Variable"
var FunctionType Type = "Function"
var ArgumentType Type = "Argument"
var ReturnTypeType Type = "ReturnType"

// Constant to represent:
// - basic literal type
// - composite literal type
// - type casting
// - struct type expression
type Constant struct {
	gotypes.DataType
}

func (c *Constant) GetType() Type {
	return ConstantType
}

// Variable to represent:
// - local variable: usually in <file>:<ln>:<name> form
// - global variable
// - qid-ed variable
type Variable struct {
	Name    string
	Package string
}

func (v *Variable) GetType() Type {
	return VariableType
}

func VariableFromSymbolDef(def *symboltable.SymbolDef) *Variable {
	return &Variable{
		Name:    fmt.Sprintf("%v:%v", def.Pos, def.Name),
		Package: def.Package,
	}
}

func TypeVar2String(tv Interface) string {
	switch d := tv.(type) {
	case *Constant:
		return fmt.Sprintf("TypeVar.Constant: %#v", d.DataType)
	case *Variable:
		return fmt.Sprintf("TypeVar.Variable: (%v) %v", d.Package, d.Name)
	case *Function:
		return fmt.Sprintf("TypeVar.Function: (%v) %v", d.Package, d.Name)
	case *ReturnType:
		return fmt.Sprintf("TypeVar.ReturnType: (%v) %v at %v", d.Package, d.Name, d.Index)
	case *Argument:
		return fmt.Sprintf("TypeVar.Argument: (%v) %v at %v", d.Package, d.Name, d.Index)
	default:
		fmt.Printf("\nTypeVar %#v\n\n", tv)
		panic("Unrecognized TypeVar")
	}
}

type Function struct {
	Name    string
	Package string
}

func (v *Function) GetType() Type {
	return FunctionType
}

// Argument represent an address of a function/method argument
type Argument struct {
	// Location of a function
	Function
	// Return type position
	Index int
}

func (a *Argument) GetType() Type {
	return ArgumentType
}

type ReturnType struct {
	// Location of a function
	Function
	// Return type position
	Index int
}

func (a *ReturnType) GetType() Type {
	return ReturnTypeType
}
