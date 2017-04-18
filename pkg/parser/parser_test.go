package parser

import (
	"os"
	"path"
	"testing"
)

func TestDataTypes(t *testing.T) {
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata"
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, "datatypes.go")
	err := NewParser(gopkg).Parse(gofile)
	if err != nil {
		t.Errorf("Unable to parse %v: %v", gofile, err)
	}
}
