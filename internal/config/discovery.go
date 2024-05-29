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
	discoveryCfgKey               = "discovery"
	arpPingCfgKey                 = "arpping"
	icmpPingCfgKey                = "icmpping"
	bootstrapOnFirstRunCfgKey     = "bootstraponfirstrun"
	networkScanIntervalCfgKey     = "networkscaninterval"
	maxNetworkScannersCfgKey      = "maxnetworkscanners"
	autoDiscoverNewNetworksCfgKey = "autodiscovernewnetworks"
	arpTableRescanIntervalCfgKey  = "arptablerescaninterval"
	interfaceRescanIntervalCfgKey = "interfacerescaninterval"
)

type (
	Discovery struct {
		Enabled                 *bool
		BootstrapOnFirstRun     *bool
		AutoDiscoverNewNetworks *bool
		CheckInterval           *time.Duration
		NetworkScanInterval     *time.Duration
		MaxWorkers              *int
		MaxNetworkScanners      *int
		ArpPing                 *DiscoveryArpConfig
		IcmpPing                *DiscoveryICMPConfig
		Snmp                    *DiscoverySNMPConfig
	}

	DiscoveryArpConfig struct {
		Enabled *bool
		Timeout *time.Duration
	}

	DiscoveryICMPConfig struct {
		Enabled   *bool
		Timeout   *time.Duration
		PingCount *int
	}

	DiscoverySNMPConfig struct {
		Enabled                 *bool
		Timeout                 *time.Duration
		Community               []string
		Ports                   []int
		ArpTableRescanInterval  *time.Duration
		InterfaceRescanInterval *time.Duration
	}
)

var (
	cfgDiscoEnabledKey              = Key(discoveryCfgKey, EnabledCfgKey)
	cfgDiscoBootstrapOnFirstRunKey  = Key(discoveryCfgKey, bootstrapOnFirstRunCfgKey)
	cfgDiscoAutoDiscoNewNetworksKey = Key(
		discoveryCfgKey,
		autoDiscoverNewNetworksCfgKey,
	)
	cfgDiscoCheckIntervalKey       = Key(discoveryCfgKey, CheckIntervalCfgKey)
	cfgDiscoNetworkScanIntervalKey = Key(discoveryCfgKey, networkScanIntervalCfgKey)
	cfgDiscoMaxWorkersKey          = Key(discoveryCfgKey, MaxWorkersCfgKey)
	cfgDiscoMaxNetworkScannersKey  = Key(discoveryCfgKey, maxNetworkScannersCfgKey)
	cfgDiscoArpPingEnabledKey      = Key(
		discoveryCfgKey,
		arpPingCfgKey,
		EnabledCfgKey,
	)
	cfgDiscoArpPingTimeoutKey = Key(
		discoveryCfgKey,
		arpPingCfgKey,
		TimeoutCfgKey,
	)
	cfgDiscoIcmpPingEnabledKey = Key(
		discoveryCfgKey,
		icmpPingCfgKey,
		EnabledCfgKey,
	)
	cfgDiscoIcmpPingTimeoutKey = Key(
		discoveryCfgKey,
		icmpPingCfgKey,
		TimeoutCfgKey,
	)
	cfgDiscoIcmpPingPingCountKey = Key(
		discoveryCfgKey,
		icmpPingCfgKey,
		PingCountCfgKey,
	)
	cfgDiscoSnmpEnabledKey = Key(
		discoveryCfgKey,
		SnmpCfgKey,
		EnabledCfgKey,
	)
	cfgDiscoSnmpTimeoutKey = Key(
		discoveryCfgKey,
		SnmpCfgKey,
		TimeoutCfgKey,
	)
	cfgDiscoSnmpCommunityKey = Key(
		discoveryCfgKey,
		SnmpCfgKey,
		CommunityCfgKey,
	)
	cfgDiscoSnmpPortsKey = Key(
		discoveryCfgKey,
		SnmpCfgKey,
		PortsCfgKey,
	)
	cfgDiscoSnmpArpTableRescanIntervalKey = Key(
		discoveryCfgKey,
		SnmpCfgKey,
		arpTableRescanIntervalCfgKey,
	)
	cfgDiscoSnmpInterfaceRescanIntervalKey = Key(
		discoveryCfgKey,
		SnmpCfgKey,
		interfaceRescanIntervalCfgKey,
	)
)

