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

type Field struct {
	X     *Variable
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
		return fmt.Sprintf("TypeVar.Constant: %#v, Package: %v", d.DataType, d.Package)
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
	case *ReturnType:
		return fmt.Sprintf("TypeVar.ReturnType: (%v) at %v", TypeVar2String(d.Function), d.Index)
	case *Argument:
		return fmt.Sprintf("TypeVar.Argument: (%v) at %v", TypeVar2String(d.Function), d.Index)
	case *Field:
		if d.Name == "" {
			return fmt.Sprintf("TypeVar.Field: %#v at index %v", d.X, d.Index)
		}
		return fmt.Sprintf("TypeVar.Field: %#v with field %q", d.X, d.Name)
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

// Argument represent an address of a function/method argument
type Argument struct {
	// Location of a function
	Function *Variable
	// Return type position
	Index int
}

func (a *Argument) GetType() Type {
	return ArgumentType
}

type ReturnType struct {
	// Location of a function
	Function *Variable
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

func MakeConstant(pkg string, datatype gotypes.DataType) *Constant {
	return &Constant{
		DataType: datatype,
		Package:  pkg,
	}
}

func MakeArgument(v *Variable, index int) *Argument {
	return &Argument{
		Function: v,
		Index:    index,
	}
}

func MakeReturn(v *Variable, index int) *ReturnType {
	return &ReturnType{
		Function: v,
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

func MakeConstantListValue(c *Constant) *ListValue {
	return &ListValue{
		Interface: c,
	}
}

func MakeMapKey(i Interface) *MapKey {
	return &MapKey{
		Interface: i,
	}
}

func MakeConstantMapKey(c *Constant) *MapKey {
	return &MapKey{
		Interface: c,
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

func MakeConstantMapValue(c *Constant) *MapValue {
	return &MapValue{
		Interface: c,
	}
}
func MakeField(v *Variable, field string, index int) *Field {
	return &Field{
		X:     v,
		Name:  field,
		Index: index,
	}
}
