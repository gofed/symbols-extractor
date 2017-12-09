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

	ra := &a

	chanA := make(chan int)
	chanValA := <-chanA

	uopa := ^1
	uopb := -1
	uopc := !true
	uopd := +1

	// case token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR:
	bopa := 1 == 2
	bopa = 1 != 2
	bopa = 1 <= 2
	bopa = 1 < 2
	bopa = 1 >= 2
	bopa = 1 > 2
	// case token.SHL, token.SHR
	bopb := 8.0 << 1
	bopb = 8.0 >> 1
	var bopc float32
	bopc = bopc << 1
	bopc = bopc >> 1
	// case token.AND, token.OR, token.MUL, token.SUB, token.QUO, token.ADD, token.AND_NOT, token.REM, token.XOR:
	bopd := true & false
	bopd = true | false
	bopd = true &^ false
	bopd = true ^ false
	bope := 1 * 1
	bope = 1 - 1
	bope = 1 / 1
	bope = 1 + 1
	bope = 1 % 1
	var bopf int16
	bope = bopf % 1
	bopd = true && false
	var bopg Int
	bope = bopg + 1
	bope = 1 + bopg

	fmt.Print(a, list, mapV, structV, structV2, listV2, ra)

	return b
}