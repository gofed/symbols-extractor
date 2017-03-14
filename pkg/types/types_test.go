package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

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
		name     string
		dataType string
		value    DataType
		empty    DataType
	}{
		// Simple identifier
		{
			name:  "Identifier",
			value: id,
			empty: &Identifier{},
		},
		// Simple channel
		{
			name:  "Channel",
			value: channel,
			empty: &Channel{},
		},
		// Slice with structured type
		{
			name:  "Slice",
			value: slice,
			empty: &Slice{},
		},
		// Function
		{
			name: "Function",
			value: &Function{
				Params:  []DataType{id, slice, channel},
				Results: []DataType{id, slice, channel},
			},
			empty: &Function{},
		},
	}

	for _, test := range tests {
		newObject := test.empty
		byteSlice, _ := json.Marshal(test.value)

		t.Logf("\nBefore: %v", string(byteSlice))

		if err := json.Unmarshal(byteSlice, newObject); err != nil {
			t.Errorf("Error: %v", err)
		}

		// compare
		if !reflect.DeepEqual(test.value, newObject) {
			t.Errorf("%#v != %#v", test.value, newObject)
		}

		t.Logf("\nAfter: %#v", newObject)
	}

}
