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
	"flag"

	"github.com/gofed/symbols-extractor/pkg/updater"
)

const (
	destDefault = "github.com/gofed/symbols-extractor/tests"
	destUsage = "set `destination` directory path; " +
		"destination directory must have golang workspace structure " +
		"and its path must be relative to GOPATH"
	srcDefault = "github.com/gofed/symbols-extractor/tests/src/specs"
	srcUsage = "set `source` directory path; " +
		"source directory path must be relative to GOPATH"
)

type TestGenCommand struct {
	name        string
	description string
	flagset     *flag.FlagSet
	destination *string
	source      *string
}

func (cmd *TestGenCommand) CommandName() string {
	return cmd.name
}

func (cmd *TestGenCommand) CommandDescription() string {
	return cmd.description
}

func (cmd *TestGenCommand) Run(argv []string) error {
	return nil
}

func NewTestGenCommand(name, description string) *TestGenCommand {
	flagset := flag.NewFlagSet(name, flag.ExitOnError)
	return &TestGenCommand{
		name: name,
		description: description,
		flagset: flagset,
		destination: flagset.String("dest", destDefault, destUsage),
		source: flagset.String("src", srcDefault, srcUsage),
	}
}

func init() {
	updater.AddCommand(NewTestGenCommand(
		"tests", "update or generate tests from YAML files",
	))
}
