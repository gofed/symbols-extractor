package typepropagation

import (
	"fmt"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/runner"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation/testdata"
	filename := "basic.go"

	fileParser, config, err := utils.InitFileParser(gopkg)
	if err != nil {
		t.Error(err)
	}

	if err := cutils.ParsePackage(t, config, fileParser, config.PackageName, filename, config.PackageName); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("contracts: %#v\n", config.ContractTable.Contracts())
	fmt.Printf("package: %v\n", config.PackageName)
	runner.New(config).Run()
}
