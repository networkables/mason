// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Server struct {
	TUIPort   *string
	WebPort   *string
	IgnoreCap bool
}

const (
	serverCfgKey    = "server"
	tuiPortCfgKey   = "tuiport"
	webPortCfgKey   = "webport"
	ignoreCapCfgKey = "ignorecap"
)

var (
	cfgServerTuiPortKey   = Key(serverCfgKey, tuiPortCfgKey)
	cfgServerWebPortKey   = Key(serverCfgKey, webPortCfgKey)
	cfgServerIgnoreCapKey = Key(serverCfgKey, ignoreCapCfgKey)
)

func cfgServerSetDefaults() {
	viper.SetDefault(cfgServerTuiPortKey, "4322")
	viper.SetDefault(cfgServerWebPortKey, "4380")
	viper.SetDefault(cfgServerIgnoreCapKey, false)
}

func CfgServerBuildAndBindFlags(pflags *pflag.FlagSet, cfg *Server) {
	pflags.StringVar(
		cfg.TUIPort,
		cfgServerTuiPortKey,
		*cfg.TUIPort,
		"Port the tui listens on",
	)
	viper.BindPFlag(cfgServerTuiPortKey, pflags.Lookup(cfgServerTuiPortKey))

	pflags.StringVar(
		cfg.WebPort,
		cfgServerWebPortKey,
		*cfg.WebPort,
		"Port the web ui listens on",
	)
	viper.BindPFlag(cfgServerWebPortKey, pflags.Lookup(cfgServerWebPortKey))

	pflags.BoolVar(
		&cfg.IgnoreCap,
		cfgServerIgnoreCapKey,
		cfg.IgnoreCap,
		"Ignore if the required system capabilities are present",
	)
	viper.BindPFlag(cfgServerIgnoreCapKey, pflags.Lookup(cfgServerIgnoreCapKey))
}
