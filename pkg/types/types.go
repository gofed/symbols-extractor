package types

import "encoding/json"

// DataType is
type DataType interface {
	GetType() string
}

const FunctionType = "function"

type Function struct {
	Params  []DataType `json:"params"`
	Results []DataType `json:"results"`
}

func (o *Function) GetType() string {
	return FunctionType
}

func (o *Function) MarshalJSON() (b []byte, e error) {
	type Copy Function
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: FunctionType,
		Copy: (*Copy)(o),
	})
}

func (o *Function) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Params field
	{
		if objMap["params"] != nil {
			var l []*json.RawMessage
			if err := json.Unmarshal(*objMap["params"], &l); err != nil {
				return err
			}

			o.Params = make([]DataType, 0)
			for _, item := range l {
				var m map[string]interface{}
				if err := json.Unmarshal(*item, &m); err != nil {
					return err
				}
				switch dataType := m["type"]; dataType {

				case IdentifierType:
					r := &Identifier{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case BuiltinType:
					r := &Builtin{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case SelectorType:
					r := &Selector{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case ChannelType:
					r := &Channel{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case SliceType:
					r := &Slice{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case ArrayType:
					r := &Array{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case MapType:
					r := &Map{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case PointerType:
					r := &Pointer{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case EllipsisType:
					r := &Ellipsis{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case FunctionType:
					r := &Function{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case MethodType:
					r := &Method{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case InterfaceType:
					r := &Interface{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				case StructType:
					r := &Struct{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Params = append(o.Params, r)

				}
			}
		}
	}

	// block for Results field
	{
		if objMap["results"] != nil {
			var l []*json.RawMessage
			if err := json.Unmarshal(*objMap["results"], &l); err != nil {
				return err
			}

			o.Results = make([]DataType, 0)
			for _, item := range l {
				var m map[string]interface{}
				if err := json.Unmarshal(*item, &m); err != nil {
					return err
				}
				switch dataType := m["type"]; dataType {

				case IdentifierType:
					r := &Identifier{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case BuiltinType:
					r := &Builtin{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case SelectorType:
					r := &Selector{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case ChannelType:
					r := &Channel{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case SliceType:
					r := &Slice{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case ArrayType:
					r := &Array{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case MapType:
					r := &Map{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case PointerType:
					r := &Pointer{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case FunctionType:
					r := &Function{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case MethodType:
					r := &Method{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case InterfaceType:
					r := &Interface{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				case StructType:
					r := &Struct{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Results = append(o.Results, r)

				}
			}
		}
	}

	return nil
}

const MapType = "map"

type Map struct {
	Keytype   DataType `json:"keytype"`
	Valuetype DataType `json:"valuetype"`
}

func (o *Map) GetType() string {
	return MapType
}

func (o *Map) MarshalJSON() (b []byte, e error) {
	type Copy Map
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: MapType,
		Copy: (*Copy)(o),
	})
}

func (o *Map) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Keytype field
	{
		if objMap["keytype"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["keytype"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["keytype"], &r); err != nil {
					return err
				}
				o.Keytype = r

			}
		}
	}

	// block for Valuetype field
	{
		if objMap["valuetype"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["valuetype"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["valuetype"], &r); err != nil {
					return err
				}
				o.Valuetype = r

			}
		}
	}

	return nil
}

const SliceType = "slice"

type Slice struct {
	Elmtype DataType `json:"elmtype"`
}

func (o *Slice) GetType() string {
	return SliceType
}

func (o *Slice) MarshalJSON() (b []byte, e error) {
	type Copy Slice
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: SliceType,
		Copy: (*Copy)(o),
	})
}

func (o *Slice) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Elmtype field
	{
		if objMap["elmtype"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["elmtype"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			}
		}
	}

	return nil
}

const StructFieldsItemType = "structfieldsitem"

type StructFieldsItem struct {
	Name string `json:"name"`

	Def DataType `json:"def"`
}

func (o *StructFieldsItem) GetType() string {
	return StructFieldsItemType
}

func (o *StructFieldsItem) MarshalJSON() (b []byte, e error) {
	type Copy StructFieldsItem
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: StructFieldsItemType,
		Copy: (*Copy)(o),
	})
}

func (o *StructFieldsItem) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["name"] actually exists
	if err := json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	// block for Def field
	{
		if objMap["def"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["def"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			}
		}
	}

	return nil
}

const StructType = "struct"

type Struct struct {
	Fields []StructFieldsItem `json:"fields"`
}

func (o *Struct) GetType() string {
	return StructType
}

func (o *Struct) MarshalJSON() (b []byte, e error) {
	type Copy Struct
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: StructType,
		Copy: (*Copy)(o),
	})
}

func (o *Struct) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Fields field
	{
		if objMap["fields"] != nil {
			var l []*json.RawMessage
			if err := json.Unmarshal(*objMap["fields"], &l); err != nil {
				return err
			}

			o.Fields = make([]StructFieldsItem, 0)
			for _, item := range l {
				var m map[string]interface{}
				if err := json.Unmarshal(*item, &m); err != nil {
					return err
				}
				switch dataType := m["type"]; dataType {

				case StructFieldsItemType:
					r := &StructFieldsItem{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Fields = append(o.Fields, *r)

				}
			}
		}
	}

	return nil
}

const PointerType = "pointer"

type Pointer struct {
	Def DataType `json:"def"`
}

func (o *Pointer) GetType() string {
	return PointerType
}

func (o *Pointer) MarshalJSON() (b []byte, e error) {
	type Copy Pointer
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: PointerType,
		Copy: (*Copy)(o),
	})
}

func (o *Pointer) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Def field
	{
		if objMap["def"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["def"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			}
		}
	}

	return nil
}

const SelectorType = "selector"

type Selector struct {
	Item string `json:"item"`

	Prefix DataType `json:"prefix"`
}

func (o *Selector) GetType() string {
	return SelectorType
}

func (o *Selector) MarshalJSON() (b []byte, e error) {
	type Copy Selector
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: SelectorType,
		Copy: (*Copy)(o),
	})
}

func (o *Selector) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["item"] actually exists
	if err := json.Unmarshal(*objMap["item"], &o.Item); err != nil {
		return err
	}

	// block for Prefix field
	{
		if objMap["prefix"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["prefix"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["prefix"], &r); err != nil {
					return err
				}
				o.Prefix = r

			case PackagequalifierType:
				r := &Packagequalifier{}
				if err := json.Unmarshal(*objMap["prefix"], &r); err != nil {
					return err
				}
				o.Prefix = r

			}
		}
	}

	return nil
}

const BuiltinType = "builtin"

type Builtin struct {
	Untyped bool   `json:"untyped"`
	Def     string `json:"def"`
}

func (o *Builtin) GetType() string {
	return BuiltinType
}

func (o *Builtin) MarshalJSON() (b []byte, e error) {
	type Copy Builtin
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: BuiltinType,
		Copy: (*Copy)(o),
	})
}

func (o *Builtin) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["untyped"] actually exists
	if err := json.Unmarshal(*objMap["untyped"], &o.Untyped); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["def"] actually exists
	if err := json.Unmarshal(*objMap["def"], &o.Def); err != nil {
		return err
	}

	return nil
}

const InterfaceMethodsItemType = "interfacemethodsitem"

type InterfaceMethodsItem struct {
	Name string `json:"name"`

	Def DataType `json:"def"`
}

func (o *InterfaceMethodsItem) GetType() string {
	return InterfaceMethodsItemType
}

func (o *InterfaceMethodsItem) MarshalJSON() (b []byte, e error) {
	type Copy InterfaceMethodsItem
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: InterfaceMethodsItemType,
		Copy: (*Copy)(o),
	})
}

