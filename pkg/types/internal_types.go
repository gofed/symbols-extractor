package types

import "encoding/json"

const NilType = "nil"

type Nil struct {
}

func (n *Nil) GetType() string {
	return NilType
}

func (o *Nil) MarshalJSON() (b []byte, e error) {
	type Copy Nil
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Copy
	}{
		Type: NilType,
		Copy: (*Copy)(o),
	})
}

func (o *Nil) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	return nil
}

// BuiltinLiteralType is a type constant for built-in literal
const BuiltinLiteralType = "builtinliteral"

// BuiltinLiteral represents literals like true or false
type BuiltinLiteral struct {
	Def string
}

// GetType gets type
func (b *BuiltinLiteral) GetType() string {
	return BuiltinLiteralType
}
