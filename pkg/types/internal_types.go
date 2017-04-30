package types

const NilType = "nil"

type Nil struct {
}

func (n *Nil) GetType() string {
	return NilType
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
