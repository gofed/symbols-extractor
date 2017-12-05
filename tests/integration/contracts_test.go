package file

import (
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

func checkContractTypeVarName(t *testing.T, i int, substr string, tv typevars.Interface, tvname string, ctype string) {
	if substr != "" {
		switch v := tv.(type) {
		case *typevars.Constant:
			if substr != "@const" {
				t.Errorf("Unexpected typevars.Constant %#v", v)
			}
		case *typevars.Variable:
			if !strings.Contains(v.Name, substr) {
				t.Errorf("expected %v.Name of %v-th %q contract to contain %q string, got %q instead", tvname, i, ctype, substr, v.Name)
			}
		default:
			t.Errorf("Unrecognized data type: %#v", tv)
		}
	}
}

// TODO(jchaloup): replace this test with generated tests
func TestDataTypes(t *testing.T) {
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata/contracts"
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, "typevars.go")

	f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, 0)
	if err != nil {
		t.Fatalf("AST Parse error: %v", err)
	}

	payload, err := fileparser.MakePayload(f)
	if err != nil {
		t.Errorf("Unable to make a payload due to: %v", err)
	}

	parser, config, err := utils.InitFileParser(gopkg)
	if err != nil {
		t.Error(err)
	}

	if err := parser.Parse(payload); err != nil {
		t.Errorf("Unable to parse file %v: %v", gofile, err)
	}

	tests := []struct {
		// contract type
		contract string
		// string expected to be contained in a TypeVar's name
		x, y, z string
	}{
		{
			contract: "PropagatesTo",
			x:        "@const",
			y:        ":c",
		},
		{
			contract: "IsCompatibleWith",
			x:        "@const",
			y:        ":d",
		},
		{
			contract: "BinaryOp",
			x:        "@const",
			y:        "@const",
			z:        "virtual.var.1",
		},
		{
			contract: "PropagatesTo",
			x:        "virtual.var.1",
			y:        ":e",
		},
		{
			contract: "BinaryOp",
			x:        ":d",
			y:        "@const",
			z:        "virtual.var.2",
		},
		{
			contract: "PropagatesTo",
			x:        "virtual.var.2",
			y:        ":f",
		},
		{
			contract: "BinaryOp",
			x:        ":e",
			y:        ":f",
			z:        "virtual.var.3",
		},
		{
			contract: "PropagatesTo",
			x:        "virtual.var.3",
			y:        ":ff",
		},
	}

	contractsList := config.ContractTable.Contracts()

	if len(contractsList) < len(tests) {
		t.Errorf("Expected at least %q contracts, got %q instead", len(tests), len(contractsList))
	}

	for i, exp := range tests {
		t.Logf("Checking contract: %#v\n", contractsList[i])
		switch exp.contract {
		case "PropagatesTo":
			c, ok := contractsList[i].(*contracts.PropagatesTo)
			if !ok {
				t.Errorf("%v-th contract must be 'PropagatesTo'", i)
			}
			checkContractTypeVarName(t, i, exp.x, c.X, "X", "PropagatesTo")
			checkContractTypeVarName(t, i, exp.y, c.Y, "Y", "PropagatesTo")
		case "IsCompatibleWith":
			c, ok := contractsList[i].(*contracts.IsCompatibleWith)
			if !ok {
				t.Errorf("%v-th contract must be 'IsCompatibleWith'", i)
			}
			checkContractTypeVarName(t, i, exp.x, c.X, "X", "IsCompatibleWith")
			checkContractTypeVarName(t, i, exp.y, c.Y, "Y", "IsCompatibleWith")
		case "BinaryOp":
			c, ok := contractsList[i].(*contracts.BinaryOp)
			if !ok {
				t.Errorf("%v-th contract must be 'BinaryOp'", i)
			}
			checkContractTypeVarName(t, i, exp.x, c.X, "X", "BinaryOp")
			checkContractTypeVarName(t, i, exp.y, c.Y, "Y", "BinaryOp")
			checkContractTypeVarName(t, i, exp.z, c.Z, "Z", "BinaryOp")
		default:
			t.Errorf("%v-th contract %#v not recognized", i, contractsList[i])
		}
	}
}
