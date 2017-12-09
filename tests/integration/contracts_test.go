package file

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

func compareTypeVars(t *testing.T, expected, tested typevars.Interface) {
	if expected.GetType() != tested.GetType() {
		t.Errorf("Expected %v, got %v instead", expected.GetType(), tested.GetType())
	}
	switch x := expected.(type) {
	case *typevars.Constant:
		y := tested.(*typevars.Constant)
		if !reflect.DeepEqual(x.DataType, y.DataType) {
			xByteSlice, _ := json.Marshal(x.DataType)
			yByteSlice, _ := json.Marshal(y.DataType)
			t.Errorf("Got Constant.DataType:\n\t%v\nexpected:\n\t%v", string(xByteSlice), string(yByteSlice))
		}
	case *typevars.Variable:
		y := tested.(*typevars.Variable)
		if x.Name != y.Name {
			t.Errorf("Got Variable.Name %v, expected %v", x.Name, y.Name)
		}
	case *typevars.Function:
		y := tested.(*typevars.Function)
		if x.Name != y.Name {
			t.Errorf("Got Variable.Name %v, expected %v", x.Name, y.Name)
		}
	case *typevars.Argument:
		y := tested.(*typevars.Argument)
		compareTypeVars(t, &x.Function, &y.Function)
		if x.Index != y.Index {
			t.Errorf("Got Argument.Index %v, expected %v", x.Index, y.Index)
		}
	case *typevars.ReturnType:
		y := tested.(*typevars.ReturnType)
		compareTypeVars(t, &x.Function, &y.Function)
		if x.Index != y.Index {
			t.Errorf("Got ReturnType.Index %v, expected %v", x.Index, y.Index)
		}
	case *typevars.ListValue:
		y := tested.(*typevars.ListValue)
		compareTypeVars(t, &x.Constant, &y.Constant)
	case *typevars.MapKey:
		y := tested.(*typevars.MapKey)
		compareTypeVars(t, &x.Constant, &y.Constant)
	case *typevars.MapValue:
		y := tested.(*typevars.MapValue)
		compareTypeVars(t, &x.Constant, &y.Constant)
	case *typevars.Field:
		y := tested.(*typevars.Field)
		compareTypeVars(t, &x.Variable, &y.Variable)
		if x.Name != y.Name {
			t.Errorf("Got Field.Name %v, expected %v", x.Name, y.Name)
		}
		if x.Name == "" {
			if x.Index != y.Index {
				t.Errorf("Got Field.Index %v, expected %v", x.Index, y.Index)
			}
		}
	default:
		t.Errorf("Unrecognized data type: %#v", x)
	}
}

func compareContracts(t *testing.T, expected, tested contracts.Contract) {
	if expected.GetType() != tested.GetType() {
		t.Errorf("Expected %v, got %v instead", expected.GetType(), tested.GetType())
	}
	switch x := expected.(type) {
	case *contracts.PropagatesTo:
		y := tested.(*contracts.PropagatesTo)
		compareTypeVars(t, x.X, y.X)
		compareTypeVars(t, x.Y, y.Y)
	case *contracts.IsCompatibleWith:
		y := tested.(*contracts.IsCompatibleWith)
		compareTypeVars(t, x.X, y.X)
		compareTypeVars(t, x.Y, y.Y)
	case *contracts.BinaryOp:
		y := tested.(*contracts.BinaryOp)
		compareTypeVars(t, x.X, y.X)
		compareTypeVars(t, x.Y, y.Y)
		compareTypeVars(t, x.Z, y.Z)
		if x.OpToken != y.OpToken {
			t.Errorf("Expected BinaryOp.Op %q, got %q instead", x.OpToken, y.OpToken)
		}
	case *contracts.UnaryOp:
		y := tested.(*contracts.UnaryOp)
		compareTypeVars(t, x.X, y.X)
		compareTypeVars(t, x.Y, y.Y)
		if x.OpToken != y.OpToken {
			t.Errorf("Expected UnaryOp.Op %q, got %q instead", x.OpToken, y.OpToken)
		}
	case *contracts.IsInvocable:
		y := tested.(*contracts.IsInvocable)
		compareTypeVars(t, x.F, y.F)
		if x.ArgsCount != y.ArgsCount {
			t.Errorf("Expected number of arguments to be %v, got %v instead", x.ArgsCount, y.ArgsCount)
		}
	case *contracts.HasField:
		y := tested.(*contracts.HasField)
		compareTypeVars(t, &x.X, &y.X)
		if x.Field != y.Field {
			t.Errorf("Expected HasField.Field %q, got %q instead", x.Field, y.Field)
		}
		if x.Field == "" {
			if x.Index != y.Index {
				t.Errorf("Expected HasField.Index %q, got %q instead", x.Index, y.Index)
			}
		}
	case *contracts.IsDereferenceable:
		y := tested.(*contracts.IsDereferenceable)
		compareTypeVars(t, x.X, y.X)
	case *contracts.IsReferenceable:
		y := tested.(*contracts.IsReferenceable)
		compareTypeVars(t, x.X, y.X)
	case *contracts.DereferenceOf:
		y := tested.(*contracts.DereferenceOf)
		compareTypeVars(t, x.X, y.X)
		compareTypeVars(t, x.Y, y.Y)
	case *contracts.ReferenceOf:
		y := tested.(*contracts.ReferenceOf)
		compareTypeVars(t, x.X, y.X)
		compareTypeVars(t, x.Y, y.Y)
	default:
		t.Errorf("Contract %#v not recognized", expected)
	}
}

