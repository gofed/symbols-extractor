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

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/runner"
	"github.com/gofed/symbols-extractor/pkg/parser"
	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	contractglobal "github.com/gofed/symbols-extractor/pkg/parser/contracts/global"
	"github.com/gofed/symbols-extractor/pkg/snapshots"
	"github.com/gofed/symbols-extractor/pkg/snapshots/glide"
	"github.com/gofed/symbols-extractor/pkg/snapshots/godeps"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	"github.com/golang/glog"
)

type flags struct {
	packagePath     *string
	packagePrefix   *string
	symbolTablePath *string
	cgoSymbolsPath  *string

	stdlib        *bool
	allocated     *bool
	recursiveFrom *string
	filterPrefix  *string
	perfile       *bool
	allallocated  *bool
	tojson        *bool
	glidefile     *string
	godepsfile    *string
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

func printPackageAllocTables(allocTable *allocglobal.Table, globalTable *global.Table, contractTable *contractglobal.Table, f *flags) error {

	pkgs := strings.Split(*f.packagePath, ":")
	perfile := *f.perfile
	allallocated := *f.allallocated
	tojson := *f.tojson

	packageAllocTables := make(map[string]allocglobal.PackageTable)

	// collect all packages with a given prefix
	if *f.recursiveFrom != "" {
		processed := make(map[string]struct{})
		toProcess := pkgs
		for len(toProcess) > 0 {
			st, err := globalTable.Lookup(toProcess[0])
			if err != nil {
				return err
			}

			for _, p := range st.(*tables.Table).Imports {
				if !strings.HasPrefix(p, *f.recursiveFrom) {
					continue
				}
				if _, ok := processed[p]; ok {
					continue
				}
				processed[p] = struct{}{}
				toProcess = append(toProcess, p)
			}
			toProcess = toProcess[1:]
		}

		for p := range processed {
			tt, err := allocTable.LookupPackage(p)
			if err != nil {
				return err
			}
			if !tt.FilterOut("github.com/coreos/etcd") {
				packageAllocTables[p] = tt
				// Get the dynamic allocations
				ct, err := contractTable.Lookup(p)
				if err != nil {
					return err
				}
				r := runner.New(p, globalTable, allocTable, ct)
				if err := r.Run(); err != nil {
					return fmt.Errorf("Unable to evaluate contracts for %v: %v", p, err)
				}
			}
		}
	} else {
		for _, pkg := range pkgs {
			// Get the dynamic allocations
			ct, err := contractTable.Lookup(pkg)
			if err != nil {
				return err
			}
			r := runner.New(pkg, globalTable, allocTable, ct)
			if err := r.Run(); err != nil {
				return fmt.Errorf("Unable to evaluate contracts for %v: %v", pkg, err)
			}
			if perfile {
				tt, err := allocTable.LookupPackage(pkg)
				if err != nil {
					return err
				}
				packageAllocTables[pkg] = tt
			} else {
				if _, err := allocTable.LookupPackage(pkg); err != nil {
					return err
				}
				tt, err := allocTable.MergeFiles(pkg)
				if err != nil {
					return err
				}
				packageAllocTables[pkg] = tt
			}
		}
	}

	// Collect dynamic allocations

	if tojson {
		byteSlice, err := json.Marshal(packageAllocTables)
		if err != nil {
			return fmt.Errorf("Unable to convert print json: %v", err)
		}
		fmt.Printf("%v\n", string(byteSlice))
		return nil
	}

	if perfile {
		for _, pkg := range pkgs {
			allocTable.Print(pkg, allallocated)
		}
	} else {
		for p, files := range packageAllocTables {
			for file, t := range files {
				fmt.Printf("file: %v\n", path.Join(p, file))
				t.Print(allallocated)
			}
		}
	}
	return nil
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
		packagePrefix:   flag.String("package-prefix", "", "Package import path prefix in a PACKAGE:COMMIT form"),
		symbolTablePath: flag.String("symbol-table-dir", "", "Directory with preprocessed symbol tables"),
		// TODO(jchaloup): extend it with a hiearchy of cgo symbol files
		cgoSymbolsPath: flag.String("cgo-symbols-path", "", "Symbol table with CGO symbols (per entire project space)"),
		stdlib:         flag.Bool("stdlib", false, "Parse system Go std library"),
		allocated:      flag.Bool("allocated", false, "Extract allocation of symbols"),
		recursiveFrom:  flag.String("recursive-from", "", "Extract allocation of symbols from all prefixed paths"),
		filterPrefix:   flag.String("filter-prefix", "", "Filter out all imported packages from allocated symbols that does not match the prefix"),
		perfile:        flag.Bool("per-file", false, "Display allocated symbols per file"),
		allallocated:   flag.Bool("all-allocated", false, "Display all allocated symbols"),
		tojson:         flag.Bool("json", false, "Display allocated symbols in JSON"),
		glidefile:      flag.String("glidefile", "", "Glide.lock with dependencies"),
		godepsfile:     flag.String("godepsfile", "", "Godeps.json with dependencies"),
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
			p, err := parser.New(generatedDir, *(f.cgoSymbolsPath), goversion, nil)
			if err != nil {
				panic(err)
			}
			if err := p.Parse(pkg, false); err != nil {
				glog.Fatalf("Parse error when parsing (%v): %v", pkg, err)
			}
		}
		return
	}

	var snapshot snapshots.Snapshot
	if *f.glidefile != "" {
		sn, err := glide.GlideFromFile(*f.glidefile)
		if err != nil {
			panic(err)
		}
		if *f.packagePrefix != "" {
			parts := strings.Split(*f.packagePrefix, ":")
			if len(parts) != 2 {
				panic(fmt.Errorf("Expected --package-prefix in a PACKAGE:COMMIT form"))
			}
			sn.MainPackageCommit(parts[0], parts[1])
		}
		snapshot = sn
	} else if *f.godepsfile != "" {
		sn, err := godeps.FromFile(*f.godepsfile)
		if err != nil {
			panic(err)
		}
		if *f.packagePrefix != "" {
			parts := strings.Split(*f.packagePrefix, ":")
			if len(parts) != 2 {
				panic(fmt.Errorf("Expected --package-prefix in a PACKAGE:COMMIT form"))
			}
			sn.MainPackageCommit(parts[0], parts[1])
		}
		snapshot = sn
	}

	p, err := parser.New(*(f.symbolTablePath), *(f.cgoSymbolsPath), goversion, snapshot)
	if err != nil {
		panic(err)
	}
	for _, pkgPath := range strings.Split(*f.packagePath, ":") {
		if err := p.Parse(pkgPath, *(f.allocated)); err != nil {
			glog.Fatalf("Parse error: %v", err)
		}
	}

	if *(f.allocated) {
		if err := printPackageAllocTables(p.GlobalAllocTable(), p.GlobalSymbolTable(), p.GlobalContractsTable(), f); err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	}
}
