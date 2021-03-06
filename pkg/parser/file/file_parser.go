package file

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/parser/types"
	"github.com/gofed/symbols-extractor/pkg/symbols"
	"github.com/gofed/symbols-extractor/pkg/symbols/tables"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
	"k8s.io/klog/v2"
)

// Payload stores symbols for parsing/processing
type Payload struct {
	DataTypes         []*ast.TypeSpec
	Variables         []*ast.ValueSpec
	Constants         []types.ConstSpec
	Functions         []*ast.FuncDecl
	FunctionDeclsOnly bool
	Reprocessing      bool
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

func (fp *FileParser) MakePackagequalifier(spec *ast.ImportSpec) *gotypes.Packagequalifier {
	q := &gotypes.Packagequalifier{
		Path: strings.Replace(spec.Path.Value, "\"", "", -1),
	}

	if spec.Name == nil {
		// TODO(jchaloup): get the q.Name from the spec.Path's package symbol table
		st, err := fp.Config.GlobalSymbolTable.Lookup(q.Path)
		if err != nil {
			panic(err)
		}
		switch d := st.(type) {
		case *tables.Table:
			q.Name = d.PackageQID
		case *tables.CGOTable:
			q.Name = d.PackageQID
		default:
			panic(st)
		}
	} else {
		q.Name = spec.Name.Name
	}

	return q
}

func MakePayload(f *ast.File) (*Payload, error) {
	payload := &Payload{}

	for _, spec := range f.Imports {
		payload.Imports = append(payload.Imports, spec)
	}

	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			// var lastValueSpecType ast.Expr
			// var lastValueSpecValue ast.Expr
			switch decl.Tok {
			case token.TYPE:
				for _, spec := range decl.Specs {
					payload.DataTypes = append(payload.DataTypes, spec.(*ast.TypeSpec))
				}
			case token.VAR:
				for _, spec := range decl.Specs {
					payload.Variables = append(payload.Variables, spec.(*ast.ValueSpec))
				}
			case token.CONST:
				var lastConstSpecValue []ast.Expr
				var lastConstSpecType ast.Expr
				for i, spec := range decl.Specs {
					constSpec := spec.(*ast.ValueSpec)
					// Is the value empty? Copy the previous expression
					if constSpec.Values == nil {
						if lastConstSpecValue == nil {
							return nil, fmt.Errorf("Unable to re-costruct a value for const %#v", constSpec)
						}
						constSpec.Values = lastConstSpecValue
					} else {
						lastConstSpecType = nil
					}
					if constSpec.Type == nil && lastConstSpecType != nil {
						constSpec.Type = lastConstSpecType
					}
					for j, constName := range constSpec.Names {
						cSpec := &ast.ValueSpec{
							Doc:     constSpec.Doc,
							Names:   []*ast.Ident{constName},
							Type:    constSpec.Type,
							Values:  nil,
							Comment: constSpec.Comment,
						}

						if constSpec.Values != nil && j < len(constSpec.Values) {
							cSpec.Values = []ast.Expr{constSpec.Values[j]}
						}

						payload.Constants = append(payload.Constants, types.ConstSpec{
							IotaIdx: uint(i),
							Spec:    cSpec,
						})
					}
					lastConstSpecValue = constSpec.Values
					// once there is a type it applies to all following untyped constants until there is a new one
					if constSpec.Type != nil {
						lastConstSpecType = constSpec.Type
					}
				}
			case token.IMPORT:
				// already processed
				continue
			default:
				panic(fmt.Sprintf("Unexpected ast.GenDecl %#v", decl))
			}

			// for _, spec := range decl.Specs {
			// 	switch d := spec.(type) {
			// 	case *ast.TypeSpec:
			// 		payload.DataTypes = append(payload.DataTypes, d)
			// 	case *ast.ValueSpec:
			// 		// Either error or iota as a value
			// 		if d.Type == nil && d.Values == nil {
			// 			if lastValueSpecValue == nil {
			// 				return nil, fmt.Errorf("Missing Type and Value for ValueSpec %#v", d)
			// 			}
			// 			d.Values = []ast.Expr{lastValueSpecValue}
			// 			if lastValueSpecType != nil {
			// 				d.Type = lastValueSpecType
			// 			}
			// 		}
			//
			// 		if len(d.Values) == 1 {
			// 			lastValueSpecValue = d.Values[0]
			// 			lastValueSpecType = d.Type
			// 		} else {
			// 			lastValueSpecValue = nil
			// 			lastValueSpecType = nil
			// 		}
			// 		payload.Variables = append(payload.Variables, d)
			// 	}
			// }
		case *ast.FuncDecl:
			payload.Functions = append(payload.Functions, decl)
		}
	}

	return payload, nil
}

