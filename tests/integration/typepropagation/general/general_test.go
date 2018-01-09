package general

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/general"

	var vars = map[string]string{
		"c1":      typevars.MakeLocalVar("c1", ":722").String(),
		"cv":      typevars.MakeLocalVar("cv", ":307").String(),
		"fI":      typevars.MakeLocalVar("fI", ":932").String(),
		"fV":      typevars.MakeLocalVar("fV", ":944").String(),
		"i":       typevars.MakeLocalVar("i", ":873").String(),
		"id":      typevars.MakeLocalVar("id", ":446").String(),
		"ifI":     typevars.MakeLocalVar("ifI", ":1051").String(),
		"ii":      typevars.MakeLocalVar("ii", ":903").String(),
		"intD":    typevars.MakeLocalVar("intD", ":469").String(),
		"intOk":   typevars.MakeLocalVar("intOk", ":475").String(),
		"l":       typevars.MakeLocalVar("l", ":372").String(),
		"m":       typevars.MakeLocalVar("m", ":405").String(),
		"mok":     typevars.MakeLocalVar("mok", ":432").String(),
		"msg1_c1": typevars.MakeLocalVar("msg1", ":753").String(),
		"msg1_c3": typevars.MakeLocalVar("msg1", ":792").String(),
		"msg1Ok":  typevars.MakeLocalVar("msg1Ok", ":798").String(),
		"mv":      typevars.MakeLocalVar("mv", ":428").String(),
		"ok":      typevars.MakeLocalVar("ok", ":392").String(),
		"s":       typevars.MakeLocalVar("s", ":511").String(),
		"sd_c1":   typevars.MakeLocalVar("sd", ":641").String(),
		"sd_c2":   typevars.MakeLocalVar("sd", ":653").String(),
		"swA":     typevars.MakeLocalVar("swA", ":585").String(),
		"v":       typevars.MakeLocalVar("v", ":389").String(),
		"vv":      typevars.MakeLocalVar("vv", ":907").String(),
		"a":       typevars.MakeVar(gopkg, "a", ":218").String(),
		"b":       typevars.MakeVar(gopkg, "b", ":221").String(),
		"c":       typevars.MakeVar(gopkg, "c", ":48").String(),
		"d":       typevars.MakeVar(gopkg, "d", ":82").String(),
		"e":       typevars.MakeVar(gopkg, "e", ":121").String(),
		"f":       typevars.MakeVar(gopkg, "f", ":136").String(),
		"ff":      typevars.MakeVar(gopkg, "ff", ":151").String(),
		"j":       typevars.MakeVar(gopkg, "j", ":201").String(),
		"k":       typevars.MakeVar(gopkg, "k", ":204").String(),
		"g_l":     typevars.MakeVar(gopkg, "l", ":207").String(),
		"ta":      typevars.MakeVar(gopkg, "ta", ":235").String(),
		"tb":      typevars.MakeVar(gopkg, "tb", ":239").String(),
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
		"../../testdata/general.go",
		[]cutils.VarTableTest{
			makeLocal("c1", &gotypes.Channel{Dir: "3", Value: &gotypes.Identifier{Package: "builtin", Def: "string"}}),
			makeLocal("cv", &gotypes.Channel{Dir: "3", Value: &gotypes.Identifier{Package: "builtin", Def: "int"}}),
			makeLocal("fI", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("fV", &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeLocal("i", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("id", &gotypes.Interface{}),
			makeLocal("ifI", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("ii", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("intD", &gotypes.Pointer{Def: &gotypes.Identifier{Package: "builtin", Def: "int"}}),
			makeLocal("intOk", &gotypes.Identifier{Package: "builtin", Def: "bool"}),
			makeLocal("l", &gotypes.Slice{Elmtype: &gotypes.Identifier{Package: "builtin", Def: "string"}}),
			makeLocal("m", &gotypes.Map{Keytype: &gotypes.Identifier{Package: "builtin", Def: "int"}, Valuetype: &gotypes.Identifier{Package: "builtin", Def: "string"}}),
			makeLocal("mok", &gotypes.Identifier{Package: "builtin", Def: "bool"}),
			makeLocal("msg1_c1", &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeLocal("msg1_c3", &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeLocal("msg1Ok", &gotypes.Identifier{Package: "builtin", Def: "bool"}),
			makeLocal("mv", &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeLocal("ok", &gotypes.Identifier{Package: "builtin", Def: "bool"}),
			makeLocal("s", &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "a",
						Def:  &gotypes.Identifier{Package: "builtin", Def: "int"},
					},
				}}),
			makeLocal("sd_c1", &gotypes.Pointer{Def: &gotypes.Identifier{Package: "builtin", Def: "int"}}),
			makeLocal("sd_c2", &gotypes.Interface{}),
			makeLocal("swA", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("v", &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeLocal("vv", &gotypes.Identifier{Package: "builtin", Def: "string"}),

			makeLocal("c", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("d", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("e", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("f", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("ff", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("j", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("k", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("g_l", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("a", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("b", &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeLocal("ta", &gotypes.Identifier{Package: "builtin", Def: "float32"}),
			makeLocal("tb", &gotypes.Identifier{Package: "builtin", Def: "float32"}),

			makeVirtual(1, &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeVirtual(2, &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeVirtual(3, &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeVirtual(4, &gotypes.Slice{Elmtype: &gotypes.Identifier{Package: "builtin", Def: "string"}}),
			makeVirtual(5, &gotypes.Map{Keytype: &gotypes.Identifier{Package: "builtin", Def: "int"}, Valuetype: &gotypes.Identifier{Package: "builtin", Def: "string"}}),
			makeVirtual(6, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "1"}),
			makeVirtual(7, &gotypes.Pointer{Def: &gotypes.Identifier{Package: "builtin", Def: "int"}}),
			makeVirtual(8, &gotypes.Interface{}),
			makeVirtual(9, &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true, Literal: "0"}),
			makeVirtual(10, &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "a",
						Def:  &gotypes.Identifier{Package: "builtin", Def: "int"},
					},
				}}),
			makeVirtual(11, &gotypes.Pointer{Def: &gotypes.Identifier{Package: "builtin", Def: "int"}}),
			makeVirtual(12, &gotypes.Identifier{Package: "builtin", Def: "int"}),
			makeVirtual(13, &gotypes.Function{Package: gopkg}),
			makeVirtual(14, &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeVirtual(15, &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeVirtual(16, &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeVirtual(17, &gotypes.Identifier{Package: "builtin", Def: "string"}),
			makeVirtual(18, &gotypes.Builtin{Def: "bool", Untyped: true}),
			makeVirtual(19, &gotypes.Function{Package: gopkg}),
			makeVirtual(20, &gotypes.Builtin{Def: "bool", Untyped: true}),
		},
	)
}
