package contracts

func g(a, b, c string) string {
	return "kjj"
}

func g1(a string) string {
	return "kjj"
}

func g2(a string, b ...int) string {
	return "kjj"
}

func f() {
	a := "ahoj"
	// This panics cause it takes the f variable, not the f function, it's a bug
	// b := a + f()
	aa := a + g1(a)
	ab := a + g2(a)
	b := a + g(a, a, a)
}
