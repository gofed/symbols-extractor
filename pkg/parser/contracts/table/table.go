package table

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/golang/glog"
)

type Table struct {
	Contracts              map[string][]contracts.Contract
	PackageName            string
	virtualVariableCounter int
	prefix                 string
}

func (t *Table) AddContract(contract contracts.Contract) {
	if _, ok := t.Contracts[t.prefix]; !ok {
		t.Contracts[t.prefix] = make([]contracts.Contract, 0)
	}
	t.Contracts[t.prefix] = append(t.Contracts[t.prefix], contract)
	glog.Infof("Adding contract: %v\n", contracts.Contract2String(contract))
}

func (t *Table) SetPrefix(prefix string) {
	t.prefix = prefix
}

func (t *Table) UnsetPrefix() {
	t.prefix = ""
}

func (t *Table) DropPrefixContracts(prefix string) {
	delete(t.Contracts, prefix)
}

func (t *Table) NewVirtualVar() *typevars.Variable {
	t.virtualVariableCounter++
	return typevars.MakeVirtualVar(t.virtualVariableCounter)
}

func (t *Table) List() map[string][]contracts.Contract {
	return t.Contracts
}

func (t *Table) Save(contracttabledir string) error {
	file := path.Join(contracttabledir, fmt.Sprintf("%v-contracts.json", strings.Replace(t.PackageName, "/", ".", -1)))
	if _, err := os.Stat(file); err == nil {
		// already stored
		return nil
	}

	byteSlice, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("Unable to save %q contract table: %v", t.PackageName, err)
	}

	if err := ioutil.WriteFile(file, byteSlice, 0644); err != nil {
		return err
	}
	return nil
}

func New(packageName string) *Table {
	return &Table{
		Contracts:              make(map[string][]contracts.Contract, 0),
		PackageName:            packageName,
		virtualVariableCounter: 0,
	}
}
