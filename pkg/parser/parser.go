package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func printDataType(dataType gotypes.DataType) {
	byteSlice, _ := json.Marshal(dataType)
	fmt.Printf("\n%v\n", string(byteSlice))
}

type FileParser struct {
	*types.Config
	// TODO(jchaloup):
	// - create a project scoped symbol table in a higher level struct (e.g. ProjectParser)
	// - merge all per file symbol tables continuously in the higher level struct (each time a new symbol definition is process)
	//
}

func NewParser(packageName string) *FileParser {
	c := &types.Config{
		PackageName:           packageName,
		SymbolTable:           symboltable.NewStack(),
		AllocatedSymbolsTable: alloctable.New(),
	}

	c.SymbolTable.Push()

	c.TypeParser = typeparser.New(c)
	c.ExprParser = exprparser.New(c)
	c.StmtParser = stmtparser.New(c)

	return &FileParser{c}
}

func (fp *FileParser) parseImportSpec(spec *ast.ImportSpec) error {
	fmt.Printf("spec: %#v\n", spec)

	q := &gotypes.PackageQualifier{
		Path: strings.Replace(spec.Path.Value, "\"", "", -1),
	}

	if spec.Name == nil {
		// TODO(jchaloup): get the q.Name from the spec.Path's package symbol table
		q.Name = path.Base(q.Path)
	} else {
		q.Name = spec.Name.Name
	}

	// TODO(jchaloup): store non-qualified imports as well
	if q.Name != "." {
		fp.SymbolTable.AddVariable(&gotypes.SymbolDef{Name: q.Name, Def: q})
		fmt.Printf("PQ added: %#v\n", &gotypes.SymbolDef{Name: q.Name, Def: q})
	}

	fmt.Printf("PQ: %#v\n", q)
	return nil
}

func (fp *FileParser) Parse(gofile string) error {
	fset := token.NewFileSet()
	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, gofile, nil, 0)
	if err != nil {
		return err
	}

	// Print the imports from the file's AST.
	for _, d := range f.Decls {
		//fmt.Printf("%v\n", d)
		// accessing dynamic_value := interface_variable.(typename)
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch d := spec.(type) {
				case *ast.ImportSpec:
					// process imports first
					//fmt.Printf("%+v\n", d)
					err := fp.parseImportSpec(d)
					if err != nil {
						return nil
					}
				case *ast.ValueSpec:
					// process value and constants as third
					//fmt.Printf("%+v\n", d)
					defs, err := fp.StmtParser.ParseValueSpec(d)
					if err != nil {
						return err
					}
					for _, def := range defs {
						// TPDP(jchaloup): we should store all variables or non.
						// Given the error is set only if the variable already exists, it should not matter so much.
						if err := fp.SymbolTable.AddVariable(def); err != nil {
							return nil
						}
					}
				case *ast.TypeSpec:
					// process type definitions as second
					//fmt.Printf("%#v\n", d)
					// Store a symbol just with a name and origin.
					// Setting the symbol's definition to nil means the symbol is being parsed (somewhere in the chain)
					if err := fp.SymbolTable.AddDataType(&gotypes.SymbolDef{
						Name:    d.Name.Name,
						Package: fp.PackageName,
						Def:     nil,
					}); err != nil {
						return err
					}

					// TODO(jchaloup): capture the current state of the allocated symbol table
					// JIC the parsing ends with end error. Which can result into re-parsing later on.
					// Which can result in re-allocation. It should be enough two-level allocated symbol table.
					typeDef, err := fp.TypeParser.Parse(d.Type)
					if err != nil {
						return err
					}

					if err := fp.SymbolTable.AddDataType(&gotypes.SymbolDef{
						Name:    d.Name.Name,
						Package: fp.PackageName,
						Def:     typeDef,
					}); err != nil {
						return err
					}
				}
			}
		case *ast.FuncDecl:
			// process function definitions as the last
			//fmt.Printf("%+v\n", d)
			funcDef, err := fp.StmtParser.ParseFuncDecl(decl)
			if err != nil {
				return err
			}

			if err := fp.SymbolTable.AddFunction(&gotypes.SymbolDef{
				Name:    decl.Name.Name,
				Package: fp.PackageName,
				Def:     funcDef,
			}); err != nil {
				return err
			}

			// if an error is returned, put the function's AST into a context list
			// and continue with other definition
			fp.StmtParser.ParseFuncBody(decl)
		}
	}

	fmt.Printf("\n\n")
	fp.AllocatedSymbolsTable.Print()

	// byteSlice, _ := json.Marshal(fp.SymbolTable)
	// fmt.Printf("\nSymbol table: %v\n", string(byteSlice))
	//
	// newObject := &symboltable.Table{}
	// if err := json.Unmarshal(byteSlice, newObject); err != nil {
	// 	fmt.Printf("Error: %v", err)
	// }
	//
	// fmt.Printf("\nAfter: %#v", newObject)

	return nil
}
