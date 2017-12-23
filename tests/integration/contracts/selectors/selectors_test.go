package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/selectors"

func TestSelectorsTypes(t *testing.T) {
	var vars = map[string]string{
		"frA": ":110",
		"mA":  ":153",
		"mB":  ":164",
		"ia":  ":301",
		"ib":  ":316",
		"ida": ":406",
		"idb": ":419",
		"idc": ":442",
		"idd": ":455",
		"ide": ":540",
		"idf": ":568",
		"idg": ":656",
		"idh": ":677",
		"idi": ":700",
		"idj": ":727",
		"idk": ":742",
		"idl": ":790",
	}
	utils.ParseAndCompareContracts(t,
		packageName,
		"testdata/selectors.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y: typevars.MakeVirtualVar(1),
			},
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(1),
				Field: "method",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVirtualVar(1), "method", 0),
				Y: typevars.MakeLocalVar("frA", vars["frA"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeLocalVar("mA", vars["mA"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("mA", vars["mA"]),
				Field: "method",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("mA", vars["mA"]), "method", 0),
				Y: typevars.MakeVirtualVar(3),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(3),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(3), 0),
				Y: typevars.MakeLocalVar("mB", vars["mB"]),
			},
			// var ia D2 = &D3{}
			// ib := ia.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D3",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D3",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.UnaryOp{
				X:       typevars.MakeVirtualVar(4),
				Y:       typevars.MakeVirtualVar(5),
				OpToken: token.AND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeLocalVar("ia", vars["ia"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("ia", vars["ia"]),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("ia", vars["ia"]), "imethod", 0),
				Y: typevars.MakeVirtualVar(6),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(6),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(6), 0),
				Y: typevars.MakeLocalVar("ib", vars["ib"]),
			},
			// type D4 D3
			// func (d *D4) imethod() int { return 0 }
			// ida := D4{}
			// idb := ida.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D4",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(7),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D4",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(7),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(7),
				Y: typevars.MakeLocalVar("ida", vars["ida"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("ida", vars["ida"]),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("ida", vars["ida"]), "imethod", 0),
				Y: typevars.MakeVirtualVar(8),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(8),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(8), 0),
				Y: typevars.MakeLocalVar("idb", vars["idb"]),
			},
			// idc := &ida
			// idd := idc.imethod()
			&contracts.UnaryOp{
				X:       typevars.MakeLocalVar("ida", vars["ida"]),
				Y:       typevars.MakeVirtualVar(9),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(9),
				Y: typevars.MakeLocalVar("idc", vars["idc"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("idc", vars["idc"]),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("idc", vars["idc"]), "imethod", 0),
				Y: typevars.MakeVirtualVar(10),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(10),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(10), 0),
				Y: typevars.MakeLocalVar("idd", vars["idd"]),
			},
			// ide := &struct{ d int }{2}
			// idf := ide.d
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(11),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(11), "", 0),
				Y: typevars.MakeConstant(packageName,
					&gotypes.Builtin{Untyped: true, Def: "int"},
				),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "d",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(11),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "d",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(11),
			},
			&contracts.UnaryOp{
				X:       typevars.MakeVirtualVar(11),
				Y:       typevars.MakeVirtualVar(12),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(12),
				Y: typevars.MakeLocalVar("ide", vars["ide"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("ide", vars["ide"]),
				Field: "d",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("ide", vars["ide"]), "d", 0),
				Y: typevars.MakeLocalVar("idf", vars["idf"]),
			},
			// type D6 string
			// func (d *D6) imethod() int { return 0 }
			// idg := D6("string")
			// idh := idg.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D6",
					Package: packageName,
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D6",
					Package: packageName,
				}),
				Y: typevars.MakeLocalVar("idg", vars["idg"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("idg", vars["idg"]),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("idg", vars["idg"]), "imethod", 0),
				Y: typevars.MakeVirtualVar(13),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(13),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(13), 0),
				Y: typevars.MakeLocalVar("idh", vars["idh"]),
			},
			// idi := struct{ d int }{2}
			// idj := idi.d
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(14),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(14), "", 0),
				Y: typevars.MakeConstant(packageName,
					&gotypes.Builtin{Untyped: true, Def: "int"},
				),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "d",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(14),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "d",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(14),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(14),
				Y: typevars.MakeLocalVar("idi", vars["idi"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("idi", vars["idi"]),
				Field: "d",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("idi", vars["idi"]), "d", 0),
				Y: typevars.MakeLocalVar("idj", vars["idj"]),
			},
			// idk := (interface {
			// 	imethod() int
			// })(&D3{})
			// idl := idk.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D3",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(15),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{
					Def:     "D3",
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(15),
			},
			&contracts.UnaryOp{
				X:       typevars.MakeVirtualVar(15),
				Y:       typevars.MakeVirtualVar(16),
				OpToken: token.AND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(16),
				Y: typevars.MakeConstant(packageName, &gotypes.Interface{
					Methods: []gotypes.InterfaceMethodsItem{
						gotypes.InterfaceMethodsItem{
							Name: "imethod",
							Def: &gotypes.Function{
								Package: packageName,
								Results: []gotypes.DataType{
									&gotypes.Builtin{Untyped: false, Def: "int"},
								},
							},
						},
					},
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Interface{
					Methods: []gotypes.InterfaceMethodsItem{
						gotypes.InterfaceMethodsItem{
							Name: "imethod",
							Def: &gotypes.Function{
								Package: packageName,
								Results: []gotypes.DataType{
									&gotypes.Builtin{Untyped: false, Def: "int"},
								},
							},
						},
					},
				}),
				Y: typevars.MakeLocalVar("idk", vars["idk"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("idk", vars["idk"]),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("idk", vars["idk"]), "imethod", 0),
				Y: typevars.MakeVirtualVar(17),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(17),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(17), 0),
				Y: typevars.MakeLocalVar("idl", vars["idl"]),
			},
		})
}