var packageName = "github.com/gofed/symbols-extractor/pkg/parser/testdata/contracts"

var vars = map[string]string{
	"c":        ":68:c",
	"d":        ":102:d",
	"e":        ":141:e",
	"f":        ":156:f",
	"ff":       ":171:ff",
	"a":        ":363:a",
	"g":        ":183:g",
	"g1":       ":232:g1",
	"g2":       ":276:g2",
	"aa":       ":471:aa",
	"ab":       ":488:ab",
	"b":        ":505:b",
	"list":     ":527:list",
	"mapV":     ":557:mapV",
	"structV":  ":606:structV",
	"structV2": ":687:structV2",
	"listV2":   ":757:listV2",
	"ra":       ":823:ra",
	"chanA":    ":834:chanA",
	"chanValA": ":859:chanValA",
	"uopa":     ":881:uopa",
	"uopb":     ":893:uopb",
	"uopc":     ":905:uopc",
	"uopd":     ":920:uopd",
	"uope":     ":893:uope",
	"bopa":     ":1008:bopa",
	"bopb":     ":1127:bopb",
	"bopc":     ":1166:bopc",
	"bopd":     ":1327:bopd",
	"bope":     ":1413:bope",
	"bopf":     ":1488:bopf",
	"bopg":     ":1543:bopg",
	"da":       ":1601:da",
}

type TestSuite struct {
	group     string
	contracts []contracts.Contract
}

func createGeneralTestSuite() *TestSuite {
	return &TestSuite{
		group: "general",
		contracts: []contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(vars["c"], packageName),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeVar(vars["d"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(1),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(vars["e"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["d"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(2),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(vars["f"], packageName),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["e"], packageName),
				Y:       typevars.MakeVar(vars["f"], packageName),
				Z:       typevars.MakeVirtualVar(3),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(3),
				Y: typevars.MakeVar(vars["ff"], packageName),
			},
		},
	}
}

func createFunctionInvocationTestSuite() *TestSuite {
	return &TestSuite{
		group: "function invocation",
		contracts: []contracts.Contract{
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeVar(vars["a"], packageName),
			},
			//
			// aa := a + g1(a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(vars["g1"], packageName),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g1"], packageName, 0),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeReturn(vars["g1"], packageName, 0),
				Z:       typevars.MakeVirtualVar(4),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeVar(vars["aa"], packageName),
			},
			//
			// ab := a + g2(a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(vars["g2"], packageName),
				ArgsCount: 1,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g2"], packageName, 0),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeReturn(vars["g2"], packageName, 0),
				Z:       typevars.MakeVirtualVar(5),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(5),
				Y: typevars.MakeVar(vars["ab"], packageName),
			},
			//
			// b := a + g(a, a, a)
			//
			&contracts.IsInvocable{
				F:         typevars.MakeFunction(vars["g"], packageName),
				ArgsCount: 3,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g"], packageName, 0),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g"], packageName, 1),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar(vars["a"], packageName),
				Y: typevars.MakeArgument(vars["g"], packageName, 2),
			},
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeReturn(vars["g"], packageName, 0),
				Z:       typevars.MakeVirtualVar(6),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(6),
				Y: typevars.MakeVar(vars["b"], packageName),
			},
		},
	}
}

