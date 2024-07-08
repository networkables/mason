// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oui

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type store struct {
	initialized bool
	filename    string
	url         string
	db          []Entry
}

var (
	once      sync.Once
	singleton *store
)

func getstore() *store {
	once.Do(func() {
		singleton = &store{filename: defaultFilename, db: make([]Entry, 0)}
		// load(singleton)
	})
	return singleton
}

func Load(opts ...Option) {
	s := getstore()
	load(s, opts...)
}

func load(s *store, opts ...Option) {
	var err error
	popts := applyOptionsToDefault(opts...)
	ensureDirectory(popts.directory)
	datafile := filepath.Join(popts.directory, popts.filename)
	s.filename = datafile
	s.url = popts.url

	s.initialized, s.db, err = getdb(s.url, s.filename)
	if err != nil {
		log.Fatal("oui load: ", err)
	}
}

func Lookup(mac net.HardwareAddr) (name string) {
	if len(mac) < 3 {
		return ""
	}
	s := getstore()
	if !s.initialized {
		return name
	}
	pfx := prefix(mac)

	idx, found := slices.BinarySearchFunc(
		s.db,
		pfx,
		func(e Entry, t string) int {
			return strings.Compare(e.Prefix, pfx)
		},
	)
	if found {
		name = s.db[idx].Name
	}
	return name
}

func prefix(mac net.HardwareAddr) string {
	return fmt.Sprintf("%02X%02X%02X", mac[0], mac[1], mac[2])
}

func ensureDirectory(dir string) {
	if dir == "" {
		return
	}
	stat, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	if stat.IsDir() {
		return
	}
	log.Fatal("not a directory", "dir", dir)
}
