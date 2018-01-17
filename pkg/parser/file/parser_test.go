package file

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	contracttable "github.com/gofed/symbols-extractor/pkg/parser/contracts/table"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	parsertypes "github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/stack"
)

func parseBuiltin(config *parsertypes.Config) error {
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return fmt.Errorf("GOROOT env not set")
	}
	gofile := path.Join(goroot, "src", "builtin/builtin.go")

	f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, 0)
	if err != nil {
		return fmt.Errorf("AST Parse error: %v", err)
	}

	payload, err := MakePayload(f)
	if err != nil {
		return err
	}
	if err := NewParser(config).Parse(payload); err != nil {
		return fmt.Errorf("Unable to parse file %v: %v", gofile, err)
	}

	table, err := config.SymbolTable.Table(0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Global storing builtin\n")
	config.GlobalSymbolTable.Add("builtin", table, false)

	return nil
}

// TODO(jchaloup): replace this test with generated tests
func TestDataTypes(t *testing.T) {
	gopkg := "github.com/gofed/symbols-extractor/pkg/parser/testdata"
	gofile := path.Join(os.Getenv("GOPATH"), "src", gopkg, "datatypes.go")

	f, err := parser.ParseFile(token.NewFileSet(), gofile, nil, 0)
	if err != nil {
		t.Fatalf("AST Parse error: %v", err)
	}

	gtable := global.New("", "")

	config := &parsertypes.Config{
		PackageName:           "builtin",
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New("", ""),
		GlobalSymbolTable:     gtable,
		ContractTable:         contracttable.New("builtin", "", ""),
	}
	config.SymbolsAccessor = accessors.NewAccessor(config.GlobalSymbolTable).SetCurrentTable(config.PackageName, config.SymbolTable)

	config.SymbolTable.Push()

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	if err := parseBuiltin(config); err != nil {
		t.Fatal(err)
	}

	config = &parsertypes.Config{
		PackageName:           gopkg,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New("", ""),
		GlobalSymbolTable:     gtable,
		ContractTable:         contracttable.New(gopkg, "", ""),
	}
	config.SymbolsAccessor = accessors.NewAccessor(config.GlobalSymbolTable).SetCurrentTable(config.PackageName, config.SymbolTable)

	config.SymbolTable.Push()

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	payload, err := MakePayload(f)
	if err != nil {
		t.Errorf("Unable to make a payload due to: %v", err)
	}
	if err := NewParser(config).Parse(payload); err != nil {
		t.Errorf("Unable to parse file %v: %v", gofile, err)
	}
}
