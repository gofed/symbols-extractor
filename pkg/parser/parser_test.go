package parser

import "testing"

// func TestDataTypes(t *testing.T) {
// 	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata"
// 	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, "unordered.go")
// 	err := NewParser(gopkg).Parse(gofile)
// 	if err != nil {
// 		t.Errorf("Unable to parse %v: %v", gofile, err)
// 	}
// }

func TestProjectParser(t *testing.T) {
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata"
	pp := New(gopkg)
	pp.Parse()
}
