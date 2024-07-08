// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package discovery

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type (
	Config struct {
		Enabled                 bool
		BootstrapOnFirstRun     bool
		AutoDiscoverNewNetworks bool
		CheckInterval           time.Duration
		NetworkScanInterval     time.Duration
		MaxWorkers              int
		Arp                     *ArpConfig
		Icmp                    *ICMPConfig
		Snmp                    *SNMPConfig
	}

	ArpConfig struct {
		Enabled bool
		Timeout time.Duration
	}

	ICMPConfig struct {
		Enabled      bool
		Privileged   bool
		Timeout      time.Duration
		PingCount    int
		SleepBetween time.Duration
	}

	SNMPConfig struct {
		Enabled                 bool
		Timeout                 time.Duration
		Community               []string
		Ports                   []int
		ArpTableRescanInterval  time.Duration
		InterfaceRescanInterval time.Duration
	}
)

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	cfg.Arp = &ArpConfig{}
	cfg.Icmp = &ICMPConfig{}
	cfg.Snmp = &SNMPConfig{}
	configMajorKey := "discovery"

	// Base
	flagset.Bool(
		fs,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		true,
		"Allow automatic discovery of devices",
	)
	flagset.Bool(
		fs,
		&cfg.BootstrapOnFirstRun,
		configMajorKey,
		"bootstraponfirstrun",
		true,
		"Run bootstrap sequence when no saved networks exist",
	)
	flagset.Bool(
		fs,
		&cfg.AutoDiscoverNewNetworks,
		configMajorKey,
		"autodiscovernewnetworks",
		true,
		"Start discovery of devices once a new network is saved",
	)
	flagset.Duration(
		fs,
		&cfg.CheckInterval,
		configMajorKey,
		"checkinterval",
		time.Hour,
		"interval to check if any networks are ready for rescan",
	)
	flagset.Duration(
		fs,
		&cfg.NetworkScanInterval,
		configMajorKey,
		"networkscaninterval",
		24*time.Hour,
		"Interval between checking for new devices on a network",
	)
	flagset.Int(
		fs,
		&cfg.MaxWorkers,
		configMajorKey,
		"maxworkers",
		2,
		"number of workers to use for device discovery",
	)

	// Arp
	arpMajorKey := flagset.Key(configMajorKey, "arp")

	flagset.Bool(
		fs,
		&cfg.Arp.Enabled,
		arpMajorKey,
		"enabled",
		false,
		"use arp ping during discovery (requires net admin privileges)",
	)
	flagset.Duration(
		fs,
		&cfg.Arp.Timeout,
		arpMajorKey,
		"timeout",
		10*time.Millisecond,
		"how long to wait for an arp ping reply",
	)

	// Icmp
	icmpMajorKey := flagset.Key(configMajorKey, "icmp")
	flagset.Bool(
		fs,
		&cfg.Icmp.Enabled,
		icmpMajorKey,
		"enabled",
		true,
		"use icmp ping during discovery",
	)
	flagset.Bool(
		fs,
		&cfg.Icmp.Privileged,
		icmpMajorKey,
		"privileged",
		false,
		"use raw socket access (requires net admin privileges)",
	)
	flagset.Duration(
		fs,
		&cfg.Icmp.Timeout,
		icmpMajorKey,
		"timeout",
		10*time.Millisecond,
		"how long to wait for icmp replies",
	)
	flagset.Int(
		fs,
		&cfg.Icmp.PingCount,
		icmpMajorKey,
		"pingcount",
		2,
		"number of requests to send (only 1 needs to come back)",
	)
	flagset.Duration(
		fs,
		&cfg.Icmp.SleepBetween,
		icmpMajorKey,
		"sleepbetween",
		time.Millisecond,
		"sleep time in-between pings",
	)

	// Snmp
	snmpMajorKey := flagset.Key(configMajorKey, "snmp")
	flagset.Bool(
		fs,
		&cfg.Snmp.Enabled,
		snmpMajorKey,
		"enabled",
		true,
		"use snmp during discovery (arp and interface table data)",
	)
	flagset.Duration(
		fs,
		&cfg.Snmp.Timeout,
		snmpMajorKey,
		"timeout",
		15*time.Millisecond,
		"how long to wait for snmp reply",
	)
	flagset.StringSlice(
		fs,
		&cfg.Snmp.Community,
		snmpMajorKey,
		"community",
		[]string{"public"},
		"community strings to test during discovery",
	)
	flagset.IntSlice(
		fs,
		&cfg.Snmp.Ports,
		snmpMajorKey,
		"ports",
		[]int{161},
		"ports to test during discovery",
	)
	flagset.Duration(
		fs,
		&cfg.Snmp.ArpTableRescanInterval,
		snmpMajorKey,
		"arptablerescaninterval",
		time.Hour,
		"time between arp table scans",
	)
	flagset.Duration(
		fs,
		&cfg.Snmp.InterfaceRescanInterval,
		snmpMajorKey,
		"interfacerescaninterval",
		24*time.Hour,
		"time between interface table scans",
	)
}
