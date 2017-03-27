package symboltable

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func TestMarshalUnmarshal(t *testing.T) {
	id := &gotypes.Identifier{Def: "Poker"}
	channel := &gotypes.Channel{
		Dir:   "bi-directional",
		Value: id,
	}
	slice := &gotypes.Slice{
		Elmtype: channel,
	}

	var sym *SymbolDef
	var pkg *Package = &Package{}
	pkg.Init()
	pkg.Name = "paegas"

	pkgMap := make(PackageMap)
	pkgMap[pkg.Name] = pkg

	// fill symbol table
	//NOTE: use zero values for DeclPos, because is not part of JSON scheme
	//      now and DeepEqueal fails otherwise
	{
		sym = &SymbolDef{
			Pos:  DeclPos{"", 0},
			Name: "someChannel",
			Def:  channel,
		}
		pkg.AddDataType(sym.GetName(), sym)

		sym = &SymbolDef{
			Pos:  DeclPos{"", 0},
			Name: "sliceChan",
			Def:  slice,
		}
		pkg.AddDataType(sym.GetName(), sym)
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
