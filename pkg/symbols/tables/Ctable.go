package tables

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gofed/symbols-extractor/pkg/symbols"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	yaml "gopkg.in/yaml.v2"
)

// TODO:
// - define 'type unknown' that is returned if C.symbol data type is unknown, e.g. CgoSymbol{}
// - define a special symbol table that never returns and error is a C symbol is looked up (just try to determine if the symbol is function or type based on the context)
//

// Examples
// func Print(s string) {
//     cs := C.CString(s)
//     defer C.free(unsafe.Pointer(cs))
//     C.fputs(cs, (*C.FILE)(C.stdout))
// }
// Analysis:
// - C.CString is stored into allocated symbol table as gotypes.CgoSymbol{Def: "CString", Package: "C"}
// - cs is out gotypes.CgoSymbol{Def: "CString", Package: "C"}
// - unsafe.Pointer is processed as usually
// - C.free is stored as gotypes.CgoSymbol{Def: "free", Package: "C"}
// - C.stdout is stored as gotypes.CgoSymbol{Def: "stdout", Package: "C"}
// - C.FILE is stored as gotypes.CgoSymbol{Def: "FILE", Package: "C"}
// - C.fputs is stored as gotypes.CgoSymbol{Def: "fputs", Package: "C"}
//
// Basically, if a gotypes.DataType is of gotypes.CgoSymbol type, it is interpreted based on a context.
// E.g. if it is a part of an expression, it is interpreted as it would have one result value
// E.g. it it is a part of a multi-var assignmet, each variable is of type gotypes.CgoSymbol{Def: "", Package: "C"} unless the type can be deduced
//
// C.GoString convers C zero-terminated array of chars to Go string
// C.int convers Go int to C int
// C.CString covers Go string to C string
//
// Or go tool cgo -godefs translates all C-like structs into G-like ones
type CGOTable struct {
	*Table
}

func NewCGOTable() *CGOTable {
	return &CGOTable{
		Table: NewTable(),
	}
}

func (c *CGOTable) Flush() {
	c.Table = NewTable()
}

type DataType struct {
	gotypes.DataType
}

func processDataType(m map[interface{}]interface{}) gotypes.DataType {
	if ident, ok := m["identifier"]; ok {
		str := ident.(string)
		l := len(str)
		if l > 1 && str[0] == 'C' && str[1] == '.' {
			return &gotypes.Identifier{
				Def:     str[2:l],
				Package: "C",
			}
		}
		// TODO(jchaloup): replace the Builtin with Identifier
		return &gotypes.Builtin{
			Def: ident.(string),
		}
	}
	if ident, ok := m["constant"]; ok {
		str := ident.(string)
		l := len(str)
		if l > 1 && str[0] == 'C' && str[1] == '.' {
			return &gotypes.Constant{
				Def:     str[2:l],
				Package: "C",
				Untyped: true,
				Literal: "*",
			}
		}
		return &gotypes.Constant{
			Def:     ident.(string),
			Package: "builtin",
			Untyped: true,
			Literal: "*",
		}
	}
	if pointer, ok := m["pointer"]; ok {
		return &gotypes.Pointer{
			Def: processDataType(pointer.(map[interface{}]interface{})),
		}
	}
	if structType, ok := m["struct"]; ok {
		var fields []gotypes.StructFieldsItem
		if structType == nil {
			return &gotypes.Struct{}
		}
		for _, field := range structType.([]interface{}) {
			fieldExpr := field.(map[interface{}]interface{})
			item := gotypes.StructFieldsItem{
				Name: fieldExpr["name"].(string),
				Def:  processDataType(fieldExpr),
			}
			fields = append(fields, item)
		}
		return &gotypes.Struct{
			Fields: fields,
		}
	}
	panic(fmt.Errorf("Unknown %#v", m))
}

func (d *DataType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var m map[interface{}]interface{}
	if err := unmarshal(&m); err != nil {
		return err
	}
	d.DataType = processDataType(m)
	return nil
}

type cgoFile struct {
	Name  string   `yaml:"name"`
	Files []string `yaml:"files"`
	Types []struct {
		Name string   `yaml:"name"`
		Type DataType `yaml:"type"`
	}
	Functions []struct {
		Name   string     `yaml:"name"`
		Params []DataType `yaml:"params"`
		Result []DataType `yaml:"result"`
	}
	Variables []struct {
		Name string   `yaml:"name"`
		Type DataType `yaml:"type"`
	}
	Constants []struct {
		Name string   `yaml:"name"`
		Type DataType `yaml:"type"`
	}
}

func (c *CGOTable) LoadFromFile(file string) error {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Unable to load file %v: %v", file, err)
	}
	var cgo cgoFile
	err = yaml.Unmarshal(yamlFile, &cgo)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	for _, typeDef := range cgo.Types {
		if err := c.AddDataType(&symbols.SymbolDef{
			Name:    typeDef.Name,
			Package: "C",
			Def:     typeDef.Type.DataType,
		}); err != nil {
			return fmt.Errorf("AddDataType %q failed: %v", typeDef.Name, err)
		}
	}

	for _, varDef := range cgo.Constants {
		if err := c.AddVariable(&symbols.SymbolDef{
			Name:    varDef.Name,
			Package: "C",
			Def:     varDef.Type.DataType,
		}); err != nil {
			return fmt.Errorf("AddVariable %q failed: %v", varDef.Name, err)
		}
	}

	for _, varDef := range cgo.Variables {
		if err := c.AddVariable(&symbols.SymbolDef{
			Name:    varDef.Name,
			Package: "C",
			Def:     varDef.Type.DataType,
		}); err != nil {
			return fmt.Errorf("AddVariable %q failed: %v", varDef.Name, err)
		}
	}

	for _, fncDef := range cgo.Functions {
		var params []gotypes.DataType
		for _, param := range fncDef.Params {
			params = append(params, param.DataType)
		}

		var results []gotypes.DataType
		for _, result := range fncDef.Result {
			results = append(results, result.DataType)
		}
		f := &gotypes.Function{
			Params:  params,
			Results: results,
		}
		if err := c.AddFunction(&symbols.SymbolDef{
			Name:    fncDef.Name,
			Package: "C",
			Def:     f,
		}); err != nil {
			return fmt.Errorf("AddFunction %q failed: %v", fncDef.Name, err)
		}
	}

	return nil
}

var _ symbols.SymbolLookable = &CGOTable{}
