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
	name := GetCommandNameFromArgv(argv)
	if len(argv) < 2 {
		return nil, CommandError(name, ErrNArgs)
	}
	cmd = GetCommand(argv[1])
	if cmd == nil {
		return nil, CommandError(name, ErrCmdNotFound, argv[1])
	}
	return cmd, nil
}

func GetCommandNameFromArgv(argv []string) string {
	if len(argv) < 1 {
		return ""
	}
	return argv[0]
}

func CommandError(name, format string, args ...interface{}) error {
	var msg string
	sargs := fmt.Sprintf(format, args...)

	if name == "" {
		msg = fmt.Sprintf("%s\n", sargs)
	} else {
		msg = fmt.Sprintf("%s: %s\n", name, sargs)
	}
	return NewUpdaterError(msg)
}

const (
	banner = "%s - updates Symbols Extractor sources.\n"
	usagePrologue = "Usage: %s <command> <flags>\n" +
		"where <command> is one of:\n\n"
	usageEpilogue = "\nFor the detailed informations about <command>," +
		" type `<command> --help`.\n"
)

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
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
