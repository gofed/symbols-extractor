//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/testgen/loader.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-23 12:58:13 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Test specification loader.
//
package testgen

import (
	"fmt"
	"strings"
	"os"
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/gofed/symbols-extractor/pkg/updater"
)

type TestSpecLoader struct {
	cmd       *TestGenCommand
	nfiles    int
	TestSpecs []*TestSpecMain
	Groups    map[string][]string
	Commands  map[string]string
	Templates map[string]string
}

type TestSpecBatch struct {
	All []string `yaml:"all"`
}

type TestSpecMain struct {
	Version string           `yaml:"spec-version"`
	Info    *TestSpecInfo    `yaml:"info"`
	Config  *TestSpecConfig  `yaml:"config"`
	Groups  []*TestSpecGroup `yaml:"groups"`
	Tests   []*TestSpecTest  `yaml:"tests"`
}

type TestSpecInfo struct {
	Name    string   `yaml:"name"`
	Author  []string `yaml:"author"`
	Version string   `yaml:"version"`
	Comment string   `yaml:"comment"`
}

type TestSpecConfig struct {
	Commands  string `yaml:"commands"`
	Templates string `yaml:"templates"`
}

type TestSpecGroup struct {
	Name      string   `yaml:"name"`
	Subgroups []string `yaml:"subgroups"`
}

type TestSpecTest struct {
	Name      string             `yaml:"name"`
	Group     string             `yaml:"group"`
	Comment   string             `yaml:"comment"`
	Fileset   []*TestSpecFileDef `yaml:"fileset"`
	Testers   []*TestSpecTester  `yaml:"testers"`
	Generator string             `yaml:"generate"`
}

type TestSpecFileDef struct {
	Name      string      `yaml:"name"`
	Version   string      `yaml:"version"`
	Content   interface{} `yaml:"content"`
	Generator string      `yaml:"generate"`
}

type TestSpecTester struct {
	Command        string      `yaml:"command"`
	ExpectedResult interface{} `yaml:"expected-result"`
	Generator      string      `yaml:"generate"`
}

type TestSpecLibrary struct {
	Name    string              `yaml:"name"`
	Entries []*TestSpecLibEntry `yaml:"entries"`
}

type TestSpecLibEntry struct {
	Name       string `yaml:"name"`
	Definition string `yaml:"definition"`
}

var TestSpecVersion = updater.Version{1, 0, 0}

func (ld *TestSpecLoader) LoadTestSpec(path string) error {
	// 1) Canonize the source path:
	cpath, err := updater.CanonizePath(path)
	if err != nil {
		return err
	}
	// 2) Resolve path:
	fpath, isbatch, err := ld.ResolvePath(cpath)
	if err != nil {
		return err
	}
	// 3) Load:
	if isbatch {
		return ld.LoadTestSpecFromBatch(fpath)
	}
	return ld.LoadTestSpecFile(fpath)
}

func (ld *TestSpecLoader) ResolvePath(path string) (fpath string, isbatch bool, err error) {
	fpath, isdir, err := updater.NeedsDirOrYamlFile(
		path, *ld.cmd.TsForceExt,
	)
	if err != nil {
		return fpath, false, err
	}
	// It is a file?
	if !isdir {
		return fpath, false, nil
	}
	// It is a directory.
	// Look for batch file first.
	if *ld.cmd.TsBatch != "" {
		var oldfpath = fpath
		fpath = filepath.Join(fpath, *ld.cmd.TsBatch)
		fpath, err = updater.NeedsYamlFile(fpath, *ld.cmd.TsForceExt)
		if err == nil {
			return fpath, true, nil
		}
		// Batch file not found.
		fpath = oldfpath
	}
	// Batch file not found or ignored.
	fpath = filepath.Join(fpath, *ld.cmd.TsMain)
	fpath, err = updater.NeedsYamlFile(fpath, *ld.cmd.TsForceExt)
	return fpath, false, err
}