func createCompositeLiteralTestSuite() *TestSuite {
	return &TestSuite{
		group: "composite literals",
		contracts: []contracts.Contract{
			//
			// list := []Int{
			// 	1,
			// 	2,
			// }
			//
			// Int <-> 1
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(&gotypes.Identifier{
					Def:     "Int",
					Package: packageName,
				}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// Int <-> 2
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(&gotypes.Identifier{
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
				X: typevars.MakeMapKey(&gotypes.Builtin{Untyped: false, Def: "string"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// 3 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// "4" <-> string
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapKey(&gotypes.Builtin{Untyped: false, Def: "string"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// 4 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
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
				X:     *typevars.MakeVirtualVar(7),
				Field: "key1",
			},
			// key1 <-> string
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(7), "key1", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// key2 exists
			&contracts.HasField{
				X:     *typevars.MakeVirtualVar(7),
				Field: "key2",
			},
			// key2 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(7), "key2", 0),
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
				Y: typevars.MakeVirtualVar(7),
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
				X:     *typevars.MakeVirtualVar(8),
				Index: 0,
			},
			// key at pos 0 <-> string
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(8), "", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
			},
			// key at pos 1 used
			&contracts.HasField{
				X:     *typevars.MakeVirtualVar(8),
				Index: 1,
			},
			// key at pos 1 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(8), "", 1),
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
				Y: typevars.MakeVirtualVar(8),
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
				X: typevars.MakeListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// 2 <-> int
			// TODO(jchaloup): documents this as "constant contract" a.k.a does not consume any data type
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// []int <-> {1,2}
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(&gotypes.Slice{
					Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeConstant(&gotypes.Slice{
					Elmtype: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
			},
			// 3 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// 4 <-> int
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(&gotypes.Builtin{Untyped: false, Def: "int"}),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// []int <-> {3,4}
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(&gotypes.Slice{
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
		},
	}
}

func createUnaryOperatorTestSuite() *TestSuite {
	return &TestSuite{
		group: "unary operator",
		contracts: []contracts.Contract{
			//
			// ra := &a
			//
			&contracts.UnaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeVirtualVar(9),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(9),
				Y: typevars.MakeVar(vars["ra"], packageName),
			},
			//
			// chanA := make(chan int)
			//
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Channel{
					Dir:   "3",
					Value: &gotypes.Builtin{Untyped: false, Def: "int"},
				}),
				Y: typevars.MakeVar(vars["chanA"], packageName),
			},
			//
			// chanValA := <-chanA
			//
			&contracts.UnaryOp{
				X:       typevars.MakeVar(vars["chanA"], packageName),
				Y:       typevars.MakeVirtualVar(10),
				OpToken: token.ARROW,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(10),
				Y: typevars.MakeVar(vars["chanValA"], packageName),
			},
			//
			// uopa := ^1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(11),
				OpToken: token.XOR,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(11),
				Y: typevars.MakeVar(vars["uopa"], packageName),
			},
			//
			// uopb := -1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(12),
				OpToken: token.SUB,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(12),
				Y: typevars.MakeVar(vars["uopb"], packageName),
			},
			//
			// uopc := !true
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeVirtualVar(13),
				OpToken: token.NOT,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(13),
				Y: typevars.MakeVar(vars["uopc"], packageName),
			},
			//
			// uopd := +1
			//
			&contracts.UnaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVirtualVar(14),
				OpToken: token.ADD,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(14),
				Y: typevars.MakeVar(vars["uopd"], packageName),
			},
		},
	}
}

func createBinaryOperatorTestSuite() *TestSuite {
	return &TestSuite{
		group: "binary operator",
		contracts: []contracts.Contract{
			//
			// bopa := 1 == 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(15),
				OpToken: token.EQL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(15),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 != 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(16),
				OpToken: token.NEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(16),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 <= 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(17),
				OpToken: token.LEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(17),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 < 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(18),
				OpToken: token.LSS,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(18),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 >= 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(19),
				OpToken: token.GEQ,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(19),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopa = 1 > 2
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(20),
				OpToken: token.GTR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(20),
				Y: typevars.MakeVar(vars["bopa"], packageName),
			},
			//
			// bopb := 8.0 << 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "float"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(21),
				OpToken: token.SHL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(21),
				Y: typevars.MakeVar(vars["bopb"], packageName),
			},
			//
			// bopb = 8.0 >> 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "float"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(22),
				OpToken: token.SHR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(22),
				Y: typevars.MakeVar(vars["bopb"], packageName),
			},
			//
			// bopc = bopc << 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopc"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(23),
				OpToken: token.SHL,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(23),
				Y: typevars.MakeVar(vars["bopc"], packageName),
			},
			//
			// bopc = bopc >> 1
			//
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopc"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(24),
				OpToken: token.SHR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(24),
				Y: typevars.MakeVar(vars["bopc"], packageName),
			},
			// bopd := true & false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(25),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(25),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bopd = true | false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(26),
				OpToken: token.OR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(26),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bopd = true &^ false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(27),
				OpToken: token.AND_NOT,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(27),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bopd = true ^ false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(28),
				OpToken: token.XOR,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(28),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// bope := 1 * 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(29),
				OpToken: token.MUL,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(29),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 - 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(30),
				OpToken: token.SUB,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(30),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 / 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(31),
				OpToken: token.QUO,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(31),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 + 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(32),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(32),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope := 1 % 1
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(33),
				OpToken: token.REM,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(33),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// var bopf int16
			// bope = bopf % 1
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopf"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(34),
				OpToken: token.REM,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(34),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bopd := true && false
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: false, Def: "bool"}),
				Z:       typevars.MakeVirtualVar(35),
				OpToken: token.LAND,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(35),
				Y: typevars.MakeVar(vars["bopd"], packageName),
			},
			// var bopg Int
			// bope = bopg + 1
			&contracts.BinaryOp{
				X:       typevars.MakeVar(vars["bopg"], packageName),
				Y:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Z:       typevars.MakeVirtualVar(36),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(36),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
			// bope = 1 + bopg
			&contracts.BinaryOp{
				X:       typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y:       typevars.MakeVar(vars["bopg"], packageName),
				Z:       typevars.MakeVirtualVar(37),
				OpToken: token.ADD,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVirtualVar(37),
				Y: typevars.MakeVar(vars["bope"], packageName),
			},
		},
	}
}

