package main

import (
	"flag"
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/parser"
	"github.com/golang/glog"
)

type flags struct {
	packagePath     *string
	symbolTablePath *string
	cgoSymbolsPath  *string
	goVersion       *string
}

func (f *flags) parse() error {
	flag.Parse()

	if *(f.packagePath) == "" {
		return fmt.Errorf("--package-path is not set")
	}

	return nil
}

// Flow:
// 1) if no go version is set, the system go stdlib is processed
// 1.1) process builtin first (as it declares all built-in types (e.g. int, string, float),
//      functions (e.g. make, panic) and variables (e.g. true, false, nil))
// 1.2) start processing a package given by package-path
//
// 2) if go version is set, load the go stdlib from gofed/data (or other source)
// 2.1) if the version is not available fallback to processing the system go stdlib
// 2.2) start processing a package given by package-path
//
// Optionaly, if a package symbol table is provided, it is automatically loaded into the global symbol table
//
// TODO(jchaloup): account commits of individual packages
// The first implementation will expect all packages (and its deps) locally available.
// Later, one will specify a path to package symbol tables (each marked with corresponding commit)
func main() {

	f := &flags{
		packagePath:     flag.String("package-path", "", "Package entry point"),
		symbolTablePath: flag.String("symbol-table-dir", "", "Directory with preprocessed symbol tables"),
		// TODO(jchaloup): extend it with a hiearchy of cgo symbol files
		cgoSymbolsPath: flag.String("cgo-symbols-path", "", "Symbol table with CGO symbols (per entire project space)"),
		goVersion:      flag.String("go-version", "", "Go std library version"),
	}

	if err := f.parse(); err != nil {
		glog.Fatal(err)
	}

	// TODO(jchaloup): Check the version is of the form d.d for now (later extend with alpha/beta/rc...)

	fmt.Printf("Parsing: %v\n", *(f.packagePath))
	if err := parser.New(*(f.packagePath), *(f.symbolTablePath), *(f.cgoSymbolsPath)).Parse(); err != nil {
		glog.Fatalf("Parse error: %v", err)
	}

	fmt.Printf("PASSED\n")
}
