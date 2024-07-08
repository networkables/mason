// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	"github.com/maragudk/migrate"
	_ "github.com/mattn/go-sqlite3"

	"github.com/networkables/mason/internal/model"
)

type Store struct {
	DB                    *sqlx.DB
	url                   string
	maxOpenConnections    int
	maxIdleConnections    int
	connectionMaxLifetime time.Duration
	connectionMaxIdleTime time.Duration
	log                   *log.Logger
	// --- old elow

	directory string
	filename  string
	networks  []model.Network
	devices   []model.Device
}

func newSqliteDatabase(cfg *Config) *Store {
	var url string

	ensureDirectory(cfg.Directory)
	if cfg.Filename != "" {
		url = "file:"
		if cfg.Directory != "" {
			url += cfg.Directory + "/"
		}
		url += cfg.Filename
	}

	// if opts.Log == nil {
	// 	opts.Log = log.New(io.Discard, "", 0)
	// }

	// - Set WAL mode (not strictly necessary each time because it's persisted in the database, but good for first run)
	// - Set busy timeout, so concurrent writers wait on each other instead of erroring immediately
	// - Enable foreign key checks
	url += cfg.URL + "?_journal=WAL&_timeout=5000&_fk=true"

	cs := &Store{
		url:                   url,
		filename:              cfg.Filename,
		maxOpenConnections:    cfg.MaxOpenConnections,
		maxIdleConnections:    cfg.MaxIdleConnections,
		connectionMaxLifetime: cfg.ConnectionMaxLifetime,
		connectionMaxIdleTime: cfg.ConnectionMaxIdle,
	}
	return cs
}

func (d *Store) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	d.DB, err = sqlx.ConnectContext(ctx, "sqlite3", d.url)
	if err != nil {
		return err
	}

	d.DB.SetMaxOpenConns(d.maxOpenConnections)
	d.DB.SetMaxIdleConns(d.maxIdleConnections)
	d.DB.SetConnMaxLifetime(d.connectionMaxLifetime)
	d.DB.SetConnMaxIdleTime(d.connectionMaxIdleTime)

	return nil
}

func New(cfg *Config) (*Store, error) {
	ctx := context.TODO()

	cs := newSqliteDatabase(cfg)

	err := cs.connect()
	if err != nil {
		return nil, err
	}

	err = cs.readNetworksInitial(ctx)
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
	return cs.DB.Close()
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

//go:embed migrations
var migrations embed.FS

func (cs *Store) MigrateUp(ctx context.Context) error {
	fsys := cs.getMigrations()
	return migrate.Up(ctx, cs.DB.DB, fsys)
}

func (cs *Store) MigrateDown(ctx context.Context) error {
	fsys := cs.getMigrations()
	return migrate.Down(ctx, cs.DB.DB, fsys)
}

func (cs *Store) getMigrations() fs.FS {
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		panic(err)
	}
	return fsys
}
