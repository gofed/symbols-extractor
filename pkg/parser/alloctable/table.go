package alloctable

import (
	"fmt"

	"github.com/golang/glog"
)

// - count a number of each symbol used (to watch how intensively is a given symbol used)
// - each AS table is per file, AS package is a union of file AS tables
//

type Table struct {
	File    string `json:"file"`
	Package string `json:"package"`
	// symbol's name is in a PACKAGE.ID form
	// if the PACKAGE is empty, the ID is considired as the embedded symbol
	Symbols map[string]int `json:"symbols"`
	locked  bool
}

func New() *Table {
	return &Table{
		Symbols: make(map[string]int),
	}
}

func (ast *Table) Lock() {
	ast.locked = true
}

func (ast *Table) Unlock() {
	ast.locked = false
}

func (ast *Table) AddDataTypeField(origin, dataType, field string) {
	ast.AddSymbol(origin, fmt.Sprintf("%v.%v", dataType, field))
}

func (ast *Table) AddSymbol(origin, id string) {
	if ast.locked {
		return
	}
	glog.Infof("Adding symbol into alloc table: origin=%q\tid=%q", origin, id)
	var key string
	if origin != "" {
		key = fmt.Sprintf("%v.%v", origin, id)
	} else {
		key = id
	}

	count, ok := ast.Symbols[key]
	if !ok {
		ast.Symbols[key] = 1
	} else {
		ast.Symbols[key] = count + 1
	}
}

func (ast *Table) Print() {
	glog.Infof("======================================================================================================\n")
	for key := range ast.Symbols {
		glog.Infof("%v:\t%v\n", key, ast.Symbols[key])
	}
	glog.Infof("======================================================================================================\n")
}
