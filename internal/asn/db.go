// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package asn

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/charmbracelet/log"
	"go4.org/netipx"

	"github.com/networkables/mason/internal/cachedb"
	"github.com/networkables/mason/internal/model"
)

type CacheEntry struct {
	Asn   string
	Range netipx.IPRange
}

type asnEntry struct {
	Range netipx.IPRange
	Asn   string
	Name  string
}

type asnCountryEntry struct {
	Range   netipx.IPRange
	Asn     string
	Country string
	Name    string
}

type ctentry struct {
	Range   netipx.IPRange
	Country string
}

type asnstorer interface {
	StartAsnLoad() func(*error)
	UpsertAsn(context.Context, model.Asn) error
}

func fullToModelAsn(f asnCountryEntry) model.Asn {
	return model.Asn{
		IPRange: model.IPRangeToModelIPRange(f.Range),
		Asn:     f.Asn,
		Country: f.Country,
		Name:    f.Name,
	}
}

func getdb(
	asnurl string,
	countryurl string,
	cachefilename string,
	store asnstorer,
) (initialized bool, memdb []CacheEntry) {
	var err error
	if !cachedb.Exists(cachefilename) {
		log.Info("building asn local cache (roughly 60s)")
		ctx := context.Background()
		var fulldb []asnCountryEntry
		memdb, fulldb, err = builddb(asnurl, countryurl)
		if err != nil {
			log.Fatal("getdb: ", err)
		}
		err = savefulldb(ctx, store, fulldb)
		if err != nil {
			log.Fatal("store asn: ", err)
		}
		err = cachedb.Write(cachefilename, memdb)
		log.Info("finished building asn local cache", "count", len(memdb))
		return true, memdb
	}
	memdb, err = cachedb.Read[CacheEntry](cachefilename)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("loaded asn from local cache", "count", len(memdb))
	return true, memdb
}

func savefulldb(ctx context.Context, store asnstorer, db []asnCountryEntry) (err error) {
	if store == nil {
		return errors.New("asnstorer is nil")
	}

	fn := store.StartAsnLoad()
	defer func() {
		fn(&err)
	}()
	for _, entry := range db {
		err = store.UpsertAsn(ctx, fullToModelAsn(entry))
		if err != nil {
			return err
		}
	}
	return nil
}

func builddb(
	asnurl string,
	countryurl string,
) (memdb []CacheEntry, fulldb []asnCountryEntry, err error) {
	asndat, err := download(asnurl)
	if err != nil {
		return memdb, fulldb, err
	}
	var asndb []asnEntry
	memdb, asndb = buildAsnDBs(asndat)

	countrydat, err := download(countryurl)
	if err != nil {
		return memdb, fulldb, err
	}
	ctdb := buildCtDB(countrydat)

	fulldb = buildFullDB(asndb, ctdb)

	return memdb, fulldb, err
}

func download(url string) (dat []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return dat, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func buildCtDB(raw []byte) (db []ctentry) {
	ctbuff := bytes.NewBuffer(raw)
	ctcsvr := csv.NewReader(ctbuff)
	ctrecs, err := ctcsvr.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	db = make([]ctentry, len(ctrecs))
	for idx, rec := range ctrecs {
		rng := netipx.MustParseIPRange(rec[0] + "-" + rec[1])
		db[idx] = ctentry{
			Range:   rng,
			Country: rec[2],
		}
	}
	slices.SortFunc(db, func(a, b ctentry) int {
		return a.Range.From().Compare(b.Range.From())
	})
	return db
}

func buildAsnDBs(raw []byte) (cachedb []CacheEntry, asndb []asnEntry) {
	asnbuff := bytes.NewBuffer(raw)
	asncsvr := csv.NewReader(asnbuff)
	asnrecs, err := asncsvr.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	asndb = make([]asnEntry, len(asnrecs))
	cachedb = make([]CacheEntry, len(asnrecs))
	for idx, rec := range asnrecs {
		asn := rec[2]
		name := rec[3]
		rng := netipx.MustParseIPRange(rec[0] + "-" + rec[1])
		asndb[idx] = asnEntry{
			Asn:   asn,
			Name:  name,
			Range: rng,
		}
		cachedb[idx] = CacheEntry{
			Range: rng,
			Asn:   asn,
		}
	}
	slices.SortFunc(asndb, func(a, b asnEntry) int {
		return a.Range.From().Compare(b.Range.From())
	})
	slices.SortFunc(cachedb, func(a, b CacheEntry) int {
		return a.Range.From().Compare(b.Range.From())
	})

	return cachedb, asndb
}

func buildFullDB(asndb []asnEntry, ctdb []ctentry) (full []asnCountryEntry) {
	var (
		ctidx int
		found bool
	)
	full = make([]asnCountryEntry, len(asndb))

	for idx, rec := range asndb {
		ctidx, found = slices.BinarySearchFunc(
			ctdb,
			rec.Range,
			func(e ctentry, t netipx.IPRange) int {
				from := (t.Prefixes()[0]).Addr()
				if e.Range.Contains(from) {
					return 0
				}
				if e.Range.From().Compare(from) == -1 {
					return -1
				}
				if e.Range.To().Compare(from) == 1 {
					return 1
				}
				fmt.Print("should not get here\n")
				return 0
			},
		)
		full[idx] = asnCountryEntry{
			Asn:   rec.Asn,
			Name:  rec.Name,
			Range: rec.Range,
		}
		if found {
			full[idx].Country = ctdb[ctidx].Country
		}
	}
	slices.SortFunc(full, func(a, b asnCountryEntry) int {
		return a.Range.From().Compare(b.Range.From())
	})

	return full
}
