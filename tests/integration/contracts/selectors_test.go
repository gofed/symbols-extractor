package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func TestSelectorsTypes(t *testing.T) {
	var vars = map[string]string{
		"frA": ":124:frA",
		"mA":  ":167:mA",
		"mB":  ":178:mB",
		"ia":  ":335:ia",
		"ib":  ":350:ib",
		"ida": ":440:ida",
		"idb": ":453:idb",
		"idc": ":476:idc",
		"idd": ":489:idd",
		"ide": ":574:ide",
		"idf": ":602:idf",
		"idg": ":690:idg",
		"idh": ":711:idh",
		"idi": ":734:idi",
		"idj": ":761:idj",
		"idk": ":776:idk",
		"idl": ":824:idl",
	}
	compareContracts(t,
		packageName,
		"selectors.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Struct{
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
				Y: typevars.MakeVar(vars["frA"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y: typevars.MakeVar(vars["mA"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["mA"], packageName),
				Field: "method",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["mA"], packageName), "method", 0),
				Y: typevars.MakeVirtualVar(3),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(3)),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(3).Name, packageName, 0),
				Y: typevars.MakeVar(vars["mB"], packageName),
			},
			// var ia D2 = &D3{}
			// ib := ia.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.UnaryOp{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y:       typevars.MakeVirtualVar(5),
				OpToken: token.AND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeVar(vars["ia"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["ia"], packageName),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["ia"], packageName), "imethod", 0),
				Y: typevars.MakeVirtualVar(6),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(6)),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(6).Name, packageName, 0),
				Y: typevars.MakeVar(vars["ib"], packageName),
			},
			// type D4 D3
			// func (d *D4) imethod() int { return 0 }
			// ida := D4{}
			// idb := ida.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y: typevars.MakeVirtualVar(7),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y: typevars.MakeVar(vars["ida"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["ida"], packageName),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["ida"], packageName), "imethod", 0),
				Y: typevars.MakeVirtualVar(8),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(8)),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(8).Name, packageName, 0),
				Y: typevars.MakeVar(vars["idb"], packageName),
			},
			// idc := &ida
			// idd := idc.imethod()
			&contracts.UnaryOp{
				X:       typevars.MakeVar(vars["ida"], packageName),
				Y:       typevars.MakeVirtualVar(9),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(9),
				Y: typevars.MakeVar(vars["idc"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["idc"], packageName),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["idc"], packageName), "imethod", 0),
				Y: typevars.MakeVirtualVar(10),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(10)),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(10).Name, packageName, 0),
				Y: typevars.MakeVar(vars["idd"], packageName),
			},
			// ide := &struct{ d int }{2}
			// idf := ide.d
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(11),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVar(typevars.MakeVirtualVar(11).Name, packageName), "", 0),
				Y: typevars.MakeConstant(
					&gotypes.Builtin{Untyped: true, Def: "int"},
				),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
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
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "d",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y:       typevars.MakeVirtualVar(12),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(12),
				Y: typevars.MakeVar(vars["ide"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["ide"], packageName),
				Field: "d",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["ide"], packageName), "d", 0),
				Y: typevars.MakeVar(vars["idf"], packageName),
			},
			// type D6 string
			// func (d *D6) imethod() int { return 0 }
			// idg := D6("string")
			// idh := idg.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeConstant(&gotypes.Identifier{
					Def:     "D6",
					Package: packageName,
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Identifier{
					Def:     "D6",
					Package: packageName,
				}),
				Y: typevars.MakeVar(vars["idg"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["idg"], packageName),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["idg"], packageName), "imethod", 0),
				Y: typevars.MakeVirtualVar(13),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(13)),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(13).Name, packageName, 0),
				Y: typevars.MakeVar(vars["idh"], packageName),
			},
			// idi := struct{ d int }{2}
			// idj := idi.d
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(14),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVar(typevars.MakeVirtualVar(14).Name, packageName), "", 0),
				Y: typevars.MakeConstant(
					&gotypes.Builtin{Untyped: true, Def: "int"},
				),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
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
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "d",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVar(vars["idi"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["idi"], packageName),
				Field: "d",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["idi"], packageName), "d", 0),
				Y: typevars.MakeVar(vars["idj"], packageName),
			},
			// idk := (interface {
			// 	imethod() int
			// })(&D3{})
			// idl := idk.imethod()
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y: typevars.MakeVirtualVar(15),
			},
			&contracts.UnaryOp{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{},
				}),
				Y:       typevars.MakeVirtualVar(16),
				OpToken: token.AND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(16),
				Y: typevars.MakeConstant(&gotypes.Interface{
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
				X: typevars.MakeConstant(&gotypes.Interface{
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
				Y: typevars.MakeVar(vars["idk"], packageName),
			},
			&contracts.HasField{
				X:     typevars.MakeVar(vars["idk"], packageName),
				Field: "imethod",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeVar(vars["idk"], packageName), "imethod", 0),
				Y: typevars.MakeVirtualVar(17),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(17)),
				ArgsCount: 0,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeReturn(typevars.MakeVirtualVar(17).Name, packageName, 0),
				Y: typevars.MakeVar(vars["idl"], packageName),
			},
		})
}
