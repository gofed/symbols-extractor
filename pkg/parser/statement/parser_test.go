package statement

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

/**** HELP FUNCTIONS ****/

func prepareParser(pkgName string) *types.Config {
	c := &types.Config{
		PackageName:           pkgName,
		SymbolTable:           symboltable.NewStack(),
		AllocatedSymbolsTable: alloctable.New(),
	}

	c.SymbolTable.Push()
	c.TypeParser = typeparser.New(c)
	c.ExprParser = exprparser.New(c)
	c.StmtParser = New(c)

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
					// case *ast.ValueSpec:
					// 	_, err := config.TypeParser.Parse(d.Type)
					// 	if err != nil {
					// 		return err
					// 	}
					// 	config.SymbolTable.AddVariable(&gotypes.SymbolDef{
					// 		Name:    d.Names[0].Name,
					// 		Package: config.PackageName,
					// 		Def: &gotypes.Identifier{
					// 			Def: d.Names[0].Name,
					// 		},
					// 	})
				}
			}
		default:
			continue
		}
	}

	return nil
}

func iterFunc(astF *ast.File) []*ast.FuncDecl {
	var decls []*ast.FuncDecl

	for _, decl := range astF.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok {
			decls = append(decls, fdecl)
		}
	}

	return decls
}

/**** TEST FUNCTIONS ****/
//TODO(pstodulk)
// - test ParseReceiver
//func TestParseFunction(t *testing.T) {
//	// prepare testing data
//	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata/"
//	funcSrc := "package funcdecl\ntype Sstr string\n"
//	testFDecl := []struct {
//		fDecl     string
//		expectRes *gotypes.SymbolDef
//	}{
//		{
//			fDecl:     "func (s *Sstr, t *Sstr) Foo() Sstr { return Sstr{} }",
//			expectRes: nil,
//		},
//	}
//		  ff := fdecl.Recv.List[0]
//}

func createSymDefFunc(gopkg, recv, name string, method, pntr bool) *gotypes.SymbolDef {
	// help function which just create simple method/function SymbolDef
	// - the test below is much more readable now
	symdef := &gotypes.SymbolDef{
		Name:    name,
		Package: gopkg,
		Def:     &gotypes.Function{},
	}
	if method == false {
		return symdef
	} else if pntr {
		symdef.Def = &gotypes.Method{
			Receiver: &gotypes.Pointer{&gotypes.Identifier{recv}},
		}
	} else {
		symdef.Def = &gotypes.Method{Receiver: &gotypes.Identifier{recv}}
	}

	return symdef
}

func TestParseFuncDecl(t *testing.T) {
	//TODO(pstodulk): add Pos into preSymDef when it will be stored later
	// prepare testing data
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata/"
	funcSrc := "package funcdecl\ntype Sstr string\ntype Foo string\n"
	testFDecl := []struct {
		fDecl     []string
		expectRes []*gotypes.SymbolDef
	}{
		{
			[]string{"func (s *Sstr, t *Sstr) Foo() Sstr { return Sstr{} }"},
			[]*gotypes.SymbolDef{nil},
		},
		{
			[]string{"func () Foo() Sstr { return Sstr{} }"},
			[]*gotypes.SymbolDef{nil},
		},
		{
			[]string{"func (s *Sstra) Foo() Sstr { return Sstr{} }"},
			[]*gotypes.SymbolDef{nil},
		},
		{
			[]string{"func (s *int) Foo() Sstr { return Sstr{} }"},
			[]*gotypes.SymbolDef{nil},
		},
		{
			[]string{"func (s *Sstr) Foo() Sstr { return Sstr{} }"},
			[]*gotypes.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, true),
			},
		},
		{
			[]string{"func (s *Foo) Foo() Sstr { return Sstr{} }"},
			[]*gotypes.SymbolDef{
				createSymDefFunc(gopkg, "Foo", "Foo", true, true),
			},
		},
		{
			[]string{
				"func (s *Sstr) Foo() Sstr { return Sstr{} }",
				"func (s *Sstr) Foo2() Sstr { return Sstr{} }",
			},
			[]*gotypes.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, true),
				createSymDefFunc(gopkg, "Sstr", "Foo2", true, true),
			},
		},
		{
			[]string{
				"func (s Sstr) Foo() Sstr { return Sstr{} }",
				"func (s Sstr) Foo2() Sstr { return Sstr{} }",
			},
			[]*gotypes.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, false),
				createSymDefFunc(gopkg, "Sstr", "Foo2", true, false),
			},
		},
		{
			[]string{
				"func (s *Sstr) Foo() Sstr { return Sstr{} }",
				"func (s *Foo) Foo() Sstr { return Sstr{} }",
			},
			[]*gotypes.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, true),
				createSymDefFunc(gopkg, "Foo", "Foo", true, true),
			},
		},
		{
			[]string{
				"func (s *Sstr) Foo() Sstr { return Sstr{} }",
				"func Foo() Sstr { return Sstr{} }",
			},
			[]*gotypes.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, true),
				createSymDefFunc(gopkg, "", "Foo", false, false),
			},
		},
		/* these tests are valid, but always fail now
		 * TODO TODO: probably it is because of idea about parsing when symbol is unknown yet
		 * TODO: Write all other possible combinations - parameters and ret val is not
		 *       so important here, as it is deal of parseFunction, which will be
		 *       tested in different test. Important is testing of
		 *          - errors,
		 *          - returned symdef (function vs method, pointer vs identifier),
		 *          - store inside symbol table
		 */

		//		{
		//			[]string{
		//				"func Foo() Sstr { return Sstr{} }",
		//				"func Foo() Sstr { return Sstr{} }",
		//			},
		//			[]*gotypes.SymbolDef{
		//				createSymDefFunc(gopkg, "", "Foo", false, false),
		//				nil,
		//			},
		//		},
	}

	// NOW: walk through all test cases and:
	//   - at first part prepare "environment" (create AST, complete data)
	//   - then test
	// when test fails or there is nothing more can be done for curr test,
	// move to another test
