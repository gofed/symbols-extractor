package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func TestFunctionLiteralsContracts(t *testing.T) {
	var vars = map[string]string{
		"ffA": ":33:ffA",
		"ffB": ":70:ffB",
	}
	compareContracts(
		t,
		packageName,
		"function_literals.go",
		[]contracts.Contract{
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
				Y: typevars.MakeVirtualFunction(typevars.MakeVirtualVar(1)),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualFunction(typevars.MakeVirtualVar(1)),
				Y: typevars.MakeVar(vars["ffA"], packageName),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(vars["ffA"], packageName),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeArgument(vars["ffA"], packageName, 0),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(vars["ffA"], packageName, 0),
				Y: typevars.MakeVar(vars["ffB"], packageName),
			},
		})
}
