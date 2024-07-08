// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type Config struct {
	Enabled               bool
	Directory             string
	URL                   string
	Filename              string
	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
	ConnectionMaxIdle     time.Duration
}

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	configMajorKey := "store.sqlite"

	flagset.Bool(
		fs,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		true,
		"enable use of sqlite data store",
	)
	flagset.String(
		fs,
		&cfg.Directory,
		configMajorKey,
		"directory",
		"data/db",
		"directory to store files",
	)
	flagset.String(
		fs,
		&cfg.URL,
		configMajorKey,
		"url",
		"",
		"additional sqlite url parameters",
	)
	flagset.String(
		fs,
		&cfg.Filename,
		configMajorKey,
		"filename",
		"mason.db",
		"database filename",
	)
	flagset.Int(
		fs,
		&cfg.MaxOpenConnections,
		configMajorKey,
		"maxopenconnections",
		5,
		"max open connections",
	)
	flagset.Int(
		fs,
		&cfg.MaxIdleConnections,
		configMajorKey,
		"maxidleconnections",
		5,
		"max idle connections",
	)
	flagset.Duration(
		fs,
		&cfg.ConnectionMaxLifetime,
		configMajorKey,
		"connectionmaxlifetime",
		time.Hour,
		"max time a connection can be open",
	)
	flagset.Duration(
		fs,
		&cfg.ConnectionMaxIdle,
		configMajorKey,
		"connectionmaxidle",
		time.Hour,
		"max time a connection can be idle",
	)
}
