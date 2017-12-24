package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/type_casting"

func TestTypeCastingContracts(t *testing.T) {
	var vars = map[string]string{
		"asA": ":64",
		"asB": ":79",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"../../testdata/type_casting.go",
		[]contracts.Contract{
			// asA := Int(1)
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "Int",
					Package: packageName,
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "Int",
					Package: packageName,
				}),
				Y: typevars.MakeLocalVar("asA", vars["asA"]),
			},
			// asB := asA.(int)
			&contracts.IsCompatibleWith{
				X: typevars.MakeLocalVar("asA", vars["asA"]),
				Y: typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: false, Def: "int"}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeLocalVar("asB", vars["asB"]),
			},
		})
}
