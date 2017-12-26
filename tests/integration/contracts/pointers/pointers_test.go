package contracts

import (
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/pointers"

func TestPointersContracts(t *testing.T) {
	var vars = map[string]string{
		"a":  ":39",
		"ra": ":52",
		"da": ":75",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"../../testdata/pointers.go",
		[]contracts.Contract{
			// a := "ahoj"
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(packageName, &gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeLocalVar("a", vars["a"]),
			},
			// ra := &a
			&contracts.IsReferenceable{
				X: typevars.MakeLocalVar("a", vars["a"]),
			},
			&contracts.ReferenceOf{
				X: typevars.MakeLocalVar("a", vars["a"]),
				Y: typevars.MakeVirtualVar(1),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeLocalVar("ra", vars["ra"]),
			},
			// da := *ra
			&contracts.IsDereferenceable{
				X: typevars.MakeLocalVar("ra", vars["ra"]),
			},
			&contracts.DereferenceOf{
				X: typevars.MakeLocalVar("ra", vars["ra"]),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeLocalVar("da", vars["da"]),
			},
		})
}
