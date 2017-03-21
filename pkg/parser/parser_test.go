package parser

import (
	"os"
	"path"
	"testing"
)

func TestDataTypes(t *testing.T) {

	gofile := path.Join(os.Getenv("GOPATH"), "src/github.com/gofed/symbols-extractor/pkg/parser/testdata/datatypes.go")
	err := NewParser().Parse(gofile)
	if err != nil {
		t.Errorf("Unable to parse %v: %v", gofile, err)
	}
}
