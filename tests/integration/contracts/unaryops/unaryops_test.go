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
		"a":        ":39:a",
		"ra":       ":52:ra",
		"chanA":    ":63:chanA",
		"chanValA": ":88:chanValA",
		"uopa":     ":110:uopa",
		"uopb":     ":122:uopb",
		"uopc":     ":134:uopc",
		"uopd":     ":149:uopd",
	}
	utils.CompareContracts(
		t,
		packageName,
		"testdata/unaryops.go",
		[]contracts.Contract{
			// a := "ahoj"
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeVar(vars["a"], packageName),
			},
			//
			// ra := &a
			//
			&contracts.UnaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeVirtualVar(1),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(vars["ra"], packageName),
			},
			//
			// chanA := make(chan int)
			//
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Channel{
					Dir:   "3",
					Value: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeVar(vars["chanA"], packageName),
			},
			//
			// chanValA := <-chanA
			//
			&contracts.IsReceiveableFrom{
				X: typevars.MakeVar(vars["chanA"], packageName),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(vars["chanValA"], packageName),
			},
			//
			// uopa := ^1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(3),
				OpToken: token.XOR,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeVar(vars["uopa"], packageName),
			},
			//
			// uopb := -1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(4),
				OpToken: token.SUB,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeVar(vars["uopb"], packageName),
			},
			//
			// uopc := !true
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeVirtualVar(5),
				OpToken: token.NOT,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeVar(vars["uopc"], packageName),
			},
			//
			// uopd := +1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(6),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(6),
				Y: typevars.MakeVar(vars["uopd"], packageName),
			},
		})
}
