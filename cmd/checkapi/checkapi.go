package main

import (
	"flag"
	"fmt"
	"go/ast"
	"path"
	"reflect"
	"runtime"
	"strings"

	allocglobal "github.com/gofed/symbols-extractor/pkg/parser/alloctable/global"
	"github.com/gofed/symbols-extractor/pkg/snapshots"
	"github.com/gofed/symbols-extractor/pkg/snapshots/glide"
	"github.com/gofed/symbols-extractor/pkg/symbols/accessors"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables/global"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"

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
	allocated          *string
	allocatedGlidefile *string
	packagePrefix      *string
	packageCommit      *string
	symbolTablePath    *string
	cgoSymbolsPath     *string
	goVersion          *string
	glidefile          *string
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

	if *(f.glidefile) == "" {
		return fmt.Errorf("-glidefile is not set")
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
	}
	return global.New(*(f.symbolTablePath), *(f.goVersion), nil), nil, nil
}

func initExercisedGlobaST(f *flags) (*global.Table, snapshots.Snapshot, error) {
	snapshot, err := glide.GlideFromFile(*f.glidefile)
	if err != nil {
		return nil, nil, err
	}
	snapshot.MainPackageCommit(*f.packagePrefix, *f.packageCommit)

	return global.New(*(f.symbolTablePath), *(f.goVersion), snapshot), snapshot, nil
}

type Triplet struct {
	Parent, Name string
	Pos          []string
}

type ApiDiff struct {
	DatatypesMissing    []string
	Datatypes           []string
	FunctionsMissing    []string
	Functions           []string
	VariablesMissing    []string
	Variables           []string
	MethodsMissing      []Triplet
	Methods             []Triplet
	StructFieldsMissing []Triplet
	StructFields        []Triplet
}

