package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser"
	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	"github.com/golang/glog"
)

type flags struct {
	packagePath     *string
	symbolTablePath *string
	cgoSymbolsPath  *string
	stdlib          *bool
	allocated       *bool
	perfile         *bool
	allallocated    *bool
	tojson          *bool
}

func (f *flags) parse() error {
	flag.Parse()

	if !(*f.stdlib) && *(f.packagePath) == "" {
		return fmt.Errorf("--package-path is not set")
	}

	return nil
}

func PrintAllocTables(allocTable *allocglobal.Table) {
	for _, pkg := range allocTable.Packages() {
		fmt.Printf("Package: %v\n", pkg)
		filesList := allocTable.Files(pkg)
		sort.Strings(filesList)
		for _, f := range filesList {
			fmt.Printf("\tFile: %v\n", f)
			at, err := allocTable.Lookup(pkg, f)
			if err != nil {
				glog.Fatalf("AC error: %v", err)
			}
			at.Print(true)
		}
	}
}

func printPackageAllocTables(allocTable *allocglobal.Table, pkg string, perfile bool, allallocated bool, tojson bool) {
	if perfile {
		filesList := allocTable.Files(pkg)
		sort.Strings(filesList)
		for _, f := range filesList {
			fmt.Printf("\tFile: %v\n", f)
			at, err := allocTable.Lookup(pkg, f)
			if err != nil {
				glog.Fatalf("AC error: %v", err)
			}
			at.Print(allallocated)
		}
	} else {
		tt, err := allocTable.MergeFiles(pkg)
		if err != nil {
			return
		}

		if tojson {
			byteSlice, err := json.Marshal(tt)
			if err != nil {
				fmt.Printf("Unable to convert print json: %v", err)
				os.Exit(1)
			}
			fmt.Printf("%v\n", string(byteSlice))
		} else {
			tt.Print(allallocated)
		}
	}
}

func getStdlibPackages() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Unable to get current WD: %v", err)
	}

	defer func() {
		// ignore the error
		os.Chdir(cwd)
	}()

	var packages []string

	output, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Unable to get GOROOT env: %v", err)
	}
	goroot := strings.Split(string(output), "\n")[0]

	err = os.Chdir(path.Join(goroot, "src"))
	if err != nil {
		return nil, fmt.Errorf("Unable to change dir to %v: %v", path.Join(goroot, "src"), err)
	}

	for _, pkg := range []string{"", "vendor/"} {
		output, err = exec.Command("go", "list", fmt.Sprintf("./%v...", pkg)).CombinedOutput()
		if err != nil {
			panic(fmt.Errorf("Unable to list packages under %v: %v", path.Join(goroot, "src"), err))
		}

		for _, line := range strings.Split(string(output), "\n") {
			if line == "" {
				continue
			}
			packages = append(packages, line)
		}
	}

	return packages, nil
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
		stdlib:         flag.Bool("stdlib", false, "Parse system Go std library"),
		allocated:      flag.Bool("allocated", false, "Extract allocation of symbols"),
		perfile:        flag.Bool("per-file", false, "Display allocated symbols per file"),
		allallocated:   flag.Bool("all-allocated", false, "Display all allocated symbols"),
		tojson:         flag.Bool("json", false, "Display allocated symbols in JSON"),
	}

	if err := f.parse(); err != nil {
		glog.Fatal(err)
	}

	// Otherwise it can eat all the CPU power
	runtime.GOMAXPROCS(1)

	output, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		glog.Fatal(fmt.Errorf("Error running `go version`: %v", err))
	}
	// TODO(jchaloup): Check the version is of the form d.d for now (later extend with alpha/beta/rc...)
	goversion := strings.Split(string(output), " ")[2][2:]

	// parse the standard library
	if *f.stdlib {
		packages, err := getStdlibPackages()
		if err != nil {
			glog.Fatal(err)
		}

		generatedDir := path.Join(*(f.symbolTablePath), "golang", goversion)

		for _, pkg := range packages {
			fmt.Printf("Parsing %q...\n", pkg)
			p := parser.New(pkg, generatedDir, *(f.cgoSymbolsPath), goversion)
			if err := p.Parse(false); err != nil {
				glog.Fatalf("Parse error when parsing (%v): %v", pkg, err)
			}
		}
		return
	}

	fmt.Printf("Parsing: %v\n", *(f.packagePath))
	p := parser.New(*(f.packagePath), *(f.symbolTablePath), *(f.cgoSymbolsPath), goversion)
	if err := p.Parse(*(f.allocated)); err != nil {
		glog.Fatalf("Parse error: %v", err)
	}

	if *(f.allocated) {
		printPackageAllocTables(p.GlobalAllocTable(), *(f.packagePath), *(f.perfile), *(f.allallocated), *(f.tojson))
	}
}
