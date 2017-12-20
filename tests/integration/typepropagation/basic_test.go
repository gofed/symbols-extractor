package typepropagation

import (
	"fmt"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/analyzers/typepropagation"
	cutils "github.com/gofed/symbols-extractor/tests/integration/contracts"
	"github.com/gofed/symbols-extractor/tests/integration/utils"
)

func TestSelfTypePropagation(t *testing.T) {

	gopkg := "github.com/gofed/symbols-extractor/tests/integration/typepropagation"
	filename := "testdata/basic.go"

	fileParser, config, err := utils.InitFileParser(gopkg)
	if err != nil {
		t.Error(err)
	}

	if err := cutils.ParsePackage(t, config, fileParser, gopkg, filename, gopkg); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("contracts: %#v\n", config.ContractTable.Contracts())
	typepropagation.New(config).Run()
}
