package main

import (
	"encoding/json"
	goflag "flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strings"

	util "github.com/gofed/symbols-extractor/cmd/go"
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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type SymbolsExtractorExtractCommand struct {
	// Go package entry point
	packagePath string
	// Import path prefix for the entry point
	packagePrefix string
	// location of already extracted symbols (used for storing extracted symbols as well)
	symbolTablePath string
	// location of symbols imported from the "C" package
	cgoSymbolsPath string

	stdlib        bool
	allocated     bool
	recursiveFrom string
	filterPrefix  string
	perfile       bool
	pertree       bool
	allallocated  bool
	tojson        bool
	glidefile     string
	godepsfile    string
	// Interpret entry point as a library instead of a reachability tree
	library bool
}

func (command *SymbolsExtractorExtractCommand) Run() error {
	if !command.stdlib && command.packagePath == "" {
		return fmt.Errorf("--package-path is not set")
	}

	// Otherwise it can eat all the CPU power
	runtime.GOMAXPROCS(1)

	output, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running `go version`: %v", err)
	}
	// TODO(jchaloup): Check the version is of the form d.d for now (later extend with alpha/beta/rc...)
	goversion := strings.Split(string(output), " ")[2][2:]

	// parse the standard library
	if command.stdlib {
		processStdlib(command.symbolTablePath, command.cgoSymbolsPath, goversion)
		return nil
	}

	snapshot, err := buildSnapshot(command.glidefile, command.godepsfile, command.packagePrefix)
	if err != nil {
		return err
	}

	p, err := parser.New(command.symbolTablePath, command.cgoSymbolsPath, goversion, snapshot)
	if err != nil {
		return err
	}

	entryPoints, _ := buildEntryPoints(command.packagePath, command.library)

	for _, pkgPath := range entryPoints {
		if err := p.Parse(pkgPath, command.allocated); err != nil {
			return fmt.Errorf("Parse error: %v", err)
		}
	}

	if command.allocated {
		if err := printPackageAllocTables(p.GlobalAllocTable(), p.GlobalSymbolTable(), p.GlobalContractsTable(), entryPoints, &cmdFlags); err != nil {
			return err
		}
	}

	return nil
}

var cmdFlags = SymbolsExtractorExtractCommand{}

func NewSymbolsExtractorExtractCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "symbols-extract",
		Short: "Go symbols extractor",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmdFlags.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&cmdFlags.packagePath, "package-path", cmdFlags.packagePath, "Package entry point")
	flags.StringVar(&cmdFlags.packagePrefix, "package-prefix", cmdFlags.packagePrefix, "Package import path prefix in a PACKAGE:COMMIT form")
	flags.StringVar(&cmdFlags.symbolTablePath, "symbol-table-dir", cmdFlags.symbolTablePath, "Directory with preprocessed symbol tables")
	// TODO(jchaloup): extend it with a hiearchy of cgo symbol files
	flags.StringVar(&cmdFlags.cgoSymbolsPath, "cgo-symbols-path", cmdFlags.cgoSymbolsPath, "Symbol table with CGO symbols (per entire project space)")
	flags.BoolVar(&cmdFlags.stdlib, "stdlib", cmdFlags.stdlib, "Parse system Go std library")
	flags.BoolVar(&cmdFlags.allocated, "allocated", cmdFlags.allocated, "Extract allocation of symbols")
	flags.StringVar(&cmdFlags.recursiveFrom, "recursive-from", cmdFlags.recursiveFrom, "Extract allocation of symbols from all prefixed paths")
	flags.StringVar(&cmdFlags.filterPrefix, "filter-prefix", cmdFlags.filterPrefix, "Filter out all imported packages from allocated symbols that does not match the prefix")
	flags.BoolVar(&cmdFlags.perfile, "per-file", cmdFlags.perfile, "Display allocated symbols per file")
	flags.BoolVar(&cmdFlags.pertree, "per-tree", cmdFlags.pertree, "Display allocated symbols per entire tree")
	flags.BoolVar(&cmdFlags.allallocated, "all-allocated", cmdFlags.allallocated, "Display all allocated symbols")
	flags.BoolVar(&cmdFlags.tojson, "json", cmdFlags.tojson, "Display allocated symbols in JSON")
	flags.StringVar(&cmdFlags.glidefile, "glidefile", cmdFlags.glidefile, "Glide.lock with dependencies")
	flags.StringVar(&cmdFlags.godepsfile, "godepsfile", cmdFlags.godepsfile, "Godeps.json with dependencies")
	flags.BoolVar(&cmdFlags.library, "library", cmdFlags.library, "Interpret package entry point as a library")

	return cmd
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

func printPackageAllocTables(allocTable *allocglobal.Table, globalTable *global.Table, contractTable *contractglobal.Table, entryPoints []string, cmdFlags *SymbolsExtractorExtractCommand) error {

	perfile := cmdFlags.perfile
	allallocated := cmdFlags.allallocated
	tojson := cmdFlags.tojson

	packageAllocTables := make(map[string]allocglobal.PackageTable)

	// collect all packages with a given prefix reachable from all the entry points
	if cmdFlags.recursiveFrom != "" {
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
			if !strings.HasPrefix(p, cmdFlags.recursiveFrom) {
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

			if !tt.FilterOut(cmdFlags.filterPrefix) {
				packageAllocTables[p] = tt
			}
		}

		// if per-project is set, merge all tables into one
		if cmdFlags.pertree {
			cpTable := allocglobal.NewPackageTable()
			for p, tt := range packageAllocTables {
				cpTable.MergeWith(&tt)
				delete(packageAllocTables, p)
			}
			packageAllocTables[cmdFlags.recursiveFrom] = *cpTable
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

func buildEntryPoints(packagePath string, library bool) ([]string, error) {
	if library {
		var entryPoints []string

		for _, ep := range strings.Split(packagePath, ":") {
			c := util.NewPackageInfoCollector(nil, nil)
			if err := c.CollectPackageInfos(ep); err != nil {
				return nil, err
			}

			pkgs, err := c.BuildPackageTree(true, false)
			if err != nil {
				return nil, err
			}
			entryPoints = append(entryPoints, pkgs...)
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

	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	command := NewSymbolsExtractorExtractCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
