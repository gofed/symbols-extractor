package testdata

// Test literal contract
var c = 2

// Test typed variable
var d int = 3

// Test binary contract
var e = 2 + 1

var f = d + 1

var ff = e + f

// Does not generate any contract
var j, k, l int

var a, b = 1, 2

var ta, tb float32 = 1, 2

func fnc() {
	// Test IsSendableTo contract
	var cv chan int
	cv <- 2

	// Test IsIncDecable contract
	c++
	c--

	l := []string{}
	v, ok := l[1]

	m := map[int]string{}
	mv, mok := m[1]

	id := interface{}(&c)
	intD, intOk := id.(*int)

	m[0] = "ahoj"
	s := struct{ a int }{a: 2}
	s.a = 2

	*(&c) = 2

	go func() {}()

	switch swA := 1; swA {
	case a:
	}

	switch sd := id.(type) {
	case *int:
	case *string, *float32:
	}

	switch id.(type) {
	case *int:
	}

	var c1 chan string
	select {
	case msg1 := <-c1:
	case l[0] = <-c1:
	case msg1, msg1Ok := <-c1:
	case l[0], ok = <-c1:
	case c1 <- "a":
	default:
	}

	for i := 0; i < 1; i++ {
	}

	for ii, vv := range m {
	}

	var fI int
	var fV string
	for fI, fV = range m {
	}

	for _, _ = range m {
	}

	for range m {
	}

	defer func() {}()

	if ifI := 1; ifI < 2 {
	} else {
	}
}
