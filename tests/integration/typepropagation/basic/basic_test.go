package basic

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/basic"

	var vars = map[string]string{
		"a":  ":167",
		"x":  ":180",
		"y":  ":183",
		"a2": ":243",
		"b2": ":323",
		"C":  ":217",
	}

	A := &gotypes.Identifier{Package: gopkg, Def: "A"}
	ptrA := &gotypes.Pointer{Def: A}
	bInt := &gotypes.Builtin{Def: "int"}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/basic.go",
		[]cutils.VarTableTest{
			{
				Name:     typevars.MakeLocalVar("a", vars["a"]).String(),
				DataType: ptrA,
			},
			{
				Name:     typevars.MakeLocalVar("x", vars["x"]).String(),
				DataType: bInt,
			},
			{
				Name:     typevars.MakeLocalVar("y", vars["y"]).String(),
				DataType: bInt,
			},
			{
				Name:     typevars.MakeVar(gopkg, "C", vars["C"]).String(),
				DataType: ptrA,
			},
			{
				Name:     typevars.MakeLocalVar("a", vars["a2"]).String(),
				DataType: ptrA,
			},
			{
				Name:     typevars.MakeLocalVar("b", vars["b2"]).String(),
				DataType: bInt,
			},
			// virtual variables
			{
				Name:     typevars.MakeVirtualVar(1).String(),
				DataType: A,
			},
			{
				Name:     typevars.MakeVirtualVar(2).String(),
				DataType: ptrA,
			},
			{
				Name:     typevars.MakeVirtualVar(3).String(),
				DataType: bInt,
			},
			{
				Name:     typevars.MakeVirtualVar(4).String(),
				DataType: A,
			},
			{
				Name: typevars.MakeVirtualVar(5).String(),
				DataType: &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "f",
							Def:  &gotypes.Builtin{Def: "float32"},
						},
					},
				},
			},
			{
				Name:     typevars.MakeVirtualVar(6).String(),
				DataType: ptrA,
			},
			{
				Name: typevars.MakeVirtualVar(7).String(),
				DataType: &gotypes.Method{
					Receiver: ptrA,
					Def: &gotypes.Function{
						Package: gopkg,
						Params: []gotypes.DataType{
							bInt,
							bInt,
						},
						Results: []gotypes.DataType{
							bInt,
						},
					},
				},
			},
		},
	)
}
