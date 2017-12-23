package runner

import (
	"fmt"
	"sort"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/symbols"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type var2Contract struct {
	vars map[string][]contracts.Contract
}

func newVar2Contract() *var2Contract {
	return &var2Contract{
		vars: make(map[string][]contracts.Contract),
	}
}

func (v *var2Contract) addVar(name string, c contracts.Contract) {
	// Storing virtual variables only => no package scope
	if _, ok := v.vars[name]; !ok {
		v.vars[name] = make([]contracts.Contract, 0)
	}
	v.vars[name] = append(v.vars[name], c)
}

type contractPayload struct {
	items map[string][]contracts.Contract
}

func newContractPayload(ctrs map[string][]contracts.Contract) *contractPayload {
	if ctrs == nil {
		return &contractPayload{
			items: make(map[string][]contracts.Contract),
		}
	}
	return &contractPayload{
		items: ctrs,
	}
}

func (cp *contractPayload) addContract(funcName string, c contracts.Contract) {
	if _, ok := cp.items[funcName]; !ok {
		cp.items[funcName] = make([]contracts.Contract, 0)
	}
	cp.items[funcName] = append(cp.items[funcName], c)
}

func (cp *contractPayload) contracts() map[string][]contracts.Contract {
	return cp.items
}

func (cp *contractPayload) isEmpty() bool {
	return len(cp.items) == 0
}

func (cp *contractPayload) dump() {
	for _, d := range cp.items {
		for _, c := range d {
			fmt.Printf("  %v\n", contracts.Contract2String(c))
		}
	}
}

/////////////////////////////////////////////////////////////
// Mapping of variables to its actual data type definition //
// virtual.var.1: ...
// virtual.var.1#Field(name): ...
// virtual.var.1#Field(name): ...
// virtual.var.1#MapValue: ...
// virtual.var.1#ListValue: ...
//
type varTableItem struct {
	dataType    gotypes.DataType
	symbolTable symbols.SymbolLookable
	packageName string
}

func (v *varTableItem) DataType() gotypes.DataType {
	return v.dataType
}

type varTable struct {
	// variable name
	variables map[string]*varTableItem
	// variable name, field
	fields map[string]map[string]*varTableItem
}

func newVarTable() *varTable {
	return &varTable{
		variables: make(map[string]*varTableItem),
		fields:    make(map[string]map[string]*varTableItem),
	}
}

func (v *varTable) Names() []string {
	var names []string
	for name := range v.variables {
		names = append(names, name)
	}
	return names
}

func (v *varTable) SetVariable(name string, item *varTableItem) {
	v.variables[name] = item
}

func (v *varTable) GetVariable(name string) *varTableItem {
	// TODO(jchaloup): handle case when the variable does not exist
	return v.variables[name]
}

func (v *varTable) SetField(name, field string, item *varTableItem) {
	if _, ok := v.fields[name]; !ok {
		v.fields[name] = make(map[string]*varTableItem)
	}
	v.fields[name][field] = item
}

func (v *varTable) GetField(name, field string) (*varTableItem, bool) {
	item, ok := v.fields[name][field]
	return item, ok
}

func (v varTable) Dump() {
	var keys []string
	for key := range v.variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("varsTable[%v]: %#v\n", key, v.variables[key])
	}
}
