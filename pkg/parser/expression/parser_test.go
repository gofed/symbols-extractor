package expression

import (
	"fmt"
	"go/ast"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/testing/utils"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func prepareParser(pkgName string) *types.Config {
	c := &types.Config{
		PackageName:           pkgName,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     global.New(""),
	}

	c.GlobalSymbolTable.Add("builtin", utils.BuiltinSymbolTable())

	c.SymbolTable.Push()
	c.TypeParser = typeparser.New(c)
	c.ExprParser = New(c)
	c.StmtParser = stmtparser.New(c)

	return c
}

//FIXME(pstodulk): incompatible with current changes. fix
//                 - just check correct parsing of expr. doesn't matter
//                   if outside or inside function body
func TestBinaryExpr(t *testing.T) {
	// prepare test
	var userType = "FInt"

	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata/valid"
	gocode := "package exprtest\ntype FInt int\n"
	testExpr := []struct {
		expRes        gotypes.DataType
		expr          string
		expectedError bool
	}{
		{&gotypes.Builtin{}, "var FooB1 int     = 1 + 4", false},
		// 1.2 + 4 is an untyped constant so combination of int and float is permitted
		{&gotypes.Builtin{}, "var FooB2 float32 = 1.2 + 4", false},
		{&gotypes.Identifier{Def: userType}, "var FooU3 FInt    = FInt(12) + FInt(4)", false},
		{&gotypes.Identifier{Def: userType}, "var FooU4 FInt    = FooU3 * FooU3", false},
		{&gotypes.Builtin{}, "var FooB5 uint32  = uint32(FooU3) + uint32(FooU4)", false},
		{&gotypes.Builtin{}, "var FooB5 uint32  = uint16(FooU3) + int(FooU4)", true},
	}

	// complete source code
	for _, test := range testExpr {
		gocode += test.expr + "\n"
	}

	astF, _, err := utils.GetAst(gopkg, "", gocode)
	if err != nil {
		t.Errorf("Wrong input data: %v", err)
		return
	}

	// create functionParser & parse general declarations
	config := prepareParser(gopkg)
	utils.ParseNonFunc(config, astF) // parse things outside of func (data types, vars)

	failed := false

	for i, valSpec := range utils.IterVar(astF) {
		fmt.Printf("Proceessing: %#v\n", testExpr[i].expr)
		binExpr := valSpec.Values[0].(*ast.BinaryExpr)
		res, err := config.ExprParser.(*Parser).parseBinaryExpr(binExpr)
		fmt.Printf("BinaryExprResult: %#v\terr: %v\n", res, err)
		// check that error is returned when is is expected

		if testExpr[i].expectedError && err == nil {
			t.Errorf("Expected error for '%s' did not occur", testExpr[i].expr)
			failed = true
			continue
		}

		if !testExpr[i].expectedError && err != nil {
			t.Errorf("Unexpected error for '%s' occurred: %v", testExpr[i].expr, err)
			failed = true
			continue
		}

		if !testExpr[i].expectedError && err == nil {
			// compare returned and expected value
			if testExpr[i].expRes.GetType() != res.GetType() {
				t.Errorf("Expected '%s' type, got '%s' instead. Line: '%s'",
					testExpr[i].expRes.GetType(),
					res.GetType(),
					testExpr[i].expr,
				)
				failed = true
			}
		}
	}

	if failed {
		t.Errorf("\n==== GOCODE ====\n%s\n==== END ====", gocode)
	}
}
