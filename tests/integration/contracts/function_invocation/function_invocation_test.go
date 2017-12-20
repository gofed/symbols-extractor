package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/function_invocation"

func TestFunctionInvocationContracts(t *testing.T) {
	var vars = map[string]string{
		"a":  ":179",
		"g":  ":20",
		"g1": ":69",
		"g2": ":113",
		"aa": ":287",
		"ab": ":304",
		"b":  ":321",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"testdata/function_invocation.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeLocalVar("a", vars["a"]),
			},
			//
			// aa := a + g1(a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(packageName, "g1", vars["g1"]),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeLocalVar("a", vars["a"]),
				Y: typevars.MakeArgument(typevars.MakeFunction(packageName, "g1", vars["g1"]), 0),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("a", vars["a"]),
				Y:       typevars.MakeReturn(typevars.MakeFunction(packageName, "g1", vars["g1"]), 0),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeLocalVar("aa", vars["aa"]),
			},
			//
			// ab := a + g2(a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(packageName, "g2", vars["g2"]),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeLocalVar("a", vars["a"]),
				Y: typevars.MakeArgument(typevars.MakeFunction(packageName, "g2", vars["g2"]), 0),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("a", vars["a"]),
				Y:       typevars.MakeReturn(typevars.MakeFunction(packageName, "g2", vars["g2"]), 0),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeLocalVar("ab", vars["ab"]),
			},
			//
			// b := a + g(a, a, a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(packageName, "g", vars["g"]),
				ArgsCount: 3,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeLocalVar("a", vars["a"]),
				Y: typevars.MakeArgument(typevars.MakeFunction(packageName, "g", vars["g"]), 0),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeLocalVar("a", vars["a"]),
				Y: typevars.MakeArgument(typevars.MakeFunction(packageName, "g", vars["g"]), 1),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeLocalVar("a", vars["a"]),
				Y: typevars.MakeArgument(typevars.MakeFunction(packageName, "g", vars["g"]), 2),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("a", vars["a"]),
				Y:       typevars.MakeReturn(typevars.MakeFunction(packageName, "g", vars["g"]), 0),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeLocalVar("b", vars["b"]),
			},
		})
}
