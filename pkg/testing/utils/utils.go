package utils

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"

	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func BuiltinSymbolTable() *symboltable.Table {
	st := symboltable.NewTable()
	addType := func(name string) {
		st.AddDataType(&symboltable.SymbolDef{
			Name:    name,
			Package: "builtin",
			Def: &gotypes.Builtin{
				Def: name,
			},
		})
	}

	addType("int")
	addType("string")
	addType("uint32")
	addType("uint16")
	addType("float32")

	return st
}

func GetAst(gopkg, filename string, gocode interface{}) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, filename)
	f, err := parser.ParseFile(fset, gofile, gocode, 0)
	if err != nil {
		return nil, fset, err
	}

	return f, fset, nil
}

func ParseNonFunc(config *types.Config, astF *ast.File) error {
	//TODO: later parsing of values will be required
	for _, d := range astF.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				//fmt.Printf("=== %#v", spec)
				switch d := spec.(type) {
				case *ast.TypeSpec:
					if err := config.SymbolTable.AddDataType(&symboltable.SymbolDef{
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

					if err := config.SymbolTable.AddDataType(&symboltable.SymbolDef{
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
					config.SymbolTable.AddVariable(&symboltable.SymbolDef{
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

func IterVar(astF *ast.File) []*ast.ValueSpec {
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

func IterFunc(astF *ast.File) []*ast.FuncDecl {
	var decls []*ast.FuncDecl

	for _, decl := range astF.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok {
			decls = append(decls, fdecl)
		}
	}

	return decls
}