func (fp *FileParser) parseImportSpec(spec *ast.ImportSpec) error {
	q := fp.MakePackagequalifier(spec)

	// TODO(jchaloup): store non-qualified imports as well
	if q.Name == "." {
		return nil
	}

	err := fp.SymbolTable.AddImport(&symbols.SymbolDef{Name: q.Name, Def: q})
	klog.V(2).Infof("Packagequalifier added: %#v\n", &symbols.SymbolDef{Name: q.Name, Def: q})
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
	fp.Config.ContractTable.UnsetPrefix()
	defer fp.Config.ContractTable.UnsetPrefix()

	for _, spec := range specs {
		fp.Config.ContractTable.SetPrefix(fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Pos()))
		// Store a symbol just with a name and origin.
		// Setting the symbol's definition to nil means the symbol is being parsed (somewhere in the chain)
		if err := fp.SymbolTable.AddDataType(&symbols.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Pos:     fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Pos()),
			Def:     nil,
		}); err != nil {
			fp.Config.ContractTable.DropPrefixContracts(fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Pos()))
			return nil, err
		}

		// TODO(jchaloup): capture the current state of the allocated symbol table
		// JIC the parsing ends with end error. Which can result into re-parsing later on.
		// Which can result in re-allocation. It should be enough two-level allocated symbol table.
		typeDef, err := fp.TypeParser.Parse(spec.Type)
		if err != nil {
			klog.V(2).Infof("File parse TypeParser error: %v\n", err)
			postponed = append(postponed, spec)
			continue
		}

		if err := fp.SymbolTable.AddDataType(&symbols.SymbolDef{
			Name:    spec.Name.Name,
			Package: fp.PackageName,
			Pos:     fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Pos()),
			Def:     typeDef,
		}); err != nil {
			return nil, err
		}
	}
	return postponed, nil
}

func (fp *FileParser) parseConstValueSpecs(specs []types.ConstSpec, reprocessing bool) ([]types.ConstSpec, error) {
	var postponed []types.ConstSpec
	fp.Config.ContractTable.UnsetPrefix()
	defer fp.Config.ContractTable.UnsetPrefix()

	for {
		for _, spec := range specs {
			fp.Config.ContractTable.SetPrefix(fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Spec.Names[0].Pos()))
			defs, err := fp.StmtParser.ParseConstValueSpec(spec)
			if err != nil {
				klog.V(2).Infof("File parse ValueSpec %#v error: %v\n", spec, err)
				postponed = append(postponed, spec)
				fp.Config.ContractTable.DropPrefixContracts(fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Spec.Names[0].Pos()))
				continue
			}
			for _, def := range defs {
				// TPDP(jchaloup): we should store all variables or non.
				// Given the error is set only if the variable already exists, it should not matter so much.
				if err := fp.SymbolTable.AddVariable(def); err != nil {
					return nil, err
				}
			}
		}
		if !reprocessing {
			break
		}

		if len(postponed) < len(specs) {
			specs = postponed
			postponed = nil
			continue
		}
		break
	}

	return postponed, nil
}

func (fp *FileParser) parseValueSpecs(specs []*ast.ValueSpec, reprocessing bool) ([]*ast.ValueSpec, error) {
	var postponed []*ast.ValueSpec
	fp.Config.ContractTable.UnsetPrefix()
	defer fp.Config.ContractTable.UnsetPrefix()

	for {
		for _, spec := range specs {
			fp.Config.ContractTable.SetPrefix(fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Names[0].Pos()))
			defs, err := fp.StmtParser.ParseValueSpec(spec)
			if err != nil {
				klog.V(2).Infof("File parse ValueSpec %q error: %v\n", spec.Names[0].Name, err)
				postponed = append(postponed, spec)
				fp.Config.ContractTable.DropPrefixContracts(fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Names[0].Pos()))
				continue
			}

			for _, def := range defs {
				// TPDP(jchaloup): we should store all variables or non.
				// Given the error is set only if the variable already exists, it should not matter so much.
				if err := fp.SymbolTable.AddVariable(def); err != nil {
					return nil, err
				}
			}
		}
		if !reprocessing {
			break
		}

		if len(postponed) < len(specs) {
			specs = postponed
			postponed = nil
			continue
		}
		break
	}

	return postponed, nil
}

