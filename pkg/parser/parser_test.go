package parser

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/types"
)

func TestProjectParser(t *testing.T) {
	pp := New("github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered")
	if err := pp.Parse(); err != nil {
		t.Errorf("Parse error: %v", err)
	}
	gtable := pp.GlobalSymbolTable()
	atable := pp.GlobalAllocTable()

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
			Def: &types.Identifier{
				Def:     "int",
				Package: "builtin",
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
			t.Errorf("Symbol table %q mismatch. \nGot:\n%v\nExpected:\n%v", expectedPackages["unordered"], string(x), string(y))
		}

		at := alloctable.New()
		// type Struct struct {
		// 	i MyInt
		// 	impa *pkg.Imp
		// 	impb *pkgb.Imp
		// }
		at.AddSymbol(expectedPackages["unordered"], "MyInt")
		at.AddSymbol(expectedPackages["pkg"], "Imp")
		at.AddSymbol(expectedPackages["pkgb"], "Imp")
		// type MyInt int
		at.AddSymbol("builtin", "int")
		// func Nic() *pkg.Imp {
		// 	pkg.Nic().Imp.Size
		// 	return &pkg.Imp{
		// 		Name: "haluz",
		// 		Size: 2,
		// 	}
		// }
		at.AddSymbol(expectedPackages["pkg"], "Imp")
		at.AddSymbol(expectedPackages["pkg"], "Nic")
		// accounted as it is returned by pkg.Nic() and its field Imp is accessed
		at.AddSymbol(expectedPackages["pkg"], "Imp")
		// temporary bump by one once the allocated symbol is extended with file.go:linenumber
		at.AddSymbol(expectedPackages["pkg"], "Imp")
		// Accessing field Imp of type Imp does not imply allocation of the type Imp itself,
		// Return type of the Nic() can change but it the field Imp is still available,
		// The change is backward compatible. Thus, the type Imp must not be allocated.
		at.AddDataTypeField(expectedPackages["pkg"], "Imp", "Imp")
		at.AddDataTypeField(expectedPackages["pkgb"], "Imp", "Size")
		at.AddSymbol(expectedPackages["pkg"], "Imp")
		at.AddDataTypeField(expectedPackages["pkg"], "Imp", "Name")
		at.AddDataTypeField(expectedPackages["pkg"], "Imp", "Size")

		pat, _ := atable.Lookup(expectedPackages["unordered"], "unordered.go")
		if !reflect.DeepEqual(at, pat) {
			x, _ := json.Marshal(pat)
			y, _ := json.Marshal(at)
			t.Errorf("Alloc symbol table %q mismatch.\nGot:\n%v\nExpected:\n%v", expectedPackages["pkg"], string(x), string(y))
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
						Def: &types.Identifier{
							Def:     "string",
							Package: "builtin",
						},
					},
					{
						Name: "Size",
						Def: &types.Identifier{
							Def:     "int",
							Package: "builtin",
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
			t.Errorf("Symbol table %q mismatch. \nGot:\n%v\nExpected:\n%v", expectedPackages["pkg"], string(x), string(y))
		}
		at := alloctable.New()
		at.AddSymbol("builtin", "string")
		at.AddSymbol("builtin", "int")
		at.AddSymbol("github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkgb", "Imp")
		at.AddSymbol("github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkg", "Imp")
		at.AddSymbol("github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkg", "Imp")
		// temporary bump by one once the allocated symbol is extended with file.go:linenumber
		at.AddSymbol("github.com/gofed/symbols-extractor/pkg/parser/testdata/unordered/pkg", "Imp")

		pat, _ := atable.Lookup(expectedPackages["pkg"], "pkg.go")
		if !reflect.DeepEqual(at, pat) {
			x, _ := json.Marshal(pat)
			y, _ := json.Marshal(at)
			t.Errorf("Alloc symbol table %q mismatch.\nGot:\n%v\nExpected:\n%v", expectedPackages["pkg"], string(x), string(y))
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
						Def: &types.Identifier{
							Def:     "string",
							Package: "builtin",
						},
					},
					{
						Name: "Size",
						Def: &types.Identifier{
							Def:     "int",
							Package: "builtin",
						},
					},
				},
			},
		})
		pst, _ := gtable.Lookup(expectedPackages["pkgb"])
		if !reflect.DeepEqual(st, pst) {
			x, _ := json.Marshal(pst)
			y, _ := json.Marshal(st)
			t.Errorf("Symbol table %q mismatch.\nGot:\n%v\nExpected:\n%v", expectedPackages["pkgb"], string(x), string(y))
		}
		at := alloctable.New()
		at.AddSymbol("builtin", "string")
		at.AddSymbol("builtin", "int")

		pat, _ := atable.Lookup(expectedPackages["pkgb"], "pkg.go")
		if !reflect.DeepEqual(at, pat) {
			x, _ := json.Marshal(pat)
			y, _ := json.Marshal(at)
			t.Errorf("Alloc symbol table %q mismatch.\nGot:\n%v\nExpected:\n%v", expectedPackages["pkgb"], string(x), string(y))
		}
	}
}
