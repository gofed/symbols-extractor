package parser

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/types"
)

func TestProjectParser(t *testing.T) {
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata"
	pp := New(gopkg)
	if err := pp.Parse(); err != nil {
		t.Errorf("Parse error: %v", err)
	}
	gtable := pp.GlobalSymbolTable()

	// Check all packages and its symbols are available
	pkgs := make(map[string]struct{}, 0)
	for _, key := range gtable.Packages() {
		pkgs[key] = struct{}{}
	}

	expectedPackages := map[string]string{
		"unordered": "github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered",
		"pkg":       "github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkg",
		"pkgb":      "github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkgb",
	}

	for _, pkg := range expectedPackages {
		if _, ok := pkgs[pkg]; !ok {
			t.Errorf("Package %q not found", pkg)
		}
	}

	// Check unordered ST
	{
		st := symboltable.NewTable()
		st.AddDataType(&types.SymbolDef{
			Name:    "Struct",
			Package: expectedPackages["unordered"],
			Def: &types.Struct{
				Fields: []types.StructFieldsItem{
					{
						Name: "i",
						Def: &types.Identifier{
							Def:     "MyInt",
							Package: expectedPackages["unordered"],
						},
					},
					{
						Name: "impa",
						Def: &types.Pointer{
							Def: &types.Selector{
								Item: "Imp",
								Prefix: &types.Packagequalifier{
									Name: "pkg",
									Path: expectedPackages["pkg"],
								},
							},
						},
					},
					{
						Name: "impb",
						Def: &types.Pointer{
							Def: &types.Selector{
								Item: "Imp",
								Prefix: &types.Packagequalifier{
									Name: "pkgb",
									Path: expectedPackages["pkgb"],
								},
							},
						},
					},
				},
			},
		})
		st.AddDataType(&types.SymbolDef{
			Name:    "MyInt",
			Package: expectedPackages["unordered"],
			Def: &types.Builtin{
				Def: "int",
			},
		})
		st.AddFunction(&types.SymbolDef{
			Name:    "Nic",
			Package: expectedPackages["unordered"],
			Def: &types.Function{
				Params: nil,
				// *pkg.Imp
				Results: []types.DataType{
					&types.Pointer{
						Def: &types.Selector{
							Item: "Imp",
							Prefix: &types.Packagequalifier{
								Name: "pkg",
								Path: expectedPackages["pkg"],
							},
						},
					},
				},
			},
		})

		pst, _ := gtable.Lookup(expectedPackages["unordered"])
		if !reflect.DeepEqual(st, pst) {
			x, _ := json.Marshal(pst)
			y, _ := json.Marshal(st)
			t.Errorf("Symbol table %q mismatch. Got:\n%v\nExpected:\n%v", expectedPackages["unordered"], string(x), string(y))
		}
	}

	// Check unordered/pkg ST
	{
		st := symboltable.NewTable()
		st.AddDataType(&types.SymbolDef{
			Name:    "Imp",
			Package: expectedPackages["pkg"],
			Def: &types.Struct{
				Fields: []types.StructFieldsItem{
					{
						Name: "Name",
						Def: &types.Builtin{
							Def: "string",
						},
					},
					{
						Name: "Size",
						Def: &types.Builtin{
							Def: "int",
						},
					},
					{
						Name: "Imp",
						Def: &types.Pointer{
							Def: &types.Selector{
								Item: "Imp",
								Prefix: &types.Packagequalifier{
									Name: "pkgb",
									Path: expectedPackages["pkgb"],
								},
							},
						},
					},
				},
			},
		})
		st.AddFunction(&types.SymbolDef{
			Name:    "Nic",
			Package: expectedPackages["pkg"],
			Def: &types.Function{
				Params: nil,
				Results: []types.DataType{
					&types.Pointer{
						Def: &types.Identifier{
							Def:     "Imp",
							Package: expectedPackages["pkg"],
						},
					},
				},
			},
		})
		pst, _ := gtable.Lookup(expectedPackages["pkg"])
		if !reflect.DeepEqual(st, pst) {
			x, _ := json.Marshal(pst)
			y, _ := json.Marshal(st)
			t.Errorf("Symbol table %q mismatch. Got:\n%v\nExpected:\n%v", expectedPackages["pkg"], string(x), string(y))
		}
	}

	// Check unordered/pkgb ST
	{
		st := symboltable.NewTable()
		st.AddDataType(&types.SymbolDef{
			Name:    "Imp",
			Package: expectedPackages["pkgb"],
			Def: &types.Struct{
				Fields: []types.StructFieldsItem{
					{
						Name: "Name",
						Def: &types.Builtin{
							Def: "string",
						},
					},
					{
						Name: "Size",
						Def: &types.Builtin{
							Def: "int",
						},
					},
				},
			},
		})
		pst, _ := gtable.Lookup(expectedPackages["pkgb"])
		if !reflect.DeepEqual(st, pst) {
			x, _ := json.Marshal(pst)
			y, _ := json.Marshal(st)
			t.Errorf("Symbol table %q mismatch. Got:\n%v\nExpected:\n%v", expectedPackages["pkgb"], string(x), string(y))
		}
	}
}