func cfgDiscoSetDefaults() {
	viper.SetDefault(cfgDiscoEnabledKey, true)
	viper.SetDefault(cfgDiscoBootstrapOnFirstRunKey, true)
	viper.SetDefault(cfgDiscoAutoDiscoNewNetworksKey, true)
	viper.SetDefault(cfgDiscoCheckIntervalKey, 30*time.Minute)
	viper.SetDefault(cfgDiscoNetworkScanIntervalKey, 24*time.Hour)
	viper.SetDefault(cfgDiscoMaxWorkersKey, 2)
	viper.SetDefault(cfgDiscoMaxNetworkScannersKey, 1)
	viper.SetDefault(cfgDiscoArpPingEnabledKey, true)
	viper.SetDefault(cfgDiscoArpPingTimeoutKey, 50*time.Millisecond)
	viper.SetDefault(cfgDiscoIcmpPingEnabledKey, true)
	viper.SetDefault(cfgDiscoIcmpPingTimeoutKey, 100*time.Millisecond)
	viper.SetDefault(cfgDiscoIcmpPingPingCountKey, 2)
	viper.SetDefault(cfgDiscoSnmpEnabledKey, true)
	viper.SetDefault(cfgDiscoSnmpTimeoutKey, 50*time.Millisecond)
	viper.SetDefault(cfgDiscoSnmpCommunityKey, []string{"public"})
	viper.SetDefault(cfgDiscoSnmpPortsKey, []int{161})
	viper.SetDefault(cfgDiscoSnmpArpTableRescanIntervalKey, time.Hour)
	viper.SetDefault(cfgDiscoSnmpInterfaceRescanIntervalKey, 24*time.Hour)
}

