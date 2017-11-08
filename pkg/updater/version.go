//                                                       -*- coding: utf-8 -*-
// File:    ./pkg/updater/version.go
// Author:  Jiří Kučera, <jkucera AT redhat DOT com>
// Stamp:   2017-10-25 17:54:04 (UTC+0100, DST+0100)
// Project: Symbols Extractor
// Version: See VERSION.
// License: See LICENSE.
// Brief:   Data type holding version.
//
package updater

import (
	"fmt"
	"strings"
	"strconv"
)

const (
	ErrVersionFormat = "version number should contain at most 3 parts!"
)

type Version struct {
	MajorLevel, MinorLevel, PatchLevel uint
}

func (v *Version) String() string {
	return fmt.Sprintf(
		"%v.%v.%v", v.MajorLevel, v.MinorLevel, v.PatchLevel,
	)
}

func Compare(v1, v2 *Version) int {
	r := int(v1.MajorLevel) - int(v2.MajorLevel)
	if r != 0 {
		return r
	}
	r = int(v1.MinorLevel) - int(v2.MinorLevel)
	if r != 0 {
		return r
	}
	return int(v1.PatchLevel) - int(v2.PatchLevel)
}

func ToVersion(s string) (v *Version, err error) {
	var parts [3]uint
	sparts := strings.Split(s, ".")
	i := 0
	for _, v := range(sparts) {
		if i >= 3 {
			return nil, NewUpdaterError(ErrVersionFormat)
		}
		tv := strings.TrimSpace(v)
		n, err := strconv.ParseUint(tv, 10, 32)
		if err != nil {
			return nil, err
		}
		parts[i] = uint(n)
		i++
	}
	return &Version{parts[0], parts[1], parts[2]}, nil
}
