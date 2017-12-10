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
var MapKeyType Type = "MapKey"
var MapValueType Type = "MapValue"
var ListKeyType Type = "ListKey"
var ListValueType Type = "ListValue"
var FieldType Type = "Field"
var CGOType Type = "CGO"

// Constant to represent:
// - basic literal type
// - composite literal type
// - type casting
// - struct type expression
type Constant struct {
	gotypes.DataType
	Package string
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

func FunctionFromSymbolDef(def *symboltable.SymbolDef) *Function {
	return &Function{
		Name:    fmt.Sprintf("%v:%v", def.Pos, def.Name),
		Package: def.Package,
	}
}

type Field struct {
	Interface
	Name  string
	Index int
}

func (f *Field) GetType() Type {
	return FieldType
}

type CGO struct{}

func (c *CGO) GetType() Type {
	return CGOType
}

func TypeVar2String(tv Interface) string {
	switch d := tv.(type) {
	case *Constant:
		return fmt.Sprintf("TypeVar.Constant: %#v", d.DataType)
	case *Variable:
		return fmt.Sprintf("TypeVar.Variable: (%v) %v", d.Package, d.Name)
	case *ListKey:
		return fmt.Sprintf("TypeVar.ListKey: int")
	case *ListValue:
		return fmt.Sprintf("TypeVar.ListValue: %#v", d.Constant.DataType)
	case *MapKey:
		return fmt.Sprintf("TypeVar.MapKey: %#v", d.Constant.DataType)
	case *MapValue:
		return fmt.Sprintf("TypeVar.MapValue: %#v", d.Constant.DataType)
	case *Function:
		return fmt.Sprintf("TypeVar.Function: (%v) %v", d.Package, d.Name)
	case *ReturnType:
		return fmt.Sprintf("TypeVar.ReturnType: (%v) %v at %v", d.Package, d.Name, d.Index)
	case *Argument:
		return fmt.Sprintf("TypeVar.Argument: (%v) %v at %v", d.Package, d.Name, d.Index)
	case *Field:
		if d.Name == "" {
			return fmt.Sprintf("TypeVar.Field: %#v at index %v", d.Interface, d.Index)
		}
		return fmt.Sprintf("TypeVar.Field: %#v with field %q", d.Interface, d.Name)
	case *CGO:
		return fmt.Sprintf("TypeVar.CGO")
	default:
		fmt.Printf("\nTypeVar %#v\n\n", tv)
		panic("Unrecognized TypeVar")
	}
}

type ListKey struct{}

func (l *ListKey) GetType() Type {
	return ListKeyType
}

type ListValue struct {
	Constant
}

func (l *ListValue) GetType() Type {
	return ListValueType
}

type MapKey struct {
	Constant
}

func (m *MapKey) GetType() Type {
	return MapKeyType
}

type MapValue struct {
	Constant
}

func (m *MapValue) GetType() Type {
	return MapValueType
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

func MakeVar(name, packageName string) *Variable {
	return &Variable{
		Name:    name,
		Package: packageName,
	}
}

func MakeVirtualVar(index int) *Variable {
	return MakeVar(fmt.Sprintf("virtual.var.%v", index), "")
}

func MakeConstant(datatype gotypes.DataType) *Constant {
	return &Constant{
		DataType: datatype,
	}
}

func MakeFunction(name, packageName string) *Function {
	return &Function{
		Name:    name,
		Package: packageName,
	}
}

func MakeVirtualFunction(v *Variable) *Function {
	return &Function{
		Name:    v.Name,
		Package: v.Package,
	}
}

func MakeArgument(name, packageName string, index int) *Argument {
	return &Argument{
		Function: *MakeFunction(name, packageName),
		Index:    index,
	}
}

func MakeReturn(name, packageName string, index int) *ReturnType {
	return &ReturnType{
		Function: *MakeFunction(name, packageName),
		Index:    index,
	}
}

func MakeListKey() *ListKey {
	return &ListKey{}
}

func MakeListValue(datatype gotypes.DataType) *ListValue {
	return &ListValue{
		Constant: *MakeConstant(datatype),
	}
}

func MakeMapKey(datatype gotypes.DataType) *MapKey {
	return &MapKey{
		Constant: *MakeConstant(datatype),
	}
}

func MakeMapValue(datatype gotypes.DataType) *MapValue {
	return &MapValue{
		Constant: *MakeConstant(datatype),
	}
}

func MakeField(i Interface, field string, index int) *Field {
	return &Field{
		Interface: i,
		Name:      field,
		Index:     index,
	}
}
