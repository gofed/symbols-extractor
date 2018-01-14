package binaryops

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/binaryops"

	var vars = map[string]string{
		"bopa": ":120",
		"bopb": ":239",
		"bopc": ":278",
		"bopd": ":437",
		"bope": ":495",
		"bopf": ":570",
		"bopg": ":625",
	}

	// A := &gotypes.Identifier{Package: gopkg, Def: "A"}
	// ptrA := &gotypes.Pointer{Def: A}
	bInt := &gotypes.Identifier{Package: "builtin", Def: "int"}
	bBool := &gotypes.Identifier{Package: "builtin", Def: "bool"}
	//bTInt := &gotypes.Identifier{Package: "builtin", Def: "int"}
	bInt32 := &gotypes.Identifier{Package: "builtin", Def: "int32"}
	bInt16 := &gotypes.Identifier{Package: "builtin", Def: "int16"}
	Int := &gotypes.Identifier{Def: "Int", Package: gopkg}

	makeLocal := func(name string, dataType gotypes.DataType) cutils.VarTableTest {
		return cutils.VarTableTest{
			Name:     typevars.MakeLocalVar(name, vars[name]).String(),
			DataType: dataType,
		}
	}

	makeVirtual := func(idx int, dataType gotypes.DataType) cutils.VarTableTest {
		return cutils.VarTableTest{
			Name:     typevars.MakeVirtualVar(idx).String(),
			DataType: dataType,
		}
	}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/binaryop.go",
		[]cutils.VarTableTest{
			makeLocal("bopa", bBool),
			makeLocal("bopb", bInt),
			makeLocal("bopc", bInt32),
			makeLocal("bopd", bInt),
			makeLocal("bope", bInt),
			makeLocal("bopf", bInt16),
			makeLocal("bopg", Int),
			makeVirtual(1, &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: "false"}),
			makeVirtual(2, &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: "true"}),
			makeVirtual(3, &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: "true"}),
			makeVirtual(4, &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: "true"}),
			makeVirtual(5, &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: "false"}),
			makeVirtual(6, &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: "false"}),
			makeVirtual(7, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "16"}),
			makeVirtual(8, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "4"}),
			makeVirtual(9, bInt32),
			makeVirtual(10, bInt32),
			makeVirtual(11, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "0"}),
			makeVirtual(12, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "1"}),
			makeVirtual(13, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "1"}),
			makeVirtual(14, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "1"}),
			makeVirtual(15, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "1"}),
			makeVirtual(16, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "0"}),
			makeVirtual(17, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "1"}),
			makeVirtual(18, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "2"}),
			makeVirtual(19, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "0"}),
			makeVirtual(20, bInt16),
			makeVirtual(21, &gotypes.Constant{Package: "builtin", Def: "bool", Untyped: true, Literal: "false"}),
			makeVirtual(22, Int),
			makeVirtual(23, Int),
		},
	)
}