func (ld *TestSpecLoader) LoadTestSpecFromBatch(path string) error {
	sbatch, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// Helps to break cyclic dependencies between files.
	ld.nfiles++
	defer func() { ld.nfiles-- }()
	if ld.nfiles > *ld.cmd.FiDepth {
		return updater.NewUpdaterError(
			"Maximal inclusion file depth (%v) was exceeded!",
			*ld.cmd.FiDepth,
		)
	}
	var batch TestSpecBatch
	if *ld.cmd.TsStrict {
		err = yaml.UnmarshalStrict(sbatch, &batch)
	} else {
		err = yaml.Unmarshal(sbatch, &batch)
	}
	if err != nil {
		return err
	}
	err = ld.VerifyBatch(path, &batch)
	if err != nil {
		return err
	}
	d := filepath.Dir(path)
	for _, p := range(batch.All) {
		pp := strings.TrimSpace(p)
		if pp == "" {
			return updater.NewUpdaterError(
				"%q contains an empty record!", path,
			)
		}
		err := ld.LoadTestSpec(filepath.Join(d, pp))
		if err != nil {
			return err
		}
	}
	return nil
}

func (ld *TestSpecLoader) LoadTestSpecFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// Helps to break cyclic dependencies between files.
	ld.nfiles++
	defer func() { ld.nfiles-- }()
	if ld.nfiles > *ld.cmd.FiDepth {
		return updater.NewUpdaterError(
			"Maximal inclusion file depth (%v) was exceeded!",
			*ld.cmd.FiDepth,
		)
	}
	var tsmain TestSpecMain
	if *ld.cmd.TsStrict {
		err = yaml.UnmarshalStrict(data, &tsmain)
	} else {
		err = yaml.Unmarshal(data, &tsmain)
	}
	if err != nil {
		return err
	}
	err = ld.VerifyTestSpec(path, &tsmain)
	if err != nil {
		return err
	}
	ld.TestSpecs = append(ld.TestSpecs, &tsmain)
	return nil
}

func (ld *TestSpecLoader) LoadLibraryFromDirByName(dir, name, storage string) error {
	var soft = false
	if name == "" {
		name = fmt.Sprintf("%ss", storage)
		soft = true
	}
	return ld.LoadLibraryFromPath(filepath.Join(dir, name), storage, soft)
}

func (ld *TestSpecLoader) LoadLibraryFromPath(path, storage string, soft bool) error {
	p, isdir, err := updater.NeedsDirOrYamlFile(path, *ld.cmd.TsForceExt)
	if err != nil {
		if soft {
			return nil
		}
		return err
	}
	if !isdir {
		err := ld.LoadLibrary(p, storage)
		if soft {
			return nil
		}
		return err
	}
	err = filepath.Walk(p, func (pp string, ii os.FileInfo, ee error) error {
		if ii != nil && ii.IsDir() {
			return filepath.SkipDir
		}
		if ee != nil {
			if soft {
				return nil
			}
			return ee
		}
		ee = ld.LoadLibrary(pp, storage)
		if soft {
			return nil
		}
		return ee
	})
	if soft {
		return nil
	}
	return err
}

func (ld *TestSpecLoader) LoadLibrary(path, storage string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	var lib TestSpecLibrary
	if *ld.cmd.TsStrict {
		err = yaml.UnmarshalStrict(data, &lib)
	} else {
		err = yaml.Unmarshal(data, &lib)
	}
	ns := strings.TrimSpace(lib.Name)
	for _, x := range(lib.Entries) {
		name := strings.TrimSpace(x.Name)
		if name == "" {
			return updater.NewUpdaterError(
				"Empty %s name!", storage,
			)
		}
		if ns != "" {
			name = fmt.Sprintf("%s.%s", ns, name)
		}
		defn := strings.TrimSpace(x.Definition)
		if defn == "" {
			return updater.NewUpdaterError(
				"Empty %s definition!", storage,
			)
		} else {
			defn += "\n"
		}
		switch storage {
		case "command":
			ld.Commands[name] = defn
		case "template":
			ld.Templates[name] = defn
		default:
			return updater.NewUpdaterError(
				"Unknown storage name (%s)!", storage,
			)
		}
	}
	return nil
}

func (cmd *TestGenCommand) NewTestSpecLoader() *TestSpecLoader {
	return &TestSpecLoader{
		cmd: cmd,
		TestSpecs: make([]*TestSpecMain, 0),
		Groups: make(map[string][]string),
		Commands: make(map[string]string),
		Templates: make(map[string]string),
	}
}
