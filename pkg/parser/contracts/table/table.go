package table

import (
	"encoding/json"
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"k8s.io/klog/v2"
)

type PackageContracts map[string][]contracts.Contract

type Table struct {
	Contracts              PackageContracts `json:"contracts"`
	PackageName            string           `json:"packagename"`
	virtualVariableCounter int
	prefix                 string
	symbolTableDir         string
	goVersion              string
}

func (t *Table) AddContract(contract contracts.Contract) {
	if _, ok := t.Contracts[t.prefix]; !ok {
		t.Contracts[t.prefix] = make([]contracts.Contract, 0)
	}
	t.Contracts[t.prefix] = append(t.Contracts[t.prefix], contract)
	klog.V(2).Infof("Adding contract: %v\n", contracts.Contract2String(contract))
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

func (pc *PackageContracts) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	pContracts := make(PackageContracts)

	for fncName, item := range objMap {
		pContracts[fncName] = make([]contracts.Contract, 0)
		var m []map[string]interface{}
		if err := json.Unmarshal(*item, &m); err != nil {
			return err
		}
		var a []*json.RawMessage
		if err := json.Unmarshal(*item, &a); err != nil {
			return err
		}

		for i, cItem := range m {
			switch contracts.Type(cItem["type"].(string)) {
			case contracts.PropagatesToType:
				r := &contracts.PropagatesTo{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)
			case contracts.BinaryOpType:
				r := &contracts.BinaryOp{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.HasFieldType:
				r := &contracts.HasField{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.UnaryOpType:
				r := &contracts.UnaryOp{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.TypecastsToType:
				r := &contracts.TypecastsTo{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsCompatibleWithType:
				r := &contracts.IsCompatibleWith{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsInvocableType:
				r := &contracts.IsInvocable{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsReferenceableType:
				r := &contracts.IsReferenceable{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.ReferenceOfType:
				r := &contracts.ReferenceOf{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsDereferenceableType:
				r := &contracts.IsDereferenceable{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.DereferenceOfType:
				r := &contracts.DereferenceOf{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsIndexableType:
				r := &contracts.IsIndexable{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsSendableToType:
				r := &contracts.IsSendableTo{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsReceiveableFromType:
				r := &contracts.IsReceiveableFrom{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsIncDecableType:
				r := &contracts.IsIncDecable{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			case contracts.IsRangeableType:
				r := &contracts.IsRangeable{}
				if err := json.Unmarshal(*a[i], &r); err != nil {
					return err
				}
				pContracts[fncName] = append(pContracts[fncName], r)

			default:
				panic(fmt.Errorf("Unrecognized contract %v", cItem["type"].(string)))
			}
		}
	}

	*pc = pContracts

	return nil
}

func New(packageName, symbolTableDir, goVersion string) *Table {
	return &Table{
		Contracts:              make(map[string][]contracts.Contract, 0),
		PackageName:            packageName,
		virtualVariableCounter: 0,
		symbolTableDir:         symbolTableDir,
		goVersion:              goVersion,
	}
}
