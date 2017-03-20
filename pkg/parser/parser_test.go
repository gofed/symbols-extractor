package parser

import (
	"os"
	"path"
	"testing"

	gost "github.com/gofed/symbols-extractor/pkg/types/symboltable"
)

// this is basic test used mainly for purposes of developer to see whether
// parsing of basic ast without "conflicts" and dependencies between types
// still works
func TestDataTypes(t *testing.T) {

	gofile := path.Join(os.Getenv("GOPATH"), "src/github.com/gofed/symbols-extractor/pkg/parser/testdata/datatypes.go")
	err := NewParser().Parse(gofile)
	if err != nil {
		t.Errorf("Unable to parse %v: %v", gofile, err)
	}
}

// this will be basic test to check whther def. symtab is created correctly
// for now, it is used ad-hoc for fast testing of various output
// TODO: create tests for symboltable itself in symboltable_test.go
// TODO: should be created test which tests that JSON is created correctly from
//       this part of symtab
// TODO: should be created test for correct creating of whole JSON according
//       to given JSON scheme,
//FIXME: why I put new symtab and then return symtab when it is pointer???
//NOTE: instead of ParseFiles should be created ParsePackage...
func TestBasicSymbolTable(t *testing.T) {
	gofile := []string{path.Join(os.Getenv("GOPATH"), "src/github.com/gofed/symbols-extractor/pkg/parser/testdata/datatypes.go")}
	var symtab gost.SymbolTable = make(gost.HST)
	var err error
	symtab, err = ParseFiles(gofile, symtab)
	if err != nil {
		t.Errorf("Something bad happened: %v", err)
	}
	//log.Print("--- type: ", symtab)
}
