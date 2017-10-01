# Symbols Extractor

![tTravis CI](https://api.travis-ci.org/gofed/symbols-extractor.svg?branch=master)

*Symbols Extractor* is a tool that gathers a lot of useful informations about
symbols used in a given [Go](https://golang.org) project and stores them in the
[JSON](http://www.json.org/) format for further processing.

## Example

TODO

## What kind of informations could symbol carry?

```go
type PackageInfo struct {
	Name     string // name of the package (can be alias)
	FullName string // fully qualified and globally unique name of the package
}

type SourceFile struct {
	Path	 string // globally unique path of the file (e.g. github.com/user/goproject/main.go)
	Revision string // SHA-1 of commit in case of git or empty if not available
	Content	 string // file content
}

// A list of scope names, the higher index the deeper scope.
// A scope name has the following format:
//
//   {<statement_name>, <file_offset>}
//
// <statement_name> is derived as follows:
//
//   "{" -> "block"
//   "if" -> "if"
//   ...
//   "func X() {...}" -> "X"
//   "func (...) {...}" -> "func"
//
// Example:
//
//   sp := []{{"Connect", 1256}, {"switch", 1287}, {"func", 1399}}
//
type ScopePath []struct{
	Name     string
	Position int
}

type Label interface {
	LabelTag()
	Name() string
}

type Scope struct {
	Label      Label
	Parent     *Scope
	Siblings   []*Scope
	Begin, End int
	Symbols    []*Symbol
}

type LocationInfo struct {
	PackageInfoRef *PackageInfo // object's package
	SourceFileRef  *SourceFile  // object's source file
	Position       int          // object's offset from the source file beginning
	Scope          *Scope       // object's scope
}

type TypeInfo struct {
	Location *LocationInfo
	Repr
}

type Symbol struct {
	Location       *LocationInfo // where I am
	Name           string        // my name
	FullName       string        // my fully qualified name, if I am visible from outside (if not, it should be empty)
	BounderRef     *Symbol       // nil if I am not bounded or pointer to bounder
	Origin         *Symbol       // where I was first defined (nil if I am the one and only)
	IsConst        bool          // am I constant?
	TypeRef        *TypeInfo     // what I am (nil if I am mysterious)
}

type AstVisitor interface {
	Visit()
}

type SymbolExtractor struct {
}

func (se *SymbolExtractor) Visit() {
}
```
