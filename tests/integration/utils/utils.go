package utils

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	contracttable "github.com/gofed/symbols-extractor/pkg/parser/contracts/table"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	parsertypes "github.com/gofed/symbols-extractor/pkg/parser/types"
)

func parseBuiltin(config *parsertypes.Config) error {
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return fmt.Errorf("GOROOT env not set")
	}
	gofile := path.Join(goroot, "src", "builtin/builtin.go")

	f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, 0)
	if err != nil {
		return fmt.Errorf("AST Parse error: %v", err)
	}

	payload, err := fileparser.MakePayload(f)
	if err != nil {
		return err
	}
	if err := fileparser.NewParser(config).Parse(payload); err != nil {
		return fmt.Errorf("Unable to parse file %v: %v", gofile, err)
	}

	table, err := config.SymbolTable.Table(0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Global storing builtin\n")
	config.GlobalSymbolTable.Add("builtin", table)

	return nil
}

func InitFileParser(gopkg string) (*fileparser.FileParser, *types.Config, error) {
	gtable := global.New("")

	config := &parsertypes.Config{
		PackageName:           "builtin",
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     gtable,
		ContractTable:         contracttable.New(),
	}

	config.SymbolTable.Push()

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	if err := parseBuiltin(config); err != nil {
		return nil, nil, err
	}

	config = &parsertypes.Config{
		PackageName:           gopkg,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     gtable,
		ContractTable:         contracttable.New(),
	}

	config.SymbolTable.Push()

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	return fileparser.NewParser(config), config, nil
}

func CompareTypeVars(t *testing.T, expected, tested typevars.Interface) {
	if expected.GetType() != tested.GetType() {
		t.Errorf("Expected %v, got %v instead", expected.GetType(), tested.GetType())
	}
	switch x := expected.(type) {
	case *typevars.Constant:
		y := tested.(*typevars.Constant)
		if !reflect.DeepEqual(x.DataType, y.DataType) {
			xByteSlice, _ := json.Marshal(x.DataType)
			yByteSlice, _ := json.Marshal(y.DataType)
			t.Errorf("Got Constant.DataType:\n\t%v\nexpected:\n\t%v", string(yByteSlice), string(xByteSlice))
		}
	case *typevars.Variable:
		y := tested.(*typevars.Variable)
		if x.Name != y.Name {
			t.Errorf("Got Variable.Name %v, expected %v", y.Name, x.Name)
		}
	case *typevars.Function:
		y := tested.(*typevars.Function)
		if x.Name != y.Name {
			t.Errorf("Got Variable.Name %v, expected %v", y.Name, x.Name)
		}
	case *typevars.Argument:
		y := tested.(*typevars.Argument)
		CompareTypeVars(t, &x.Function, &y.Function)
		if x.Index != y.Index {
			t.Errorf("Got Argument.Index %v, expected %v", y.Index, x.Index)
		}
	case *typevars.ReturnType:
		y := tested.(*typevars.ReturnType)
		CompareTypeVars(t, &x.Function, &y.Function)
		if x.Index != y.Index {
			t.Errorf("Got ReturnType.Index %v, expected %v", y.Index, x.Index)
		}
	case *typevars.ListKey:
	case *typevars.ListValue:
		y := tested.(*typevars.ListValue)
		CompareTypeVars(t, x.Interface, y.Interface)
	case *typevars.MapKey:
		y := tested.(*typevars.MapKey)
		CompareTypeVars(t, x.Interface, y.Interface)
	case *typevars.MapValue:
		y := tested.(*typevars.MapValue)
		CompareTypeVars(t, x.Interface, y.Interface)
	case *typevars.RangeKey:
		y := tested.(*typevars.RangeKey)
		CompareTypeVars(t, x.Interface, y.Interface)
	case *typevars.RangeValue:
		y := tested.(*typevars.RangeValue)
		CompareTypeVars(t, x.Interface, y.Interface)
	case *typevars.Field:
		y := tested.(*typevars.Field)
		CompareTypeVars(t, x.Interface, y.Interface)
		if x.Name != y.Name {
			t.Errorf("Got Field.Name %v, expected %v", y.Name, x.Name)
		}
		if x.Name == "" {
			if x.Index != y.Index {
				t.Errorf("Got Field.Index %v, expected %v", y.Index, x.Index)
			}
		}
	default:
		t.Errorf("Unrecognized data type: %#v", x)
	}
}

func CompareContracts(t *testing.T, expected, tested contracts.Contract) {
	if expected.GetType() != tested.GetType() {
		t.Errorf("Expected %v, got %v instead", expected.GetType(), tested.GetType())
	}
	switch x := expected.(type) {
	case *contracts.PropagatesTo:
		y := tested.(*contracts.PropagatesTo)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
	case *contracts.IsCompatibleWith:
		y := tested.(*contracts.IsCompatibleWith)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
		if x.Weak != y.Weak {
			t.Errorf("Expected IsCompatibleWith.Week %q, got %q instead", x.Weak, y.Weak)
		}
	case *contracts.BinaryOp:
		y := tested.(*contracts.BinaryOp)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
		CompareTypeVars(t, x.Z, y.Z)
		if x.OpToken != y.OpToken {
			t.Errorf("Expected BinaryOp.Op %q, got %q instead", x.OpToken, y.OpToken)
		}
	case *contracts.UnaryOp:
		y := tested.(*contracts.UnaryOp)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
		if x.OpToken != y.OpToken {
			t.Errorf("Expected UnaryOp.Op %q, got %q instead", x.OpToken, y.OpToken)
		}
	case *contracts.IsInvocable:
		y := tested.(*contracts.IsInvocable)
		CompareTypeVars(t, x.F, y.F)
		if x.ArgsCount != y.ArgsCount {
			t.Errorf("Expected number of arguments to be %v, got %v instead", x.ArgsCount, y.ArgsCount)
		}
	case *contracts.HasField:
		y := tested.(*contracts.HasField)
		CompareTypeVars(t, x.X, y.X)
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
		CompareTypeVars(t, x.X, y.X)
	case *contracts.IsReferenceable:
		y := tested.(*contracts.IsReferenceable)
		CompareTypeVars(t, x.X, y.X)
	case *contracts.DereferenceOf:
		y := tested.(*contracts.DereferenceOf)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
	case *contracts.ReferenceOf:
		y := tested.(*contracts.ReferenceOf)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
	case *contracts.IsIndexable:
		y := tested.(*contracts.IsIndexable)
		CompareTypeVars(t, x.X, y.X)
	case *contracts.IsSendableTo:
		y := tested.(*contracts.IsSendableTo)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
	case *contracts.IsReceiveableFrom:
		y := tested.(*contracts.IsReceiveableFrom)
		CompareTypeVars(t, x.X, y.X)
		CompareTypeVars(t, x.Y, y.Y)
	case *contracts.IsIncDecable:
		y := tested.(*contracts.IsIncDecable)
		CompareTypeVars(t, x.X, y.X)
	case *contracts.IsRangeable:
		y := tested.(*contracts.IsRangeable)
		CompareTypeVars(t, x.X, y.X)
	default:
		t.Errorf("Contract %#v not recognized", expected)
	}
}