func (fp *FileParser) parseFuncDeclaration(spec *ast.FuncDecl) (bool, error) {
	klog.V(2).Infof("Parsing function %q declaration", spec.Name.Name)
	// process function definitions as the last
	funcDef, err := fp.StmtParser.ParseFuncDecl(spec)
	if err != nil {
		// if the function declaration is not fully processed (e.g. missing data type)
		// skip it and postponed its processing
		klog.V(2).Infof("Postponing function %q declaration parseFuncDecls processing due to: %v", spec.Name.Name, err)
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
		klog.V(2).Infof("Looking I up for %q.%q\terr: %v\tDef: %#v\tRecv: %#v\n", receiverDataType, spec.Name.Name, err, def, spec.Recv.List)
		// method declaration already exist
		if err == nil {
			return false, nil
		}
	} else {
		def, err := fp.SymbolTable.LookupFunction(spec.Name.Name)
		klog.V(2).Infof("Looking II up for %q\terr: %v\tDef: %#v\tRecv: %#v\n", spec.Name.Name, err, def, spec.Recv)
		// function declaration already exist
		if err == nil {
			return false, nil
		}
	}

	if err := fp.SymbolTable.AddFunction(&symbols.SymbolDef{
		Name:    spec.Name.Name,
		Package: fp.PackageName,
		Pos:     fmt.Sprintf("%v:%v", fp.Config.FileName, spec.Pos()),
		Def:     funcDef,
	}); err != nil {
		klog.V(2).Infof("Error during parsing of function %q declaration: %v", spec.Name.Name, err)
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
	fp.Config.ContractTable.UnsetPrefix()
	defer fp.Config.ContractTable.UnsetPrefix()

	for _, spec := range specs {
		postpone, err := fp.parseFuncDeclaration(spec)
		if err != nil {
			// If this errors -> no more parsing of the file
			// If anything gets reprocessed again, it can not error here
			return nil, err
		}
		if postpone {
			postponed = append(postponed, spec)
		}

		// if an error is returned, put the function's AST into a context list
		// and continue with other definition
		// TODO(jchaloup): reset the alloc table back to the state before the body got processed
		// in case the processing ended with an error
		fp.Config.ContractTable.SetPrefix(spec.Name.Name)
		if err := fp.StmtParser.ParseFuncBody(spec); err != nil {
			fp.Config.ContractTable.DropPrefixContracts(spec.Name.Name)
			klog.V(2).Infof("File %q/%q parse %q Funcs error: %v\n", fp.Config.PackageName, fp.Config.FileName, spec.Name.Name, err)
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

	printDataTypeNames := func(types []*ast.TypeSpec) []string {
		var names []string
		for _, t := range types {
			names = append(names, t.Name.Name)
		}
		return names
	}

	// Data definitions as second
	{
		klog.V(2).Infof("\n\nBefore parseTypeSpecs: %v\n\tNames: %v\n", len(p.DataTypes), strings.Join(printDataTypeNames(p.DataTypes), ","))
		postponed, err := fp.parseTypeSpecs(p.DataTypes)
		if err != nil {
			return err
		}
		p.DataTypes = postponed
		klog.V(2).Infof("\n\nAfter parseTypeSpecs: %v", len(p.DataTypes))
	}

	// Function/Method declarations
	{
		klog.V(2).Infof("\n\nBefore parseFuncDecls: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
		if err := fp.parseFuncDecls(p.Functions); err != nil {
			return err
		}
		klog.V(2).Infof("\n\nAfter parseFuncDecls: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
	}

	// Constants
	{
		var cNames []string
		for _, spec := range p.Constants {
			cNames = append(cNames, printVarNames([]*ast.ValueSpec{spec.Spec})...)
		}
		klog.V(2).Infof("\n\nBefore const parseValueSpecs: %v\tNames: %v\n", len(p.Constants), strings.Join(cNames, ","))
		fp.Config.IsConst = true
		postponed, err := fp.parseConstValueSpecs(p.Constants, p.Reprocessing)
		if err != nil {
			return err
		}
		p.Constants = postponed
		cNames = nil
		for _, spec := range p.Constants {
			cNames = append(cNames, printVarNames([]*ast.ValueSpec{spec.Spec})...)
		}
		klog.V(2).Infof("\n\nAfter const parseValueSpecs: %v\tNames: %v\n", len(p.Constants), strings.Join(cNames, ","))
	}

	// Vars
	{
		klog.V(2).Infof("\n\nBefore parseValueSpecs: %v\tNames: %v\n", len(p.Variables), strings.Join(printVarNames(p.Variables), ","))
		fp.Config.IsConst = false
		postponed, err := fp.parseValueSpecs(p.Variables, p.Reprocessing)
		if err != nil {
			return err
		}
		p.Variables = postponed
		klog.V(2).Infof("\n\nAfter parseValueSpecs: %v\tNames: %v\n", len(p.Variables), strings.Join(printVarNames(p.Variables), ","))
	}

	// // Funcs
	if !p.FunctionDeclsOnly {
		klog.V(2).Infof("\n\nBefore parseFuncs: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
		fp.Config.IsConst = false
		postponed, err := fp.parseFuncs(p.Functions)
		if err != nil {
			return err
		}
		p.Functions = postponed
		klog.V(2).Infof("\n\nAfter parseFuncs: %v\tNames: %v\n", len(p.Functions), strings.Join(printFuncNames(p.Functions), ","))
	}

	// fmt.Printf("AllocST for %q\n", fp.PackageName)
	// fp.AllocatedSymbolsTable.Print()
	// fp.SymbolTable.Json()

	return nil
}