func collectApiDiffs(table allocglobal.PackageTable, packagePrefix string, refGlobalST, exercisedGlobalST *global.Table) (*ApiDiff, error) {
	apidiff := ApiDiff{}

	refSTable, err := refGlobalST.Lookup(packagePrefix)
	if err != nil {
		return nil, err
	}

	exerSTable, err := exercisedGlobalST.Lookup(packagePrefix)
	if err != nil {
		return nil, err
	}

	refAccessor := accessors.NewAccessor(refGlobalST)
	exerAccessor := accessors.NewAccessor(exercisedGlobalST)

	for file, fileItem := range table {
		glog.V(1).Infof("Processing %q of %q\n", file, fileItem.Package)
		for pkg, symbolsSet := range fileItem.Symbols {
			// skip anonymous symbols
			if pkg == "" {
				continue
			}

			if pkg != packagePrefix {
				continue
			}

			glog.V(1).Infof("Processing %q\n", pkg)

			dtNames := make(map[string]struct{}, 0)
			fncNames := make(map[string]struct{}, 0)
			varNames := make(map[string]struct{}, 0)
			fieldNames := make(map[struct{ parent, field string }][]string, 0)
			methodNames := make(map[struct{ parent, name string }][]string, 0)

			// Just collect names without positions
			// Datatypes
			for _, item := range symbolsSet.Datatypes {
				if _, ok := dtNames[item.Name]; ok {
					continue
				}
				if !ast.IsExported(item.Name) {
					continue
				}
				dtNames[item.Name] = struct{}{}
			}
			// Functions
			for _, item := range symbolsSet.Functions {
				if _, ok := fncNames[item.Name]; ok {
					continue
				}
				if !ast.IsExported(item.Name) {
					continue
				}
				fncNames[item.Name] = struct{}{}
			}
			// Variables
			for _, item := range symbolsSet.Variables {
				if _, ok := varNames[item.Name]; ok {
					continue
				}
				if !ast.IsExported(item.Name) {
					continue
				}
				varNames[item.Name] = struct{}{}
			}
			// Struct fields
			for _, item := range symbolsSet.Structfields {
				key := struct{ parent, field string }{item.Parent, item.Field}
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
				key := struct{ parent, name string }{item.Parent, item.Name}
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

			for name := range dtNames {
				glog.V(1).Infof("Checking %q type", name)
				refSDef, err := refSTable.LookupDataType(name)
				if err != nil {
					return nil, fmt.Errorf("Type %q of %q package not found", name, packagePrefix)
				}

				exerSDef, err := exerSTable.LookupDataType(name)
				if err != nil {
					apidiff.DatatypesMissing = append(apidiff.DatatypesMissing, name)
					continue
				}

				// Compare both symbols
				if !reflect.DeepEqual(refSDef.Def, exerSDef.Def) {
					apidiff.Datatypes = append(apidiff.Datatypes, name)
				}
			}

			for name := range fncNames {
				glog.V(1).Infof("Checking %q function", name)
				refSDef, err := refSTable.LookupFunction(name)
				if err != nil {
					return nil, fmt.Errorf("Function %q of %q package not found", name, packagePrefix)
				}

				exerSDef, err := exerSTable.LookupFunction(name)
				if err != nil {
					apidiff.FunctionsMissing = append(apidiff.FunctionsMissing, name)
					continue
				}

				// Compare both symbols
				if !reflect.DeepEqual(refSDef.Def, exerSDef.Def) {
					apidiff.Functions = append(apidiff.Functions, name)
				}
			}

			for name := range varNames {
				glog.V(1).Infof("Checking %q variable", name)
				refSDef, err := refSTable.LookupVariable(name)
				if err != nil {
					return nil, fmt.Errorf("Variable %q of %q package not found", name, packagePrefix)
				}

				exerSDef, err := exerSTable.LookupVariable(name)
				if err != nil {
					apidiff.VariablesMissing = append(apidiff.VariablesMissing, name)
					continue
				}

				// Compare both symbols
				if !reflect.DeepEqual(refSDef.Def, exerSDef.Def) {
					apidiff.Variables = append(apidiff.Variables, name)
				}
			}

			for item := range methodNames {
				glog.V(1).Infof("Checking %q method", fmt.Sprintf("%v.%v", item.parent, item.name))

				refSDef, err := refSTable.LookupDataType(item.parent)
				if err != nil {
					return nil, fmt.Errorf("Type %q of %q package not found", item.parent, packagePrefix)
				}

				refMethodDef, err := refAccessor.RetrieveDataTypeField(
					accessors.NewFieldAccessor(refSTable, refSDef, &ast.Ident{Name: item.name}),
				)
				if err != nil {
					return nil, fmt.Errorf("Method %q of %q Type in %q package not found", item.name, item.parent, packagePrefix)
				}

				exerSDef, err := exerSTable.LookupDataType(item.parent)
				if err != nil {
					apidiff.MethodsMissing = append(apidiff.MethodsMissing, Triplet{
						Parent: item.parent,
						Name:   item.name,
						Pos:    methodNames[item],
					})
					continue
				}

				exerMethodDef, err := exerAccessor.RetrieveDataTypeField(
					accessors.NewFieldAccessor(exerSTable, exerSDef, &ast.Ident{Name: item.name}),
				)
				if err != nil {
					apidiff.MethodsMissing = append(apidiff.MethodsMissing, Triplet{
						Parent: item.parent,
						Name:   item.name,
						Pos:    methodNames[item],
					})
					continue
				}

				// Compare both symbols
				if !reflect.DeepEqual(refMethodDef.DataType, exerMethodDef.DataType) {
					apidiff.Methods = append(apidiff.Methods, Triplet{
						Parent: item.parent,
						Name:   item.name,
						Pos:    methodNames[item],
					})
				}
			}

			for item := range fieldNames {
				glog.V(1).Infof("Checking %q field", fmt.Sprintf("%v.%v", item.parent, item.field))

				refSDef, err := refSTable.LookupDataType(item.parent)
				if err != nil {
					return nil, fmt.Errorf("Type %q of %q package not found", item.parent, packagePrefix)
				}

				refFieldDef, err := refAccessor.RetrieveDataTypeField(
					accessors.NewFieldAccessor(refSTable, refSDef, &ast.Ident{Name: item.field}),
				)
				if err != nil {
					return nil, fmt.Errorf("Field %q of %q Type in %q package not found", item.field, item.parent, packagePrefix)
				}

				exerSDef, err := exerSTable.LookupDataType(item.parent)
				if err != nil {
					apidiff.StructFieldsMissing = append(apidiff.StructFieldsMissing, Triplet{
						Parent: item.parent,
						Name:   item.field,
						Pos:    fieldNames[item],
					})
					continue
				}

				exerFieldDef, err := exerAccessor.RetrieveDataTypeField(
					accessors.NewFieldAccessor(exerSTable, exerSDef, &ast.Ident{Name: item.field}),
				)

				if err != nil {
					apidiff.StructFieldsMissing = append(apidiff.StructFieldsMissing, Triplet{
						Parent: item.parent,
						Name:   item.field,
						Pos:    fieldNames[item],
					})
					continue
				}

				// Compare both symbols
				if !reflect.DeepEqual(refFieldDef.DataType, exerFieldDef.DataType) {
					apidiff.StructFields = append(apidiff.StructFields, Triplet{
						Parent: item.parent,
						Name:   item.field,
						Pos:    fieldNames[item],
					})
				}
			}
		}
	}

	return &apidiff, nil
}

const CLR_B = "\x1b[34;1m"
const CLR_N = "\x1b[0m"
const CLR_R = "\x1b[31;1m"

func main() {

	f := &flags{
		allocated:          flag.String("allocated", "", "Allocated symbol table"),
		allocatedGlidefile: flag.String("allocated-glidefile", "", "Glide.lock with dependencies of allocated symbol table"),
		packagePrefix:      flag.String("package-prefix", "", "Package entry point"),
		packageCommit:      flag.String("package-commit", "", "Package commit entry point"),
		goVersion:          flag.String("go-version", "", "Go stdlib version"),
		symbolTablePath:    flag.String("symbol-table-dir", "", "Directory with preprocessed symbol tables"),
		cgoSymbolsPath:     flag.String("cgo-symbols-path", "", "Symbol table with CGO symbols (per entire project space)"),
		glidefile:          flag.String("glidefile", "", "Glide.lock with dependencies"),
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
	var table allocglobal.PackageTable
	if err := table.Load(*(f.allocated)); err != nil {
		fmt.Printf("table.Load: %v\n", err)
		return
	}

	// Global symbols Accessing
	// - the symbols from the allocated list are accessed through one global symbols table
	// - the exercised API (and its dependencies) are accessed through another global symbols table
	//

	// 2. Find each symbol in a global symbol table

	// 3. Compare if the symbol definition is the same as in the allocated
	apidiff, _ := collectApiDiffs(table, *f.packagePrefix, refGlobalST, exercisedGlobalST)

	fmt.Printf("apidiff: %#v\n\n", apidiff)

	pkg := *f.packagePrefix
	refCommit, _ := refSnapshot.Commit(pkg)
	exercisedCommit, _ := exercisedSnapshot.Commit(pkg)

	fmt.Printf("Comparing %v:%v with %v:%v\n", pkg, refCommit, pkg, exercisedCommit)

	// 4. Report differences (if there are any)
	for _, name := range apidiff.Datatypes {
		refSTable, _ := refGlobalST.Lookup(pkg)
		exerSTable, _ := exercisedGlobalST.Lookup(pkg)
		refSDef, _ := refSTable.LookupDataType(name)
		exerSDef, _ := exerSTable.LookupDataType(name)

		// structs?
		if refSDef.Def.GetType() == gotypes.StructType && exerSDef.Def.GetType() == gotypes.StructType {
			refFields := make(map[string]gotypes.DataType)
			exerFields := make(map[string]gotypes.DataType)
			for _, item := range refSDef.Def.(*gotypes.Struct).Fields {
				if item.Name == "" {
					panic("JJJJ")
				}
				if !ast.IsExported(item.Name) {
					continue
				}
				refFields[item.Name] = item.Def
			}

			for _, item := range exerSDef.Def.(*gotypes.Struct).Fields {
				if item.Name == "" {
					panic("JJJJ")
				}
				if !ast.IsExported(item.Name) {
					continue
				}
				exerFields[item.Name] = item.Def
			}

			// any missing fields?
			for field := range refFields {
				if _, ok := exerFields[field]; ok {
					if !reflect.DeepEqual(refFields[field], exerFields[field]) {
						fmt.Printf("?field %v.%v definition changed\n", name, field)
					}
				}
			}
			continue
		}

		fmt.Printf("Definition of type %q changed\n", name)
	}

	for _, item := range apidiff.StructFieldsMissing {
		fmt.Printf("%v-field %v.%v missing%v\n\tused at %v\n", CLR_R, item.Parent, item.Name, CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.StructFields {
		fmt.Printf("%v?field %v.%v changed%v\n\tused at %v\n", CLR_B, item.Parent, item.Name, CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.MethodsMissing {
		fmt.Printf("%v-method %v.%v missing%v\n\tused at %v\n", CLR_R, item.Parent, item.Name, CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

	for _, item := range apidiff.Methods {
		fmt.Printf("%v?method %v.%v changed%v\n\tused at %v\n", CLR_B, item.Parent, item.Name, CLR_N, strings.Join(item.Pos, "\n\tused at "))
	}

}
