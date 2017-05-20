package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/golang/glog"
)

type Table struct {
	tables map[string]*symboltable.Table
}

func (t *Table) Lookup(pkg string) (*symboltable.Table, error) {
	table, ok := t.tables[pkg]
	if !ok {
		return nil, fmt.Errorf("Unable to find symbol table for %q", pkg)
	}
	return table, nil
}

func (t *Table) Exists(pkg string) bool {
	_, ok := t.tables[pkg]
	return ok
}

func (t *Table) Add(pkg string, st *symboltable.Table) error {
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

func New() *Table {
	return &Table{
		tables: make(map[string]*symboltable.Table, 0),
	}
}

func (t *Table) Save(symboltabledir string) error {
	for key, symbolTable := range t.tables {
		file := path.Join(symboltabledir, fmt.Sprintf("%v.json", strings.Replace(key, "/", ".", -1)))
		if _, err := os.Stat(file); err == nil {
			continue
		}
		byteSlice, err := json.Marshal(symbolTable)
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
	t.tables = make(map[string]*symboltable.Table, 0)

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
		var table symboltable.Table
		if err := json.Unmarshal(raw, &table); err != nil {
			return nil
		}
		t.tables[name] = &table
		fmt.Printf("Symbol table %q loaded\n", name)
	}

	return nil
}
