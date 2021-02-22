package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gofed/symbols-extractor/pkg/snapshots"
	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables"
	"k8s.io/klog/v2"
)

type Table struct {
	symbolTableDir string
	goVersion      string
	glide          snapshots.Snapshot
	tables         map[string]symbols.SymbolTable
	fromFile       map[string]struct{}
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

func (t *Table) loadFromFile(pkg string) (symbols.SymbolTable, error) {
	if t.symbolTableDir == "" {
		return nil, fmt.Errorf("Unable to load %q, symbol table dir not set", pkg)
	}

	file := path.Join(t.getPackagePath(pkg), "api.json")
	klog.V(2).Infof("Global symbol table %q loading", file)

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		klog.V(2).Infof("Global symbol table %q loading failed: %v", file, err)
		return nil, fmt.Errorf("Unable to load %q symbol table from %q: %v", pkg, file, err)
	}

	var table tables.Table
	if err := json.Unmarshal(raw, &table); err != nil {
		return nil, fmt.Errorf("Unable to load %q symbol table from %q: %v", pkg, file, err)
	}

	t.fromFile[pkg] = struct{}{}
	return &table, nil
}

func (t *Table) Lookup(pkg string) (symbols.SymbolTable, error) {
	if table, ok := t.tables[pkg]; ok {
		klog.V(2).Infof("Global symbol table %q found", pkg)
		return table, nil
	}

	// load the package on-demand
	table, err := t.loadFromFile(pkg)
	if err != nil {
		return nil, err
	}

	t.tables[pkg] = table
	klog.V(2).Infof("Global symbol table %q loaded", pkg)

	return table, nil
}

func (t *Table) Exists(pkg string) bool {
	_, ok := t.tables[pkg]
	if ok {
		return true
	}

	// check if the symbol table is available locally
	if _, err := os.Stat(path.Join(t.getPackagePath(pkg), "api.json")); err == nil {
		return true
	}

	return false
}

func (t *Table) Add(pkg string, table symbols.SymbolTable, store bool) error {
	if _, ok := t.tables[pkg]; ok {
		return fmt.Errorf("Symbol table for %q already exist in the global symbol table", pkg)
	}

	t.tables[pkg] = table

	if store {
		st, ok := table.(*tables.Table)
		if !ok {
			return nil
		}
		t.store(pkg, st)
	}
	return nil
}

func (t *Table) Drop(pkg string) {
	delete(t.tables, pkg)
}

func (t *Table) Packages() []string {
	var keys []string
	for key, _ := range t.tables {
		keys = append(keys, key)
	}
	return keys
}

func New(symbolTableDir, goVersion string, snapshot snapshots.Snapshot) *Table {
	return &Table{
		symbolTableDir: symbolTableDir,
		goVersion:      goVersion,
		glide:          snapshot,
		tables:         make(map[string]symbols.SymbolTable, 0),
		fromFile:       make(map[string]struct{}, 0),
	}
}

func (t *Table) store(pkg string, table *tables.Table) error {
	// Don't save tables that ware loaded from a file
	if _, ok := t.fromFile[pkg]; ok {
		return nil
	}

	packagePath := t.getPackagePath(pkg)

	pErr := os.MkdirAll(packagePath, 0777)
	if pErr != nil {
		return fmt.Errorf("Unable to create package path %v: %v", packagePath, pErr)
	}

	file := path.Join(packagePath, "api.json")
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

func (t *Table) Save(symboltabledir string) error {
	// create the dir if it does not exist
	err := os.MkdirAll(symboltabledir, 0777)
	if err != nil {
		return fmt.Errorf("Unable to create directory path %v: %v", symboltabledir, err)
	}

	for key, symbolTable := range t.tables {
		st, ok := symbolTable.(*tables.Table)
		if !ok {
			continue
		}

		if err := t.store(key, st); err != nil {
			return err
		}
	}
	return nil
}
