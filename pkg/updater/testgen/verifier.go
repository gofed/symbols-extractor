//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/testgen/verifier.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-25 10:39:26 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Test specification verifying tools.
//
package testgen

import (
	"strings"
	"path/filepath"

	"github.com/gofed/symbols-extractor/pkg/updater"
)

func (ld *TestSpecLoader) VerifyBatch(batch_path string, batch *TestSpecBatch) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	assertBatchIsNotEmpty(ld, batch_path, batch)
	return nil
}

func assertBatchIsNotEmpty(ld *TestSpecLoader, p string, b *TestSpecBatch) {
	if *ld.cmd.TsStrict && len(b.All) == 0 {
		m := "%q has not items to be processed."
		updater.Warning(ld.cmd.CommandName(), m, p)
	}
}

func (ld *TestSpecLoader) VerifyTestSpec(ts_path string, ts *TestSpecMain) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	ld.VerifyTestSpecMain(ts_path, ts)
	return nil
}

func (ld *TestSpecLoader) VerifyTestSpecMain(ts_path string, ts *TestSpecMain) {
	ld.VerifyTestSpecVersion(ts_path, ts)
	ld.VerifyTestSpecInfo(ts_path, ts)
	ld.VerifyTestSpecConfig(ts_path, ts)
	ld.VerifyTestSpecGroups(ts_path, ts)
	ld.VerifyTestSpecTests(ts_path, ts)
}

func (ld *TestSpecLoader) VerifyTestSpecVersion(ts_path string, ts *TestSpecMain) {
	assertTestGenVersion(ts_path, getVersion(ts_path, ts))
}

func getVersion(p string, ts *TestSpecMain) *updater.Version {
	v, err := updater.ToVersion(ts.Version)
	if err != nil {
		panic(updater.NewUpdaterError("%q: %s", p, err))
	}
	return v
}

func assertTestGenVersion(p string, v *updater.Version) {
	if updater.Compare(v, &TestSpecVersion) > 0 {
		m := "at least %s version of test generator engine is required"
		m += " to process %q; current version is %s!"
		panic(updater.NewUpdaterError(m, v, p, &TestSpecVersion))
	}
}

func (ld *TestSpecLoader) VerifyTestSpecInfo(ts_path string, ts *TestSpecMain) {
	assertInfo(ts_path, ts)
	info := ts.Info
	info.Name = strings.TrimSpace(info.Name)
	assertInfoName(ts_path, info)
	assertInfoAuthor(ts_path, info)
	for i, author := range(info.Author) {
		author = strings.TrimSpace(author)
		assertInfoAuthorItem(ts_path, author, i)
		info.Author[i] = author
	}
	info.Version = strings.TrimSpace(info.Version)
	assertInfoVersion(ts_path, info)
	info.Comment = strings.TrimSpace(info.Comment)
	if info.Comment != "" {
		info.Comment += "\n"
	}
}

func assertInfo(p string, ts *TestSpecMain) {
	if ts.Info == nil {
		m := "%q: \"info\" field is required!"
		panic(updater.NewUpdaterError(m, p))
	}
}

func assertInfoName(p string, info *TestSpecInfo) {
	if info.Name == "" {
		m := "%q: \"info.name\" field is required!"
		panic(updater.NewUpdaterError(m, p))
	}
}

func assertInfoAuthor(p string, info *TestSpecInfo) {
	if len(info.Author) == 0 {
		m := "%q: \"info.author\" field is required!"
		panic(updater.NewUpdaterError(m, p))
	}
}

func assertInfoAuthorItem(p, a string, i int) {
	if a == "" {
		m := "%q: info.author: #%v item is empty!"
		panic(updater.NewUpdaterError(m, p, i))
	}
}

func assertInfoVersion(p string, info *TestSpecInfo) {
	_, err := updater.ToVersion(info.Version)
	if err != nil {
		m := "%q: info.version: %s"
		panic(updater.NewUpdaterError(m, p, err))
	}
}

func (ld *TestSpecLoader) VerifyTestSpecConfig(ts_path string, ts *TestSpecMain) {
	if ts.Config == nil {
		ts.Config = &TestSpecConfig{}
	}
	d := filepath.Dir(ts_path)
	config := ts.Config
	config.Commands = strings.TrimSpace(config.Commands)
	err := ld.LoadLibraryFromDirByName(d, config.Commands, "command")
	if err != nil {
		panic(err)
	}
	config.Templates = strings.TrimSpace(config.Templates)
	err = ld.LoadLibraryFromDirByName(d, config.Templates, "template")
	if err != nil {
		panic(err)
	}
}

