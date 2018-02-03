package alloctable

import (
	"fmt"
	"go/ast"
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

func New(pkg, file string) *Table {
	return &Table{
		Symbols: make(map[string]*Package),
		Package: pkg,
		File:    file,
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

func (allSt *Table) MergeWith(pt *Table) {
	for pkg, symbolSet := range pt.Symbols {
		if _, ok := allSt.Symbols[pkg]; ok {
			for _, item := range symbolSet.Datatypes {
				allSt.AddDataType(pkg, item.Name, item.Pos)
			}
			for _, item := range symbolSet.Functions {
				allSt.AddFunction(pkg, item.Name, item.Pos)
			}
			for _, item := range symbolSet.Variables {
				allSt.AddVariable(pkg, item.Name, item.Pos)
			}
			for _, item := range symbolSet.Structfields {
				allSt.AddStructField(pkg, item.Parent, item.Field, item.Pos)
			}
			for _, item := range symbolSet.Methods {
				allSt.AddStructField(pkg, item.Parent, item.Name, item.Pos)
			}
		} else {
			allSt.Symbols[pkg] = symbolSet
		}
	}
}

func (allSt *Table) Lock() {
	allSt.locked = true
}

func (allSt *Table) Unlock() {
	allSt.locked = false
}

func (allSt *Table) AddDataType(pkg, name, pos string) {
	items, exists := allSt.Symbols[pkg]
	if !exists {
		allSt.Symbols[pkg] = newPackage()
		items = allSt.Symbols[pkg]
	}

	k := toKey(name, pos)
	if _, ok := items.Datatypes[k]; ok {
		return
	}

	items.Datatypes[k] = Datatype{
		Name: name,
		Pos:  pos,
	}
}

func (allSt *Table) AddVariable(pkg, name, pos string) {
	items, exists := allSt.Symbols[pkg]
	if !exists {
		allSt.Symbols[pkg] = newPackage()
		items = allSt.Symbols[pkg]
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

func (allSt *Table) AddFunction(pkg, name, pos string) {
	items, exists := allSt.Symbols[pkg]
	if !exists {
		allSt.Symbols[pkg] = newPackage()
		items = allSt.Symbols[pkg]
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

func (allSt *Table) AddStructField(pkg, parent, field string, pos string) {
	items, exists := allSt.Symbols[pkg]
	if !exists {
		allSt.Symbols[pkg] = newPackage()
		items = allSt.Symbols[pkg]
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

func (allSt *Table) AddMethod(pkg, parent, name, pos string) {
	items, exists := allSt.Symbols[pkg]
	if !exists {
		allSt.Symbols[pkg] = newPackage()
		items = allSt.Symbols[pkg]
	}
	items.Methods = append(items.Methods, Method{
		Parent: parent,
		Name:   name,
		Pos:    pos,
	})
}

func (allSt *Table) AddDataTypeField(origin, dataType, field, pos string) {

	allSt.AddSymbol(origin, fmt.Sprintf("%v.%v", dataType, field), pos)
}

func (allSt *Table) AddSymbol(origin, id, pos string) {

}

func (allSt *Table) Print(all bool) {
	fmt.Printf("======================================================================================================\n")
	symPos := make(map[string][]string)
	maxKeyLen := 0
	var keys []string
	for pkg := range allSt.Symbols {
		// agregate positions by data type
		for _, dt := range allSt.Symbols[pkg].Datatypes {
			if !all && !ast.IsExported(dt.Name) {
				continue
			}
			qidid := fmt.Sprintf("D: %v.%v", pkg, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range allSt.Symbols[pkg].Functions {
			if !all && !ast.IsExported(dt.Name) {
				continue
			}
			qidid := fmt.Sprintf("F: %v.%v", pkg, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range allSt.Symbols[pkg].Methods {
			if !all && !ast.IsExported(dt.Name) {
				continue
			}
			qidid := fmt.Sprintf("M: %v.%v.%v", pkg, dt.Parent, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range allSt.Symbols[pkg].Variables {
			if !all && !ast.IsExported(dt.Name) {
				continue
			}
			qidid := fmt.Sprintf("V: %v.%v", pkg, dt.Name)
			if len(qidid) > maxKeyLen {
				maxKeyLen = len(qidid)
			}
			if _, ok := symPos[qidid]; !ok {
				symPos[qidid] = make([]string, 0)
				keys = append(keys, qidid)
			}
			symPos[qidid] = append(symPos[qidid], dt.Pos)
		}
		for _, dt := range allSt.Symbols[pkg].Structfields {
			if !all && !ast.IsExported(dt.Field) {
				continue
			}
			qidid := fmt.Sprintf("S: %v.%v.%v", pkg, dt.Parent, dt.Field)
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
