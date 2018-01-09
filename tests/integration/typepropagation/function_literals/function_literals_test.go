package function_literals

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/function_literals"

	var vars = map[string]string{
		"ffA": typevars.MakeLocalVar("ffA", ":33").String(),
		"a":   typevars.MakeLocalVar("a", ":45").String(),
		"ffB": typevars.MakeLocalVar("ffB", ":70").String(),
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

	F := &gotypes.Function{
		Package: gopkg,
		Params: []gotypes.DataType{
			&gotypes.Identifier{Package: "builtin", Def: "int"},
		},
		Results: []gotypes.DataType{
			&gotypes.Identifier{Package: "builtin", Def: "int"},
		},
	}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/function_literals.go",
		[]cutils.VarTableTest{
			makeLocal("a", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("ffA", F),
			makeLocal("ffB", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeVirtual(1, F),
		},
	)
}
