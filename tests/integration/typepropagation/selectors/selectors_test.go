package selectors

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/selectors"

	var vars = map[string]string{
		"dD":  typevars.MakeLocalVar("d", ":42").String(),
		"dD3": typevars.MakeLocalVar("d", ":247").String(),
		"dD4": typevars.MakeLocalVar("d", ":356").String(),
		"dD6": typevars.MakeLocalVar("d", ":606").String(),
		"frA": typevars.MakeLocalVar("frA", ":110").String(),
		"ia":  typevars.MakeLocalVar("ia", ":301").String(),
		"ib":  typevars.MakeLocalVar("ib", ":316").String(),
		"ida": typevars.MakeLocalVar("ida", ":406").String(),
		"idb": typevars.MakeLocalVar("idb", ":419").String(),
		"idc": typevars.MakeLocalVar("idc", ":442").String(),
		"idd": typevars.MakeLocalVar("idd", ":455").String(),
		"ide": typevars.MakeLocalVar("ide", ":540").String(),
		"idf": typevars.MakeLocalVar("idf", ":568").String(),
		"idg": typevars.MakeLocalVar("idg", ":656").String(),
		"idh": typevars.MakeLocalVar("idh", ":677").String(),
		"idi": typevars.MakeLocalVar("idi", ":700").String(),
		"idj": typevars.MakeLocalVar("idj", ":727").String(),
		"idk": typevars.MakeLocalVar("idk", ":742").String(),
		"idl": typevars.MakeLocalVar("idl", ":790").String(),
		"mA":  typevars.MakeLocalVar("mA", ":153").String(),
		"mB":  typevars.MakeLocalVar("mB", ":164").String(),
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

	DPointer := &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D", Package: gopkg}}
	Method := &gotypes.Method{
		Receiver: DPointer,
		Def: &gotypes.Function{
			Package: gopkg,
			Results: []gotypes.DataType{
				&gotypes.Builtin{Def: "int"},
			},
		}}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/selectors.go",
		[]cutils.VarTableTest{
			makeLocal("dD", DPointer),
			makeLocal("dD3", &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D3", Package: gopkg}}),
			makeLocal("dD4", &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D4", Package: gopkg}}),
			makeLocal("dD6", &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D6", Package: gopkg}}),
			makeLocal("frA", Method),
			makeLocal("ia", &gotypes.Identifier{Def: "D2", Package: gopkg}),
			makeLocal("ib", &gotypes.Builtin{Def: "int"}),
			makeLocal("ida", &gotypes.Identifier{Def: "D4", Package: gopkg}),
			makeLocal("idb", &gotypes.Builtin{Def: "int"}),
			makeLocal("idc", &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D4", Package: gopkg}}),
			makeLocal("idd", &gotypes.Builtin{Def: "int"}),
			makeLocal("ide", &gotypes.Pointer{
				Def: &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "d",
							Def:  &gotypes.Builtin{Def: "int"},
						},
					},
				}}),
			makeLocal("idf", &gotypes.Builtin{Def: "int"}),
			makeLocal("idg", &gotypes.Identifier{Def: "D6", Package: gopkg}),
			makeLocal("idh", &gotypes.Builtin{Def: "int"}),
			makeLocal("idi", &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "d",
						Def:  &gotypes.Builtin{Def: "int"},
					},
				}}),
			makeLocal("idj", &gotypes.Builtin{Def: "int"}),
			makeLocal("idk", &gotypes.Interface{
				Methods: []gotypes.InterfaceMethodsItem{
					{
						Name: "imethod",
						Def: &gotypes.Function{
							Package: gopkg,
							Results: []gotypes.DataType{
								&gotypes.Builtin{Def: "int"},
							},
						},
					},
				}}),
			makeLocal("idl", &gotypes.Builtin{Def: "int"}),
			makeLocal("mA", &gotypes.Identifier{Def: "D", Package: gopkg}),
			makeLocal("mB", &gotypes.Builtin{Def: "int"}),
			makeVirtual(1, &gotypes.Identifier{Def: "D", Package: gopkg}),
			makeVirtual(2, &gotypes.Identifier{Def: "D", Package: gopkg}),
			makeVirtual(3, Method),
			makeVirtual(4, &gotypes.Identifier{Def: "D3", Package: gopkg}),
			makeVirtual(5, &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D3", Package: gopkg}}),
			makeVirtual(6, &gotypes.Function{
				Package: gopkg,
				Results: []gotypes.DataType{
					&gotypes.Builtin{Def: "int"},
				}}),
			makeVirtual(7, &gotypes.Identifier{Def: "D4", Package: gopkg}),
			makeVirtual(8, &gotypes.Method{
				Receiver: &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D4", Package: gopkg}},
				Def: &gotypes.Function{
					Package: gopkg,
					Results: []gotypes.DataType{
						&gotypes.Builtin{Def: "int"},
					},
				}}),
			makeVirtual(9, &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D4", Package: gopkg}}),
			makeVirtual(10, &gotypes.Method{
				Receiver: &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D4", Package: gopkg}},
				Def: &gotypes.Function{
					Package: gopkg,
					Results: []gotypes.DataType{
						&gotypes.Builtin{Def: "int"},
					},
				}}),
			makeVirtual(11, &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "d",
						Def:  &gotypes.Builtin{Def: "int"},
					},
				}}),
			makeVirtual(12, &gotypes.Pointer{Def: &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "d",
						Def:  &gotypes.Builtin{Def: "int"},
					},
				},
			}}),
			makeVirtual(13, &gotypes.Method{
				Receiver: &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D6", Package: gopkg}},
				Def: &gotypes.Function{
					Package: gopkg,
					Results: []gotypes.DataType{
						&gotypes.Builtin{Def: "int"},
					},
				}}),
			makeVirtual(14, &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "d",
						Def:  &gotypes.Builtin{Def: "int"},
					},
				}}),
			makeVirtual(15, &gotypes.Identifier{Def: "D3", Package: gopkg}),
			makeVirtual(16, &gotypes.Pointer{Def: &gotypes.Identifier{Def: "D3", Package: gopkg}}),
			makeVirtual(17, &gotypes.Function{
				Package: gopkg,
				Results: []gotypes.DataType{
					&gotypes.Builtin{Def: "int"},
				}}),
		},
	)
}
