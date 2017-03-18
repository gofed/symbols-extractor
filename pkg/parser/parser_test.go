package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"testing"
)

func TestDataTypes(t *testing.T) {
	fset := token.NewFileSet()
	gofile := path.Join(os.Getenv("GOPATH"), "src/github.com/gofed/symbols-extractor/pkg/parser/testdata/datatypes.go")
	f, err := parser.ParseFile(fset, gofile, nil, 0)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch d := spec.(type) {
				case *ast.TypeSpec:
					_, err := parseTypeSpec(d)
					if err != nil {
						t.Errorf("Unable to parse %#v: %v", d, err)
					}
				}
			}
		}
	}
}
