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
var ArgumentType Type = "Argument"

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
	default:
		fmt.Printf("\nTypeVar %#v\n\n", tv)
		panic("Unrecognized TypeVar")
	}
}

// Argument represent an address of a function/method argument
type Argument struct {
}

func (a *Argument) GetType() Type {
	return ArgumentType
}
