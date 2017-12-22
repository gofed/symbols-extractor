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
		"list": ":45",
		"mapV": ":63",
		"la":   ":107",
		"lb":   ":124",
		"ma":   ":137",
		"sa":   ":154",
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
			&contracts.IsIndexable{
				X: typevars.MakeVirtualVar(1),
			},
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
				Y: typevars.MakeLocalVar("list", vars["list"]),
			},
			//
			// mapV := map[string]int{
			// 	"3": 3,
			// }
			//
			&contracts.IsIndexable{
				X: typevars.MakeVirtualVar(2),
			},
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
				X: typevars.MakeConstant(&gotypes.Map{
					Keytype:   &gotypes.Builtin{Untyped: false, Def: "string"},
					Valuetype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeLocalVar("mapV", vars["mapV"]),
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
				X: typevars.MakeLocalVar("list", vars["list"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeLocalVar("list", vars["list"]),
				Y: typevars.MakeLocalVar("la", vars["la"]),
			},
			//lb := la[3]
			&contracts.IsIndexable{
				X: typevars.MakeLocalVar("la", vars["la"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeListValue(typevars.MakeLocalVar("la", vars["la"])),
				Y: typevars.MakeLocalVar("lb", vars["lb"]),
			},
			// ma := mapV["3"]
			&contracts.IsIndexable{
				X: typevars.MakeLocalVar("mapV", vars["mapV"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeConstantMapKey(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeMapValue(typevars.MakeLocalVar("mapV", vars["mapV"])),
				Y: typevars.MakeLocalVar("ma", vars["ma"]),
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
				Y: typevars.MakeLocalVar("sa", vars["sa"]),
			},
		})
}
