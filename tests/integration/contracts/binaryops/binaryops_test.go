package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/binaryops"

func TestBinaryOpsContracts(t *testing.T) {
	var vars = map[string]string{
		"bopa": ":120",
		"bopb": ":239",
		"bopc": ":278",
		"bopd": ":437",
		"bope": ":495",
		"bopf": ":570",
		"bopg": ":625",
	}

	utils.ParseAndCompareContracts(
		t,
		packageName,
		"../../testdata/binaryop.go",
		[]contracts.Contract{
			//
			// bopa := 1 == 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.EQL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeLocalVar("bopa", vars["bopa"]),
			},
			//
			// bopa = 1 != 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.NEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeLocalVar("bopa", vars["bopa"]),
			},
			//
			// bopa = 1 <= 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.LEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeLocalVar("bopa", vars["bopa"]),
			},
			//
			// bopa = 1 < 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Z:       typevars.MakeVirtualVar(4),
				OpToken: token.LSS,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeLocalVar("bopa", vars["bopa"]),
			},
			//
			// bopa = 1 >= 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Z:       typevars.MakeVirtualVar(5),
				OpToken: token.GEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeLocalVar("bopa", vars["bopa"]),
			},
			//
			// bopa = 1 > 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Z:       typevars.MakeVirtualVar(6),
				OpToken: token.GTR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(6),
				Y: typevars.MakeLocalVar("bopa", vars["bopa"]),
			},
			//
			// bopb := 8.0 << 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "float64", Literal: "8.0"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(7),
				OpToken: token.SHL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(7),
				Y: typevars.MakeLocalVar("bopb", vars["bopb"]),
			},
			//
			// bopb = 8.0 >> 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "float64", Literal: "8.0"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(8),
				OpToken: token.SHR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(8),
				Y: typevars.MakeLocalVar("bopb", vars["bopb"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int32"}),
				Y: typevars.MakeLocalVar("bopc", vars["bopc"]),
			},
			//
			// bopc = bopc << 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("bopc", vars["bopc"]),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(9),
				OpToken: token.SHL,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(9),
				Y: typevars.MakeLocalVar("bopc", vars["bopc"]),
			},
			//
			// bopc = bopc >> 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("bopc", vars["bopc"]),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(10),
				OpToken: token.SHR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(10),
				Y: typevars.MakeLocalVar("bopc", vars["bopc"]),
			},
			// bopd := 1 & 0
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Z:       typevars.MakeVirtualVar(11),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(11),
				Y: typevars.MakeLocalVar("bopd", vars["bopd"]),
			},
			// bopd = 1 | 0
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Z:       typevars.MakeVirtualVar(12),
				OpToken: token.OR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(12),
				Y: typevars.MakeLocalVar("bopd", vars["bopd"]),
			},
			// bopd = 1 &^ 0
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Z:       typevars.MakeVirtualVar(13),
				OpToken: token.AND_NOT,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(13),
				Y: typevars.MakeLocalVar("bopd", vars["bopd"]),
			},
			// bopd = 1 ^ 0
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Z:       typevars.MakeVirtualVar(14),
				OpToken: token.XOR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(14),
				Y: typevars.MakeLocalVar("bopd", vars["bopd"]),
			},
			// bope := 1 * 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(15),
				OpToken: token.MUL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(15),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
			// bope := 1 - 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(16),
				OpToken: token.SUB,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(16),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
			// bope := 1 / 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(17),
				OpToken: token.QUO,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(17),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
			// bope := 1 + 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(18),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(18),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
			// bope := 1 % 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(19),
				OpToken: token.REM,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(19),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
			// var bopf int16
			// bope = bopf % 1
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int16"}),
				Y: typevars.MakeLocalVar("bopf", vars["bopf"]),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("bopf", vars["bopf"]),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(20),
				OpToken: token.REM,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(20),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
			// bopd := true && false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "bool", Literal: "true"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "bool", Literal: "false"}),
				Z:       typevars.MakeVirtualVar(21),
				OpToken: token.LAND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(21),
				Y: typevars.MakeLocalVar("bopd", vars["bopd"]),
			},
			// var bopg Int
			// bope = bopg + 1
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: packageName, Def: "Int"}),
				Y: typevars.MakeLocalVar("bopg", vars["bopg"]),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("bopg", vars["bopg"]),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(22),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(22),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
			// bope = 1 + bopg
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y:       typevars.MakeLocalVar("bopg", vars["bopg"]),
				Z:       typevars.MakeVirtualVar(23),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(23),
				Y: typevars.MakeLocalVar("bope", vars["bope"]),
			},
		})
}
