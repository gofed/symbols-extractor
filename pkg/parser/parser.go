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
	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
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

	DataTypes []*ast.TypeSpec
	Variables []*ast.ValueSpec
	Functions []*ast.FuncDecl

	AllocatedSymbolsTable *alloctable.Table
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
	// per package symbol table
	SymbolTable *stack.Stack

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
	// Global symbol table
	globalSymbolTable *global.Table
	// For each package and its file store its alloc symbol table
	globalAllocSymbolTable *allocglobal.Table
	// package stack
	packageStack []*PackageContext
}

func New(packagePath string) *ProjectParser {
	return &ProjectParser{
		packagePath:            packagePath,
		packageStack:           make([]*PackageContext, 0),
		globalSymbolTable:      global.New(),
		globalAllocSymbolTable: allocglobal.New(),
	}
}

func (pp *ProjectParser) processImports(imports []*ast.ImportSpec) (missingImports []*gotypes.Packagequalifier) {
	for _, spec := range imports {
		q := fileparser.MakePackagequalifier(spec)
		// Check if the imported package is already processed
		_, err := pp.globalSymbolTable.Lookup(q.Path)
		if err != nil {
			missingImports = append(missingImports, q)
			glog.Infof("Package %q not yet processed\n", q.Path)
		}
		// TODO(jchaloup): Check if the package is already in the package queue
		//                 If it is it is an error (import cycles are not permitted)
	}
	return
}

func (pp *ProjectParser) getPackageFiles(packagePath string) (files []string, packageLocation string, err error) {
	var godirs []string

	// e.g. /usr/lib/golang
	goroot := os.Getenv("GOROOT")
	if goroot != "" {
		godirs = append(godirs, path.Join(goroot, "src", packagePath))
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return nil, "", fmt.Errorf("GOPATH env not set")
	}

	godirs = append(godirs, path.Join(gopath, "src", packagePath))

	for _, godir := range godirs {
		fileInfo, err := ioutil.ReadDir(godir)
		if err == nil {
			glog.Infof("Checking %v...\n", godir)
			for _, file := range fileInfo {
				if !file.Mode().IsRegular() {
					continue
				}
				// TODO(jchaloup): filter out unacceptable files (only *.go and *.s allowed)
				if !strings.HasSuffix(file.Name(), ".go") {
					continue
				}
				files = append(files, file.Name())
			}
			return files, godir, nil
		}
	}

	return nil, "", fmt.Errorf("Package %q not found in any of %s locations", packagePath, strings.Join(godirs, ":"))
}

func (pp *ProjectParser) createPackageContext(packagePath string) (*PackageContext, error) {
	c := &PackageContext{
		PackagePath: packagePath,
		FileIndex:   0,
		SymbolTable: stack.New(),
	}

	files, path, err := pp.getPackageFiles(packagePath)
	if err != nil {
		return nil, err
	}
	c.PackageDir = path
	for _, file := range files {
		fc := &FileContext{
			Filename:              file,
			AllocatedSymbolsTable: alloctable.New(),
		}
		c.Files = append(c.Files, fc)
		pp.globalAllocSymbolTable.Add(packagePath, file, fc.AllocatedSymbolsTable)
	}

	config := &types.Config{
		PackageName:       packagePath,
		SymbolTable:       c.SymbolTable,
		GlobalSymbolTable: pp.globalSymbolTable,
	}

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	c.Config = config

	glog.Infof("PackageContextCreated: %#v\n\n", c)
	return c, nil
}

func (pp *ProjectParser) reprocessDataTypes(p *PackageContext) error {
	fLen := len(p.Files)
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		glog.Infof("File %q processing...", fileContext.Filename)
		if fileContext.DataTypes != nil {
			payload := &fileparser.Payload{
				DataTypes: fileContext.DataTypes,
			}
			glog.Infof("Types before processing: %#v\n", payload.DataTypes)
			for _, spec := range fileContext.FileAST.Imports {
				payload.Imports = append(payload.Imports, spec)
			}
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			glog.Infof("Types after processing: %#v\n", payload.DataTypes)
			if payload.DataTypes != nil {
				return fmt.Errorf("There are still some postponed data types to process after the second round: %v", p.PackagePath)
			}
		}
	}
	return nil
}

