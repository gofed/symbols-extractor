package compatibility

import (
	"fmt"
	"reflect"

	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

func InterfacesEqual(x, y *gotypes.Interface) bool {
	fmt.Printf("I1: %#v\n", x)
	fmt.Printf("I2: %#v\n", y)

	iXMethods := make(map[string]*gotypes.Function)
	iYMethods := make(map[string]*gotypes.Function)

	if len(x.Methods) != len(y.Methods) {
		return false
	}

	for _, m := range x.Methods {
		iXMethods[m.Name] = m.Def.(*gotypes.Function)
	}

	for _, m := range y.Methods {
		if _, ok := iXMethods[m.Name]; !ok {
			return false
		}
		iYMethods[m.Name] = m.Def.(*gotypes.Function)
	}

	for name := range iXMethods {
		fmt.Printf("%v:\n\t%#v\n\t%#v\n", name, iXMethods[name], iYMethods[name])
		if !FunctionSignaturesEqual(iXMethods[name], iYMethods[name]) {
			return false
		}
	}

	// Two interface types are identical if they have the same set of methods with the same names and identical function types. Non-exported method names from different packages are always different. The order of the methods is irrelevant.
	// TODO(jchaloup): what about embedded interfaces?

	return true
}

func TypeImplementsInterface(x gotypes.DataType, i *gotypes.Interface) bool {
	return false
}

func FunctionSignaturesEqual(x, y *gotypes.Function) bool {
	// compare params
	if len(x.Params) != len(y.Params) || len(x.Results) != len(y.Results) {
		return false
	}
	for i := range x.Params {
		fmt.Printf("%v:\n\t%#v\n\t%#v\n", i, x.Params[i], y.Params[i])
		if !DataTypesEqual(x.Params[i], y.Params[i]) {
			return false
		}
	}

	for i := range x.Results {
		fmt.Printf("%v:\n\t%#v\n\t%#v\n", i, x.Results[i], y.Results[i])
		if !DataTypesEqual(x.Results[i], y.Results[i]) {
			return false
		}
	}

	return true
}

func DataTypesEqual(x, y gotypes.DataType) bool {
	if x.GetType() != y.GetType() {
		return false
	}

	switch x.GetType() {
	case gotypes.IdentifierType:
		i1 := x.(*gotypes.Identifier)
		i2 := x.(*gotypes.Identifier)
		return i1.Package == i2.Package && i1.Def == i2.Def
	default:
		panic(fmt.Errorf("DataTypesEqual: NYI for %v", x.GetType()))
	}
}

// Untyped constant propagation translates to warning rather than a real check.
// Otherwise once would have to implement builtin types evaluation which is
// a runtime aspect of the analysis, not the static one.
// Only the syntax checks will get carried.
// Maybe, extend each warning with a Go code snippet that can be run so the user
// can check if an untyped constant would fit a types variable.

func IsXCompatibleWithY(xDataType, yDataType gotypes.DataType) (bool, error) {
	// Two data type are compatible when they are identical
	if reflect.DeepEqual(xDataType, yDataType) {
		return true, nil
	}

	// Untyped vs typed builtins
	if xDataType.GetType() == gotypes.BuiltinType && yDataType.GetType() == gotypes.BuiltinType {
		xBuiltin := xDataType.(*gotypes.Builtin)
		yBuiltin := yDataType.(*gotypes.Builtin)
		// Check all the categories:
		// - int
		// - float
		// - complex
		// ...
		if xBuiltin.Def == "float" && xBuiltin.Untyped {

		}
		if xBuiltin.Def == yBuiltin.Def {
			return true, nil
		}
		return false, fmt.Errorf("Two different builtin types: %v != %v", xBuiltin.Def, yBuiltin.Def)
	}

	fmt.Printf("X: %#v\n", xDataType)
	fmt.Printf("Y: %#v\n", yDataType)

	panic("IsCompatible NYI")
}
