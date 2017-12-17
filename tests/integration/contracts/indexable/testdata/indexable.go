package testdata

type Int int

func f() {
	list := []Int{1}
	mapV := map[string]int{"3": 3}
	// indeces
	la := list[0:1]
	lb := la[3]
	ma := mapV["3"]
	sa := "ahoj"[0]
}