func (ld *TestSpecLoader) VerifyTestSpecGroups(ts_path string, ts *TestSpecMain) {
	for i, gp := range ts.Groups {
		if gp == nil {
			continue
		}
		k := strings.TrimSpace(gp.Name)
		assertGroupName(ld, ts_path, k, i)
		assertGroupUndefined(ld, ts_path, k, i)
		ld.Groups[k] = make([]string, 0)
		var m = make(map[string]bool)
		for j, sg := range gp.Subgroups {
			sg = strings.TrimSpace(sg)
			if !assertSubgroupName(ld, ts_path, sg, j, k, i) {
				continue
			}
			if !assertSubgroupUnique(ld, ts_path, m, sg, j, k, i) {
				continue
			}
			ld.Groups[k] = append(ld.Groups[k], sg)
			m[sg] = true
		}
	}
	assertAllGroupsAreDefined(ld, ts_path)
}

func assertGroupName(ld *TestSpecLoader, p, g string, i int) {
	if g == "" {
		m := "%q: groups: group #%v has empty name%s"
		if *ld.cmd.TsStrict {
			panic(updater.NewUpdaterError(m, p, i, "!"))
		} else {
			updater.Warning(ld.cmd.CommandName(), m, p, i, ".")
		}
	}
}

func assertGroupUndefined(ld *TestSpecLoader, p, g string, i int) {
	if _, ok := ld.Groups[g]; ok {
		m := "%q: groups: group #%v (%q) is already defined%s"
		if *ld.cmd.TsStrict {
			panic(updater.NewUpdaterError(m, p, i, g, "!"))
		} else {
			updater.Warning(ld.cmd.CommandName(), m, p, i, g, ".")
		}
	}
}

func assertSubgroupName(ld *TestSpecLoader, p, sg string, j int, g string, i int) bool {
	if sg == "" {
		m := "%q: groups: subgroup #%v of group #%v (%q) has empty"
		m += " name%s"
		if *ld.cmd.TsStrict {
			panic(updater.NewUpdaterError(m, p, j, i, g, "!"))
		} else {
			c := ld.cmd.CommandName()
			updater.Warning(c, m, p, j, i, g, ".")
		}
		return false
	}
	return true
}

func assertSubgroupUnique(ld *TestSpecLoader, p string, m map[string]bool, sg string, j int, g string, i int) bool {
	if _, ok := m[sg]; ok {
		m := "%q: groups: subgroup #%v (%q) of group #%v (%q) was"
		m += " introduced before in this list%s"
		if *ld.cmd.TsStrict {
			panic(updater.NewUpdaterError(m, p, j, sg, i, g, "!"))
		} else {
			c := ld.cmd.CommandName()
			updater.Warning(c, m, p, j, sg, i, g, ".")
		}
		return false
	}
	return true
}

func assertAllGroupsAreDefined(ld *TestSpecLoader, p string) {
	m := "%q: groups.%s.#%v: Undefined group %q!"
	for g, sgs := range ld.Groups {
		for i, sg := range sgs {
			if _, ok := ld.Groups[sg]; !ok {
				panic(updater.NewUpdaterError(m, p, g, i, sg))
			}
		}
	}
}

func (ld *TestSpecLoader) VerifyTestSpecTests(ts_path string, ts *TestSpecMain) {
	assertTests(ts_path, ts)
	for i, t := range ts.Tests {
		ld.VerifyOneTest(ts_path, i, t)
	}
}

func assertTests(p string, ts *TestSpecMain) {
	if len(ts.Tests) == 0 {
		m := "there are no tests in %q!"
		panic(updater.NewUpdaterError(m, p))
	}
}

func (ld *TestSpecLoader) VerifyOneTest(ts_path string, i int, t *TestSpecTest) {
	if t.Generator != "" {
		t.Generator = strings.TrimSpace(t.Generator)
		if t.Generator != "" {
			t.Generator += "\n"
		}
		assertTestGenerator(ts_path, t, i)
		assertTestOnlyGenerator(ts_path, t, i)
		return
	}
	t.Name = strings.TrimSpace(t.Name)
	assertTestName(ts_path, t, i)
	t.Group = strings.TrimSpace(t.Group)
	assertTestGroup(ld, ts_path, t, i)
	t.Comment = strings.TrimSpace(t.Comment)
	if t.Comment != "" {
		t.Comment += "\n"
	}
	ld.VerifyTestFileset(ts_path, i, t)
	ld.VerifyTestTesters(ts_path, i, t)
}

