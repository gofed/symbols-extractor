package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

// Context participants:
// - package (fully qualified package name, e.g. github.com/coreos/etcd/pkg/wait)
// - package file (package + its underlying filename)
// - package symbol definition (AST of a symbol definition)

// FileContext storing context for a file
type FileContext struct {
	// package's underlying filename
	Filename string
	// AST of a file (so the AST is not constructed again once all file's dependencies are processed)
	FileAST *ast.File
	// If set, there are still some symbols that needs processing
	ImportsProcessed bool
}

// Payload stores symbols for parsing/processing
type Payload struct {
	DataTypes []*ast.TypeSpec
	Variables []*ast.ValueSpec
	Functions []*ast.FuncDecl
	Imports   []*ast.ImportSpec
}

// PackageContext storing context for a package
type PackageContext struct {
	// fully qualified package name
	PackagePath string

	// files attached to a package
	PackageDir string
	FileIndex  int
	Files      []*FileContext

	Config *types.Config

	// package name
	PackageName string
	// per file symbol table
	SymbolTable *stack.Stack
	// per file allocatable ST
	AllocatedSymbolsTable *alloctable.Table

	// symbol definitions postponed (only variable/constants and function bodies definitions affected)
	DataTypes []*ast.TypeSpec
	Variables []*ast.ValueSpec
	Functions []*ast.FuncDecl
}

// Idea:
// - process the input package
// - retrieve all input package files
// - process each file of the input package
// - process a list of imported packages in each file
// - if any of the imported packages is not yet parsed out the package at the top of the package stack
// - pick a package from the top of the package stack
// - repeat the process until all imported packages are processed
// - continue processing declarations/definitions in the file
// - if any of the decls/defs in the file are not processed completely, put it in the postponed list
// - once all files in the package are processed, start re-processing the decls/defs in the postponed list
// - once all decls/defs are processed, clear the package and pick new package from the package stack

type ProjectParser struct {
	packagePath string
	// For each package and its file store its symbol table
	globalSymbolTable map[string]map[string]*symboltable.Table
	// For each package and its file store its alloc symbol table
	allocSymbolTable map[string]map[string]*alloctable.Table

	// package stack
	packageStack []*PackageContext
}

func New(packagePath string) *ProjectParser {
	return &ProjectParser{
		packagePath:  packagePath,
		packageStack: make([]*PackageContext, 0),
	}
}

func processImportSpec(spec *ast.ImportSpec) *gotypes.PackageQualifier {
	q := &gotypes.PackageQualifier{
		Path: strings.Replace(spec.Path.Value, "\"", "", -1),
	}

	if spec.Name == nil {
		// TODO(jchaloup): get the q.Name from the spec.Path's package symbol table
		q.Name = path.Base(q.Path)
	} else {
		q.Name = spec.Name.Name
	}

	return q
}

func (pp *ProjectParser) processImports(imports []*ast.ImportSpec) (missingImports []*gotypes.PackageQualifier) {
	for _, spec := range imports {
		q := processImportSpec(spec)
		// Check if the imported package is already processed
		_, ok := pp.globalSymbolTable[q.Path]
		if !ok {
			missingImports = append(missingImports, q)
			fmt.Printf("Package %q not yet processed\n", q.Path)
		}
		// TODO(jchaloup): Check if the package is already in the package queue
		//                 If it is it is an error (import cycles are not permitted)
	}
	return
}

func (pp *ProjectParser) getPackageFiles(packagePath string) (files []string, packageLocation string, err error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return nil, "", fmt.Errorf("GOPATH env not set")
	}

	// TODO(jchaloup): detect the GOROOT env from `go env` command
	goroot := "/usr/lib/golang/"
	godirs := []string{
		path.Join(goroot, "src", packagePath),
		path.Join(gopath, "src", packagePath),
	}
	for _, godir := range godirs {
		fileInfo, err := ioutil.ReadDir(godir)
		if err == nil {
			fmt.Printf("Checking %v...\n", godir)
			for _, file := range fileInfo {
				if !file.Mode().IsRegular() {
					continue
				}
				fmt.Printf("DirFile: %v\n", file.Name())
				// TODO(jchaloup): filter out unacceptable files (only *.go and *.s allowed)
				files = append(files, file.Name())
			}
			fmt.Printf("\n\n")
			return files, godir, nil
		}
	}

	return nil, "", fmt.Errorf("Package %q not found in any of %s locations", packagePath, strings.Join(godirs, ":"))
}

