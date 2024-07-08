// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package enrichment

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type (
	Config struct {
		Enabled    bool
		MaxWorkers int
		Dns        *DnsConfig
		Oui        *OuiConfig
		PortScan   *PortScanConfig
		Snmp       *SnmpConfig
	}

	DnsConfig struct {
		Enabled bool
	}

	OuiConfig struct {
		Enabled bool
	}

	PortScanConfig struct {
		Enabled             bool
		Timeout             time.Duration
		MaxWorkers          int
		DefaultScanInterval time.Duration
		ServerScanInterval  time.Duration
		PortList            string
	}

	SnmpConfig struct {
		Enabled   bool
		Timeout   time.Duration
		Community []string
		Ports     []int
	}
)

func SetFlags(fs *pflag.FlagSet, cfg *Config) {
	cfg.Dns = &DnsConfig{}
	cfg.Oui = &OuiConfig{}
	cfg.PortScan = &PortScanConfig{}
	cfg.Snmp = &SnmpConfig{}

	configMajorKey := "enrichment"

	flagset.Bool(
		fs,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		true,
		"enable discovery of additional attributes for devices (dns, manufacturer, portscan, snmp)",
	)
	flagset.Int(
		fs,
		&cfg.MaxWorkers,
		configMajorKey,
		"maxworkers",
		2,
		"max number of devices to simultaneously enrich",
	)

	dnsConfigMajorKey := flagset.Key(configMajorKey, "dns")
	flagset.Bool(
		fs,
		&cfg.Dns.Enabled,
		dnsConfigMajorKey,
		"enabled",
		true,
		"use reverse ip dns lookup",
	)

	ouiConfigMajorKey := flagset.Key(configMajorKey, "oui")
	flagset.Bool(
		fs,
		&cfg.Oui.Enabled,
		ouiConfigMajorKey,
		"enabled",
		true,
		"lookup device MAC in oui table to determine manufacturer",
	)

	psConfigMajorKey := flagset.Key(configMajorKey, "portscan")
	flagset.Bool(
		fs,
		&cfg.PortScan.Enabled,
		psConfigMajorKey,
		"enabled",
		true,
		"perform portscan against device",
	)
	flagset.Int(
		fs,
		&cfg.PortScan.MaxWorkers,
		psConfigMajorKey,
		"maxworkers",
		2,
		"max portscan workers",
	)
	flagset.Duration(
		fs,
		&cfg.PortScan.Timeout,
		psConfigMajorKey,
		"timeout",
		20*time.Millisecond,
		"amount of time to wait for a response to a open port request",
	)
	flagset.Duration(
		fs,
		&cfg.PortScan.DefaultScanInterval,
		psConfigMajorKey,
		"defaultscaninterval",
		7*24*time.Hour,
		"duration between port scans for non-server devices",
	)
	flagset.Duration(
		fs,
		&cfg.PortScan.ServerScanInterval,
		psConfigMajorKey,
		"serverscaninterval",
		24*time.Hour,
		"duration between port scans for server devices",
	)
	flagset.String(
		fs,
		&cfg.PortScan.PortList,
		psConfigMajorKey,
		"portlist",
		"general",
		"portlist set to use for scanning [all,general,privileged,common]",
	)

	snmpConfigMajorKey := flagset.Key(configMajorKey, "snmp")
	flagset.Bool(
		fs,
		&cfg.Snmp.Enabled,
		snmpConfigMajorKey,
		"enabled",
		true,
		"enable snmp scanning against device",
	)
	flagset.Duration(
		fs,
		&cfg.Snmp.Timeout,
		snmpConfigMajorKey,
		"timeout",
		50*time.Millisecond,
		"max time to wait for snmp response",
	)
	flagset.StringSlice(
		fs,
		&cfg.Snmp.Community,
		snmpConfigMajorKey,
		"community",
		[]string{"public"},
		"list of strings to test for community string",
	)
	flagset.IntSlice(
		fs,
		&cfg.Snmp.Ports,
		snmpConfigMajorKey,
		"ports",
		[]int{161},
		"list of ports to test for snmp port",
	)
}
