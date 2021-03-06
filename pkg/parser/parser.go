package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	util "github.com/gofed/symbols-extractor/cmd/go"
	"github.com/gofed/symbols-extractor/pkg/analyzers/type/runner"
	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	contractglobal "github.com/gofed/symbols-extractor/pkg/parser/contracts/global"
	contracttable "github.com/gofed/symbols-extractor/pkg/parser/contracts/table"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/snapshots"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/stack"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"k8s.io/klog/v2"
)

var GOOS = map[string]struct{}{
	"android":   {},
	"darwin":    {},
	"dragonfly": {},
	"freebsd":   {},
	"linux":     {},
	"nacl":      {},
	"netbsd":    {},
	"openbsd":   {},
	"plan9":     {},
	"solaris":   {},
	"windows":   {},
}

var GOARCH = map[string]struct{}{
	"386":      {},
	"amd64":    {},
	"amd64p32": {},
	"arm64":    {},
	"arm":      {},
	"mips64":   {},
	"mips64le": {},
	"ppc64":    {},
	"ppc64le":  {},
	"s390x":    {},
}

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
	Constants []types.ConstSpec
	Functions []*ast.FuncDecl

	AllocatedSymbolsTable *alloctable.Table
}

// PackageContext storing context for a package
type PackageContext struct {
	// fully qualified package name
	PackagePath string
	PackageQID  string

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
	packagePath          string
	symbolTableDirectory string
	cgoSymbolsPath       string
	// Global symbol table
	globalSymbolTable *global.Table
	cgoSymbolTable    *tables.CGOTable
	// For each package and its file store its alloc symbol table
	globalAllocSymbolTable *allocglobal.Table
	globalContractsTable   *contractglobal.Table
	// package stack
	packageStack []*PackageContext

	goVersion string
	allocated bool
}

func New(symbolTableDir, cgoSymbolsPath, goVersion string, snapshot snapshots.Snapshot) (*ProjectParser, error) {
	pp := &ProjectParser{
		symbolTableDirectory:   symbolTableDir,
		cgoSymbolsPath:         cgoSymbolsPath,
		packageStack:           make([]*PackageContext, 0),
		globalSymbolTable:      global.New(symbolTableDir, goVersion, snapshot),
		globalAllocSymbolTable: allocglobal.New(symbolTableDir, goVersion, snapshot),
		globalContractsTable:   contractglobal.New(symbolTableDir, goVersion, snapshot),
		goVersion:              goVersion,
	}

	// set C pseudo-package
	pp.cgoSymbolTable = tables.NewCGOTable()
	pp.cgoSymbolTable.PackageQID = "C"
	if err := pp.globalSymbolTable.Add("C", pp.cgoSymbolTable, false); err != nil {
		return nil, fmt.Errorf("Unable to add C pseudo-package symbol table: %v", err)
	}

	return pp, nil
}

func (pp *ProjectParser) processImports(file string, imports []*ast.ImportSpec) (missingImports []*gotypes.Packagequalifier) {
	for _, spec := range imports {
		qPath := strings.Replace(spec.Path.Value, "\"", "", -1)
		// 'C' is a pseudo-package
		// See https://golang.org/cmd/cgo/
		if qPath == "C" {
			if pp.cgoSymbolsPath == "" {
				klog.Fatalf("Unable to load C symbol table. cgoSymbolsPath not set.")
			}
			pp.cgoSymbolTable.Flush()
			if err := pp.cgoSymbolTable.LoadFromFile(pp.cgoSymbolsPath); err != nil {
				panic(err)
			}
			continue
		}
		// Check if the imported package is already processed
		_, err := pp.globalSymbolTable.Lookup(qPath)
		if err != nil {
			missingImports = append(missingImports, &gotypes.Packagequalifier{Path: qPath})
			klog.V(2).Infof("Package %q not yet processed\n", qPath)
		}
		// TODO(jchaloup): Check if the package is already in the package queue
		//                 If it is it is an error (import cycles are not permitted)
	}
	return
}

