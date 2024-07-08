// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package combostore

import (
	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type Config struct {
	Enabled      bool
	Directory    string
	WSPRetention string
}

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	configMajorKey := "store.combo"

	flagset.Bool(
		fs,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		false,
		"enable use of combo store (msgpack/whisper data store)",
	)
	flagset.String(
		fs,
		&cfg.Directory,
		configMajorKey,
		"directory",
		"data",
		"directory to store files",
	)
	flagset.String(
		fs,
		&cfg.WSPRetention,
		configMajorKey,
		"wspretention",
		"10m:3d,1h:3w",
		"whisper retention settings",
	)
}
