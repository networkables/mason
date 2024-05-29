// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Bus struct {
	MaxEvents               *int
	MaxErrors               *int
	InboundSize             *int
	EnableDebugLog          *bool
	EnableErrorLog          *bool
	CheckErrorTypeOnPublish *bool
}

const (
	busCfgKey            = "bus"
	maxEventsCfgKey      = "maxevents"
	maxErrorsCfgKey      = "maxerrors"
	inboundSizeCfgKey    = "inboundsize"
	enableErrorLogCfgKey = "enableerrorlog"
)

var (
	cfgBusMaxEventsKey      = Key(busCfgKey, maxEventsCfgKey)
	cfgBusMaxErrorsKey      = Key(busCfgKey, maxErrorsCfgKey)
	cfgBusEnableDebugLogKey = Key(busCfgKey, EnableDebugLogCfgKey)
	cfgBusEnableErrorLogKey = Key(busCfgKey, enableErrorLogCfgKey)
	cfgBusInboundSizeKey    = Key(busCfgKey, inboundSizeCfgKey)
)

func cfgBusSetDefaults() {
	viper.SetDefault(cfgBusMaxEventsKey, 100)
	viper.SetDefault(cfgBusMaxErrorsKey, 100)
	viper.SetDefault(cfgBusEnableDebugLogKey, false)
	viper.SetDefault(cfgBusEnableErrorLogKey, true)
	viper.SetDefault(cfgBusInboundSizeKey, 200)
}

func CfgBusBuildAndBindFlags(pflags *pflag.FlagSet, cfg *Bus) {
	pflags.IntVar(
		cfg.MaxEvents,
		cfgBusMaxEventsKey,
		*cfg.MaxEvents,
		"Max number of events to retain",
	)
	viper.BindPFlag(cfgBusMaxEventsKey, pflags.Lookup(cfgBusMaxEventsKey))

	pflags.IntVar(
		cfg.MaxErrors,
		cfgBusMaxErrorsKey,
		*cfg.MaxErrors,
		"Max number of errors to retain",
	)
	viper.BindPFlag(cfgBusMaxErrorsKey, pflags.Lookup(cfgBusMaxErrorsKey))

	pflags.BoolVar(
		cfg.EnableDebugLog,
		cfgBusEnableDebugLogKey,
		*cfg.EnableDebugLog,
		"Enable debug logging",
	)
	viper.BindPFlag(cfgBusEnableDebugLogKey, pflags.Lookup(cfgBusEnableDebugLogKey))

	pflags.BoolVar(
		cfg.EnableErrorLog,
		cfgBusEnableErrorLogKey,
		*cfg.EnableErrorLog,
		"Enable error logging",
	)
	viper.BindPFlag(cfgBusEnableErrorLogKey, pflags.Lookup(cfgBusEnableErrorLogKey))

	pflags.IntVar(
		cfg.InboundSize,
		cfgBusInboundSizeKey,
		*cfg.InboundSize,
		"Size of the inbound events buffer",
	)
	viper.BindPFlag(cfgBusInboundSizeKey, pflags.Lookup(cfgBusInboundSizeKey))
}
