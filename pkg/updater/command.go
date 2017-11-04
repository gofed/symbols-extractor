//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/command.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-19 14:58:30 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Updater's command interface.
//
package updater

import (
	"os"
	"io"
	"fmt"
)

type Command interface {
	CommandName() string
	CommandDescription() string
	Help()
	Run(argv []string) error
}

var commands map[string]Command

func init() {
	commands = make(map[string]Command)
}

func AddCommand(cmd Command) {
	if cmd != nil {
		commands[cmd.CommandName()] = cmd
	}
}

func GetCommand(name string) Command {
	return commands[name]
}

func GetCommandByArgv(argv []string) (cmd Command, err error) {
	if len(argv) < 2 {
		return nil, NewUpdaterError(ErrNArgs)
	}
	cmd = GetCommand(argv[1])
	if cmd == nil {
		return nil, NewUpdaterError(ErrCmdNotFound, argv[1])
	}
	return cmd, nil
}

const (
	banner = "%s - updates Symbols Extractor sources.\n"
	usagePrologue = "Usage: %s <command> <flags>\n" +
		"where <command> is one of:\n\n"
	usageEpilogue = "\nFor the detailed informations about <command>," +
		" type `<command> --help`.\n"
)

func Warning(name, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[Warning] %s: %s\n", name, s)
}

func PrintError(name string, err error) {
	fmt.Fprintf(os.Stderr, "[Error] %s: %s\n", name, err)
}

func PrintBanner(output io.Writer) {
	fmt.Fprintf(output, banner, os.Args[0])
}

func PrintUsage(output io.Writer) {
	fmt.Fprintf(output, usagePrologue, os.Args[0])
	for _, cmd := range(commands) {
		fmt.Fprintf(output, "    %s\t%s\n",
			cmd.CommandName(), cmd.CommandDescription(),
		)
	}
	fmt.Fprintf(output, usageEpilogue)
}
