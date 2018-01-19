package typevars

import (
	"encoding/json"
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
	gotypes.DataType `json:"datatype"`
	Package          string `json:"package"`
}

func (c *Constant) GetType() Type {
	return ConstantType
}

func (o *Constant) MarshalJSON() (b []byte, e error) {
	type Copy Constant
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(ConstantType),
		Copy: (*Copy)(o),
	})
}

func (o *Constant) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["package"], &(o.Package)); err != nil {
		return err
	}

	dt, err := UnmarshalDataType(objMap["datatype"])
	if err != nil {
		return err
	}

	o.DataType = dt

	return nil
}

func UnmarshalDataType(rawMessage *json.RawMessage) (gotypes.DataType, error) {
	// TODO(jchaloup): this is wrong, the data type must be always set
	if rawMessage == nil {
		return nil, nil
	}

	var a map[string]interface{}
	if err := json.Unmarshal(*rawMessage, &a); err != nil {
		return nil, err
	}

	// TODO(jchaloup): this is wrong, the data type must be always set
	if _, ok := a["type"]; !ok {
		return nil, nil
	}

	switch a["type"].(string) {
	case gotypes.IdentifierType:
		r := &gotypes.Identifier{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.BuiltinType:
		r := &gotypes.Builtin{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.ConstantType:
		r := &gotypes.Constant{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.PackagequalifierType:
		r := &gotypes.Packagequalifier{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.SelectorType:
		r := &gotypes.Selector{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.ChannelType:
		r := &gotypes.Channel{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.SliceType:
		r := &gotypes.Slice{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.ArrayType:
		r := &gotypes.Array{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.MapType:
		r := &gotypes.Map{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.PointerType:
		r := &gotypes.Pointer{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.EllipsisType:
		r := &gotypes.Ellipsis{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.FunctionType:
		r := &gotypes.Function{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.MethodType:
		r := &gotypes.Method{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.InterfaceType:
		r := &gotypes.Interface{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.StructType:
		r := &gotypes.Struct{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case gotypes.NilType:
		r := &gotypes.Nil{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	default:
		panic(fmt.Errorf("Unrecognized data type %v", a["type"].(string)))
	}
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

func (o *Variable) MarshalJSON() (b []byte, e error) {
	type Copy Variable
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(VariableType),
		Copy: (*Copy)(o),
	})
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
	Pos   string
}

func (f *Field) GetType() Type {
	return FieldType
}

func (o *Field) MarshalJSON() (b []byte, e error) {
	type Copy Field
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(FieldType),
		Copy: (*Copy)(o),
	})
}

type CGO struct {
	gotypes.DataType `json:"datatype"`
	Package          string `json:"package"`
}

func (c *CGO) GetType() Type {
	return CGOType
}

func (o *CGO) MarshalJSON() (b []byte, e error) {

	type Copy CGO
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(CGOType),
		Copy: (*Copy)(o),
	})
}

func (o *CGO) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["package"], &(o.Package)); err != nil {
		return err
	}

	dt, err := UnmarshalDataType(objMap["datatype"])
	if err != nil {
		return err
	}

	o.DataType = dt

	return nil
}

func TypeVar2String(tv Interface) string {
	switch d := tv.(type) {
	case *Constant:
		return fmt.Sprintf("TypeVar.Constant: %#v, Package: %v", d.DataType, d.Package)
	case *Variable:
		return fmt.Sprintf("TypeVar.Variable: (%v) %v at %v", d.Package, d.Name, d.Pos)
	case *ListKey:
		return fmt.Sprintf("TypeVar.ListKey: integer type")
	case *ListValue:
		return fmt.Sprintf("TypeVar.ListValue: %#v", d.X)
	case *MapKey:
		return fmt.Sprintf("TypeVar.MapKey: %#v", d.X)
	case *MapValue:
		return fmt.Sprintf("TypeVar.MapValue: %#v", d.X)
	case *RangeKey:
		return fmt.Sprintf("TypeVar.RangeKey: %#v", d.X)
	case *RangeValue:
		return fmt.Sprintf("TypeVar.RangeValue: %#v", d.X)
	case *ReturnType:
		return fmt.Sprintf("TypeVar.ReturnType: (%v) at %v", TypeVar2String(d.Function), d.Index)
	case *Argument:
		return fmt.Sprintf("TypeVar.Argument: (%v) at %v", TypeVar2String(d.Function), d.Index)
	case *Field:
		if d.Name == "" {
			return fmt.Sprintf("TypeVar.Field: %#v at index %v at pos %v", d.X, d.Index, d.Pos)
		}
		return fmt.Sprintf("TypeVar.Field: %#v with field %q at pos %v", d.X, d.Name, d.Pos)
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

func (o *ListKey) MarshalJSON() (b []byte, e error) {

	type Copy ListKey
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(ListKeyType),
		Copy: (*Copy)(o),
	})
}

type ListValue struct {
	X *Variable
}

func (l *ListValue) GetType() Type {
	return ListValueType
}

func (o *ListValue) MarshalJSON() (b []byte, e error) {

	type Copy ListValue
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(ListValueType),
		Copy: (*Copy)(o),
	})
}

type MapKey struct {
	X *Variable
}

func (m *MapKey) GetType() Type {
	return MapKeyType
}

func (o *MapKey) MarshalJSON() (b []byte, e error) {

	type Copy MapKey
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(MapKeyType),
		Copy: (*Copy)(o),
	})
}

type MapValue struct {
	X *Variable
}

func (m *MapValue) GetType() Type {
	return MapValueType
}

func (o *MapValue) MarshalJSON() (b []byte, e error) {

	type Copy MapValue
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(MapValueType),
		Copy: (*Copy)(o),
	})
}

type RangeKey struct {
	X *Variable
}

func (m *RangeKey) GetType() Type {
	return RangeKeyType
}

func (o *RangeKey) MarshalJSON() (b []byte, e error) {

	type Copy RangeKey
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

type RangeValue struct {
	X *Variable
}

func (m *RangeValue) GetType() Type {
	return RangeValueType
}

func (o *RangeValue) MarshalJSON() (b []byte, e error) {

	type Copy RangeValue
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
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

func (o *Argument) MarshalJSON() (b []byte, e error) {

	type Copy Argument
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
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

func (o *ReturnType) MarshalJSON() (b []byte, e error) {

	type Copy ReturnType
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
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

func MakeListValue(v *Variable) *ListValue {
	return &ListValue{
		X: v,
	}
}

func MakeMapKey(v *Variable) *MapKey {
	return &MapKey{
		X: v,
	}
}

func MakeMapValue(v *Variable) *MapValue {
	return &MapValue{
		X: v,
	}
}

func MakeRangeKey(v *Variable) *RangeKey {
	return &RangeKey{
		X: v,
	}
}

func MakeRangeValue(v *Variable) *RangeValue {
	return &RangeValue{
		X: v,
	}
}

func MakeField(v *Variable, field string, index int, pos string) *Field {
	return &Field{
		X:     v,
		Name:  field,
		Index: index,
		Pos:   pos,
	}
}
