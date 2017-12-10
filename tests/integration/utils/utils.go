package utils

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	contracttable "github.com/gofed/symbols-extractor/pkg/parser/contracts/table"
	exprparser "github.com/gofed/symbols-extractor/pkg/parser/expression"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	stmtparser "github.com/gofed/symbols-extractor/pkg/parser/statement"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable/stack"
	typeparser "github.com/gofed/symbols-extractor/pkg/parser/type"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	parsertypes "github.com/gofed/symbols-extractor/pkg/parser/types"
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

	payload, err := fileparser.MakePayload(f)
	if err != nil {
		return err
	}
	if err := fileparser.NewParser(config).Parse(payload); err != nil {
		return fmt.Errorf("Unable to parse file %v: %v", gofile, err)
	}

	table, err := config.SymbolTable.Table(0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Global storing builtin\n")
	config.GlobalSymbolTable.Add("builtin", table)

	return nil
}

func InitFileParser(gopkg string) (*fileparser.FileParser, *types.Config, error) {
	gtable := global.New("")

	config := &parsertypes.Config{
		PackageName:           "builtin",
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     gtable,
		ContractTable:         contracttable.New(),
	}

	config.SymbolTable.Push()

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	if err := parseBuiltin(config); err != nil {
		return nil, nil, err
	}

	config = &parsertypes.Config{
		PackageName:           gopkg,
		SymbolTable:           stack.New(),
		AllocatedSymbolsTable: alloctable.New(),
		GlobalSymbolTable:     gtable,
		ContractTable:         contracttable.New(),
	}

	config.SymbolTable.Push()

	config.TypeParser = typeparser.New(config)
	config.ExprParser = exprparser.New(config)
	config.StmtParser = stmtparser.New(config)

	return fileparser.NewParser(config), config, nil
}
