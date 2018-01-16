package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables"
	"github.com/golang/glog"
)

type Table struct {
	symbolTableDir string
	goVersion      string
	tables         map[string]symbols.SymbolTable
}

func (t *Table) loadFromFile(pkg string) (symbols.SymbolTable, error) {
	if t.symbolTableDir == "" {
		return nil, fmt.Errorf("Unable to load %q, symbol table dir not set", pkg)
	}

	// check if the symbol table is available locally
	packagePath := path.Join(t.symbolTableDir, "golang", t.goVersion, pkg)
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		packagePath = path.Join(t.symbolTableDir, pkg)
	}

	file := path.Join(packagePath, "api.json")
	glog.Infof("Global symbol table %q loading", file)

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		glog.Infof("Global symbol table %q loading failed: %v", file, err)
		return nil, fmt.Errorf("Unable to load %q symbol table from %q: %v", pkg, file, err)
	}

	var table tables.Table
	if err := json.Unmarshal(raw, &table); err != nil {
		return nil, fmt.Errorf("Unable to load %q symbol table from %q: %v", pkg, file, err)
	}
	return &table, nil
}

func (t *Table) Lookup(pkg string) (symbols.SymbolTable, error) {
	if table, ok := t.tables[pkg]; ok {
		glog.Infof("Global symbol table %q found", pkg)
		return table, nil
	}

	// load the package on-demand
	table, err := t.loadFromFile(pkg)
	if err != nil {
		return nil, err
	}

	t.tables[pkg] = table
	glog.Infof("Global symbol table %q loaded", pkg)

	return table, nil
}

func (t *Table) Exists(pkg string) bool {
	_, ok := t.tables[pkg]
	return ok
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

func (t *Table) Packages() []string {
	var keys []string
	for key, _ := range t.tables {
		keys = append(keys, key)
	}
	return keys
}

func New(symbolTableDir, goVersion string) *Table {
	return &Table{
		symbolTableDir: symbolTableDir,
		goVersion:      goVersion,
		tables:         make(map[string]symbols.SymbolTable, 0),
	}
}

func (t *Table) store(pkg string, table *tables.Table) error {
	packagePath := path.Join(t.symbolTableDir, pkg)
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

func (t *Table) Load(symboltabledir string) error {
	t.tables = make(map[string]symbols.SymbolTable, 0)

	files, err := ioutil.ReadDir(symboltabledir)
	if err != nil {
		return err
	}
	for _, f := range files {
		file := path.Join(symboltabledir, f.Name())
		glog.Infof("Loading %q", file)
		raw, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		parts := strings.Split(f.Name(), ".")
		name := strings.Join(parts[:len(parts)-1], "/")
		var table tables.Table
		if err := json.Unmarshal(raw, &table); err != nil {
			return nil
		}
		t.tables[name] = &table
		fmt.Printf("Symbol table %q loaded\n", name)
	}

	return nil
}
