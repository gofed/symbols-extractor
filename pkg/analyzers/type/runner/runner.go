package runner

import (
	"fmt"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/analyzers/type/propagation"
	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts"
	"github.com/gofed/symbols-extractor/pkg/parser/contracts/typevars"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

type Runner struct {
	v2c           *var2Contract
	entryTypevars map[string]*typevars.Variable
	// data type propagation through virtual variables
	varTable *VarTable

	// contracts
	waitingContracts *contractPayload

	globalSymbolTable      *global.Table
	allocatedSymbolsTable  map[string]*alloctable.Table
	globalAllocSymbolTable *allocglobal.Table
	packageName            string

	symbolAccessor *accessors.Accessor
}

func New(config *types.Config, globalAllocSymbolTable *allocglobal.Table) *Runner {
	r := &Runner{
		v2c:                    newVar2Contract(),
		entryTypevars:          make(map[string]*typevars.Variable, 0),
		varTable:               newVarTable(),
		waitingContracts:       newContractPayload(nil),
		globalSymbolTable:      config.GlobalSymbolTable,
		symbolAccessor:         accessors.NewAccessor(config.GlobalSymbolTable),
		globalAllocSymbolTable: globalAllocSymbolTable,
		packageName:            config.PackageName,
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
			case *contracts.TypecastsTo:
				if isVariable(d.X) {
					storeVar(d.X.(*typevars.Variable), c)
				}
				if isVariable(d.Y) {
					storeVar(d.Y.(*typevars.Variable), c)
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
		// is the variable a builtin function?
		// - make, len, complex, real, ...
		switch d.Package {
		case "builtin":
			switch d.Name {
			case "len", "make", "copy", "append", "panic", "recover", "print", "new", "println":
				return true
			}
		case "unsafe":
			switch d.Name {
			case "Sizeof", "Offsetof", "Alignof":
				return true
			}
		}

		v, ok := r.varTable.GetVariable(d.String())
		// glog.Infof("isTypevarEvaluated: %v, ok: %v", d.String(), ok)
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
	case *typevars.CGO:
		return true
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
			case *contracts.TypecastsTo:
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

	return readyPayload, unreadyPayload
}

func (r *Runner) evaluateContract(c contracts.Contract) error {
	getVar := func(i typevars.Interface) (*varTableItem, bool) {
		k, ok := r.varTable.GetVariable(i.(*typevars.Variable).String())
		// glog.Infof("Getting %v: %#v", i.(*typevars.Variable).String(), k)
		return k, ok
	}

	setVar := func(i typevars.Interface, item *varTableItem) {
		// glog.Infof("Setting %v to %#v", i.(*typevars.Variable).String(), item)
		r.varTable.SetVariable(i.(*typevars.Variable).String(), item)
	}

	// glog.Infof("Checking %v...", contracts.Contract2String(c))

	typevar2varTableItem := func(i typevars.Interface) (*varTableItem, error) {
		//glog.Infof("typevar2varTableItem, i=%#v", i)
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
			dt, err := r.symbolAccessor.FindFirstNonIdSymbol(item.dataType)
			if err != nil {
				return nil, err
			}

			switch fDef := dt.(type) {
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
				return nil, fmt.Errorf("typevars.ReturnType expected to be a funtion/method, got %#v instead", dt)
			}
		case *typevars.ListKey:
			st, err := r.globalSymbolTable.Lookup("builtin")
			if err != nil {
				panic(err)
			}
			return &varTableItem{
				dataType:    &gotypes.Constant{Package: "builtin", Def: "int", Untyped: true},
				packageName: "builtin",
				symbolTable: st,
			}, nil
		case *typevars.ListValue:
			item, ok := getVar(td.X)
			if !ok {
				return nil, fmt.Errorf("Variable %v does not exist", td.X.String())
			}
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
		case *typevars.CGO:
			st, err := r.globalSymbolTable.Lookup(td.Package)
			if err != nil {
				panic(err)
			}
			return &varTableItem{
				dataType:    td.DataType,
				packageName: td.Package,
				symbolTable: st,
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

		//glog.Infof("xVarItem: %#v, dataType: %#v", xVarItem, xVarItem.dataType)

		yDataType, err := propagation.New(r.symbolAccessor).UnaryExpr(
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

		//glog.Infof("xVarItem: %#v, dataType: %#v", xVarItem, xVarItem.dataType)
		//glog.Infof("yVarItem: %#v, dataType: %#v", yVarItem, yVarItem.dataType)

		// xDt, err := r.symbolAccessor.FindFirstNonIdSymbol(xVarItem.dataType)
		// if err != nil {
		// 	return err
		// }
		//
		// glog.Infof("xDt: %#v", xDt)
		//
		// yDt, err := r.symbolAccessor.FindFirstNonIdSymbol(yVarItem.dataType)
		// if err != nil {
		// 	return err
		// }
		//
		// glog.Infof("yDt: %#v", yDt)

		// have the BinaryExpr return symbol table and the package name as well
		// it will be either builtin or the current package
		zDataType, err := propagation.New(r.symbolAccessor).BinaryExpr(
			d.OpToken,
			xVarItem.dataType,
			yVarItem.dataType,
		)
		if err != nil {
			return err
		}
		//glog.Infof("zDataType: %#v\n", zDataType)
		setVar(d.Z, &varTableItem{
			dataType:    zDataType,
			packageName: xVarItem.packageName,
			symbolTable: xVarItem.symbolTable,
		})
	case *contracts.HasField:
		xVarItem, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		if d.Field != "" {
			fieldAttribute, err := propagation.New(r.symbolAccessor).SelectorExpr(
				xVarItem.dataType,
				d.Field,
			)
			if err != nil {
				panic(err)
				return err
			}
			yDataType := fieldAttribute.DataType

			r.varTable.SetField(d.X.(*typevars.Variable).String(), d.Field, &varTableItem{
				dataType: yDataType,
				// once RetrieveDataTypeField returns the symbol table and the package name, set the two below properly
				packageName: xVarItem.packageName,
				symbolTable: xVarItem.symbolTable,
			})

			// method of a struct or a data type?
			if method, ok := yDataType.(*gotypes.Method); ok {
				switch recvr := method.Receiver.(type) {
				case *gotypes.Pointer:
					ident, ok := recvr.Def.(*gotypes.Identifier)
					if ok {
						allocTable, err := r.globalAllocSymbolTable.Lookup(r.packageName, strings.Split(d.Pos, ":")[0])
						if err != nil {
							return nil
						}
						// TODO(jchaloup): pick the entry data type instead of the data type that actually defines the method
						// var fieldCache struct {
						//         sync.RWMutex
						//         m map[reflect.Type][]field
						// }
						//
						// fieldCache.RLock() is actually a call of the RLock through the anonymous struct
						allocTable.AddMethod(ident.Package, ident.Def, d.Field, d.Pos)
					} else {
						return fmt.Errorf("Receiver expected to be a pointer to identifier or an identifier, got pointer to %#v instead", recvr.Def)
					}
				case *gotypes.Identifier:
					allocTable, err := r.globalAllocSymbolTable.Lookup(r.packageName, strings.Split(d.Pos, ":")[0])
					if err != nil {
						return nil
					}
					// TODO(jchaloup): pick the entry data type instead of the data type that actually defines the method
					// var fieldCache struct {
					//         sync.RWMutex
					//         m map[reflect.Type][]field
					// }
					//
					// fieldCache.RLock() is actually a call of the RLock through the anonymous struct
					allocTable.AddMethod(recvr.Package, recvr.Def, d.Field, d.Pos)
				default:
					return fmt.Errorf("Receiver expected to be a pointer to identifier or an identifier, got %#v instead", method.Receiver)
				}
			} else {
				// struct field?
				if !fieldAttribute.IsMethod {
					found := false
					for i := len(fieldAttribute.Origin) - 1; i >= 0; i-- {
						item := fieldAttribute.Origin[i]
						if item.Def.GetType() == gotypes.StructType {
							allocTable, err := r.globalAllocSymbolTable.Lookup(r.packageName, strings.Split(d.Pos, ":")[0])
							if err != nil {
								return nil
							}
							allocTable.AddStructField(item.Package, item.Name, d.Field, d.Pos)
							found = true
							break
						}
					}
					if !found {
						panic("Field origin not found")
					}
				}
			}
		} else {
			dt, err := r.symbolAccessor.FindFirstNonidDataType(xVarItem.dataType)
			if err != nil {
				panic(err)
				return err
			}
			// pointer? Maybe
			// TODO(jchaloup): only one pointer should be allowed
			pointer, ok := dt.(*gotypes.Pointer)
			for {
				pointer, ok = dt.(*gotypes.Pointer)
				if !ok {
					break
				}
				dt = pointer.Def
			}

			dt, err = r.symbolAccessor.FindFirstNonidDataType(dt)
			if err != nil {
				panic(err)
				return err
			}

			structDef, ok := dt.(*gotypes.Struct)
			if !ok {
				return fmt.Errorf("Trying to retrieve field at index %v from non-struct data type %#v", d.Index, dt)
			}
			// retrieve positional field
			field, err := r.symbolAccessor.RetrieveStructFieldAtIndex(structDef, d.Index)
			if err != nil {
				panic(err)
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
		item, err := typevar2varTableItem(d.X)
		if err != nil {
			return err
		}

		if d.ToVariable {
			if constant, ok := item.dataType.(*gotypes.Constant); ok {
				item = &varTableItem{
					dataType:    &gotypes.Identifier{Package: constant.Package, Def: constant.Def},
					symbolTable: item.symbolTable,
					packageName: item.packageName,
					entry:       item.entry,
				}
			}
		}

		setVar(d.Y, item)
		// fmt.Println(contracts.Contract2String(d))
	case *contracts.IsInvocable:
		item, xErr := typevar2varTableItem(d.F)
		if xErr != nil {
			return xErr
		}
		// count the variable is allocated
		if item.entry {
			variable, ok := d.F.(*typevars.Variable)
			if !ok {
				panic("Expected variable")
			}
			allocTable, err := r.globalAllocSymbolTable.Lookup(r.packageName, strings.Split(variable.Pos, ":")[0])
			if err != nil {
				return nil
			}
			allocTable.AddFunction(variable.Package, variable.Name, variable.Pos)
		}
		dt, err := r.symbolAccessor.FindFirstNonIdSymbol(item.dataType)
		if err != nil {
			return err
		}

		// TODO(jchaloup): call getFunctionDef in case the DataType is an identifier
		if _, ok := dt.(*gotypes.Method); ok {
			return nil
		}
		if _, ok := dt.(*gotypes.Function); ok {
			return nil
		}
		return fmt.Errorf("d.F of contracts.IsInvocable expected to be a funtion/method, got %#v instead", dt)
	case *contracts.IsIndexable:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		dt, err := r.symbolAccessor.FindFirstNonidDataType(item.dataType)
		if err != nil {
			return err
		}

		if pointer, ok := dt.(*gotypes.Pointer); ok {
			dt, err = r.symbolAccessor.FindFirstNonidDataType(pointer.Def)
			if err != nil {
				return err
			}
		}

		// pkg.A = []string, pkg.Index int = 0
		// pkg.A = map[string]string, pkg.Index string = "0"
		// as long as I use the pkg.A as:
		//   pkg.A[pkg.Index]
		// the change is backward compatible
		if _, ok := dt.(*gotypes.Ellipsis); ok {
			return nil
		}
		if _, ok := dt.(*gotypes.Slice); ok {
			// Both min:max must be compatible with Integer type which is checked in the
			// IsCompatibleWith(ListKey, Min/Max) contract
			return nil
		}
		if _, ok := dt.(*gotypes.Map); ok {
			// This is the most benevolent case, index type can change to basically anything
			// that can be an index, not just an Integer type
			return nil
		}
		if _, ok := dt.(*gotypes.Array); ok {
			if d.IsSlice {
				return nil
			}
			// The index must be an Integer type
			indexItem, err := typevar2varTableItem(d.Key)
			if err != nil {
				return err
			}
			// TODO(jchaloup): check the index is compatible with Integer type
			fmt.Printf("Checking index type %#v compatibility with Integer type not yet implemented", indexItem)
			return nil
		}
		switch d := dt.(type) {
		case *gotypes.Builtin:
			if d.Def == "string" {
				// TODO(jchaloup): check the index is compatible with Integer type
				return nil
			}
		case *gotypes.Constant:
			if d.Package == "builtin" && d.Def == "string" {
				// TODO(jchaloup): check the index is compatible with Integer type
				return nil
			}
		case *gotypes.Identifier:
			if d.Package == "builtin" && d.Def == "string" {
				// TODO(jchaloup): check the index is compatible with Integer type
				return nil
			}
		default:
			panic(fmt.Errorf("Unsupported indexable typevar %#v", dt))
		}
		// error
		fmt.Printf("Unsupported indexable typevar %#v", dt)
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

		channel, err := r.symbolAccessor.FindFirstNonIdSymbol(item.dataType)
		if err != nil {
			return err
		}

		if channel.GetType() != gotypes.ChannelType {
			return fmt.Errorf("Expected channel, got %#v instead", channel)
		}
		// TODO(jchaloup): check the direction as well
		setVar(d.Y, &varTableItem{
			dataType:    channel.(*gotypes.Channel).Value,
			packageName: item.packageName,
			symbolTable: item.symbolTable,
		})
	case *contracts.IsSendableTo:
		yItem, yErr := typevar2varTableItem(d.Y)
		if yErr != nil {
			return yErr
		}
		channel, err := r.symbolAccessor.FindFirstNonIdSymbol(yItem.dataType)
		if err != nil {
			return err
		}
		if channel.GetType() != gotypes.ChannelType {
			return fmt.Errorf("Expected channel, got %#v instead", channel)
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
		fmt.Printf("About to check %#v can be inc/decremented\n", xItem)
	case *contracts.IsRangeable:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		dt, err := r.symbolAccessor.FindFirstNonidDataType(item.dataType)
		if err != nil {
			return err
		}

		if pointer, ok := dt.(*gotypes.Pointer); ok {
			dt, err = r.symbolAccessor.FindFirstNonidDataType(pointer.Def)
			if err != nil {
				return err
			}
		}

		switch dt.GetType() {
		case gotypes.SliceType, gotypes.MapType, gotypes.ArrayType, gotypes.EllipsisType, gotypes.ChannelType:
			return nil
		case gotypes.BuiltinType:
			if dt.(*gotypes.Builtin).Def == "string" {
				return nil
			}
			return fmt.Errorf("Builtin expected to be a string, got %v instead", dt.(*gotypes.Builtin).Def)
		case gotypes.IdentifierType:
			ident := dt.(*gotypes.Identifier)
			if ident.Package == "builtin" && ident.Def == "string" {
				return nil
			}
			return fmt.Errorf("Builtin expected to be a string, got %v.%v instead", ident.Package, ident.Def)
		case gotypes.ConstantType:
			ident := dt.(*gotypes.Constant)
			if ident.Package == "builtin" && ident.Def == "string" {
				return nil
			}
			return fmt.Errorf("Builtin expected to be a string, got %v.%v instead", ident.Package, ident.Def)
		default:
			return fmt.Errorf("Expression %#v not rangeable", dt)
		}
	case *contracts.TypecastsTo:
		item, xErr := typevar2varTableItem(d.X)
		if xErr != nil {
			return xErr
		}
		// TODO(jchaloup): check the X can be type-casted to Type
		fmt.Printf("TODO: need to check %#v is compatible with %#v\n", item, d.Type)

		// d.Type must be always a constant
		tc, ok := d.Type.(*typevars.Constant)
		if !ok {
			return fmt.Errorf("Expected constant when type-casting, got %#v instead", d.Type)
		}

		// get type's origin
		dt := tc.DataType
		var dtOrigin string
		pointer, ok := dt.(*gotypes.Pointer)
		for {
			pointer, ok = dt.(*gotypes.Pointer)
			if !ok {
				break
			}
			dt = pointer.Def
		}

		switch d := dt.(type) {
		case *gotypes.Identifier:
			dtOrigin = d.Package
		case *gotypes.Selector:
			qid, ok := d.Prefix.(*gotypes.Packagequalifier)
			if !ok {
				fmt.Printf("slector not qid: %#v\n", d.Prefix)
				panic("SELECTOR not QID!!!")
			}
			dtOrigin = qid.Path
		default:
			dtOrigin = r.packageName
		}

		st, err := r.globalSymbolTable.Lookup(r.packageName)
		if err != nil {
			return err
		}

		castedDef, err := propagation.New(r.symbolAccessor).TypecastExpr(item.dataType, tc.DataType)
		yItem := &varTableItem{
			dataType:    castedDef,
			packageName: dtOrigin,
			symbolTable: st,
		}
		//glog.Infof("yItem: %#v, yItem.dataType: %#v\n", yItem, yItem.dataType)
		setVar(d.Y, yItem)
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
	for _, ev := range r.entryTypevars {
		st, err := r.globalSymbolTable.Lookup(ev.Package)
		if err != nil {
			return err
		}
		sDef, _, sErr := st.Lookup(ev.Name)
		if sErr != nil {
			return sErr
		}
		r.varTable.SetVariable(ev.String(), &varTableItem{
			dataType:    sDef.Def,
			packageName: ev.Package,
			symbolTable: st,
			entry:       true,
		})
	}

	ready, unready := r.splitContracts(newContractPayload(r.waitingContracts.contracts()))
	for !ready.isEmpty() {
		fmt.Printf("Ready:\n")
		ready.dump()
		fmt.Printf("Unready:\n")
		unready.dump()
		for _, cs := range ready.contracts() {
			for _, c := range cs {
				//for _, c := range ready.sortedContracts() {
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
		unready.dump()
		return fmt.Errorf("There are still some unprocessed contract: %v", unready.len())
	}

	return nil
}

func (r *Runner) VarTable() *VarTable {
	return r.varTable
}
