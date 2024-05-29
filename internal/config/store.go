// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfgStoreDirKey          = Key(storeCfgKey, dirCfgKey)
	cfgStoreWspRetentionKey = Key(storeCfgKey, wspRetentionCfgKey)
)

func cfgStoreSetDefaults() {
	viper.SetDefault(cfgStoreDirKey, "data/store")
	viper.SetDefault(cfgStoreWspRetentionKey, "10m:3d,1h:3w")
}

type (
	// StoresConfig holds configuration needed to configure the data stores used in Mason
	Store struct {
		Directory    *string
		WSPRetention *string
	}
)

const (
	storeCfgKey        = "store"
	dirCfgKey          = "directory"
	wspRetentionCfgKey = "wspretention"
)

func CfgStoreBuildAndBindFlags(flags *pflag.FlagSet, cfg *Store) {
	flags.StringVar(
		cfg.Directory,
		cfgStoreDirKey,
		*cfg.Directory,
		"Directory to save data",
	)
	viper.BindPFlag(cfgStoreDirKey, flags.Lookup(cfgStoreDirKey))

	flags.StringVar(
		cfg.WSPRetention,
		cfgStoreWspRetentionKey,
		*cfg.WSPRetention,
		"Whisper retention settings for timeseries data",
	)
	viper.BindPFlag(cfgStoreWspRetentionKey, flags.Lookup(cfgStoreWspRetentionKey))
}
