package file

import (
	"fmt"
	"go/ast"
	"path"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"github.com/golang/glog"
)

// Payload stores symbols for parsing/processing
type Payload struct {
	DataTypes         []*ast.TypeSpec
	Variables         []*ast.ValueSpec
	Functions         []*ast.FuncDecl
	FunctionDeclsOnly bool
	Imports           []*ast.ImportSpec
}

type FileParser struct {
	*types.Config
}

func NewParser(config *types.Config) *FileParser {
	return &FileParser{
		Config: config,
	}
}

func MakePackagequalifier(spec *ast.ImportSpec) *gotypes.Packagequalifier {
	q := &gotypes.Packagequalifier{
		Path: strings.Replace(spec.Path.Value, "\"", "", -1),
	}

	if spec.Name == nil {
		// TODO(jchaloup): get the q.Name from the spec.Path's package symbol table
		q.Name = path.Base(q.Path)
	} else {
		q.Name = spec.Name.Name
	}

	return q
}

func MakePayload(f *ast.File) *Payload {
	payload := &Payload{}

	for _, spec := range f.Imports {
		payload.Imports = append(payload.Imports, spec)
	}

	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch d := spec.(type) {
				case *ast.TypeSpec:
					payload.DataTypes = append(payload.DataTypes, d)
				case *ast.ValueSpec:
					payload.Variables = append(payload.Variables, d)
				}
			}
		case *ast.FuncDecl:
			payload.Functions = append(payload.Functions, decl)
		}
	}
	return payload
}

func (fp *FileParser) parseImportSpec(spec *ast.ImportSpec) error {
	q := MakePackagequalifier(spec)

	// TODO(jchaloup): store non-qualified imports as well
	if q.Name == "." {
		return nil
	}

	err := fp.SymbolTable.AddImport(&gotypes.SymbolDef{Name: q.Name, Def: q})
	glog.Infof("Packagequalifier added: %#v\n", &gotypes.SymbolDef{Name: q.Name, Def: q})
	return err
}

func (fp *FileParser) parseImportedPackages(specs []*ast.ImportSpec) error {
	for _, spec := range specs {
		err := fp.parseImportSpec(spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fp *FileParser) parseTypeSpecs(specs []*ast.TypeSpec) ([]*ast.TypeSpec, error) {
	var postponed []*ast.TypeSpec
	for _, spec := range specs {
		// Store a symbol just with a name and origin.
		// Setting the symbol's definition to nil means the symbol is being parsed (somewhere in the chain)
		if err := fp.SymbolTable.AddDataType(&gotypes.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Def:     nil,
		}); err != nil {
			return nil, err
		}

		// TODO(jchaloup): capture the current state of the allocated symbol table
		// JIC the parsing ends with end error. Which can result into re-parsing later on.
		// Which can result in re-allocation. It should be enough two-level allocated symbol table.
		typeDef, err := fp.TypeParser.Parse(spec.Type)
		if err != nil {
			glog.Warningf("File parse TypeParser error: %v\n", err)
			postponed = append(postponed, spec)
			continue
		}

		if err := fp.SymbolTable.AddDataType(&gotypes.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Def:     typeDef,
		}); err != nil {
			return nil, err
		}
	}
	return postponed, nil
}

func (fp *FileParser) parseValueSpecs(specs []*ast.ValueSpec) ([]*ast.ValueSpec, error) {
	var postponed []*ast.ValueSpec
	for _, spec := range specs {

		defs, err := fp.StmtParser.ParseValueSpec(spec)
		if err != nil {
			glog.Warningf("File parse ValueSpec %q error: %v\n", spec.Names[0].Name, err)
			postponed = append(postponed, spec)
			continue
		}
		for _, def := range defs {
			// TPDP(jchaloup): we should store all variables or non.
			// Given the error is set only if the variable already exists, it should not matter so much.
			if err := fp.SymbolTable.AddVariable(def); err != nil {
				return nil, nil
			}
		}
	}

	return postponed, nil
}

