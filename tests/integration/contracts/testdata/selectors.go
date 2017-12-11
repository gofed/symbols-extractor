package testdata

import "fmt"

type D struct{}

func (d *D) method() int { return 0 }

func frr() {
	// data type method
	frA := (*D).method

	// method invocation
	mA := D{}
	mB := mA.method()

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
