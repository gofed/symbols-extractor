package multi_packages

import (
	"go/token"
	"sort"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	conutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

func TestMultiPackageContracts(t *testing.T) {

	packageName := "github.com/gofed/symbols-extractor/tests/integration/contracts/multi_packages"

	// Parse builtin and initialize the file parser
	fileParser, config, err := utils.InitFileParser("pkgA")
	if err != nil {
		t.Error(err)
	}

	// Parse pkgA package
	if err := conutils.ParsePackage(t, config, fileParser, packageName, "testdata/pkgA/pkg.go", "pkgA"); err != nil {
		t.Error(err)
		return
	}
	// Parse pkgB package
	if err := conutils.ParsePackage(t, config, fileParser, packageName, "testdata/pkgB/pkg.go", "pkgB"); err != nil {
		t.Error(err)
		return
	}

	var vars = map[string]string{
		"a": ":47",
		"b": ":80",
		"c": ":117",
	}

	var genContracts []contracts.Contract
	cs := config.ContractTable.Contracts()
	var keys []string
	for fncName := range cs {
		keys = append(keys, fncName)
	}
	sort.Strings(keys)
	for _, key := range keys {
		genContracts = append(genContracts, cs[key]...)
	}

	conutils.CompareContracts(
		t,
		genContracts,
		[]contracts.Contract{
			// a := &pkgA.A{}
			// a.method(1, 2)
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar("pkgA", "A", ""),
				Y: typevars.MakeVirtualVar(1),
			},
			&contracts.UnaryOp{
				X:       typevars.MakeVirtualVar(1),
				Y:       typevars.MakeVirtualVar(2),
				OpToken: token.AND,
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(2),
				Y: typevars.MakeLocalVar("a", vars["a"]),
			},
			&contracts.HasField{
				X:     typevars.MakeLocalVar("a", vars["a"]),
				Field: "method",
			},
			&contracts.PropagatesTo{
				X: typevars.MakeField(typevars.MakeLocalVar("a", vars["a"]), "method", 0),
				Y: typevars.MakeVirtualVar(3),
			},
			&contracts.IsInvocable{
				F:         typevars.MakeVirtualFunction(typevars.MakeVirtualVar(3)),
				ArgsCount: 2,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeArgument(typevars.MakeVirtualFunction(typevars.MakeVirtualVar(3)), 0),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
				Y: typevars.MakeArgument(typevars.MakeVirtualFunction(typevars.MakeVirtualVar(3)), 1),
			},
			// b := pkgA.B{
			//   pkgA.A{1},
			//   {2},
			// }
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(5),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(5), "", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// The first item has its type given explicitly
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar("pkgA", "A", ""),
				Y: typevars.MakeVirtualVar(5),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(4)),
				Y: typevars.MakeVirtualVar(5),
			},
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(6),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(6), "", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			// The second item has not its type given explicitly => no contract
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(4)),
				Y: typevars.MakeVirtualVar(6),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar("pkgA", "B", ""),
				Y: typevars.MakeVirtualVar(4),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(4),
				Y: typevars.MakeLocalVar("b", vars["b"]),
			},
			// c := pkgA.C{
			// 	0: pkgA.B{
			// 		{1},
			// 		{},
			// 	},
			// 	1: {
			// 		pkgA.A{f: 2},
			// 	},
			// }
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapKey(typevars.MakeVirtualVar(7)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(9),
				Index: 0,
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(9), "", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(8)),
				Y: typevars.MakeVirtualVar(9),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(8)),
				Y: typevars.MakeVirtualVar(10),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar("pkgA", "B", ""),
				Y: typevars.MakeVirtualVar(8),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapValue(typevars.MakeVirtualVar(7)),
				Y: typevars.MakeVirtualVar(8),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapKey(typevars.MakeVirtualVar(7)),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.HasField{
				X:     typevars.MakeVirtualVar(12),
				Field: "f",
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeField(typevars.MakeVirtualVar(12), "f", 0),
				Y: typevars.MakeConstant(&gotypes.Builtin{Untyped: true, Def: "int"}),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar("pkgA", "A", ""),
				Y: typevars.MakeVirtualVar(12),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeListValue(typevars.MakeVirtualVar(11)),
				Y: typevars.MakeVirtualVar(12),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeMapValue(typevars.MakeVirtualVar(7)),
				Y: typevars.MakeVirtualVar(11),
			},
			&contracts.IsCompatibleWith{
				X: typevars.MakeVar("pkgA", "C", ""),
				Y: typevars.MakeVirtualVar(7),
			},
			&contracts.PropagatesTo{
				X: typevars.MakeVirtualVar(7),
				Y: typevars.MakeLocalVar("c", vars["c"]),
			},
		},
	)
}
