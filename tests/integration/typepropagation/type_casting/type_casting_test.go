package type_casting

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/type_casting"

	var vars = map[string]string{
		"asA": typevars.MakeLocalVar("asA", ":64").String(),
		"asB": typevars.MakeLocalVar("asB", ":79").String(),
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
		"../../testdata/type_casting.go",
		[]cutils.VarTableTest{
			makeLocal("asA", &gotypes.Identifier{Def: "Int", Package: gopkg}),
			makeLocal("asB", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeVirtual(1, &gotypes.Constant{Package: gopkg, Untyped: false, Def: "Int", Literal: "1"}),
		},
	)
}
