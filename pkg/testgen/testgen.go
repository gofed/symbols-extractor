package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	path "path/filepath"
	"strings"
	"text/template"
	//"github.com/tonnerre/golang-pretty"
)

// generated test files use this prefix
const prefix string = "gen_"

var builtinTypes = map[string]struct{}{
	"uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
	"int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {},
	"float32": {}, "float64": {},
	"complex64": {}, "complex128": {},
	"string": {}, "byte": {}, "rune": {},
	"chan": {}, "bool": {},
	"uintptr": {}, "error": {},
}

func isBuiltin(s string) bool {
	_, ok := builtinTypes[s]

	return ok
}

func BuiltinOrIdent(s string) string {
	if isBuiltin(s) {
		return "Builtin"
	}

	return "Identifier"
}

/******* define types and parse YAML file *******/

type Test struct {
	Code     string `yaml:"code"`
	Expected string `yaml:"expected"`
}

type Target struct {
	Target string `yaml:"target"`
	Tests  []Test `yaml:"tests"`
}

type TestRecipe struct {
	SymTab  []string `yaml:"symboltable"`
	Targets []Target `yaml:"targets"`
}

func (t *TestRecipe) ParseYaml(fname string) error {
	// load and parse YAML
	yamlFile, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("yamlFile.Get: #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, t)
	if err != nil {
		return fmt.Errorf("Unmarshal YAML file: %v", err)
	}

	return nil
}

func genTestFile(tmplPath string, testRecipe TestRecipe) error {
	// Generate file with unit tests corresponding to the given testRecipe.
	//
	var dstF *os.File
	pkgName := path.Base(path.Dir(tmplPath))
	testFileName := path.Join(path.Dir(tmplPath),
		prefix+pkgName+"_test.go")
	funcMap := template.FuncMap{
		"tittle":         strings.Title,
		"isBuiltin":      isBuiltin,
		"BuiltinOrIdent": BuiltinOrIdent,
	}

	tmplt, err := template.New("test.tmpl").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	dstF, err = os.OpenFile(testFileName, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	defer dstF.Close()

	err = tmplt.Execute(dstF, testRecipe)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "fmt", testFileName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("go fmt: %v\n==== Stderr:\n%s====\n", err, stderr.String())
	}

	return nil
}

func getPkgPath() string {
	// return absolute path corresponding to given CLI argument
	// 1. When ABS path is used, it is kept as it is.
	// 2. When script is processed without any input parameter, it is used
	//    current working directory.
	// 3. Otherwise input is accepted as the existing package.
	//
	// In case of wrong path, the script will crash on loading of YAML
	// recipe.
	var pkgPath string

	if len(os.Args) > 1 {
		// path is specified
		if path.IsAbs(os.Args[1]) {
			pkgPath = os.Args[1]
		} else {
			pkgPath = path.Join(os.Getenv("GOPATH"), "src", os.Args[1])
		}
	} else {
		pkgPath, _ = path.Abs(".")
	}

	return pkgPath
}

/* TODO(pstodulk):
 * -? fill symboltable - troubles with function
 * - complete templates (binexp complete)
 */

func main() {
	pkgPath := getPkgPath()
	ymlPath := path.Join(pkgPath, "test.yml")
	tmplPath := path.Join(pkgPath, "test.tmpl")

	var testRecipe TestRecipe
	if err := testRecipe.ParseYaml(ymlPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if err := genTestFile(tmplPath, testRecipe); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	return
}