func (pp *ProjectParser) createPackageContext(packagePath string) (*PackageContext, error) {
	c := &PackageContext{
		PackagePath:           packagePath,
		FileIndex:             0,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
	}

	files, path, err := pp.getPackageFiles(packagePath)
	if err != nil {
		return nil, err
	}
	c.PackageDir = path
	for _, file := range files {
		c.Files = append(c.Files, &FileContext{Filename: file})
	}

	config := &types.Config{
		PackageName:           packagePath,
		SymbolTable:           c.SymbolTable,
		AllocatedSymbolsTable: c.AllocatedSymbolsTable,
	}

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	c.Config = config

	c.Config.SymbolTable.Push()

	fmt.Printf("PackageContextCreated: %#v\n\n", c)
	return c, nil
}

func (pp *ProjectParser) makePayload(f *ast.File) *Payload {
	payload := &Payload{}

	for _, spec := range f.Imports {
		payload.Imports = append(payload.Imports, spec)
	}

	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch d := spec.(type) {
				case *ast.TypeSpec:
					payload.DataTypes = append(payload.DataTypes, d)
				case *ast.ValueSpec:
					payload.Variables = append(payload.Variables, d)
				}
			}
		case *ast.FuncDecl:
			payload.Functions = append(payload.Functions, decl)
		}
	}
	return payload
}

func (pp *ProjectParser) Parse() error {
	// Process the input package
	c, err := pp.createPackageContext("github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered")
	if err != nil {
		return err
	}
	// Push the input package into the package stack
	pp.packageStack = append(pp.packageStack, c)

PACKAGE_STACK:
	for len(pp.packageStack) > 0 {
		// Process the package stack
		p := pp.packageStack[0]
		fmt.Printf("========PS processing %#v...========\n", p.PackageDir)
		// Process the files
		fLen := len(p.Files)
		for i := p.FileIndex; i < fLen; i++ {
			fileContext := p.Files[i]
			fmt.Printf("fx: %#v\n", fileContext)
			if fileContext.FileAST == nil {
				f, err := parser.ParseFile(token.NewFileSet(), path.Join(p.PackageDir, fileContext.Filename), nil, 0)
				if err != nil {
					return err
				}
				fileContext.FileAST = f
			}
			fmt.Printf("FileAST:\t\t%#v\n", fileContext.FileAST)
			// processed imported packages
			fmt.Printf("FileAST.Imports:\t%#v\n", fileContext.FileAST.Imports)
			if !fileContext.ImportsProcessed {
				missingImports := pp.processImports(fileContext.FileAST.Imports)
				fmt.Printf("Missing:\t\t%#v\n\n", missingImports)
				if len(missingImports) > 0 {
					for _, spec := range missingImports {
						fmt.Printf("Spec:\t\t\t%#v\n", spec)
						c, err := pp.createPackageContext(spec.Path)
						if err != nil {
							return err
						}

						pp.packageStack = append([]*PackageContext{c}, pp.packageStack...)

						fmt.Printf("PackageContext: %#v\n\n", c)
						// byteSlice, _ := json.Marshal(c)
						// fmt.Printf("\nPC: %v\n", string(byteSlice))
					}
					// At least one imported package is not yet processed
					fmt.Printf("----Postponing %v\n\n", p.PackageDir)
					fileContext.ImportsProcessed = true
					continue PACKAGE_STACK
				}
			}
			// All imported packages known => process the AST
			// TODO(jchaloup): reset the ST
			// Keep only the top-most ST
			if err := p.Config.SymbolTable.Reset(0); err != nil {
				panic(err)
			}
			payload := pp.makePayload(fileContext.FileAST)
			if err := NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			fmt.Printf("Types: %#v\n", payload.DataTypes)
			fmt.Printf("Vars: %#v\n", payload.Variables)
			fmt.Printf("Funcs: %#v\n", payload.Functions)
			p.DataTypes = append(p.DataTypes, payload.DataTypes...)
			p.Variables = append(p.Variables, payload.Variables...)
			p.Functions = append(p.Functions, payload.Functions...)
			p.FileIndex++
		}

		// Re-process postponed symbols
		if p.DataTypes != nil || p.Variables != nil || p.Functions != nil {
			// List of imports is already processed
			payload := &Payload{
				DataTypes: p.DataTypes,
				Variables: p.Variables,
				Functions: p.Functions,
			}
			if err := NewParser(p.Config).Parse(payload); err != nil {
				return err
			}

			fmt.Printf("Types: %#v\n", payload.DataTypes)
			fmt.Printf("Vars: %#v\n", payload.Variables)
			fmt.Printf("Funcs: %#v\n", payload.Functions)

			// All files are parsed. Check if there are any postponed symbols
			// If there are it is an error.
			if payload.DataTypes != nil || payload.Variables != nil || payload.Functions != nil {
				return fmt.Errorf("There are still some postponed symbols to process after the second round: %v", p.PackagePath)
			}
		}

		// Put the package ST into the global one
		byteSlice, _ := json.Marshal(p.SymbolTable)
		fmt.Printf("\nSymbol table: %v\n\n", string(byteSlice))

		// Pop the package from the package stack
		pp.packageStack = pp.packageStack[1:]
	}
	return nil
}

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