func (o *InterfaceMethodsItem) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["name"] actually exists
	if err := json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	// block for Def field
	{
		if objMap["def"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["def"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			}
		}
	}

	return nil
}

const InterfaceType = "interface"

type Interface struct {
	Methods []InterfaceMethodsItem `json:"methods"`
}

func (o *Interface) GetType() string {
	return InterfaceType
}

func (o *Interface) MarshalJSON() (b []byte, e error) {
	type Copy Interface
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: InterfaceType,
		Copy: (*Copy)(o),
	})
}

func (o *Interface) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Methods field
	{
		if objMap["methods"] != nil {
			var l []*json.RawMessage
			if err := json.Unmarshal(*objMap["methods"], &l); err != nil {
				return err
			}

			o.Methods = make([]InterfaceMethodsItem, 0)
			for _, item := range l {
				var m map[string]interface{}
				if err := json.Unmarshal(*item, &m); err != nil {
					return err
				}
				switch dataType := m["type"]; dataType {

				case InterfaceMethodsItemType:
					r := &InterfaceMethodsItem{}
					if err := json.Unmarshal(*item, &r); err != nil {
						return err
					}
					o.Methods = append(o.Methods, *r)

				}
			}
		}
	}

	return nil
}

const EllipsisType = "ellipsis"

type Ellipsis struct {
	Def DataType `json:"def"`
}

func (o *Ellipsis) GetType() string {
	return EllipsisType
}

func (o *Ellipsis) MarshalJSON() (b []byte, e error) {
	type Copy Ellipsis
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: EllipsisType,
		Copy: (*Copy)(o),
	})
}

func (o *Ellipsis) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Def field
	{
		if objMap["def"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["def"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			}
		}
	}

	return nil
}

const ArrayType = "array"

type Array struct {
	Elmtype DataType `json:"elmtype"`
}

func (o *Array) GetType() string {
	return ArrayType
}

func (o *Array) MarshalJSON() (b []byte, e error) {
	type Copy Array
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: ArrayType,
		Copy: (*Copy)(o),
	})
}

