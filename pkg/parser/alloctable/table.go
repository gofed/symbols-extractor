package alloctable

import "fmt"

// - count a number of each symbol used (to watch how intensively is a given symbol used)
// - each AS table is per file, AS package is a union of file AS tables
//
type Datatype struct {
	Name string `json:"name"`
	Pos  string `json:"pos"`
}

type Function struct {
	Name string `json:"name"`
	Pos  string `json:"pos"`
}

type Method struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
	Pos    string `json:"pos"`
}

type StructField struct {
	Parent string   `json:"parent"`
	Field  string   `json:"field"`
	Chain  []string `json:"chain"`
	Pos    string   `json:"pos"`
}

type Package struct {
	Datatypes    []Datatype    `json:"datatypes"`
	Functions    []Function    `json:"functions"`
	Methods      []Method      `json:"methods"`
	Structfields []StructField `json:"structfields"`
}

type Table struct {
	File    string              `json:"file"`
	Package string              `json:"package"`
	Symbols map[string]*Package `json:"symbols"`
	locked  bool
}

func New() *Table {
	return &Table{
		Symbols: make(map[string]*Package),
	}
}

func (ast *Table) Lock() {
	ast.locked = true
}

func (ast *Table) Unlock() {
	ast.locked = false
}

func (ast *Table) AddDataType(pkg, name, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = &Package{}
		items = ast.Symbols[pkg]
	}
	items.Datatypes = append(items.Datatypes, Datatype{
		Name: name,
		Pos:  pos,
	})
}

func (ast *Table) AddFunction(pkg, name, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = &Package{}
		items = ast.Symbols[pkg]
	}
	items.Functions = append(items.Functions, Function{
		Name: name,
		Pos:  pos,
	})
}

func (ast *Table) AddStructField(pkg, parent, field string, chain []string, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = &Package{}
		items = ast.Symbols[pkg]
	}
	items.Structfields = append(items.Structfields, StructField{
		Parent: parent,
		Field:  field,
		Chain:  chain,
		Pos:    pos,
	})
}

func (ast *Table) AddMethod(pkg, parent, name, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = &Package{}
		items = ast.Symbols[pkg]
	}
	items.Methods = append(items.Methods, Method{
		Parent: parent,
		Name:   name,
		Pos:    pos,
	})
}

func (ast *Table) AddDataTypeField(origin, dataType, field string) {
	ast.AddSymbol(origin, fmt.Sprintf("%v.%v", dataType, field), "")
}

func (ast *Table) AddSymbol(origin, id, pos string) {

}

func (ast *Table) Print() {
	fmt.Printf("======================================================================================================\n")
	for key := range ast.Symbols {
		fmt.Printf("%v:\t%v\n", key, ast.Symbols[key])
	}
	fmt.Printf("======================================================================================================\n")
}
