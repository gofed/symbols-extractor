package expression

import (
	"fmt"
	"go/ast"
	"go/parser"
	"reflect"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"

	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/testing/utils"

	// these are for parseBuiltin function
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	parsertypes "github.com/gofed/symbols-extractor/pkg/parser/types"
	"go/token"
	"os"
	"path"
	//"github.com/tonnerre/golang-pretty"
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
		return fmt.Errorf("Cannot make a payload: %v", err)
	}
	if err := fileparser.NewParser(config).Parse(payload); err != nil {
		return fmt.Errorf("Unable to parse file %v: %v", gofile, err)
	}

	table, err := config.SymbolTable.Table(0)
	if err != nil {
		panic(err)
	}

	config.GlobalSymbolTable.Add("builtin", table)

	return nil
}

func prepareParser(pkgName string) (*types.Config, error) {
	c := &types.Config{
		PackageName:           pkgName,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     global.New(""),
	}

	c.SymbolTable.Push()
	c.TypeParser = typeparser.New(c)
	c.ExprParser = New(c)
	c.StmtParser = stmtparser.New(c)

	// load all Builtin functions
	if err := parseBuiltin(c); err != nil {
		return nil, err
	}

	return c, nil
}

func builtinOrIdent(config *types.Config, str string, untyped bool) gotypes.DataType {
	if config.IsBuiltin(str) {
		return &gotypes.Builtin{Untyped: untyped, Def: str}
	}

	return &gotypes.Identifier{Def: str, Package: gopkg}
}

func parseTypes(config *types.Config, strTypes ...string) ([]gotypes.DataType, error) {
	parsedTypes := []gotypes.DataType{}

	for _, str := range strTypes {
		astType, err := parser.ParseExpr(str)
		if err != nil {
			return nil, err
		}

		parsedType, err := config.TypeParser.Parse(astType)
		if err != nil {
			return nil, err
		}

		parsedTypes = append(parsedTypes, parsedType)
	}

	return parsedTypes, nil
}

const gopkg string = "github.com/gofed/symbols-extractor/pkg/parser/testdata/valid"
const gocode string = `
package exprtest

type MyInt int
type MyStruct struct { Foo int }
var FooMyInt MyInt
var pFooMyInt *MyInt
var FooMyStruct MyStruct
func MyFunc(MyInt) MyInt
func MyFuncStr(MyInt) string
func (i *MyInt) Inc(increment MyInt)
func (i MyInt) GetAbs() uint
type iFace interface {}
var iFaceVarFloat iFace = float64(15.1)
var iFaceVarMyInt iFace = MyInt(15)
var mapStrInt map[string]int = map[string]int { "foo": 10}
var arrUint []uint = []uint{1,2,3,4,5,}
func Sum(a, b int) int
func SumEllipsis(i ...int) int
func Difference(a int, b int) int
func Div(a,b int32) (int, error)
func TransformMyStruct (a MyStruct) MyStruct
func Mapping() map[string]MyStruct
func fibonacci() func() int
func (s *MyStruct) getFoo() int
func (i *MyInt) Op(fce func (MyInt, MyInt) MyInt)
func Closure(s string, i int) func(int) (string, error)
func DClosure(fn func(int) uint, s string) (func(string) (int,error), func(int) uint)
`

func initST() (*types.Config, error) {
	var err error
	var config *types.Config
	var astF *ast.File

	config, err = prepareParser(gopkg)
	if err != nil {
		return nil, fmt.Errorf("Parser has not been prepared: %v", err)
	}

	astF, _, err = utils.GetAst(gopkg, "", gocode)
	if err != nil {
		return nil, fmt.Errorf("Broken test! Fix test suite: %v", err)
	}
	if err = utils.ParseNonFunc(config, astF); err != nil {
		return nil, fmt.Errorf("utils.ParseNonFunc: %v", err)
	}

	if err = utils.ParseFuncDecls(config, astF); err != nil {
		return nil, fmt.Errorf("utils.ParseFuncDecls: %v", err)
	}

	return config, nil
}

func initExprTest(expr_str string) (*types.Config, ast.Expr, error) {
	config, errT := initST()
	if errT != nil {
		return nil, nil, errT
	}

	expr, errE := parser.ParseExpr(expr_str)
	if errE != nil {
		return nil, nil, fmt.Errorf("Broken test! Fix test suite: %v", errE)
	}

	return config, expr, nil
}

func parseBinaryExprTest(expr_str, expected string, untyped bool) error {
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		return errE
	}
	expected_type := builtinOrIdent(config, expected, untyped)
	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		return fmt.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		return fmt.Errorf(msgf, expected_type, current_type, expr_str)
	}

	return nil
}

func parseCallExprTest(expr_str string, expected ...string) error {
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		return errE
	}

	expected_type, err := parseTypes(config, expected...)
	if err != nil {
		return fmt.Errorf("Unexpected error: Cannot get/parse expected types: %v\n", err)
	}

	current_type, err := config.ExprParser.(*Parser).parseCallExpr(expr.(*ast.CallExpr))
	if err != nil {
		msgf := "Unexpected error for expr '%s': %v\n"
		return fmt.Errorf(msgf, expr_str, err)
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		return fmt.Errorf(msgf, expected_type, current_type, expr_str)
	}

	return nil
}

