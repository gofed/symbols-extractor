//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/errors.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-19 19:24:52 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Updater's command interface.
//
package updater

const (
	ErrNArgs = "Insuficient number of arguments!"
	ErrCmdNotFound = "%s is not a valid command name!"
)

type UpdaterError struct {
	detail string
}

func (e *UpdaterError) Error() string {
	return e.detail
}

func NewUpdaterError(detail string) error {
	return &UpdaterError{detail}
}