TEST_LOOP:
	for _, tfdecl := range testFDecl {
		// create AST
		gocode := funcSrc + strings.Join(tfdecl.fDecl, "\n")
		astF, _, err := getAst(gopkg, "funcDecl.go", gocode)
		if err != nil {
			t.Errorf("Wrong input data: %v", err)
			return
		}

		//ast.Print(f, astF)

		// complete pre-test data (add Function type)
		config := prepareParser(gopkg)
		err = parseNonFunc(config, astF)
		if err != nil {
			t.Errorf("Can't parse types outside of body: %#v", err)
			continue
		}

		for i, fdecl := range iterFunc(astF) {
			if tfdecl.expectRes[i] == nil {
				continue
			}

			if funcDef, err := config.TypeParser.Parse(fdecl.Type); err == nil {
				if tfdecl.expectRes[i].Def.GetType() == gotypes.MethodType {
					tfdecl.expectRes[i].Def.(*gotypes.Method).Def = funcDef
				} else {
					tfdecl.expectRes[i].Def = funcDef
				}
			}
		}

		// TEST
		for i, fdecl := range iterFunc(astF) {

			// test that function returns expected values
			if ret, err := config.StmtParser.ParseFuncDecl(fdecl); err != nil {
				if tfdecl.expectRes[i] != nil {
					msgf := "Found error instead of expected result: %v" +
						"\n==== GOCODE ====\n%s\n==== END ===="
					t.Errorf(msgf, err, gocode)
				}
				// we are done here
				continue TEST_LOOP
			} else if tfdecl.expectRes[i] == nil {
				msgf := "Returned DataType instead of error.Code:\n%#v" +
					"\n==== GOCODE ====\n%s\n==== END ===="
				t.Errorf(msgf, ret, gocode)
				continue TEST_LOOP
			} else if !reflect.DeepEqual(tfdecl.expectRes[i].Def, ret) {
				msgf := "Returned different value:\nExpected: %#v\nGet: %#v" +
					"\n==== GOCODE ====\n%s\n==== END ===="
				t.Errorf(msgf, tfdecl.expectRes[i].Def, ret, gocode)
				continue TEST_LOOP
			}
			// complete curr symbol

			//fmt.Printf("-- type: %T -- val: %#v", ff, ff)
			i++
		}

		//FIXME: need modification of symbol table test will fail for last case
		// test that symbol table contains record about the symbol
		//for _, symdef := range tfdecl.expectRes {
		//  if tp.Lookup(symdef)
		//}
	}
	fmt.Println()
}

// func Test(t *testing.T) {}