func NewParser(config *types.Config) *FileParser {
	return &FileParser{
		Config: config,
	}
}

func (fp *FileParser) parseImportSpec(spec *ast.ImportSpec) error {
	q := processImportSpec(spec)

	// TODO(jchaloup): store non-qualified imports as well
	if q.Name == "." {
		return nil
	}

	err := fp.SymbolTable.AddVariable(&gotypes.SymbolDef{Name: q.Name, Def: q})
	fmt.Printf("PQ added: %#v\n", &gotypes.SymbolDef{Name: q.Name, Def: q})
	fmt.Printf("PQ: %#v\n", q)
	return err
}

func (fp *FileParser) parseImportedPackages(specs []*ast.ImportSpec) error {
	for _, spec := range specs {
		err := fp.parseImportSpec(spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fp *FileParser) parseTypeSpecs(specs []*ast.TypeSpec) ([]*ast.TypeSpec, error) {
	var postponed []*ast.TypeSpec
	for _, spec := range specs {
		// Store a symbol just with a name and origin.
		// Setting the symbol's definition to nil means the symbol is being parsed (somewhere in the chain)
		if err := fp.SymbolTable.AddDataType(&gotypes.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Def:     nil,
		}); err != nil {
			return nil, err
		}

		// TODO(jchaloup): capture the current state of the allocated symbol table
		// JIC the parsing ends with end error. Which can result into re-parsing later on.
		// Which can result in re-allocation. It should be enough two-level allocated symbol table.
		typeDef, err := fp.TypeParser.Parse(spec.Type)
		if err != nil {
			postponed = append(postponed, spec)
			continue
		}

		if err := fp.SymbolTable.AddDataType(&gotypes.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Def:     typeDef,
		}); err != nil {
			return nil, err
		}
	}
	return postponed, nil
}

func (fp *FileParser) parseValueSpecs(specs []*ast.ValueSpec) ([]*ast.ValueSpec, error) {
	var postponed []*ast.ValueSpec
	for _, spec := range specs {

		defs, err := fp.StmtParser.ParseValueSpec(spec)
		if err != nil {
			postponed = append(postponed, spec)
			continue
		}
		for _, def := range defs {
			// TPDP(jchaloup): we should store all variables or non.
			// Given the error is set only if the variable already exists, it should not matter so much.
			if err := fp.SymbolTable.AddVariable(def); err != nil {
				return nil, nil
			}
		}
	}

	return postponed, nil
}

func (fp *FileParser) parseFuncs(specs []*ast.FuncDecl) ([]*ast.FuncDecl, error) {
	var postponed []*ast.FuncDecl
	for _, spec := range specs {

		// process function definitions as the last
		//fmt.Printf("%+v\n", d)
		funcDef, err := fp.StmtParser.ParseFuncDecl(spec)
		if err != nil {
			// TODO(jchaloup): this should not error (ever)
			// It can happen an identifier is unknown. Given all imported packages
			// are already processed, the identifier comes from the same package, just
			// from a different file. So the type parser should assume that and choose
			// symbol's origin accordingaly.
			postponed = append(postponed, spec)
			continue
		}

		if err := fp.SymbolTable.AddFunction(&gotypes.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Def:     funcDef,
		}); err != nil {
			return nil, err
		}

		// if an error is returned, put the function's AST into a context list
		// and continue with other definition
		if err := fp.StmtParser.ParseFuncBody(spec); err != nil {
			postponed = append(postponed, spec)
			continue
		}
	}

	return postponed, nil
}

func (fp *FileParser) Parse(p *Payload) error {

	// Given:
	// type My1 My2
	// type My2 My3
	// ...
	// type Myn Myn-1
	//
	// We can not re-process a symbol processing (due to unknown symbol) until we know
	// we process the unknown symbol before it is needed.
	// Given a type name is stored into the symbol table before its definition is parsed
	// all data type names are known in the second round of the postponed type definition processing

	// Imported packages first
	{
		err := fp.parseImportedPackages(p.Imports)
		if err != nil {
			return err
		}
	}

	// Data definitions as second
	{
		postponed, err := fp.parseTypeSpecs(p.DataTypes)
		if err != nil {
			return err
		}
		p.DataTypes = postponed
	}

	// // Vars/constants
	{
		postponed, err := fp.parseValueSpecs(p.Variables)
		if err != nil {
			return err
		}
		p.Variables = postponed
	}

	// // Funcs
	{
		postponed, err := fp.parseFuncs(p.Functions)
		if err != nil {
			return err
		}
		p.Functions = postponed
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
