package types

// For ast.BasicLit embedded types
// token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
const IntegerType = "integer"

type Integer struct {
}

func (i *Integer) GetType() string {
	return IntegerType
}

const FloatType = "float"

type Float struct {
}

func (f *Float) GetType() string {
	return FloatType
}

const ImagType = "imag"

type Imag struct {
}

func (i *Imag) GetType() string {
	return ImagType
}

const CharType = "char"

type Char struct {
}

func (c *Char) GetType() string {
	return CharType
}

const StringType = "string"

type String struct {
}

func (s *String) GetType() string {
	return StringType
}

const NilType = "nil"

type Nil struct {
}

func (n *Nil) GetType() string {
	return NilType
}