func assertTestGenerator(p string, t *TestSpecTest, i int) {
	if t.Generator == "" {
		m := "%q: tests.#%v: \"generate\" field has no value!"
		panic(updater.NewUpdaterError(m, p, i))
	}
}

func assertTestOnlyGenerator(p string, t *TestSpecTest, i int) {
	if t.Name != "" || t.Group != "" || t.Comment != "" ||
	len(t.Fileset) > 0 || len(t.Testers) > 0 {
		m := "%q: tests.#%v: \"generate\" is mixed with other fields!"
		panic(updater.NewUpdaterError(m, p, i))
	}
}

func assertTestName(p string, t *TestSpecTest, i int) {
	if t.Name == "" {
		m := "%q: tests.#%v: \"name\" field is required!"
		panic(updater.NewUpdaterError(m, p, i))
	}
}

func assertTestGroup(ld *TestSpecLoader, p string, t *TestSpecTest, i int) {
	if t.Group != "" {
		if _, ok := ld.Groups[t.Group]; !ok {
			m := "%q: tests.#%v.group: undefined group %q!"
			panic(updater.NewUpdaterError(m, p, i, t.Group))
		}
	}
}

func (ld *TestSpecLoader) VerifyTestFileset(ts_path string, i int, t *TestSpecTest) {
	assertTestFileset(ts_path, t, i)
	for j, f := range t.Fileset {
		if f.Generator != "" {
			f.Generator = strings.TrimSpace(f.Generator)
			if f.Generator != "" {
				f.Generator += "\n"
			}
			assertTestFilesetGenerator(ts_path, f, i, j)
			assertTestFilesetOnlyGenerator(ts_path, f, i, j)
			continue
		}
		f.Name = strings.TrimSpace(f.Name)
		assertTestFilesetName(ts_path, f, i, j)
		f.Version = strings.TrimSpace(f.Version)
		assertTestFilesetVersion(ts_path, f, i, j)
		assertTestFilesetContent(ts_path, f, i, j)
		switch v := f.Content.(type) {
		case string:
			v = strings.TrimSpace(v)
			if v != "" {
				v += "\n"
			}
			assertFileContentString(ts_path, v, i, j)
			f.Content = v
		case map[string]string:
			assertFileContentGenerator(ts_path, v, i, j)
			v["generate"] = strings.TrimSpace(v["generate"])
			if v["generate"] != "" {
				v["generate"] += "\n"
			}
			assertFileContentGeneratorValue(ts_path, v, i, j)
			f.Content = v
		default:
			badFileContentType(ts_path, v, i, j)
		}
	}
}

func assertTestFileset(p string, t *TestSpecTest, i int) {
	if len(t.Fileset) == 0 {
		m := "%q: tests.#%v: empty or no fileset!"
		panic(updater.NewUpdaterError(m, p, i))
	}
}

