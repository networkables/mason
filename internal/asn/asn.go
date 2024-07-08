// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package asn

import (
	"errors"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type store struct {
	initialized   bool
	cachefilename string
	asnurl        string
	countryurl    string
	db            []CacheEntry
}

var (
	once      sync.Once
	singleton *store
)

func getstore() *store {
	once.Do(func() {
		singleton = &store{cachefilename: defaultCacheFilename, db: make([]CacheEntry, 0)}
		// load(singleton)
	})
	return singleton
}

func Load(opts ...Option) {
	s := getstore()
	load(s, opts...)
}

func load(s *store, opts ...Option) {
	popts := applyOptionsToDefault(opts...)
	ensureDirectory(popts.directory)
	datafile := filepath.Join(popts.directory, popts.cachefilename)
	s.cachefilename = datafile
	s.asnurl = popts.asnurl
	s.countryurl = popts.countryurl

	s.initialized, s.db = getdb(s.asnurl, s.countryurl, s.cachefilename, popts.store)
}

func FindAsn(addr netip.Addr) (asn string) {
	s := getstore()
	idx, found := slices.BinarySearchFunc(
		s.db,
		addr,
		func(e CacheEntry, ip netip.Addr) int {
			if e.Range.Contains(ip) {
				return 0
			}
			return e.Range.Prefixes()[0].Addr().Compare(ip)
		},
	)
	if found {
		return s.db[idx].Asn
	}
	return asn
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
