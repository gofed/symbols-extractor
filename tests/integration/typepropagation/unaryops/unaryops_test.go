package unaryops

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/unaryops"

	var vars = map[string]string{
		"a":        typevars.MakeLocalVar("a", ":39").String(),
		"chanA":    typevars.MakeLocalVar("chanA", ":63").String(),
		"chanValA": typevars.MakeLocalVar("chanValA", ":88").String(),
		"ra":       typevars.MakeLocalVar("ra", ":52").String(),
		"uopa":     typevars.MakeLocalVar("uopa", ":110").String(),
		"uopb":     typevars.MakeLocalVar("uopb", ":122").String(),
		"uopc":     typevars.MakeLocalVar("uopc", ":134").String(),
		"uopd":     typevars.MakeLocalVar("uopd", ":149").String(),
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

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/unaryops.go",
		[]cutils.VarTableTest{
			makeLocal("a", &gotypes.Builtin{Def: "string", Untyped: true}),
			makeLocal("ra", &gotypes.Pointer{Def: &gotypes.Builtin{Def: "string", Untyped: true}}),
			makeLocal("chanA", &gotypes.Channel{Dir: "3", Value: &gotypes.Builtin{Def: "int"}}),
			makeLocal("chanValA", &gotypes.Builtin{Def: "int"}),
			makeLocal("uopa", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("uopb", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("uopc", &gotypes.Builtin{Def: "bool"}),
			makeLocal("uopd", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeVirtual(1, &gotypes.Pointer{Def: &gotypes.Builtin{Def: "string", Untyped: true}}),
			makeVirtual(2, &gotypes.Builtin{Def: "int"}),
			makeVirtual(3, &gotypes.Builtin{Def: "int", Untyped: true}),
			makeVirtual(4, &gotypes.Builtin{Def: "int", Untyped: true}),
			makeVirtual(5, &gotypes.Builtin{Def: "bool"}),
			makeVirtual(6, &gotypes.Builtin{Def: "int", Untyped: true}),
		},
	)
}