func createPointersTestSuite() *TestSuite {
	return &TestSuite{
		group: "pointers",
		contracts: []contracts.Contract{
			// da := *ra
			&contracts.IsDereferenceable{
				X: typevars.MakeVar(vars["ra"], packageName),
			},
			&contracts.DereferenceOf{
				X: typevars.MakeVar(vars["ra"], packageName),
				Y: typevars.MakeVirtualVar(38),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(38),
				Y: typevars.MakeVar(vars["da"], packageName),
			},
		},
	}
}

func createEmptyTestSuite() *TestSuite {
	return &TestSuite{
		group:     "empty",
		contracts: []contracts.Contract{},
	}
}

// TODO(jchaloup): replace this test with generated tests
func TestDataTypes(t *testing.T) {
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata/contracts"
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, "typevars.go")

	f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, 0)
	if err != nil {
		t.Fatalf("AST Parse error: %v", err)
	}

	payload, err := fileparser.MakePayload(f)
	if err != nil {
		t.Errorf("Unable to make a payload due to: %v", err)
	}

	parser, config, err := utils.InitFileParser(gopkg)
	if err != nil {
		t.Error(err)
	}

	if err := parser.Parse(payload); err != nil {
		t.Errorf("Unable to parse file %v: %v", gofile, err)
	}

	tests := []*TestSuite{
		createGeneralTestSuite(),
		createFunctionInvocationTestSuite(),
		createCompositeLiteralTestSuite(),
		createUnaryOperatorTestSuite(),
		createBinaryOperatorTestSuite(),
		createPointersTestSuite(),
	}

	contractsList := config.ContractTable.Contracts()

	testsTotal := 0
	for _, test := range tests {
		testsTotal += len(test.contracts)
	}

	if len(contractsList) < testsTotal {
		t.Errorf("Expected at least %q contracts, got %q instead", testsTotal, len(contractsList))
	}

	c := 0
	for _, testSuite := range tests {
		t.Logf("Checking %q group:", testSuite.group)
		for j, exp := range testSuite.contracts {
			t.Logf("Checking %v-th contract: %v\n", j, contracts.Contract2String(contractsList[c]))
			compareContracts(t, exp, contractsList[c])
			c++
		}
	}
	for i := c; i < len(contractsList); i++ {
		t.Logf("\n\n\n\nAbout to check %v-th contract: %v\n", i, contracts.Contract2String(contractsList[i]))
	}
}
