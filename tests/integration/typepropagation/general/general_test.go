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
			makeLocal("c1", &gotypes.Channel{Dir: "3", Value: &gotypes.Builtin{Def: "string"}}),
			makeLocal("cv", &gotypes.Channel{Dir: "3", Value: &gotypes.Builtin{Def: "int"}}),
			makeLocal("fI", &gotypes.Builtin{Def: "int"}),
			makeLocal("fV", &gotypes.Builtin{Def: "string"}),
			makeLocal("i", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("id", &gotypes.Interface{}),
			makeLocal("ifI", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("ii", &gotypes.Builtin{Def: "int"}),
			makeLocal("intD", &gotypes.Pointer{Def: &gotypes.Builtin{Def: "int"}}),
			makeLocal("intOk", &gotypes.Builtin{Def: "bool"}),
			makeLocal("l", &gotypes.Slice{Elmtype: &gotypes.Builtin{Def: "string"}}),
			makeLocal("m", &gotypes.Map{Keytype: &gotypes.Builtin{Def: "int"}, Valuetype: &gotypes.Builtin{Def: "string"}}),
			makeLocal("mok", &gotypes.Builtin{Def: "bool"}),
			makeLocal("msg1_c1", &gotypes.Builtin{Def: "string"}),
			makeLocal("msg1_c3", &gotypes.Builtin{Def: "string"}),
			makeLocal("msg1Ok", &gotypes.Builtin{Def: "bool"}),
			makeLocal("mv", &gotypes.Builtin{Def: "string"}),
			makeLocal("ok", &gotypes.Builtin{Def: "bool"}),
			makeLocal("s", &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "a",
						Def:  &gotypes.Builtin{Def: "int"},
					},
				}}),
			makeLocal("sd_c1", &gotypes.Pointer{Def: &gotypes.Builtin{Def: "int"}}),
			makeLocal("sd_c2", &gotypes.Interface{}),
			makeLocal("swA", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("v", &gotypes.Builtin{Def: "string"}),
			makeLocal("vv", &gotypes.Builtin{Def: "string"}),

			makeLocal("c", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("d", &gotypes.Builtin{Def: "int"}),
			makeLocal("e", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("f", &gotypes.Builtin{Def: "int"}),
			makeLocal("ff", &gotypes.Builtin{Def: "int"}),
			makeLocal("j", &gotypes.Builtin{Def: "int"}),
			makeLocal("k", &gotypes.Builtin{Def: "int"}),
			makeLocal("g_l", &gotypes.Builtin{Def: "int"}),
			makeLocal("a", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("b", &gotypes.Builtin{Def: "int", Untyped: true}),
			makeLocal("ta", &gotypes.Builtin{Def: "float32"}),
			makeLocal("tb", &gotypes.Builtin{Def: "float32"}),

			makeVirtual(1, &gotypes.Builtin{Def: "int", Untyped: true}),
			makeVirtual(2, &gotypes.Builtin{Def: "int"}),
			makeVirtual(3, &gotypes.Builtin{Def: "int"}),
			makeVirtual(4, &gotypes.Slice{Elmtype: &gotypes.Builtin{Def: "string"}}),
			makeVirtual(5, &gotypes.Map{Keytype: &gotypes.Builtin{Def: "int"}, Valuetype: &gotypes.Builtin{Def: "string"}}),
			makeVirtual(6, &gotypes.Builtin{Def: "int", Untyped: true}),
			makeVirtual(7, &gotypes.Pointer{Def: &gotypes.Builtin{Def: "int", Untyped: true}}),
			makeVirtual(8, &gotypes.Builtin{Def: "int", Untyped: true}),
			makeVirtual(9, &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "a",
						Def:  &gotypes.Builtin{Def: "int"},
					},
				}}),
			makeVirtual(10, &gotypes.Pointer{Def: &gotypes.Builtin{Def: "int", Untyped: true}}),
			makeVirtual(11, &gotypes.Builtin{Def: "int", Untyped: true}),
			makeVirtual(12, &gotypes.Function{Package: gopkg}),
			makeVirtual(13, &gotypes.Builtin{Def: "string"}),
			makeVirtual(14, &gotypes.Builtin{Def: "string"}),
			makeVirtual(15, &gotypes.Builtin{Def: "string"}),
			makeVirtual(16, &gotypes.Builtin{Def: "string"}),
			makeVirtual(17, &gotypes.Builtin{Def: "bool"}),
			makeVirtual(18, &gotypes.Function{Package: gopkg}),
			makeVirtual(19, &gotypes.Builtin{Def: "bool"}),
		},
	)
}
