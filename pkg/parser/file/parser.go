package file

import (
	"fmt"
	"go/ast"
	"path"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

// Payload stores symbols for parsing/processing
type Payload struct {
	DataTypes []*ast.TypeSpec
	Variables []*ast.ValueSpec
	Functions []*ast.FuncDecl
	Imports   []*ast.ImportSpec
}

type FileParser struct {
	*types.Config
}

func NewParser(config *types.Config) *FileParser {
	return &FileParser{
		Config: config,
	}
}

func MakePackageQualifier(spec *ast.ImportSpec) *gotypes.PackageQualifier {
	q := &gotypes.PackageQualifier{
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
	q := MakePackageQualifier(spec)

	// TODO(jchaloup): store non-qualified imports as well
	if q.Name == "." {
		return nil
	}

	err := fp.SymbolTable.AddImport(&gotypes.SymbolDef{Name: q.Name, Def: q})
	fmt.Printf("PQ added: %#v\n", &gotypes.SymbolDef{Name: q.Name, Def: q})
	fmt.Printf("PQ: %#v\n", q)
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

func (fp *FileParser) parseFuncs(specs []*ast.FuncDecl) ([]*ast.FuncDecl, error) {
	var postponed []*ast.FuncDecl
	for _, spec := range specs {

		// process function definitions as the last
		//fmt.Printf("%+v\n", d)
		funcDef, err := fp.StmtParser.ParseFuncDecl(spec)
		if err != nil {
			// TODO(jchaloup): this should not error (ever)
			// It can happen an identifier is unknown. Given all imported packages
			// are already processed, the identifier comes from the same package, just
			// from a different file. So the type parser should assume that and choose
			// symbol's origin accordingaly.
			postponed = append(postponed, spec)
			continue
		}

		if err := fp.SymbolTable.AddFunction(&gotypes.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Def:     funcDef,
		}); err != nil {
			return nil, err
		}

		// if an error is returned, put the function's AST into a context list
		// and continue with other definition
		if err := fp.StmtParser.ParseFuncBody(spec); err != nil {
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

	// Data definitions as second
	{
		postponed, err := fp.parseTypeSpecs(p.DataTypes)
		if err != nil {
			return err
		}
		p.DataTypes = postponed
	}

	// // Vars/constants
	{
		postponed, err := fp.parseValueSpecs(p.Variables)
		if err != nil {
			return err
		}
		p.Variables = postponed
	}

	// // Funcs
	{
		postponed, err := fp.parseFuncs(p.Functions)
		if err != nil {
			return err
		}
		p.Functions = postponed
	}

	fmt.Printf("\n\n")
	fp.AllocatedSymbolsTable.Print()

	return nil
}
