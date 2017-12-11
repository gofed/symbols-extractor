package testdata

type Int int

func f() {
	// type assertion
	asA := Int(1)
	asB := asA.(int)
}