func (pp *ProjectParser) getPackageFiles(packagePath string) (files []string, packageLocation string, err error) {

	listGoFiles := func(packagePath string, cgo bool) ([]string, error) {

		collectFiles := func(output string) []string {
			line := strings.Split(string(output), "\n")[0]
			line = line[1 : len(line)-1]
			if line == "" {
				return nil
			}
			return strings.Split(line, " ")
		}
		// check GOPATH/packagePath
		filter := "{{.GoFiles}}"
		if cgo {
			filter = "{{.CgoFiles}}"
		}
		cmd := exec.Command("go", "list", "-f", filter, packagePath)
		output, e := cmd.CombinedOutput()
		if e == nil {
			return collectFiles(string(output)), nil
		}

		if strings.Contains(string(output), "no buildable Go source files in") {
			return nil, nil
		}

		return nil, e
	}

	files, ppath, e := func() ([]string, string, error) {
		var searched []string
		// First searched the vendor directories
		pathParts := strings.Split(pp.packagePath, string(os.PathSeparator))
		for i := len(pathParts); i >= 0; i-- {
			vendorpath := path.Join(path.Join(pathParts[:i]...), "vendor", packagePath)
			klog.V(1).Infof("Checking %v directory", vendorpath)
			if l, e := listGoFiles(vendorpath, false); e == nil {
				return l, vendorpath, e
			}
			searched = append(searched, vendorpath)
		}

		klog.V(1).Infof("Checking %v directory", packagePath)
		if l, e := listGoFiles(packagePath, false); e == nil {
			return l, packagePath, e
		}
		searched = append(searched, packagePath)

		return nil, "", fmt.Errorf("Unable to find %q in any of:\n\t\t%v\n", packagePath, strings.Join(searched, "\n\t\t"))
	}()

	if e != nil {
		return nil, "", e
	}

	// cgo files enabled?
	cgoFiles, e := listGoFiles(ppath, true)
	if e != nil {
		return nil, "", e
	}

	if len(cgoFiles) > 0 {
		files = append(files, cgoFiles...)
	}

	{
		cmd := exec.Command("go", "list", "-f", "{{.Dir}}", ppath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, "", fmt.Errorf("go list -f {{.Dir}} %v failed: %v", ppath, err)
		}
		lines := strings.Split(string(output), "\n")
		packageLocation = string(lines[0])
	}

	return files, packageLocation, nil
}

func (pp *ProjectParser) createPackageContext(packagePath string) (*PackageContext, error) {
	c := &PackageContext{
		PackagePath: packagePath,
		FileIndex:   0,
		SymbolTable: stack.New(),
	}

	files, path, err := util.GetPackageFiles(pp.packagePath, packagePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to get %q package files: %v", packagePath, err)
	}
	c.PackageDir = path
	for _, file := range files {
		fc := &FileContext{
			Filename:              file,
			AllocatedSymbolsTable: alloctable.New(packagePath, file),
		}
		c.Files = append(c.Files, fc)
	}

	config := &types.Config{
		PackageName:       packagePath,
		SymbolTable:       c.SymbolTable,
		GlobalSymbolTable: pp.globalSymbolTable,
		ContractTable:     contracttable.New(packagePath, pp.symbolTableDirectory, pp.goVersion),
		SymbolsAccessor:   accessors.NewAccessor(pp.globalSymbolTable).SetCurrentTable(packagePath, c.SymbolTable),
	}

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	c.Config = config

	klog.V(2).Infof("PackageContextCreated: %#v\n\n", c)
	return c, nil
}

func (pp *ProjectParser) reprocessDataTypes(p *PackageContext) error {
	fLen := len(p.Files)
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		klog.V(2).Infof("File %q reprocessing...", path.Join(p.PackageDir, fileContext.Filename))
		if fileContext.DataTypes != nil {
			payload := &fileparser.Payload{
				DataTypes: fileContext.DataTypes,
			}
			klog.V(2).Infof("Types before processing: %#v\n", payload.DataTypes)
			for _, spec := range fileContext.FileAST.Imports {
				payload.Imports = append(payload.Imports, spec)
			}
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			p.Config.FileName = fileContext.Filename
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			klog.V(2).Infof("Types after processing: %#v\n", payload.DataTypes)
			if payload.DataTypes != nil {
				return fmt.Errorf("There are still some postponed data types to process after the second round: %v", p.PackagePath)
			}
		}
	}
	return nil
}

