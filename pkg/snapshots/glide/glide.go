package glide

import (
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type GlideImport struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Glide struct {
	Hash        string        `json:"hash"`
	Updated     string        `json:"updated"`
	Imports     []GlideImport `json:"imports"`
	TestImports []GlideImport `json:"testImports"`

	importsList   map[string]string
	mainPkg       string
	mainPkgCommit string
}

func GlideFromFile(file string) (*Glide, error) {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to load file %v: %v", file, err)
	}

	var glide Glide
	err = yaml.Unmarshal(yamlFile, &glide)
	if err != nil {
		panic(err)
	}

	glide.importsList = make(map[string]string)
	for _, item := range glide.Imports {
		glide.importsList[item.Name] = item.Version
	}

	return &glide, nil
}

func (g *Glide) Commit(pkg string) (string, error) {
	if g.mainPkg != "" && g.mainPkgCommit != "" && strings.HasPrefix(pkg, g.mainPkg) {
		return g.mainPkgCommit, nil
	}
	if commit, ok := g.importsList[pkg]; ok {
		return commit, nil
	}
	return "", fmt.Errorf("Commit for package %q not found", pkg)
}

func (g *Glide) MainPackageCommit(pkg string, commit string) {
	g.mainPkg = pkg
	g.mainPkgCommit = commit
}
