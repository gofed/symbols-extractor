package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func TestBinaryOpsContracts(t *testing.T) {
	var vars = map[string]string{
		"bopa": ":120:bopa",
		"bopb": ":239:bopb",
		"bopc": ":278:bopc",
		"bopd": ":439:bopd",
		"bope": ":525:bope",
		"bopf": ":600:bopf",
		"bopg": ":655:bopg",
	}
	compareContracts(
		t,
		packageName,
		"binaryop.go",
		[]contracts.Contract{
			//
			// bopa := 1 == 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.EQL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 != 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.NEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 <= 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.LEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 < 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(4),
				OpToken: token.LSS,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 >= 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(5),
				OpToken: token.GEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 > 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(6),
				OpToken: token.GTR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(6),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopb := 8.0 << 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "float"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(7),
				OpToken: token.SHL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(7),
				Y: typevars.MakeVar(vars["bopb"], packageName),
			},
			//
			// bopb = 8.0 >> 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "float"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(8),
				OpToken: token.SHR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(8),
				Y: typevars.MakeVar(vars["bopb"], packageName),
			},
			//
			// bopc = bopc << 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopc"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(9),
				OpToken: token.SHL,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(9),
				Y: typevars.MakeVar(vars["bopc"], packageName),
			},
			//
			// bopc = bopc >> 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopc"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(10),
				OpToken: token.SHR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(10),
				Y: typevars.MakeVar(vars["bopc"], packageName),
			},
			// bopd := true & false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(11),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(11),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bopd = true | false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(12),
				OpToken: token.OR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(12),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bopd = true &^ false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(13),
				OpToken: token.AND_NOT,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(13),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bopd = true ^ false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(14),
				OpToken: token.XOR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(14),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bope := 1 * 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(15),
				OpToken: token.MUL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(15),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 - 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(16),
				OpToken: token.SUB,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(16),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 / 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(17),
				OpToken: token.QUO,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(17),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 + 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(18),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(18),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 % 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(19),
				OpToken: token.REM,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(19),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// var bopf int16
			// bope = bopf % 1
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopf"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(20),
				OpToken: token.REM,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(20),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bopd := true && false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(21),
				OpToken: token.LAND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(21),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// var bopg Int
			// bope = bopg + 1
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopg"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(22),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(22),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope = 1 + bopg
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVar(vars["bopg"], packageName),
				Z:       typevars.MakeVirtualVar(23),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(23),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
		})
}
