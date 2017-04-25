package main

import (
	"fmt"
	"reflect"
)

type DataType interface {
	GetType() string
}

const IdentifierType = "identifier"

type Identifier struct {
	Def string `json:"def"`
}

func (o *Identifier) GetType() string {
	return IdentifierType
}

type LocalSlice []*Identifier

func main() {

	id := &Identifier{Def: "Poker"}

	tests := []struct {
		name, dataType string
		value          DataType
		empty          DataType
		ident          map[string]*Identifier
	}{
		// Simple identifier
		{
			name:  "Identifier",
			value: id,
			empty: &Identifier{},
			ident: map[string]*Identifier{
				"field1": &Identifier{
					Def: "Field1",
				},
				"field2": &Identifier{
					Def: "Field2",
				},
			},
		},
	}

	fmt.Printf("tests = %#v\n", tests)

	cl1 := []struct {
		field1 string
		field2 int
	}{
		{"", 0},
		{field1: "", field2: 0},
	}

	fmt.Printf("cl1: %#v\n", cl1)

	// only the array item type can be omitted, can not omit the struct in the item's struct
	cl2 := []struct {
		field1 string
		field2 struct {
			field1 string
			field2 int
		}
	}{
		{
			field1: "",
			field2: struct {
				field1 string
				field2 int
			}{"", 0}},
		{},
	}

	fmt.Printf("cl2: %#v\n", cl2)

	// double * does not work :)
	cl3 := []*struct {
		field1 string
		field2 int
	}{
		{"", 0},
		{field1: "", field2: 0},
	}

	fmt.Printf("cl3: %#v\n", cl3)

	cl4 := map[string]map[int]*struct {
		field1 string
		field2 int
	}{
		"f1": {
			0: {"", 0},
		},
	}

	fmt.Printf("cl4: %#v\n", cl4)

	hh := &struct {
		field1 string
		field2 int
	}{"", 3}

	// we can dereffer an anonymous type in parenthesis (but only once, &(&(...)) will not work)
	clhh := map[string]map[int]**struct {
		field1 string
		field2 int
	}{
		"f1": {
			0: &hh,
		},
	}

	fmt.Printf("clhh: %#v\n", clhh)

	cl5 := map[string]*map[int][]*map[string]*struct {
		field1 string
		field2 int
	}{
		"f1": {
			0: {
				33: {
					"1": {"", 0},
				},
			},
		},
	}

	fmt.Printf("cl5: %#v\n", cl5)

	// hhh := &(&struct {
	// 	field1 string
	// 	field2 int
	// }{field1: "", field2: 3})
	//hhh := &(&Identifier{"iii"})

	hhh := struct {
		field1 string
		field2 string
	}{"", ""}

	fmt.Printf("hhh: %#v\n", hhh)

	aaa := []*struct {
		field1 string
		field2 string
	}{{"", ""}}

	fmt.Printf("%#v\n", aaa)

	// This errors
	//bbb := []*Identifier{Identifier{""}}

	bbb := LocalSlice{
		{""},
	}

	fmt.Printf("%#v\n", bbb)

	if _ = 1 + 2; true {
		fmt.Printf("True\n")
	}

	// Playing with http://stackoverflow.com/questions/38664844/multiple-return-value-and-in-go
	type My int
	type My2 int

	res, err := "", My2(2)
	res2, err := "a", 2+2
	fmt.Printf("res: %v, err: %v\n", res, err)
	fmt.Printf("res2: %v, type(err): %v, err: %v\n", res2, reflect.TypeOf(err), err)

	type M1 struct {
		i int
	}

	r := M1{i: 3}
	r = struct{ i int }{i: 4}
	fmt.Printf("r: %v, type(r): %v\n", r, reflect.TypeOf(r))
}
