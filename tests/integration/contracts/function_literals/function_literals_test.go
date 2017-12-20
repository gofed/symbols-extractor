package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/function_literals"

func TestFunctionLiteralsContracts(t *testing.T) {
	var vars = map[string]string{
		"a":   ":45",
		"ffA": ":33",
		"ffB": ":70",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"testdata/function_literals.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Def: "int"}),
				Y: typevars.MakeLocalVar("a", vars["a"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Function{
					Params: []gotypes.DataType{
						&gotypes.Builtin{Untyped: false, Def: "int"},
					},
					Results: []gotypes.DataType{
						&gotypes.Builtin{Untyped: false, Def: "int"},
					},
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(1),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualFunction(typevars.MakeVirtualVar(1)),
				Y: typevars.MakeLocalVar("ffA", vars["ffA"]),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(packageName, "ffA", vars["ffA"]),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeArgument(typevars.MakeFunction(packageName, "ffA", vars["ffA"]), 0),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeFunction(packageName, "ffA", vars["ffA"]), 0),
				Y: typevars.MakeLocalVar("ffB", vars["ffB"]),
			},
		})
}
