package testdata

import (
	"net/http"
	"testing"
)

type Interface interface {
	JustAnotherFunction(a, b int, list map[string]string) (str string, err error)
}

type Struct struct {
	// Simple ID
	simpleID uint64
	// Pointer
	simplePointerID *string
	// Channel
	simpleChannel chan struct{}
	// Struct
	simpleStruct struct {
		simpleID uint64
	}
	// Map
	simpleMap [string]struct{}
	// Array
	simpleSlice  []string
	simpleSlice2 []Struct
	simpleArray  [4]string
	// Function
	simpleMethod       func(arg1, arg2 string) (string, error)
	simpleStringMethod func(a, b string) Struct
	// interface
	simpleInterface interface {
		simpleMethod(arg1, arg2 string) (string, error)
	}
	// simple method with interface
	simpleMethodWithInterface func(a, b string) Interface
	// Ellipsis
	simpleMethodWithEllipsis func(arg1 string, ellipsis ...string) (string, error)

	// fully qualified ID
	qid http.MethodGet
}

func (s *Struct) JustAnotherFunction(a, b int, list map[string]string) (str string, err error) {
	return "", nil
}

func reallyAFunction() string {
	return "NotAnEmptyString"
}

func (s *Struct) JustAFunction(a, b int, list map[string]string) (string, error) {
	return list["neco"], nil
	return "", nil
	return "" + "", nil
	return JustAnotherFunction(a, b, list)
	return "" + reallyAFunction(), nil
	return "" + s.simpleArray[2]
	return s.simpleStringMethod(",", ",").simpleSlice2[2].simpleSlice[1]
	return s.simpleMethodWithInterface("", "").(*Struct).simpleMethodWithInterface("", "").(*Struct)
}

type DataType interface {
	GetType() string
}

type Identifier struct {
	Def string `json:"def"`
}

type Channel struct {
	Dir string `json:"dir"`

	Value DataType `json:"value"`
}

type Slice struct {
	Elmtype DataType `json:"elmtype"`
}

func TestMarshalUnmarshal(t *testing.T) {

	id := &Identifier{Def: "Poker"}
	channel := &Channel{
		Dir:   "bi-directional",
		Value: id,
	}
	slice := &Slice{
		Elmtype: channel,
	}

	tests := []struct {
		name, dataType string
		value          DataType
		empty          DataType
		ident          map[string]*Identifier
	}{
		// Simple identifier
		{
			name:  "Identifier",
			value: id,
			empty: &Identifier{},
			ident: map[string]*Identifier{
				"field1": &Identifier{
					Def: "Field1",
				},
				"field2": &Identifier{
					Def: "Field2",
				},
			},
		},
	}
	//
	// tests2 := [][][]*Identifier{}

	// 		// Simple channel
	// 		{
	// 			name:  "Channel",
	// 			value: channel,
	// 			empty: &Channel{},
	// 		},
	// 		// Slice with structured type
	// 		{
	// 			name:  "Slice",
	// 			value: slice,
	// 			empty: &Slice{},
	// 		},
	// 		// Function
	// 		{
	// 			name: "Function",
	// 			value: &Function{
	// 				Params:  []DataType{id, slice, channel},
	// 				Results: []DataType{id, slice, channel},
	// 			},
	// 			empty: &Function{},
	// 		},
	// 	}
	//
	// 	for _, test := range tests {
	// 		{
	// 			newObject := test.empty
	// 			byteSlice, _ := json.Marshal(test.value)
	//
	// 			t.Logf("\nBefore: %v", string(byteSlice))
	//
	// 			if err := json.Unmarshal(byteSlice, newObject); err != nil {
	// 				t.Errorf("Error: %v", err)
	// 			}
	//
	// 			// compare
	// 			if !reflect.DeepEqual(test.value, newObject) {
	// 				t.Errorf("%#v != %#v", test.value, newObject)
	// 			}
	//
	// 			t.Logf("\nAfter: %#v", newObject)
	// 		}
	//
	// 		// wrap the type into a symbol
	// 		{
	// 			def := &SymbolDef{
	// 				Pos:  "Ppzice",
	// 				Name: "Name",
	// 				Def:  test.value,
	// 			}
	//
	// 			t.Logf("%v\n", def)
	// 			byteSlice, _ := json.Marshal(def)
	//
	// 			t.Logf("\nBefore: %v", string(byteSlice))
	// 			newObject := &SymbolDef{}
	// 			if err := json.Unmarshal(byteSlice, newObject); err != nil {
	// 				t.Errorf("Error: %v", err)
	// 			}
	//
	// 			// compare
	// 			if !reflect.DeepEqual(def, newObject) {
	// 				t.Errorf("%#v != %#v", def, newObject)
	// 			}
	//
	// 			t.Logf("\nAfter: %#v", newObject)
	// 		}
	// 	}
}
