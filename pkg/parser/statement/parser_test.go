package statement

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/stack"
	"github.com/gofed/symbols-extractor/pkg/testing/utils"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func prepareParser(pkgName string) *types.Config {
	c := &types.Config{
		PackageName:           pkgName,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     global.New("", ""),
	}
	c.SymbolsAccessor = accessors.NewAccessor(c.GlobalSymbolTable).SetCurrentTable(c.PackageName, c.SymbolTable)

	c.GlobalSymbolTable.Add("builtin", utils.BuiltinSymbolTable(), false)

	c.SymbolTable.Push()
	c.TypeParser = typeparser.New(c)
	c.ExprParser = exprparser.New(c)
	c.StmtParser = New(c)

	return c
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
//		expectRes *symboltable.SymbolDef
//	}{
//		{
//			fDecl:     "func (s *Sstr, t *Sstr) Foo() Sstr { return Sstr{} }",
//			expectRes: nil,
//		},
//	}
//		  ff := fdecl.Recv.List[0]
//}

func createSymDefFunc(gopkg, recv, name string, method, pntr bool) *symbols.SymbolDef {
	// help function which just create simple method/function SymbolDef
	// - the test below is much more readable now
	symdef := &symbols.SymbolDef{
		Name:    name,
		Package: gopkg,
		Def:     &gotypes.Function{},
	}
	if !method {
		return symdef
	}

	if pntr {
		symdef.Def = &gotypes.Method{
			Receiver: &gotypes.Pointer{&gotypes.Identifier{Def: recv, Package: gopkg}},
		}
	} else {
		symdef.Def = &gotypes.Method{
			Receiver: &gotypes.Identifier{Def: recv, Package: gopkg},
		}
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
		expectRes []*symbols.SymbolDef
	}{
		{
			[]string{"func (s *Sstr, t *Sstr) Foo() Sstr { return Sstr{} }"},
			[]*symbols.SymbolDef{nil},
		},
		{
			[]string{"func () Foo() Sstr { return Sstr{} }"},
			[]*symbols.SymbolDef{nil},
		},
		{
			[]string{"func (s *Sstra) Foo() Sstr { return Sstr{} }"},
			[]*symbols.SymbolDef{nil},
		},
		{
			[]string{"func (s *int) Foo() Sstr { return Sstr{} }"},
			[]*symbols.SymbolDef{nil},
		},
		{
			[]string{"func (s *Sstr) Foo() Sstr { return Sstr{} }"},
			[]*symbols.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, true),
			},
		},
		{
			[]string{"func (s *Foo) Foo() Sstr { return Sstr{} }"},
			[]*symbols.SymbolDef{
				createSymDefFunc(gopkg, "Foo", "Foo", true, true),
			},
		},
		{
			[]string{
				"func (s *Sstr) Foo() Sstr { return Sstr{} }",
				"func (s *Sstr) Foo2() Sstr { return Sstr{} }",
			},
			[]*symbols.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, true),
				createSymDefFunc(gopkg, "Sstr", "Foo2", true, true),
			},
		},
		{
			[]string{
				"func (s Sstr) Foo() Sstr { return Sstr{} }",
				"func (s Sstr) Foo2() Sstr { return Sstr{} }",
			},
			[]*symbols.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, false),
				createSymDefFunc(gopkg, "Sstr", "Foo2", true, false),
			},
		},
		{
			[]string{
				"func (s *Sstr) Foo() Sstr { return Sstr{} }",
				"func (s *Foo) Foo() Sstr { return Sstr{} }",
			},
			[]*symbols.SymbolDef{
				createSymDefFunc(gopkg, "Sstr", "Foo", true, true),
				createSymDefFunc(gopkg, "Foo", "Foo", true, true),
			},
		},
		{
			[]string{
				"func (s *Sstr) Foo() Sstr { return Sstr{} }",
				"func Foo() Sstr { return Sstr{} }",
			},
			[]*symbols.SymbolDef{
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
		//			[]*symboltable.SymbolDef{
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
		astF, _, err := utils.GetAst(gopkg, "funcDecl.go", gocode)
		if err != nil {
			t.Errorf("Wrong input data: %v", err)
			return
		}

		//ast.Print(f, astF)

		// complete pre-test data (add Function type)
		config := prepareParser(gopkg)
		err = utils.ParseNonFunc(config, astF)
		if err != nil {
			t.Errorf("Can't parse types outside of body: %#v", err)
			continue
		}

		for i, fdecl := range utils.IterFunc(astF) {
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
		for i, fdecl := range utils.IterFunc(astF) {
			fmt.Printf("Parsing %#v\n", tfdecl.fDecl[i])
			// test that function returns expected values
			ret, err := config.StmtParser.ParseFuncDecl(fdecl)
			if err != nil {
				if tfdecl.expectRes[i] != nil {
					msgf := "Found error instead of expected result: %v" +
						"\n==== GOCODE ====\n%s\n==== END ===="
					t.Errorf(msgf, err, gocode)
				}
				// we are done here
				continue TEST_LOOP
			}

			if tfdecl.expectRes[i] == nil {
				msgf := "Returned DataType instead of error.Code:\n%#v" +
					"\n==== GOCODE ====\n%s\n==== END ===="
				t.Errorf(msgf, ret, gocode)
				continue TEST_LOOP
			}

			if !reflect.DeepEqual(tfdecl.expectRes[i].Def, ret) {
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
