// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cachedb

import (
	"compress/gzip"
	"io"
	"os"

	"github.com/vmihailenco/msgpack/v5"
)

type format uint8

const (
	mpz1 format = iota
)

func (f format) String() string {
	switch f {
	case mpz1:
		return "mpz1"
	}
	panic("unknown format")
}

const currentformat = mpz1

func mpz1read[T any](filename string) (db []T, err error) {
	filename = fname(filename, mpz1)
	f, err := os.Open(filename)
	if err != nil {
		return db, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return db, err
	}
	dat, err := io.ReadAll(gz)
	if err != nil {
		return db, err
	}
	gz.Close()
	err = msgpack.Unmarshal(dat, &db)
	if err != nil {
		return db, err
	}
	return db, nil
}

func mpz1write[T any](filename string, db []T) (err error) {
	filename = fname(filename, mpz1)
	dat, err := msgpack.Marshal(db)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(f)
	_, err = gw.Write(dat)
	if err != nil {
		return err
	}
	err = gw.Close()
	if err != nil {
		return err
	}
	return f.Close()
}
