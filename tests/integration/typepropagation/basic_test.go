package typepropagation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/runner"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata"
	filename := "basic.go"

	fileParser, config, err := utils.InitFileParser(gopkg)
	if err != nil {
		t.Error(err)
	}

	if err := cutils.ParsePackage(t, config, fileParser, config.PackageName, filename, config.PackageName); err != nil {
		t.Error(err)
		return
	}

	t.Logf("contracts: %#v\n", config.ContractTable.Contracts())
	t.Logf("package: %v\n", config.PackageName)
	r := runner.New(config)
	if err := r.Run(); err != nil {
		t.Fatal(err)
	}

	varTable := r.VarTable()
	names := varTable.Names()

	expected := []struct {
		name     string
		dataType gotypes.DataType
	}{
		{
			name: "#a#:167",
			dataType: &gotypes.Pointer{
				Def: &gotypes.Identifier{
					Def:     "A",
					Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
				},
			},
		},
		{
			name:     "#x#:180",
			dataType: &gotypes.Builtin{Def: "int"},
		},
		{
			name:     "#y#:183",
			dataType: &gotypes.Builtin{Def: "int"},
		},
		{
			name: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata#C#:217",
			dataType: &gotypes.Pointer{
				Def: &gotypes.Identifier{
					Def:     "A",
					Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
				},
			},
		},
		{
			name: "#a#:243",
			dataType: &gotypes.Pointer{
				Def: &gotypes.Identifier{
					Def:     "A",
					Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
				},
			},
		},
		{
			name:     "#b#:323",
			dataType: &gotypes.Builtin{Def: "int"},
		},
		// virtual variables
		{
			name: "#virtual.var.1#",
			dataType: &gotypes.Identifier{
				Def:     "A",
				Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
			},
		},
		{
			name: "#virtual.var.2#",
			dataType: &gotypes.Pointer{
				Def: &gotypes.Identifier{
					Def:     "A",
					Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
				},
			},
		},
		{
			name:     "#virtual.var.3#",
			dataType: &gotypes.Builtin{Def: "int"},
		},
		{
			name: "#virtual.var.4#",
			dataType: &gotypes.Identifier{
				Def:     "A",
				Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
			},
		},

		{
			name: "#virtual.var.5#",
			dataType: &gotypes.Struct{
				Fields: []gotypes.StructFieldsItem{
					{
						Name: "f",
						Def:  &gotypes.Builtin{Def: "float32"},
					},
				},
			},
		},
		{
			name: "#virtual.var.6#",
			dataType: &gotypes.Pointer{
				Def: &gotypes.Identifier{
					Def:     "A",
					Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
				},
			},
		},
		{
			name: "#virtual.var.7#",
			dataType: &gotypes.Method{
				Receiver: &gotypes.Pointer{
					Def: &gotypes.Identifier{
						Def:     "A",
						Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
					},
				},
				Def: &gotypes.Function{
					Package: "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata",
					Params: []gotypes.DataType{
						&gotypes.Builtin{Def: "int"},
						&gotypes.Builtin{Def: "int"},
					},
					Results: []gotypes.DataType{
						&gotypes.Builtin{Def: "int"},
					},
				},
			},
		},
	}

	for _, e := range expected {
		tested := varTable.GetVariable(e.name)
		fmt.Printf("Checking %q variable...\n", e.name)
		if !reflect.DeepEqual(tested.DataType(), e.dataType) {
			tByteSlice, _ := json.Marshal(tested.DataType())
			eByteSlice, _ := json.Marshal(e.dataType)
			t.Errorf("Got %v, expected %v", string(tByteSlice), string(eByteSlice))
		}
	}

	if len(names) > len(expected) {
		t.Errorf("There is %v variables not yet checked", len(names)-len(expected))
	}

	// sort.Strings(names)
	// for _, name := range names {
	// 	fmt.Printf("Name: %v\tDataType: %#v\n", name, varTable.GetVariable(name).DataType())
	// }
}
