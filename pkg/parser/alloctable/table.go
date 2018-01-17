package alloctable

import (
	"fmt"
	"sort"
	"strings"
)

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

type Variable struct {
	Name string `json:"name"`
	Pos  string `json:"pos"`
}

type Method struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
	Pos    string `json:"pos"`
}

type StructField struct {
	Parent string `json:"parent"`
	Field  string `json:"field"`
	//Chain  []*symbols.SymbolDef `json:"chain"`
	Pos string `json:"pos"`
}

type Package struct {
	Datatypes    map[string]Datatype    `json:"datatypes"`
	Functions    map[string]Function    `json:"functions"`
	Variables    map[string]Variable    `json:"variables"`
	Methods      []Method               `json:"methods"`
	Structfields map[string]StructField `json:"structfields"`
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

func newPackage() *Package {
	return &Package{
		Functions:    make(map[string]Function),
		Datatypes:    make(map[string]Datatype),
		Variables:    make(map[string]Variable),
		Structfields: make(map[string]StructField),
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
		ast.Symbols[pkg] = newPackage()
		items = ast.Symbols[pkg]
	}

	k := toKey(name, pos)
	if _, ok := items.Functions[k]; ok {
		return
	}

	items.Datatypes[k] = Datatype{
		Name: name,
		Pos:  pos,
	}
}

func (ast *Table) AddVariable(pkg, name, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = newPackage()
		items = ast.Symbols[pkg]
	}

	k := toKey(name, pos)
	if _, ok := items.Variables[k]; ok {
		return
	}

	items.Variables[k] = Variable{
		Name: name,
		Pos:  pos,
	}
}

func toKey(name, pos string) string {
	return fmt.Sprintf("%v:%v", name, pos)
}

func (ast *Table) AddFunction(pkg, name, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = newPackage()
		items = ast.Symbols[pkg]
	}

	k := toKey(name, pos)
	if _, ok := items.Functions[k]; ok {
		return
	}

	items.Functions[k] = Function{
		Name: name,
		Pos:  pos,
	}
}

func (ast *Table) AddStructField(pkg, parent, field string, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = newPackage()
		items = ast.Symbols[pkg]
	}

	k := toKey(field, pos)
	if _, ok := items.Structfields[k]; ok {
		return
	}

	items.Structfields[k] = StructField{
		Parent: parent,
		Field:  field,
		Pos:    pos,
	}
}

func (ast *Table) AddMethod(pkg, parent, name, pos string) {
	items, exists := ast.Symbols[pkg]
	if !exists {
		ast.Symbols[pkg] = newPackage()
		items = ast.Symbols[pkg]
	}
	items.Methods = append(items.Methods, Method{
		Parent: parent,
		Name:   name,
		Pos:    pos,
	})
}

func (ast *Table) AddDataTypeField(origin, dataType, field, pos string) {

	ast.AddSymbol(origin, fmt.Sprintf("%v.%v", dataType, field), pos)
}

func (ast *Table) AddSymbol(origin, id, pos string) {

}

func (ast *Table) Print() {
	fmt.Printf("======================================================================================================\n")
	symPos := make(map[string][]string)
	maxKeyLen := 0
	var keys []string
	for key := range ast.Symbols {
		// agregate positions by data type
		for _, dt := range ast.Symbols[key].Datatypes {
			qidid := fmt.Sprintf("D: %v.%v", key, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range ast.Symbols[key].Functions {
			qidid := fmt.Sprintf("F: %v.%v", key, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range ast.Symbols[key].Methods {
			qidid := fmt.Sprintf("M: %v.%v.%v", key, dt.Parent, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range ast.Symbols[key].Variables {
			qidid := fmt.Sprintf("V: %v.%v", key, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range ast.Symbols[key].Structfields {
			qidid := fmt.Sprintf("S: %v.%v.%v", key, dt.Parent, dt.Field)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
	}
	sort.Strings(keys)
	for _, qidid := range keys {
		fmt.Printf("\t%v:%v\t%v\n", qidid, strings.Repeat(" ", (maxKeyLen-len(qidid))), len(symPos[qidid]))
	}
	fmt.Printf("======================================================================================================\n")
}
