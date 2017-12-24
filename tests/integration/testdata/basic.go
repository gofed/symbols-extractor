package testdata

// Just get a list of all contracts and propagate all data types on more time

type A struct {
	i int
	s string
	f struct {
		f float32
	}
}

func (a *A) method(x, y int) int {
	return x + y
}

var C = &A{}

func g() int {
	a := &A{
		i: 1,
		s: "1",
		f: struct {
			f float32
		}{
			f: 1.0,
		},
	}

	b := a.method(a.i, a.i)
	return b
}
