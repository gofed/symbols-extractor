package indexable

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/indexable"

	var vars = map[string]string{
		"la":   typevars.MakeLocalVar("la", ":107").String(),
		"lb":   typevars.MakeLocalVar("lb", ":124").String(),
		"list": typevars.MakeLocalVar("list", ":45").String(),
		"ma":   typevars.MakeLocalVar("ma", ":137").String(),
		"mapV": typevars.MakeLocalVar("mapV", ":63").String(),
		"sa":   typevars.MakeLocalVar("sa", ":154").String(),
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

	Int := &gotypes.Identifier{Package: gopkg, Def: "Int"}
	Slice := &gotypes.Slice{Elmtype: Int}
	Map := &gotypes.Map{
		Keytype:   &gotypes.Builtin{Def: "string"},
		Valuetype: &gotypes.Builtin{Def: "int"},
	}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/indexable.go",
		[]cutils.VarTableTest{
			makeLocal("la", Slice),
			makeLocal("list", Slice),
			makeLocal("mapV", Map),
			makeLocal("lb", Int),
			makeLocal("ma", &gotypes.Builtin{Def: "int"}),
			makeLocal("sa", &gotypes.Builtin{Def: "uint8"}),
			makeVirtual(1, Slice),
			makeVirtual(2, Map),
			makeVirtual(3, &gotypes.Builtin{Def: "string", Untyped: true}),
			makeVirtual(4, &gotypes.Builtin{Def: "string", Untyped: true}),
		},
	)
}
