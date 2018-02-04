package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	// Go package entry point
	packagePath *string
	// Import path prefix for the entry point
	packagePrefix *string
	// location of already extracted symbols (used for storing extracted symbols as well)
	symbolTablePath *string
	// location of symbols imported from the "C" package
	cgoSymbolsPath *string

	stdlib        *bool
	allocated     *bool
	recursiveFrom *string
	filterPrefix  *string
	perfile       *bool
	pertree       *bool
	allallocated  *bool
	tojson        *bool
	glidefile     *string
	godepsfile    *string
	// Interpret entry point as a library instead of a reachability tree
	library *bool
}

func buildFlags() *flags {
	return &flags{
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
		pertree:        flag.Bool("per-tree", false, "Display allocated symbols per entire tree"),
		allallocated:   flag.Bool("all-allocated", false, "Display all allocated symbols"),
		tojson:         flag.Bool("json", false, "Display allocated symbols in JSON"),
		glidefile:      flag.String("glidefile", "", "Glide.lock with dependencies"),
		godepsfile:     flag.String("godepsfile", "", "Godeps.json with dependencies"),
		library:        flag.Bool("library", false, "Interpret package entry point as a library"),
	}
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

func printPackageAllocTables(allocTable *allocglobal.Table, globalTable *global.Table, contractTable *contractglobal.Table, entryPoints []string, f *flags) error {

	perfile := *f.perfile
	allallocated := *f.allallocated
	tojson := *f.tojson

	packageAllocTables := make(map[string]allocglobal.PackageTable)

	// collect all packages with a given prefix reachable from all the entry points
	if *f.recursiveFrom != "" {
		processed := make(map[string]struct{})
		toProcess := entryPoints
		for len(toProcess) > 0 {
			if _, ok := processed[toProcess[0]]; ok {
				toProcess = toProcess[1:]
				continue
			}

			glog.V(1).Infof("Processing %v\n", toProcess[0])
			st, err := globalTable.Lookup(toProcess[0])
			if err != nil {
				return err
			}

			processed[toProcess[0]] = struct{}{}

			for _, p := range st.(*tables.Table).Imports {
				if _, ok := processed[p]; ok {
					continue
				}
				processed[p] = struct{}{}
				toProcess = append(toProcess, p)
			}
			toProcess = toProcess[1:]
		}

		for p := range processed {
			if !strings.HasPrefix(p, *f.recursiveFrom) {
				delete(processed, p)
				continue
			}

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

		for p := range processed {
			tt, err := allocTable.LookupPackage(p)
			if err != nil {
				return err
			}

			if !tt.FilterOut(*f.filterPrefix) {
				packageAllocTables[p] = tt
			}
		}

		// if per-project is set, merge all tables into one
		if *f.pertree {
			cpTable := allocglobal.NewPackageTable()
			for p, tt := range packageAllocTables {
				cpTable.MergeWith(&tt)
				delete(packageAllocTables, p)
			}
			packageAllocTables[*f.recursiveFrom] = *cpTable
		}

	} else {
		for _, pkg := range entryPoints {
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

	if tojson {
		byteSlice, err := json.Marshal(packageAllocTables)
		if err != nil {
			return fmt.Errorf("Unable to convert print json: %v", err)
		}
		fmt.Printf("%v\n", string(byteSlice))
		return nil
	}

	if perfile {
		for _, pkg := range entryPoints {
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

func processStdlib(symbolTablePath, cgoSymbolsPath, goversion string) {
	packages, err := getStdlibPackages()
	if err != nil {
		glog.Fatal(err)
	}

	generatedDir := path.Join(symbolTablePath, "golang", goversion)

	for _, pkg := range packages {
		fmt.Printf("Parsing %q...\n", pkg)
		p, err := parser.New(generatedDir, cgoSymbolsPath, goversion, nil)
		if err != nil {
			panic(err)
		}
		if err := p.Parse(pkg, false); err != nil {
			glog.Fatalf("Parse error when parsing (%v): %v", pkg, err)
		}
	}
	return
}

func buildSnapshot(glidefile, godepsfile, packagePrefix string) (snapshots.Snapshot, error) {
	if glidefile != "" {
		sn, err := glide.GlideFromFile(glidefile)
		if err != nil {
			return nil, err
		}
		if packagePrefix != "" {
			parts := strings.Split(packagePrefix, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("Expected --package-prefix in a PACKAGE:COMMIT form")
			}
			sn.MainPackageCommit(parts[0], parts[1])
		}
		return sn, nil
	}
	if godepsfile != "" {
		sn, err := godeps.FromFile(godepsfile)
		if err != nil {
			return nil, err
		}
		if packagePrefix != "" {
			parts := strings.Split(packagePrefix, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("Expected --package-prefix in a PACKAGE:COMMIT form")
			}
			sn.MainPackageCommit(parts[0], parts[1])
		}
		return sn, nil
	}
	panic("glidefile or godepsfile must be nonempty")
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func buildEntryPoints(packagePath string, library bool) ([]string, error) {
	if library {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			return nil, fmt.Errorf("GOPATH not set")
		}

		var abspath, pathPrefix string

		for _, ep := range strings.Split(packagePath, ":") {
			// first, find the absolute path
			for _, gpath := range strings.Split(gopath, ":") {
				abspath = path.Join(gpath, "src", ep)

				if e, err := exists(abspath); err == nil && e {
					pathPrefix = path.Join(gpath, "src")
					break
				}
			}
		}

		pkgs := make(map[string]struct{})

		visit := func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if strings.HasSuffix(path, ".go") {
					pkgs[filepath.Dir(path[len(pathPrefix)+1:])] = struct{}{}
				}
			}
			return nil
		}

		err := filepath.Walk(abspath+"/", visit)
		if err != nil {
			return nil, err
		}

		var entryPoints []string
		for p := range pkgs {
			entryPoints = append(entryPoints, p)
		}

		return entryPoints, nil
	}
	return strings.Split(packagePath, ":"), nil
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

	f := buildFlags()

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
		processStdlib(*f.symbolTablePath, *f.cgoSymbolsPath, goversion)
		return
	}

	snapshot, err := buildSnapshot(*f.glidefile, *f.godepsfile, *f.packagePrefix)
	if err != nil {
		glog.Fatal(err)
	}

	p, err := parser.New(*(f.symbolTablePath), *(f.cgoSymbolsPath), goversion, snapshot)
	if err != nil {
		glog.Fatal(err)
	}

	entryPoints, _ := buildEntryPoints(*f.packagePath, *f.library)

	for _, pkgPath := range entryPoints {
		if err := p.Parse(pkgPath, *(f.allocated)); err != nil {
			glog.Fatalf("Parse error: %v", err)
		}
	}

	if *(f.allocated) {
		if err := printPackageAllocTables(p.GlobalAllocTable(), p.GlobalSymbolTable(), p.GlobalContractsTable(), entryPoints, f); err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	}
}
