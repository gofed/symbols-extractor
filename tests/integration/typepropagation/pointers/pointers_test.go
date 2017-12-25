package pointers

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/pointers"

	var vars = map[string]string{
		"a":  typevars.MakeLocalVar("a", ":39").String(),
		"da": typevars.MakeLocalVar("da", ":75").String(),
		"ra": typevars.MakeLocalVar("ra", ":52").String(),
	}

	makeLocal := func(name string, dataType gotypes.DataType) cutils.VarTableTest {
		return cutils.VarTableTest{
			Name:     vars[name],
			DataType: dataType,
		}
	}

	makeVirtual := func(idx int, dataType gotypes.DataType) cutils.VarTableTest {
		return cutils.VarTableTest{
			Name:     typevars.MakeVirtualVar(idx).String(),
			DataType: dataType,
		}
	}

	Pointer := &gotypes.Pointer{Def: &gotypes.Builtin{Def: "string", Untyped: true}}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/pointers.go",
		[]cutils.VarTableTest{
			makeLocal("a", &gotypes.Builtin{Def: "string", Untyped: true}),
			makeLocal("da", &gotypes.Builtin{Def: "string", Untyped: true}),
			makeLocal("ra", Pointer),
			makeVirtual(1, Pointer),
			makeVirtual(2, &gotypes.Builtin{Def: "string", Untyped: true}),
		},
	)
}
