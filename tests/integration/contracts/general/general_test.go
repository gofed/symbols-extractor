package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/general"

var gVars = map[string]string{
	"c":      ":48:c",
	"d":      ":82:d",
	"e":      ":121:e",
	"f":      ":136:f",
	"ff":     ":151:ff",
	"a":      ":218:a",
	"b":      ":221:b",
	"ta":     ":235:ta",
	"tb":     ":239:tb",
	"cv":     ":307:cv",
	"l":      ":372:l",
	"v":      ":389:v",
	"ok":     ":392:ok",
	"m":      ":405:m",
	"mv":     ":428:mv",
	"mok":    ":432:mok",
	"id":     ":446:id",
	"intD":   ":469:intD",
	"intOk":  ":475:intOk",
	"s":      ":511:s",
	"swA":    ":585:swA",
	"sd:c1":  ":641:sd",
	"sd:c2":  ":653:sd",
	"c1":     ":722:c1",
	"msg1":   ":753:msg1",
	"msg1:2": ":792:msg1",
	"msg1Ok": ":798:msg1Ok",
	"i":      ":873:i",
	"fI":     ":932:fI",
	"fV":     ":944:fV",
	"ii":     ":903:ii",
	"vv":     ":907:vv",
	"ifI":    ":1051:ifI",
}

