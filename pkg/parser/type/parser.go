package typeparser

import (
	"fmt"
	"go/ast"

	"github.com/gofed/symbols-extractor/pkg/parser/alloctable"
	"github.com/gofed/symbols-extractor/pkg/parser/symboltable"
	"github.com/gofed/symbols-extractor/pkg/parser/types"
	gotypes "github.com/gofed/symbols-extractor/pkg/types"
)

var builtinTypes = map[string]struct{}{
	//TODO(pstodulk):
	//  - generate the map from: https://golang.org/src/builtin/builtin.go
	//  - NOTE: there are just documentation types too (like Type1) which
	//          should not be part of this map
	//          - append all types except of Doc. only?
	"uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
	"int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {},
	"float32": {}, "float64": {},
	"complex64": {}, "complex128": {},
	"string": {}, "byte": {}, "rune": {},
	"chan": {}, "bool": {},
	"uintptr": {}, "error": {},
}

func isBuiltin(ident string) bool {
	_, ok := builtinTypes[ident]

	return ok
}

// Parser parsing Go types
type Parser struct {
	// package name
	packageName string
	// per file symbol table
	symbolTable *symboltable.Stack
	// per file allocatable ST
	allocatedSymbolsTable *alloctable.Table
	// name of the currently processed data type
	currentDataTypeName string

	// TODO(jchaloup):
	// - create a project scoped symbol table in a higher level struct (e.g. ProjectParser)
	// - merge all per file symbol tables continuously in the higher level struct (each time a new symbol definition is process)
	//
}

func (p *Parser) parseIdentifier(typedExpr *ast.Ident) (gotypes.DataType, error) {
	// TODO(jchaloup): store symbol's origin as well (in a case a symbol is imported without qid)
	// Check if the identifier is in the any of the global symbol tables (in a case a symbol is imported without qid).
	// If it is, the origin is known. If it is not, the origin is the current package.
	// Check if the identifier is embedded type first. Then the origin is empty.

	// Every data type definition consists of a set of identifiers.
	// Whenever an identifier is used in the definition,
	// it is allocated.

	// Check if the identifier is a built-in type
	if isBuiltin(typedExpr.Name) {
		p.allocatedSymbolsTable.AddSymbol("", typedExpr.Name)
		return &gotypes.Builtin{}, nil
	}

	// Check if the identifier is available in the symbol table
	def, err := p.symbolTable.Lookup(typedExpr.Name)
	if err != nil {
		return nil, fmt.Errorf("Unable to find symbol %v in the symbol table", typedExpr.Name)
	}

	// TODO(jchaloup): consider if we should count the recursive use of a data type into its allocation count
	p.allocatedSymbolsTable.AddSymbol(def.Package, def.Name)

	return &gotypes.Identifier{
		Def: typedExpr.Name,
	}, nil
}

func (p *Parser) parseStar(typedExpr *ast.StarExpr) (*gotypes.Pointer, error) {
	// X.Sel
	def, err := p.Parse(typedExpr.X)
	if err != nil {
		return nil, err
	}
	return &gotypes.Pointer{
		Def: def,
	}, nil
}

func (p *Parser) parseChan(typedExpr *ast.ChanType) (*gotypes.Channel, error) {
	def, err := p.Parse(typedExpr.Value)
	if err != nil {
		return nil, err
	}

	channel := &gotypes.Channel{
		Value: def,
	}

	switch typedExpr.Dir {
	case ast.SEND:
		channel.Dir = "1"
	case ast.RECV:
		channel.Dir = "2"
	default:
		channel.Dir = "3"
	}

	return channel, nil
}

func (p *Parser) parseEllipsis(typedExpr *ast.Ellipsis) (*gotypes.Ellipsis, error) {
	// X.Sel
	def, err := p.Parse(typedExpr.Elt)
	if err != nil {
		return nil, err
	}
	return &gotypes.Ellipsis{
		Def: def,
	}, nil
}

func (p *Parser) parseSelector(typedExpr *ast.SelectorExpr) (*gotypes.Selector, error) {
	// X.Sel a.k.a Prefix.Item

	id, ok := typedExpr.X.(*ast.Ident)
	// if the prefix is identifier only, the selector is in a form qui.identifier,
	// i.e. fully qualified identifier
	// TODO(jchaloup): check id.function(), is that selector as well or not?
	if ok {
		// TODO(jchaloup): replace the id.Name with full package path
		// E.g. instead of (qid, identifier) use (github.com/coreos/etcd/pkg/wait, Wait)
		p.allocatedSymbolsTable.AddSymbol(id.Name, typedExpr.Sel.Name)
		return &gotypes.Selector{
			Item: typedExpr.Sel.Name,
			Prefix: &gotypes.Identifier{
				Def: id.Name,
			},
		}, nil
	}

	def, err := p.Parse(typedExpr.X)
	if err != nil {
		return nil, err
	}
	return &gotypes.Selector{
		Item:   typedExpr.Sel.Name,
		Prefix: def,
	}, nil
}

