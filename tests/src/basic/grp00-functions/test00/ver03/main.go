// Version 03: local 'c' changed, local 'b' used.
package main

import "fmt"

func sum(a, b int) int {
	return a + b
}

func repr(a, b int) string {
	return fmt.Sprintf("%v.%v", a, b)
}

var a = sum
var b = a
var c = b

func main() {
	var a = sum
	b := a
	c := repr
	ftm.Printf("%T, %T\n", c, b)
}