func TestGeneralContracts(t *testing.T) {
	utils.CompareContracts(
		t,
		packageName,
		"testdata/general.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["c"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["d"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(gVars["e"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(gVars["d"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(gVars["f"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(gVars["e"], packageName),
				Y:       typevars.MakeVar(gVars["f"], packageName),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeVar(gVars["ff"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["a"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["b"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["ta"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["tb"], packageName),
			},
			&contracts.IsSendableTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["cv"], packageName),
			},
			&contracts.IsIncDecable{
				X: typevars.MakeVar(gVars["c"], packageName),
			},
			&contracts.IsIncDecable{
				X: typevars.MakeVar(gVars["c"], packageName),
			},
			// l := []string{}
			// v, ok := l[1]
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Builtin{Untyped: false, Def: "string"},
				}),
				Y: typevars.MakeVar(gVars["l"], packageName),
			},
			&contracts.IsIndexable{
				X: typevars.MakeVar(gVars["l"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeListValue(typevars.MakeVar(gVars["l"], packageName)),
				Y: typevars.MakeVar(gVars["v"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y: typevars.MakeVar(gVars["ok"], packageName),
			},
			// m := map[int]string{}
			// mv, mok := m[1]
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Map{
					Keytype:   &gotypes.Builtin{Untyped: false, Def: "int"},
					Valuetype: &gotypes.Builtin{Untyped: false, Def: "string"},
				}),
				Y: typevars.MakeVar(gVars["m"], packageName),
			},
			&contracts.IsIndexable{
				X: typevars.MakeVar(gVars["m"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeMapKey(typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"})),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeMapValue(typevars.MakeVar(gVars["m"], packageName)),
				Y: typevars.MakeVar(gVars["mv"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y: typevars.MakeVar(gVars["mok"], packageName),
			},
			// id := interface{}(&c)
			// intD, intOk := id.(*int)
			&contracts.UnaryOp{
				X:       typevars.MakeVar(gVars["c"], packageName),
				Y:       typevars.MakeVirtualVar(4),
				OpToken: token.AND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeConstant(&gotypes.Interface{
					Methods: nil,
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Interface{
					Methods: nil,
				}),
				Y: typevars.MakeVar(gVars["id"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(gVars["id"], packageName),
				Y: typevars.MakeConstant(&gotypes.Pointer{
					Def: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Pointer{
					Def: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeVar(gVars["intD"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y: typevars.MakeVar(gVars["intOk"], packageName),
			},
			// m[0] = "ahoj"
			&contracts.IsIndexable{
				X: typevars.MakeVar(gVars["m"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeConstantMapKey(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeMapValue(typevars.MakeVar(gVars["m"], packageName)),
			},
			// s := struct{ a int }{a: 2}
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(5),
				Field: "a",
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(5), "a", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "a",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(5),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "a",
							Def:  &gotypes.Builtin{Untyped: false, Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVar(gVars["s"], packageName),
			},
			// s.a = 2
			&contracts.HasField{
				X:     typevars.MakeVar(gVars["s"], packageName),
				Field: "a",
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeField(typevars.MakeVar(gVars["s"], packageName), "a", 0),
			},
			// *(&c) = 2
			&contracts.UnaryOp{
				X:       typevars.MakeVar(gVars["c"], packageName),
				Y:       typevars.MakeVirtualVar(6),
				OpToken: token.AND,
			},
			&contracts.IsDereferenceable{
				X: typevars.MakeVirtualVar(6),
			},
			&contracts.DereferenceOf{
				X: typevars.MakeVirtualVar(6),
				Y: typevars.MakeVirtualVar(7),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVirtualVar(7),
			},
			// go func() {}()
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Function{
					Package: packageName,
				}),
				Y: typevars.MakeVirtualFunction(typevars.MakeVirtualVar(8)),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(8)),
				ArgsCount: 0,
			},
			// switch swA := 1; swA {
			// case a:
			// }
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["swA"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(gVars["a"], packageName),
				Y: typevars.MakeVar(gVars["swA"], packageName),
			},
			// switch d := id.(type) {
			// case *int:
			// case *string, *float32:
			// }
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Pointer{
					Def: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y:    typevars.MakeVar(gVars["id"], packageName),
				Weak: true,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Pointer{
					Def: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeVar(gVars["sd:c1"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Pointer{
					Def: &gotypes.Builtin{Untyped: false, Def: "string"},
				}),
				Y:    typevars.MakeVar(gVars["id"], packageName),
				Weak: true,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Pointer{
					Def: &gotypes.Builtin{Untyped: false, Def: "float32"},
				}),
				Y:    typevars.MakeVar(gVars["id"], packageName),
				Weak: true,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Interface{}),
				Y: typevars.MakeVar(gVars["sd:c2"], packageName),
			},
			// switch id.(type) {
			// case *int:
			// }
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Pointer{
					Def: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y:    typevars.MakeVar(gVars["id"], packageName),
				Weak: true,
			},
			// var c1 chan string
			// select {
			// case msg1 := <-c1:
			// case l[0] = <-c1:
			// case msg1, msg1Ok := <-c1:
			// case l[0], ok = <-c1:
			// case c1 <- "a":
			// default:
			// }
			&contracts.IsReceiveableFrom{
				X: typevars.MakeVar(gVars["c1"], packageName),
				Y: typevars.MakeVirtualVar(9),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(9),
				Y: typevars.MakeVar(gVars["msg1"], packageName),
			},
			&contracts.IsReceiveableFrom{
				X: typevars.MakeVar(gVars["c1"], packageName),
				Y: typevars.MakeVirtualVar(10),
			},
			&contracts.IsIndexable{
				X: typevars.MakeVar(gVars["l"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(10),
				Y: typevars.MakeListValue(typevars.MakeVar(gVars["l"], packageName)),
			},
			&contracts.IsReceiveableFrom{
				X: typevars.MakeVar(gVars["c1"], packageName),
				Y: typevars.MakeVirtualVar(11),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(11),
				Y: typevars.MakeVar(gVars["msg1:2"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Def: "bool"}),
				Y: typevars.MakeVar(gVars["msg1Ok"], packageName),
			},
			&contracts.IsReceiveableFrom{
				X: typevars.MakeVar(gVars["c1"], packageName),
				Y: typevars.MakeVirtualVar(12),
			},
			&contracts.IsIndexable{
				X: typevars.MakeVar(gVars["l"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(12),
				Y: typevars.MakeListValue(typevars.MakeVar(gVars["l"], packageName)),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Def: "bool"}),
				Y: typevars.MakeVar(gVars["ok"], packageName),
			},
			&contracts.IsSendableTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeVar(gVars["c1"], packageName),
			},
			// for i := 0; i < 1; i++ {
			// }
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["i"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(gVars["i"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(13),
				OpToken: token.LSS,
			},
			&contracts.IsIncDecable{
				X: typevars.MakeVar(gVars["i"], packageName),
			},
			// var fI int
			// var fV string
			// for fI, fV = range m {
			// }
			&contracts.IsRangeable{
				X: typevars.MakeVar(gVars["m"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeRangeKey(typevars.MakeVar(gVars["m"], packageName)),
				Y: typevars.MakeVar(gVars["ii"], packageName),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeRangeValue(typevars.MakeVar(gVars["m"], packageName)),
				Y: typevars.MakeVar(gVars["vv"], packageName),
			},
			&contracts.IsRangeable{
				X: typevars.MakeVar(gVars["m"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeRangeKey(typevars.MakeVar(gVars["m"], packageName)),
				Y: typevars.MakeVar(gVars["fI"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeRangeValue(typevars.MakeVar(gVars["m"], packageName)),
				Y: typevars.MakeVar(gVars["fV"], packageName),
			},
			// for _, _ = range m {
			// }
			&contracts.IsRangeable{
				X: typevars.MakeVar(gVars["m"], packageName),
			},
			// for range m {
			// }
			&contracts.IsRangeable{
				X: typevars.MakeVar(gVars["m"], packageName),
			},
			// defer func() {}()
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Function{
					Package: packageName,
				}),
				Y: typevars.MakeVirtualFunction(typevars.MakeVirtualVar(14)),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(14)),
				ArgsCount: 0,
			},
			// if ifI := 1; ifI < 2 {
			// } else {
			// }
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(gVars["ifI"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(gVars["ifI"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(15),
				OpToken: token.LSS,
			},
		})

}
