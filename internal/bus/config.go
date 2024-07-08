// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bus

import (
	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type Config struct {
	MaxEvents            int
	MaxErrors            int
	InboundSize          int
	MinimumPriorityLevel int
	EnableDebugLog       bool
	EnableErrorLog       bool
}

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	configMajorKey := "bus"

	flagset.Int(
		fs,
		&cfg.MaxEvents,
		configMajorKey,
		"maxevents",
		100,
		"max number of events to retain",
	)
	flagset.Int(
		fs,
		&cfg.MaxErrors,
		configMajorKey,
		"maxerrors",
		100,
		"max number of errors to retain",
	)
	flagset.Int(fs, &cfg.InboundSize, configMajorKey, "inboundsize", 0, "inbound channel size")
	flagset.Int(
		fs,
		&cfg.MinimumPriorityLevel,
		configMajorKey,
		"minimumprioritylevel",
		20,
		"minimum event priority level to save in history",
	)
	flagset.Bool(
		fs,
		&cfg.EnableErrorLog,
		configMajorKey,
		"enableerrorlog",
		true,
		"output bus errors to console",
	)
	flagset.Bool(
		fs,
		&cfg.EnableDebugLog,
		configMajorKey,
		"enabledebuglog",
		true,
		"output debug bus messages to console",
	)
}
