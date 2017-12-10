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

	// pointers
	da := *ra

	// indeces
	la := list[0:1]
	lb := la[3]
	ma := mapV["3"]
	sa := "ahoj"[0]

	// type assertion
	asA := Int(uopb)
	asB := asA.(int)

	return b
}

type D struct{}

func (d *D) method() int { return 0 }

func frr() {
	ffA := func(a int) int { return 0 }
	ffB := ffA(2)

	// data type method
	frA := (*D).method

	// method invocation
	mA := D{}
	mB := mA.method()

	// struct type casting
	//sA := (struct{})(struct{}{})

	// chan type casting
	// c0 := make(chan int)
	// cA := (<-chan int)(c0)

	fmt.Print("Neco")
}

type D2 interface {
	imethod() int
}

type D3 struct{}

func (d *D3) imethod() int { return 0 }

func frr2() {
	var ia D2 = &D3{}
	ib := ia.imethod()
}

type D4 D3

func (d *D4) imethod() int { return 0 }

func frr2() {
	ida := D4{}
	idb := ida.imethod()

	idc := &ida
	idd := idc.imethod()
}

type D5 struct {
	d struct {
		a string
	}
}

func frr3() {
	ide := &struct{ d int }{2}
	idf := ide.d
}

type D6 string

func (d *D6) imethod() int { return 0 }

func frr3() {
	idg := D6("string")
	idh := idg.imethod()

	idi := struct{ d int }{2}
	idj := idi.d

	idk := (interface {
		imethod() int
	})(&D3{})
	idl := idk.imethod()
}
