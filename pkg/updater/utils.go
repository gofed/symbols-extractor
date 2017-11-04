//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/utils.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-23 15:57:19 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Updater's utilities.
//
package updater

import (
	"os"
	"path/filepath"
)

const (
	GOPATH_NAME = "GOPATH"
)

func CanonizePath(path string) (rpath string, err error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	gopath, ok := os.LookupEnv(GOPATH_NAME)
	if ok && gopath != "" {
		err = NeedsDir(gopath)
		if err != nil {
			return gopath, err
		}
		return filepath.Join(gopath, path), nil
	}
	// GOPATH is not set -> use cwd:
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, path), nil
}

func NeedsDir(path string) error {
	finfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if finfo.IsDir() {
		return nil
	}
	return NewUpdaterError("%q is not a directory!", path)
}

func NeedsFile(path string) error {
	finfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if finfo.IsDir() {
		return NewUpdaterError("%q is not a file!", path)
	}
	return nil
}

func NeedsYamlFile(path string, forceext bool) (fpath string, err error) {
	if forceext && filepath.Ext(path) == "" {
		path += ".yml"
	}
	err = NeedsFile(path)
	if err == nil {
		return path, nil
	}
	if filepath.Ext(path) != "" {
		return path, err
	}
	fpath = path + ".yml"
	err = NeedsFile(fpath)
	if err != nil {
		return fpath, err
	}
	return fpath, nil
}

func NeedsDirOrYamlFile(path string, forceext bool) (fpath string, isdir bool, err error) {
	err = NeedsDir(path)
	if err == nil {
		return path, true, nil
	}
	fpath, err = NeedsYamlFile(path, forceext)
	return fpath, false, err
}
