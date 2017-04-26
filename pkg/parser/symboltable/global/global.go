package global

import (
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
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
