//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/testgen/init.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-21 12:31:20 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Implementation of `update tests` command.
//
package testgen

import (
	"fmt"
	"flag"

	"github.com/gofed/symbols-extractor/pkg/updater"
)

const (
	helpDefault = false
	helpUsage = "print this screen"
	destDefault = "github.com/gofed/symbols-extractor/tests"
	destUsage = "set destination directory `path`; " +
		"destination directory must have golang workspace structure " +
		"and its path must be relative to GOPATH"
	srcDefault = "github.com/gofed/symbols-extractor/tests/src/specs"
	srcUsage = "set source directory `path`; " +
		"source directory path must be relative to GOPATH"
	fidpDefault = 16
	fidpUsage = "maximal `number` of the file inclusion depth level"
	tsmainDefault = "tests.yml"
	tsmainUsage = "set the name of the main `file` the test generator " +
		"is looking for"
	tsbatchDefault = "all.yml"
	tsbatchUsage = "set the name of the batch `file`; the empty value " +
		"means that no batch file will be used"
	tsforceextDefault = false
	tsforceextUsage = "requires .yml suffix for all YAML files"
	tsstrictDefault = false
	tsstrictUsage = "loads YAML files under the strict mode, i.e. all " +
		"unknown keys will be treated as errors"
)

type TestGenCommand struct {
	name        string
	description string
	flags       *flag.FlagSet
	HelpFlag    *bool
	Destination *string
	Source      *string
	FiDepth     *int
	TsMain      *string
	TsBatch     *string
	TsForceExt  *bool
	TsStrict    *bool
}

func (cmd *TestGenCommand) CommandName() string {
	return cmd.name
}

func (cmd *TestGenCommand) CommandDescription() string {
	return cmd.description
}

func (cmd *TestGenCommand) Help() {
	cmd.flags.Usage()
}

func (cmd *TestGenCommand) Run(argv []string) error {
	cmd.flags.Parse(argv)
	if *cmd.HelpFlag {
		cmd.Help()
		return nil
	}
	loader := cmd.NewTestSpecLoader()
	err := loader.LoadTestSpec(*cmd.Source)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", loader.TestSpecs[0])
	return nil
}

func NewTestGenCommand(name, description string) *TestGenCommand {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	return &TestGenCommand{
		name: name,
		description: description,
		flags: flags,
		HelpFlag: flags.Bool("help", helpDefault, helpUsage),
		Destination: flags.String("dest", destDefault, destUsage),
		Source: flags.String("src", srcDefault, srcUsage),
		FiDepth: flags.Int("fidp", fidpDefault, fidpUsage),
		TsMain: flags.String("main", tsmainDefault, tsmainUsage),
		TsBatch: flags.String("batch", tsbatchDefault, tsbatchUsage),
		TsForceExt: flags.Bool(
			"force-ext", tsforceextDefault, tsforceextUsage,
		),
		TsStrict: flags.Bool(
			"strict", tsstrictDefault, tsstrictUsage,
		),
	}
}

func init() {
	updater.AddCommand(NewTestGenCommand(
		"tests", "update or generate tests from YAML files",
	))
}
