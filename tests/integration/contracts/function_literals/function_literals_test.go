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
		"a":    ":45",
		"ffA":  ":33",
		"ffA2": ":77",
		"ffB":  ":70",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"../../testdata/function_literals.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int"}),
				Y: typevars.MakeLocalVar("a", vars["a"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Function{
					Params: []gotypes.DataType{
						&gotypes.Identifier{Package: "builtin", Def: "int"},
					},
					Results: []gotypes.DataType{
						&gotypes.Identifier{Package: "builtin", Def: "int"},
					},
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(1),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeLocalVar("ffA", vars["ffA"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeLocalVar("ffA", vars["ffA"]),
				Y: typevars.MakeLocalVar("ffA", vars["ffA2"]),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeLocalVar("ffA", vars["ffA2"]),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y: typevars.MakeArgument(typevars.MakeLocalVar("ffA", vars["ffA2"]), 0),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeLocalVar("ffA", vars["ffA2"]), 0),
				Y: typevars.MakeLocalVar("ffB", vars["ffB"]),
			},
		})
}
