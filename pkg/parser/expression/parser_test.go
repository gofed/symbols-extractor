package expression

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

/**** HELP FUNCTIONS ****/

func prepareParser(pkgName string) *Parser {
	symtab := symboltable.NewStack()
	allocSymtab := alloctable.New()
	tp := typeparser.New(pkgName, symtab, allocSymtab)

	symtab.Push()
	return New(pkgName, symtab, allocSymtab, tp)
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

func parseNonFunc(parser *Parser, astF *ast.File) error {
	//TODO: later parsing of values will be required
	tp := parser.typesParser
	for _, d := range astF.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				//fmt.Printf("=== %#v", spec)
				switch d := spec.(type) {
				case *ast.TypeSpec:
					if _, err := tp.Parse(d); err != nil {
						return err
					}
				case *ast.ValueSpec:
					//TODO(pstodulk):
					//  - maybe identifier will be added automatically
					//    by typeparser into the symtab. Watch..
					//  - store type into the variable - now it is not possible
					//    varType, err := tp.ParseTypeExpr(d.Type)
					_, err := tp.ParseTypeExpr(d.Type)
					if err != nil {
						return err
					}
					parser.symbolTable.AddVariable(&gotypes.SymbolDef{
						Name:    d.Names[0].Name,
						Package: parser.packageName,
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

func iterFunc(astF *ast.File) <-chan *ast.FuncDecl {
	// iterate over AST and returns function declarations
	fchan := make(chan *ast.FuncDecl)
	go func() {
		for _, decl := range astF.Decls {
			if fdecl, ok := decl.(*ast.FuncDecl); ok {
				fchan <- fdecl
			}
		}
		close(fchan)
	}()

	return fchan
}

func iterVar(astF *ast.File) <-chan *ast.ValueSpec {
	// iterate over AST and returns function declarations
	fchan := make(chan *ast.ValueSpec)
	go func() {
		for _, decl := range astF.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				if genDecl.Tok == token.VAR {
					fchan <- genDecl.Specs[0].(*ast.ValueSpec)
				}
			}
		}
		close(fchan)
	}()

	return fchan
}

/**** TEST FUNCTIONS ****/
//FIXME(pstodulk): incompatible with current changes. fix
//                 - just check correct parsing of expr. doesn't matter
//                   if outside or inside function body
func TestBinaryExpr(t *testing.T) {
	// prepare test
	var builtType string = gotypes.BuiltinType
	var userType string = "FInt"

	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata/valid"
	gocode := "package exprtest\ntype FInt int\n"
	testExpr := []struct {
		expRes *string
		expr   string
	}{
		{&builtType, "var FooB1 int     = 1 + 4"},
		{&builtType, "var FooB2 float32 = 1.2 + 4"},
		{&userType, "var FooU3 FInt    = FInt(12) + FInt(4)"},
		{&userType, "var FooU4 FInt    = FooU3 * FooU3"},
		{&builtType, "var FooB5 uint32  = uint32(FooU3) + uint32(FooU4)"},
		//{nil,        "var FooB5 uint32  = uint16(FooU3) + int(FooU4)"},
	}

	// complete source code
	for _, test := range testExpr {
		gocode += test.expr + "\n"
	}

	astF, _, err := getAst(gopkg, "exprtest.go", gocode)
	if err != nil {
		t.Errorf("Wrong input data: %v", err)
		return
	}

	// create functionParser & parse general declarations
	parser := prepareParser(gopkg)
	parseNonFunc(parser, astF) // parse things autside of func (data types, vars)
	//ast.Print(f, astF)

	// test itself
	failed := false
	i := 0
	for valSpec := range iterVar(astF) {
		binExpr := valSpec.Values[0].(*ast.BinaryExpr)
		res, err := parser.parseBinaryExpr(binExpr)

		// check that error is returned when is is expected
		if err != nil {
			if testExpr[i].expRes != nil {
				t.Errorf("error on line: '%s': %v\n", testExpr[i].expr, err)
				failed = true
				i++
				continue
			}
		} else {
			if testExpr[i].expRes == nil {
				// expected error instead of value
				msgf := "Returned the '%v' value instead of error. Line: '%s'"
				t.Errorf(msgf, res, testExpr[i].expr)
				failed = true
				i++
				continue
			}
		}

		// compare returned and expected value
		if *testExpr[i].expRes != res.GetType() {
			msgf := "Expected the %s type instead of %s. Line: '%s'"
			t.Errorf(msgf, *testExpr[i].expRes, res.GetType(), testExpr[i].expr)
			failed = true
		}
		i++
	}

	if failed {
		t.Errorf("\n==== GOCODE ====\n%s\n==== END ====", gocode)
	}
}
