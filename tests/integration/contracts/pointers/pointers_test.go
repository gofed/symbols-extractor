package contracts

import (
	"go/token"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	utils "github.com/gofed/symbols-extractor/tests/integration/contracts"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/pointers"

func TestPointersContracts(t *testing.T) {
	var vars = map[string]string{
		"a":  ":39:a",
		"ra": ":52:ra",
		"da": ":75:da",
	}
	utils.ParseAndCompareContracts(
		t,
		packageName,
		"testdata/pointers.go",
		[]contracts.Contract{
			// a := "ahoj"
			&contracts.PropagatesTo{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "string"}),
				Y: typevars.MakeVar(vars["a"], packageName),
			},
			// ra := &a
			&contracts.UnaryOp{
				X:       typevars.MakeVar(vars["a"], packageName),
				Y:       typevars.MakeVirtualVar(1),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(1),
				Y: typevars.MakeVar(vars["ra"], packageName),
			},
			// da := *ra
			&contracts.IsDereferenceable{
				X: typevars.MakeVar(vars["ra"], packageName),
			},
			&contracts.DereferenceOf{
				X: typevars.MakeVar(vars["ra"], packageName),
				Y: typevars.MakeVirtualVar(2),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeVar(vars["da"], packageName),
			},
		})
}