func (p *Parser) parseStruct(typedExpr *ast.StructType) (*gotypes.Struct, error) {
	structType := &gotypes.Struct{}
	structType.Fields = make([]gotypes.StructFieldsItem, 0)

	if typedExpr.Fields.List == nil {
		return structType, nil
	}

	for _, field := range typedExpr.Fields.List {
		// anonymous field?
		if field.Names == nil {
			def, err := p.Parse(field.Type)
			if err != nil {
				return nil, err
			}

			item := gotypes.StructFieldsItem{
				Name: "",
				Def:  def,
			}

			structType.Fields = append(structType.Fields, item)
			// named fields
		} else {
			for _, name := range field.Names {
				def, err := p.Parse(field.Type)
				if err != nil {
					return nil, err
				}

				item := gotypes.StructFieldsItem{
					Name: name.Name,
					Def:  def,
				}
				structType.Fields = append(structType.Fields, item)
			}
		}
	}
	return structType, nil
}

func (p *Parser) parseMap(typedExpr *ast.MapType) (*gotypes.Map, error) {
	keyDef, keyErr := p.Parse(typedExpr.Key)
	if keyErr != nil {
		return nil, keyErr
	}

	valueDef, valueErr := p.Parse(typedExpr.Value)
	if valueErr != nil {
		return nil, valueErr
	}

	return &gotypes.Map{
		Keytype:   keyDef,
		Valuetype: valueDef,
	}, nil
}

func (p *Parser) parseArray(typedExpr *ast.ArrayType) (gotypes.DataType, error) {
	def, err := p.Parse(typedExpr.Elt)
	if err != nil {
		return nil, err
	}
	if typedExpr.Len == nil {
		return &gotypes.Slice{
			Elmtype: def,
		}, nil
	}

	// TODO(jchaloup): store array's length as well
	return &gotypes.Array{
		Elmtype: def,
	}, nil
}

func (p *Parser) parseInterface(typedExpr *ast.InterfaceType) (*gotypes.Interface, error) {
	// TODO(jchaloup): extend the interface definition with embedded interfaces
	interfaceObj := &gotypes.Interface{}
	var methods []gotypes.InterfaceMethodsItem
	for _, m := range typedExpr.Methods.List {
		def, err := p.Parse(m.Type)
		if err != nil {
			return nil, err
		}

		for _, name := range m.Names {
			item := gotypes.InterfaceMethodsItem{
				Name: name.Name,
				Def:  def,
			}
			methods = append(methods, item)
		}
	}
	if len(methods) > 0 {
		interfaceObj.Methods = methods
	}

	return interfaceObj, nil
}

func (p *Parser) parseFunction(typedExpr *ast.FuncType) (*gotypes.Function, error) {
	functionType := &gotypes.Function{}

	var params []gotypes.DataType
	var results []gotypes.DataType

	if typedExpr.Params != nil {
		for _, field := range typedExpr.Params.List {
			def, err := p.Parse(field.Type)
			if err != nil {
				return nil, err
			}

			// field.Names list must be singletion at least
			for i := 0; i < len(field.Names); i++ {
				params = append(params, def)
			}
		}

		if len(params) > 0 {
			functionType.Params = params
		}
	}

	if typedExpr.Results != nil {
		for _, field := range typedExpr.Results.List {
			def, err := p.Parse(field.Type)
			if err != nil {
				return nil, err
			}

			// results can be identifier free
			if len(field.Names) == 0 {
				results = append(results, def)
			} else {
				for i := 0; i < len(field.Names); i++ {
					results = append(results, def)
				}
			}

		}

		if len(results) > 0 {
			functionType.Results = results
		}
	}

	return functionType, nil
}

func (p *Parser) Parse(expr ast.Expr) (gotypes.DataType, error) {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		return p.parseIdentifier(typedExpr)
	case *ast.StarExpr:
		return p.parseStar(typedExpr)
	case *ast.ChanType:
		return p.parseChan(typedExpr)
	case *ast.Ellipsis:
		return p.parseEllipsis(typedExpr)
	case *ast.SelectorExpr:
		return p.parseSelector(typedExpr)
	case *ast.MapType:
		return p.parseMap(typedExpr)
	case *ast.ArrayType:
		return p.parseArray(typedExpr)
	case *ast.StructType:
		return p.parseStruct(typedExpr)
	case *ast.InterfaceType:
		return p.parseInterface(typedExpr)
	case *ast.FuncType:
		return p.parseFunction(typedExpr)
	}
	return nil, fmt.Errorf("ast.Expr (%#v) not recognized", expr)
}

// New creates an instance of the type Parser
func New(packageName string, symbolTable *symboltable.Stack, allocTable *alloctable.Table) types.TypeParser {
	p := &Parser{
		packageName:           packageName,
		symbolTable:           symbolTable,
		allocatedSymbolsTable: allocTable,
	}
	p.symbolTable.Push()
	return p
}