func (pp *ProjectParser) reprocessVariables(p *PackageContext) error {
	fLen := len(p.Files)
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		glog.Infof("File %q processing...", fileContext.Filename)
		if fileContext.Variables != nil {
			payload := &fileparser.Payload{
				Variables: fileContext.Variables,
			}
			glog.Infof("Vars before processing: %#v\n", payload.Variables)
			for _, spec := range fileContext.FileAST.Imports {
				payload.Imports = append(payload.Imports, spec)
			}
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			glog.Infof("Vars after processing: %#v\n", payload.Variables)
			if payload.Variables != nil {
				return fmt.Errorf("There are still some postponed variables to process after the second round: %v", p.PackagePath)
			}
		}
	}
	return nil
}

func (pp *ProjectParser) reprocessFunctions(p *PackageContext) error {
	fLen := len(p.Files)
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		glog.Infof("File %q processing...", fileContext.Filename)
		if fileContext.Functions != nil {
			payload := &fileparser.Payload{
				Functions: fileContext.Functions,
			}
			fmt.Printf("Funcs before processing: %#v\n", payload.Functions)
			for _, spec := range fileContext.FileAST.Imports {
				payload.Imports = append(payload.Imports, spec)
			}
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			fmt.Printf("Funcs after processing: %#v\n", payload.Functions)
			if payload.Functions != nil {
				for _, name := range payload.Functions {
					glog.Errorf("Function %q not processed", name.Name)
				}
				return fmt.Errorf("There are still some postponed functions to process after the second round: %v", p.PackagePath)
			}
		}
	}
	return nil
}

func (pp *ProjectParser) Parse() error {
	// Process the input package
	c, err := pp.createPackageContext(pp.packagePath)
	if err != nil {
		return err
	}
	// Push the input package into the package stack
	pp.packageStack = append(pp.packageStack, c)

PACKAGE_STACK:
	for len(pp.packageStack) > 0 {
		// Process the package stack
		p := pp.packageStack[0]
		glog.Infof("PS processing %#v\n", p.PackageDir)
		// Process the files
		fLen := len(p.Files)
		for i := p.FileIndex; i < fLen; i++ {
			fileContext := p.Files[i]
			if fileContext.FileAST == nil {
				f, err := parser.ParseFile(token.NewFileSet(), path.Join(p.PackageDir, fileContext.Filename), nil, 0)
				if err != nil {
					return err
				}
				fileContext.FileAST = f
			}
			// processed imported packages
			if !fileContext.ImportsProcessed {
				missingImports := pp.processImports(fileContext.FileAST.Imports)
				glog.Infof("Unknown imports:\t\t%#v\n\n", missingImports)
				if len(missingImports) > 0 {
					for _, spec := range missingImports {
						c, err := pp.createPackageContext(spec.Path)
						if err != nil {
							return err
						}

						pp.packageStack = append([]*PackageContext{c}, pp.packageStack...)
					}
					// At least one imported package is not yet processed
					glog.Infof("----Postponing %v\n\n", p.PackageDir)
					fileContext.ImportsProcessed = true
					continue PACKAGE_STACK
				}
			}
			// All imported packages known => process the AST
			// TODO(jchaloup): reset the ST
			// Keep only the top-most ST
			if err := p.Config.SymbolTable.Reset(); err != nil {
				panic(err)
			}
			payload := fileparser.MakePayload(fileContext.FileAST)
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			fileContext.DataTypes = payload.DataTypes
			fileContext.Variables = payload.Variables
			fileContext.Functions = payload.Functions
			p.FileIndex++
		}

		// re-process data types
		if err := pp.reprocessDataTypes(p); err != nil {
			return err
		}

		// re-process variables
		if err := pp.reprocessVariables(p); err != nil {
			return err
		}
		// re-process functions
		if err := pp.reprocessFunctions(p); err != nil {
			return err
		}

		// Put the package ST into the global one
		byteSlice, _ := json.Marshal(p.SymbolTable)
		fmt.Printf("\nSymbol table: %v\n\n", string(byteSlice))

		table, err := p.SymbolTable.Table(0)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Global storing %q\n", p.PackagePath)
		if err := pp.globalSymbolTable.Add(p.PackagePath, table); err != nil {
			panic(err)
		}

		// Pop the package from the package stack
		pp.packageStack = pp.packageStack[1:]
	}

	return nil
}

func (pp *ProjectParser) GlobalSymbolTable() *global.Table {
	return pp.globalSymbolTable
}

func (pp *ProjectParser) GlobalAllocTable() *allocglobal.Table {
	return pp.globalAllocSymbolTable
}

func printDataType(dataType gotypes.DataType) {
	byteSlice, _ := json.Marshal(dataType)
	fmt.Printf("\n%v\n", string(byteSlice))
}
