package symbols

import (
	"encoding/json"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type SymbolDef struct {
	Pos     string           `json:"pos"`
	Name    string           `json:"name"`
	Package string           `json:"package"`
	Def     gotypes.DataType `json:"def"`
	Block   int              `json:"block"`
}

func (o *SymbolDef) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["pos"], &o.Pos); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	if err := json.Unmarshal(*objMap["package"], &o.Package); err != nil {
		return err
	}

	if _, ok := objMap["block"]; ok {
		if err := json.Unmarshal(*objMap["block"], &o.Block); err != nil {
			return err
		}
	}

	var m map[string]interface{}
	if err := json.Unmarshal(*objMap["def"], &m); err != nil {
		return err
	}

	switch dataType := m["type"]; dataType {

	case gotypes.IdentifierType:
		r := &gotypes.Identifier{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.BuiltinType:
		r := &gotypes.Builtin{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.ConstantType:
		r := &gotypes.Constant{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.PackagequalifierType:
		r := &gotypes.Packagequalifier{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.SelectorType:
		r := &gotypes.Selector{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.ChannelType:
		r := &gotypes.Channel{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.SliceType:
		r := &gotypes.Slice{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.ArrayType:
		r := &gotypes.Array{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.MapType:
		r := &gotypes.Map{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.PointerType:
		r := &gotypes.Pointer{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.EllipsisType:
		r := &gotypes.Ellipsis{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.FunctionType:
		r := &gotypes.Function{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.MethodType:
		r := &gotypes.Method{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.InterfaceType:
		r := &gotypes.Interface{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case gotypes.StructType:
		r := &gotypes.Struct{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	}

	return nil
}
