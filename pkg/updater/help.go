//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/help.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-23 10:53:17 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Updater's help command.
//
package updater

import (
	"os"
	"flag"
)

type HelpCommand struct {
	name        string
	description string
	flags       *flag.FlagSet
	help        *bool
}

func (cmd *HelpCommand) CommandName() string {
	return cmd.name
}

func (cmd *HelpCommand) CommandDescription() string {
	return cmd.description
}

func (cmd *HelpCommand) Run(argv []string) error {
	cmd.flags.Parse(argv)
	if *cmd.help {
		cmd.flags.Usage()
		return nil
	}
	PrintBanner(os.Stdout)
	PrintUsage(os.Stdout)
	return nil
}

func NewHelpCommand(name, description string) Command {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	return &HelpCommand{
		name: name,
		description: description,
		flags: flags,
		help: flags.Bool("help", false, "print this screen"),
	}
}

func init() {
	AddCommand(NewHelpCommand("help", "print this screen"))
}
