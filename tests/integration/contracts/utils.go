package contracts

import (
	"go/parser"
	"go/token"
	"os"
	"path"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

var packageName = "github.com/gofed/symbols-extractor/tests/integration/contracts/testdata"

func compareContracts(t *testing.T, gopkg, filename string, tests []contracts.Contract) {
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, filename)

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

	contractsList := config.ContractTable.Contracts()

	testsTotal := len(tests)

	if len(contractsList) < testsTotal {
		t.Errorf("Expected at least %v contracts, got %v instead", testsTotal, len(contractsList))
	}

	c := 0
	for j, exp := range tests {
		t.Logf("Checking %v-th contract: %v\n", j, contracts.Contract2String(contractsList[c]))
		utils.CompareContracts(t, exp, contractsList[c])
		c++
	}

	for i := c; i < len(contractsList); i++ {
		t.Logf("\n\n\n\nAbout to check %v-th contract: %v\n", i, contracts.Contract2String(contractsList[i]))
	}
}
