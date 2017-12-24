package runner

import (
	"fmt"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/propagation"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type Runner struct {
	v2c           *var2Contract
	entryTypevars map[string]typevars.Interface
	// data type propagation through virtual variables
	*varTable

	// contracts
	waitingContracts *contractPayload

	globalSymbolTable *global.Table

	symbolAccessor *accessors.Accessor
}

func New(config *types.Config) *Runner {
	r := &Runner{
		v2c:               newVar2Contract(),
		entryTypevars:     make(map[string]typevars.Interface, 0),
		varTable:          newVarTable(),
		waitingContracts:  newContractPayload(nil),
		globalSymbolTable: config.GlobalSymbolTable,
		symbolAccessor:    accessors.NewAccessor(config.GlobalSymbolTable),
	}

	st, err := config.GlobalSymbolTable.Lookup(config.PackageName)
	if err != nil {
		panic(err)
	}

	r.symbolAccessor.SetCurrentTable(config.PackageName, st)

	storeVar := func(v *typevars.Variable, c contracts.Contract) {
		// Allocate table of variables for data type propagation
		r.varTable.SetVariable(v.String(), nil)
		r.v2c.addVar(v.String(), c)
		if v.Package != "" {
			// Entry points are all global scope variables.
			// Except for var a = ... cases where the type is not known (handled later).
			r.entryTypevars[v.String()] = v
		}
	}

	isVariable := func(i typevars.Interface) bool {
		_, ok := i.(*typevars.Variable)
		return ok
	}

	for funcName, cs := range config.ContractTable.Contracts() {
		for _, c := range cs {
			r.waitingContracts.addContract(funcName, c)
			switch d := c.(type) {
			case *contracts.UnaryOp:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.BinaryOp:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
				if isVariable(d.Y) {
					storeVar(d.Y.(*typevars.Variable), c)
				}
			case *contracts.HasField:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.IsCompatibleWith:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
				if isVariable(d.Y) {
					storeVar(d.Y.(*typevars.Variable), c)
				}
			case *contracts.PropagatesTo:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
				if isVariable(d.Y) {
					storeVar(d.Y.(*typevars.Variable), c)
				}
			case *contracts.IsInvocable:
				if isVariable(d.F) {
					storeVar(d.F.(*typevars.Variable), c)
				}
			default:
				panic(fmt.Sprintf("Unrecognized contract: %#v", c))
			}
		}
	}

	// Find all global variables that are on RHS of the PropagatesTo contract
	// and remove them from the entry variables
	for _, cs := range config.ContractTable.Contracts() {
		for _, c := range cs {
			if d, ok := c.(*contracts.PropagatesTo); ok {
				if v, ok := d.Y.(*typevars.Variable); ok {
					if v.Package != "" {
						delete(r.entryTypevars, v.String())
					}
				}
			}
		}
	}

	return r
}

func (r *Runner) isTypevarEvaluated(i typevars.Interface) bool {
	switch d := i.(type) {
	case *typevars.Constant:
		return true
	case *typevars.Variable:
		if r.varTable.GetVariable(d.String()) != nil {
			return true
		}
		return false
	case *typevars.Field:
		if r.isTypevarEvaluated(d.X) {
			_, exists := r.varTable.GetField(d.X.String(), d.Name)
			return exists
		}
		return false
	case *typevars.ReturnType:
		return r.isTypevarEvaluated(d.Function)
	case *typevars.Argument:
		return r.isTypevarEvaluated(d.Function)
	default:
		panic(fmt.Sprintf("Unrecognized typevar: %#v", i))
	}
}

func (r *Runner) splitContracts(ctrs *contractPayload) (*contractPayload, *contractPayload) {
	// create two groups:
	// - contracts ready for evaluation
	// - contracts not ready for evaluation

	readyPayload := newContractPayload(nil)
	unreadyPayload := newContractPayload(nil)

	for key, cs := range ctrs.contracts() {
		for _, c := range cs {
			ready := false
			switch d := c.(type) {
			case *contracts.UnaryOp:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.BinaryOp:
				if r.isTypevarEvaluated(d.X) && r.isTypevarEvaluated(d.Y) {
					ready = true
				}
			case *contracts.HasField:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.IsCompatibleWith:
				if r.isTypevarEvaluated(d.X) && r.isTypevarEvaluated(d.Y) {
					ready = true
				}
			case *contracts.PropagatesTo:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.IsInvocable:
				if r.isTypevarEvaluated(d.F) {
					ready = true
				}
			default:
				panic(fmt.Sprintf("Unrecognized contract: %#v", c))
			}

			if ready {
				readyPayload.addContract(key, c)
			} else {
				unreadyPayload.addContract(key, c)
			}
		}
	}

	// fmt.Printf("\n\nReady:\n")
	// readyPayload.dump()
	// fmt.Printf("\nUnready:\n")
	// unreadyPayload.dump()

	return readyPayload, unreadyPayload
}