func assertTestFilesetGenerator(p string, f *TestSpecFileDef, i, j int) {
	if f.Generator == "" {
		m := "%q: tests.#%v.fileset.#%v:"
		m += " \"generate\" field has no value!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTestFilesetOnlyGenerator(p string, f *TestSpecFileDef, i, j int) {
	if f.Name != "" || f.Version != "" || f.Content != nil {
		m := "%q: tests.#%v.fileset.#%v:"
		m += " \"generate\" is mixed with other fields!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTestFilesetName(p string, f *TestSpecFileDef, i, j int) {
	if f.Name == "" {
		m := "%q: tests.#%v.fileset.#%v: \"name\" field is required!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTestFilesetVersion(p string, f *TestSpecFileDef, i, j int) {
	_, err := updater.ToVersion(f.Version)
	if err != nil {
		m := "%q: tests.#%v.fileset.#%v.version: %s"
		panic(updater.NewUpdaterError(m, p, i, j, err))
	}
}

func assertTestFilesetContent(p string, f *TestSpecFileDef, i, j int) {
	if f.Content == nil {
		m := "%q: tests.#%v.fileset.#%v:"
		m += " \"content\" field is required!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertFileContentString(p, v string, i, j int) {
	if v == "" {
		m := "%q: tests.#%v.fileset.#%v.content: empty file content!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertFileContentGenerator(p string, v map[string]string, i, j int) {
	if _, ok := v["generate"]; len(v) != 1 || !ok {
		m := "%q: tests.#%v.fileset.#%v.content: ill-formed generator!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertFileContentGeneratorValue(p string, v map[string]string, i, j int) {
	if v["generate"] == "" {
		m := "%q: tests.#%v.fileset.#%v.content: empty generator!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func badFileContentType(p string, v interface{}, i, j int) {
	m := "%q: tests.#%v.fileset.#%v.content:"
	m += " bad type of file content (%T)!"
	panic(updater.NewUpdaterError(m, p, i, j, v))
}

func (ld *TestSpecLoader) VerifyTestTesters(ts_path string, i int, t *TestSpecTest) {
	assertTestTesters(ts_path, t, i)
	for j, tt := range t.Testers {
		if tt.Generator != "" {
			tt.Generator = strings.TrimSpace(tt.Generator)
			if tt.Generator != "" {
				tt.Generator += "\n"
			}
			assertTestTesterGenerator(ts_path, tt, i, j)
			assertTestTesterOnlyGenerator(ts_path, tt, i, j)
			continue
		}
		tt.Command = strings.TrimSpace(tt.Command)
		if tt.Command != "" {
			tt.Command += "\n"
		}
		assertTestTesterCommand(ts_path, tt, i, j)
		assertTestTesterExpectedResult(ts_path, tt, i, j)
		switch v := tt.ExpectedResult.(type) {
		case string:
			v = strings.TrimSpace(v)
			if v != "" {
				v += "\n"
			}
			assertTesterExpectedResultString(ts_path, v, i, j)
			tt.ExpectedResult = v
		case map[string]string:
			assertTesterExpectedResultGenerator(ts_path, v, i, j)
			v["generate"] = strings.TrimSpace(v["generate"])
			if v["generate"] != "" {
				v["generate"] += "\n"
			}
			assertTesterExpectedResultGeneratorValue(
				ts_path, v, i, j,
			)
			tt.ExpectedResult = v
		default:
			badTesterExpectedResultType(ts_path, v, i, j)
		}
	}
}

func assertTestTesters(p string, t *TestSpecTest, i int) {
	if len(t.Testers) == 0 {
		m := "%q: tests.#%v: missing testers!"
		panic(updater.NewUpdaterError(m, p, i))
	}
}

func assertTestTesterGenerator(p string, tt *TestSpecTester, i, j int) {
	if tt.Generator == "" {
		m := "%q: tests.#%v.testers.#%v: empty generator!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTestTesterOnlyGenerator(p string, tt *TestSpecTester, i, j int) {
	if tt.Command != "" || tt.ExpectedResult != nil {
		m := "%q: tests.#%v.testers.#%v: generator is mixed with other"
		m += " fields!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTestTesterCommand(p string, tt *TestSpecTester, i, j int) {
	if tt.Command == "" {
		m := "%q: tests.#%v.testers.#%v: \"command\" field is"
		m += " required!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTestTesterExpectedResult(p string, tt *TestSpecTester, i, j int) {
	if tt.ExpectedResult == nil {
		m := "%q: tests.#%v.testers.#%v: \"expected-result\" field is"
		m += " required!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTesterExpectedResultString(p, v string, i, j int) {
	if v == "" {
		m := "%q: tests.#%v.testers.#%v.expected-result: this field"
		m += " should not be empty!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTesterExpectedResultGenerator(p string, v map[string]string, i, j int) {
	if _, ok := v["generate"]; len(v) != 1 || !ok {
		m := "%q: tests.#%v.testers.#%v.expected-result: ill-formed"
		m += " generator!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func assertTesterExpectedResultGeneratorValue(p string, v map[string]string, i, j int) {
	if v["generate"] == "" {
		m := "%q: tests.#%v.testers.#%v.expected-result: empty"
		m += " generator!"
		panic(updater.NewUpdaterError(m, p, i, j))
	}
}

func badTesterExpectedResultType(p string, v interface{}, i, j int) {
	m := "%q: tests.#%v.testers.#%v.expected-result: bad type of expected"
	m += " result (%T)!"
	panic(updater.NewUpdaterError(m, p, i, j, v))
}
