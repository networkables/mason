// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oui

import (
	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type Config struct {
	Enabled   bool
	Url       string
	Directory string
	Filename  string
}

const (
	defaultUrl      = "https://standards-oui.ieee.org/oui/oui.txt"
	defaultFilename = "oui.mpz1"
	newDBSize       = 38_000
)

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	configMajorKey := "oui"

	flagset.Bool(
		fs,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		false,
		"enable usage of ieee oui database",
	)
	flagset.String(
		fs,
		&cfg.Url,
		configMajorKey,
		"url",
		defaultUrl,
		"url to fetch oui listing if local db is not found",
	)
	flagset.String(
		fs,
		&cfg.Directory,
		configMajorKey,
		"directory",
		"data/oui",
		"directory to store local db",
	)
	flagset.String(
		fs,
		&cfg.Filename,
		configMajorKey,
		"filename",
		defaultFilename,
		"filename to store local db",
	)
}
