package function_invocation

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/function_invocation"

	var vars = map[string]string{
		"g":    typevars.MakeVar(gopkg, "g", ":20").String(),
		"g_a":  typevars.MakeLocalVar("a", ":27").String(),
		"g_b":  typevars.MakeLocalVar("b", ":30").String(),
		"g_c":  typevars.MakeLocalVar("c", ":33").String(),
		"g1":   typevars.MakeVar(gopkg, "g1", ":69").String(),
		"g1_a": typevars.MakeLocalVar("a", ":77").String(),
		"g2":   typevars.MakeVar(gopkg, "g2", ":113").String(),
		"g2_a": typevars.MakeLocalVar("a", ":121").String(),
		"g2_b": typevars.MakeLocalVar("b", ":131").String(),
		"a":    typevars.MakeLocalVar("a", ":179").String(),
		"aa":   typevars.MakeLocalVar("aa", ":287").String(),
		"ab":   typevars.MakeLocalVar("ab", ":304").String(),
		"b":    typevars.MakeLocalVar("b", ":321").String(),
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

	str := &gotypes.Builtin{Def: "string"}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/function_invocation.go",
		[]cutils.VarTableTest{
			makeLocal("g", &gotypes.Function{
				Package: gopkg,
				Params: []gotypes.DataType{
					str,
					str,
					str,
				},
				Results: []gotypes.DataType{
					str,
				},
			}),
			makeLocal("g_a", str),
			makeLocal("g_b", str),
			makeLocal("g_c", str),
			makeLocal("g1", &gotypes.Function{
				Package: gopkg,
				Params: []gotypes.DataType{
					str,
				},
				Results: []gotypes.DataType{
					str,
				},
			}),
			makeLocal("g1_a", str),
			makeLocal("g2", &gotypes.Function{
				Package: gopkg,
				Params: []gotypes.DataType{
					str,
					&gotypes.Ellipsis{Def: &gotypes.Builtin{Def: "int"}},
				},
				Results: []gotypes.DataType{
					str,
				},
			}),
			makeLocal("g2_a", str),
			makeLocal("g2_b", &gotypes.Ellipsis{Def: &gotypes.Builtin{Def: "int"}}),
			makeLocal("a", &gotypes.Builtin{Def: "string", Untyped: true}),
			makeLocal("aa", str),
			makeLocal("ab", str),
			makeLocal("b", str),
			makeVirtual(1, str),
			makeVirtual(2, str),
			makeVirtual(3, str),
		},
	)
}
