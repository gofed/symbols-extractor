package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"io/ioutil"
	"path"
	"reflect"
	"runtime"
	"strings"

	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	"github.com/gofed/symbols-extractor/pkg/snapshots"
	"github.com/gofed/symbols-extractor/pkg/snapshots/glide"
	"github.com/gofed/symbols-extractor/pkg/snapshots/godeps"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"

	"github.com/golang/glog"
)

//////////
// checkapi --allocated <ALLOCATED>.json --package-path <EXERCISED_DEPENDENCY> --symbol-table-dir <PACKAGE_APIS> --cgo-symbols-path <CGO>
//
// The --symbol-table-dir is a `:` separated list of paths with generated directory
//
// E.g.
// checkapi --alocated toml.json --package-path github.com/coreos/etcd:commit --symbol-table-dir generated:updates/1.10/generated
//

type flags struct {
	allocated           *string
	allocatedGlidefile  *string
	allocatedGodepsfile *string
	packagePrefix       *string
	packageCommit       *string
	symbolTablePath     *string
	cgoSymbolsPath      *string
	goVersion           *string
	glidefile           *string
	godepsfile          *string
}

func (f *flags) parse() error {
	flag.Parse()

	if *(f.allocated) == "" {
		return fmt.Errorf("--allocated is not set")
	}

	if *(f.packagePrefix) == "" {
		return fmt.Errorf("--package-prefix is not set")
	}

	if *(f.packageCommit) == "" {
		return fmt.Errorf("--package-commit is not set")
	}

	if *(f.symbolTablePath) == "" {
		return fmt.Errorf("--symbol-table-dir is not set")
	}

	if *(f.goVersion) == "" {
		return fmt.Errorf("--go-version is not set")
	}

	if *(f.glidefile) == "" && *f.godepsfile == "" {
		return fmt.Errorf("-glidefile or -godepsfile is not set")
	}

	return nil
}

func initRefGlobaST(f *flags) (*global.Table, snapshots.Snapshot, error) {
	if *f.allocatedGlidefile != "" {
		snapshot, err := glide.GlideFromFile(*f.allocatedGlidefile)
		if err != nil {
			return nil, nil, err
		}
		return global.New(*(f.symbolTablePath), *(f.goVersion), snapshot), snapshot, nil
	} else if *f.allocatedGodepsfile != "" {
		snapshot, err := godeps.FromFile(*f.allocatedGodepsfile)
		if err != nil {
			return nil, nil, err
		}
		return global.New(*(f.symbolTablePath), *(f.goVersion), snapshot), snapshot, nil
	}
	return global.New(*(f.symbolTablePath), *(f.goVersion), nil), nil, nil
}

func initExercisedGlobaST(f *flags) (*global.Table, snapshots.Snapshot, error) {
	if *f.glidefile != "" {
		snapshot, err := glide.GlideFromFile(*f.glidefile)
		if err != nil {
			return nil, nil, err
		}
		snapshot.MainPackageCommit(*f.packagePrefix, *f.packageCommit)
		return global.New(*(f.symbolTablePath), *(f.goVersion), snapshot), snapshot, nil
	}
	snapshot, err := godeps.FromFile(*f.godepsfile)
	if err != nil {
		return nil, nil, err
	}
	snapshot.MainPackageCommit(*f.packagePrefix, *f.packageCommit)
	return global.New(*(f.symbolTablePath), *(f.goVersion), snapshot), snapshot, nil
}

type SymbolInfo struct {
	Package string
	Parent  string
	Name    string
	Pos     []string
}

func (s *SymbolInfo) str() string {
	if s.Parent == "" {
		return fmt.Sprintf("%v.%v", s.Package, s.Name)
	}
	return fmt.Sprintf("%v.%v.%v", s.Package, s.Parent, s.Name)
}

