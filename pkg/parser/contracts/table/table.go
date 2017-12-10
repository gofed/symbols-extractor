package table

import (
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/golang/glog"
)

type Table struct {
	contracts              []contracts.Contract
	virtualVariableCounter int
	prefix                 string
}

func (t *Table) AddContract(contract contracts.Contract) {
	t.contracts = append(t.contracts, contract)
	glog.Infof("Adding contract: %v\n", contracts.Contract2String(contract))
}

func (t *Table) SetPrefix(prefix string) {
	t.prefix = prefix
}

func (t *Table) NewVariable() string {
	t.virtualVariableCounter++
	return fmt.Sprintf("virtual.var.%v", t.virtualVariableCounter)
}

func (t *Table) Contracts() []contracts.Contract {
	return t.contracts
}

func New() *Table {
	return &Table{
		contracts:              make([]contracts.Contract, 0),
		virtualVariableCounter: 0,
	}
}
