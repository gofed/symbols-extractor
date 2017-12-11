package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/testdata"

func TestTypeCastingContracts(t *testing.T) {
	var vars = map[string]string{
		"asA": ":64:asA",
		"asB": ":79:asB",
	}
	utils.CompareContracts(
		t,
		packageName,
		"type_casting.go",
		[]contracts.Contract{
			// asA := Int(1)
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Identifier{
					Def:     "Int",
					Package: packageName,
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Identifier{
					Def:     "Int",
					Package: packageName,
				}),
				Y: typevars.MakeVar(vars["asA"], packageName),
			},
			// asB := asA.(int)
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["asA"], packageName),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "int"}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeVar(vars["asB"], packageName),
			},
		})
}
