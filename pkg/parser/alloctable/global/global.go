package global

import (
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
)

// Table captures list of allocated symbols for each package and its files
type Table struct {
	tables map[string]map[string]*alloctable.Table
}

func New() *Table {
	return &Table{
		tables: make(map[string]map[string]*alloctable.Table, 0),
	}
}

func (t *Table) Add(packagePath, file string, table *alloctable.Table) {
	if _, ok := t.tables[packagePath]; !ok {
		t.tables[packagePath] = make(map[string]*alloctable.Table, 0)
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
