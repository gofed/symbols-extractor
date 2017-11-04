// Version 01: Original version.
package main

import (
	"fmt"
)

type Shape interface {
	GetCoords() (float64, float64)
	Draw()
}

type struct Canvas {
	w, h uint
	data []byte
}

type struct Point {
	x, y float64
}

type struct Circle {
	*Point
	r float64
}

type struct Ellipsis {
	*Point
	e, f float64
}

func (p *Point) GetCoords() (x float64, y float64) {
	return p.x, p.y
}

func (p *Point) Draw() {
	canvas.data[p.x + py*canvas.w] = 1
}

func (c *Circle) GetCoords() (x float64, y float64) {
	return c.x, c.y
}

func (c *Circle) Draw() {
	canvas.data[p.x + py*canvas.w] = 1
}

func (e *Ellipsis) GetCoords() (x float64, y float64) {
	return e.x, e.y
}

func (e *Ellipsis) Draw() {
	canvas.data[p.x + py*canvas.w] = 1
}

var canvas = &Canvas{128, 128, make([]byte, 128*128)}

func main() {
	var e = &Ellipsis{16, 32, 10, 5}
	e.Draw()
}
/*
  Sequences:
    - sequences with asterisk are used in code

    (main.Circle.GetCoords, main.Point.GetCoords)
    (main.Ellipsis.GetCoords, main.Point.GetCoords)
    (main.Circle.Draw, main.Point.Draw)
    *(main.Ellipsis.Draw, main.Point.Draw)

    (main.Point.Draw.canvas.data, main.canvas.data)
    (main.Circle.Draw.canvas.data, main.canvas.data)
    *(main.Ellipsis.Draw.canvas.data, main.canvas.data)
*/