func CfgDiscoBuildAndBindFlags(pflags *pflag.FlagSet, cfg *Discovery) {
	pflags.BoolVar(
		cfg.Enabled,
		cfgDiscoEnabledKey,
		*cfg.Enabled,
		"Allow automatic discovery operate",
	)
	viper.BindPFlag(cfgDiscoEnabledKey, pflags.Lookup(cfgDiscoEnabledKey))

	pflags.BoolVar(
		cfg.BootstrapOnFirstRun,
		cfgDiscoBootstrapOnFirstRunKey,
		*cfg.BootstrapOnFirstRun,
		"Run bootstrap on startup, if no saved networks found",
	)
	viper.BindPFlag(cfgDiscoBootstrapOnFirstRunKey, pflags.Lookup(cfgDiscoBootstrapOnFirstRunKey))

	pflags.BoolVar(
		cfg.AutoDiscoverNewNetworks,
		cfgDiscoAutoDiscoNewNetworksKey,
		*cfg.AutoDiscoverNewNetworks,
		"When a new network is saved, run discovery on the network space",
	)
	viper.BindPFlag(cfgDiscoAutoDiscoNewNetworksKey, pflags.Lookup(cfgDiscoAutoDiscoNewNetworksKey))

	pflags.DurationVar(
		cfg.CheckInterval,
		cfgDiscoCheckIntervalKey,
		*cfg.CheckInterval,
		"When a new network is saved, run discovery on the network space",
	)
	viper.BindPFlag(cfgDiscoCheckIntervalKey, pflags.Lookup(cfgDiscoCheckIntervalKey))

	pflags.DurationVar(
		cfg.NetworkScanInterval,
		cfgDiscoNetworkScanIntervalKey,
		*cfg.NetworkScanInterval,
		"Interval between network discovery scans",
	)
	viper.BindPFlag(cfgDiscoNetworkScanIntervalKey, pflags.Lookup(cfgDiscoNetworkScanIntervalKey))

	pflags.IntVar(
		cfg.MaxWorkers,
		cfgDiscoMaxWorkersKey,
		*cfg.MaxWorkers,
		"Max network address space worker pool size",
	)
	viper.BindPFlag(cfgDiscoMaxWorkersKey, pflags.Lookup(cfgDiscoMaxWorkersKey))

	pflags.IntVar(
		cfg.MaxNetworkScanners,
		cfgDiscoMaxNetworkScannersKey,
		*cfg.MaxNetworkScanners,
		"Max network scanner pool size (enumerates IPs to feed device discovery, leave at 1)",
	)
	viper.BindPFlag(cfgDiscoMaxNetworkScannersKey, pflags.Lookup(cfgDiscoMaxNetworkScannersKey))

	// ARP
	pflags.BoolVar(
		cfg.ArpPing.Enabled,
		cfgDiscoArpPingEnabledKey,
		*cfg.ArpPing.Enabled,
		"Allow ARP requests be used during discovery",
	)
	viper.BindPFlag(cfgDiscoArpPingEnabledKey, pflags.Lookup(cfgDiscoArpPingEnabledKey))

	pflags.DurationVar(
		cfg.ArpPing.Timeout,
		cfgDiscoArpPingTimeoutKey,
		*cfg.ArpPing.Timeout,
		"How long to wait for a response to an ARP request",
	)
	viper.BindPFlag(cfgDiscoArpPingTimeoutKey, pflags.Lookup(cfgDiscoArpPingTimeoutKey))

	// ICMP
	pflags.BoolVar(
		cfg.IcmpPing.Enabled,
		cfgDiscoIcmpPingEnabledKey,
		*cfg.IcmpPing.Enabled,
		"Allow ICMP echo requests be used during discovery",
	)
	viper.BindPFlag(cfgDiscoIcmpPingEnabledKey, pflags.Lookup(cfgDiscoIcmpPingEnabledKey))

	pflags.DurationVar(
		cfg.IcmpPing.Timeout,
		cfgDiscoIcmpPingTimeoutKey,
		*cfg.IcmpPing.Timeout,
		"How long to wait for a response to an ICMP echo request",
	)
	viper.BindPFlag(cfgDiscoIcmpPingTimeoutKey, pflags.Lookup(cfgDiscoIcmpPingTimeoutKey))

	pflags.IntVar(
		cfg.IcmpPing.PingCount,
		cfgDiscoIcmpPingPingCountKey,
		*cfg.IcmpPing.PingCount,
		"How many ICMP echo requests to send (only 1 needs to be received)",
	)
	viper.BindPFlag(cfgDiscoIcmpPingPingCountKey, pflags.Lookup(cfgDiscoIcmpPingPingCountKey))

	// SNMP
	pflags.BoolVar(
		cfg.Snmp.Enabled,
		cfgDiscoSnmpEnabledKey,
		*cfg.Snmp.Enabled,
		"Allow SNMP requests be used during discovery",
	)
	viper.BindPFlag(cfgDiscoSnmpEnabledKey, pflags.Lookup(cfgDiscoSnmpEnabledKey))

	pflags.DurationVar(
		cfg.Snmp.Timeout,
		cfgDiscoSnmpTimeoutKey,
		*cfg.Snmp.Timeout,
		"How long to wait for a response to an SNMP discovery request",
	)
	viper.BindPFlag(cfgDiscoSnmpTimeoutKey, pflags.Lookup(cfgDiscoSnmpTimeoutKey))

	pflags.StringSliceVar(
		&(cfg.Snmp.Community),
		cfgDiscoSnmpCommunityKey,
		cfg.Snmp.Community,
		"How long to wait for a response to an SNMP discovery request",
	)
	viper.BindPFlag(cfgDiscoSnmpCommunityKey, pflags.Lookup(cfgDiscoSnmpCommunityKey))

	pflags.IntSliceVar(
		&(cfg.Snmp.Ports),
		cfgDiscoSnmpPortsKey,
		cfg.Snmp.Ports,
		"How long to wait for a response to an SNMP discovery request",
	)
	viper.BindPFlag(cfgDiscoSnmpPortsKey, pflags.Lookup(cfgDiscoSnmpPortsKey))

	pflags.DurationVar(
		cfg.Snmp.ArpTableRescanInterval,
		cfgDiscoSnmpArpTableRescanIntervalKey,
		*cfg.Snmp.ArpTableRescanInterval,
		"Interval to rescan SNMP ARP tables",
	)
	viper.BindPFlag(
		cfgDiscoSnmpArpTableRescanIntervalKey,
		pflags.Lookup(cfgDiscoSnmpArpTableRescanIntervalKey),
	)

	pflags.DurationVar(
		cfg.Snmp.InterfaceRescanInterval,
		cfgDiscoSnmpInterfaceRescanIntervalKey,
		*cfg.Snmp.InterfaceRescanInterval,
		"Interval to rescan SNMP Interface tables",
	)
	viper.BindPFlag(
		cfgDiscoSnmpInterfaceRescanIntervalKey,
		pflags.Lookup(cfgDiscoSnmpInterfaceRescanIntervalKey),
	)
}
