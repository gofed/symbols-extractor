package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

var gVars = map[string]string{
	"c":  ":68:c",
	"d":  ":102:d",
	"e":  ":141:e",
	"f":  ":156:f",
	"ff": ":171:ff",
}

func TestGeneralContracts(t *testing.T) {
	compareContracts(
		t,
		packageName,
		"general.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["c"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["d"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(gVars["e"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(gVars["d"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(gVars["f"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(gVars["e"], packageName),
				Y:       typevars.MakeVar(gVars["f"], packageName),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeVar(gVars["ff"], packageName),
			},
		})
}
