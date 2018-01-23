package godeps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type Dep struct {
	ImportPath string
	Rev        string
	Comment    string
}

type Godeps struct {
	ImportPath   string
	GoVersion    string
	GodepVersion string
	Packages     []string
	Deps         []Dep

	importsList   map[string]string
	mainPkg       string
	mainPkgCommit string
}

func FromFile(file string) (*Godeps, error) {
	text, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to load file %v: %v", file, err)
	}

	var godeps Godeps
	err = json.Unmarshal(text, &godeps)
	if err != nil {
		panic(err)
	}

	godeps.importsList = make(map[string]string)
	for _, item := range godeps.Deps {
		godeps.importsList[item.ImportPath] = item.Rev
	}

	return &godeps, nil
}

func (g *Godeps) Commit(pkg string) (string, error) {
	if g.mainPkg != "" && g.mainPkgCommit != "" && strings.HasPrefix(pkg, g.mainPkg) {
		return g.mainPkgCommit, nil
	}
	if commit, ok := g.importsList[pkg]; ok {
		return commit, nil
	}
	return "", fmt.Errorf("Commit for package %q not found", pkg)
}

func (g *Godeps) MainPackageCommit(pkg string, commit string) {
	g.mainPkg = pkg
	g.mainPkgCommit = commit

	g.importsList[pkg] = commit
}
