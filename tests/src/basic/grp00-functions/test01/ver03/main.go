// Version 03: 'TA' and 'TB' used.
package main

import (
	"fmt"
	"strconv"
)

type FuncA func(string, float32) rune
type FuncB func(float32, string) bool

type TA struct {
	fn FuncA
}

type TB struct {
	fn FuncB
}

func resolve(a string, b float32) rune {
	if a == "abc" && b == 0.5 {
		return 'A'
	}
	return 'a'
}

func cmp(a float32, b string) bool {
	return a == strconv.ParseFloat(b, 32)
}

func main() {
	a := resolve
	b := cmp
	c := TA{a}
	d := c.fn
	e := TB{b}
	f := e.fn
	fmt.Printf("%T, %T\n", d, f)
}
/*
  Sequences:

    (main.TA.fn, main.main.a, main.resolve)
    (main.TB.fn, main.main.b, main.cmp)
*/
