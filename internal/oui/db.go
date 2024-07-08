// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oui

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/cachedb"
)

type Entry struct {
	Prefix string
	Name   string
}

func getdb(url string, filename string) (initialized bool, db []Entry, err error) {
	if !cachedb.Exists(filename) {
		log.Info("building oui local cache (roughly 10s)")
		db, err = builddb(url)
		if err != nil {
			log.Fatal("getdb: ", err)
		}
		err = cachedb.Write(filename, db)
		if err != nil {
			return false, db, err
		}
		log.Info("finished building oui local cache", "count", len(db))
		return true, db, nil
	}
	db, err = cachedb.Read[Entry](filename)
	if err != nil {
		return initialized, db, err
	}
	log.Info("loaded oui from local", "count", len(db))
	return true, db, nil
}

func download(url string) (dat []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return dat, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func builddb(url string) (db []Entry, err error) {
	db = make([]Entry, 0, newDBSize)

	dat, err := download(url)
	if err != nil {
		return db, err
	}

	db = parsedata(dat, db)

	slices.SortFunc(db, func(a, b Entry) int {
		return strings.Compare(a.Prefix, b.Prefix)
	})

	return db, err
}

func parsedata(dat []byte, db []Entry) []Entry {
	var (
		fields []string
		prefix string
		name   string
	)

	b := bufio.NewScanner(bytes.NewBuffer(dat))
	for b.Scan() {
		line := b.Text()
		if strings.Contains(line, "(base 16)") {
			fields = strings.Split(line, "     ")
			prefix = fields[0]
			fields = strings.Split(fields[1], "\t")
			name = fields[2]
			db = append(db, Entry{Prefix: prefix, Name: name})
		}
	}
	if b.Err() != nil {
		log.Fatal("parsedata scanner: ", b.Err())
	}
	return db
}
