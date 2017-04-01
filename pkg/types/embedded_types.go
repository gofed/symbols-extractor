package types

const BuiltinType = "builtin"

type Builtin struct{}

func (b *Builtin) GetType() string {
	return BuiltinType
}

const NilType = "nil"

type Nil struct {
}

func (n *Nil) GetType() string {
	return NilType
}
