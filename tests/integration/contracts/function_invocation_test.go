package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func TestFunctionInvocationContracts(t *testing.T) {
	var vars = map[string]string{
		"a":  ":179:a",
		"g":  ":20:g",
		"g1": ":69:g1",
		"g2": ":113:g2",
		"aa": ":287:aa",
		"ab": ":304:ab",
		"b":  ":321:b",
	}
	compareContracts(
		t,
		packageName,
		"function_invocation.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeVar(vars["a"], packageName),
			},
			//
			// aa := a + g1(a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(vars["g1"], packageName),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g1"], packageName, 0),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeReturn(vars["g1"], packageName, 0),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(vars["aa"], packageName),
			},
			//
			// ab := a + g2(a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(vars["g2"], packageName),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g2"], packageName, 0),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeReturn(vars["g2"], packageName, 0),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(vars["ab"], packageName),
			},
			//
			// b := a + g(a, a, a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(vars["g"], packageName),
				ArgsCount: 3,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g"], packageName, 0),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g"], packageName, 1),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g"], packageName, 2),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeReturn(vars["g"], packageName, 0),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeVar(vars["b"], packageName),
			},
		})
}
