package typevars

import (
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/symbols"
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
var RangeKeyType Type = "RangeKey"
var RangeValueType Type = "RangeValue"

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
	Pos     string
	Package string
}

func (v *Variable) GetType() Type {
	return VariableType
}

func (v *Variable) String() string {
	return fmt.Sprintf("%v#%v#%v", v.Package, v.Name, v.Pos)
}

func VariableFromSymbolDef(def *symbols.SymbolDef) *Variable {
	return &Variable{
		Name:    def.Name,
		Pos:     def.Pos,
		Package: def.Package,
	}
}

func FunctionFromSymbolDef(def *symbols.SymbolDef) *Function {
	return &Function{
		Name:    def.Name,
		Pos:     def.Pos,
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
		return fmt.Sprintf("TypeVar.ListKey: integer type")
	case *ListValue:
		return fmt.Sprintf("TypeVar.ListValue: %#v", d.Interface)
	case *MapKey:
		return fmt.Sprintf("TypeVar.MapKey: %#v", d.Interface)
	case *MapValue:
		return fmt.Sprintf("TypeVar.MapValue: %#v", d.Interface)
	case *RangeKey:
		return fmt.Sprintf("TypeVar.RangeKey: %#v", d.Interface)
	case *RangeValue:
		return fmt.Sprintf("TypeVar.RangeValue: %#v", d.Interface)
	case *Function:
		return fmt.Sprintf("TypeVar.Function: (%v) %v", d.Package, d.Name)
	case *ReturnType:
		return fmt.Sprintf("TypeVar.ReturnType: (%v) at %v", TypeVar2String(d.Function), d.Index)
	case *Argument:
		return fmt.Sprintf("TypeVar.Argument: (%v) at %v", TypeVar2String(d.Function), d.Index)
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
	Interface
}

func (l *ListValue) GetType() Type {
	return ListValueType
}

type MapKey struct {
	Interface
}

func (m *MapKey) GetType() Type {
	return MapKeyType
}

type MapValue struct {
	Interface
}

func (m *MapValue) GetType() Type {
	return MapValueType
}

type RangeKey struct {
	Interface
}

func (m *RangeKey) GetType() Type {
	return RangeKeyType
}

type RangeValue struct {
	Interface
}

func (m *RangeValue) GetType() Type {
	return RangeValueType
}

type Function struct {
	Package string
	Name    string
	Pos     string
}

func (f *Function) String() string {
	return fmt.Sprintf("%v#%v#%v", f.Package, f.Name, f.Pos)
}

func (v *Function) GetType() Type {
	return FunctionType
}

// Argument represent an address of a function/method argument
type Argument struct {
	// Location of a function
	Function Interface
	// Return type position
	Index int
}

func (a *Argument) GetType() Type {
	return ArgumentType
}

type ReturnType struct {
	// Location of a function
	Function Interface
	// Return type position
	Index int
}

func (a *ReturnType) GetType() Type {
	return ReturnTypeType
}

func MakeVar(packageName, name, pos string) *Variable {
	return &Variable{
		Package: packageName,
		Name:    name,
		Pos:     pos,
	}
}

func MakeLocalVar(name, pos string) *Variable {
	return &Variable{
		Name: name,
		Pos:  pos,
	}
}

func MakeVirtualVar(index int) *Variable {
	return &Variable{
		Name: fmt.Sprintf("virtual.var.%v", index),
	}
}

func MakeConstant(datatype gotypes.DataType) *Constant {
	return &Constant{
		DataType: datatype,
	}
}

func MakeFunction(packageName, name, pos string) *Function {
	return &Function{
		Package: packageName,
		Name:    name,
		Pos:     pos,
	}
}

func MakeVirtualFunction(v *Variable) *Function {
	return &Function{
		Name:    v.Name,
		Package: v.Package,
	}
}

func MakeArgument(i Interface, index int) *Argument {
	return &Argument{
		Function: i,
		Index:    index,
	}
}

func MakeReturn(i Interface, index int) *ReturnType {
	return &ReturnType{
		Function: i,
		Index:    index,
	}
}

func MakeListKey() *ListKey {
	return &ListKey{}
}

func MakeListValue(i Interface) *ListValue {
	return &ListValue{
		Interface: i,
	}
}

func MakeConstantListValue(datatype gotypes.DataType) *ListValue {
	return &ListValue{
		Interface: MakeConstant(datatype),
	}
}

func MakeMapKey(i Interface) *MapKey {
	return &MapKey{
		Interface: i,
	}
}

func MakeConstantMapKey(datatype gotypes.DataType) *MapKey {
	return &MapKey{
		Interface: MakeConstant(datatype),
	}
}

func MakeMapValue(i Interface) *MapValue {
	return &MapValue{
		Interface: i,
	}
}

func MakeRangeKey(i Interface) *RangeKey {
	return &RangeKey{
		Interface: i,
	}
}

func MakeRangeValue(i Interface) *RangeValue {
	return &RangeValue{
		Interface: i,
	}
}

func MakeConstantMapValue(datatype gotypes.DataType) *MapValue {
	return &MapValue{
		Interface: MakeConstant(datatype),
	}
}
func MakeField(i Interface, field string, index int) *Field {
	return &Field{
		Interface: i,
		Name:      field,
		Index:     index,
	}
}
