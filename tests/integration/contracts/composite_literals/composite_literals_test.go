package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/testdata"

func TestCompositeLiteralsContracts(t *testing.T) {
	var vars = map[string]string{
		"list":     ":45:list",
		"mapV":     ":75:mapV",
		"structV":  ":124:structV",
		"structV2": ":205:structV2",
		"listV2":   ":275:listV2",
	}
	utils.CompareContracts(
		t,
		packageName,
		"composite_literals.go",
		[]contracts.Contract{
			//
			// list := []Int{
			// 	1,
			// 	2,
			// }
			//
			// Int <-> 1
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Identifier{
					Def:     "Int",
					Package: packageName,
				}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// Int <-> 2
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Identifier{
					Def:     "Int",
					Package: packageName,
				}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// list <-> []Int
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Identifier{
						Def:     "Int",
						Package: packageName,
					},
				}),
				Y: typevars.MakeVar(vars["list"], packageName),
			},
			//
			// mapV := map[string]int{
			// 	"3": 3,
			// 	"4": 4,
			// }
			//
			// "3" <-> string
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantMapKey(&gotypes.Builtin{Untyped: false, Def: "string"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// 3 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantMapValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// "4" <-> string
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantMapKey(&gotypes.Builtin{Untyped: false, Def: "string"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// 4 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantMapValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// mapV <-> map[string]int
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Map{
					Keytype:   &gotypes.Builtin{Untyped: false, Def: "string"},
					Valuetype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeVar(vars["mapV"], packageName),
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
				X:     typevars.MakeVirtualVar(1),
				Field: "key1",
			},
			// key1 <-> string
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(1), "key1", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// key2 exists
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(1),
				Field: "key2",
			},
			// key2 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(1), "key2", 0),
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
				Y: typevars.MakeVirtualVar(1),
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
				Y: typevars.MakeVar(vars["structV"], packageName),
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
				X:     typevars.MakeVirtualVar(2),
				Index: 0,
			},
			// key at pos 0 <-> string
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(2), "", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// key at pos 1 used
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(2),
				Index: 1,
			},
			// key at pos 1 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(2), "", 1),
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
				Y: typevars.MakeVirtualVar(2),
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
				Y: typevars.MakeVar(vars["structV2"], packageName),
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
			// 1 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// 2 <-> int
			// TODO(jchaloup): documents this as "constant contract" a.k.a does not consume any data type
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// []int <-> {1,2}
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Slice{
					Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
			},
			// 3 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// 4 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// []int <-> {3,4}
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstantListValue(&gotypes.Slice{
					Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
			},
			// [][]int <-> {{1,2},{3,4}}
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Slice{
						Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
					},
				}),
				Y: typevars.MakeVar(vars["listV2"], packageName),
			},
		})
}
