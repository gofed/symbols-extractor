package file

import (
	"go/parser"
	"go/token"
	"os"
	"path"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	parsertypes "github.com/gofed/symbols-extractor/pkg/parser/types"
)

// TODO(jchaloup): replace this test with generated tests
func TestDataTypes(t *testing.T) {
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata"
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, "datatypes.go")

	f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, 0)
	if err != nil {
		t.Fatalf("AST Parse error: %v", err)
	}

	config := &parsertypes.Config{
		PackageName:           gopkg,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     global.New(),
	}

	config.SymbolTable.Push()

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	payload := MakePayload(f)
	if err := NewParser(config).Parse(payload); err != nil {
		t.Errorf("Unable to parse file %v: %v", gofile, err)
	}
}