// Given a variable/const type can be undefined until its expression is parsed
// the variables/consts need to be reprocessed iteratively as long as number
// of unprocessed specs decreases
func (pp *ProjectParser) reprocessVariables(p *PackageContext) error {
	fLen := len(p.Files)
	pkgPostponedCounter := 0
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		pkgPostponedCounter += len(fileContext.Variables)
	}

	for {
		counter := 0
		for i := 0; i < fLen; i++ {
			fileContext := p.Files[i]
			klog.V(2).Infof("File %q reprocessing...", path.Join(p.PackageDir, fileContext.Filename))
			if fileContext.Variables != nil {
				payload := &fileparser.Payload{
					Variables:    fileContext.Variables,
					Reprocessing: true,
				}
				klog.V(2).Infof("Vars before processing: %#v\n", payload.Variables)
				for _, spec := range fileContext.FileAST.Imports {
					payload.Imports = append(payload.Imports, spec)
				}
				p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
				p.Config.FileName = fileContext.Filename
				if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
					return err
				}
				klog.V(2).Infof("Vars after processing: %#v\n", payload.Variables)
				fileContext.Variables = payload.Variables
				counter += len(payload.Variables)
			}
		}
		klog.V(2).Infof("len(postponed) before: %v, len(postponsed) after: %v\n", pkgPostponedCounter, counter)
		if counter < pkgPostponedCounter {
			pkgPostponedCounter = counter
			continue
		}
		break
	}
	return nil
}

func (pp *ProjectParser) reprocessConstants(p *PackageContext) error {
	fLen := len(p.Files)
	pkgPostponedCounter := 0
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		pkgPostponedCounter += len(fileContext.Constants)
	}

	for {
		counter := 0
		for i := 0; i < fLen; i++ {
			fileContext := p.Files[i]
			klog.V(2).Infof("File %q reprocessing...", path.Join(p.PackageDir, fileContext.Filename))
			if fileContext.Constants != nil {
				payload := &fileparser.Payload{
					Constants:    fileContext.Constants,
					Reprocessing: true,
				}
				klog.V(2).Infof("Constants before reprocessing: %#v\n", payload.Variables)
				for _, spec := range fileContext.FileAST.Imports {
					payload.Imports = append(payload.Imports, spec)
				}
				p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
				p.Config.FileName = fileContext.Filename
				if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
					return err
				}
				klog.V(2).Infof("Constants after reprocessing: %#v\n", payload.Constants)
				fileContext.Constants = payload.Constants
				counter += len(payload.Variables)
			}
		}
		klog.V(2).Infof("len(postponed) before: %v, len(postponsed) after: %v\n", pkgPostponedCounter, counter)
		if counter < pkgPostponedCounter {
			pkgPostponedCounter = counter
			continue
		}
		break
	}
	return nil
}

func (pp *ProjectParser) reprocessFunctionDeclarations(p *PackageContext) error {
	printFuncNames := func(funcs []*ast.FuncDecl) []string {
		var names []string
		for _, f := range funcs {
			names = append(names, f.Name.Name)
		}
		return names
	}
	fLen := len(p.Files)
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		klog.V(2).Infof("File %q reprocessing...", path.Join(p.PackageDir, fileContext.Filename))
		klog.V(2).Infof("Postponed symbols for file %q:\n\tD: %v\tV: %v\tF: %v", fileContext.Filename, len(fileContext.DataTypes), len(fileContext.Variables), len(fileContext.Functions))
		if fileContext.Functions != nil {
			payload := &fileparser.Payload{
				Functions:         fileContext.Functions,
				FunctionDeclsOnly: true,
			}
			klog.V(2).Infof("Funcs before processing: %#v\n", strings.Join(printFuncNames(payload.Functions), ","))
			for _, spec := range fileContext.FileAST.Imports {
				payload.Imports = append(payload.Imports, spec)
			}
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			p.Config.FileName = fileContext.Filename
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			klog.V(2).Infof("Funcs after processing: %#v\n", strings.Join(printFuncNames(payload.Functions), ","))
			if payload.Functions != nil {
				for _, name := range payload.Functions {
					klog.V(2).Infof("Function declaration of %q not yet processed", name.Name)
				}
			}
		}
	}
	return nil
}

func (pp *ProjectParser) reprocessFunctions(p *PackageContext) error {
	printFuncNames := func(funcs []*ast.FuncDecl) []string {
		var names []string
		for _, f := range funcs {
			names = append(names, f.Name.Name)
		}
		return names
	}
	fLen := len(p.Files)
	for i := 0; i < fLen; i++ {
		fileContext := p.Files[i]
		klog.V(2).Infof("File %q reprocessing...", path.Join(p.PackageDir, fileContext.Filename))
		klog.V(2).Infof("Postponed symbols for file %q:\n\tD: %v\tV: %v\tF: %v", fileContext.Filename, len(fileContext.DataTypes), len(fileContext.Variables), len(fileContext.Functions))
		if fileContext.Functions != nil {
			payload := &fileparser.Payload{
				Functions: fileContext.Functions,
			}
			klog.V(2).Infof("Funcs before processing: %#v\n", strings.Join(printFuncNames(payload.Functions), ","))
			for _, spec := range fileContext.FileAST.Imports {
				payload.Imports = append(payload.Imports, spec)
			}
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			p.Config.FileName = fileContext.Filename
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			klog.V(2).Infof("Funcs after processing: %#v\n", strings.Join(printFuncNames(payload.Functions), ","))
			if payload.Functions != nil {
				for _, name := range payload.Functions {
					klog.Errorf("Function %q not processed", name.Name)
				}
				panic(fmt.Errorf("There are still some postponed functions to process after the second round: %v", strings.Join(printFuncNames(payload.Functions), ",")))
			}
		}
	}
	return nil
}

