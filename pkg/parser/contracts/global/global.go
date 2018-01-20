package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	contracttable "github.com/gofed/symbols-extractor/pkg/parser/contracts/table"
	"github.com/gofed/symbols-extractor/pkg/snapshots"
	"github.com/golang/glog"
)

type PackageTable map[string]*contracttable.Table

type Table struct {
	tables         PackageTable
	symbolTableDir string
	goVersion      string
	glide          snapshots.Snapshot
}

func New(symbolTableDir, goVersion string, snapshot snapshots.Snapshot) *Table {
	return &Table{
		tables:         make(PackageTable),
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

func (t *Table) Add(packagePath string, table *contracttable.Table) {
	t.tables[packagePath] = table
}

func (t *Table) Packages() []string {
	var packages []string
	for key := range t.tables {
		packages = append(packages, key)
	}
	return packages
}

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

	file := path.Join(packagePath, "contracts.json")
	if _, err := os.Stat(file); err == nil {
		return nil
	}

	byteSlice, err := json.Marshal(*table)
	if err != nil {
		return fmt.Errorf("Unable to save %q symbol table: %v", pkg, err)
	}

	if err := ioutil.WriteFile(file, byteSlice, 0644); err != nil {
		return err
	}

	return nil
}

func (t *Table) Load(pkg string) (*contracttable.Table, error) {
	if t.symbolTableDir == "" {
		return nil, fmt.Errorf("Unable to load %q, symbol table dir not set", pkg)
	}

	file := path.Join(t.getPackagePath(pkg), "contracts.json")
	glog.V(2).Infof("Global contract table %q loading", file)

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		glog.V(2).Infof("Global contracts table %q loading failed: %v", file, err)
		return nil, fmt.Errorf("Unable to load %q contracts table from %q: %v", pkg, file, err)
	}

	var cTable contracttable.Table
	if err := json.Unmarshal(raw, &cTable); err != nil {
		return nil, fmt.Errorf("Unable to load %q symbol table from %q: %v", pkg, file, err)
	}

	return &cTable, nil
}

func (t *Table) Exists(pkg string) bool {
	_, ok := t.tables[pkg]
	if ok {
		return true
	}

	if _, err := os.Stat(path.Join(t.getPackagePath(pkg), "contracts.json")); err == nil {
		return true
	}

	return false
}

func (t *Table) Lookup(pkg string) (*contracttable.Table, error) {
	// the table must have at least one file processed
	if table, ok := t.tables[pkg]; ok {
		glog.V(2).Infof("Package-level contracts table %q found", pkg)
		return table, nil
	}

	// load the package on-demand
	table, err := t.Load(pkg)
	if err != nil {
		return nil, err
	}

	t.tables[pkg] = table
	glog.V(2).Infof("Package-level contracts table %q loaded", pkg)

	return table, nil
}

func (t *Table) Drop(pkg string) {
	delete(t.tables, pkg)
}
