package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
)

type PackageTable map[string]*alloctable.Table

func NewPackageTable() *PackageTable {
	tt := make(PackageTable, 0)
	return &tt
}

func (t *PackageTable) Load(file string) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Unable to load symbol table from %q: %v", file, err)
	}

	var table PackageTable
	if err := json.Unmarshal(raw, &table); err != nil {
		return fmt.Errorf("Unable to load symbol table from %q: %v", file, err)
	}

	*t = table
	return nil
}

func (t *PackageTable) FilterOut(prefix string) bool {
	emptyTable := true
	for file, tab := range *t {
		empty := true
		for pkg := range tab.Symbols {
			if !strings.HasPrefix(pkg, prefix) {
				delete(tab.Symbols, pkg)
				continue
			}
			empty = false
		}
		if empty {
			delete(*t, file)
			continue
		}
		emptyTable = false
	}
	return emptyTable
}

// Consolidate merges all per-file allocated tables into one
func (t *PackageTable) Consolidate() {
	cTable := alloctable.New("", "")

	for file, table := range *t {
		cTable.MergeWith(table)
		delete(*t, file)
	}
	(*t)[""] = cTable
}

// Merge current table with passed one
func (t *PackageTable) MergeWith(pt *PackageTable) {
	t.Consolidate()

	for _, table := range *pt {
		(*t)[""].MergeWith(table)
	}
}
