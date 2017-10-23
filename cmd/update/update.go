//                                                       -*- coding: utf-8 -*-
// File:    ./cmd/update.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-19 13:15:16 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Updates Symbols Extractor's project files.
//
package main

import (
	"os"

	"github.com/gofed/symbols-extractor/pkg/updater"
	_ "github.com/gofed/symbols-extractor/pkg/updater/testgen"
)

func main() {
	cmd, err := updater.GetCommandByArgv(os.Args)
	if err != nil {
		updater.PrintError(err)
		updater.PrintUsage(os.Stderr)
		os.Exit(2)
	}
	err = cmd.Run(os.Args[2:])
	if err != nil {
		updater.PrintError(err)
		os.Exit(2)
	}
}