func TestParseBinaryExpr0(t *testing.T) {
	if err := parseBinaryExprTest("2 + 4", "int", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr1(t *testing.T) {
	if err := parseBinaryExprTest("8 * 3.1", "float64", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr2(t *testing.T) {
	if err := parseBinaryExprTest("3 * 4 * 5", "int", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr3(t *testing.T) {
	if err := parseBinaryExprTest("float32(4) - float32(3)", "float32", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr4(t *testing.T) {
	if err := parseBinaryExprTest("MyInt(2) + 3", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr5(t *testing.T) {
	if err := parseBinaryExprTest("MyFunc(2) + 3", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr6(t *testing.T) {
	if err := parseBinaryExprTest("2 + MyFunc(3)", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr7(t *testing.T) {
	if err := parseBinaryExprTest("FooMyInt + 5", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr8(t *testing.T) {
	if err := parseBinaryExprTest("MyFunc(MyInt(3)) + MyFunc(MyInt(3))", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr9(t *testing.T) {
	if err := parseBinaryExprTest("MyInt(MyFunc(MyInt(3))) + MyFunc(MyInt(3))", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr10(t *testing.T) {
	if err := parseBinaryExprTest("\"Hello\" + \" Johnny!\"", "string", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr11(t *testing.T) {
	if err := parseBinaryExprTest("2 + int(MyFunc(3))", "int", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr12(t *testing.T) {
	if err := parseBinaryExprTest("2 - 4", "int", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr13(t *testing.T) {
	if err := parseBinaryExprTest("2 / 4", "int", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr14(t *testing.T) {
	if err := parseBinaryExprTest("2 % 4", "int", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr15(t *testing.T) {
	if err := parseBinaryExprTest("false || true", "bool", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr16(t *testing.T) {
	if err := parseBinaryExprTest("false && true", "bool", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr17(t *testing.T) {
	if err := parseBinaryExprTest("!bool(false) && bool(true)", "bool", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr18(t *testing.T) {
	if err := parseBinaryExprTest("5 > 1", "bool", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr19(t *testing.T) {
	if err := parseBinaryExprTest("5 < 1", "bool", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr20(t *testing.T) {
	if err := parseBinaryExprTest("5 <= 1", "bool", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr21(t *testing.T) {
	if err := parseBinaryExprTest("5 >= 1", "bool", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr22(t *testing.T) {
	if err := parseBinaryExprTest("MyFunc(15) == 15 ", "bool", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr23(t *testing.T) {
	if err := parseBinaryExprTest("5 != 1", "bool", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr24(t *testing.T) {
	if err := parseBinaryExprTest("!bool(true) ==  (1 != 5)", "bool", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr25(t *testing.T) {
	if err := parseBinaryExprTest("true || ((5+1) * (MyFunc(11) / MyInt(14))) > 0", "bool", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr26(t *testing.T) {
	if err := parseBinaryExprTest("(5 + 1) + (10 * 1)", "int", true); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr27(t *testing.T) {
	if err := parseBinaryExprTest("iota + iota", "int", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr28(t *testing.T) {
	if err := parseBinaryExprTest("*pFooMyInt + 10", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr29(t *testing.T) {
	if err := parseBinaryExprTest("iFaceVarFloat.(float64) - 10", "float64", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr30(t *testing.T) {
	if err := parseBinaryExprTest("iFaceVarMyInt.(MyInt) * 1", "MyInt", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr31(t *testing.T) {
	if err := parseBinaryExprTest("mapStrInt[\"foo\"] * 10", "int", false); err != nil {
		t.Error(err)
	}
}

func TestParseBinaryExpr32(t *testing.T) {
	if err := parseBinaryExprTest("arrUint[2] + 1", "uint", false); err != nil {
		t.Error(err)
	}
}

func TestParseCallExpr0(t *testing.T) {
	if err := parseCallExprTest("Sum(5, 3)", "int"); err != nil {
		t.Error(err)
	}

}

func TestParseCallExpr1(t *testing.T) {
	if err := parseCallExprTest("Div(5, 3)", "int", "error"); err != nil {
		t.Error(err)
	}

}

func TestParseCallExpr2(t *testing.T) {
	if err := parseCallExprTest("SumEllipsis(5, 3, 4)", "int"); err != nil {
		t.Error(err)
	}

}

func TestParseCallExpr3(t *testing.T) {
	if err := parseCallExprTest("MyFunc(MyInt(3))", "MyInt"); err != nil {
		t.Error(err)
	}

}

func TestParseCallExpr4(t *testing.T) {
	if err := parseCallExprTest("MyFuncStr(MyInt(2))", "string"); err != nil {
		t.Error(err)
	}

}

func TestParseCallExpr5(t *testing.T) {
	if err := parseCallExprTest("TransformMyStruct(FooMyStruct)", "MyStruct"); err != nil {
		t.Error(err)
	}

}

func TestParseCallExpr6(t *testing.T) {
	if err := parseCallExprTest("FooMyStruct.GetAbs()", "uint"); err != nil {
		t.Error(err)
	}

}

func TestParseCallExpr7(t *testing.T) {
	if err := parseCallExprTest("FooMyStruct.Inc(4)"); err != nil {
		t.Error(err)
	}

}
