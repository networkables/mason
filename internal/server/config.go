// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/networkables/mason/internal/asn"
	"github.com/networkables/mason/internal/bus"
	"github.com/networkables/mason/internal/combostore"
	"github.com/networkables/mason/internal/discovery"
	"github.com/networkables/mason/internal/enrichment"
	"github.com/networkables/mason/internal/flagset"
	"github.com/networkables/mason/internal/netflows"
	"github.com/networkables/mason/internal/oui"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/internal/sqlitestore"
)

type Store struct {
	Combo  *combostore.Config
	Sqlite *sqlitestore.Config
}

type TuiConfig struct {
	Enabled       bool
	ListenAddress string
	KeyDirectory  string
}

type WuiConfig struct {
	Enabled       bool
	ListenAddress string
}

type Config struct {
	ConfigDirectory string
	Store           *Store
	Wui             *WuiConfig
	Tui             *TuiConfig
	Bus             *bus.Config
	Discovery       *discovery.Config
	Pinger          *pinger.Config
	Enrichment      *enrichment.Config
	NetFlows        *netflows.Config
	Asn             *asn.Config
	Oui             *oui.Config
}

var (
	configonce      sync.Once
	configsingleton *Config
)

func ptr[T any](x T) *T {
	return &x
}

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	flagset.String(
		fs,
		&cfg.ConfigDirectory,
		"config",
		"directory",
		"config",
		"location of config file(s)",
	)

	wuiConfigMajorKey := "wui"

	flagset.Bool(fs, &cfg.Wui.Enabled, wuiConfigMajorKey, "enabled", true, "enable the web ui")
	flagset.String(
		fs,
		&cfg.Wui.ListenAddress,
		wuiConfigMajorKey,
		"listenaddress",
		":4380",
		"address to list for http requests",
	)

	tuiConfigMajorKey := "tui"

	flagset.Bool(
		fs,
		&cfg.Tui.Enabled,
		tuiConfigMajorKey,
		"enabled",
		true,
		"enable the ssh terminal ui",
	)
	flagset.String(
		fs,
		&cfg.Tui.ListenAddress,
		tuiConfigMajorKey,
		"listenaddress",
		":4322",
		"address to list for ssh connections",
	)
	flagset.String(
		fs,
		&cfg.Tui.KeyDirectory,
		tuiConfigMajorKey,
		"keydirectory",
		"data/ssh",
		"directory to store ssh key, current directory if not specifed",
	)
}

func GetConfig() *Config {
	configonce.Do(func() {
		configsingleton = defaultConfig()
	})
	return configsingleton
}

const (
	configName = "mason"
	configType = "yaml"
)

func defaultConfig() *Config {
	c := &Config{
		Store: &Store{
			Combo:  &combostore.Config{},
			Sqlite: &sqlitestore.Config{},
		},
		Wui:        &WuiConfig{},
		Tui:        &TuiConfig{},
		Bus:        &bus.Config{},
		Discovery:  &discovery.Config{},
		Pinger:     &pinger.Config{},
		Enrichment: &enrichment.Config{},
		NetFlows:   &netflows.Config{},
		Asn:        &asn.Config{},
		Oui:        &oui.Config{},
	}

	// viper.SetConfigName(configName)
	// viper.SetConfigType(configType)
	// viper.AddConfigPath(".")    // optionally look for config in the working directory
	// err := viper.ReadInConfig() // Find and read the config file
	// if err != nil {             // Handle errors reading the config file
	// 	log.Fatal("viper read", "error", err)
	// }
	// viper.SetEnvPrefix("MASON")

	err := viper.Unmarshal(c)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func (c Config) Save() error {
	fname := fmt.Sprintf("%s.%d.%s", configName, time.Now().Unix(), configType)
	return viper.WriteConfigAs(fname)
}
