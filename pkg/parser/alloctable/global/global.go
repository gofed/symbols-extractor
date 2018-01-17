package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/golang/glog"
)

type PackageTable map[string]*alloctable.Table

// Table captures list of allocated symbols for each package and its files
type Table struct {
	tables         map[string]PackageTable
	symbolTableDir string
	goVersion      string
}

func New(symbolTableDir, goVersion string) *Table {
	return &Table{
		tables:         make(map[string]PackageTable, 0),
		symbolTableDir: symbolTableDir,
		goVersion:      goVersion,
	}
}

func (t *Table) Add(packagePath, file string, table *alloctable.Table) {
	if _, ok := t.tables[packagePath]; !ok {
		t.tables[packagePath] = make(PackageTable, 0)
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

func (t *Table) Lookup(packagePath, file string) (*alloctable.Table, error) {
	files, ok := t.tables[packagePath]
	if !ok {
		return nil, fmt.Errorf("Unable to find package-level allocated symbol table for package %q", packagePath)
	}
	table, exists := files[file]
	if !exists {
		return nil, fmt.Errorf("Unable to find file-level allocated symbol table for package %q and file %q", packagePath, file)
	}
	return table, nil
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
func (t *Table) Store(pkg string) error {
	table, ok := t.tables[pkg]
	if !ok {
		return fmt.Errorf("Allocated table for %q does not exist", pkg)
	}

	packagePath := path.Join(t.symbolTableDir, pkg)
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
	packagePath := path.Join(t.symbolTableDir, "golang", t.goVersion, pkg)
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		packagePath = path.Join(t.symbolTableDir, pkg)
	}

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
