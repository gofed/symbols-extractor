package contracts

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/runner"
	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	fileparser "github.com/gofed/symbols-extractor/pkg/parser/file"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
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

	config.GlobalSymbolTable.Add(config.PackageName, table, false)
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

	table, err := config.SymbolTable.Table(0)
	if err != nil {
		panic(err)
	}
	table.PackageQID = pkg
	fmt.Printf("Global storing %v\n", pkg)
	config.GlobalSymbolTable.Add(pkg, table, false)

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
	cs := config.ContractTable.List()
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

type VarTableTest struct {
	Name     string
	DataType gotypes.DataType
}

func CompareVarTable(t *testing.T, expected []VarTableTest, testedVarTable *runner.VarTable) {
	t.Logf("#### Checking variables\n")
	for _, e := range expected {
		tested, exists := testedVarTable.GetVariable(e.Name)
		if !exists {
			t.Errorf("Variable %v does not exist", e.Name)
			continue
		}
		t.Logf("Checking %q variable...\n", e.Name)
		if !reflect.DeepEqual(tested.DataType(), e.DataType) {
			tByteSlice, _ := json.Marshal(tested.DataType())
			eByteSlice, _ := json.Marshal(e.DataType)
			t.Errorf("%v: got\n%v, expected\n%v", e.Name, string(tByteSlice), string(eByteSlice))
		}
	}

	names := testedVarTable.Names()
	if len(names) > len(expected) {
		var eNames []string
		eNamesMap := map[string]struct{}{}
		for _, n := range expected {
			eNames = append(eNames, n.Name)
			eNamesMap[n.Name] = struct{}{}
		}
		sort.Strings(eNames)
		sort.Strings(names)
		t.Logf("\n#### Once all expected variables are set, both columns will be equal\n")
		for i := 0; i < len(names); i++ {
			eName := ""
			if _, exists := eNamesMap[names[i]]; exists {
				eName = names[i]
			}
			fmt.Printf("test.name: %v\t\te.name: %v\n", names[i], eName)
		}

		if len(names)-len(expected) > 0 {
			t.Logf("\n#### There is %v variables not yet checked\n", len(names)-len(expected))
			for _, name := range names {
				if _, exists := eNamesMap[name]; !exists {
					t.Errorf("%v variables not yet checked", name)
				}
			}
		}
	}

	// sort.Strings(names)
	// for _, name := range names {
	// 	fmt.Printf("Name: %v\tDataType: %#v\n", name, varTable.GetVariable(name).DataType())
	// }
}

func ParseAndCompareVarTable(t *testing.T, gopkg, filename string, expected []VarTableTest) {
	fileParser, config, err := utils.InitFileParser(gopkg)
	if err != nil {
		t.Error(err)
	}

	if err := ParsePackage(t, config, fileParser, config.PackageName, filename, config.PackageName); err != nil {
		t.Error(err)
		return
	}

	r := runner.New(config.PackageName, config.GlobalSymbolTable, allocglobal.New("", "", nil), config.ContractTable)
	if err := r.Run(); err != nil {
		t.Fatal(err)
	}

	r.VarTable().Dump()

	CompareVarTable(t, expected, r.VarTable())
}
