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
	default:
		t.Errorf("Contract %#v not recognized", expected)
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

	packageName := "github.com/gofed/symbols-extractor/pkg/parser/testdata/contracts"

	vars := map[string]string{
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
	}

	tests := []contracts.Contract{
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
	}

	contractsList := config.ContractTable.Contracts()

	if len(contractsList) < len(tests) {
		t.Errorf("Expected at least %q contracts, got %q instead", len(tests), len(contractsList))
	}

	c := 0
	for i, exp := range tests {
		c = i
		t.Logf("Checking %v-th contract: %v\n", i, contracts.Contract2String(contractsList[i]))
		compareContracts(t, exp, contractsList[i])
	}
	c++
	if c < len(contractsList) {
		t.Logf("\n\n\n\nAbout to check %v-th contract: %v\n", c, contracts.Contract2String(contractsList[c]))
	}
}
