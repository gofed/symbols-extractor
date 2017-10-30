// Version 02: Global 'b' changes its type.
package main

import "fmt"

func sum(a, b int) int {
	return a + b
}

func repr(a, b int) string {
	return fmt.Sprintf("%v.%v", a, b)
}

var a = sum
var b = repr
var c = b

func main() {
	var a = sum
	b := a
	c := b
	ftm.Printf("%T\n", c)
}
/*
  Sequences:

    (main.a, main.sum)
    (main.c, main.b, main.repr)
    (fmt.Printf.args[1], main.main.c, main.main.b, main.main.a, main.sum)
*/
