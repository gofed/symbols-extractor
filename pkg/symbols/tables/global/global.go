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
	tables         map[string]symbols.SymbolTable
}

func (t *Table) loadFromFile(pkg string) (symbols.SymbolTable, error) {
	if t.symbolTableDir == "" {
		return nil, fmt.Errorf("Symbol table dir not set")
	}
	// check if the symbol table is available locally
	parts := strings.Split(pkg, "/")
	parts = append(parts, "json")
	filename := strings.Join(parts, ".")
	file := path.Join(t.symbolTableDir, filename)
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

func (t *Table) Add(pkg string, st symbols.SymbolTable) error {
	if _, ok := t.tables[pkg]; ok {
		return fmt.Errorf("Symbol table for %q already exist in the global symbol table", pkg)
	}

	t.tables[pkg] = st
	return nil
}

func (t *Table) Packages() []string {
	var keys []string
	for key, _ := range t.tables {
		keys = append(keys, key)
	}
	return keys
}

func New(symbolTableDir string) *Table {
	return &Table{
		symbolTableDir: symbolTableDir,
		tables:         make(map[string]symbols.SymbolTable, 0),
	}
}

func (t *Table) Save(symboltabledir string) error {
	for key, symbolTable := range t.tables {
		file := path.Join(symboltabledir, fmt.Sprintf("%v.json", strings.Replace(key, "/", ".", -1)))
		if _, err := os.Stat(file); err == nil {
			continue
		}
		st, ok := symbolTable.(*tables.Table)
		if !ok {
			continue
		}
		byteSlice, err := json.Marshal(st)
		if err != nil {
			return fmt.Errorf("Unable to save %q symbol table: %v", key, err)
		}

		if err := ioutil.WriteFile(file, byteSlice, 0644); err != nil {
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
