package pkg

import "github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkgb"

type Imp struct {
	Name string
	Size int
	Imp  *pkgb.Imp
}

func Nic() *Imp {
	return &Imp{}
}
