package unordered

import (
	"github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkg"
	"github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkgb"
)

type Struct struct {
	i MyInt
	//n http.MethodDelete
	impa *pkg.Imp
	impb *pkgb.Imp
}

type MyInt int

func Nic() *pkg.Imp {
	pkg.Nic().Imp.Size
	return &pkg.Imp{
		Name: "haluz",
		Size: 2,
	}
}
