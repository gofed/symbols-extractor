package type_casting

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/type_casting"

	var vars = map[string]string{
		"asA": typevars.MakeLocalVar("asA", ":64").String(),
		"asB": typevars.MakeLocalVar("asB", ":79").String(),
	}

	makeLocal := func(name string, dataType gotypes.DataType) cutils.VarTableTest {
		return cutils.VarTableTest{
			Name:     vars[name],
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
		"../../testdata/type_casting.go",
		[]cutils.VarTableTest{
			makeLocal("asA", &gotypes.Identifier{Def: "Int", Package: gopkg}),
			makeLocal("asB", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeVirtual(1, &gotypes.Constant{Package: gopkg, Untyped: false, Def: "Int", Literal: "1"}),
		},
	)
}

func TestUntypedToUintPropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/type_casting"

	makeVirtual := func(idx int, dataType gotypes.DataType) cutils.VarTableTest {
		return cutils.VarTableTest{
			Name:     typevars.MakeVirtualVar(idx).String(),
			DataType: dataType,
		}
	}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/untyped.go",
		[]cutils.VarTableTest{
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "pallocChunksL1Bits", ":28").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "13"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "_PageShift", ":54").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "13"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "GoosDarwin", ":80").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "0"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "GoarchArm64", ":105").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "0"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "GoarchMipsle", ":130").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "0"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "GoarchMips", ":155").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "0"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "GoarchWasm", ":180").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "0"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "_64bit", ":205").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "2"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "pageShift", ":257").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "13"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "logPallocChunkPages", ":291").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "9"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "logPallocChunkBytes", ":316").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "22"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "heapAddrBits", ":371").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "64"},
			},
			cutils.VarTableTest{
				Name:     typevars.MakeVar(gopkg, "pallocChunksL2Bits", ":531").String(),
				DataType: &gotypes.Constant{Def: "int", Package: "builtin", Untyped: true, Literal: "29"},
			},

			makeVirtual(1, &gotypes.Constant{Package: "builtin", Untyped: false, Def: "uintptr", Literal: "0"}),
			makeVirtual(2, &gotypes.Constant{Package: "builtin", Untyped: false, Def: "uintptr", Literal: "18446744073709551615"}),
			makeVirtual(3, &gotypes.Constant{Package: "builtin", Untyped: false, Def: "uintptr", Literal: "2"}),
			makeVirtual(4, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "4"}),
			makeVirtual(5, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
			makeVirtual(6, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "22"}),
			makeVirtual(7, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
			makeVirtual(8, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
			makeVirtual(9, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
			makeVirtual(10, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
			makeVirtual(11, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
			makeVirtual(12, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "96"}),
			makeVirtual(13, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "-1"}),
			makeVirtual(14, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "-1"}),
			makeVirtual(15, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
			makeVirtual(16, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "32"}),
			makeVirtual(17, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "-32"}),
			makeVirtual(18, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "64"}),
			makeVirtual(19, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
			makeVirtual(20, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
			makeVirtual(21, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "64"}),
			makeVirtual(22, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "42"}),
			makeVirtual(23, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "29"}),

			makeVirtual(24, &gotypes.Constant{Package: "builtin", Untyped: false, Def: "uint", Literal: "1"}),
			makeVirtual(25, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "536870912"}),
			makeVirtual(26, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "536870911"}),
			makeVirtual(27, &gotypes.Constant{Package: "builtin", Untyped: false, Def: "uint", Literal: "1"}),
		},
	)
}
