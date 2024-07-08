// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cachedb

import (
	"os"
	"strings"
)

func Exists(filename string) bool {
	filename = Filename(filename)

	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	// if errors.Is(err, os.ErrNotExist) {
	// 	return false
	// }
	return false
}

func Read[T any](filename string) (db []T, err error) {
	switch currentformat {
	case mpz1:
		return mpz1read[T](filename)
	}
	panic("unknown format")
}

func Write[T any](filename string, db []T) (err error) {
	switch currentformat {
	case mpz1:
		return mpz1write(filename, db)
	}
	panic("unknown format")
}

func Filename(filename string) string {
	return fname(filename, currentformat)
}

func fname(filename string, f format) string {
	ext := f.String()
	if !strings.HasSuffix(filename, ext) {
		filename += "." + ext
	}
	return filename
}