func (fp *FileParser) parseFuncDeclaration(spec *ast.FuncDecl) (bool, error) {
	glog.Infof("Parsing function %q declaration", spec.Name.Name)
	// process function definitions as the last
	funcDef, err := fp.StmtParser.ParseFuncDecl(spec)
	if err != nil {
		// if the function declaration is not fully processed (e.g. missing data type)
		// skip it and postponed its processing
		glog.Infof("Postponing function %q declaration parseFuncDecls processing due to: %v", spec.Name.Name, err)
		return true, nil
	}

	if method, ok := funcDef.(*gotypes.Method); ok {
		var receiverDataType string
		switch rExpr := method.Receiver.(type) {
		case *gotypes.Identifier:
			receiverDataType = rExpr.Def
		case *gotypes.Pointer:
			if ident, ok := rExpr.Def.(*gotypes.Identifier); ok {
				receiverDataType = ident.Def
			} else {
				return false, fmt.Errorf("Expecting a pointer to an identifier as a receiver for %q method, got %#v instead", spec.Name.Name, rExpr)
			}
		default:
			return false, fmt.Errorf("Expecting an identifier or a pointer to an identifier as a receiver for %q method, got %#v instead", spec.Name.Name, method.Receiver)
		}

		def, err := fp.SymbolTable.LookupMethod(receiverDataType, spec.Name.Name)
		glog.Infof("Looking I up for %q.%q\terr: %v\tDef: %#v\tRecv: %#v\n", receiverDataType, spec.Name.Name, err, def, spec.Recv.List)
		// method declaration already exist
		if err == nil {
			return false, nil
		}
	} else {
		def, err := fp.SymbolTable.LookupFunction(spec.Name.Name)
		glog.Infof("Looking II up for %q\terr: %v\tDef: %#v\tRecv: %#v\n", spec.Name.Name, err, def, spec.Recv)
		// function declaration already exist
		if err == nil {
			return false, nil
		}
	}

	if err := fp.SymbolTable.AddFunction(&gotypes.SymbolDef{
		Name:    spec.Name.Name,
		Package: fp.PackageName,
		Def:     funcDef,
	}); err != nil {
		glog.Info("Error during parsing of function %q declaration: %v", spec.Name.Name, err)
		return false, err
	}
	return false, nil
}

func (fp *FileParser) parseFuncDecls(specs []*ast.FuncDecl) error {
	for _, spec := range specs {
		if _, err := fp.parseFuncDeclaration(spec); err != nil {
			return err
		}
	}

	return nil
}

func (fp *FileParser) parseFuncs(specs []*ast.FuncDecl) ([]*ast.FuncDecl, error) {
	var postponed []*ast.FuncDecl
	for _, spec := range specs {
		postpone, err := fp.parseFuncDeclaration(spec)
		if err != nil {
			return nil, err
		}
		if postpone {
			postponed = append(postponed, spec)
		}

		// if an error is returned, put the function's AST into a context list
		// and continue with other definition
		// TODO(jchaloup): reset the alloc table back to the state before the body got processed
		// in case the processing ended with an error
		if err := fp.StmtParser.ParseFuncBody(spec); err != nil {
			glog.Warningf("File parse %q Funcs error: %v\n", spec.Name.Name, err)
			postponed = append(postponed, spec)
			continue
		}
	}

	return postponed, nil
}

func (fp *FileParser) Parse(p *Payload) error {

	// Given:
	// type My1 My2
	// type My2 My3
	// ...
	// type Myn Myn-1
	//
	// We can not re-process a symbol processing (due to unknown symbol) until we know
	// we process the unknown symbol before it is needed.
	// Given a type name is stored into the symbol table before its definition is parsed
	// all data type names are known in the second round of the postponed type definition processing

	// Imported packages first
	{
		err := fp.parseImportedPackages(p.Imports)
		if err != nil {
			return err
		}
	}

	printFuncNames := func(funcs []*ast.FuncDecl) []string {
		var names []string
		for _, f := range funcs {
			names = append(names, f.Name.Name)
		}
		return names
	}

	printVarNames := func(vars []*ast.ValueSpec) []string {
		var names []string
		for _, f := range vars {
			for _, n := range f.Names {
				names = append(names, n.Name)
			}
		}
		return names
	}

	// Data definitions as second
	{
		glog.Infof("\n\nBefore parseTypeSpecs: %v", len(p.DataTypes))
		postponed, err := fp.parseTypeSpecs(p.DataTypes)
		if err != nil {
			return err
		}
		p.DataTypes = postponed
		glog.Infof("\n\nAfter parseTypeSpecs: %v", len(p.DataTypes))
	}

	// Function/Method declarations
	{
		glog.Infof("\n\nBefore parseFuncDecls: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
		if err := fp.parseFuncDecls(p.Functions); err != nil {
			return err
		}
		glog.Infof("\n\nAfter parseFuncDecls: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
	}

	// // Vars/constants
	{
		glog.Infof("\n\nBefore parseValueSpecs: %v\tNames: %v\n", len(p.Variables), strings.Join(printVarNames(p.Variables), ","))
		postponed, err := fp.parseValueSpecs(p.Variables)
		if err != nil {
			return err
		}
		p.Variables = postponed
		glog.Infof("\n\nAfter parseValueSpecs: %v\tNames: %v\n", len(p.Variables), strings.Join(printVarNames(p.Variables), ","))
	}

	// // Funcs
	if !p.FunctionDeclsOnly {
		glog.Infof("\n\nBefore parseFuncs: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
		postponed, err := fp.parseFuncs(p.Functions)
		if err != nil {
			return err
		}
		p.Functions = postponed
		glog.Infof("\n\nAfter parseFuncs: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
	}

	fmt.Printf("AllocST for %q\n", fp.PackageName)
	fp.AllocatedSymbolsTable.Print()
	//fp.SymbolTable.Json()

	return nil
}
