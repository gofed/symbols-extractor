package contracts

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

func makePayload(t *testing.T, packageName, filename string) *fileparser.Payload {
	gofile := path.Join(os.Getenv("GOPATH"), "src", packageName, filename)

	f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, 0)
	if err != nil {
		t.Fatalf("Unable to parse file %v, AST Parse error: %v", gofile, err)
	}

	payload, err := fileparser.MakePayload(f)
	if err != nil {
		t.Errorf("Unable to parse file %v, unable to make a payload due to: %v", gofile, err)
	}
	return payload
}

func storePackage(config *types.Config) {
	table, err := config.SymbolTable.Table(0)
	if err != nil {
		panic(err)
	}

	config.GlobalSymbolTable.Add(config.PackageName, table)
}

func ParsePackage(t *testing.T, config *types.Config, fileParser *fileparser.FileParser, packageName, filename, pkg string) error {
	config.PackageName = pkg
	config.SymbolsAccessor.SetCurrentTable(pkg, config.SymbolTable)

	payload := makePayload(t, packageName, filename)
	if e := fileParser.Parse(payload); e != nil {
		return fmt.Errorf("Unable to parse file %v: %v", filename, e)
	}

	storePackage(config)

	if len(payload.DataTypes) > 0 {
		return fmt.Errorf("Payload not fully consumed, missing %v DataTypes", len(payload.DataTypes))
	}

	if len(payload.Variables) > 0 {
		return fmt.Errorf("Payload not fully consumed, missing %v Variables", len(payload.Variables))
	}

	if len(payload.Functions) > 0 {
		return fmt.Errorf("Payload not fully consumed, missing %v Functions", len(payload.Functions))
	}

	// reset symbol table stack
	config.SymbolTable.Pop()
	config.SymbolTable.Push()

	return nil
}

func ParseAndCompareContracts(t *testing.T, gopkg, filename string, tests []contracts.Contract) {
	fileParser, config, err := utils.InitFileParser(gopkg)
	if err != nil {
		t.Error(err)
	}

	if err := ParsePackage(t, config, fileParser, gopkg, filename, gopkg); err != nil {
		t.Error(err)
		return
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
	CompareContracts(t, genContracts, tests)
}

func CompareContracts(t *testing.T, contractsList, tests []contracts.Contract) {
	testsTotal := len(tests)

	if len(contractsList) < testsTotal {
		t.Errorf("Expected at least %v contracts, got %v instead", testsTotal, len(contractsList))
	}
	t.Logf("Got %v tests, %v contracts", testsTotal, len(contractsList))
	c := 0
	for j, exp := range tests {
		t.Logf("Checking %v-th contract: %v\n", j, contracts.Contract2String(contractsList[c]))
		utils.CompareContracts(t, exp, contractsList[c])
		c++
	}

	for i := c; i < len(contractsList); i++ {
		t.Logf("\n\n\n\nAbout to check %v-th contract: %v\n", i, contracts.Contract2String(contractsList[i]))
		//t.Errorf("\n\n\n\nUnprocessed %v-th contract: %v\n", i, contracts.Contract2String(contractsList[i]))
	}
}