func (o *Array) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Elmtype field
	{
		if objMap["elmtype"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["elmtype"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["elmtype"], &r); err != nil {
					return err
				}
				o.Elmtype = r

			}
		}
	}

	return nil
}

const IdentifierType = "identifier"

type Identifier struct {
	Def     string `json:"def"`
	Package string `json:"package"`
}

func (o *Identifier) GetType() string {
	return IdentifierType
}

func (o *Identifier) MarshalJSON() (b []byte, e error) {
	type Copy Identifier
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: IdentifierType,
		Copy: (*Copy)(o),
	})
}

func (o *Identifier) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["def"] actually exists
	if err := json.Unmarshal(*objMap["def"], &o.Def); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["package"] actually exists
	if err := json.Unmarshal(*objMap["package"], &o.Package); err != nil {
		return err
	}

	return nil
}

const PackagequalifierType = "packagequalifier"

type Packagequalifier struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func (o *Packagequalifier) GetType() string {
	return PackagequalifierType
}

func (o *Packagequalifier) MarshalJSON() (b []byte, e error) {
	type Copy Packagequalifier
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: PackagequalifierType,
		Copy: (*Copy)(o),
	})
}

func (o *Packagequalifier) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["path"] actually exists
	if err := json.Unmarshal(*objMap["path"], &o.Path); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["name"] actually exists
	if err := json.Unmarshal(*objMap["name"], &o.Name); err != nil {
		return err
	}

	return nil
}

const MethodType = "method"

type Method struct {
	Def      DataType `json:"def"`
	Receiver DataType `json:"receiver"`
}

func (o *Method) GetType() string {
	return MethodType
}

func (o *Method) MarshalJSON() (b []byte, e error) {
	type Copy Method
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: MethodType,
		Copy: (*Copy)(o),
	})
}

func (o *Method) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// block for Def field
	{
		if objMap["def"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["def"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["def"], &r); err != nil {
					return err
				}
				o.Def = r

			}
		}
	}

	// block for Receiver field
	{
		if objMap["receiver"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["receiver"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["receiver"], &r); err != nil {
					return err
				}
				o.Receiver = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["receiver"], &r); err != nil {
					return err
				}
				o.Receiver = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["receiver"], &r); err != nil {
					return err
				}
				o.Receiver = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["receiver"], &r); err != nil {
					return err
				}
				o.Receiver = r

			}
		}
	}

	return nil
}

const ChannelType = "channel"

type Channel struct {
	Dir string `json:"dir"`

	Value DataType `json:"value"`
}

func (o *Channel) GetType() string {
	return ChannelType
}

func (o *Channel) MarshalJSON() (b []byte, e error) {
	type Copy Channel
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: ChannelType,
		Copy: (*Copy)(o),
	})
}

func (o *Channel) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// TODO(jchaloup): check the objMap["dir"] actually exists
	if err := json.Unmarshal(*objMap["dir"], &o.Dir); err != nil {
		return err
	}

	// block for Value field
	{
		if objMap["value"] != nil {
			var m map[string]interface{}
			if err := json.Unmarshal(*objMap["value"], &m); err != nil {
				return err
			}

			switch dataType := m["type"]; dataType {

			case IdentifierType:
				r := &Identifier{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case BuiltinType:
				r := &Builtin{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case SelectorType:
				r := &Selector{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case ChannelType:
				r := &Channel{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case SliceType:
				r := &Slice{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case ArrayType:
				r := &Array{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case MapType:
				r := &Map{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case PointerType:
				r := &Pointer{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case FunctionType:
				r := &Function{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case MethodType:
				r := &Method{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case InterfaceType:
				r := &Interface{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			case StructType:
				r := &Struct{}
				if err := json.Unmarshal(*objMap["value"], &r); err != nil {
					return err
				}
				o.Value = r

			}
		}
	}

	return nil
}

type SymbolDef struct {
	Pos     string   `json:"pos"`
	Name    string   `json:"name"`
	Package string   `json:"package"`
	Def     DataType `json:"def"`
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

	var m map[string]interface{}
	if err := json.Unmarshal(*objMap["def"], &m); err != nil {
		return err
	}

	switch dataType := m["type"]; dataType {

	case IdentifierType:
		r := &Identifier{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case BuiltinType:
		r := &Builtin{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case PackagequalifierType:
		r := &Packagequalifier{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case SelectorType:
		r := &Selector{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case ChannelType:
		r := &Channel{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case SliceType:
		r := &Slice{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case ArrayType:
		r := &Array{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case MapType:
		r := &Map{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case PointerType:
		r := &Pointer{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case EllipsisType:
		r := &Ellipsis{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case FunctionType:
		r := &Function{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case MethodType:
		r := &Method{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case InterfaceType:
		r := &Interface{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	case StructType:
		r := &Struct{}
		if err := json.Unmarshal(*objMap["def"], &r); err != nil {
			return err
		}
		o.Def = r

	}

	return nil
}
