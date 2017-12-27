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
		"bopd": ":439",
		"bope": ":525",
		"bopf": ":600",
		"bopg": ":655",
	}

	// A := &gotypes.Identifier{Package: gopkg, Def: "A"}
	// ptrA := &gotypes.Pointer{Def: A}
	bInt := &gotypes.Builtin{Def: "int", Untyped: true}
	bBool := &gotypes.Builtin{Def: "bool"}
	//bTInt := &gotypes.Builtin{Def: "int"}
	bFloat := &gotypes.Builtin{Def: "float32"}
	bInt16 := &gotypes.Builtin{Def: "int16"}
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
			makeLocal("bopc", bFloat),
			makeLocal("bopd", bBool),
			makeLocal("bope", bInt),
			makeLocal("bopf", bInt16),
			makeLocal("bopg", Int),
			makeVirtual(1, bBool),
			makeVirtual(2, bBool),
			makeVirtual(3, bBool),
			makeVirtual(4, bBool),
			makeVirtual(5, bBool),
			makeVirtual(6, bBool),
			makeVirtual(7, bInt),
			makeVirtual(8, bInt),
			makeVirtual(9, bFloat),
			makeVirtual(10, bFloat),
			makeVirtual(11, bBool),
			makeVirtual(12, bBool),
			makeVirtual(13, bBool),
			makeVirtual(14, bBool),
			makeVirtual(15, bInt),
			makeVirtual(16, bInt),
			makeVirtual(17, bInt),
			makeVirtual(18, bInt),
			makeVirtual(19, bInt),
			makeVirtual(20, bInt16),
			makeVirtual(21, bBool),
			makeVirtual(22, Int),
			makeVirtual(23, Int),
		},
	)
}
