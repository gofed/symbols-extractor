package expression

import (
	"fmt"
	"go/ast"
	"go/parser"
	"reflect"
	"testing"

	//    "github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	//    "github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/parser/types"

	//    typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

const gopkg string = "github.com/gofed/symbols-extractor/pkg/parser/testdata/valid"
const gocode string = `
package exprtest

  type MyInt int

  var FooMyInt MyInt

  type MyStruct struct { Foo int }

  func MyFunc(MyInt) MyInt

  func MyFuncStr(MyInt) string

`

func initST() (*types.Config, error) {
	config := prepareParser(gopkg)
	astF, _, err := getAst(gopkg, "", gocode)
	if err != nil {
		return nil, fmt.Errorf("Broken test! Fix test suite: %v", err)
	}
	if err = parseNonFunc(config, astF); err != nil {
		return nil, err
	}

	if err = parseFuncDecls(config, astF); err != nil {
		return nil, fmt.Errorf("parseFuncDecls: %v", err)
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

func TestParseBinaryExpr0(t *testing.T) {
	expr_str := "2 + 4"
	expected_type := &gotypes.Builtin{
		Def: "int",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr1(t *testing.T) {
	expr_str := "8 * 3.1"
	expected_type := &gotypes.Builtin{
		Def: "float64",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr2(t *testing.T) {
	expr_str := "3 * 4 * 5"
	expected_type := &gotypes.Builtin{
		Def: "int",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr3(t *testing.T) {
	expr_str := "float32(4) - float32(3)"
	expected_type := &gotypes.Builtin{
		Def: "float32",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr4(t *testing.T) {
	expr_str := "MyInt(2) + 3"
	expected_type := &gotypes.Identifier{
		Def:     "MyInt",
		Package: gopkg,
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr5(t *testing.T) {
	expr_str := "MyFunc(2) + 3"
	expected_type := &gotypes.Identifier{
		Def:     "MyInt",
		Package: gopkg,
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr6(t *testing.T) {
	expr_str := "2 + MyFunc(3)"
	expected_type := &gotypes.Identifier{
		Def:     "MyInt",
		Package: gopkg,
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr7(t *testing.T) {
	expr_str := "FooMyInt + 5"
	expected_type := &gotypes.Identifier{
		Def:     "MyInt",
		Package: gopkg,
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr8(t *testing.T) {
	expr_str := "MyFunc(MyInt(3)) + MyFunc(MyInt(3))"
	expected_type := &gotypes.Identifier{
		Def:     "MyInt",
		Package: gopkg,
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr9(t *testing.T) {
	expr_str := "MyInt(MyFunc(MyInt(3))) + MyFunc(MyInt(3))"
	expected_type := &gotypes.Identifier{
		Def:     "MyInt",
		Package: gopkg,
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr10(t *testing.T) {
	expr_str := "\"Hello\" + \" Johnny!\""
	expected_type := &gotypes.Builtin{
		Def: "string",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr11(t *testing.T) {
	expr_str := "2 + int(MyFunc(3))"
	expected_type := &gotypes.Builtin{
		Def: "int",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr12(t *testing.T) {
	expr_str := "2 - 4"
	expected_type := &gotypes.Builtin{
		Def: "int",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr13(t *testing.T) {
	expr_str := "2 / 4"
	expected_type := &gotypes.Builtin{
		Def: "int",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr14(t *testing.T) {
	expr_str := "2 % 4"
	expected_type := &gotypes.Builtin{
		Def: "int",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr15(t *testing.T) {
	expr_str := "false || true"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr16(t *testing.T) {
	expr_str := "false && true"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr17(t *testing.T) {
	expr_str := "!bool(false) && bool(true)"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr18(t *testing.T) {
	expr_str := "5 > 1"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr19(t *testing.T) {
	expr_str := "5 < 1"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr20(t *testing.T) {
	expr_str := "5 <= 1"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr21(t *testing.T) {
	expr_str := "5 >= 1"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr22(t *testing.T) {
	expr_str := "MyFunc(15) == 15 "
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr23(t *testing.T) {
	expr_str := "5 != 1"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr24(t *testing.T) {
	expr_str := "!bool(true) ==  (1 != 5)"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}

func TestParseBinaryExpr25(t *testing.T) {
	expr_str := "true || ((5+1) * (MyFunc(11) / MyInt(14))) > 0"
	expected_type := &gotypes.Builtin{
		Def: "bool",
	}
	config, expr, errE := initExprTest(expr_str)
	if errE != nil {
		t.Error(errE)
		return
	}

	current_type, err := config.ExprParser.(*Parser).parseBinaryExpr(expr.(*ast.BinaryExpr))
	if err != nil {
		t.Errorf("Unexpected error for expr '%s': %v\n", expr_str, err)
		return
	}

	if !reflect.DeepEqual(current_type, expected_type) {
		msgf := "Expected type '%#v', received '%#v'. Expr: '%s' "
		t.Errorf(msgf, expected_type, current_type, expr_str)
	}
}
