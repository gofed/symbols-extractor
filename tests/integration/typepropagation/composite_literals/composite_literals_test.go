package composite_literals

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/composite_literals"

	var vars = map[string]string{
		"list":     ":45",
		"listV2":   ":275",
		"mapV":     ":75",
		"structV":  ":124",
		"structV2": ":205",
	}

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

	Slice := &gotypes.Slice{
		Elmtype: &gotypes.Identifier{Def: "Int", Package: gopkg},
	}
	Map := &gotypes.Map{
		Keytype:   &gotypes.Builtin{Def: "string"},
		Valuetype: &gotypes.Builtin{Def: "int"},
	}
	Struct := &gotypes.Struct{
		Fields: []gotypes.StructFieldsItem{
			{
				Name: "key1",
				Def:  &gotypes.Builtin{Def: "string"},
			},
			{
				Name: "key2",
				Def:  &gotypes.Builtin{Def: "int"},
			},
		},
	}

	intSlice := &gotypes.Slice{
		Elmtype: &gotypes.Builtin{Def: "int"},
	}
	SliceSlice := &gotypes.Slice{
		Elmtype: intSlice,
	}

	cutils.ParseAndCompareVarTable(
		t,
		gopkg,
		"../../testdata/composite_literals.go",
		[]cutils.VarTableTest{
			makeLocal("list", Slice),
			makeLocal("mapV", Map),
			makeLocal("structV", Struct),
			makeLocal("structV2", Struct),
			makeLocal("listV2", SliceSlice),
			makeVirtual(1, Slice),
			makeVirtual(2, Map),
			makeVirtual(3, Struct),
			makeVirtual(4, Struct),
			makeVirtual(5, SliceSlice),
			makeVirtual(6, intSlice),
			makeVirtual(7, intSlice),
		},
	)
}
