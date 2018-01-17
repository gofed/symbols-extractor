package contracts

import (
	"encoding/json"
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
var UnaryOpType Type = "unaryop"
var PropagatesToType Type = "propagatesto"
var TypecastsToType Type = "typecaststo"
var IsCompatibleWithType Type = "iscompatiblewith"
var IsInvocableType Type = "isinvocable"
var HasFieldType Type = "hasfield"
var IsReferenceableType Type = "isreferenceable"
var ReferenceOfType Type = "referenceof"
var IsDereferenceableType Type = "Isdereferenceable"
var DereferenceOfType Type = "dereferenceOf"
var IsIndexableType Type = "isindexable"
var IsSendableToType Type = "issendableto"
var IsReceiveableFromType Type = "isreceiveablefrom"
var IsIncDecableType Type = "isincdecable"
var IsRangeableType Type = "israngeable"

func Contract2String(c Contract) string {
	switch d := c.(type) {
	case *BinaryOp:
		return fmt.Sprintf("BinaryOpContract:\n\tX=%v,\n\tY=%v,\n\tZ=%v,\n\top=%v\n\tPos=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), typevars.TypeVar2String(d.Z), d.OpToken, d.Pos)
	case *UnaryOp:
		return fmt.Sprintf("UnaryOpContract:\n\tX=%v,\n\tY=%v,\n\top=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.OpToken)
	case *PropagatesTo:
		return fmt.Sprintf("PropagatesTo:\n\tX=%v,\n\tY=%v,\n\tE=%#v,\n\tToVariable=%v,\n\tPos=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.ExpectedType, d.ToVariable, d.Pos)
	case *TypecastsTo:
		return fmt.Sprintf("TypecastsTo:\n\tX=%v,\n\tY=%v,\n\tType=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), typevars.TypeVar2String(d.Type))
	case *IsCompatibleWith:
		return fmt.Sprintf("IsCompatibleWith:\n\tX=%v\n\tY=%v\n\tWeak=%v\n\tE=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y), d.Weak, d.ExpectedType)
	case *IsInvocable:
		return fmt.Sprintf("IsInvocable:\n\tF=%v,\n\targCount=%v", typevars.TypeVar2String(d.F), d.ArgsCount)
	case *IsReferenceable:
		return fmt.Sprintf("IsReferenceable:\n\tX=%v", typevars.TypeVar2String(d.X))
	case *IsDereferenceable:
		return fmt.Sprintf("IsDereferenceable:\n\tX=%v", typevars.TypeVar2String(d.X))
	case *ReferenceOf:
		return fmt.Sprintf("ReferenceOf:\n\tX=%v,\n\tY=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y))
	case *DereferenceOf:
		return fmt.Sprintf("DereferenceOf:\n\tX=%v,\n\tY=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y))
	case *HasField:
		return fmt.Sprintf("HasField:\n\tX=%v,\n\tField=%v,\n\tIndex=%v,\n\tPos=%v", typevars.TypeVar2String(d.X), d.Field, d.Index, d.Pos)
	case *IsIndexable:
		return fmt.Sprintf("IsIndexable:\n\tX=%v\n\tKey=%#v\n\tIsSlice=%v,\n\tPos=%v", typevars.TypeVar2String(d.X), d.Key, d.IsSlice, d.Pos)
	case *IsSendableTo:
		return fmt.Sprintf("IsSendableTo:\n\tX=%v\n\tY=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y))
	case *IsReceiveableFrom:
		return fmt.Sprintf("IsReceiveableFrom:\n\tX=%v\n\tY=%v", typevars.TypeVar2String(d.X), typevars.TypeVar2String(d.Y))
	case *IsIncDecable:
		return fmt.Sprintf("IsIncDecable:\n\tX=%v", typevars.TypeVar2String(d.X))
	case *IsRangeable:
		return fmt.Sprintf("IsRangeable:\n\tX=%v,\n\tPos=%v", typevars.TypeVar2String(d.X), d.Pos)
	default:
		panic(fmt.Sprintf("Contract %#v not recognized", c))
	}
}

// BinaryOp represents contract between two typevars
type BinaryOp struct {
	// OpToken gives information about particular binary operation.
	// E.g. '+' can be used with integers and strings, '-' can not be used with strings.
	// As long as the operands are compatible with the operation, the contract holds.
	OpToken token.Token `json:"optoken"`
	// Z = X op Y
	X            typevars.Interface `json:"x"`
	Y            typevars.Interface `json:"y"`
	Z            typevars.Interface `json:"z"`
	ExpectedType gotypes.DataType   `json:"expectedtype"`
	Pos          string             `json:"pos"`
}

func (o *BinaryOp) MarshalJSON() (b []byte, e error) {
	type Copy BinaryOp
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *BinaryOp) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["z"])
		if err != nil {
			return err
		}
		o.Z = ti
	}

	{
		dt, err := typevars.UnmarshalDataType(objMap["expectedtype"])
		if err != nil {
			return err
		}
		o.ExpectedType = dt
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["optoken"], &(o.OpToken)); err != nil {
		return err
	}

	return nil
}

type UnaryOp struct {
	OpToken token.Token `json:"optoken"`
	// Y = op X
	X            typevars.Interface `json:"x"`
	Y            typevars.Interface `json:"y"`
	ExpectedType gotypes.DataType   `json:"expectedtype"`
	Pos          string             `json:"pos"`
}

func (o *UnaryOp) MarshalJSON() (b []byte, e error) {
	type Copy UnaryOp
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *UnaryOp) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	{
		dt, err := typevars.UnmarshalDataType(objMap["expectedtype"])
		if err != nil {
			return err
		}
		o.ExpectedType = dt
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["optoken"], &(o.OpToken)); err != nil {
		return err
	}

	return nil
}

type PropagatesTo struct {
	X            typevars.Interface `json:"x"`
	Y            typevars.Interface `json:"y"`
	ExpectedType gotypes.DataType   `json:"expectedtype"`
	ToVariable   bool               `json:"tovariable"`
	Pos          string             `json:"pos"`
}

func (o *PropagatesTo) MarshalJSON() (b []byte, e error) {
	type Copy PropagatesTo
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func UnmarshalTypevar(rawMessage *json.RawMessage) (typevars.Interface, error) {
	if rawMessage == nil {
		return nil, nil
	}
	var a map[string]interface{}
	if err := json.Unmarshal(*rawMessage, &a); err != nil {
		return nil, err
	}

	switch typevars.Type(a["type"].(string)) {
	case typevars.ConstantType:
		r := &typevars.Constant{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.VariableType:
		r := &typevars.Variable{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.ArgumentType:
		r := &typevars.Argument{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.ReturnTypeType:
		r := &typevars.ReturnType{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.MapKeyType:
		r := &typevars.MapKey{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.MapValueType:
		r := &typevars.MapValue{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.ListKeyType:
		r := &typevars.ListKey{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.ListValueType:
		r := &typevars.ListValue{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.FieldType:
		r := &typevars.Field{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.CGOType:
		r := &typevars.CGO{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.RangeKeyType:
		r := &typevars.RangeKey{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	case typevars.RangeValueType:
		r := &typevars.RangeValue{}
		if err := json.Unmarshal(*rawMessage, &r); err != nil {
			return nil, err
		}
		return r, nil
	default:
		panic(fmt.Errorf("Unrecognized typevar %v", a["type"].(string)))
	}
}

func (o *PropagatesTo) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	{
		dt, err := typevars.UnmarshalDataType(objMap["expectedtype"])
		if err != nil {
			return err
		}
		o.ExpectedType = dt
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type TypecastsTo struct {
	X            typevars.Interface `json:"x"`
	Type         typevars.Interface `json:"castedtype"`
	Y            typevars.Interface `json:"y"`
	ExpectedType gotypes.DataType   `json:"expectedtype"`
	Pos          string             `json:"pos"`
}

func (o *TypecastsTo) MarshalJSON() (b []byte, e error) {
	type Copy TypecastsTo
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *TypecastsTo) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return fmt.Errorf("TypecastsTo.UnmarshalJSON: %v", err)
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return fmt.Errorf("TypecastsTo.UnmarshalTypevar(x): %v", err)
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return fmt.Errorf("TypecastsTo.UnmarshalTypevar(y): %v", err)
		}
		o.Y = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["castedtype"])
		if err != nil {
			return fmt.Errorf("TypecastsTo.UnmarshalTypevar(type): %v", err)
		}
		o.Type = ti
	}

	{
		dt, err := typevars.UnmarshalDataType(objMap["expectedtype"])
		if err != nil {
			return fmt.Errorf("TypecastsTo.UnmarshalDataType(expectedtype): %v", err)
		}
		o.ExpectedType = dt
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return fmt.Errorf("TypecastsTo.Unmarshal(pos): %v", err)
	}

	return nil
}

type IsCompatibleWith struct {
	X            typevars.Interface `json:"x"`
	Y            typevars.Interface `json:"y"`
	ExpectedType gotypes.DataType   `json:"expectedtype"`
	// As long as MapKey is compatible with integer, it is compatible with ListKey as well
	// TODO(jchaloup): make sure this principle is applied during the compatibility analysis
	Weak bool   `json:"weak"`
	Pos  string `json:"pos"`
}

func (o *IsCompatibleWith) MarshalJSON() (b []byte, e error) {
	type Copy IsCompatibleWith
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsCompatibleWith) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	{
		dt, err := typevars.UnmarshalDataType(objMap["expectedtype"])
		if err != nil {
			return err
		}
		o.ExpectedType = dt
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["weak"], &(o.Weak)); err != nil {
		return err
	}

	return nil
}

type IsInvocable struct {
	F         typevars.Interface `json:"f"`
	ArgsCount int                `json:"argscount"`
	Pos       string             `json:"pos"`
}

func (o *IsInvocable) MarshalJSON() (b []byte, e error) {
	type Copy IsInvocable
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsInvocable) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["f"])
		if err != nil {
			return err
		}
		o.F = ti
	}

	if err := json.Unmarshal(*objMap["argscount"], &(o.ArgsCount)); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type HasField struct {
	X     typevars.Interface `json:"x"`
	Field string             `json:"field"`
	Index int                `json:"index"`
	Pos   string             `json:"pos"`
}

func (o *HasField) MarshalJSON() (b []byte, e error) {
	type Copy HasField
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *HasField) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["field"], &(o.Field)); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["index"], &(o.Index)); err != nil {
		return err
	}

	return nil
}

type IsReferenceable struct {
	X   typevars.Interface `json:"x"`
	Pos string             `json:"pos"`
}

func (o *IsReferenceable) MarshalJSON() (b []byte, e error) {
	type Copy IsReferenceable
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsReferenceable) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type ReferenceOf struct {
	X   typevars.Interface `json:"x"`
	Y   typevars.Interface `json:"y"`
	Pos string             `json:"pos"`
}

func (o *ReferenceOf) MarshalJSON() (b []byte, e error) {
	type Copy ReferenceOf
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *ReferenceOf) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type IsDereferenceable struct {
	X   typevars.Interface `json:"x"`
	Pos string             `json:"pos"`
}

func (o *IsDereferenceable) MarshalJSON() (b []byte, e error) {
	type Copy IsDereferenceable
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsDereferenceable) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type DereferenceOf struct {
	X   typevars.Interface `json:"x"`
	Y   typevars.Interface `json:"y"`
	Pos string             `json:"pos"`
}

func (o *DereferenceOf) MarshalJSON() (b []byte, e error) {
	type Copy DereferenceOf
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *DereferenceOf) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type IsIndexable struct {
	X       typevars.Interface `json:"x"`
	Key     typevars.Interface `json:"key"`
	IsSlice bool               `json:"isslice"`
	Pos     string             `json:"pos"`
}

func (o *IsIndexable) MarshalJSON() (b []byte, e error) {
	type Copy IsIndexable
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsIndexable) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["key"])
		if err != nil {
			return err
		}
		o.Key = ti
	}

	if err := json.Unmarshal(*objMap["isslice"], &(o.IsSlice)); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type IsSendableTo struct {
	X   typevars.Interface `json:"x"`
	Y   typevars.Interface `json:"y"`
	Pos string             `json:"pos"`
}

func (o *IsSendableTo) MarshalJSON() (b []byte, e error) {
	type Copy IsSendableTo
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsSendableTo) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type IsReceiveableFrom struct {
	X            typevars.Interface `json:"x"`
	Y            typevars.Interface `json:"y"`
	ExpectedType gotypes.DataType   `json:"expectedtype"`
	Pos          string             `json:"pos"`
}

func (o *IsReceiveableFrom) MarshalJSON() (b []byte, e error) {
	type Copy IsReceiveableFrom
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsReceiveableFrom) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	{
		ti, err := UnmarshalTypevar(objMap["y"])
		if err != nil {
			return err
		}
		o.Y = ti
	}

	{
		dt, err := typevars.UnmarshalDataType(objMap["expectedtype"])
		if err != nil {
			return err
		}
		o.ExpectedType = dt
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type IsIncDecable struct {
	X   typevars.Interface `json:"x"`
	Pos string             `json:"pos"`
}

func (o *IsIncDecable) MarshalJSON() (b []byte, e error) {
	type Copy IsIncDecable
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsIncDecable) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

type IsRangeable struct {
	X   typevars.Interface `json:"x"`
	Pos string             `json:"pos"`
}

func (o *IsRangeable) MarshalJSON() (b []byte, e error) {
	type Copy IsRangeable
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: string(o.GetType()),
		Copy: (*Copy)(o),
	})
}

func (o *IsRangeable) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	{
		ti, err := UnmarshalTypevar(objMap["x"])
		if err != nil {
			return err
		}
		o.X = ti
	}

	if err := json.Unmarshal(*objMap["pos"], &(o.Pos)); err != nil {
		return err
	}

	return nil
}

func (b *BinaryOp) GetType() Type {
	return BinaryOpType
}

func (b *UnaryOp) GetType() Type {
	return UnaryOpType
}

func (p *PropagatesTo) GetType() Type {
	return PropagatesToType
}

func (p *TypecastsTo) GetType() Type {
	return TypecastsToType
}

func (i *IsCompatibleWith) GetType() Type {
	return IsCompatibleWithType
}

func (i *IsInvocable) GetType() Type {
	return IsInvocableType
}

func (i *HasField) GetType() Type {
	return HasFieldType
}

func (i *IsReferenceable) GetType() Type {
	return IsReferenceableType
}

func (i *IsDereferenceable) GetType() Type {
	return IsDereferenceableType
}

func (i *ReferenceOf) GetType() Type {
	return ReferenceOfType
}

func (i *DereferenceOf) GetType() Type {
	return DereferenceOfType
}

func (i *IsIndexable) GetType() Type {
	return IsIndexableType
}

func (i *IsSendableTo) GetType() Type {
	return IsSendableToType
}

func (i *IsReceiveableFrom) GetType() Type {
	return IsReceiveableFromType
}

func (i *IsIncDecable) GetType() Type {
	return IsIncDecableType
}

func (i *IsRangeable) GetType() Type {
	return IsRangeableType
}
