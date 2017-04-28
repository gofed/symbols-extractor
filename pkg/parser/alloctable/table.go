package alloctable

import "fmt"

// - count a number of each symbol used (to watch how intensively is a given symbol used)
// - each AS table is per file, AS package is a union of file AS tables
//

type Table struct {
	File    string
	Package string
	// symbol's name is in a PACKAGE.ID form
	// if the PACKAGE is empty, the ID is considired as the embedded symbol
	symbols map[string]int
}

func New() *Table {
	return &Table{
		symbols: make(map[string]int),
	}
}

func (ast *Table) AddDataTypeField(origin, dataType, field string) {
	ast.AddSymbol(origin, fmt.Sprintf("%v.%v", dataType, field))
}

func (ast *Table) AddSymbol(origin, id string) {
	var key string
	if origin != "" {
		key = fmt.Sprintf("%v.%v", origin, id)
	} else {
		key = id
	}

	count, ok := ast.symbols[key]
	if !ok {
		ast.symbols[key] = 1
	} else {
		ast.symbols[key] = count + 1
	}
}

func (ast *Table) Print() {
	fmt.Printf("======================================================================================================\n")
	for key := range ast.symbols {
		fmt.Printf("%v:\t%v\n", key, ast.symbols[key])
	}
	fmt.Printf("======================================================================================================\n")
}
