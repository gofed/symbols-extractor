package expression

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

/**** HELP FUNCTIONS ****/

func prepareParser(pkgName string) *types.Config {
	c := &types.Config{
		PackageName:           pkgName,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
	}

	c.SymbolTable.Push()
	c.TypeParser = typeparser.New(c)
	c.ExprParser = New(c)

	return c
}

func getAst(gopkg, filename string, gocode interface{}) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, filename)
	f, err := parser.ParseFile(fset, gofile, gocode, 0)
	if err != nil {
		return nil, fset, err
	}

	return f, fset, nil
}

func parseNonFunc(config *types.Config, astF *ast.File) error {
	//TODO: later parsing of values will be required
	for _, d := range astF.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				//fmt.Printf("=== %#v", spec)
				switch d := spec.(type) {
				case *ast.TypeSpec:
					if err := config.SymbolTable.AddDataType(&gotypes.SymbolDef{
						Name:    d.Name.Name,
						Package: config.PackageName,
						Def:     nil,
					}); err != nil {
						return err
					}

					typeDef, err := config.TypeParser.Parse(d.Type)
					if err != nil {
						return err
					}

					if err := config.SymbolTable.AddDataType(&gotypes.SymbolDef{
						Name:    d.Name.Name,
						Package: config.PackageName,
						Def:     typeDef,
					}); err != nil {
						return err
					}
				case *ast.ValueSpec:
					//TODO(pstodulk):
					//  - maybe identifier will be added automatically
					//    by typeparser into the symtab. Watch..
					//  - store type into the variable - now it is not possible
					//    varType, err := tp.ParseTypeExpr(d.Type)
					_, err := config.TypeParser.Parse(d.Type)
					if err != nil {
						return err
					}
					config.SymbolTable.AddVariable(&gotypes.SymbolDef{
						Name:    d.Names[0].Name,
						Package: config.PackageName,
						Def: &gotypes.Identifier{
							Def: d.Names[0].Name,
						},
					})
				}
			}
		default:
			continue
		}
	}

	return nil
}

func iterVar(astF *ast.File) []*ast.ValueSpec {
	var specs []*ast.ValueSpec
	for _, decl := range astF.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, valSpec := range genDecl.Specs {
			varDecl, ok := valSpec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			specs = append(specs, varDecl)
		}
	}
	return specs
}

/**** TEST FUNCTIONS ****/
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
		expectedError error
	}{
		{&gotypes.Builtin{}, "var FooB1 int     = 1 + 4", nil},
		{&gotypes.Builtin{}, "var FooB2 float32 = 1.2 + 4", nil},
		{&gotypes.Identifier{Def: userType}, "var FooU3 FInt    = FInt(12) + FInt(4)", nil},
		{&gotypes.Identifier{Def: userType}, "var FooU4 FInt    = FooU3 * FooU3", nil},
		{&gotypes.Builtin{}, "var FooB5 uint32  = uint32(FooU3) + uint32(FooU4)", nil},
		{&gotypes.Builtin{}, "var FooB5 uint32  = uint16(FooU3) + int(FooU4)", nil},
	}

	// complete source code
	for _, test := range testExpr {
		gocode += test.expr + "\n"
	}

	astF, _, err := getAst(gopkg, "", gocode)
	if err != nil {
		t.Errorf("Wrong input data: %v", err)
		return
	}

	// create functionParser & parse general declarations
	config := prepareParser(gopkg)
	parseNonFunc(config, astF) // parse things outside of func (data types, vars)

	failed := false

	for i, valSpec := range iterVar(astF) {
		binExpr := valSpec.Values[0].(*ast.BinaryExpr)
		res, err := config.ExprParser.(*Parser).parseBinaryExpr(binExpr)
		fmt.Printf("BinaryExprResult: %#v\n", res)
		// check that error is returned when is is expected

		if testExpr[i].expectedError != err {
			t.Errorf("Unexpected error for '%s': %v\n", testExpr[i].expr, err)
			failed = true
			continue
		}

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

	if failed {
		t.Errorf("\n==== GOCODE ====\n%s\n==== END ====", gocode)
	}
}
