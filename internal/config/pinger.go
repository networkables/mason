// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type (
	Pinger struct {
		Enabled         *bool
		MaxWorkers      *int
		PingCount       *int
		Timeout         *time.Duration
		CheckInterval   *time.Duration
		DefaultInterval *time.Duration
		ServerInterval  *time.Duration
	}
)

const (
	perfPingCfgKey = "performancepinger"
)

var (
	cfgPerfPingEnabledKey         = Key(perfPingCfgKey, EnabledCfgKey)
	cfgPerfPingMaxWorkersKey      = Key(perfPingCfgKey, MaxWorkersCfgKey)
	cfgPerfPingPingCountKey       = Key(perfPingCfgKey, PingCountCfgKey)
	cfgPerfPingTimeoutKey         = Key(perfPingCfgKey, TimeoutCfgKey)
	cfgPerfPingCheckIntervalKey   = Key(perfPingCfgKey, CheckIntervalCfgKey)
	cfgPerfPingDefaultIntervalKey = Key(perfPingCfgKey, DefaultIntervalCfgKey)
	cfgPerfPingServerIntervalKey  = Key(perfPingCfgKey, ServerIntervalCfgKey)
)

func cfgPerfPingSetDefaults() {
	viper.SetDefault(cfgPerfPingEnabledKey, true)
	viper.SetDefault(cfgPerfPingMaxWorkersKey, 2)
	viper.SetDefault(cfgPerfPingPingCountKey, 3)
	viper.SetDefault(cfgPerfPingTimeoutKey, 200*time.Millisecond)
	viper.SetDefault(cfgPerfPingCheckIntervalKey, 5*time.Minute)
	viper.SetDefault(cfgPerfPingDefaultIntervalKey, time.Hour)
	viper.SetDefault(cfgPerfPingServerIntervalKey, 10*time.Minute)
}

func CfgPerfPingeBuildAndBindFlags(pflags *pflag.FlagSet, cfg *Pinger) {
	pflags.BoolVar(
		cfg.Enabled,
		cfgPerfPingEnabledKey,
		*cfg.Enabled,
		"Enable performing pinging",
	)
	viper.BindPFlag(cfgPerfPingEnabledKey, pflags.Lookup(cfgPerfPingEnabledKey))

	pflags.IntVar(
		cfg.MaxWorkers,
		cfgPerfPingMaxWorkersKey,
		*cfg.MaxWorkers,
		"Max number of workers for performance pinging",
	)
	viper.BindPFlag(cfgPerfPingMaxWorkersKey, pflags.Lookup(cfgPerfPingMaxWorkersKey))

	pflags.IntVar(
		cfg.PingCount,
		cfgPerfPingPingCountKey,
		*cfg.PingCount,
		"How many pings to send to a device",
	)
	viper.BindPFlag(cfgPerfPingPingCountKey, pflags.Lookup(cfgPerfPingPingCountKey))

	pflags.DurationVar(
		cfg.Timeout,
		cfgPerfPingTimeoutKey,
		*cfg.Timeout,
		"How long to wait for a response to a ping",
	)
	viper.BindPFlag(cfgPerfPingTimeoutKey, pflags.Lookup(cfgPerfPingTimeoutKey))

	pflags.DurationVar(
		cfg.CheckInterval,
		cfgPerfPingCheckIntervalKey,
		*cfg.CheckInterval,
		"How often to check if any devices are ready for pinging",
	)
	viper.BindPFlag(cfgPerfPingCheckIntervalKey, pflags.Lookup(cfgPerfPingCheckIntervalKey))

	pflags.DurationVar(
		cfg.DefaultInterval,
		cfgPerfPingDefaultIntervalKey,
		*cfg.DefaultInterval,
		"Time between a ping sesssion for a regular device",
	)
	viper.BindPFlag(cfgPerfPingDefaultIntervalKey, pflags.Lookup(cfgPerfPingDefaultIntervalKey))

	pflags.DurationVar(
		cfg.ServerInterval,
		cfgPerfPingServerIntervalKey,
		*cfg.ServerInterval,
		"Time between a ping sesssion for a server device",
	)
	viper.BindPFlag(cfgPerfPingServerIntervalKey, pflags.Lookup(cfgPerfPingServerIntervalKey))
}
