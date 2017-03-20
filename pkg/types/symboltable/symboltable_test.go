package symboltable

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

//func TestAddSymbol(t *testing.T) {
//  var st SymbolTable = make(HST)
//
//}

func TestMarshalUnmarshal(t *testing.T) {
	id := &gotypes.Identifier{Def: "Poker"}
	channel := &gotypes.Channel{
		Dir:   "bi-directional",
		Value: id,
	}
	slice := &gotypes.Slice{
		Elmtype: channel,
	}

	var sym Symbol
	var st SymbolTable = make(HST)
	var pkg *Package = &Package{
		Project:     "", //TODO: need change in schema to store complete ID
		Imports:     nil,
		SymbolTable: st,
		Name:        "paegas",
	}
	pkgMap := make(PackageMap)
	pkgMap[pkg.Name] = pkg

	// fill symbol table
	//NOTE: use zero values for DeclPos, because is not part of JSON scheme
	//      now and DeepEqueal fails otherwise
	{
		sym = &DeclType{
			Pos:  DeclPos{"", 0},
			Name: "someChannel",
			Def:  channel,
		}
		st.AddSymbol(sym.GetName(), sym)

		sym = &DeclType{
			Pos:  DeclPos{"", 0},
			Name: "sliceChan",
			Def:  slice,
		}
		st.AddSymbol(sym.GetName(), sym)
	}

	byteSlice, err := json.Marshal(&pkgMap)
	if err != nil {
		t.Errorf("JSON serialisation failed: %v", err)
		return
	}
	fmt.Println(string(byteSlice))

	var newPkgMap PackageMap = make(PackageMap)
	if err := json.Unmarshal(byteSlice, &newPkgMap); err != nil {
		t.Errorf("JSON deserialisation failed: %v", err)
	}

	if !reflect.DeepEqual(pkgMap, newPkgMap) {
		t.Errorf("%#v != %#v", pkgMap, newPkgMap)
	}
}
