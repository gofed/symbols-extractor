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
	"github.com/golang/glog"
)

type Runner struct {
	v2c           *var2Contract
	entryTypevars map[string]*typevars.Variable
	// data type propagation through virtual variables
	varTable *VarTable

	// contracts
	waitingContracts *contractPayload

	globalSymbolTable *global.Table

	symbolAccessor *accessors.Accessor
}

func New(config *types.Config) *Runner {
	r := &Runner{
		v2c:               newVar2Contract(),
		entryTypevars:     make(map[string]*typevars.Variable, 0),
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
			case *contracts.IsIndexable:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
				if d.Key != nil && isVariable(d.Key) {
					storeVar(d.Key.(*typevars.Variable), c)
				}
			case *contracts.IsDereferenceable:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.DereferenceOf:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.IsReferenceable:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.ReferenceOf:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.IsReceiveableFrom:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.IsSendableTo:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
				if isVariable(d.Y) {
					storeVar(d.Y.(*typevars.Variable), c)
				}
			case *contracts.IsIncDecable:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
			case *contracts.IsRangeable:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
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
		v, ok := r.varTable.GetVariable(d.String())
		if !ok {
			return false
		}
		return v != nil
	case *typevars.Field:
		if r.isTypevarEvaluated(d.X) {
			if d.Name != "" {
				_, exists := r.varTable.GetField(d.X.String(), d.Name)
				return exists
			}
			_, exists := r.varTable.GetFieldAt(d.X.String(), d.Index)
			return exists
		}
		return false
	case *typevars.ReturnType:
		return r.isTypevarEvaluated(d.Function)
	case *typevars.Argument:
		return r.isTypevarEvaluated(d.Function)
	case *typevars.ListValue:
		return r.isTypevarEvaluated(d.X)
	case *typevars.ListKey:
		return true
	case *typevars.MapKey:
		return r.isTypevarEvaluated(d.X)
	case *typevars.MapValue:
		return r.isTypevarEvaluated(d.X)
	case *typevars.RangeKey:
		return r.isTypevarEvaluated(d.X)
	case *typevars.RangeValue:
		return r.isTypevarEvaluated(d.X)
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
			case *contracts.IsIndexable:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
				if d.Key != nil && !r.isTypevarEvaluated(d.Key) {
					ready = false
				}
			case *contracts.IsDereferenceable:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.DereferenceOf:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.IsReferenceable:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.ReferenceOf:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.IsReceiveableFrom:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.IsSendableTo:
				if r.isTypevarEvaluated(d.X) && r.isTypevarEvaluated(d.Y) {
					ready = true
				}
			case *contracts.IsIncDecable:
				if r.isTypevarEvaluated(d.X) {
					ready = true
				}
			case *contracts.IsRangeable:
				if r.isTypevarEvaluated(d.X) {
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
	getVar := func(i typevars.Interface) (*varTableItem, bool) {
		k, ok := r.varTable.GetVariable(i.(*typevars.Variable).String())
		return k, ok
	}

	setVar := func(i typevars.Interface, item *varTableItem) {
		glog.Infof("Setting %v to %#v", i.(*typevars.Variable).String(), item)
		r.varTable.SetVariable(i.(*typevars.Variable).String(), item)
	}

	glog.Infof("Checking %v...", contracts.Contract2String(c))

	typevar2varTableItem := func(i typevars.Interface) (*varTableItem, error) {
		switch td := i.(type) {
		case *typevars.Constant:
			st, err := r.globalSymbolTable.Lookup(td.Package)
			if err != nil {
				panic(err)
			}
			return &varTableItem{
				dataType:    td.DataType,
				packageName: td.Package,
				symbolTable: st,
			}, nil
		case *typevars.Variable:
			v, ok := getVar(td)
			if ok {
				return v, nil
			}
			return nil, fmt.Errorf("Variable %v does not exist", td.String())
		case *typevars.Field:
			item, _ := r.varTable.GetField(td.X.String(), td.Name)
			return item, nil
		case *typevars.ReturnType:
			item, ok := getVar(td.Function)
			if !ok {
				return nil, fmt.Errorf("Variable %v does not exist", td.Function.String())
			}
			// fmt.Printf("item: %#v\n", item)
			switch fDef := item.dataType.(type) {
			case *gotypes.Method:
				return &varTableItem{
					dataType:    fDef.Def.(*gotypes.Function).Results[td.Index],
					packageName: item.packageName,
					symbolTable: item.symbolTable,
				}, nil
			case *gotypes.Function:
				return &varTableItem{
					dataType:    fDef.Results[td.Index],
					packageName: item.packageName,
					symbolTable: item.symbolTable,
				}, nil
			default:
				return nil, fmt.Errorf("typevars.ReturnType expected to be a funtion/method, got %#v instead", item.dataType)
			}
		case *typevars.ListValue:
			item, ok := getVar(td.X)
			if !ok {
				return nil, fmt.Errorf("Variable %v does not exist", td.X.String())
			}
			fmt.Printf("item: %#v, ok: %v\n", item, ok)
			yDataType, _, err := propagation.New(r.symbolAccessor).IndexExpr(
				item.dataType,
				nil,
			)
			if err != nil {
				return nil, err
			}
			return &varTableItem{
				dataType:    yDataType,
				packageName: item.packageName,
				symbolTable: item.symbolTable,
			}, nil
		case *typevars.MapValue:
			item, ok := getVar(td.X)
			if !ok {
				return nil, fmt.Errorf("Variable %v does not exist", td.X.String())
			}
			fmt.Printf("item: %#v, ok: %v\n", item, ok)
			yDataType, _, err := propagation.New(r.symbolAccessor).IndexExpr(
				item.dataType,
				nil,
			)
			if err != nil {
				return nil, err
			}
			return &varTableItem{
				dataType:    yDataType,
				packageName: item.packageName,
				symbolTable: item.symbolTable,
			}, nil
		case *typevars.RangeKey:
			item, ok := getVar(td.X)
			if !ok {
				return nil, fmt.Errorf("Variable %v does not exist", td.X.String())
			}
			keyDT, _, err := propagation.New(r.symbolAccessor).RangeExpr(item.dataType)
			if err != nil {
				return nil, err
			}
			return &varTableItem{
				dataType:    keyDT,
				packageName: item.packageName,
				symbolTable: item.symbolTable,
			}, nil
		case *typevars.RangeValue:
			item, ok := getVar(td.X)
			if !ok {
				return nil, fmt.Errorf("Variable %v does not exist", td.X.String())
			}
			_, valueDT, err := propagation.New(r.symbolAccessor).RangeExpr(item.dataType)
			if err != nil {
				return nil, err
			}
			return &varTableItem{
				dataType:    valueDT,
				packageName: item.packageName,
				symbolTable: item.symbolTable,
			}, nil
		default:
			panic(fmt.Sprintf("Unrecognized typevar %#v", i))
		}
	}

	switch d := c.(type) {
	case *contracts.UnaryOp:
		xVarItem, err := typevar2varTableItem(d.X)
		if err != nil {
			return err
		}

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
		xVarItem, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		yVarItem, yErr := typevar2varTableItem(d.Y)
		if yErr != nil {
			return yErr
		}

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
		xVarItem, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		// fmt.Println(contracts.Contract2String(d))

		if d.Field != "" {
			fmt.Printf("xVarItem: %#v\n", xVarItem)
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
			structDef, ok := xVarItem.dataType.(*gotypes.Struct)
			if !ok {
				return fmt.Errorf("Trying to retrieve field at index %v from non-struct data type %#v", d.Index, xVarItem.dataType)
			}
			// retrieve positional field
			field, err := r.symbolAccessor.RetrieveStructFieldAtIndex(structDef, d.Index)
			if err != nil {
				return err
			}
			r.varTable.SetFieldAt(d.X.(*typevars.Variable).String(), d.Index, &varTableItem{
				dataType:    field,
				packageName: xVarItem.packageName,
				symbolTable: xVarItem.symbolTable,
			})
		}
	case *contracts.IsCompatibleWith:
	case *contracts.PropagatesTo:
		// fmt.Printf("PT.X: %#v\n", d.X)
		// fmt.Printf("PT.Y: %#v\n", d.Y)

		item, err := typevar2varTableItem(d.X)
		if err != nil {
			return err
		}

		setVar(d.Y, item)
		// fmt.Println(contracts.Contract2String(d))
	case *contracts.IsInvocable:
		item, xErr := typevar2varTableItem(d.F)
		if xErr != nil {
			return xErr
		}
		// TODO(jchaloup): call getFunctionDef in case the DataType is an identifier
		// fmt.Printf("d.F: %#v\n", item.dataType)
		if _, ok := item.dataType.(*gotypes.Method); ok {
			return nil
		}
		if _, ok := item.dataType.(*gotypes.Function); ok {
			return nil
		}
		return fmt.Errorf("d.F of contracts.IsInvocable expected to be a funtion/method, got %#v instead", item.dataType)
	case *contracts.IsIndexable:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		// pkg.A = []string, pkg.Index int = 0
		// pkg.A = map[string]string, pkg.Index string = "0"
		// as long as I use the pkg.A as:
		//   pkg.A[pkg.Index]
		// the change is backward compatible
		if _, ok := item.dataType.(*gotypes.Slice); ok {
			// Both min:max must be compatible with Integer type which is checked in the
			// IsCompatibleWith(ListKey, Min/Max) contract
			return nil
		}
		if _, ok := item.dataType.(*gotypes.Map); ok {
			// This is the most benevolent case, index type can change to basically anything
			// that can be an index, not just an Integer type
			return nil
		}
		if _, ok := item.dataType.(*gotypes.Array); ok {
			// The index must be an Integer type
			indexItem, err := typevar2varTableItem(d.Key)
			if err != nil {
				return err
			}
			// TODO(jchaloup): check the index is compatible with Integer type
			fmt.Printf("Checking index type %#v compatibility with Integer type not yet implemented", indexItem)
			return nil
		}
		if d, ok := item.dataType.(*gotypes.Builtin); ok {
			if d.Def == "string" {
				// TODO(jchaloup): check the index is compatible with Integer type
				return nil
			}
			// error
		}
		// error
		fmt.Printf("item: %#v\n", item)
		panic("|||")
	case *contracts.IsDereferenceable:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		if item.dataType.GetType() != gotypes.PointerType {
			return fmt.Errorf("Expected pointer, got %#v instead", item.dataType)
		}
	case *contracts.DereferenceOf:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		setVar(d.Y, &varTableItem{
			dataType:    item.dataType.(*gotypes.Pointer).Def,
			packageName: item.packageName,
			symbolTable: item.symbolTable,
		})
	case *contracts.IsReferenceable:
		// TODO(jchaloup): is there anything to check?
	case *contracts.ReferenceOf:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		setVar(d.Y, &varTableItem{
			dataType:    &gotypes.Pointer{Def: item.dataType},
			packageName: item.packageName,
			symbolTable: item.symbolTable,
		})
	case *contracts.IsReceiveableFrom:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		if item.dataType.GetType() != gotypes.ChannelType {
			return fmt.Errorf("Expected channel, got %#v instead", item.dataType)
		}
		// TODO(jchaloup): check the direction as well
		setVar(d.Y, &varTableItem{
			dataType:    item.dataType.(*gotypes.Channel).Value,
			packageName: item.packageName,
			symbolTable: item.symbolTable,
		})
	case *contracts.IsSendableTo:
		yItem, yErr := typevar2varTableItem(d.Y)
		if yErr != nil {
			return yErr
		}
		if yItem.dataType.GetType() != gotypes.ChannelType {
			return fmt.Errorf("Expected channel, got %#v instead", yItem.dataType)
		}
		xItem, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		// TODO(chaloup): check the xItem (value) is compatible with yItem.Value
		fmt.Printf("Checking item %#v compatibility with channel", xItem)
	case *contracts.IsIncDecable:
		xItem, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		// TODO(jchaloup): Check the
		fmt.Printf("About to check %#v can be inc/decremented", xItem)
	case *contracts.IsRangeable:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		switch item.dataType.GetType() {
		case gotypes.SliceType, gotypes.MapType, gotypes.ArrayType:
			return nil
		case gotypes.BuiltinType:
			if item.dataType.(*gotypes.Builtin).Def == "string" {
				return nil
			}
			return fmt.Errorf("Builtin expected to be a string, got %v instead", item.dataType.(*gotypes.Builtin).Def)
		default:
			return fmt.Errorf("Expression %#v not rangeable", item.dataType)
		}
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
	for key, ev := range r.entryTypevars {
		fmt.Printf("ev: %#v\n", key)
		st, err := r.globalSymbolTable.Lookup(ev.Package)
		if err != nil {
			return err
		}
		sDef, _, sErr := st.LookupVariableLikeSymbol(ev.Name)
		if sErr != nil {
			return sErr
		}
		r.varTable.SetVariable(ev.String(), &varTableItem{
			dataType:    sDef.Def,
			packageName: ev.Package,
			symbolTable: st,
		})
	}

	ready, unready := r.splitContracts(newContractPayload(r.waitingContracts.contracts()))
	for !ready.isEmpty() {
		fmt.Printf("Ready:\n")
		ready.dump()
		fmt.Printf("Unready:\n")
		unready.dump()
		for _, d := range ready.contracts() {
			for _, c := range d {
				if err := r.evaluateContract(c); err != nil {
					return err
				}
			}
		}
		ready, unready = r.splitContracts(unready)
	}
	fmt.Printf("Ready:\n")
	ready.dump()
	fmt.Printf("Unready:\n")
	unready.dump()
	if !unready.isEmpty() {
		return fmt.Errorf("There are still some unprocessed contract")
	}

	return nil
}

func (r *Runner) VarTable() *VarTable {
	return r.varTable
}
