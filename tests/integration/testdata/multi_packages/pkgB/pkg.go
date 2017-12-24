package pkgB

import (
	"pkgA"
)

func f() {
	a := &pkgA.A{}
	a.method(1, 2)

	b := pkgA.B{
		pkgA.A{1},
		{2},
	}
	c := pkgA.C{
		0: pkgA.B{
			{1},
			{},
		},
		1: {
			pkgA.A{f: 2},
		},
	}
}
