package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/unaryops"

func TestUnaryOpContracts(t *testing.T) {
	var vars = map[string]string{
		"a":        ":39",
		"ra":       ":52",
		"chanA":    ":63",
		"chanValA": ":88",
		"uopa":     ":110",
		"uopb":     ":122",
		"uopc":     ":134",
		"uopd":     ":149",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"../../testdata/unaryops.go",
		[]contracts.Contract{
			// a := "ahoj"
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeLocalVar("a", vars["a"]),
			},
			//
			// ra := &a
			//
			&contracts.IsReferenceable{
				X: typevars.MakeLocalVar("a", vars["a"]),
			},
			&contracts.ReferenceOf{
				X: typevars.MakeLocalVar("a", vars["a"]),
				Y: typevars.MakeVirtualVar(1),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeLocalVar("ra", vars["ra"]),
			},
			//
			// chanA := make(chan int)
			//
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Channel{
					Dir:   "3",
					Value: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeLocalVar("chanA", vars["chanA"]),
			},
			//
			// chanValA := <-chanA
			//
			&contracts.IsReceiveableFrom{
				X: typevars.MakeLocalVar("chanA", vars["chanA"]),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeLocalVar("chanValA", vars["chanValA"]),
			},
			//
			// uopa := ^1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(3),
				OpToken: token.XOR,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeLocalVar("uopa", vars["uopa"]),
			},
			//
			// uopb := -1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(4),
				OpToken: token.SUB,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeLocalVar("uopb", vars["uopb"]),
			},
			//
			// uopc := !true
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeVirtualVar(5),
				OpToken: token.NOT,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeLocalVar("uopc", vars["uopc"]),
			},
			//
			// uopd := +1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(6),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(6),
				Y: typevars.MakeLocalVar("uopd", vars["uopd"]),
			},
		})
}