type ApiDiff struct {
	DatatypesMissing    []SymbolInfo
	Datatypes           []SymbolInfo
	FunctionsMissing    []SymbolInfo
	Functions           []SymbolInfo
	VariablesMissing    []SymbolInfo
	Variables           []SymbolInfo
	MethodsMissing      []SymbolInfo
	Methods             []SymbolInfo
	StructFieldsMissing []SymbolInfo
	StructFields        []SymbolInfo
}

type ident struct {
	pkg, name, field string
}

func (i ident) str() string {
	if i.field == "" {
		return fmt.Sprintf("%v.%v", i.pkg, i.name)
	}
	return fmt.Sprintf("%v.%v.%v", i.pkg, i.name, i.field)
}

type typeSymbols map[ident][]string

func collectApiDiffs(tables map[string]allocglobal.PackageTable, packagePrefix string, refGlobalST, exercisedGlobalST *global.Table) (*ApiDiff, error) {
	apidiff := ApiDiff{}

	refAccessor := accessors.NewAccessor(refGlobalST)
	exerAccessor := accessors.NewAccessor(exercisedGlobalST)

	dtNames := make(typeSymbols, 0)
	fncNames := make(typeSymbols, 0)
	varNames := make(typeSymbols, 0)
	fieldNames := make(typeSymbols, 0)
	methodNames := make(typeSymbols, 0)

	for tablePkg, table := range tables {
		for file, fileItem := range table {
			glog.V(1).Infof("Processing %q of %q\n", file, tablePkg)
			for pkg, symbolsSet := range fileItem.Symbols {
				// skip anonymous symbols
				if pkg == "" {
					continue
				}

				if !strings.HasPrefix(pkg, packagePrefix) {
					continue
				}

				glog.V(1).Infof("Processing %q\n", pkg)

				// Datatypes
				for _, item := range symbolsSet.Datatypes {
					key := ident{pkg: pkg, name: item.Name}
					if _, ok := dtNames[key]; ok {
						dtNames[key] = append(dtNames[key], path.Join(fileItem.Package, item.Pos))
						continue
					}
					if !ast.IsExported(item.Name) {
						continue
					}
					dtNames[key] = []string{path.Join(fileItem.Package, item.Pos)}
				}
				// Functions
				for _, item := range symbolsSet.Functions {
					key := ident{pkg: pkg, name: item.Name}
					if _, ok := fncNames[key]; ok {
						fncNames[key] = append(fncNames[key], path.Join(fileItem.Package, item.Pos))
						continue
					}
					if !ast.IsExported(item.Name) {
						continue
					}
					fncNames[key] = []string{path.Join(fileItem.Package, item.Pos)}
				}
				// Variables
				for _, item := range symbolsSet.Variables {
					key := ident{pkg: pkg, name: item.Name}
					if _, ok := varNames[key]; ok {
						varNames[key] = append(varNames[key], path.Join(fileItem.Package, item.Pos))
						continue
					}
					if !ast.IsExported(item.Name) {
						continue
					}
					varNames[key] = []string{path.Join(fileItem.Package, item.Pos)}
				}
				// Struct fields
				for _, item := range symbolsSet.Structfields {
					key := ident{pkg: pkg, name: item.Parent, field: item.Field}
					if _, ok := fieldNames[key]; ok {
						fieldNames[key] = append(fieldNames[key], path.Join(fileItem.Package, item.Pos))
						continue
					}
					if !ast.IsExported(item.Parent) {
						continue
					}
					if !ast.IsExported(item.Field) {
						continue
					}
					fieldNames[key] = []string{path.Join(fileItem.Package, item.Pos)}
				}
				// Methods
				for _, item := range symbolsSet.Methods {
					key := ident{pkg: pkg, name: item.Parent, field: item.Name}
					if _, ok := methodNames[key]; ok {
						methodNames[key] = append(methodNames[key], path.Join(fileItem.Package, item.Pos))
						continue
					}
					if !ast.IsExported(item.Parent) {
						continue
					}
					if !ast.IsExported(item.Name) {
						continue
					}
					methodNames[key] = []string{path.Join(fileItem.Package, item.Pos)}
				}
			}
		}
	}

	for symbolItem, positions := range dtNames {
		glog.V(1).Infof("Checking %q type", symbolItem.str())
		refSTable, err := refGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			return nil, err
		}
		refSDef, err := refSTable.LookupDataType(symbolItem.name)
		if err != nil {
			return nil, fmt.Errorf("Type %q of %q package not found", symbolItem.name, symbolItem.pkg)
		}

		exerSTable, err := exercisedGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			apidiff.DatatypesMissing = append(apidiff.DatatypesMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
			continue
		}
		exerSDef, err := exerSTable.LookupDataType(symbolItem.name)
		if err != nil {
			apidiff.DatatypesMissing = append(apidiff.DatatypesMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
			continue
		}

		// Compare both symbols
		if !reflect.DeepEqual(refSDef.Def, exerSDef.Def) {
			apidiff.Datatypes = append(apidiff.Datatypes, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
		}
	}

	for symbolItem, positions := range fncNames {
		glog.V(1).Infof("Checking %q function", symbolItem.str())
		refSTable, err := refGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			return nil, err
		}
		refSDef, err := refSTable.LookupFunction(symbolItem.name)
		if err != nil {
			return nil, fmt.Errorf("Function %q of %q package not found", symbolItem.name, symbolItem.pkg)
		}

		exerSTable, err := exercisedGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			apidiff.FunctionsMissing = append(apidiff.FunctionsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
			continue
		}
		exerSDef, err := exerSTable.LookupFunction(symbolItem.name)
		if err != nil {
			apidiff.FunctionsMissing = append(apidiff.FunctionsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
			continue
		}

		// Compare both symbols
		if !reflect.DeepEqual(refSDef.Def, exerSDef.Def) {
			apidiff.Functions = append(apidiff.Functions, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
		}
	}

	for symbolItem, positions := range varNames {
		glog.V(1).Infof("Checking %q variable", symbolItem.str())
		refSTable, err := refGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			return nil, err
		}
		refSDef, err := refSTable.LookupVariable(symbolItem.name)
		if err != nil {
			return nil, fmt.Errorf("Type %q of %q package not found", symbolItem.name, symbolItem.pkg)
		}

		exerSTable, err := exercisedGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			apidiff.VariablesMissing = append(apidiff.VariablesMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
			continue
		}
		exerSDef, err := exerSTable.LookupVariable(symbolItem.name)
		if err != nil {
			apidiff.VariablesMissing = append(apidiff.VariablesMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
			continue
		}

		// Compare both symbols
		if !reflect.DeepEqual(refSDef.Def, exerSDef.Def) {
			apidiff.Variables = append(apidiff.Variables, SymbolInfo{
				Package: symbolItem.pkg,
				Name:    symbolItem.name,
				Pos:     positions,
			})
		}
	}

	for symbolItem, positions := range methodNames {
		glog.V(1).Infof("Checking %q method", symbolItem.str())

		refSTable, err := refGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			return nil, err
		}
		refSDef, err := refSTable.LookupDataType(symbolItem.name)
		if err != nil {
			return nil, fmt.Errorf("Method %q of %q package not found", symbolItem.name, symbolItem.pkg)
		}

		refMethodDef, err := refAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(refSTable, refSDef, &ast.Ident{Name: symbolItem.field}),
		)
		if err != nil {
			return nil, fmt.Errorf("Method %q of %q Type in %q package not found", symbolItem.field, symbolItem.name, symbolItem.pkg)
		}

		exerSTable, err := exercisedGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			apidiff.MethodsMissing = append(apidiff.MethodsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
			continue
		}
		exerSDef, err := exerSTable.LookupDataType(symbolItem.name)
		if err != nil {
			apidiff.MethodsMissing = append(apidiff.MethodsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
			continue
		}

		exerMethodDef, err := exerAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(exerSTable, exerSDef, &ast.Ident{Name: symbolItem.field}),
		)
		if err != nil {
			apidiff.MethodsMissing = append(apidiff.MethodsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
			continue
		}

		// Compare both symbols
		if !reflect.DeepEqual(refMethodDef.DataType, exerMethodDef.DataType) {
			apidiff.Methods = append(apidiff.Methods, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
		}
	}

	for symbolItem, positions := range fieldNames {
		glog.V(1).Infof("Checking %q field", symbolItem.str())

		refSTable, err := refGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			return nil, err
		}
		refSDef, err := refSTable.LookupDataType(symbolItem.name)
		if err != nil {
			return nil, fmt.Errorf("Field %q of %q package not found", symbolItem.name, symbolItem.pkg)
		}

		refMethodDef, err := refAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(refSTable, refSDef, &ast.Ident{Name: symbolItem.field}),
		)
		if err != nil {
			return nil, fmt.Errorf("Field %q of %q Type in %q package not found", symbolItem.field, symbolItem.name, symbolItem.pkg)
		}

		exerSTable, err := exercisedGlobalST.Lookup(symbolItem.pkg)
		if err != nil {
			apidiff.StructFieldsMissing = append(apidiff.StructFieldsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
			continue
		}
		exerSDef, err := exerSTable.LookupDataType(symbolItem.name)
		if err != nil {
			apidiff.StructFieldsMissing = append(apidiff.StructFieldsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
			continue
		}

		exerMethodDef, err := exerAccessor.RetrieveDataTypeField(
			accessors.NewFieldAccessor(exerSTable, exerSDef, &ast.Ident{Name: symbolItem.field}),
		)
		if err != nil {
			apidiff.StructFieldsMissing = append(apidiff.StructFieldsMissing, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
			continue
		}

		// Compare both symbols
		if !reflect.DeepEqual(refMethodDef.DataType, exerMethodDef.DataType) {
			apidiff.StructFields = append(apidiff.StructFields, SymbolInfo{
				Package: symbolItem.pkg,
				Parent:  symbolItem.name,
				Name:    symbolItem.field,
				Pos:     positions,
			})
		}
	}

	return &apidiff, nil
}

const CLR_B = "\x1b[34;1m"
const CLR_N = "\x1b[0m"
const CLR_R = "\x1b[31;1m"

func main() {

	f := &flags{
		allocated:           flag.String("allocated", "", "Allocated symbol table"),
		allocatedGlidefile:  flag.String("allocated-glidefile", "", "Glide.lock with dependencies of allocated symbol table"),
		allocatedGodepsfile: flag.String("allocated-godepsfile", "", "Godeps.json with dependencies of allocated symbol table"),
		packagePrefix:       flag.String("package-prefix", "", "Package entry point"),
		packageCommit:       flag.String("package-commit", "", "Package commit entry point"),
		goVersion:           flag.String("go-version", "", "Go stdlib version"),
		symbolTablePath:     flag.String("symbol-table-dir", "", "Directory with preprocessed symbol tables"),
		cgoSymbolsPath:      flag.String("cgo-symbols-path", "", "Symbol table with CGO symbols (per entire project space)"),
		glidefile:           flag.String("glidefile", "", "Glide.lock with dependencies"),
		godepsfile:          flag.String("godepsfile", "", "Godeps.json with dependencies"),
	}

	if err := f.parse(); err != nil {
		glog.Fatal(err)
	}

	// Otherwise it can eat all the CPU power
	runtime.GOMAXPROCS(1)

	refGlobalST, refSnapshot, err := initRefGlobaST(f)
	if err != nil {
		panic(err)
	}

	exercisedGlobalST, exercisedSnapshot, err := initExercisedGlobaST(f)
	if err != nil {
		panic(err)
	}

	// TODO(jchaloup): construct the snapshot for the refGlobalST from the allocated table

	// 1. Load the allocated symbols Table
	var tables map[string]allocglobal.PackageTable
	raw, err := ioutil.ReadFile(*(f.allocated))
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(raw, &tables); err != nil {
		panic(err)
	}

	// Global symbols Accessing
	// - the symbols from the allocated list are accessed through one global symbols table
	// - the exercised API (and its dependencies) are accessed through another global symbols table
	//

	// 2. Find each symbol in a global symbol table

	// 3. Compare if the symbol definition is the same as in the allocated
	apidiff, err := collectApiDiffs(tables, *f.packagePrefix, refGlobalST, exercisedGlobalST)
	if err != nil {
		panic(err)
	}

	pkg := *f.packagePrefix
	refCommit, _ := refSnapshot.Commit(pkg)
	exercisedCommit, _ := exercisedSnapshot.Commit(pkg)

	fmt.Printf("Comparing %v:%v with %v:%v\n", pkg, refCommit, pkg, exercisedCommit)

	// 4. Report differences (if there are any)
	for _, item := range apidiff.Datatypes {
		// refSTable, _ := refGlobalST.Lookup(pkg)
		// exerSTable, _ := exercisedGlobalST.Lookup(pkg)
		// refSDef, _ := refSTable.LookupDataType(item.Name)
		// exerSDef, _ := exerSTable.LookupDataType(item.Name)
		//
		// // structs?
		// if refSDef.Def.GetType() == gotypes.StructType && exerSDef.Def.GetType() == gotypes.StructType {
		// 	refFields := make(map[string]gotypes.DataType)
		// 	exerFields := make(map[string]gotypes.DataType)
		// 	for _, item := range refSDef.Def.(*gotypes.Struct).Fields {
		// 		if item.Name == "" {
		// 			panic("JJJJ")
		// 		}
		// 		if !ast.IsExported(item.Name) {
		// 			continue
		// 		}
		// 		refFields[item.Name] = item.Def
		// 	}
		//
		// 	for _, item := range exerSDef.Def.(*gotypes.Struct).Fields {
		// 		if item.Name == "" {
		// 			panic("JJJJ")
		// 		}
		// 		if !ast.IsExported(item.Name) {
		// 			continue
		// 		}
		// 		exerFields[item.Name] = item.Def
		// 	}
		//
		// 	// any missing fields?
		// 	for field := range refFields {
		// 		if _, ok := exerFields[field]; ok {
		// 			if !reflect.DeepEqual(refFields[field], exerFields[field]) {
		// 				fmt.Printf("?field %v.%v definition changed\n", item.Name, field)
		// 			}
		// 		}
		// 	}
		// 	continue
		// }

		fmt.Printf("%v?type %q changed%v\n\tused at %v\n", CLR_B, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.DatatypesMissing {
		fmt.Printf("%v-type %q missing%v\n\tused at %v\n", CLR_R, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.VariablesMissing {
		fmt.Printf("%v-function %q missing%v\n\tused at %v\n", CLR_R, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.Variables {
		fmt.Printf("%v?function %q changed%v\n\tused at %v\n", CLR_B, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.FunctionsMissing {
		fmt.Printf("%v-function %q missing%v\n\tused at %v\n", CLR_R, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.Functions {
		fmt.Printf("%v?function %q changed%v\n\tused at %v\n", CLR_B, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.StructFieldsMissing {
		fmt.Printf("%v-field %q missing%v\n\tused at %v\n", CLR_R, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.StructFields {
		fmt.Printf("%v?field %q changed%v\n\tused at %v\n", CLR_B, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.MethodsMissing {
		fmt.Printf("%v-method %q missing%v\n\tused at %v\n", CLR_R, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.Methods {
		fmt.Printf("%v?method %q changed%v\n\tused at %v\n", CLR_B, item.str(), CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

}
