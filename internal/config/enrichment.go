// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	enrichmentCfgKey  = "enrichment"
	portScanCfgKey    = "portscanner"
	dnsLookupCfgKey   = "dnslookup"
	ouiLookupCfgKey   = "ouilookup"
	portListCfgKey    = "portlist"
	portTimeoutCfgKey = "porttimeout"
	SnmpCfgKey        = "snmp"
)

type (
	Enrichment struct {
		Enabled     *bool
		MaxWorkers  *int
		DnsLookup   *DnsLookupConfig
		OuiLookup   *OuiLookupConfig
		PortScanner *PortScannerConfig
		SNMP        *SNMPEnrichmentConfig
	}

	DnsLookupConfig struct {
		Enabled *bool
	}

	OuiLookupConfig struct {
		Enabled *bool
	}

	PortScannerConfig struct {
		Enabled         *bool
		PortTimeout     *time.Duration
		MaxWorkers      *int
		DefaultInterval *time.Duration
		ServerInterval  *time.Duration
		PortList        *string
	}

	SNMPEnrichmentConfig struct {
		Enabled   *bool
		Timeout   *time.Duration
		Community []string
		Ports     []int
	}
)

var (
	cfgEnrichEnabledKey    = Key(enrichmentCfgKey, EnabledCfgKey)
	cfgEnrichMaxWorkersKey = Key(enrichmentCfgKey, MaxWorkersCfgKey)
	cfgEnrichDnsEnabledKey = Key(
		enrichmentCfgKey,
		dnsLookupCfgKey,
		EnabledCfgKey,
	)
	cfgEnrichOuiEnabledKey = Key(
		enrichmentCfgKey,
		ouiLookupCfgKey,
		EnabledCfgKey,
	)
	cfgEnrichPortScanEnabledKey = Key(
		enrichmentCfgKey,
		portScanCfgKey,
		EnabledCfgKey,
	)
	cfgEnrichPortScanPortTimeoutKey = Key(
		enrichmentCfgKey,
		portScanCfgKey,
		portTimeoutCfgKey,
	)
	cfgEnrichPortScanMaxWorkersKey = Key(
		enrichmentCfgKey,
		portScanCfgKey,
		MaxWorkersCfgKey,
	)
	cfgEnrichPortScanPortDefaultIntervalKey = Key(
		enrichmentCfgKey,
		portScanCfgKey,
		DefaultIntervalCfgKey,
	)
	cfgEnrichPortScanPortServerIntervalKey = Key(
		enrichmentCfgKey,
		portScanCfgKey,
		ServerIntervalCfgKey,
	)
	cfgEnrichPortScanPortPortListKey = Key(enrichmentCfgKey, portScanCfgKey, portListCfgKey)
	cfgEnrichSnmpEnabledKey          = Key(
		enrichmentCfgKey,
		SnmpCfgKey,
		EnabledCfgKey,
	)
	cfgEnrichSnmpTimeoutKey = Key(
		enrichmentCfgKey,
		SnmpCfgKey,
		TimeoutCfgKey,
	)
	cfgEnrichSnmpCommunityKey = Key(
		enrichmentCfgKey,
		SnmpCfgKey,
		CommunityCfgKey,
	)
	cfgEnrichSnmpPortsKey = Key(
		enrichmentCfgKey,
		SnmpCfgKey,
		PortsCfgKey,
	)
)

func cfgEnrichSetDefaults() {
	viper.SetDefault(cfgEnrichEnabledKey, true)
	viper.SetDefault(cfgEnrichMaxWorkersKey, 2)
	viper.SetDefault(cfgEnrichDnsEnabledKey, true)
	viper.SetDefault(cfgEnrichOuiEnabledKey, true)
	viper.SetDefault(cfgEnrichPortScanEnabledKey, true)
	viper.SetDefault(cfgEnrichPortScanPortTimeoutKey, 20*time.Millisecond)
	viper.SetDefault(cfgEnrichPortScanMaxWorkersKey, 2)
	viper.SetDefault(cfgEnrichPortScanPortDefaultIntervalKey, 7*24*time.Hour)
	viper.SetDefault(cfgEnrichPortScanPortServerIntervalKey, 24*time.Hour)
	viper.SetDefault(cfgEnrichPortScanPortPortListKey, "general")
	viper.SetDefault(cfgEnrichSnmpEnabledKey, true)
	viper.SetDefault(cfgEnrichSnmpTimeoutKey, 50*time.Millisecond)
	viper.SetDefault(cfgEnrichSnmpCommunityKey, []string{"public"})
	viper.SetDefault(cfgEnrichSnmpPortsKey, []int{161})
}

