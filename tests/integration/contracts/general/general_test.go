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

var vars = map[string]string{
	"c":      ":48",
	"d":      ":82",
	"e":      ":121",
	"f":      ":136",
	"ff":     ":151",
	"j0":     ":201",
	"k0":     ":204",
	"l0":     ":207",
	"a":      ":218",
	"b":      ":221",
	"ta":     ":235",
	"tb":     ":239",
	"cv":     ":307",
	"l":      ":372",
	"v":      ":389",
	"ok":     ":392",
	"m":      ":405",
	"mv":     ":428",
	"mok":    ":432",
	"id":     ":446",
	"intD":   ":469",
	"intOk":  ":475",
	"s":      ":511",
	"swA":    ":585",
	"sd:c1":  ":641",
	"sd:c2":  ":653",
	"c1":     ":722",
	"msg1":   ":753",
	"msg1:2": ":792",
	"msg1Ok": ":798",
	"i":      ":873",
	"fI":     ":932",
	"fV":     ":944",
	"ii":     ":903",
	"vv":     ":907",
	"ifI":    ":1051",
}

func TestGeneralContracts(t *testing.T) {
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"../../testdata/general.go",
		[]contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y: typevars.MakeVar(packageName, "c", vars["c"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "3"}),
				Y: typevars.MakeVar(packageName, "d", vars["d"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int"}),
				Y: typevars.MakeVar(packageName, "d", vars["d"]),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(packageName, "e", vars["e"]),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(packageName, "d", vars["d"]),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(packageName, "f", vars["f"]),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(packageName, "e", vars["e"]),
				Y:       typevars.MakeVar(packageName, "f", vars["f"]),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeVar(packageName, "ff", vars["ff"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int"}),
				Y: typevars.MakeVar(packageName, "j", vars["j0"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int"}),
				Y: typevars.MakeVar(packageName, "k", vars["k0"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int"}),
				Y: typevars.MakeVar(packageName, "l", vars["l0"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y: typevars.MakeVar(packageName, "a", vars["a"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y: typevars.MakeVar(packageName, "b", vars["b"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y: typevars.MakeVar(packageName, "ta", vars["ta"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant("builtin", &gotypes.Identifier{Package: "builtin", Def: "float32"}),
				Y: typevars.MakeVar(packageName, "ta", vars["ta"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y: typevars.MakeVar(packageName, "tb", vars["tb"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant("builtin", &gotypes.Identifier{Package: "builtin", Def: "float32"}),
				Y: typevars.MakeVar(packageName, "tb", vars["tb"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Channel{Dir: "3", Value: &gotypes.Identifier{Package: "builtin", Def: "int"}}),
				Y: typevars.MakeLocalVar("cv", vars["cv"]),
			},
			&contracts.IsSendableTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y: typevars.MakeLocalVar("cv", vars["cv"]),
			},
			&contracts.IsIncDecable{
				X: typevars.MakeVar(packageName, "c", vars["c"]),
			},
			&contracts.IsIncDecable{
				X: typevars.MakeVar(packageName, "c", vars["c"]),
			},
			// l := []string{}
			// v, ok := l[1]
			&contracts.IsIndexable{
				X: typevars.MakeVirtualVar(4),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Slice{
					Elmtype: &gotypes.Identifier{Package: "builtin", Def: "string"},
				}),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Slice{
					Elmtype: &gotypes.Identifier{Package: "builtin", Def: "string"},
				}),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeLocalVar("l", vars["l"]),
			},
			&contracts.IsIndexable{
				X: typevars.MakeLocalVar("l", vars["l"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeListValue(typevars.MakeLocalVar("l", vars["l"])),
				Y: typevars.MakeLocalVar("v", vars["v"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}),
				Y: typevars.MakeLocalVar("ok", vars["ok"]),
			},
			// m := map[int]string{}
			// mv, mok := m[1]
			&contracts.IsIndexable{
				X: typevars.MakeVirtualVar(5),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Map{
					Keytype:   &gotypes.Identifier{Package: "builtin", Def: "int"},
					Valuetype: &gotypes.Identifier{Package: "builtin", Def: "string"},
				}),
				Y: typevars.MakeVirtualVar(5),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Map{
					Keytype:   &gotypes.Identifier{Package: "builtin", Def: "int"},
					Valuetype: &gotypes.Identifier{Package: "builtin", Def: "string"},
				}),
				Y: typevars.MakeVirtualVar(5),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeLocalVar("m", vars["m"]),
			},
			&contracts.IsIndexable{
				X: typevars.MakeLocalVar("m", vars["m"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y: typevars.MakeVirtualVar(6),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y: typevars.MakeMapKey(typevars.MakeVirtualVar(6)),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeMapValue(typevars.MakeLocalVar("m", vars["m"])),
				Y: typevars.MakeLocalVar("mv", vars["mv"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}),
				Y: typevars.MakeLocalVar("mok", vars["mok"]),
			},
			// id := interface{}(&c)
			// intD, intOk := id.(*int)
			&contracts.IsReferenceable{
				X: typevars.MakeVar(packageName, "c", vars["c"]),
			},
			&contracts.ReferenceOf{
				X: typevars.MakeVar(packageName, "c", vars["c"]),
				Y: typevars.MakeVirtualVar(7),
			},
			&contracts.TypecastsTo{
				X: typevars.MakeVirtualVar(7),
				Y: typevars.MakeVirtualVar(8),
				Type: typevars.MakeConstant(packageName, &gotypes.Interface{
					Methods: nil,
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(8),
				Y: typevars.MakeLocalVar("id", vars["id"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeLocalVar("id", vars["id"]),
				Y: typevars.MakeConstant(packageName, &gotypes.Pointer{
					Def: &gotypes.Identifier{Package: "builtin", Def: "int"},
				}),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Pointer{
					Def: &gotypes.Identifier{Package: "builtin", Def: "int"},
				}),
				Y: typevars.MakeLocalVar("intD", vars["intD"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}),
				Y: typevars.MakeLocalVar("intOk", vars["intOk"]),
			},
			// m[0] = "ahoj"
			&contracts.IsIndexable{
				X: typevars.MakeLocalVar("m", vars["m"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Y: typevars.MakeVirtualVar(9),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Y: typevars.MakeMapKey(typevars.MakeVirtualVar(9)),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "string", Literal: "\"ahoj\""}),
				Y: typevars.MakeMapValue(typevars.MakeLocalVar("m", vars["m"])),
			},
			// s := struct{ a int }{a: 2}
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(10),
				Field: "a",
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(10), "a", 0),
				Y: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "a",
							Def:  &gotypes.Identifier{Package: "builtin", Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(10),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Struct{
					Fields: []gotypes.StructFieldsItem{
						{
							Name: "a",
							Def:  &gotypes.Identifier{Package: "builtin", Def: "int"},
						},
					},
				}),
				Y: typevars.MakeVirtualVar(10),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(10),
				Y: typevars.MakeLocalVar("s", vars["s"]),
			},
			// s.a = 2
			&contracts.HasField{
				X:     typevars.MakeLocalVar("s", vars["s"]),
				Field: "a",
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y: typevars.MakeField(typevars.MakeLocalVar("s", vars["s"]), "a", 0),
			},
			// *(&c) = 2
			&contracts.IsReferenceable{
				X: typevars.MakeVar(packageName, "c", vars["c"]),
			},
			&contracts.ReferenceOf{
				X: typevars.MakeVar(packageName, "c", vars["c"]),
				Y: typevars.MakeVirtualVar(11),
			},
			&contracts.IsDereferenceable{
				X: typevars.MakeVirtualVar(11),
			},
			&contracts.DereferenceOf{
				X: typevars.MakeVirtualVar(11),
				Y: typevars.MakeVirtualVar(12),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Y: typevars.MakeVirtualVar(12),
			},
			// go func() {}()
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Function{
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(13),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(13),
				ArgsCount: 0,
			},
			// switch swA := 1; swA {
			// case a:
			// }
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y: typevars.MakeLocalVar("swA", vars["swA"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(packageName, "a", vars["a"]),
				Y: typevars.MakeLocalVar("swA", vars["swA"]),
			},
			// switch d := id.(type) {
			// case *int:
			// case *string, *float32:
			// }
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Pointer{
					Def: &gotypes.Identifier{Package: "builtin", Def: "int"},
				}),
				Y:    typevars.MakeLocalVar("id", vars["id"]),
				Weak: true,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Pointer{
					Def: &gotypes.Identifier{Package: "builtin", Def: "int"},
				}),
				Y: typevars.MakeLocalVar("sd", vars["sd:c1"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Pointer{
					Def: &gotypes.Identifier{Package: "builtin", Def: "string"},
				}),
				Y:    typevars.MakeLocalVar("id", vars["id"]),
				Weak: true,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Pointer{
					Def: &gotypes.Identifier{Package: "builtin", Def: "float32"},
				}),
				Y:    typevars.MakeLocalVar("id", vars["id"]),
				Weak: true,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Interface{}),
				Y: typevars.MakeLocalVar("sd", vars["sd:c2"]),
			},
			// switch id.(type) {
			// case *int:
			// }
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Pointer{
					Def: &gotypes.Identifier{Package: "builtin", Def: "int"},
				}),
				Y:    typevars.MakeLocalVar("id", vars["id"]),
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
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Channel{Dir: "3", Value: &gotypes.Identifier{Package: "builtin", Def: "string"}}),
				Y: typevars.MakeLocalVar("c1", vars["c1"]),
			},
			&contracts.IsReceiveableFrom{
				X: typevars.MakeLocalVar("c1", vars["c1"]),
				Y: typevars.MakeVirtualVar(14),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(14),
				Y: typevars.MakeLocalVar("msg1", vars["msg1"]),
			},
			&contracts.IsReceiveableFrom{
				X: typevars.MakeLocalVar("c1", vars["c1"]),
				Y: typevars.MakeVirtualVar(15),
			},
			&contracts.IsIndexable{
				X: typevars.MakeLocalVar("l", vars["l"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(15),
				Y: typevars.MakeListValue(typevars.MakeLocalVar("l", vars["l"])),
			},
			&contracts.IsReceiveableFrom{
				X: typevars.MakeLocalVar("c1", vars["c1"]),
				Y: typevars.MakeVirtualVar(16),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(16),
				Y: typevars.MakeLocalVar("msg1", vars["msg1:2"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}),
				Y: typevars.MakeLocalVar("msg1Ok", vars["msg1Ok"]),
			},
			&contracts.IsReceiveableFrom{
				X: typevars.MakeLocalVar("c1", vars["c1"]),
				Y: typevars.MakeVirtualVar(17),
			},
			&contracts.IsIndexable{
				X: typevars.MakeLocalVar("l", vars["l"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Y: typevars.MakeListKey(),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(17),
				Y: typevars.MakeListValue(typevars.MakeLocalVar("l", vars["l"])),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "bool"}),
				Y: typevars.MakeLocalVar("ok", vars["ok"]),
			},
			&contracts.IsSendableTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "string", Literal: "\"a\""}),
				Y: typevars.MakeLocalVar("c1", vars["c1"]),
			},
			// for i := 0; i < 1; i++ {
			// }
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "0"}),
				Y: typevars.MakeLocalVar("i", vars["i"]),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("i", vars["i"]),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Z:       typevars.MakeVirtualVar(18),
				OpToken: token.LSS,
			},
			&contracts.IsIncDecable{
				X: typevars.MakeLocalVar("i", vars["i"]),
			},
			// for ii, vv := range m {
			// }
			&contracts.IsRangeable{
				X: typevars.MakeLocalVar("m", vars["m"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeRangeKey(typevars.MakeLocalVar("m", vars["m"])),
				Y: typevars.MakeLocalVar("ii", vars["ii"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeRangeValue(typevars.MakeLocalVar("m", vars["m"])),
				Y: typevars.MakeLocalVar("vv", vars["vv"]),
			},
			// var fI int
			// var fV string
			// for fI, fV = range m {
			// }
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "int"}),
				Y: typevars.MakeLocalVar("fI", vars["fI"]),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Identifier{Package: "builtin", Def: "string"}),
				Y: typevars.MakeLocalVar("fV", vars["fV"]),
			},
			&contracts.IsRangeable{
				X: typevars.MakeLocalVar("m", vars["m"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeRangeKey(typevars.MakeLocalVar("m", vars["m"])),
				Y: typevars.MakeLocalVar("fI", vars["fI"]),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeRangeValue(typevars.MakeLocalVar("m", vars["m"])),
				Y: typevars.MakeLocalVar("fV", vars["fV"]),
			},
			// for _, _ = range m {
			// }
			&contracts.IsRangeable{
				X: typevars.MakeLocalVar("m", vars["m"]),
			},
			// for range m {
			// }
			&contracts.IsRangeable{
				X: typevars.MakeLocalVar("m", vars["m"]),
			},
			// defer func() {}()
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Function{
					Package: packageName,
				}),
				Y: typevars.MakeVirtualVar(19),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualVar(19),
				ArgsCount: 0,
			},
			// if ifI := 1; ifI < 2 {
			// } else {
			// }
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "1"}),
				Y: typevars.MakeLocalVar("ifI", vars["ifI"]),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeLocalVar("ifI", vars["ifI"]),
				Y:       typevars.MakeConstant(packageName, &gotypes.Constant{Package: "builtin", Untyped: true, Def: "int", Literal: "2"}),
				Z:       typevars.MakeVirtualVar(20),
				OpToken: token.LSS,
			},
		})

}
