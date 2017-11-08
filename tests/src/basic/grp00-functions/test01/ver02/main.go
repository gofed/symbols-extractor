// Version 02: 'TB' used.
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
	x, _ := strconv.ParseFloat(b, 32)
	return a == x
}

func main() {
	a := cmp
	b := a
	c := TB{b}
	d := c.fn
	fmt.Printf("%T\n", d)
}
/*
  Sequences:

    (main.TB.fn, main.main.b, main.main.a, main.cmp)
*/
