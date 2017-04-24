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

func New() *Table {
	return &Table{
		tables: make(map[string]*symboltable.Table, 0),
	}
}
