package contracts

import (
	"fmt"
)

// Test literal contract
var c = 2

// Test typed variable
var d int = 3

// Test binary contract
var e = 2 + 1

var f = d + 1

var ff = e + f

func g(a, b, c string) string {
	return "kjj"
}

func g1(a string) string {
	return "kjj"
}

func g2(a string, b ...int) string {
	return "kjj"
}

type Int int

func f() string {
	a := "ahoj"
	// This panics cause it takes the f variable, not the f function, it's a bug
	// b := a + f()
	aa := a + g1(a)
	ab := a + g2(a)
	b := a + g(a, a, a)

	list := []Int{
		1,
		2,
	}

	mapV := map[string]int{
		"3": 3,
		"4": 4,
	}

	structV := struct {
		key1 string
		key2 int
	}{
		key1: "key1",
		key2: 2,
	}

	structV2 := struct {
		key1 string
		key2 int
	}{
		"key1",
		2,
	}

	listV2 := [][]int{
		{
			1,
			2,
		},
		{
			3,
			4,
		},
	}

	fmt.Printf("a: %v\n", a)
	fmt.Printf("list: %v\n", list)
	fmt.Printf("mapV: %v\n", mapV)
	fmt.Printf("structV: %v\n", structV)
	fmt.Printf("structV2: %v\n", structV2)
	fmt.Printf("listV2: %v\n", listV2)

	return b
}
