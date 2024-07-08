// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package netflows

import (
	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type Config struct {
	Enabled       bool
	ListenAddress string
	MaxWorkers    int
	PacketSize    int
}

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	configMajorKey := "netflows"

	flagset.Bool(
		fs,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		true,
		"enable listening and processing of netflow (ipfix) data",
	)
	flagset.String(
		fs,
		&cfg.ListenAddress,
		configMajorKey,
		"listenaddress",
		":2055",
		"address to listen for netflow data",
	)
	flagset.Int(
		fs,
		&cfg.MaxWorkers,
		configMajorKey,
		"maxworkers",
		1,
		"number of workers to process netflow packets",
	)
	flagset.Int(
		fs,
		&cfg.PacketSize,
		configMajorKey,
		"packetsize",
		16384,
		"max size of packet buffer when listening (this is per packet)",
	)
}
