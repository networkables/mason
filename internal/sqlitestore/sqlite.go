// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"errors"
	"os"

	"github.com/charmbracelet/log"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/networkables/mason/internal/model"
)

type Store struct {
	DB   *sqlite.Conn
	Pool *sqlitemigration.Pool
	url  string

	directory string
	filename  string
	networks  []model.Network
	devices   []model.Device
}

func newSqliteDatabase(cfg *Config) *Store {
	schema := sqlitemigration.Schema{
		Migrations: []string{
			`create table devices (
  addr text primary key,
  name text,
  mac text,
  discoveredat timestamp,
  discoveredby text,
  -- Meta
  metadnsname text,
  metamanufacturer text,
  metatags text,
  -- Server
  serverports text,
  serverlastscan timestamp,
  -- PerfPing
  perfpingfirstseen timestamp,
  perfpinglastseen timestamp,
  perfpingmeanping integer,
  perfpingmaxping integer,
  perfpinglastfailed integer,
  -- Snmp
  snmpname text,
  snmpdescription text,
  snmpcommunity text,
  snmpport integer,
  snmplastcheck timestamp,
  snmphasarptable text,
  snmplastarptablescan timestamp,
  snmphasinterfaces text,
  snmplastinterfacesscan timestamp
);`,

			`create table networks (
  prefix text primary key,
  name string,
  lastscan timestamp,
  tags text
);`,

			`create table flows (
  start timestamp,
  end timestamp,
  srcaddr text,
  srcport integer,
  srcasn text,
  dstaddr text,
  dstport integer,
  dstasn text,
  protocol text,
  bytes integer,
  packets integer
);`,

			`create table performancepings (
  start timestamp,
  addr text,
  minimum integer,
  average integer,
  maximum integer,
  loss float
);`,

			`create table asns (
  asn text primary key,
  country text,
  name text,
  iprange text,
  created timestamp
);`,
		},
	}

	var url string

	ensureDirectory(cfg.Directory)
	if cfg.Filename != "" {
		// url = "file:"
		if cfg.Directory != "" {
			url += cfg.Directory + "/"
		}
		url += cfg.Filename
	}
	url += cfg.URL

	pool := sqlitemigration.NewPool(url, schema, sqlitemigration.Options{
		Flags:    sqlite.OpenCreate | sqlite.OpenReadWrite | sqlite.OpenWAL,
		PoolSize: cfg.MaxOpenConnections,
		PrepareConn: func(conn *sqlite.Conn) error {
			return sqlitex.ExecuteTransient(conn, "PRAGMA foreign_keys = ON;", nil)
		},
		OnError: func(err error) {
			log.Error("sqlitepool", "error", err)
		},
	})

	// TODO: Need a better solution for using the pool
	conn, err := pool.Get(context.TODO())
	if err != nil {
		log.Fatal("pool get conn", "error", err)
	}

	cs := &Store{
		url:      url,
		filename: cfg.Filename,
		Pool:     pool,
		DB:       conn,
	}
	return cs
}

func New(cfg *Config) (*Store, error) {
	ctx := context.TODO()

	cs := newSqliteDatabase(cfg)

	err := cs.readNetworksInitial(ctx)
	if err != nil {
		return nil, err
	}
	err = cs.readDevicesInitial(ctx)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (cs *Store) Close() error {
	cs.Pool.Put(cs.DB)
	return cs.Pool.Close()
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