func CfgEnrichBuildAndBindFlags(pflags *pflag.FlagSet, cfg *Enrichment) {
	pflags.BoolVar(
		cfg.Enabled,
		cfgEnrichEnabledKey,
		*cfg.Enabled,
		"Enable device enrichment after discovery",
	)
	viper.BindPFlag(cfgEnrichEnabledKey, pflags.Lookup(cfgEnrichEnabledKey))

	pflags.IntVar(
		cfg.MaxWorkers,
		cfgEnrichMaxWorkersKey,
		*cfg.MaxWorkers,
		"Max number of workers running for device enrichment",
	)
	viper.BindPFlag(cfgEnrichMaxWorkersKey, pflags.Lookup(cfgEnrichMaxWorkersKey))

	pflags.BoolVar(
		cfg.DnsLookup.Enabled,
		cfgEnrichDnsEnabledKey,
		*cfg.DnsLookup.Enabled,
		"Enable dns resolution of device address",
	)
	viper.BindPFlag(cfgEnrichDnsEnabledKey, pflags.Lookup(cfgEnrichDnsEnabledKey))

	pflags.BoolVar(
		cfg.OuiLookup.Enabled,
		cfgEnrichOuiEnabledKey,
		*cfg.OuiLookup.Enabled,
		"Enable manufacturer lookup using device MAC",
	)
	viper.BindPFlag(cfgEnrichOuiEnabledKey, pflags.Lookup(cfgEnrichOuiEnabledKey))

	pflags.BoolVar(
		cfg.PortScanner.Enabled,
		cfgEnrichPortScanEnabledKey,
		*cfg.PortScanner.Enabled,
		"Enable open port scanning of device",
	)
	viper.BindPFlag(cfgEnrichPortScanEnabledKey, pflags.Lookup(cfgEnrichPortScanEnabledKey))

	pflags.DurationVar(
		cfg.PortScanner.PortTimeout,
		cfgEnrichPortScanPortTimeoutKey,
		*cfg.PortScanner.PortTimeout,
		"Maximum duration to wait for a port open response",
	)
	viper.BindPFlag(cfgEnrichPortScanPortTimeoutKey, pflags.Lookup(cfgEnrichPortScanPortTimeoutKey))

	pflags.IntVar(
		cfg.PortScanner.MaxWorkers,
		cfgEnrichPortScanMaxWorkersKey,
		*cfg.PortScanner.MaxWorkers,
		"Max number of worker for port scanning",
	)
	viper.BindPFlag(cfgEnrichPortScanMaxWorkersKey, pflags.Lookup(cfgEnrichPortScanMaxWorkersKey))

	pflags.DurationVar(
		cfg.PortScanner.DefaultInterval,
		cfgEnrichPortScanPortDefaultIntervalKey,
		*cfg.PortScanner.DefaultInterval,
		"Duration between open port checks for regular devices",
	)
	viper.BindPFlag(
		cfgEnrichPortScanPortDefaultIntervalKey,
		pflags.Lookup(cfgEnrichPortScanPortDefaultIntervalKey),
	)

	pflags.DurationVar(
		cfg.PortScanner.ServerInterval,
		cfgEnrichPortScanPortServerIntervalKey,
		*cfg.PortScanner.ServerInterval,
		"Duration between open port checks for server devices",
	)
	viper.BindPFlag(
		cfgEnrichPortScanPortServerIntervalKey,
		pflags.Lookup(cfgEnrichPortScanPortServerIntervalKey),
	)

	pflags.StringVar(
		cfg.PortScanner.PortList,
		cfgEnrichPortScanPortPortListKey,
		*cfg.PortScanner.PortList,
		"Which port list to use during port scanning",
	)
	viper.BindPFlag(
		cfgEnrichPortScanPortPortListKey,
		pflags.Lookup(cfgEnrichPortScanPortPortListKey),
	)

	pflags.BoolVar(
		cfg.SNMP.Enabled,
		cfgEnrichSnmpEnabledKey,
		*cfg.SNMP.Enabled,
		"Enable snmp probing of device",
	)
	viper.BindPFlag(cfgEnrichSnmpEnabledKey, pflags.Lookup(cfgEnrichSnmpEnabledKey))

	pflags.DurationVar(
		cfg.SNMP.Timeout,
		cfgEnrichSnmpTimeoutKey,
		*cfg.SNMP.Timeout,
		"Duration to wait for a snmp response",
	)
	viper.BindPFlag(cfgEnrichSnmpTimeoutKey, pflags.Lookup(cfgEnrichSnmpTimeoutKey))

	pflags.StringSliceVar(
		&(cfg.SNMP.Community),
		cfgEnrichSnmpCommunityKey,
		cfg.SNMP.Community,
		"Community strings to try during snmp probing",
	)
	viper.BindPFlag(cfgEnrichSnmpCommunityKey, pflags.Lookup(cfgEnrichSnmpCommunityKey))

	pflags.IntSliceVar(
		&(cfg.SNMP.Ports),
		cfgEnrichSnmpPortsKey,
		cfg.SNMP.Ports,
		"Community strings to try during snmp probing",
	)
	viper.BindPFlag(cfgEnrichSnmpPortsKey, pflags.Lookup(cfgEnrichSnmpPortsKey))
}
