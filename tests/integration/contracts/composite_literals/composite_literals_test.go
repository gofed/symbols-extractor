package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/composite_literals"

func TestCompositeLiteralsContracts(t *testing.T) {
	var vars = map[string]string{
		"list":     ":45",
		"mapV":     ":75",
		"structV":  ":124",
		"structV2": ":205",
		"listV2":   ":275",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"testdata/composite_literals.go",
		[]contracts.Contract{
			//
			// list := []Int{
			// 	1,
			// 	2,
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
			// mapV := map[string]int{
			// 	"3": 3,
			// 	"4": 4,
			// }
			//
			// "3" <-> string
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
			//
			// structV := struct {
			// 	key1 string
			// 	key2 int
			// }{
			// 	key1: "key1",
			// 	key2: 2,
			// }
			//
			// key1 exists
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(3),
				Field: "key1",
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(3), "key1", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(3),
				Field: "key2",
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(3), "key2", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// structV <-> struct {
			// 	key1 string
			// 	key2 int
			// }
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						gotypes.StructFieldsItem{
							Name: "key1",
							Def:  &gotypes.Builtin{Untyped: false, Def: "string"},
						},
						gotypes.StructFieldsItem{
							Name: "key2",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(3),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						gotypes.StructFieldsItem{
							Name: "key1",
							Def:  &gotypes.Builtin{Untyped: false, Def: "string"},
						},
						gotypes.StructFieldsItem{
							Name: "key2",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(3),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeLocalVar("structV", vars["structV"]),
			},
			//
			// structV2 := struct {
			// 	key1 string
			// 	key2 int
			// }{
			// 	"key1",
			// 	2,
			// }
			//
			// key at pos 0 used
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(4),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(4), "", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(4),
				Index: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(4), "", 1),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// structV2 <-> struct {
			// 	key1 string
			// 	key2 int
			// }
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						gotypes.StructFieldsItem{
							Name: "key1",
							Def:  &gotypes.Builtin{Untyped: false, Def: "string"},
						},
						gotypes.StructFieldsItem{
							Name: "key2",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						gotypes.StructFieldsItem{
							Name: "key1",
							Def:  &gotypes.Builtin{Untyped: false, Def: "string"},
						},
						gotypes.StructFieldsItem{
							Name: "key2",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeLocalVar("structV2", vars["structV2"]),
			},
			//
			// listV2 := [][]int{
			// 	{
			// 		1,
			// 		2,
			// 	},
			// 	{
			// 		3,
			// 		4,
			// 	},
			// }
			//
			&contracts.IsIndexable{
				X: typevars.MakeVirtualVar(5),
			},
			&contracts.IsIndexable{
				X: typevars.MakeVirtualVar(6),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(6)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(6)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(5)),
				Y: typevars.MakeVirtualVar(6),
			},
			&contracts.IsIndexable{
				X: typevars.MakeVirtualVar(7),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(7)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(7)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(5)),
				Y: typevars.MakeVirtualVar(7),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Slice{
						Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
					},
				}),
				Y: typevars.MakeVirtualVar(5),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Slice{
						Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
					},
				}),
				Y: typevars.MakeVirtualVar(5),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeLocalVar("listV2", vars["listV2"]),
			},
		})
}