func (pp *ProjectParser) packageProcessed(pkg string) bool {
	// TODO(jchaloup):
	// We should process only the API by default for the stdlib.
	// The allocated and the contracts are not important.
	// Plus, the contracts can grow up to tens of MB so they should be
	// optional as well and rather extracted only when needed.

	// API available?
	if !pp.globalSymbolTable.Exists(pkg) {
		return false
	}

	// Static allocations available?
	if !pp.globalAllocSymbolTable.Exists(pkg) {
		return false
	}

	// Contracts available
	if !pp.globalContractsTable.Exists(pkg) {
		return false
	}

	return true
}

func (pp *ProjectParser) Parse(packagePath string, allocated bool) error {
	pp.packagePath = packagePath
	pp.allocated = allocated

	// process builtin package first
	if !pp.packageProcessed("builtin") {
		if err := pp.processPackage("builtin"); err != nil {
			return err
		}
	}

	// check if the requested package is already provided
	if pp.packageProcessed(pp.packagePath) {
		klog.V(1).Infof("Package %q already processed\n", pp.packagePath)
		return nil
	}

	// process the requested package
	return pp.processPackage(pp.packagePath)
}

func (pp *ProjectParser) processPackage(packagePath string) error {
	// Process the input package
	c, err := pp.createPackageContext(packagePath)
	if err != nil {
		return err
	}
	// Push the input package into the package stack
	pp.packageStack = append(pp.packageStack, c)

PACKAGE_STACK:
	for len(pp.packageStack) > 0 {
		// Process the package stack
		p := pp.packageStack[0]
		if p.PackageDir == "" {
			// Pop the package from the package stack
			pp.packageStack = pp.packageStack[1:]
			continue
		}

		klog.V(2).Infof("\n\n\nPS processing %#v\n", p.PackageDir)
		if pp.packageProcessed(p.PackagePath) {
			klog.V(2).Infof("\n\n\nPS %#v already processed\n", p.PackageDir)
			// Pop the package from the package stack
			pp.packageStack = pp.packageStack[1:]
			continue
		}
		// a package may be processed again in case at least one of
		// api.json, allocated.json or contracts.json is missing
		pp.globalSymbolTable.Drop(p.PackagePath)
		pp.globalAllocSymbolTable.Drop(p.PackagePath)
		pp.globalContractsTable.Drop(p.PackagePath)

		// Process the files
		fLen := len(p.Files)
		for i := p.FileIndex; i < fLen; i++ {
			fileContext := p.Files[i]
			klog.V(2).Infof("File %q processing...", path.Join(p.PackageDir, fileContext.Filename))
			if fileContext.FileAST == nil {
				f, err := parser.ParseFile(token.NewFileSet(), path.Join(p.PackageDir, fileContext.Filename), nil, 0)
				if err != nil {
					return err
				}
				fileContext.FileAST = f
				// register the package name
				pkgQID := f.Name.Name
				if pkgQID != "main" {
					if p.PackageQID == "" {
						p.PackageQID = pkgQID
					} else if p.PackageQID != pkgQID {
						return fmt.Errorf("Package %v has at least two different per-file names: %q and %q", p.PackagePath, pkgQID, p.PackageQID)
					}
				}
			}
			// processed imported packages
			if !fileContext.ImportsProcessed {
				missingImports := pp.processImports(path.Join(p.PackagePath, fileContext.Filename), fileContext.FileAST.Imports)
				klog.V(2).Infof("Unknown imports:\t\t%#v\n\n", missingImports)
				if len(missingImports) > 0 {
					for _, spec := range missingImports {
						c, err := pp.createPackageContext(spec.Path)
						if err != nil {
							return err
						}

						pp.packageStack = append([]*PackageContext{c}, pp.packageStack...)
					}
					// At least one imported package is not yet processed
					klog.V(2).Infof("----Postponing %v\n\n", p.PackageDir)
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
			payload, err := fileparser.MakePayload(fileContext.FileAST)
			if err != nil {
				return fmt.Errorf("Unable to create a payload: %v", err)
			}
			p.Config.AllocatedSymbolsTable = fileContext.AllocatedSymbolsTable
			p.Config.FileName = fileContext.Filename
			if err := fileparser.NewParser(p.Config).Parse(payload); err != nil {
				return err
			}
			fileContext.DataTypes = payload.DataTypes
			fileContext.Constants = payload.Constants
			fileContext.Variables = payload.Variables
			fileContext.Functions = payload.Functions
			klog.V(2).Infof("Storing possible postponed symbols for file %q:\n\tD: %v\tV: %v\tF: %v", fileContext.Filename, len(fileContext.DataTypes), len(fileContext.Variables), len(fileContext.Functions))
			p.FileIndex++
		}

		// re-process data types
		klog.V(2).Infof("\n\n========REPROCESSING DATA TYPES========\n\n")
		if err := pp.reprocessDataTypes(p); err != nil {
			return err
		}

		// re-process function declerations
		klog.V(2).Infof("\n\n========REPROCESSING FUNC DECLS========\n\n")
		if err := pp.reprocessFunctionDeclarations(p); err != nil {
			klog.V(2).Infof("Processing function declarations with error: %v", err)
		}

		// re-process variables
		klog.V(2).Infof("\n\n========REPROCESSING VARIABLES========\n\n")
		if err := pp.reprocessConstants(p); err != nil {
			return err
		}

		// re-process variables
		klog.V(2).Infof("\n\n========REPROCESSING VARIABLES========\n\n")
		if err := pp.reprocessVariables(p); err != nil {
			return err
		}
		// re-process functions
		klog.V(2).Infof("\n\n========REPROCESSING FUNCTIONS========\n\n")
		if err := pp.reprocessFunctions(p); err != nil {
			return err
		}

		// Put the package ST into the global one
		table, err := p.SymbolTable.Table(0)
		if err != nil {
			panic(err)
		}

		// sort unique imported packages
		paths := make(map[string]struct{})
		for _, p := range table.Imports {
			paths[p] = struct{}{}
		}

		table.Imports = nil
		for p := range paths {
			table.Imports = append(table.Imports, p)
		}

		sort.Strings(table.Imports)

		klog.V(2).Infof("Global storing %q\n", p.PackagePath)
		fmt.Fprintf(os.Stderr, "Package %q processed\n", p.PackagePath)
		// TODO(jchaloup): this is hacky, the Add of the globalSymbolTable should
		// eat tables.Table instead of the generic SymbolTable

		if p.PackageQID == "" {
			p.PackageQID = path.Base(p.PackagePath)
		}
		table.PackageQID = p.PackageQID

		if err := pp.globalSymbolTable.Add(p.PackagePath, table, true); err != nil {
			panic(err)
		}

		// Store the allocated symbols
		for i := 0; i < fLen; i++ {
			fc := p.Files[i]
			pp.globalAllocSymbolTable.Add(p.PackagePath, fc.Filename, fc.AllocatedSymbolsTable)
			pp.globalContractsTable.Add(p.PackagePath, p.Config.ContractTable)
		}

		if err := pp.globalAllocSymbolTable.Save(p.PackagePath); err != nil {
			panic(err)
		}

		if pp.allocated {
			// Evaluate contracts to collect remaining allocated symbols (so called dynamicly allocated symbols).
			// The symbols are not stored as they depend on a specific dependency package commit
			r := runner.New(p.Config.PackageName, pp.globalSymbolTable, pp.globalAllocSymbolTable, p.Config.ContractTable)
			if err := r.Run(); err != nil {
				panic(err)
			}
		}

		if err := pp.globalContractsTable.Save(p.PackagePath); err != nil {
			panic(err)
		}

		// Pop the package from the package stack
		pp.packageStack = pp.packageStack[1:]
	}

	if pp.symbolTableDirectory != "" {
		return pp.globalSymbolTable.Save(pp.symbolTableDirectory)
	}
	return nil

}

func (pp *ProjectParser) GlobalSymbolTable() *global.Table {
	return pp.globalSymbolTable
}

func (pp *ProjectParser) GlobalAllocTable() *allocglobal.Table {
	return pp.globalAllocSymbolTable
}

func (pp *ProjectParser) GlobalContractsTable() *contractglobal.Table {
	return pp.globalContractsTable
}

func printDataType(dataType gotypes.DataType) {
	byteSlice, _ := json.Marshal(dataType)
	fmt.Printf("\n%v\n", string(byteSlice))
}
