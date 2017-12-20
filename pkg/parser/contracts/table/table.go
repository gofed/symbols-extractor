package table

import (
	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/golang/glog"
)

type Table struct {
	contracts              map[string][]contracts.Contract
	virtualVariableCounter int
	prefix                 string
}

func (t *Table) AddContract(contract contracts.Contract) {
	if _, ok := t.contracts[t.prefix]; !ok {
		t.contracts[t.prefix] = make([]contracts.Contract, 0)
	}
	t.contracts[t.prefix] = append(t.contracts[t.prefix], contract)
	glog.Infof("Adding contract: %v\n", contracts.Contract2String(contract))
}

func (t *Table) SetPrefix(prefix string) {
	t.prefix = prefix
}

func (t *Table) UnsetPrefix() {
	t.prefix = ""
}

func (t *Table) DropPrefixContracts(prefix string) {
	delete(t.contracts, prefix)
}

func (t *Table) NewVirtualVar() *typevars.Variable {
	t.virtualVariableCounter++
	return typevars.MakeVirtualVar(t.virtualVariableCounter)
}

func (t *Table) Contracts() map[string][]contracts.Contract {
	return t.contracts
}

func New() *Table {
	return &Table{
		contracts:              make(map[string][]contracts.Contract, 0),
		virtualVariableCounter: 0,
	}
}
