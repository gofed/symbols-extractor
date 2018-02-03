package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/snapshots"
	"github.com/golang/glog"
)

// Table captures list of allocated symbols for each package and its files
type Table struct {
	tables         map[string]PackageTable
	symbolTableDir string
	goVersion      string
	glide          snapshots.Snapshot
}

func New(symbolTableDir, goVersion string, snapshot snapshots.Snapshot) *Table {
	return &Table{
		tables:         make(map[string]PackageTable, 0),
		symbolTableDir: symbolTableDir,
		goVersion:      goVersion,
		glide:          snapshot,
	}
}

func (t *Table) getPackagePath(pkg string) string {
	packagePath := path.Join(t.symbolTableDir, "golang", t.goVersion, pkg)
	if _, err := os.Stat(packagePath); err == nil {
		return packagePath
	}

	packagePath = path.Join(t.symbolTableDir, pkg)
	if t.glide != nil {
		if commit, err := t.glide.Commit(pkg); err == nil {
			return path.Join(packagePath, commit)
		}
	}

	return packagePath
}

func (t *Table) Add(packagePath, file string, table *alloctable.Table) {
	if _, ok := t.tables[packagePath]; !ok {
		t.tables[packagePath] = *NewPackageTable()
	}

	if _, ok := t.tables[packagePath][file]; !ok {
		t.tables[packagePath][file] = table
	}
}

func (t *Table) Packages() []string {
	var packages []string
	for key, _ := range t.tables {
		packages = append(packages, key)
	}
	return packages
}

func (t *Table) Print(packagePath string, all bool) {
	filesList := t.Files(packagePath)
	sort.Strings(filesList)
	for _, f := range filesList {
		fmt.Printf("\tFile: %v\n", f)
		at, err := t.Lookup(packagePath, f)
		if err != nil {
			panic(err)
		}
		at.Print(all)
	}
}

func (t *Table) Files(packagePath string) []string {
	var files []string
	fileTable, ok := t.tables[packagePath]
	if !ok {
		return nil
	}
	for key, _ := range fileTable {
		files = append(files, key)
	}
	return files
}

// MergeFiles merges all per-file allocated tables into one
func (t *Table) MergeFiles(packagePath string) (PackageTable, error) {
	maTable := alloctable.New(packagePath, "")
	table, ok := t.tables[packagePath]
	if !ok {
		return nil, fmt.Errorf("Unable to find %q package", packagePath)
	}
	for _, atable := range table {
		for pkg, symbolSets := range atable.Symbols {
			for _, item := range symbolSets.Datatypes {
				maTable.AddDataType(pkg, item.Name, item.Pos)
			}
			for _, item := range symbolSets.Functions {
				maTable.AddFunction(pkg, item.Name, item.Pos)
			}
			for _, item := range symbolSets.Variables {
				maTable.AddVariable(pkg, item.Name, item.Pos)
			}
			for _, item := range symbolSets.Methods {
				maTable.AddMethod(pkg, item.Parent, item.Name, item.Pos)
			}
			for _, item := range symbolSets.Structfields {
				maTable.AddStructField(pkg, item.Parent, item.Field, item.Pos)
			}
		}
	}
	return PackageTable{
		"": maTable,
	}, nil
}

func (t *Table) Lookup(packagePath, file string) (*alloctable.Table, error) {
	files, ok := t.tables[packagePath]
	if !ok {
		table, err := t.LookupPackage(packagePath)
		if err != nil {
			return nil, fmt.Errorf("Unable to find package-level allocated symbol table for package %q", packagePath)
		}
		files = table
	}
	table, exists := files[file]
	if !exists {
		return nil, fmt.Errorf("Unable to find file-level allocated symbol table for package %q and file %q", packagePath, file)
	}
	return table, nil
}

func (t *Table) Exists(pkg string) bool {
	_, ok := t.tables[pkg]
	if ok {
		return true
	}

	if _, err := os.Stat(path.Join(t.getPackagePath(pkg), "allocated.json")); err == nil {
		return true
	}

	return false
}
func (t *Table) LookupPackage(pkg string) (PackageTable, error) {
	// the table must have at least one file processed
	if table, ok := t.tables[pkg]; ok && len(table) > 0 {
		glog.V(2).Infof("Package-level allocated symbol table %q found: %#v", pkg, table)
		return table, nil
	}

	// load the package on-demand
	table, err := t.loadFromFile(pkg)
	if err != nil {
		return nil, err
	}

	t.tables[pkg] = table
	glog.V(2).Infof("Package-level allocated symbol table %q loaded", pkg)

	return table, nil
}

// Store allocation for a given package
func (t *Table) Save(pkg string) error {
	table, ok := t.tables[pkg]
	if !ok {
		return fmt.Errorf("Allocated table for %q does not exist", pkg)
	}

	packagePath := t.getPackagePath(pkg)
	pErr := os.MkdirAll(packagePath, 0777)
	if pErr != nil {
		return fmt.Errorf("Unable to create package path %v: %v", packagePath, pErr)
	}

	file := path.Join(packagePath, "allocated.json")
	if _, err := os.Stat(file); err == nil {
		return nil
	}

	byteSlice, err := json.Marshal(table)
	if err != nil {
		return fmt.Errorf("Unable to save %q symbol table: %v", pkg, err)
	}

	if err := ioutil.WriteFile(file, byteSlice, 0644); err != nil {
		return err
	}

	return nil
}

func (t *Table) loadFromFile(pkg string) (PackageTable, error) {
	if t.symbolTableDir == "" {
		return nil, fmt.Errorf("Unable to load %q, symbol table dir not set", pkg)
	}

	// check if the symbol table is available locally
	packagePath := t.getPackagePath(pkg)

	file := path.Join(packagePath, "allocated.json")
	glog.V(2).Infof("Global symbol table %q loading", file)

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		glog.V(2).Infof("Global symbol table %q loading failed: %v", file, err)
		return nil, fmt.Errorf("Unable to load %q symbol table from %q: %v", pkg, file, err)
	}

	var table PackageTable
	if err := json.Unmarshal(raw, &table); err != nil {
		return nil, fmt.Errorf("Unable to load %q symbol table from %q: %v", pkg, file, err)
	}

	return table, nil
}

func (t *Table) Drop(pkg string) {
	delete(t.tables, pkg)
}
