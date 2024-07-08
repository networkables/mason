// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pinger

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type Config struct {
	Enabled         bool
	Privileged      bool
	MaxWorkers      int
	PingCount       int
	Timeout         time.Duration
	CheckInterval   time.Duration
	DefaultInterval time.Duration
	ServerInterval  time.Duration
}

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	configMajorKey := "pinger"

	flagset.Bool(
		fs,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		true,
		"enable regular ping of devices as a performance measurement",
	)
	flagset.Bool(
		fs,
		&cfg.Privileged,
		configMajorKey,
		"privileged",
		false,
		"use privileged ping",
	)
	flagset.Int(
		fs,
		&cfg.MaxWorkers,
		configMajorKey,
		"maxworkers",
		2,
		"max number of devices to ping simultaneously",
	)
	flagset.Int(
		fs,
		&cfg.PingCount,
		configMajorKey,
		"pingcount",
		3,
		"number of pings per device",
	)
	flagset.Duration(
		fs,
		&cfg.Timeout,
		configMajorKey,
		"timeout",
		100*time.Millisecond,
		"max time to wait for response",
	)
	flagset.Duration(
		fs,
		&cfg.CheckInterval,
		configMajorKey,
		"checkinterval",
		5*time.Minute,
		"interval to check if any devices are ready retest",
	)
	flagset.Duration(
		fs,
		&cfg.DefaultInterval,
		configMajorKey,
		"defaultinterval",
		time.Hour,
		"time between pings for non-server devices",
	)
	flagset.Duration(
		fs,
		&cfg.ServerInterval,
		configMajorKey,
		"serverinterval",
		5*time.Minute,
		"time between pings for server devices",
	)
}
