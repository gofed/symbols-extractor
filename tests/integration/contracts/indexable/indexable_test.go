package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/indexable"

func TestIndexableContracts(t *testing.T) {
	var vars = map[string]string{
		"list": ":45:list",
		"mapV": ":63:mapV",
		"la":   ":107:la",
		"lb":   ":124:lb",
		"ma":   ":137:ma",
		"sa":   ":154:sa",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"testdata/indexable.go",
		[]contracts.Contract{
			//
			// list := []Int{
			// 	1,
			// }
			//
			// Int <-> 1
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(1)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Identifier{
						Def:     "Int",
						Package: packageName,
					},
				}),
				Y: typevars.MakeVirtualVar(1),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(vars["list"], packageName),
			},
			//
			// mapV := map[string]int{
			// 	"3": 3,
			// }
			//
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapKey(typevars.MakeVirtualVar(2)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapValue(typevars.MakeVirtualVar(2)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Map{
					Keytype:   &gotypes.Builtin{Untyped: false, Def: "string"},
					Valuetype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(vars["mapV"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.IsIndexable{
				X: typevars.MakeVar(vars["list"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVar(vars["list"], packageName),
				Y: typevars.MakeVar(vars["la"], packageName),
			},
			//lb := la[3]
			&contracts.IsIndexable{
				X: typevars.MakeVar(vars["la"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeListValue(typevars.MakeVar(vars["la"], packageName)),
				Y: typevars.MakeVar(vars["lb"], packageName),
			},
			// ma := mapV["3"]
			&contracts.IsIndexable{
				X: typevars.MakeVar(vars["mapV"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeConstantMapKey(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeMapValue(typevars.MakeVar(vars["mapV"], packageName)),
				Y: typevars.MakeVar(vars["ma"], packageName),
			},
			//sa := "ahoj"[0]
			&contracts.IsIndexable{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstantListValue(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeVar(vars["sa"], packageName),
			},
		})
}