func (r *Runner) evaluateContract(c contracts.Contract) error {
	getVar := func(i typevars.Interface) *varTableItem {
		return r.varTable.GetVariable(i.(*typevars.Variable).String())
	}

	setVar := func(i typevars.Interface, item *varTableItem) {
		r.varTable.SetVariable(i.(*typevars.Variable).String(), item)
	}

	switch d := c.(type) {
	case *contracts.UnaryOp:
		xVarItem := getVar(d.X)

		yDataType, err := propagation.New(nil).UnaryExpr(
			d.OpToken,
			xVarItem.dataType,
		)
		if err != nil {
			return err
		}

		setVar(d.Y, &varTableItem{
			dataType:    yDataType,
			packageName: xVarItem.packageName,
			symbolTable: xVarItem.symbolTable,
		})

		// fmt.Println(contracts.Contract2String(d))
	case *contracts.BinaryOp:
		xVarItem := getVar(d.X)
		yVarItem := getVar(d.Y)

		// have the BinaryExpr return symbol table and the package name as well
		// it will be either builtin or the current package
		zDataType, err := propagation.New(nil).BinaryExpr(
			d.OpToken,
			xVarItem.dataType,
			yVarItem.dataType,
		)
		if err != nil {
			return err
		}

		setVar(d.Z, &varTableItem{
			dataType:    zDataType,
			packageName: xVarItem.packageName,
			symbolTable: xVarItem.symbolTable,
		})

		// fmt.Println(contracts.Contract2String(d))
	case *contracts.HasField:
		xVarItem := getVar(d.X)
		// fmt.Println(contracts.Contract2String(d))

		if d.Field != "" {
			yDataType, err := propagation.New(r.symbolAccessor).SelectorExpr(
				xVarItem.dataType,
				d.Field,
			)
			// fmt.Printf("yDataType: %#v, \terr: %v\n", yDataType, err)
			if err != nil {
				return err
			}
			r.varTable.SetField(d.X.(*typevars.Variable).String(), d.Field, &varTableItem{
				dataType: yDataType,
				// once RetrieveDataTypeField returns the symbol table and the package name, set the two below properly
				packageName: xVarItem.packageName,
				symbolTable: xVarItem.symbolTable,
			})
		} else {
			// retrieve positional field
			panic("retrieve positional field NYI")
		}

	case *contracts.IsCompatibleWith:
	//   if r.isTypevarEvaluated(d.X) && r.isTypevarEvaluated(d.Y) {
	//     ready = true
	//   }
	case *contracts.PropagatesTo:
		// fmt.Printf("PT.X: %#v\n", d.X)
		// fmt.Printf("PT.Y: %#v\n", d.Y)

		switch td := d.X.(type) {
		case *typevars.Constant:
			// The RHS is always a variable
			st, err := r.globalSymbolTable.Lookup(td.Package)
			if err != nil {
				panic(err)
			}
			setVar(d.Y, &varTableItem{
				dataType:    td.DataType,
				packageName: td.Package,
				symbolTable: st,
			})
		case *typevars.Variable:
			setVar(d.Y, getVar(d.X))
		case *typevars.Field:
			item, _ := r.varTable.GetField(td.X.String(), td.Name)
			setVar(d.Y, item)
		case *typevars.ReturnType:
			// fmt.Printf("d.X: %#v\n", td.Function)
			item := getVar(td.Function)
			// fmt.Printf("item: %#v\n", item)
			switch fDef := item.dataType.(type) {
			case *gotypes.Method:
				// fmt.Printf("item(%v): %#v\n", td.Index, fDef.Def.(*gotypes.Function).Results[td.Index])
				setVar(d.Y, &varTableItem{
					dataType:    fDef.Def.(*gotypes.Function).Results[td.Index],
					packageName: item.packageName,
					symbolTable: item.symbolTable,
				})
			case *gotypes.Function:
				// fmt.Printf("item(%v): %#v\n", td.Index, fDef.Results[td.Index])
				setVar(d.Y, &varTableItem{
					dataType:    fDef.Results[td.Index],
					packageName: item.packageName,
					symbolTable: item.symbolTable,
				})
			default:
				return fmt.Errorf("d.X of typevars.ReturnType expected to be a funtion/method, got %#v instead", item.dataType)
			}
		default:
			panic(fmt.Sprintf("Unrecognized typevar %#v during PropagatesTo evaluation", d.X))
		}
		// fmt.Println(contracts.Contract2String(d))
	case *contracts.IsInvocable:
		item := getVar(d.F)
		// TODO(jchaloup): call getFunctionDef in case the DataType is an identifier
		// fmt.Printf("d.F: %#v\n", item.dataType)
		if _, ok := item.dataType.(*gotypes.Method); ok {
			return nil
		}
		if _, ok := item.dataType.(*gotypes.Function); ok {
			return nil
		}
		return fmt.Errorf("d.F of contracts.IsInvocable expected to be a funtion/method, got %#v instead", item.dataType)
	default:
		panic(fmt.Sprintf("Unrecognized contract: %#v", c))
	}

	return nil
}

func (r *Runner) dumpTypeVars() {
	fmt.Printf("\n\n")
	r.varTable.Dump()
	fmt.Printf("\n\n")
}

func (r *Runner) Run() error {
	// propagate data type definitions of all entry typevars
	// for key, _ := range r.entryTypevars {
	// 	fmt.Printf("ev: %#v\n", key)
	// 	// if _, ok := r.varsTypes[key]; !ok {
	// 	// 	r.varsTypes[key] = nil
	// 	// 	// st, err := r.globalSymbolTable.Lookup(v.Package)
	// 	// 	// if err != nil {
	// 	// 	// 	return err
	// 	// 	// }
	// 	// 	// fmt.Printf("ST: %#v\n", st)
	// 	// }
	// }

	ready, unready := r.splitContracts(newContractPayload(r.waitingContracts.contracts()))
	for !ready.isEmpty() {
		for _, d := range ready.contracts() {
			for _, c := range d {
				if err := r.evaluateContract(c); err != nil {
					return nil
				}
			}
		}
		ready, unready = r.splitContracts(unready)
	}
	// r.dumpTypeVars()
	return nil
}

func (r *Runner) VarTable() *varTable {
	return r.varTable
}