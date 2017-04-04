package valid

type MyInt int

func binExpMyInt(x, y MyInt) interface{} {
	res := x * y
	return res
}

func binExpBuiltin(x, y int) interface{} {
	res := x * y
	return res
}
